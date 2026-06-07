"""Shared fixtures for the Threads SDK test suite.

All vendor HTTP is mocked at the transport layer via :class:`httpx.MockTransport`
(first-party in httpx — no third-party mock library). No real Threads API call
is ever made, and no credentials are required to run the suite.
"""

from __future__ import annotations

from collections.abc import Callable
from typing import Any

import httpx
import pytest

from threads.client import ThreadsClient
from threads.config import THREADS_API_BASE_URL

TOKEN = "test_access_token_123"
BASE = THREADS_API_BASE_URL

# A handler maps an inbound httpx.Request to an httpx.Response.
RouteHandler = Callable[[httpx.Request], httpx.Response]


class MockRouter:
    """Minimal path-keyed router backing an :class:`httpx.MockTransport`.

    Routes are registered by exact URL path (e.g. ``/me`` or ``/me/threads``).
    Each registered path may queue multiple responses; they are returned in
    FIFO order across successive calls (so polling flows can be modelled).
    """

    def __init__(self) -> None:
        self._routes: dict[str, list[httpx.Response]] = {}
        self.requests: list[httpx.Request] = []

    def add(
        self,
        path: str,
        *,
        json: dict[str, Any] | list[Any] | None = None,
        status_code: int = 200,
    ) -> None:
        """Queue a JSON response for the given URL *path*."""
        self._routes.setdefault(path, []).append(httpx.Response(status_code=status_code, json=json))

    def _match(self, path: str) -> list[httpx.Response] | None:
        # The session targets the versioned base URL, so the actual request
        # path is e.g. "/v1.0/me". Routes are registered by their unversioned
        # suffix ("/me"); match on suffix so tests stay version-agnostic.
        if path in self._routes:
            return self._routes[path]
        for route_path, queue in self._routes.items():
            if path.endswith(route_path):
                return queue
        return None

    def handler(self, request: httpx.Request) -> httpx.Response:
        self.requests.append(request)
        queue = self._match(request.url.path)
        if not queue:
            return httpx.Response(
                status_code=404,
                json={"error": {"message": f"no mock for {request.url.path}", "code": 100}},
            )
        # Keep the last queued response once exhausted (idempotent polling).
        return queue.pop(0) if len(queue) > 1 else queue[0]

    @property
    def transport(self) -> httpx.MockTransport:
        return httpx.MockTransport(self.handler)


@pytest.fixture
def router() -> MockRouter:
    """A fresh :class:`MockRouter` per test."""
    return MockRouter()


@pytest.fixture
async def client(router: MockRouter) -> ThreadsClient:  # type: ignore[misc]
    """A ThreadsClient wired to the mock transport."""
    c = ThreadsClient(access_token=TOKEN, transport=router.transport)
    yield c  # type: ignore[misc]
    await c.aclose()
