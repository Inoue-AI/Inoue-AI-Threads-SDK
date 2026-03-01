"""User Profile API."""

from __future__ import annotations

from threads.apis.base import BaseAPI
from threads.models.user import User, UserField


class UserAPI(BaseAPI):
    """Retrieve Threads user profile information.

    Requires the ``threads_basic`` scope.
    """

    async def get_profile(
        self,
        user_id: str = "me",
        fields: list[UserField] | None = None,
    ) -> User:
        """Fetch the profile for the authenticated user (default) or a specific user.

        Parameters
        ----------
        user_id:
            User ID or ``"me"`` for the authenticated user.
        fields:
            Profile fields to retrieve. Defaults to **all** available fields.
        """
        if fields is None:
            fields = list(UserField)
        params = {"fields": ",".join(fields)}
        data = await self._session.get(user_id, params=params)
        return User.model_validate(data)
