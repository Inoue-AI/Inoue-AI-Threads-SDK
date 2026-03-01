"""Tests for the User Profile API."""

from __future__ import annotations

import re

import pytest
from aioresponses import aioresponses

from tests.conftest import BASE
from threads.client import ThreadsClient
from threads.models.user import UserField


@pytest.mark.asyncio
async def test_get_profile_defaults(
    mock_api: aioresponses,
    client: ThreadsClient,
) -> None:
    mock_api.get(
        re.compile(rf"^{re.escape(BASE)}/me\?"),
        payload={
            "id": "12345",
            "username": "testuser",
            "name": "Test User",
            "threads_profile_picture_url": "https://example.com/pic.jpg",
            "threads_biography": "Hello world",
            "is_verified": True,
        },
    )

    user = await client.user.get_profile()
    assert user.id == "12345"
    assert user.username == "testuser"
    assert user.name == "Test User"
    assert user.is_verified is True


@pytest.mark.asyncio
async def test_get_profile_specific_fields(
    mock_api: aioresponses,
    client: ThreadsClient,
) -> None:
    mock_api.get(
        re.compile(rf"^{re.escape(BASE)}/me\?"),
        payload={"id": "12345", "username": "testuser"},
    )

    user = await client.user.get_profile(
        fields=[UserField.ID, UserField.USERNAME],
    )
    assert user.id == "12345"
    assert user.username == "testuser"
    assert user.name is None


@pytest.mark.asyncio
async def test_get_profile_other_user(
    mock_api: aioresponses,
    client: ThreadsClient,
) -> None:
    mock_api.get(
        re.compile(rf"^{re.escape(BASE)}/67890\?"),
        payload={"id": "67890", "username": "other"},
    )

    user = await client.user.get_profile("67890")
    assert user.id == "67890"
