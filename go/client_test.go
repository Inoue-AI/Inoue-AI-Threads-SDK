package threads

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c := New(ClientOptions{
		AccessToken: "TKN",
		BaseURL:     srv.URL,
		Timeout:     5 * time.Second,
	})
	t.Cleanup(func() { _ = c.Close() })
	return c, srv
}

func TestNew_DefaultsHTTPClient(t *testing.T) {
	c := New(ClientOptions{AccessToken: "x"})
	defer c.Close()
	if c.HTTPClient() == http.DefaultClient {
		t.Fatal("must NOT use http.DefaultClient")
	}
	if c.HTTPClient().Timeout == 0 {
		t.Fatal("must set explicit Timeout")
	}
	tr, ok := c.HTTPClient().Transport.(*http.Transport)
	if !ok {
		t.Fatalf("expected *http.Transport, got %T", c.HTTPClient().Transport)
	}
	if tr.MaxIdleConnsPerHost == 0 || tr.IdleConnTimeout == 0 {
		t.Fatal("Transport must configure idle-conn limits")
	}
}

func TestGetUser_AppendsAccessToken(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me" {
			t.Errorf("path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("access_token"); got != "TKN" {
			t.Errorf("access_token query: %q", got)
		}
		if !strings.Contains(r.URL.Query().Get("fields"), "username") {
			t.Errorf("missing username in fields: %s", r.URL.Query().Get("fields"))
		}
		_, _ = io.WriteString(w, `{"id":"1","username":"u","name":"N"}`)
	})
	user, err := c.GetUser(context.Background(), "me", nil)
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	if user.Username != "u" || user.ID != "1" {
		t.Fatalf("unexpected user: %+v", user)
	}
}

func TestListPosts_PassesPagination(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me/threads" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("limit"); got != "5" {
			t.Errorf("limit: %s", got)
		}
		if got := r.URL.Query().Get("after"); got != "CURSOR" {
			t.Errorf("after: %s", got)
		}
		_, _ = io.WriteString(w, `{"data":[{"id":"P1","text":"hi"}],"paging":{"cursors":{"after":"NEXT"}}}`)
	})
	page, err := c.ListPosts(context.Background(), ListPostsParams{Limit: 5, After: "CURSOR"})
	if err != nil {
		t.Fatalf("ListPosts: %v", err)
	}
	if len(page.Data) != 1 || page.Data[0].ID != "P1" {
		t.Fatalf("unexpected posts: %+v", page.Data)
	}
	if page.Paging.Cursors.After != "NEXT" {
		t.Fatalf("expected next cursor")
	}
}

func TestGetPost_RequiresID(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("server must not be called")
	})
	if _, err := c.GetPost(context.Background(), "", nil); err == nil {
		t.Fatal("expected error for empty mediaID")
	}
}

func TestErrorMapping(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, `{"error":{"message":"bad","type":"OAuthException","code":190,"fbtrace_id":"T"}}`)
	})
	_, err := c.GetUser(context.Background(), "me", nil)
	apiErr, ok := AsError(err)
	if !ok {
		t.Fatalf("expected *Error, got %v", err)
	}
	if !apiErr.IsAuthError() {
		t.Fatalf("expected IsAuthError, got %+v", apiErr)
	}
	if apiErr.Code != 190 || apiErr.FBTraceID != "T" {
		t.Fatalf("unexpected error fields: %+v", apiErr)
	}
}

func TestRefreshAccessToken_Success(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/refresh_access_token" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("grant_type") != "th_refresh_token" {
			t.Errorf("missing grant_type")
		}
		if r.URL.Query().Get("access_token") != "OLD" {
			t.Errorf("missing access_token")
		}
		_, _ = io.WriteString(w, `{"access_token":"NEW","token_type":"bearer","expires_in":1234}`)
	})
	out, err := c.RefreshAccessToken(context.Background(), RefreshTokenParams{AccessToken: "OLD"})
	if err != nil {
		t.Fatalf("RefreshAccessToken: %v", err)
	}
	if out.AccessToken != "NEW" || out.ExpiresIn != 1234 {
		t.Fatalf("unexpected refresh result: %+v", out)
	}
}

func TestDoGet_RequiresAccessToken(t *testing.T) {
	c := New(ClientOptions{})
	defer c.Close()
	_, err := c.GetUser(context.Background(), "me", nil)
	if err == nil || !strings.Contains(err.Error(), "access token") {
		t.Fatalf("expected access token error, got %v", err)
	}
}
