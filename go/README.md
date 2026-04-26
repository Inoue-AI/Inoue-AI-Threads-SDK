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

| Go method | Threads endpoint |
|---|---|
| `GetUser(ctx, userID, fields)` | `GET /{user-id}` |
| `ListPosts(ctx, params)` | `GET /{user-id}/threads` |
| `GetPost(ctx, mediaID, fields)` | `GET /{media-id}` |
| `RefreshAccessToken(ctx, params)` | `GET /refresh_access_token` |

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
