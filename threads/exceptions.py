"""Exception hierarchy for the Threads SDK."""

from __future__ import annotations


class ThreadsSDKError(Exception):
    """Base exception for all SDK errors."""


class ThreadsConfigError(ThreadsSDKError):
    """Raised for invalid configuration (e.g. empty access token)."""


class ThreadsAPIError(ThreadsSDKError):
    """Raised when the Threads API returns an error response."""

    def __init__(
        self,
        message: str,
        *,
        http_status: int = 200,
        error_type: str = "",
        error_code: int = 0,
        error_subcode: int = 0,
        fbtrace_id: str = "",
    ) -> None:
        self.message = message
        self.http_status = http_status
        self.error_type = error_type
        self.error_code = error_code
        self.error_subcode = error_subcode
        self.fbtrace_id = fbtrace_id
        super().__init__(f"[{http_status}] {error_type}: {message}")


class ThreadsAuthError(ThreadsAPIError):
    """Raised for authentication / authorization failures."""


class ThreadsRateLimitError(ThreadsAPIError):
    """Raised when the API rate limit is exceeded."""


class ThreadsNotFoundError(ThreadsAPIError):
    """Raised when a requested resource does not exist."""


class ThreadsServerError(ThreadsAPIError):
    """Raised for 5xx server-side errors."""


class ThreadsPublishingError(ThreadsSDKError):
    """Raised when media container processing fails or times out."""


def build_api_error(
    message: str,
    *,
    http_status: int,
    error_type: str,
    error_code: int,
    error_subcode: int,
    fbtrace_id: str,
) -> ThreadsAPIError:
    """Return the most specific exception subclass for the given error."""
    cls: type[ThreadsAPIError]

    if http_status == 429 or error_code == 4:
        cls = ThreadsRateLimitError
    elif http_status == 401 or error_code in (190, 102) or error_type == "OAuthException":
        cls = ThreadsAuthError
    elif http_status == 404:
        cls = ThreadsNotFoundError
    elif http_status >= 500:
        cls = ThreadsServerError
    else:
        cls = ThreadsAPIError

    return cls(
        message,
        http_status=http_status,
        error_type=error_type,
        error_code=error_code,
        error_subcode=error_subcode,
        fbtrace_id=fbtrace_id,
    )
