"""Base Pydantic model and common pagination types."""

from __future__ import annotations

from pydantic import BaseModel, ConfigDict


class ThreadsBaseModel(BaseModel):
    """Immutable base model with forward-compatible defaults.

    * ``frozen``          – response objects are immutable after creation.
    * ``extra="ignore"``  – silently drop unknown fields from newer API versions.
    * ``populate_by_name`` – accept both field name and alias.
    """

    model_config = ConfigDict(
        frozen=True,
        extra="ignore",
        populate_by_name=True,
    )


class PaginationCursors(ThreadsBaseModel):
    """Cursor pair returned inside ``paging``."""

    before: str | None = None
    after: str | None = None


class Pagination(ThreadsBaseModel):
    """Pagination envelope returned by list endpoints."""

    cursors: PaginationCursors | None = None
