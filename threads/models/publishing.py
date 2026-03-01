"""Models for the Content Publishing API."""

from __future__ import annotations

from enum import StrEnum

from threads.models.base import ThreadsBaseModel


class MediaType(StrEnum):
    """Media type used when creating a container."""

    TEXT = "TEXT"
    IMAGE = "IMAGE"
    VIDEO = "VIDEO"
    CAROUSEL = "CAROUSEL"


class ReplyControl(StrEnum):
    """Controls who can reply to a thread."""

    EVERYONE = "everyone"
    ACCOUNTS_YOU_FOLLOW = "accounts_you_follow"
    MENTIONED_ONLY = "mentioned_only"


class ContainerStatus(StrEnum):
    """Processing status of a media container."""

    IN_PROGRESS = "IN_PROGRESS"
    FINISHED = "FINISHED"
    ERRORED = "ERRORED"
    EXPIRED = "EXPIRED"
    PUBLISHED = "PUBLISHED"


class ContainerStatusResponse(ThreadsBaseModel):
    """Status of a media container being processed."""

    id: str
    status: ContainerStatus | None = None
    error_message: str | None = None


class PublishResponse(ThreadsBaseModel):
    """Response from creating a container or publishing a thread."""

    id: str


class PublishingLimitConfig(ThreadsBaseModel):
    """Rate-limit configuration returned by the publishing-limit endpoint."""

    quota_total: int | None = None
    quota_duration: int | None = None


class PublishingLimit(ThreadsBaseModel):
    """Current publishing quota usage and configuration."""

    quota_usage: int | None = None
    config: PublishingLimitConfig | None = None
    reply_quota_usage: int | None = None
    reply_config: PublishingLimitConfig | None = None
