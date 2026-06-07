"""Tests for the Reply Management API."""

from __future__ import annotations

import pytest

from tests.conftest import MockRouter
from threads.client import ThreadsClient


@pytest.mark.asyncio
async def test_get_replies(
    router: MockRouter,
    client: ThreadsClient,
) -> None:
    router.add(
        "/post_1/replies",
        json={
            "data": [
                {"id": "r1", "text": "Great post!", "username": "fan1"},
                {"id": "r2", "text": "Love it", "username": "fan2"},
            ],
        },
    )

    result = await client.replies.get_replies("post_1")
    assert len(result.data) == 2
    assert result.data[0].text == "Great post!"
    assert result.data[1].username == "fan2"


@pytest.mark.asyncio
async def test_get_conversation(
    router: MockRouter,
    client: ThreadsClient,
) -> None:
    router.add(
        "/post_1/conversation",
        json={
            "data": [
                {"id": "r1", "text": "First reply"},
                {"id": "r1_1", "text": "Reply to reply"},
            ],
        },
    )

    result = await client.replies.get_conversation("post_1")
    assert len(result.data) == 2


@pytest.mark.asyncio
async def test_hide_reply(
    router: MockRouter,
    client: ThreadsClient,
) -> None:
    router.add("/reply_1/manage_reply", json={"success": True})

    result = await client.replies.hide_reply("reply_1")
    assert result is True


@pytest.mark.asyncio
async def test_unhide_reply(
    router: MockRouter,
    client: ThreadsClient,
) -> None:
    router.add("/reply_1/manage_reply", json={"success": True})

    result = await client.replies.unhide_reply("reply_1")
    assert result is True


@pytest.mark.asyncio
async def test_get_mentions(
    router: MockRouter,
    client: ThreadsClient,
) -> None:
    router.add(
        "/me/mentions",
        json={
            "data": [
                {"id": "m1", "text": "Hey @user check this!", "username": "someone"},
            ],
        },
    )

    result = await client.replies.get_mentions()
    assert len(result.data) == 1
    assert result.data[0].id == "m1"
