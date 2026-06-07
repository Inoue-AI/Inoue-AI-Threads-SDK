package threads

import (
	"context"
	"io"
	"net/http"
	"testing"
)

func TestIterPosts_FollowsCursors(t *testing.T) {
	calls := 0
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		after := r.URL.Query().Get("after")
		switch after {
		case "":
			// Page 1 returns a forward cursor.
			_, _ = io.WriteString(w, `{"data":[{"id":"t1"},{"id":"t2"}],"paging":{"cursors":{"after":"cursor_1"}}}`)
		case "cursor_1":
			// Page 2 is terminal (no cursor).
			_, _ = io.WriteString(w, `{"data":[{"id":"t3"}],"paging":{"cursors":{}}}`)
		default:
			t.Errorf("unexpected after cursor: %s", after)
		}
	})
	posts, err := c.IterPosts(context.Background(), IterPostsParams{Limit: 2})
	if err != nil {
		t.Fatalf("IterPosts: %v", err)
	}
	ids := make([]string, len(posts))
	for i, p := range posts {
		ids[i] = p.ID
	}
	if len(ids) != 3 || ids[0] != "t1" || ids[1] != "t2" || ids[2] != "t3" {
		t.Fatalf("unexpected ids: %v", ids)
	}
	if calls != 2 {
		t.Fatalf("expected 2 page fetches, got %d", calls)
	}
}

func TestIterPosts_MaxPages(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		// Always return a forward cursor; MaxPages must stop the loop.
		_, _ = io.WriteString(w, `{"data":[{"id":"x"}],"paging":{"cursors":{"after":"more"}}}`)
	})
	posts, err := c.IterPosts(context.Background(), IterPostsParams{MaxPages: 2})
	if err != nil {
		t.Fatalf("IterPosts: %v", err)
	}
	if len(posts) != 2 {
		t.Fatalf("expected MaxPages=2 to yield 2 posts, got %d", len(posts))
	}
}

func TestIterPosts_CtxCancel(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"data":[{"id":"x"}],"paging":{"cursors":{"after":"more"}}}`)
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := c.IterPosts(ctx, IterPostsParams{}); err == nil {
		t.Fatal("expected ctx cancellation error")
	}
}
