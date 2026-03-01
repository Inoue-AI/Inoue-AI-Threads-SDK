"""Base class shared by all API namespaces."""

from __future__ import annotations

from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from threads.http.session import ThreadsSession


class BaseAPI:
    """Provides the shared :class:`ThreadsSession` to every API namespace."""

    __slots__ = ("_session",)

    def __init__(self, session: ThreadsSession) -> None:
        self._session = session
