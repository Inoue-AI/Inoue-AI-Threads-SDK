"""Tests for the User Profile API."""

from __future__ import annotations

import pytest

from tests.conftest import MockRouter
from threads.client import ThreadsClient
from threads.models.user import UserField


@pytest.mark.asyncio
async def test_get_profile_defaults(
    router: MockRouter,
    client: ThreadsClient,
) -> None:
    router.add(
        "/me",
        json={
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
    router: MockRouter,
    client: ThreadsClient,
) -> None:
    router.add("/me", json={"id": "12345", "username": "testuser"})

    user = await client.user.get_profile(
        fields=[UserField.ID, UserField.USERNAME],
    )
    assert user.id == "12345"
    assert user.username == "testuser"
    assert user.name is None
    # The requested fields are forwarded as a comma-joined query param.
    sent = router.requests[-1]
    assert sent.url.params["fields"] == "id,username"
    assert sent.url.params["access_token"] == "test_access_token_123"


@pytest.mark.asyncio
async def test_get_profile_other_user(
    router: MockRouter,
    client: ThreadsClient,
) -> None:
    router.add("/67890", json={"id": "67890", "username": "other"})

    user = await client.user.get_profile("67890")
    assert user.id == "67890"
