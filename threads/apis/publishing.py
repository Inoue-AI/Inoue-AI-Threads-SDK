"""Content Publishing API."""

from __future__ import annotations

import asyncio
from typing import Any

from threads.apis.base import BaseAPI
from threads.config import (
    CAROUSEL_MAX_ITEMS,
    CAROUSEL_MIN_ITEMS,
    DEFAULT_POLL_INTERVAL,
    DEFAULT_POLL_TIMEOUT,
)
from threads.exceptions import ThreadsConfigError, ThreadsPublishingError
from threads.models.publishing import (
    ContainerStatus,
    ContainerStatusResponse,
    MediaType,
    PublishingLimit,
    PublishResponse,
    ReplyControl,
)

_TERMINAL_STATUSES = frozenset(
    {
        ContainerStatus.FINISHED,
        ContainerStatus.ERRORED,
        ContainerStatus.EXPIRED,
        ContainerStatus.PUBLISHED,
    }
)


class PublishingAPI(BaseAPI):
    """Create, publish, and manage Threads posts.

    Requires the ``threads_content_publish`` scope (and ``threads_delete``
    for :meth:`delete_post`).
    """

    # -- low-level container creation ----------------------------------------

    async def create_container(
        self,
        user_id: str,
        *,
        media_type: MediaType,
        text: str | None = None,
        image_url: str | None = None,
        video_url: str | None = None,
        children: list[str] | None = None,
        is_carousel_item: bool = False,
        reply_to_id: str | None = None,
        reply_control: ReplyControl | None = None,
        link_attachment: str | None = None,
    ) -> PublishResponse:
        """Create a media container.

        For most use-cases prefer the convenience helpers:
        :meth:`create_text_post`, :meth:`create_image_post`,
        :meth:`create_video_post`, or :meth:`create_carousel_post`.
        """
        data: dict[str, Any] = {"media_type": media_type.value}

        if text is not None:
            data["text"] = text
        if image_url is not None:
            data["image_url"] = image_url
        if video_url is not None:
            data["video_url"] = video_url
        if children is not None:
            data["children"] = ",".join(children)
        if is_carousel_item:
            data["is_carousel_item"] = "true"
        if reply_to_id is not None:
            data["reply_to_id"] = reply_to_id
        if reply_control is not None:
            data["reply_control"] = reply_control.value
        if link_attachment is not None:
            data["link_attachment"] = link_attachment

        resp = await self._session.post(f"{user_id}/threads", data=data)
        return PublishResponse.model_validate(resp)

    # -- convenience helpers -------------------------------------------------

    async def create_text_post(
        self,
        user_id: str,
        text: str,
        *,
        reply_to_id: str | None = None,
        reply_control: ReplyControl | None = None,
        link_attachment: str | None = None,
    ) -> PublishResponse:
        """Create a text-only container."""
        return await self.create_container(
            user_id,
            media_type=MediaType.TEXT,
            text=text,
            reply_to_id=reply_to_id,
            reply_control=reply_control,
            link_attachment=link_attachment,
        )

    async def create_image_post(
        self,
        user_id: str,
        image_url: str,
        *,
        text: str | None = None,
        reply_to_id: str | None = None,
        reply_control: ReplyControl | None = None,
        is_carousel_item: bool = False,
    ) -> PublishResponse:
        """Create an image container."""
        return await self.create_container(
            user_id,
            media_type=MediaType.IMAGE,
            image_url=image_url,
            text=text,
            reply_to_id=reply_to_id,
            reply_control=reply_control,
            is_carousel_item=is_carousel_item,
        )

    async def create_video_post(
        self,
        user_id: str,
        video_url: str,
        *,
        text: str | None = None,
        reply_to_id: str | None = None,
        reply_control: ReplyControl | None = None,
        is_carousel_item: bool = False,
    ) -> PublishResponse:
        """Create a video container."""
        return await self.create_container(
            user_id,
            media_type=MediaType.VIDEO,
            video_url=video_url,
            text=text,
            reply_to_id=reply_to_id,
            reply_control=reply_control,
            is_carousel_item=is_carousel_item,
        )

    async def create_carousel_post(
        self,
        user_id: str,
        children: list[str],
        *,
        text: str | None = None,
        reply_to_id: str | None = None,
        reply_control: ReplyControl | None = None,
    ) -> PublishResponse:
        """Create a carousel container from previously created item containers.

        Parameters
        ----------
        children:
            List of container IDs (2–20 items, each created with
            ``is_carousel_item=True``).
        """
        if len(children) < CAROUSEL_MIN_ITEMS:
            raise ThreadsConfigError(f"A carousel requires at least {CAROUSEL_MIN_ITEMS} items.")
        if len(children) > CAROUSEL_MAX_ITEMS:
            raise ThreadsConfigError(f"A carousel allows at most {CAROUSEL_MAX_ITEMS} items.")
        return await self.create_container(
            user_id,
            media_type=MediaType.CAROUSEL,
            children=children,
            text=text,
            reply_to_id=reply_to_id,
            reply_control=reply_control,
        )

    # -- publishing ----------------------------------------------------------

    async def publish(self, user_id: str, creation_id: str) -> PublishResponse:
        """Publish a previously created media container."""
        data = {"creation_id": creation_id}
        resp = await self._session.post(f"{user_id}/threads_publish", data=data)
        return PublishResponse.model_validate(resp)

    # -- container status ----------------------------------------------------

    async def get_container_status(self, container_id: str) -> ContainerStatusResponse:
        """Check the processing status of a media container."""
        params = {"fields": "id,status,error_message"}
        data = await self._session.get(container_id, params=params)
        return ContainerStatusResponse.model_validate(data)

    async def wait_for_container(
        self,
        container_id: str,
        *,
        poll_interval: float = DEFAULT_POLL_INTERVAL,
        timeout: float = DEFAULT_POLL_TIMEOUT,
    ) -> ContainerStatusResponse:
        """Poll a container until it reaches a terminal status.

        Raises :class:`~threads.exceptions.ThreadsPublishingError` if the
        container errors out or the *timeout* is exceeded.
        """
        elapsed = 0.0
        while elapsed < timeout:
            status = await self.get_container_status(container_id)
            if status.status in _TERMINAL_STATUSES:
                if status.status == ContainerStatus.ERRORED:
                    raise ThreadsPublishingError(
                        f"Container {container_id} failed: {status.error_message}"
                    )
                return status
            await asyncio.sleep(poll_interval)
            elapsed += poll_interval

        raise ThreadsPublishingError(f"Container {container_id} did not finish within {timeout}s.")

    # -- high-level post helpers ---------------------------------------------

    async def post_text(
        self,
        user_id: str,
        text: str,
        *,
        reply_to_id: str | None = None,
        reply_control: ReplyControl | None = None,
        link_attachment: str | None = None,
    ) -> PublishResponse:
        """Create and immediately publish a text-only thread."""
        container = await self.create_text_post(
            user_id,
            text,
            reply_to_id=reply_to_id,
            reply_control=reply_control,
            link_attachment=link_attachment,
        )
        return await self.publish(user_id, container.id)

    async def post_image(
        self,
        user_id: str,
        image_url: str,
        *,
        text: str | None = None,
        reply_to_id: str | None = None,
        reply_control: ReplyControl | None = None,
    ) -> PublishResponse:
        """Create, wait for processing, and publish an image thread."""
        container = await self.create_image_post(
            user_id,
            image_url,
            text=text,
            reply_to_id=reply_to_id,
            reply_control=reply_control,
        )
        await self.wait_for_container(container.id)
        return await self.publish(user_id, container.id)

    async def post_video(
        self,
        user_id: str,
        video_url: str,
        *,
        text: str | None = None,
        reply_to_id: str | None = None,
        reply_control: ReplyControl | None = None,
        poll_interval: float = DEFAULT_POLL_INTERVAL,
        timeout: float = DEFAULT_POLL_TIMEOUT,
    ) -> PublishResponse:
        """Create, wait for processing, and publish a video thread."""
        container = await self.create_video_post(
            user_id,
            video_url,
            text=text,
            reply_to_id=reply_to_id,
            reply_control=reply_control,
        )
        await self.wait_for_container(
            container.id,
            poll_interval=poll_interval,
            timeout=timeout,
        )
        return await self.publish(user_id, container.id)

    async def post_carousel(
        self,
        user_id: str,
        children: list[str],
        *,
        text: str | None = None,
        reply_to_id: str | None = None,
        reply_control: ReplyControl | None = None,
    ) -> PublishResponse:
        """Create and publish a carousel thread.

        The *children* must already be created container IDs (each with
        ``is_carousel_item=True``).  This method waits for the carousel
        container to finish processing before publishing.
        """
        container = await self.create_carousel_post(
            user_id,
            children,
            text=text,
            reply_to_id=reply_to_id,
            reply_control=reply_control,
        )
        await self.wait_for_container(container.id)
        return await self.publish(user_id, container.id)

    # -- deletion ------------------------------------------------------------

    async def delete_post(self, thread_id: str) -> bool:
        """Delete a published thread.

        Requires the ``threads_delete`` scope.
        """
        resp = await self._session.delete(thread_id)
        return bool(resp.get("success", False))

    # -- publishing limit ----------------------------------------------------

    async def get_publishing_limit(self) -> PublishingLimit:
        """Check the current publishing rate-limit quota."""
        params = {"fields": "quota_usage,config,reply_quota_usage,reply_config"}
        resp = await self._session.get("me/threads_publishing_limit", params=params)
        items = resp.get("data", [])
        if items:
            return PublishingLimit.model_validate(items[0])
        return PublishingLimit()
