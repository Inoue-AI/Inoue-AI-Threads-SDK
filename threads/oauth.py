"""Threads OAuth 2.0 client — authorization URL, code exchange, token refresh.

This is an *app-level* flow: it uses your Threads app's ``client_id`` and
``client_secret`` to mint and refresh user access tokens. It does not require a
pre-existing user token (that is what it produces), so it is implemented as a
standalone client rather than a namespace on :class:`~threads.client.ThreadsClient`.

Flow overview::

    1. authorization_url(...)          -> redirect the user, receive ?code=...
    2. exchange_code(code)             -> ShortLivedToken (~1 hour)
    3. exchange_for_long_lived(token)  -> LongLivedToken  (~60 days)
    4. refresh_long_lived(token)       -> LongLivedToken  (refreshed ~60 days)

No vendor secret is ever logged or embedded; credentials are supplied by the
caller (sourced from environment/secret store) and only ever sent over TLS to
the Threads OAuth host.
"""

from __future__ import annotations

from typing import Any
from urllib.parse import urlencode

import httpx

from threads.config import (
    DEFAULT_MAX_CONNECTIONS,
    DEFAULT_MAX_KEEPALIVE_CONNECTIONS,
    DEFAULT_TIMEOUT,
    THREADS_AUTHORIZE_BASE_URL,
    THREADS_OAUTH_BASE_URL,
)
from threads.exceptions import (
    ThreadsAPIError,
    ThreadsConfigError,
    build_api_error,
)
from threads.models.oauth import LongLivedToken, ShortLivedToken, ThreadsScope

_GRANT_AUTHORIZATION_CODE = "authorization_code"
_GRANT_EXCHANGE_TOKEN = "th_exchange_token"
_GRANT_REFRESH_TOKEN = "th_refresh_token"


class ThreadsOAuth:
    """Client for the Threads OAuth 2.0 authorization-code flow.

    Parameters
    ----------
    client_id:
        Your Threads app's client id (Meta App ID).
    client_secret:
        Your Threads app's client secret. Sourced from a secret store /
        environment variable — never hard-code it.
    redirect_uri:
        The redirect URI registered with your Threads app.
    timeout:
        Total per-request timeout in seconds.
    authorize_base_url / oauth_base_url:
        Overridable hosts (used by tests to point at a mock transport).
    transport:
        Optional :class:`httpx.AsyncBaseTransport` injected for testing.
    """

    __slots__ = (
        "_client_id",
        "_client_secret",
        "_redirect_uri",
        "_authorize_base_url",
        "_oauth_base_url",
        "_timeout",
        "_transport",
        "_client",
    )

    def __init__(
        self,
        client_id: str,
        client_secret: str,
        redirect_uri: str,
        *,
        timeout: float = DEFAULT_TIMEOUT,
        authorize_base_url: str = THREADS_AUTHORIZE_BASE_URL,
        oauth_base_url: str = THREADS_OAUTH_BASE_URL,
        transport: httpx.AsyncBaseTransport | None = None,
    ) -> None:
        if not client_id:
            raise ThreadsConfigError("client_id must not be empty.")
        if not client_secret:
            raise ThreadsConfigError("client_secret must not be empty.")
        if not redirect_uri:
            raise ThreadsConfigError("redirect_uri must not be empty.")

        self._client_id = client_id
        self._client_secret = client_secret
        self._redirect_uri = redirect_uri
        self._authorize_base_url = authorize_base_url.rstrip("/")
        self._oauth_base_url = oauth_base_url.rstrip("/")
        self._timeout = httpx.Timeout(timeout)
        self._transport = transport
        self._client = httpx.AsyncClient(
            base_url=self._oauth_base_url,
            timeout=self._timeout,
            limits=httpx.Limits(
                max_connections=DEFAULT_MAX_CONNECTIONS,
                max_keepalive_connections=DEFAULT_MAX_KEEPALIVE_CONNECTIONS,
            ),
            transport=self._transport,
        )

    async def __aenter__(self) -> ThreadsOAuth:
        return self

    async def __aexit__(self, *exc: object) -> None:
        await self.aclose()

    async def aclose(self) -> None:
        """Close the underlying HTTP client and release pooled connections."""
        if not self._client.is_closed:
            await self._client.aclose()

    # -- step 1: authorization URL -------------------------------------------

    def authorization_url(
        self,
        scopes: list[ThreadsScope],
        *,
        state: str | None = None,
    ) -> str:
        """Build the URL to redirect the user to for consent.

        Parameters
        ----------
        scopes:
            One or more :class:`~threads.models.oauth.ThreadsScope` values.
        state:
            Opaque CSRF / round-trip token echoed back to your redirect URI.
        """
        if not scopes:
            raise ThreadsConfigError("authorization_url requires at least one scope.")
        params: dict[str, str] = {
            "client_id": self._client_id,
            "redirect_uri": self._redirect_uri,
            "scope": ",".join(scope.value for scope in scopes),
            "response_type": "code",
        }
        if state is not None:
            params["state"] = state
        return f"{self._authorize_base_url}/oauth/authorize?{urlencode(params)}"

    # -- step 2: code -> short-lived token ------------------------------------

    async def exchange_code(self, code: str) -> ShortLivedToken:
        """Exchange an authorization ``code`` for a short-lived access token."""
        if not code:
            raise ThreadsConfigError("exchange_code requires a non-empty code.")
        data = {
            "client_id": self._client_id,
            "client_secret": self._client_secret,
            "grant_type": _GRANT_AUTHORIZATION_CODE,
            "redirect_uri": self._redirect_uri,
            "code": code,
        }
        body = await self._post("/oauth/access_token", data)
        return ShortLivedToken.model_validate(body)

    # -- step 3: short-lived -> long-lived token ------------------------------

    async def exchange_for_long_lived(self, short_lived_token: str) -> LongLivedToken:
        """Exchange a short-lived token for a long-lived (~60 day) token."""
        if not short_lived_token:
            raise ThreadsConfigError(
                "exchange_for_long_lived requires a non-empty short_lived_token."
            )
        params = {
            "grant_type": _GRANT_EXCHANGE_TOKEN,
            "client_secret": self._client_secret,
            "access_token": short_lived_token,
        }
        body = await self._get("/access_token", params)
        return LongLivedToken.model_validate(body)

    # -- step 4: refresh a long-lived token -----------------------------------

    async def refresh_long_lived(self, long_lived_token: str) -> LongLivedToken:
        """Refresh a still-valid long-lived token for a new ~60 day token.

        Threads does not issue a separate refresh token; the existing
        long-lived access token is presented to ``/refresh_access_token``.
        """
        if not long_lived_token:
            raise ThreadsConfigError("refresh_long_lived requires a non-empty long_lived_token.")
        params = {
            "grant_type": _GRANT_REFRESH_TOKEN,
            "access_token": long_lived_token,
        }
        body = await self._get("/refresh_access_token", params)
        return LongLivedToken.model_validate(body)

    # -- internal request helpers --------------------------------------------

    async def _get(self, path: str, params: dict[str, str]) -> dict[str, Any]:
        resp = await self._client.get(path, params=params)
        return self._parse_response(resp)

    async def _post(self, path: str, data: dict[str, str]) -> dict[str, Any]:
        resp = await self._client.post(path, data=data)
        return self._parse_response(resp)

    @staticmethod
    def _parse_response(resp: httpx.Response) -> dict[str, Any]:
        try:
            body: dict[str, Any] = resp.json()
        except ValueError as exc:
            raise ThreadsAPIError(
                f"HTTP {resp.status_code}",
                http_status=resp.status_code,
            ) from exc

        if isinstance(body, dict) and "error" in body:
            err = body["error"]
            # Threads OAuth errors can be either the Graph nested envelope
            # ({"error": {"message": ..., "code": ...}}) or the flat OAuth2
            # form ({"error": "invalid_grant", "error_description": ...}).
            if isinstance(err, dict):
                raise build_api_error(
                    message=err.get("message", "Unknown error"),
                    http_status=resp.status_code,
                    error_type=err.get("type", ""),
                    error_code=err.get("code", 0),
                    error_subcode=err.get("error_subcode", 0),
                    fbtrace_id=err.get("fbtrace_id", ""),
                )
            raise build_api_error(
                message=str(body.get("error_description", err)),
                http_status=resp.status_code,
                error_type=str(err),
                error_code=0,
                error_subcode=0,
                fbtrace_id="",
            )

        if resp.status_code >= 400:
            raise ThreadsAPIError(
                f"HTTP {resp.status_code}",
                http_status=resp.status_code,
            )

        return body
