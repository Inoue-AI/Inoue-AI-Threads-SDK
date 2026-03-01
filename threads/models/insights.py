"""Models for the Insights / Analytics API."""

from __future__ import annotations

from enum import StrEnum
from typing import Any

from threads.models.base import ThreadsBaseModel


class MediaMetric(StrEnum):
    """Metrics available for individual media posts."""

    VIEWS = "views"
    LIKES = "likes"
    REPLIES = "replies"
    REPOSTS = "reposts"
    QUOTES = "quotes"
    SHARES = "shares"


class AccountMetric(StrEnum):
    """Metrics available for account-level insights."""

    VIEWS = "views"
    LIKES = "likes"
    REPLIES = "replies"
    REPOSTS = "reposts"
    QUOTES = "quotes"
    SHARES = "shares"
    FOLLOWERS_COUNT = "followers_count"
    FOLLOWER_DEMOGRAPHICS = "follower_demographics"


class InsightPeriod(StrEnum):
    """Time period for account-level insights."""

    DAY = "day"
    WEEK = "week"
    DAYS_28 = "days_28"
    LIFETIME = "lifetime"


class DemographicBreakdown(StrEnum):
    """Dimensions for follower demographic breakdowns."""

    COUNTRY = "country"
    CITY = "city"
    AGE = "age"
    GENDER = "gender"


class InsightValue(ThreadsBaseModel):
    """A single insight data-point."""

    value: int | dict[str, Any] | None = None
    end_time: str | None = None


class InsightData(ThreadsBaseModel):
    """A single insight metric with its values."""

    name: str
    period: str
    values: list[InsightValue] = []
    title: str | None = None
    description: str | None = None
    id: str | None = None
    total_value: InsightValue | None = None


class InsightsResponse(ThreadsBaseModel):
    """Response from an insights endpoint."""

    data: list[InsightData] = []
