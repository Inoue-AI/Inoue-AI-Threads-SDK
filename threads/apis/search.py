"""Keyword Search API."""

from __future__ import annotations

from typing import Any

from threads.apis.base import BaseAPI
from threads.models.media import MediaField
from threads.models.search import SearchMediaType, SearchResponse


class SearchAPI(BaseAPI):
    """Search public Threads posts by keyword or hashtag.

    Requires the ``threads_keyword_search`` scope.
    """

    async def search(
        self,
        query: str,
        *,
        media_type: SearchMediaType | None = None,
        fields: list[MediaField] | None = None,
        since: str | None = None,
        until: str | None = None,
    ) -> SearchResponse:
        """Search for public threads matching a keyword or ``#topic``.

        Parameters
        ----------
        query:
            The search term.  Prefix with ``#`` to search by topic tag.
        media_type:
            Optionally filter results by media type.
        fields:
            Thread fields to include in results.
        since / until:
            Unix timestamps to bound the date range.
        """
        if fields is None:
            fields = list(MediaField)

        params: dict[str, Any] = {
            "q": query,
            "fields": ",".join(fields),
        }
        if media_type is not None:
            params["media_type"] = media_type.value
        if since:
            params["since"] = since
        if until:
            params["until"] = until

        data = await self._session.get("keyword_search", params=params)
        return SearchResponse.model_validate(data)
