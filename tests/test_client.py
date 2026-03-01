"""Tests for the ThreadsClient and error handling."""

from __future__ import annotations

import re

import pytest
from aioresponses import aioresponses

from tests.conftest import BASE
from threads.client import ThreadsClient
from threads.exceptions import (
    ThreadsAuthError,
    ThreadsConfigError,
    ThreadsNotFoundError,
    ThreadsRateLimitError,
    ThreadsServerError,
)


def test_empty_token_raises() -> None:
    with pytest.raises(ThreadsConfigError, match="access_token"):
        ThreadsClient(access_token="")


@pytest.mark.asyncio
async def test_context_manager() -> None:
    async with ThreadsClient(access_token="tok") as client:
        assert client.user is not None
        assert client.publishing is not None
        assert client.media is not None
        assert client.insights is not None
        assert client.replies is not None
        assert client.search is not None


@pytest.mark.asyncio
async def test_auth_error(
    mock_api: aioresponses,
    client: ThreadsClient,
) -> None:
    mock_api.get(
        re.compile(rf"^{re.escape(BASE)}/me\?"),
        payload={
            "error": {
                "message": "Invalid OAuth access token",
                "type": "OAuthException",
                "code": 190,
                "error_subcode": 463,
                "fbtrace_id": "trace123",
            },
        },
        status=401,
    )

    with pytest.raises(ThreadsAuthError) as exc_info:
        await client.user.get_profile()

    assert exc_info.value.error_code == 190
    assert exc_info.value.fbtrace_id == "trace123"


@pytest.mark.asyncio
async def test_rate_limit_error(
    mock_api: aioresponses,
    client: ThreadsClient,
) -> None:
    mock_api.get(
        re.compile(rf"^{re.escape(BASE)}/me\?"),
        payload={
            "error": {
                "message": "Rate limit exceeded",
                "type": "OAuthException",
                "code": 4,
                "fbtrace_id": "trace456",
            },
        },
        status=429,
    )

    with pytest.raises(ThreadsRateLimitError):
        await client.user.get_profile()


@pytest.mark.asyncio
async def test_not_found_error(
    mock_api: aioresponses,
    client: ThreadsClient,
) -> None:
    mock_api.get(
        re.compile(rf"^{re.escape(BASE)}/nonexistent\?"),
        payload={
            "error": {
                "message": "Object does not exist",
                "type": "GraphMethodException",
                "code": 100,
                "fbtrace_id": "trace789",
            },
        },
        status=404,
    )

    with pytest.raises(ThreadsNotFoundError):
        await client.media.get_thread("nonexistent")


@pytest.mark.asyncio
async def test_server_error(
    mock_api: aioresponses,
    client: ThreadsClient,
) -> None:
    mock_api.get(
        re.compile(rf"^{re.escape(BASE)}/me\?"),
        payload={
            "error": {
                "message": "Internal server error",
                "type": "ServerException",
                "code": 1,
                "fbtrace_id": "trace000",
            },
        },
        status=500,
    )

    with pytest.raises(ThreadsServerError):
        await client.user.get_profile()


@pytest.mark.asyncio
async def test_unknown_fields_ignored(
    mock_api: aioresponses,
    client: ThreadsClient,
) -> None:
    mock_api.get(
        re.compile(rf"^{re.escape(BASE)}/me\?"),
        payload={
            "id": "12345",
            "username": "testuser",
            "brand_new_field": "should be ignored",
        },
    )

    user = await client.user.get_profile()
    assert user.id == "12345"
    assert user.username == "testuser"
