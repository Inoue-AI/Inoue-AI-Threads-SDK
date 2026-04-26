package threads

import (
	"context"
	"errors"
	"net/url"
)

// RefreshTokenParams carries the inputs to the Threads token refresh endpoint.
//
// Threads issues long-lived tokens that can be refreshed via
// GET /refresh_access_token while still valid. The endpoint requires only the
// access token itself (no client_secret).
type RefreshTokenParams struct {
	AccessToken string // The current long-lived access token to refresh.
}

// RefreshTokenResult is the decoded response.
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
	q.Set("grant_type", "th_refresh_token")
	q.Set("access_token", p.AccessToken)

	out := &RefreshTokenResult{}
	if err := c.doRefresh(ctx, q, out); err != nil {
		return nil, err
	}
	return out, nil
}
