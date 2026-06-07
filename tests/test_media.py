"""Tests for the Media / Threads Retrieval API."""

from __future__ import annotations

import pytest

from tests.conftest import MockRouter
from threads.client import ThreadsClient
from threads.models.media import MediaField, ThreadMediaType


@pytest.mark.asyncio
async def test_list_threads(
    router: MockRouter,
    client: ThreadsClient,
) -> None:
    router.add(
        "/me/threads",
        json={
            "data": [
                {
                    "id": "t1",
                    "media_type": "TEXT_POST",
                    "text": "Hello",
                    "timestamp": "2024-06-01T12:00:00+0000",
                },
                {
                    "id": "t2",
                    "media_type": "IMAGE",
                    "text": "Photo",
                },
            ],
            "paging": {
                "cursors": {"before": "abc", "after": "xyz"},
            },
        },
    )

    result = await client.media.list_threads()
    assert len(result.data) == 2
    assert result.data[0].id == "t1"
    assert result.data[0].media_type == ThreadMediaType.TEXT_POST
    assert result.data[1].media_type == ThreadMediaType.IMAGE
    assert result.paging is not None
    assert result.paging.cursors is not None
    assert result.paging.cursors.after == "xyz"


@pytest.mark.asyncio
async def test_get_thread(
    router: MockRouter,
    client: ThreadsClient,
) -> None:
    router.add(
        "/t1",
        json={
            "id": "t1",
            "media_type": "VIDEO",
            "media_url": "https://example.com/vid.mp4",
            "permalink": "https://threads.net/@user/post/t1",
            "text": "Check this out",
            "owner": {"id": "12345"},
        },
    )

    thread = await client.media.get_thread("t1")
    assert thread.id == "t1"
    assert thread.media_type == ThreadMediaType.VIDEO
    assert thread.owner is not None
    assert thread.owner.id == "12345"
    assert thread.permalink == "https://threads.net/@user/post/t1"


@pytest.mark.asyncio
async def test_list_threads_with_fields(
    router: MockRouter,
    client: ThreadsClient,
) -> None:
    router.add("/me/threads", json={"data": [{"id": "t3", "text": "Hi"}]})

    result = await client.media.list_threads(
        fields=[MediaField.ID, MediaField.TEXT],
    )
    assert len(result.data) == 1
    assert result.data[0].text == "Hi"
    assert router.requests[-1].url.params["fields"] == "id,text"


@pytest.mark.asyncio
async def test_iter_threads(
    router: MockRouter,
    client: ThreadsClient,
) -> None:
    # Page 1 (has a cursor) then page 2 (terminal — no cursor).
    router.add(
        "/me/threads",
        json={
            "data": [{"id": "t1"}, {"id": "t2"}],
            "paging": {"cursors": {"after": "cursor_1"}},
        },
    )
    router.add(
        "/me/threads",
        json={
            "data": [{"id": "t3"}],
            "paging": {"cursors": {}},
        },
    )

    ids: list[str] = []
    async for thread in client.media.iter_threads():
        ids.append(thread.id or "")

    assert ids == ["t1", "t2", "t3"]


@pytest.mark.asyncio
async def test_get_thread_carousel_children(
    router: MockRouter,
    client: ThreadsClient,
) -> None:
    router.add(
        "/t_carousel",
        json={
            "id": "t_carousel",
            "media_type": "CAROUSEL_ALBUM",
            "children": {
                "data": [{"id": "child_1"}, {"id": "child_2"}],
            },
        },
    )

    thread = await client.media.get_thread("t_carousel")
    assert thread.media_type == ThreadMediaType.CAROUSEL_ALBUM
    assert thread.children is not None
    assert len(thread.children.data) == 2
    assert thread.children.data[0].id == "child_1"
