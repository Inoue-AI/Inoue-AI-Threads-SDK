"""Main ThreadsClient class — entry point for the SDK."""

from __future__ import annotations

from typing import TYPE_CHECKING

from threads.apis.insights import InsightsAPI
from threads.apis.media import MediaAPI
from threads.apis.publishing import PublishingAPI
from threads.apis.replies import RepliesAPI
from threads.apis.search import SearchAPI
from threads.apis.user import UserAPI
from threads.config import DEFAULT_TIMEOUT
from threads.exceptions import ThreadsConfigError
from threads.http.session import ThreadsSession

if TYPE_CHECKING:
    import httpx


class ThreadsClient:
    """Async client for the Meta Threads API.

    The caller is responsible for obtaining an OAuth 2.0 access token —
    the SDK does **not** handle any part of the OAuth flow.

    Usage::

        async with ThreadsClient(access_token="...") as client:
            profile = await client.user.get_profile()
            print(profile.username)

    Parameters
    ----------
    access_token:
        A valid Threads OAuth 2.0 access token.
    timeout:
        Default request timeout in seconds.
    transport:
        Optional :class:`httpx.AsyncBaseTransport` injected for testing
        (e.g. :class:`httpx.MockTransport`). Leave ``None`` in production.
    """

    __slots__ = (
        "_session",
        "user",
        "publishing",
        "media",
        "insights",
        "replies",
        "search",
    )

    def __init__(
        self,
        access_token: str,
        *,
        timeout: float = DEFAULT_TIMEOUT,
        transport: httpx.AsyncBaseTransport | None = None,
    ) -> None:
        if not access_token:
            raise ThreadsConfigError("access_token must not be empty.")

        self._session = ThreadsSession(
            access_token=access_token,
            timeout=timeout,
            transport=transport,
        )
        self.user = UserAPI(self._session)
        self.publishing = PublishingAPI(self._session)
        self.media = MediaAPI(self._session)
        self.insights = InsightsAPI(self._session)
        self.replies = RepliesAPI(self._session)
        self.search = SearchAPI(self._session)

    async def __aenter__(self) -> ThreadsClient:
        return self

    async def __aexit__(self, *exc: object) -> None:
        await self.aclose()

    async def aclose(self) -> None:
        """Close the underlying HTTP session."""
        await self._session.close()
