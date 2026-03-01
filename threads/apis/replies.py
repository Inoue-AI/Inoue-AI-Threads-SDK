"""Reply Management API."""

from __future__ import annotations

from threads.apis.base import BaseAPI
from threads.models.media import MediaField
from threads.models.replies import ReplyListResponse


class RepliesAPI(BaseAPI):
    """Read and manage replies, conversations, and mentions.

    Requires ``threads_read_replies`` and/or ``threads_manage_replies``
    scopes depending on the operation.
    """

    async def get_replies(
        self,
        media_id: str,
        *,
        fields: list[MediaField] | None = None,
        reverse: bool = False,
    ) -> ReplyListResponse:
        """Get top-level replies to a thread.

        Parameters
        ----------
        reverse:
            If ``True``, return replies in reverse chronological order.
        """
        if fields is None:
            fields = list(MediaField)
        params: dict[str, str] = {"fields": ",".join(fields)}
        if reverse:
            params["reverse"] = "true"
        data = await self._session.get(f"{media_id}/replies", params=params)
        return ReplyListResponse.model_validate(data)

    async def get_conversation(
        self,
        media_id: str,
        *,
        fields: list[MediaField] | None = None,
        reverse: bool = False,
    ) -> ReplyListResponse:
        """Get the full conversation tree (nested replies) for a thread."""
        if fields is None:
            fields = list(MediaField)
        params: dict[str, str] = {"fields": ",".join(fields)}
        if reverse:
            params["reverse"] = "true"
        data = await self._session.get(f"{media_id}/conversation", params=params)
        return ReplyListResponse.model_validate(data)

    async def hide_reply(self, reply_id: str) -> bool:
        """Hide a reply so it is no longer publicly visible."""
        resp = await self._session.post(f"{reply_id}/manage_reply", data={"hide": "true"})
        return bool(resp.get("success", False))

    async def unhide_reply(self, reply_id: str) -> bool:
        """Unhide a previously hidden reply."""
        resp = await self._session.post(f"{reply_id}/manage_reply", data={"hide": "false"})
        return bool(resp.get("success", False))

    async def get_mentions(
        self,
        *,
        fields: list[MediaField] | None = None,
    ) -> ReplyListResponse:
        """Get threads that mention the authenticated user."""
        if fields is None:
            fields = list(MediaField)
        params = {"fields": ",".join(fields)}
        data = await self._session.get("me/mentions", params=params)
        return ReplyListResponse.model_validate(data)
