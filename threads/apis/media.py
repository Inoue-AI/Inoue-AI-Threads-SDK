"""Media / Threads Retrieval API."""

from __future__ import annotations

from typing import TYPE_CHECKING, Any

from threads.apis.base import BaseAPI

if TYPE_CHECKING:
    from collections.abc import AsyncIterator
from threads.config import DEFAULT_PAGE_LIMIT
from threads.models.media import MediaField, Thread, ThreadListResponse


class MediaAPI(BaseAPI):
    """Retrieve user threads and individual media objects.

    Requires the ``threads_basic`` scope.
    """

    async def list_threads(
        self,
        user_id: str = "me",
        *,
        fields: list[MediaField] | None = None,
        limit: int = DEFAULT_PAGE_LIMIT,
        since: str | None = None,
        until: str | None = None,
        after: str | None = None,
        before: str | None = None,
    ) -> ThreadListResponse:
        """List threads for a user with cursor-based pagination.

        Parameters
        ----------
        limit:
            Number of threads per page (max 100, default 25).
        since / until:
            ISO-8601 timestamps to bound the date range.
        after / before:
            Pagination cursors from a previous response.
        """
        if fields is None:
            fields = list(MediaField)

        params: dict[str, Any] = {
            "fields": ",".join(fields),
            "limit": limit,
        }
        if since:
            params["since"] = since
        if until:
            params["until"] = until
        if after:
            params["after"] = after
        if before:
            params["before"] = before

        data = await self._session.get(f"{user_id}/threads", params=params)
        return ThreadListResponse.model_validate(data)

    async def get_thread(
        self,
        media_id: str,
        *,
        fields: list[MediaField] | None = None,
    ) -> Thread:
        """Retrieve a single thread / media object by ID."""
        if fields is None:
            fields = list(MediaField)
        params = {"fields": ",".join(fields)}
        data = await self._session.get(media_id, params=params)
        return Thread.model_validate(data)

    async def iter_threads(
        self,
        user_id: str = "me",
        *,
        fields: list[MediaField] | None = None,
        limit: int = DEFAULT_PAGE_LIMIT,
    ) -> AsyncIterator[Thread]:
        """Auto-paginating async iterator over all threads for a user.

        Usage::

            async for thread in client.media.iter_threads():
                print(thread.text)
        """
        after: str | None = None
        while True:
            page = await self.list_threads(
                user_id,
                fields=fields,
                limit=limit,
                after=after,
            )
            for thread in page.data:
                yield thread
            if not page.paging or not page.paging.cursors or not page.paging.cursors.after:
                break
            after = page.paging.cursors.after
