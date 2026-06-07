package threads

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestAuthorizationURL(t *testing.T) {
	got, err := AuthorizationURL(AuthorizeParams{
		ClientID:    "CID",
		RedirectURI: "https://app.example.com/cb",
		Scopes:      []Scope{ScopeBasic, ScopeContentPublish},
		State:       "csrf1",
	})
	if err != nil {
		t.Fatalf("AuthorizationURL: %v", err)
	}
	u, err := url.Parse(got)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if u.Host != "threads.net" || u.Path != "/oauth/authorize" {
		t.Fatalf("unexpected host/path: %s", got)
	}
	q := u.Query()
	if q.Get("client_id") != "CID" {
		t.Errorf("client_id: %q", q.Get("client_id"))
	}
	if q.Get("scope") != "threads_basic,threads_content_publish" {
		t.Errorf("scope: %q", q.Get("scope"))
	}
	if q.Get("response_type") != "code" {
		t.Errorf("response_type: %q", q.Get("response_type"))
	}
	if q.Get("state") != "csrf1" {
		t.Errorf("state: %q", q.Get("state"))
	}
}

func TestAuthorizationURL_Validation(t *testing.T) {
	if _, err := AuthorizationURL(AuthorizeParams{RedirectURI: "r", Scopes: []Scope{ScopeBasic}}); err == nil {
		t.Error("expected error for missing ClientID")
	}
	if _, err := AuthorizationURL(AuthorizeParams{ClientID: "c", Scopes: []Scope{ScopeBasic}}); err == nil {
		t.Error("expected error for missing RedirectURI")
	}
	if _, err := AuthorizationURL(AuthorizeParams{ClientID: "c", RedirectURI: "r"}); err == nil {
		t.Error("expected error for missing Scopes")
	}
}

func TestExchangeCode_Success(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oauth/access_token" {
			t.Errorf("path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("method: %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		form, _ := url.ParseQuery(string(body))
		if form.Get("grant_type") != "authorization_code" {
			t.Errorf("grant_type: %s", form.Get("grant_type"))
		}
		if form.Get("code") != "CODE" {
			t.Errorf("code: %s", form.Get("code"))
		}
		if form.Get("client_secret") != "SECRET" {
			t.Errorf("client_secret missing")
		}
		_, _ = io.WriteString(w, `{"access_token":"SHORT","user_id":42}`)
	})
	out, err := c.ExchangeCode(context.Background(), ExchangeCodeParams{
		ClientID:     "CID",
		ClientSecret: "SECRET",
		RedirectURI:  "https://app.example.com/cb",
		Code:         "CODE",
	})
	if err != nil {
		t.Fatalf("ExchangeCode: %v", err)
	}
	if out.AccessToken != "SHORT" || out.UserID != 42 {
		t.Fatalf("unexpected token: %+v", out)
	}
}

func TestExchangeCode_Validation(t *testing.T) {
	c := New(ClientOptions{AccessToken: "x"})
	defer c.Close()
	if _, err := c.ExchangeCode(context.Background(), ExchangeCodeParams{ClientSecret: "s", Code: "c", RedirectURI: "r"}); err == nil {
		t.Error("expected error for missing ClientID")
	}
	if _, err := c.ExchangeCode(context.Background(), ExchangeCodeParams{ClientID: "c", ClientSecret: "s", RedirectURI: "r"}); err == nil {
		t.Error("expected error for missing Code")
	}
}

func TestExchangeForLongLived_Success(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/access_token" {
			t.Errorf("path: %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("grant_type") != "th_exchange_token" {
			t.Errorf("grant_type: %s", q.Get("grant_type"))
		}
		if q.Get("client_secret") != "SECRET" || q.Get("access_token") != "SHORT" {
			t.Errorf("missing creds: %v", q)
		}
		_, _ = io.WriteString(w, `{"access_token":"LONG","token_type":"bearer","expires_in":5183944}`)
	})
	out, err := c.ExchangeForLongLived(context.Background(), ExchangeForLongLivedParams{
		ClientSecret:    "SECRET",
		ShortLivedToken: "SHORT",
	})
	if err != nil {
		t.Fatalf("ExchangeForLongLived: %v", err)
	}
	if out.AccessToken != "LONG" || out.ExpiresIn != 5183944 {
		t.Fatalf("unexpected token: %+v", out)
	}
}

func TestOAuthBaseURLFromGraph(t *testing.T) {
	cases := map[string]string{
		"https://graph.threads.net/v1.0": "https://graph.threads.net",
		"https://graph.threads.net":      "https://graph.threads.net",
		"http://127.0.0.1:8080/v1.0":     "http://127.0.0.1:8080",
	}
	for in, want := range cases {
		if got := oauthBaseURLFromGraph(in); got != want {
			t.Errorf("oauthBaseURLFromGraph(%q)=%q want %q", in, got, want)
		}
	}
}

func TestOAuth_StripsVersionPrefix(t *testing.T) {
	// When BaseURL carries a /v1.0 prefix, OAuth calls must hit the
	// unversioned host root, not /v1.0/refresh_access_token.
	var gotPath string
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = io.WriteString(w, `{"access_token":"NEW","expires_in":1}`)
	})
	// Re-point the client at a versioned base URL backed by the same server.
	c2 := New(ClientOptions{AccessToken: "x", BaseURL: srv.URL + "/v1.0"})
	defer c2.Close()
	if _, err := c2.RefreshAccessToken(context.Background(), RefreshTokenParams{AccessToken: "OLD"}); err != nil {
		t.Fatalf("RefreshAccessToken: %v", err)
	}
	if strings.Contains(gotPath, "v1.0") {
		t.Fatalf("OAuth call must not include version prefix, got %s", gotPath)
	}
	if gotPath != "/refresh_access_token" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	_ = c
}
