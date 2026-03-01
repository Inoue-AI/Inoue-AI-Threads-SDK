"""Low-level async HTTP session for the Threads Graph API."""

from __future__ import annotations

import asyncio
from typing import Any

import aiohttp

from threads.config import DEFAULT_TIMEOUT, THREADS_API_BASE_URL
from threads.exceptions import ThreadsAPIError, build_api_error


class ThreadsSession:
    """Thin wrapper around :class:`aiohttp.ClientSession`.

    * Lazily creates the underlying session on first request.
    * Attaches the ``access_token`` query-parameter to every request.
    * Parses Graph API error envelopes and raises typed exceptions.
    """

    __slots__ = ("_access_token", "_base_url", "_timeout", "_session", "_lock")

    def __init__(
        self,
        access_token: str,
        *,
        base_url: str = THREADS_API_BASE_URL,
        timeout: float = DEFAULT_TIMEOUT,
    ) -> None:
        self._access_token = access_token
        self._base_url = base_url.rstrip("/")
        self._timeout = aiohttp.ClientTimeout(total=timeout)
        self._session: aiohttp.ClientSession | None = None
        self._lock = asyncio.Lock()

    # -- internal helpers ----------------------------------------------------

    async def _ensure_session(self) -> aiohttp.ClientSession:
        if self._session is None or self._session.closed:
            async with self._lock:
                if self._session is None or self._session.closed:
                    connector = aiohttp.TCPConnector(
                        limit=100,
                        ttl_dns_cache=300,
                        ssl=True,
                    )
                    self._session = aiohttp.ClientSession(
                        connector=connector,
                        timeout=self._timeout,
                    )
        return self._session

    def _url(self, path: str) -> str:
        if path.startswith("http"):
            return path
        return f"{self._base_url}/{path.lstrip('/')}"

    def _auth_params(self) -> dict[str, str]:
        return {"access_token": self._access_token}

    @staticmethod
    async def _parse_response(resp: aiohttp.ClientResponse) -> dict[str, Any]:
        body: dict[str, Any] = await resp.json(content_type=None)

        if "error" in body:
            err = body["error"]
            raise build_api_error(
                message=err.get("message", "Unknown error"),
                http_status=resp.status,
                error_type=err.get("type", ""),
                error_code=err.get("code", 0),
                error_subcode=err.get("error_subcode", 0),
                fbtrace_id=err.get("fbtrace_id", ""),
            )

        if resp.status >= 400:
            raise ThreadsAPIError(
                f"HTTP {resp.status}",
                http_status=resp.status,
            )

        return body

    # -- public request methods ----------------------------------------------

    async def get(
        self,
        path: str,
        params: dict[str, Any] | None = None,
    ) -> dict[str, Any]:
        """Issue a GET request and return the parsed JSON body."""
        session = await self._ensure_session()
        merged: dict[str, Any] = {**self._auth_params(), **(params or {})}
        async with session.get(self._url(path), params=merged) as resp:
            return await self._parse_response(resp)

    async def post(
        self,
        path: str,
        data: dict[str, Any] | None = None,
        params: dict[str, Any] | None = None,
    ) -> dict[str, Any]:
        """Issue a POST request with form-encoded *data* and return parsed JSON."""
        session = await self._ensure_session()
        merged: dict[str, Any] = {**self._auth_params(), **(params or {})}
        async with session.post(self._url(path), params=merged, data=data) as resp:
            return await self._parse_response(resp)

    async def delete(
        self,
        path: str,
        params: dict[str, Any] | None = None,
    ) -> dict[str, Any]:
        """Issue a DELETE request and return the parsed JSON body."""
        session = await self._ensure_session()
        merged: dict[str, Any] = {**self._auth_params(), **(params or {})}
        async with session.delete(self._url(path), params=merged) as resp:
            return await self._parse_response(resp)

    # -- lifecycle -----------------------------------------------------------

    async def close(self) -> None:
        """Close the underlying HTTP session."""
        async with self._lock:
            if self._session and not self._session.closed:
                await self._session.close()
                self._session = None
