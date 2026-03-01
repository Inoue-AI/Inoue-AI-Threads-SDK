"""Tests for the Reply Management API."""

from __future__ import annotations

import re

import pytest
from aioresponses import aioresponses

from tests.conftest import BASE
from threads.client import ThreadsClient


@pytest.mark.asyncio
async def test_get_replies(
    mock_api: aioresponses,
    client: ThreadsClient,
) -> None:
    mock_api.get(
        re.compile(rf"^{re.escape(BASE)}/post_1/replies\?"),
        payload={
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
    mock_api: aioresponses,
    client: ThreadsClient,
) -> None:
    mock_api.get(
        re.compile(rf"^{re.escape(BASE)}/post_1/conversation\?"),
        payload={
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
    mock_api: aioresponses,
    client: ThreadsClient,
) -> None:
    mock_api.post(
        re.compile(rf"^{re.escape(BASE)}/reply_1/manage_reply\?"),
        payload={"success": True},
    )

    result = await client.replies.hide_reply("reply_1")
    assert result is True


@pytest.mark.asyncio
async def test_unhide_reply(
    mock_api: aioresponses,
    client: ThreadsClient,
) -> None:
    mock_api.post(
        re.compile(rf"^{re.escape(BASE)}/reply_1/manage_reply\?"),
        payload={"success": True},
    )

    result = await client.replies.unhide_reply("reply_1")
    assert result is True


@pytest.mark.asyncio
async def test_get_mentions(
    mock_api: aioresponses,
    client: ThreadsClient,
) -> None:
    mock_api.get(
        re.compile(rf"^{re.escape(BASE)}/me/mentions\?"),
        payload={
            "data": [
                {"id": "m1", "text": "Hey @user check this!", "username": "someone"},
            ],
        },
    )

    result = await client.replies.get_mentions()
    assert len(result.data) == 1
    assert result.data[0].id == "m1"
