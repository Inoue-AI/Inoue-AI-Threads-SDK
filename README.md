# Inoue-AI Threads SDK

Multi-language SDK for the Meta Threads Graph API. Ships both a Python client
(under `threads/`) and a Go client (under `go/`).

> **Note:** This repository was renamed from `Inoue-AI/Threads-Python-SDK` to
> `Inoue-AI/Inoue-AI-Threads-SDK`. GitHub auto-redirects the old URL, so
> existing `pip install ... @ git+https://github.com/Inoue-AI/Threads-Python-SDK.git`
> commands continue to work. The local clone keeps the legacy directory name.

## Python

The Python source remains under [`threads/`](./threads). Install:

```bash
pip install "git+https://github.com/Inoue-AI/Inoue-AI-Threads-SDK.git"
```

## Go

The Go SDK lives under [`go/`](./go). It exposes a focused subset matching
what the Inoue AI Go backend calls (user profile, posts list/get, OAuth token
refresh).

```bash
go get github.com/Inoue-AI/Inoue-AI-Threads-SDK/go@latest
```

```go
client := threads.New(threads.ClientOptions{AccessToken: "..."})
defer client.Close()
user, err := client.GetUser(ctx, "me", nil)
```

See [`go/README.md`](./go/README.md) for the full Go API.

---

## Original Python SDK

An async Python SDK for the **Meta Threads API** — covering publishing, media retrieval, insights, reply management, and keyword search.

[![Python](https://img.shields.io/pypi/pyversions/threads-python-sdk)](https://pypi.org/project/threads-python-sdk/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

## Features

- **Fully async** — built on `httpx.AsyncClient` with an `asyncio`-native interface
- **Typed responses** — every API response is a frozen Pydantic v2 model
- **Built-in OAuth 2.0** — `ThreadsOAuth` covers the authorization-code flow: consent URL, code exchange, short→long-lived exchange, and long-lived refresh
- **Six API namespaces** — `client.user`, `client.publishing`, `client.media`, `client.insights`, `client.replies`, `client.search`
- **Convenience helpers** — one-call publishing for text/image/video/carousel, container polling, auto-pagination
- **Clean exception hierarchy** — catch broad or specific errors as needed
- **Forward-compatible** — unknown fields from newer API versions are silently ignored

## Requirements

- Python ≥ 3.11
- A Meta developer account with a Threads app
- A valid user access token (OAuth 2.0 handled externally)

## Installation

```bash
pip install threads-python-sdk
```

## Quick start

```python
import asyncio
from threads import ThreadsClient, UserField, MediaField, MediaMetric

async def main() -> None:
    async with ThreadsClient(access_token="your_token_here") as client:

        # --- User Profile ---
        me = await client.user.get_profile()
        print(f"@{me.username} — {me.threads_biography}")

        # --- Publish a text post ---
        post = await client.publishing.post_text("me", "Hello Threads!")
        print(f"Published: {post.id}")

        # --- Publish an image ---
        img = await client.publishing.post_image(
            "me",
            "https://example.com/photo.jpg",
            text="Check this out",
        )

        # --- Publish a video (waits for processing automatically) ---
        vid = await client.publishing.post_video(
            "me",
            "https://example.com/clip.mp4",
            text="New video drop",
        )

        # --- Insights ---
        insights = await client.insights.get_media_insights(post.id)
        for metric in insights.data:
            print(f"{metric.name}: {metric.values[0].value}")

        # --- Paginate through all threads ---
        async for thread in client.media.iter_threads():
            print(thread.id, thread.text)

        # --- Search ---
        results = await client.search.search("#python")
        for hit in results.data:
            print(hit.text)

asyncio.run(main())
```

## API reference

### `ThreadsClient`

```python
ThreadsClient(access_token: str, *, timeout: float = 30.0)
```

The main client. Use as an async context manager or call `await client.aclose()` manually.

---

### `client.user` — User Profile API

| Method | Description |
|--------|-------------|
| `get_profile(user_id, fields)` | Authenticated user's profile (or another user by ID) |

#### Available user fields (`UserField`)

`ID`, `USERNAME`, `NAME`, `THREADS_PROFILE_PICTURE_URL`, `THREADS_BIOGRAPHY`, `IS_VERIFIED`

---

### `client.publishing` — Content Publishing API

| Method | Description |
|--------|-------------|
| `create_container(user_id, ...)` | Low-level: create a media container |
| `create_text_post(user_id, text, ...)` | Create a text-only container |
| `create_image_post(user_id, image_url, ...)` | Create an image container |
| `create_video_post(user_id, video_url, ...)` | Create a video container |
| `create_carousel_post(user_id, children, ...)` | Create a carousel container (2–20 items) |
| `publish(user_id, creation_id)` | Publish a container |
| `get_container_status(container_id)` | Check processing status |
| `wait_for_container(container_id, ...)` | Poll until terminal state |
| `post_text(user_id, text, ...)` | **Convenience:** create + publish text |
| `post_image(user_id, image_url, ...)` | **Convenience:** create + wait + publish image |
| `post_video(user_id, video_url, ...)` | **Convenience:** create + wait + publish video |
| `post_carousel(user_id, children, ...)` | **Convenience:** create + wait + publish carousel |
| `delete_post(thread_id)` | Delete a published thread |
| `get_publishing_limit()` | Check rate-limit quota |

#### Media types (`MediaType`)

`TEXT`, `IMAGE`, `VIDEO`, `CAROUSEL`

#### Reply control (`ReplyControl`)

`EVERYONE`, `ACCOUNTS_YOU_FOLLOW`, `MENTIONED_ONLY`

#### Container status (`ContainerStatus`)

`IN_PROGRESS`, `FINISHED`, `ERRORED`, `EXPIRED`, `PUBLISHED`

#### Publishing a carousel

```python
# Step 1: Create individual item containers
item1 = await client.publishing.create_image_post(
    "me", "https://example.com/img1.jpg", is_carousel_item=True,
)
item2 = await client.publishing.create_image_post(
    "me", "https://example.com/img2.jpg", is_carousel_item=True,
)

# Step 2: Create and publish carousel (convenience method)
post = await client.publishing.post_carousel(
    "me", [item1.id, item2.id], text="My carousel",
)
```

---

### `client.media` — Media Retrieval API

| Method | Description |
|--------|-------------|
| `list_threads(user_id, *, fields, limit, since, until, after, before)` | One page of the user's threads |
| `get_thread(media_id, *, fields)` | Fetch a single thread by ID |
| `iter_threads(user_id, *, fields, limit)` | Async iterator over all threads (auto-paginates) |

#### Available media fields (`MediaField`)

`ID`, `MEDIA_PRODUCT_TYPE`, `MEDIA_TYPE`, `MEDIA_URL`, `PERMALINK`, `OWNER`,
`USERNAME`, `TEXT`, `TIMESTAMP`, `SHORTCODE`, `THUMBNAIL_URL`, `CHILDREN`, `IS_QUOTE_POST`

#### Thread media types (`ThreadMediaType`)

`TEXT_POST`, `IMAGE`, `VIDEO`, `CAROUSEL_ALBUM`, `REPOST_FACADE`

---

### `client.insights` — Insights / Analytics API

| Method | Description |
|--------|-------------|
| `get_media_insights(media_id, metrics)` | Metrics for a specific post |
| `get_account_insights(*, metrics, period, since, until, breakdown)` | Account-level analytics |
| `get_metric(media_id, metric)` | Convenience: single metric for a post |

#### Media metrics (`MediaMetric`)

`VIEWS`, `LIKES`, `REPLIES`, `REPOSTS`, `QUOTES`, `SHARES`

#### Account metrics (`AccountMetric`)

All media metrics plus: `FOLLOWERS_COUNT`, `FOLLOWER_DEMOGRAPHICS`

#### Insight periods (`InsightPeriod`)

`DAY`, `WEEK`, `DAYS_28`, `LIFETIME`

#### Demographic breakdowns (`DemographicBreakdown`)

`COUNTRY`, `CITY`, `AGE`, `GENDER`

```python
# Get follower demographics by country (requires 100+ followers)
from threads import AccountMetric, InsightPeriod, DemographicBreakdown

response = await client.insights.get_account_insights(
    metrics=[AccountMetric.FOLLOWER_DEMOGRAPHICS],
    period=InsightPeriod.LIFETIME,
    breakdown=DemographicBreakdown.COUNTRY,
)
```

---

### `client.replies` — Reply Management API

| Method | Description |
|--------|-------------|
| `get_replies(media_id, *, fields, reverse)` | Top-level replies to a thread |
| `get_conversation(media_id, *, fields, reverse)` | Full conversation tree |
| `hide_reply(reply_id)` | Hide a reply |
| `unhide_reply(reply_id)` | Unhide a reply |
| `get_mentions(*, fields)` | Threads that mention the authenticated user |

---

### `client.search` — Keyword Search API

| Method | Description |
|--------|-------------|
| `search(query, *, media_type, fields, since, until)` | Search public threads by keyword or `#topic` |

#### Search media type filter (`SearchMediaType`)

`TEXT`, `IMAGE`, `VIDEO`

---

### Exception hierarchy

```
ThreadsSDKError
├── ThreadsAPIError          ← error response from the Graph API
│   ├── ThreadsAuthError     ← invalid / expired token, OAuthException
│   ├── ThreadsRateLimitError
│   ├── ThreadsNotFoundError
│   └── ThreadsServerError   ← 5xx from Meta
├── ThreadsPublishingError   ← container processing failure or timeout
└── ThreadsConfigError       ← SDK misconfiguration (e.g. empty token)
```

```python
from threads import ThreadsAuthError, ThreadsRateLimitError, ThreadsAPIError

try:
    me = await client.user.get_profile()
except ThreadsAuthError as e:
    print(f"Auth failed: {e.error_code} — {e.message}")
except ThreadsRateLimitError:
    print("Rate limited, back off and retry.")
except ThreadsAPIError as e:
    print(f"API error [{e.http_status}] {e.error_type}: {e.message}")
```

## Development

```bash
# Clone and install with dev extras
git clone https://github.com/inoue-ai/threads-python-sdk.git
cd threads-python-sdk
python3 -m venv .venv && source .venv/bin/activate
pip install -e ".[dev]"

# Run tests
pytest

# Lint and format
ruff check .
ruff format .

# Type-check
mypy threads/
```

## Authentication

The SDK ships a `ThreadsOAuth` client that implements the full Threads
OAuth 2.0 authorization-code flow. It uses your app's `client_id` /
`client_secret` (sourced from your environment or secret store — **never**
hard-coded) and mints/refreshes user access tokens:

```python
from threads import ThreadsOAuth, ThreadsScope, ThreadsClient

async with ThreadsOAuth(
    client_id="...",        # Meta App ID
    client_secret="...",    # from env / secret store
    redirect_uri="https://app.example.com/callback",
) as oauth:
    # 1. Send the user to consent.
    url = oauth.authorization_url(
        [ThreadsScope.BASIC, ThreadsScope.CONTENT_PUBLISH],
        state="csrf-token",
    )

    # 2. After the redirect back with ?code=..., exchange it.
    short = await oauth.exchange_code(code)            # ~1 hour token
    long = await oauth.exchange_for_long_lived(short.access_token)  # ~60 days

    # 3. Later, refresh the long-lived token before it expires.
    refreshed = await oauth.refresh_long_lived(long.access_token)

# Use the resulting token with the data-plane client.
async with ThreadsClient(access_token=long.access_token) as client:
    me = await client.user.get_profile()
```

You may also bring your own externally-obtained token and skip `ThreadsOAuth`
entirely — just pass it to `ThreadsClient(access_token=...)`.

Required OAuth scopes per feature:

| Feature | Scope |
|---------|-------|
| User profile | `threads_basic` |
| Publish posts | `threads_content_publish` |
| Read media | `threads_basic` |
| Insights / analytics | `threads_manage_insights` |
| Read replies | `threads_read_replies` |
| Manage replies (hide/unhide) | `threads_manage_replies` |
| Keyword search | `threads_keyword_search` |
| Delete posts | `threads_delete` |

## License

[MIT](LICENSE)
