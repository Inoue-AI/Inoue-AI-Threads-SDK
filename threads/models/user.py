"""Models for the User Profile API."""

from __future__ import annotations

from enum import StrEnum

from threads.models.base import ThreadsBaseModel


class UserField(StrEnum):
    """Fields available when retrieving user profile information."""

    ID = "id"
    USERNAME = "username"
    NAME = "name"
    THREADS_PROFILE_PICTURE_URL = "threads_profile_picture_url"
    THREADS_BIOGRAPHY = "threads_biography"
    IS_VERIFIED = "is_verified"


class User(ThreadsBaseModel):
    """Threads user profile.

    All fields are optional because the caller selects which fields to request.
    """

    id: str | None = None
    username: str | None = None
    name: str | None = None
    threads_profile_picture_url: str | None = None
    threads_biography: str | None = None
    is_verified: bool | None = None
