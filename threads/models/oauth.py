"""Models and enums for the Threads OAuth 2.0 flow."""

from __future__ import annotations

from enum import StrEnum

from threads.models.base import ThreadsBaseModel


class ThreadsScope(StrEnum):
    """OAuth scopes recognised by the Threads platform.

    Request the minimal set your integration needs. See the Threads API
    docs for the authoritative permission-to-feature mapping.
    """

    BASIC = "threads_basic"
    CONTENT_PUBLISH = "threads_content_publish"
    MANAGE_INSIGHTS = "threads_manage_insights"
    READ_REPLIES = "threads_read_replies"
    MANAGE_REPLIES = "threads_manage_replies"
    KEYWORD_SEARCH = "threads_keyword_search"
    MANAGE_MENTIONS = "threads_manage_mentions"
    DELETE = "threads_delete"


class ShortLivedToken(ThreadsBaseModel):
    """Result of exchanging an authorization ``code`` for a short-lived token.

    The Threads ``/oauth/access_token`` endpoint returns a short-lived
    (~1 hour) token plus the numeric Threads user id.
    """

    access_token: str
    user_id: str | int | None = None


class LongLivedToken(ThreadsBaseModel):
    """Result of exchanging a short-lived token for a long-lived (~60 day) token.

    Returned by both ``/access_token?grant_type=th_exchange_token`` and the
    ``/refresh_access_token?grant_type=th_refresh_token`` endpoints.
    """

    access_token: str
    token_type: str | None = None
    expires_in: int | None = None
