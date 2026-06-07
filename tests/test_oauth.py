"""Tests for the Threads OAuth 2.0 client.

All token endpoints are mocked at the transport layer. No client secret is
ever sent to a real host and no credentials are required to run the suite.
"""

from __future__ import annotations

from urllib.parse import parse_qs, urlparse

import httpx
import pytest

from threads.exceptions import ThreadsAPIError, ThreadsAuthError, ThreadsConfigError
from threads.models.oauth import ThreadsScope
from threads.oauth import ThreadsOAuth

CLIENT_ID = "1234567890"
CLIENT_SECRET = "dummy-secret-for-tests"  # not a real secret; mocked transport only
REDIRECT_URI = "https://app.example.com/callback"


def _oauth(transport: httpx.MockTransport | None = None) -> ThreadsOAuth:
    return ThreadsOAuth(
        client_id=CLIENT_ID,
        client_secret=CLIENT_SECRET,
        redirect_uri=REDIRECT_URI,
        transport=transport,
    )


def test_requires_credentials() -> None:
    with pytest.raises(ThreadsConfigError, match="client_id"):
        ThreadsOAuth(client_id="", client_secret="s", redirect_uri="r")
    with pytest.raises(ThreadsConfigError, match="client_secret"):
        ThreadsOAuth(client_id="c", client_secret="", redirect_uri="r")
    with pytest.raises(ThreadsConfigError, match="redirect_uri"):
        ThreadsOAuth(client_id="c", client_secret="s", redirect_uri="")


def test_authorization_url() -> None:
    oauth = _oauth()
    url = oauth.authorization_url(
        [ThreadsScope.BASIC, ThreadsScope.CONTENT_PUBLISH],
        state="csrf123",
    )
    parsed = urlparse(url)
    qs = parse_qs(parsed.query)
    assert parsed.path == "/oauth/authorize"
    assert qs["client_id"] == [CLIENT_ID]
    assert qs["redirect_uri"] == [REDIRECT_URI]
    assert qs["scope"] == ["threads_basic,threads_content_publish"]
    assert qs["response_type"] == ["code"]
    assert qs["state"] == ["csrf123"]
    # The client secret must never appear in the authorization URL.
    assert CLIENT_SECRET not in url


def test_authorization_url_requires_scope() -> None:
    oauth = _oauth()
    with pytest.raises(ThreadsConfigError, match="at least one scope"):
        oauth.authorization_url([])


@pytest.mark.asyncio
async def test_exchange_code() -> None:
    captured: dict[str, str] = {}

    def handler(request: httpx.Request) -> httpx.Response:
        assert request.url.path == "/oauth/access_token"
        body = parse_qs(request.content.decode())
        captured.update({k: v[0] for k, v in body.items()})
        return httpx.Response(200, json={"access_token": "SHORT", "user_id": 9988})

    async with _oauth(httpx.MockTransport(handler)) as oauth:
        tok = await oauth.exchange_code("auth_code_xyz")

    assert tok.access_token == "SHORT"
    assert tok.user_id == 9988
    assert captured["grant_type"] == "authorization_code"
    assert captured["code"] == "auth_code_xyz"
    assert captured["client_secret"] == CLIENT_SECRET
    assert captured["redirect_uri"] == REDIRECT_URI


@pytest.mark.asyncio
async def test_exchange_for_long_lived() -> None:
    def handler(request: httpx.Request) -> httpx.Response:
        assert request.url.path == "/access_token"
        qs = parse_qs(request.url.query.decode())
        assert qs["grant_type"] == ["th_exchange_token"]
        assert qs["client_secret"] == [CLIENT_SECRET]
        assert qs["access_token"] == ["SHORT"]
        return httpx.Response(
            200,
            json={"access_token": "LONG", "token_type": "bearer", "expires_in": 5183944},
        )

    async with _oauth(httpx.MockTransport(handler)) as oauth:
        tok = await oauth.exchange_for_long_lived("SHORT")

    assert tok.access_token == "LONG"
    assert tok.expires_in == 5183944


@pytest.mark.asyncio
async def test_refresh_long_lived() -> None:
    def handler(request: httpx.Request) -> httpx.Response:
        assert request.url.path == "/refresh_access_token"
        qs = parse_qs(request.url.query.decode())
        assert qs["grant_type"] == ["th_refresh_token"]
        assert qs["access_token"] == ["LONG_OLD"]
        return httpx.Response(
            200,
            json={"access_token": "LONG_NEW", "token_type": "bearer", "expires_in": 5183944},
        )

    async with _oauth(httpx.MockTransport(handler)) as oauth:
        tok = await oauth.refresh_long_lived("LONG_OLD")

    assert tok.access_token == "LONG_NEW"


@pytest.mark.asyncio
async def test_exchange_code_requires_code() -> None:
    async with _oauth() as oauth:
        with pytest.raises(ThreadsConfigError, match="code"):
            await oauth.exchange_code("")


@pytest.mark.asyncio
async def test_oauth_error_graph_envelope() -> None:
    def handler(request: httpx.Request) -> httpx.Response:
        return httpx.Response(
            400,
            json={
                "error": {
                    "message": "Invalid verification code format.",
                    "type": "OAuthException",
                    "code": 100,
                    "fbtrace_id": "tracexyz",
                },
            },
        )

    async with _oauth(httpx.MockTransport(handler)) as oauth:
        with pytest.raises(ThreadsAuthError) as exc:
            await oauth.exchange_code("bad")
    assert exc.value.fbtrace_id == "tracexyz"


@pytest.mark.asyncio
async def test_oauth_error_flat_envelope() -> None:
    def handler(request: httpx.Request) -> httpx.Response:
        return httpx.Response(
            400,
            json={"error": "invalid_grant", "error_description": "code expired"},
        )

    async with _oauth(httpx.MockTransport(handler)) as oauth:
        with pytest.raises(ThreadsAPIError) as exc:
            await oauth.exchange_code("bad")
    assert "code expired" in exc.value.message
    assert exc.value.error_type == "invalid_grant"
