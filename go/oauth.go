package threads

import (
	"context"
	"errors"
	"net/url"
	"strings"
)

// OAuth grant-type constants used by the Threads token endpoints.
const (
	grantAuthorizationCode = "authorization_code"
	grantExchangeToken     = "th_exchange_token"
	grantRefreshToken      = "th_refresh_token"
)

// DefaultAuthorizeBaseURL is the host that serves the OAuth consent UI.
// The token-exchange + refresh endpoints live on the Graph host (BaseURL).
const DefaultAuthorizeBaseURL = "https://threads.net"

// Scope is a Threads OAuth permission. Request the minimal set required.
type Scope string

// Recognised Threads OAuth scopes.
const (
	ScopeBasic          Scope = "threads_basic"
	ScopeContentPublish Scope = "threads_content_publish"
	ScopeManageInsights Scope = "threads_manage_insights"
	ScopeReadReplies    Scope = "threads_read_replies"
	ScopeManageReplies  Scope = "threads_manage_replies"
	ScopeKeywordSearch  Scope = "threads_keyword_search"
	ScopeManageMentions Scope = "threads_manage_mentions"
	ScopeDelete         Scope = "threads_delete"
)

// AuthorizeParams configures AuthorizationURL.
type AuthorizeParams struct {
	ClientID    string  // Threads app client id (required).
	RedirectURI string  // Registered redirect URI (required).
	Scopes      []Scope // At least one scope (required).
	State       string  // Optional CSRF / round-trip token.
	// AuthorizeBaseURL overrides the consent-UI host. Defaults to
	// DefaultAuthorizeBaseURL when empty.
	AuthorizeBaseURL string
}

// AuthorizationURL builds the URL to redirect the user to for consent. It
// performs no I/O. The user is sent here; on approval Threads redirects back
// to RedirectURI with a one-time ?code= the caller passes to ExchangeCode.
func AuthorizationURL(p AuthorizeParams) (string, error) {
	if p.ClientID == "" {
		return "", errors.New("threads: AuthorizationURL requires ClientID")
	}
	if p.RedirectURI == "" {
		return "", errors.New("threads: AuthorizationURL requires RedirectURI")
	}
	if len(p.Scopes) == 0 {
		return "", errors.New("threads: AuthorizationURL requires at least one Scope")
	}
	scopes := make([]string, len(p.Scopes))
	for i, s := range p.Scopes {
		scopes[i] = string(s)
	}
	base := strings.TrimRight(p.AuthorizeBaseURL, "/")
	if base == "" {
		base = DefaultAuthorizeBaseURL
	}
	q := url.Values{}
	q.Set("client_id", p.ClientID)
	q.Set("redirect_uri", p.RedirectURI)
	q.Set("scope", strings.Join(scopes, ","))
	q.Set("response_type", "code")
	if p.State != "" {
		q.Set("state", p.State)
	}
	return base + "/oauth/authorize?" + q.Encode(), nil
}

// ShortLivedToken is the result of exchanging an authorization code.
type ShortLivedToken struct {
	AccessToken string `json:"access_token"`
	UserID      int64  `json:"user_id,omitempty"`
}

// LongLivedToken is the result of exchanging a short-lived token, or of
// refreshing a long-lived token. Threads long-lived tokens last ~60 days.
type LongLivedToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type,omitempty"`
	ExpiresIn   int64  `json:"expires_in,omitempty"`
}

// ExchangeCodeParams carries the inputs to the code->token exchange.
type ExchangeCodeParams struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	Code         string
}

// ExchangeCode swaps a one-time authorization code for a short-lived access
// token. This is the first server-side step of the OAuth flow and mirrors the
// backend's in-tree threads.Client.ExchangeCode. The client secret is sent in
// a form-encoded POST body over TLS and is never logged by this package.
func (c *Client) ExchangeCode(ctx context.Context, p ExchangeCodeParams) (*ShortLivedToken, error) {
	if p.ClientID == "" || p.ClientSecret == "" {
		return nil, errors.New("threads: ExchangeCode requires ClientID and ClientSecret")
	}
	if p.Code == "" {
		return nil, errors.New("threads: ExchangeCode requires Code")
	}
	if p.RedirectURI == "" {
		return nil, errors.New("threads: ExchangeCode requires RedirectURI")
	}
	form := url.Values{}
	form.Set("client_id", p.ClientID)
	form.Set("client_secret", p.ClientSecret)
	form.Set("grant_type", grantAuthorizationCode)
	form.Set("redirect_uri", p.RedirectURI)
	form.Set("code", p.Code)

	out := &ShortLivedToken{}
	if err := c.doOAuthPost(ctx, "/oauth/access_token", form, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ExchangeForLongLivedParams carries the inputs to the short->long exchange.
type ExchangeForLongLivedParams struct {
	ClientSecret    string
	ShortLivedToken string
}

// ExchangeForLongLived swaps a short-lived token for a long-lived (~60 day)
// token via GET /access_token?grant_type=th_exchange_token.
func (c *Client) ExchangeForLongLived(ctx context.Context, p ExchangeForLongLivedParams) (*LongLivedToken, error) {
	if p.ClientSecret == "" {
		return nil, errors.New("threads: ExchangeForLongLived requires ClientSecret")
	}
	if p.ShortLivedToken == "" {
		return nil, errors.New("threads: ExchangeForLongLived requires ShortLivedToken")
	}
	q := url.Values{}
	q.Set("grant_type", grantExchangeToken)
	q.Set("client_secret", p.ClientSecret)
	q.Set("access_token", p.ShortLivedToken)

	out := &LongLivedToken{}
	if err := c.doOAuthGet(ctx, "/access_token", q, out); err != nil {
		return nil, err
	}
	return out, nil
}

// RefreshTokenParams carries the inputs to the Threads token refresh endpoint.
//
// Threads issues long-lived tokens that can be refreshed via
// GET /refresh_access_token while still valid. The endpoint requires only the
// access token itself (no client_secret).
type RefreshTokenParams struct {
	AccessToken string // The current long-lived access token to refresh.
}

// RefreshTokenResult is the decoded response. Retained as a distinct type for
// backwards compatibility; it is structurally identical to LongLivedToken.
type RefreshTokenResult struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

// RefreshAccessToken exchanges a still-valid long-lived token for a refreshed
// version. This call does not consume the *Client's stored AccessToken so the
// caller can refresh tokens on behalf of any user without reconfiguring.
func (c *Client) RefreshAccessToken(ctx context.Context, p RefreshTokenParams) (*RefreshTokenResult, error) {
	if p.AccessToken == "" {
		return nil, errors.New("threads: RefreshTokenParams requires AccessToken")
	}
	q := url.Values{}
	q.Set("grant_type", grantRefreshToken)
	q.Set("access_token", p.AccessToken)

	out := &RefreshTokenResult{}
	if err := c.doOAuthGet(ctx, "/refresh_access_token", q, out); err != nil {
		return nil, err
	}
	return out, nil
}
