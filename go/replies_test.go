package threads

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"testing"
)

func TestGetReplies(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/post_1/replies" {
			t.Errorf("path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("reverse") != "" {
			t.Errorf("reverse should be unset by default")
		}
		_, _ = io.WriteString(w, `{"data":[
			{"id":"r1","text":"Great post!","username":"fan1"},
			{"id":"r2","text":"Love it","username":"fan2"}
		]}`)
	})
	res, err := c.GetReplies(context.Background(), "post_1", nil, false)
	if err != nil {
		t.Fatalf("GetReplies: %v", err)
	}
	if len(res.Data) != 2 {
		t.Fatalf("expected 2 replies, got %d", len(res.Data))
	}
	if res.Data[0].Text != "Great post!" || res.Data[1].Username != "fan2" {
		t.Fatalf("unexpected replies: %+v", res.Data)
	}
}

func TestGetReplies_Reverse(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("reverse") != "true" {
			t.Errorf("reverse: %s", r.URL.Query().Get("reverse"))
		}
		_, _ = io.WriteString(w, `{"data":[]}`)
	})
	if _, err := c.GetReplies(context.Background(), "post_1", nil, true); err != nil {
		t.Fatalf("GetReplies: %v", err)
	}
}

func TestGetReplies_RequiresID(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("server must not be called")
	})
	if _, err := c.GetReplies(context.Background(), "", nil, false); err == nil {
		t.Fatal("expected error for empty mediaID")
	}
}

func TestGetConversation(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/post_1/conversation" {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = io.WriteString(w, `{"data":[
			{"id":"r1","text":"First reply"},
			{"id":"r1_1","text":"Reply to reply"}
		]}`)
	})
	res, err := c.GetConversation(context.Background(), "post_1", nil, false)
	if err != nil {
		t.Fatalf("GetConversation: %v", err)
	}
	if len(res.Data) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(res.Data))
	}
}

func TestHideReply(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/reply_1/manage_reply" {
			t.Errorf("path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("method: %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		form, _ := url.ParseQuery(string(body))
		if form.Get("hide") != "true" {
			t.Errorf("hide: %s", form.Get("hide"))
		}
		_, _ = io.WriteString(w, `{"success":true}`)
	})
	ok, err := c.HideReply(context.Background(), "reply_1")
	if err != nil {
		t.Fatalf("HideReply: %v", err)
	}
	if !ok {
		t.Fatal("expected success=true")
	}
}

func TestUnhideReply(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		form, _ := url.ParseQuery(string(body))
		if form.Get("hide") != "false" {
			t.Errorf("hide: %s", form.Get("hide"))
		}
		_, _ = io.WriteString(w, `{"success":true}`)
	})
	ok, err := c.UnhideReply(context.Background(), "reply_1")
	if err != nil {
		t.Fatalf("UnhideReply: %v", err)
	}
	if !ok {
		t.Fatal("expected success=true")
	}
}

func TestManageReply_RequiresID(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("server must not be called")
	})
	if _, err := c.HideReply(context.Background(), ""); err == nil {
		t.Fatal("expected error for empty replyID")
	}
}

func TestGetMentions(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me/mentions" {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = io.WriteString(w, `{"data":[{"id":"m1","text":"Hey @user check this!","username":"someone"}]}`)
	})
	res, err := c.GetMentions(context.Background(), nil)
	if err != nil {
		t.Fatalf("GetMentions: %v", err)
	}
	if len(res.Data) != 1 || res.Data[0].ID != "m1" {
		t.Fatalf("unexpected mentions: %+v", res.Data)
	}
}
