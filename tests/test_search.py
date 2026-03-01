"""Tests for the Keyword Search API."""

from __future__ import annotations

import re

import pytest
from aioresponses import aioresponses

from tests.conftest import BASE
from threads.client import ThreadsClient
from threads.models.search import SearchMediaType


@pytest.mark.asyncio
async def test_search_basic(
    mock_api: aioresponses,
    client: ThreadsClient,
) -> None:
    mock_api.get(
        re.compile(rf"^{re.escape(BASE)}/keyword_search\?"),
        payload={
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


@pytest.mark.asyncio
async def test_search_with_media_type(
    mock_api: aioresponses,
    client: ThreadsClient,
) -> None:
    mock_api.get(
        re.compile(rf"^{re.escape(BASE)}/keyword_search\?"),
        payload={"data": [{"id": "v1", "media_type": "VIDEO"}]},
    )

    result = await client.search.search("tutorial", media_type=SearchMediaType.VIDEO)
    assert len(result.data) == 1


@pytest.mark.asyncio
async def test_search_hashtag(
    mock_api: aioresponses,
    client: ThreadsClient,
) -> None:
    mock_api.get(
        re.compile(rf"^{re.escape(BASE)}/keyword_search\?"),
        payload={"data": [{"id": "h1", "text": "#technology rocks"}]},
    )

    result = await client.search.search("#technology")
    assert len(result.data) == 1
