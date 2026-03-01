"""API namespace classes."""

from threads.apis.insights import InsightsAPI
from threads.apis.media import MediaAPI
from threads.apis.publishing import PublishingAPI
from threads.apis.replies import RepliesAPI
from threads.apis.search import SearchAPI
from threads.apis.user import UserAPI

__all__ = [
    "InsightsAPI",
    "MediaAPI",
    "PublishingAPI",
    "RepliesAPI",
    "SearchAPI",
    "UserAPI",
]
