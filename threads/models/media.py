"""Models for the Media / Threads retrieval API."""

from __future__ import annotations

from enum import StrEnum

from threads.models.base import Pagination, ThreadsBaseModel


class ThreadMediaType(StrEnum):
    """Media type as returned by the API for existing threads."""

    TEXT_POST = "TEXT_POST"
    IMAGE = "IMAGE"
    VIDEO = "VIDEO"
    CAROUSEL_ALBUM = "CAROUSEL_ALBUM"
    REPOST_FACADE = "REPOST_FACADE"


class MediaField(StrEnum):
    """Fields available when retrieving a thread / media object."""

    ID = "id"
    MEDIA_PRODUCT_TYPE = "media_product_type"
    MEDIA_TYPE = "media_type"
    MEDIA_URL = "media_url"
    PERMALINK = "permalink"
    OWNER = "owner"
    USERNAME = "username"
    TEXT = "text"
    TIMESTAMP = "timestamp"
    SHORTCODE = "shortcode"
    THUMBNAIL_URL = "thumbnail_url"
    CHILDREN = "children"
    IS_QUOTE_POST = "is_quote_post"


class ThreadOwner(ThreadsBaseModel):
    """Owner reference embedded in a thread object."""

    id: str


class ThreadChild(ThreadsBaseModel):
    """Child media reference in a carousel."""

    id: str


class ThreadChildren(ThreadsBaseModel):
    """Container for carousel children (Graph API wraps them in ``{data: [...]}``\\ )."""

    data: list[ThreadChild] = []


class Thread(ThreadsBaseModel):
    """A single Threads post / media object.

    All fields are optional because the caller selects which fields to request.
    """

    id: str | None = None
    media_product_type: str | None = None
    media_type: ThreadMediaType | None = None
    media_url: str | None = None
    permalink: str | None = None
    owner: ThreadOwner | None = None
    username: str | None = None
    text: str | None = None
    timestamp: str | None = None
    shortcode: str | None = None
    thumbnail_url: str | None = None
    children: ThreadChildren | None = None
    is_quote_post: bool | None = None


class ThreadListResponse(ThreadsBaseModel):
    """Paginated list of threads."""

    data: list[Thread] = []
    paging: Pagination | None = None
