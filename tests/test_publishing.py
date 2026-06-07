"""Tests for the Content Publishing API."""

from __future__ import annotations

import pytest

from tests.conftest import MockRouter
from threads.client import ThreadsClient
from threads.exceptions import ThreadsConfigError, ThreadsPublishingError
from threads.models.publishing import ContainerStatus


@pytest.mark.asyncio
async def test_create_text_post(
    router: MockRouter,
    client: ThreadsClient,
) -> None:
    router.add("/me/threads", json={"id": "container_1"})

    resp = await client.publishing.create_text_post("me", "Hello Threads!")
    assert resp.id == "container_1"


@pytest.mark.asyncio
async def test_create_image_post(
    router: MockRouter,
    client: ThreadsClient,
) -> None:
    router.add("/me/threads", json={"id": "container_2"})

    resp = await client.publishing.create_image_post(
        "me",
        "https://example.com/image.jpg",
        text="Nice pic",
    )
    assert resp.id == "container_2"


@pytest.mark.asyncio
async def test_create_video_post(
    router: MockRouter,
    client: ThreadsClient,
) -> None:
    router.add("/me/threads", json={"id": "container_3"})

    resp = await client.publishing.create_video_post(
        "me",
        "https://example.com/video.mp4",
    )
    assert resp.id == "container_3"


@pytest.mark.asyncio
async def test_create_carousel_too_few_items(client: ThreadsClient) -> None:
    with pytest.raises(ThreadsConfigError, match="at least 2"):
        await client.publishing.create_carousel_post("me", ["item1"])


@pytest.mark.asyncio
async def test_create_carousel_too_many_items(client: ThreadsClient) -> None:
    with pytest.raises(ThreadsConfigError, match="at most 20"):
        await client.publishing.create_carousel_post("me", [f"i{i}" for i in range(21)])


@pytest.mark.asyncio
async def test_create_carousel_post(
    router: MockRouter,
    client: ThreadsClient,
) -> None:
    router.add("/me/threads", json={"id": "carousel_container"})

    resp = await client.publishing.create_carousel_post(
        "me",
        ["child_1", "child_2", "child_3"],
    )
    assert resp.id == "carousel_container"


@pytest.mark.asyncio
async def test_publish(
    router: MockRouter,
    client: ThreadsClient,
) -> None:
    router.add("/me/threads_publish", json={"id": "published_post_1"})

    resp = await client.publishing.publish("me", "container_1")
    assert resp.id == "published_post_1"


@pytest.mark.asyncio
async def test_get_container_status(
    router: MockRouter,
    client: ThreadsClient,
) -> None:
    router.add("/container_1", json={"id": "container_1", "status": "FINISHED"})

    status = await client.publishing.get_container_status("container_1")
    assert status.status == ContainerStatus.FINISHED
    assert status.error_message is None


@pytest.mark.asyncio
async def test_wait_for_container_finished(
    router: MockRouter,
    client: ThreadsClient,
) -> None:
    router.add("/c1", json={"id": "c1", "status": "IN_PROGRESS"})
    router.add("/c1", json={"id": "c1", "status": "FINISHED"})

    status = await client.publishing.wait_for_container("c1", poll_interval=0.01)
    assert status.status == ContainerStatus.FINISHED


@pytest.mark.asyncio
async def test_wait_for_container_errored(
    router: MockRouter,
    client: ThreadsClient,
) -> None:
    router.add(
        "/c2",
        json={"id": "c2", "status": "ERRORED", "error_message": "Bad format"},
    )

    with pytest.raises(ThreadsPublishingError, match="Bad format"):
        await client.publishing.wait_for_container("c2", poll_interval=0.01)


@pytest.mark.asyncio
async def test_post_text_full_flow(
    router: MockRouter,
    client: ThreadsClient,
) -> None:
    router.add("/me/threads", json={"id": "container_txt"})
    router.add("/me/threads_publish", json={"id": "published_txt"})

    resp = await client.publishing.post_text("me", "Hello!")
    assert resp.id == "published_txt"


@pytest.mark.asyncio
async def test_delete_post(
    router: MockRouter,
    client: ThreadsClient,
) -> None:
    router.add("/thread_99", json={"success": True})

    result = await client.publishing.delete_post("thread_99")
    assert result is True


@pytest.mark.asyncio
async def test_get_publishing_limit(
    router: MockRouter,
    client: ThreadsClient,
) -> None:
    router.add(
        "/me/threads_publishing_limit",
        json={
            "data": [
                {
                    "quota_usage": 5,
                    "config": {"quota_total": 250, "quota_duration": 86400},
                },
            ],
        },
    )

    limit = await client.publishing.get_publishing_limit()
    assert limit.quota_usage == 5
    assert limit.config is not None
    assert limit.config.quota_total == 250
