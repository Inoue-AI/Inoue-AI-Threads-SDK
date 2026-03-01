"""Models for the Reply Management API."""

from __future__ import annotations

from threads.models.base import Pagination, ThreadsBaseModel
from threads.models.media import Thread  # noqa: TCH001 – Pydantic needs runtime access


class ReplyListResponse(ThreadsBaseModel):
    """Paginated list of replies (or mentions)."""

    data: list[Thread] = []
    paging: Pagination | None = None


class ManageReplyResponse(ThreadsBaseModel):
    """Response from hiding or unhiding a reply."""

    success: bool = False
