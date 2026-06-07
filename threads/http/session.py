"""Low-level async HTTP session for the Threads Graph API.

Built on :mod:`httpx`. A single :class:`httpx.AsyncClient` is owned per
session, lazily constructed on first use, and explicitly closed via
:meth:`ThreadsSession.close` (or the owning :class:`~threads.client.ThreadsClient`
context manager). Connection pooling and timeouts are configured explicitly —
the SDK never relies on a global/default client.
"""

from __future__ import annotations

import asyncio
from typing import Any

import httpx

from threads.config import (
    DEFAULT_MAX_CONNECTIONS,
    DEFAULT_MAX_KEEPALIVE_CONNECTIONS,
    DEFAULT_TIMEOUT,
    THREADS_API_BASE_URL,
)
from threads.exceptions import ThreadsAPIError, build_api_error


class ThreadsSession:
    """Thin async wrapper around :class:`httpx.AsyncClient`.

    * Lazily creates the underlying client on first request.
    * Attaches the ``access_token`` query-parameter to every request.
    * Parses Graph API error envelopes and raises typed exceptions.

    Parameters
    ----------
    access_token:
        A valid Threads OAuth 2.0 access token.
    base_url:
        Graph API root. Overridable for testing.
    timeout:
        Total per-request timeout in seconds.
    transport:
        Optional :class:`httpx.AsyncBaseTransport` injected for testing
        (e.g. :class:`httpx.MockTransport`). Production code leaves this
        ``None`` so a pooled HTTP transport is used.
    """

    __slots__ = (
        "_access_token",
        "_base_url",
        "_timeout",
        "_transport",
        "_client",
        "_lock",
    )

    def __init__(
        self,
        access_token: str,
        *,
        base_url: str = THREADS_API_BASE_URL,
        timeout: float = DEFAULT_TIMEOUT,
        transport: httpx.AsyncBaseTransport | None = None,
    ) -> None:
        self._access_token = access_token
        self._base_url = base_url.rstrip("/")
        self._timeout = httpx.Timeout(timeout)
        self._transport = transport
        self._client: httpx.AsyncClient | None = None
        self._lock = asyncio.Lock()

    # -- internal helpers ----------------------------------------------------

    async def _ensure_client(self) -> httpx.AsyncClient:
        if self._client is None or self._client.is_closed:
            async with self._lock:
                if self._client is None or self._client.is_closed:
                    limits = httpx.Limits(
                        max_connections=DEFAULT_MAX_CONNECTIONS,
                        max_keepalive_connections=DEFAULT_MAX_KEEPALIVE_CONNECTIONS,
                    )
                    self._client = httpx.AsyncClient(
                        base_url=self._base_url,
                        timeout=self._timeout,
                        limits=limits,
                        transport=self._transport,
                    )
        return self._client

    def _url(self, path: str) -> str:
        # Absolute URLs (e.g. pagination "next" links) pass through unchanged;
        # otherwise return a leading-slash-normalised relative path so httpx
        # composes it against the client's configured base_url. Leading slashes
        # on a relative path would make httpx discard the base_url's own path
        # component (the API version prefix), so they are stripped.
        if path.startswith("http"):
            return path
        return path.lstrip("/")

    def _auth_params(self) -> dict[str, str]:
        return {"access_token": self._access_token}

    @staticmethod
    def _parse_response(resp: httpx.Response) -> dict[str, Any]:
        try:
            body: dict[str, Any] = resp.json()
        except ValueError as exc:
            if resp.status_code >= 400:
                raise ThreadsAPIError(
                    f"HTTP {resp.status_code}",
                    http_status=resp.status_code,
                ) from exc
            raise ThreadsAPIError(
                "Threads returned a non-JSON response",
                http_status=resp.status_code,
            ) from exc

        if isinstance(body, dict) and "error" in body:
            err = body["error"]
            raise build_api_error(
                message=err.get("message", "Unknown error"),
                http_status=resp.status_code,
                error_type=err.get("type", ""),
                error_code=err.get("code", 0),
                error_subcode=err.get("error_subcode", 0),
                fbtrace_id=err.get("fbtrace_id", ""),
            )

        if resp.status_code >= 400:
            raise ThreadsAPIError(
                f"HTTP {resp.status_code}",
                http_status=resp.status_code,
            )

        return body

    # -- public request methods ----------------------------------------------

    async def get(
        self,
        path: str,
        params: dict[str, Any] | None = None,
    ) -> dict[str, Any]:
        """Issue a GET request and return the parsed JSON body."""
        client = await self._ensure_client()
        merged: dict[str, Any] = {**self._auth_params(), **(params or {})}
        resp = await client.get(self._url(path), params=merged)
        return self._parse_response(resp)

    async def post(
        self,
        path: str,
        data: dict[str, Any] | None = None,
        params: dict[str, Any] | None = None,
    ) -> dict[str, Any]:
        """Issue a POST request with form-encoded *data* and return parsed JSON."""
        client = await self._ensure_client()
        merged: dict[str, Any] = {**self._auth_params(), **(params or {})}
        resp = await client.post(self._url(path), params=merged, data=data)
        return self._parse_response(resp)

    async def delete(
        self,
        path: str,
        params: dict[str, Any] | None = None,
    ) -> dict[str, Any]:
        """Issue a DELETE request and return the parsed JSON body."""
        client = await self._ensure_client()
        merged: dict[str, Any] = {**self._auth_params(), **(params or {})}
        resp = await client.delete(self._url(path), params=merged)
        return self._parse_response(resp)

    # -- lifecycle -----------------------------------------------------------

    async def close(self) -> None:
        """Close the underlying HTTP client and release pooled connections."""
        async with self._lock:
            if self._client is not None and not self._client.is_closed:
                await self._client.aclose()
            self._client = None
