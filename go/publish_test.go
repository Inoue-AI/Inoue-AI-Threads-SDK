package threads

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"testing"
)

func TestCreateContainer_Text(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me/threads" {
			t.Errorf("path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("method: %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		form, _ := url.ParseQuery(string(body))
		if form.Get("media_type") != "TEXT" {
			t.Errorf("media_type: %s", form.Get("media_type"))
		}
		if form.Get("text") != "hello" {
			t.Errorf("text: %s", form.Get("text"))
		}
		_, _ = io.WriteString(w, `{"id":"CONTAINER1"}`)
	})
	out, err := c.CreateContainer(context.Background(), CreateContainerParams{
		MediaType: MediaTypeText,
		Text:      "hello",
	})
	if err != nil {
		t.Fatalf("CreateContainer: %v", err)
	}
	if out.ID != "CONTAINER1" {
		t.Fatalf("unexpected id: %s", out.ID)
	}
}

func TestCreateContainer_RequiresMediaType(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("server must not be called")
	})
	if _, err := c.CreateContainer(context.Background(), CreateContainerParams{Text: "x"}); err == nil {
		t.Fatal("expected error for missing MediaType")
	}
}

func TestPublish(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me/threads_publish" {
			t.Errorf("path: %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		form, _ := url.ParseQuery(string(body))
		if form.Get("creation_id") != "CONTAINER1" {
			t.Errorf("creation_id: %s", form.Get("creation_id"))
		}
		_, _ = io.WriteString(w, `{"id":"PUBLISHED1"}`)
	})
	out, err := c.Publish(context.Background(), "me", "CONTAINER1")
	if err != nil {
		t.Fatalf("Publish: %v", err)
	}
	if out.ID != "PUBLISHED1" {
		t.Fatalf("unexpected id: %s", out.ID)
	}
}

func TestPostText_FullFlow(t *testing.T) {
	calls := 0
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		switch r.URL.Path {
		case "/me/threads":
			_, _ = io.WriteString(w, `{"id":"C1"}`)
		case "/me/threads_publish":
			_, _ = io.WriteString(w, `{"id":"P1"}`)
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	})
	out, err := c.PostText(context.Background(), "me", "hi there")
	if err != nil {
		t.Fatalf("PostText: %v", err)
	}
	if out.ID != "P1" {
		t.Fatalf("unexpected id: %s", out.ID)
	}
	if calls != 2 {
		t.Fatalf("expected create+publish (2 calls), got %d", calls)
	}
}

func TestPostText_RequiresText(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("server must not be called")
	})
	if _, err := c.PostText(context.Background(), "me", ""); err == nil {
		t.Fatal("expected error for empty text")
	}
}

func TestDeletePost(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/thread_99" {
			t.Errorf("path: %s", r.URL.Path)
		}
		if r.Method != http.MethodDelete {
			t.Errorf("method: %s", r.Method)
		}
		_, _ = io.WriteString(w, `{"success":true}`)
	})
	ok, err := c.DeletePost(context.Background(), "thread_99")
	if err != nil {
		t.Fatalf("DeletePost: %v", err)
	}
	if !ok {
		t.Fatal("expected success=true")
	}
}
