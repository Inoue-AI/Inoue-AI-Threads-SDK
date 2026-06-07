"""Constants and default configuration values for the Threads SDK."""

# Graph API data plane (read/publish/insights). Versioned host.
THREADS_API_BASE_URL = "https://graph.threads.net/v1.0"

# OAuth host. The authorize UI lives on threads.net; the token-exchange and
# token-refresh endpoints live on the graph host. They are kept separate so
# tests can stub each independently and future host migrations stay isolated.
THREADS_AUTHORIZE_BASE_URL = "https://threads.net"
THREADS_OAUTH_BASE_URL = "https://graph.threads.net"

DEFAULT_TIMEOUT: float = 30.0

# httpx connection-pool bounds (explicit; the SDK never relies on defaults).
DEFAULT_MAX_CONNECTIONS = 100
DEFAULT_MAX_KEEPALIVE_CONNECTIONS = 20

DEFAULT_POLL_INTERVAL: float = 5.0
DEFAULT_POLL_TIMEOUT: float = 300.0

CAROUSEL_MIN_ITEMS = 2
CAROUSEL_MAX_ITEMS = 20

TEXT_MAX_LENGTH = 500

DEFAULT_PAGE_LIMIT = 25
MAX_PAGE_LIMIT = 100
