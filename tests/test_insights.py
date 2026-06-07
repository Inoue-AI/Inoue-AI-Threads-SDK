"""Tests for the Insights / Analytics API."""

from __future__ import annotations

import pytest

from tests.conftest import MockRouter
from threads.client import ThreadsClient
from threads.models.insights import (
    AccountMetric,
    DemographicBreakdown,
    InsightPeriod,
    MediaMetric,
)


@pytest.mark.asyncio
async def test_get_media_insights(
    router: MockRouter,
    client: ThreadsClient,
) -> None:
    router.add(
        "/post_1/insights",
        json={
            "data": [
                {
                    "name": "views",
                    "period": "lifetime",
                    "values": [{"value": 1500}],
                    "title": "Views",
                    "id": "post_1/insights/views/lifetime",
                },
                {
                    "name": "likes",
                    "period": "lifetime",
                    "values": [{"value": 42}],
                    "title": "Likes",
                    "id": "post_1/insights/likes/lifetime",
                },
            ],
        },
    )

    response = await client.insights.get_media_insights("post_1")
    assert len(response.data) == 2
    assert response.data[0].name == "views"
    assert response.data[0].values[0].value == 1500
    assert response.data[1].name == "likes"
    assert response.data[1].values[0].value == 42


@pytest.mark.asyncio
async def test_get_account_insights(
    router: MockRouter,
    client: ThreadsClient,
) -> None:
    router.add(
        "/me/threads_insights",
        json={
            "data": [
                {
                    "name": "views",
                    "period": "day",
                    "values": [
                        {"value": 200, "end_time": "2024-06-01T07:00:00+0000"},
                        {"value": 350, "end_time": "2024-06-02T07:00:00+0000"},
                    ],
                    "title": "Views",
                    "id": "me/threads_insights/views/day",
                },
            ],
        },
    )

    response = await client.insights.get_account_insights(
        period=InsightPeriod.DAY,
        since="2024-06-01",
        until="2024-06-03",
    )
    assert len(response.data) == 1
    assert response.data[0].name == "views"
    assert len(response.data[0].values) == 2


@pytest.mark.asyncio
async def test_get_metric_convenience(
    router: MockRouter,
    client: ThreadsClient,
) -> None:
    router.add(
        "/post_2/insights",
        json={
            "data": [
                {
                    "name": "likes",
                    "period": "lifetime",
                    "values": [{"value": 99}],
                    "title": "Likes",
                    "id": "post_2/insights/likes/lifetime",
                },
            ],
        },
    )

    metric = await client.insights.get_metric("post_2", MediaMetric.LIKES)
    assert metric is not None
    assert metric.name == "likes"
    assert metric.values[0].value == 99


@pytest.mark.asyncio
async def test_get_metric_empty(
    router: MockRouter,
    client: ThreadsClient,
) -> None:
    router.add("/post_3/insights", json={"data": []})

    metric = await client.insights.get_metric("post_3", MediaMetric.SHARES)
    assert metric is None


@pytest.mark.asyncio
async def test_follower_demographics(
    router: MockRouter,
    client: ThreadsClient,
) -> None:
    router.add(
        "/me/threads_insights",
        json={
            "data": [
                {
                    "name": "follower_demographics",
                    "period": "lifetime",
                    "values": [{"value": {"US": 500, "UK": 200, "JP": 150}}],
                    "title": "Follower Demographics",
                    "id": "me/threads_insights/follower_demographics/lifetime",
                },
            ],
        },
    )

    response = await client.insights.get_account_insights(
        metrics=[AccountMetric.FOLLOWER_DEMOGRAPHICS],
        period=InsightPeriod.LIFETIME,
        breakdown=DemographicBreakdown.COUNTRY,
    )
    assert len(response.data) == 1
    val = response.data[0].values[0].value
    assert isinstance(val, dict)
    assert val["US"] == 500
