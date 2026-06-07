"""Tests for the Keyword Search API."""

from __future__ import annotations

import pytest

from tests.conftest import MockRouter
from threads.client import ThreadsClient
from threads.models.search import SearchMediaType


@pytest.mark.asyncio
async def test_search_basic(
    router: MockRouter,
    client: ThreadsClient,
) -> None:
    router.add(
        "/keyword_search",
        json={
            "data": [
                {"id": "s1", "text": "Python is great", "media_type": "TEXT_POST"},
                {"id": "s2", "text": "Learning Python", "media_type": "TEXT_POST"},
            ],
            "paging": {"cursors": {"after": "next_page"}},
        },
    )

    result = await client.search.search("Python")
    assert len(result.data) == 2
    assert result.data[0].text == "Python is great"
    assert result.paging is not None
    assert router.requests[-1].url.params["q"] == "Python"


@pytest.mark.asyncio
async def test_search_with_media_type(
    router: MockRouter,
    client: ThreadsClient,
) -> None:
    router.add("/keyword_search", json={"data": [{"id": "v1", "media_type": "VIDEO"}]})

    result = await client.search.search("tutorial", media_type=SearchMediaType.VIDEO)
    assert len(result.data) == 1
    assert router.requests[-1].url.params["media_type"] == "VIDEO"


@pytest.mark.asyncio
async def test_search_hashtag(
    router: MockRouter,
    client: ThreadsClient,
) -> None:
    router.add("/keyword_search", json={"data": [{"id": "h1", "text": "#technology rocks"}]})

    result = await client.search.search("#technology")
    assert len(result.data) == 1
