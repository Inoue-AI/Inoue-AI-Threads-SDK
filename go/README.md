# Threads Go SDK

Typed, context-aware Go client for the Meta Threads Graph API.

## Install

```bash
go get github.com/Inoue-AI/Inoue-AI-Threads-SDK/go@latest
```

## Quickstart

```go
package main

import (
	"context"
	"log"
	"time"

	threads "github.com/Inoue-AI/Inoue-AI-Threads-SDK/go"
)

func main() {
	client := threads.New(threads.ClientOptions{
		AccessToken: "USER_ACCESS_TOKEN",
		Timeout:     30 * time.Second,
	})
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	user, err := client.GetUser(ctx, "me", nil)
	if err != nil {
		log.Fatalf("get user: %v", err)
	}
	log.Printf("@%s (%s)", user.Username, user.Name)
}
```

## Methods

### Read

| Go method | Threads endpoint |
|---|---|
| `GetUser(ctx, userID, fields)` | `GET /{user-id}` |
| `ListPosts(ctx, params)` | `GET /{user-id}/threads` |
| `IterPosts(ctx, params)` | `GET /{user-id}/threads` (auto-follows cursors) |
| `GetPost(ctx, mediaID, fields)` | `GET /{media-id}` |

### Publish

| Go method | Threads endpoint |
|---|---|
| `CreateContainer(ctx, params)` | `POST /{user-id}/threads` |
| `CreateTextPost(ctx, params)` | `POST /{user-id}/threads` (TEXT) |
| `CreateImagePost(ctx, params)` | `POST /{user-id}/threads` (IMAGE) |
| `CreateVideoPost(ctx, params)` | `POST /{user-id}/threads` (VIDEO) |
| `CreateCarouselPost(ctx, params)` | `POST /{user-id}/threads` (CAROUSEL) |
| `Publish(ctx, userID, creationID)` | `POST /{user-id}/threads_publish` |
| `GetContainerStatus(ctx, containerID)` | `GET /{container-id}` (status) |
| `WaitForContainer(ctx, containerID, params)` | poll until terminal status |
| `PostText(ctx, userID, text)` | create + publish (convenience) |
| `PostImage(ctx, params)` | create + wait + publish (convenience) |
| `PostVideo(ctx, params)` | create + wait + publish (convenience) |
| `PostCarousel(ctx, params)` | create + wait + publish (convenience) |
| `DeletePost(ctx, threadID)` | `DELETE /{media-id}` |
| `GetPublishingLimit(ctx)` | `GET /me/threads_publishing_limit` |

### Insights

| Go method | Threads endpoint |
|---|---|
| `GetMediaInsights(ctx, mediaID, metrics)` | `GET /{media-id}/insights` |
| `GetAccountInsights(ctx, params)` | `GET /me/threads_insights` |
| `GetMetric(ctx, mediaID, metric)` | single-metric convenience |

### Replies

| Go method | Threads endpoint |
|---|---|
| `GetReplies(ctx, mediaID, fields, reverse)` | `GET /{media-id}/replies` |
| `GetConversation(ctx, mediaID, fields, reverse)` | `GET /{media-id}/conversation` |
| `HideReply(ctx, replyID)` | `POST /{reply-id}/manage_reply` (`hide=true`) |
| `UnhideReply(ctx, replyID)` | `POST /{reply-id}/manage_reply` (`hide=false`) |
| `GetMentions(ctx, fields)` | `GET /me/mentions` |

### Search

| Go method | Threads endpoint |
|---|---|
| `Search(ctx, params)` | `GET /keyword_search` |

### OAuth 2.0

| Go method | Threads endpoint |
|---|---|
| `AuthorizationURL(params)` | builds the consent URL (no I/O) |
| `ExchangeCode(ctx, params)` | `POST /oauth/access_token` |
| `ExchangeForLongLived(ctx, params)` | `GET /access_token` (`th_exchange_token`) |
| `RefreshAccessToken(ctx, params)` | `GET /refresh_access_token` (`th_refresh_token`) |

The OAuth token endpoints live on the **unversioned** Graph host
(`https://graph.threads.net/...`, no `/v1.0` prefix); the client derives that
host automatically from `BaseURL`.

## Relationship to the backend's in-tree client

The Inoue AI Go backend has its own minimal Threads REST client at
`internal/domain/platforms/threads/client.go` (with `ExchangeCode`,
`RefreshToken`, `GetMe`) that is tightly coupled to the backend's DB-backed
connect/sync service and circuit breaker. This SDK is the **standalone,
dependency-free** equivalent for external consumers and mirrors the same OAuth
call shapes (grant types, endpoints) so the two stay behaviourally aligned.
The backend client is intentionally **not** replaced by this SDK — it is
specialised for in-process use.

## Operating principles

- Every method takes `context.Context` first; cancellation propagates.
- Each `*Client` owns one `*http.Client` with explicit `Timeout`,
  `MaxIdleConnsPerHost`, and `IdleConnTimeout`. `http.DefaultClient` is never
  used.
- `defer client.Close()` releases idle connections.
- API errors surface as `*threads.Error` with `StatusCode`, `Type`, `Code`,
  `Subcode`, `Message`, and `FBTraceID`.

## Repository layout

The Go SDK lives in the `go/` subdirectory. The Python SDK remains under
`threads/` and is unchanged. See the top-level [README](../README.md) for the
multi-language overview.
