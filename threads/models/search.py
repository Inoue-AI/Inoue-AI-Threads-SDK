"""Models for the Keyword Search API."""

from __future__ import annotations

from enum import StrEnum

from threads.models.base import Pagination, ThreadsBaseModel
from threads.models.media import Thread  # noqa: TCH001 – Pydantic needs runtime access


class SearchMediaType(StrEnum):
    """Media-type filter for keyword search."""

    TEXT = "TEXT"
    IMAGE = "IMAGE"
    VIDEO = "VIDEO"


class SearchResponse(ThreadsBaseModel):
    """Paginated results from a keyword search."""

    data: list[Thread] = []
    paging: Pagination | None = None
