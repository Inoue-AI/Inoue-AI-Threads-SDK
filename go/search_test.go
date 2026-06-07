package threads

import (
	"context"
	"io"
	"net/http"
	"testing"
)

func TestSearch_Basic(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/keyword_search" {
			t.Errorf("path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("q"); got != "Python" {
			t.Errorf("q: %s", got)
		}
		_, _ = io.WriteString(w, `{"data":[
			{"id":"s1","text":"Python is great","media_type":"TEXT_POST"},
			{"id":"s2","text":"Learning Python","media_type":"TEXT_POST"}
		],"paging":{"cursors":{"after":"next_page"}}}`)
	})
	res, err := c.Search(context.Background(), SearchParams{Query: "Python"})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(res.Data) != 2 || res.Data[0].Text != "Python is great" {
		t.Fatalf("unexpected results: %+v", res.Data)
	}
	if res.Paging.Cursors.After != "next_page" {
		t.Fatalf("expected paging cursor, got %q", res.Paging.Cursors.After)
	}
}

func TestSearch_WithMediaType(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("media_type"); got != "VIDEO" {
			t.Errorf("media_type: %s", got)
		}
		_, _ = io.WriteString(w, `{"data":[{"id":"v1","media_type":"VIDEO"}]}`)
	})
	res, err := c.Search(context.Background(), SearchParams{
		Query:     "tutorial",
		MediaType: SearchMediaTypeVideo,
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(res.Data) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res.Data))
	}
}

func TestSearch_DateBounds(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("since") != "100" || q.Get("until") != "200" {
			t.Errorf("since/until: %s %s", q.Get("since"), q.Get("until"))
		}
		_, _ = io.WriteString(w, `{"data":[]}`)
	})
	if _, err := c.Search(context.Background(), SearchParams{
		Query: "x",
		Since: "100",
		Until: "200",
	}); err != nil {
		t.Fatalf("Search: %v", err)
	}
}

func TestSearch_RequiresQuery(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("server must not be called")
	})
	if _, err := c.Search(context.Background(), SearchParams{}); err == nil {
		t.Fatal("expected error for empty Query")
	}
}
