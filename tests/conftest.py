"""Shared fixtures for the Threads SDK test suite."""

from __future__ import annotations

import pytest
from aioresponses import aioresponses

from threads.client import ThreadsClient
from threads.config import THREADS_API_BASE_URL

TOKEN = "test_access_token_123"
BASE = THREADS_API_BASE_URL


@pytest.fixture
def mock_api() -> aioresponses:  # type: ignore[misc]
    with aioresponses() as m:
        yield m


@pytest.fixture
async def client() -> ThreadsClient:  # type: ignore[misc]
    c = ThreadsClient(access_token=TOKEN)
    yield c  # type: ignore[misc]
    await c.aclose()
