"""Insights / Analytics API."""

from __future__ import annotations

from typing import Any

from threads.apis.base import BaseAPI
from threads.models.insights import (
    AccountMetric,
    DemographicBreakdown,
    InsightData,
    InsightPeriod,
    InsightsResponse,
    MediaMetric,
)


class InsightsAPI(BaseAPI):
    """Access media-level and account-level analytics.

    Requires the ``threads_manage_insights`` scope.
    """

    async def get_media_insights(
        self,
        media_id: str,
        metrics: list[MediaMetric] | None = None,
    ) -> InsightsResponse:
        """Retrieve insights for a specific thread / media object.

        Parameters
        ----------
        metrics:
            List of metrics to fetch. Defaults to all available media metrics.
        """
        if metrics is None:
            metrics = list(MediaMetric)
        params = {"metric": ",".join(metrics)}
        data = await self._session.get(f"{media_id}/insights", params=params)
        return InsightsResponse.model_validate(data)

    async def get_account_insights(
        self,
        *,
        metrics: list[AccountMetric] | None = None,
        period: InsightPeriod = InsightPeriod.DAY,
        since: str | None = None,
        until: str | None = None,
        breakdown: DemographicBreakdown | None = None,
    ) -> InsightsResponse:
        """Retrieve account-level insights with optional date range and breakdown.

        Parameters
        ----------
        metrics:
            Which metrics to fetch. Defaults to the five core engagement
            metrics (views, likes, replies, reposts, quotes).
        period:
            Aggregation period (``day``, ``week``, ``days_28``, ``lifetime``).
        since / until:
            ISO-8601 date strings bounding the date range.
        breakdown:
            Demographic dimension (``country``, ``city``, ``age``, ``gender``).
            Requires at least 100 followers.
        """
        if metrics is None:
            metrics = [
                AccountMetric.VIEWS,
                AccountMetric.LIKES,
                AccountMetric.REPLIES,
                AccountMetric.REPOSTS,
                AccountMetric.QUOTES,
            ]

        params: dict[str, Any] = {
            "metric": ",".join(metrics),
            "period": period.value,
        }
        if since:
            params["since"] = since
        if until:
            params["until"] = until
        if breakdown:
            params["breakdown"] = breakdown.value

        data = await self._session.get("me/threads_insights", params=params)
        return InsightsResponse.model_validate(data)

    async def get_metric(
        self,
        media_id: str,
        metric: MediaMetric,
    ) -> InsightData | None:
        """Convenience: retrieve a single metric for a media object."""
        response = await self.get_media_insights(media_id, [metric])
        return response.data[0] if response.data else None
