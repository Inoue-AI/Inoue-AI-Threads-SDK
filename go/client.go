package threads

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// DefaultBaseURL is the production Threads Graph API root.
const DefaultBaseURL = "https://graph.threads.net/v1.0"

// DefaultTimeout is the per-request timeout used when ClientOptions.Timeout
// is zero. 30s matches the Python SDK default.
const DefaultTimeout = 30 * time.Second

// ClientOptions configures a *Client created via New.
type ClientOptions struct {
	// AccessToken is the user OAuth token (required for user-scoped endpoints).
	// RefreshAccessToken does not require this field.
	AccessToken string

	// BaseURL overrides the production Graph API host (useful for testing).
	BaseURL string

	// Timeout sets the per-request total timeout. Defaults to DefaultTimeout.
	Timeout time.Duration

	// HTTPClient lets callers inject a fully configured *http.Client. When nil,
	// New constructs a client with bounded idle connections and an explicit
	// Timeout. The package never falls back to http.DefaultClient.
	HTTPClient *http.Client

	// UserAgent overrides the User-Agent header.
	UserAgent string
}

// Client is the Threads API client. Safe for concurrent use. Callers must
// invoke Close (or use defer) so idle TCP connections are released.
type Client struct {
	httpClient  *http.Client
	baseURL     string
	accessToken string
	userAgent   string
	ownsHTTP    bool
}

// New constructs a *Client with the supplied options.
func New(opts ClientOptions) *Client {
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	baseURL := strings.TrimRight(opts.BaseURL, "/")
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	ua := opts.UserAgent
	if ua == "" {
		ua = "inoue-threads-sdk-go/1"
	}
	c := &Client{
		baseURL:     baseURL,
		accessToken: opts.AccessToken,
		userAgent:   ua,
	}
	if opts.HTTPClient != nil {
		c.httpClient = opts.HTTPClient
		c.ownsHTTP = false
	} else {
		c.httpClient = &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:          100,
				MaxIdleConnsPerHost:   20,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
				ResponseHeaderTimeout: timeout,
			},
		}
		c.ownsHTTP = true
	}
	return c
}

// Close releases idle TCP connections held by the underlying *http.Client.
func (c *Client) Close() error {
	if c == nil || c.httpClient == nil {
		return nil
	}
	if c.ownsHTTP {
		if t, ok := c.httpClient.Transport.(*http.Transport); ok {
			t.CloseIdleConnections()
		}
	}
	return nil
}

// HTTPClient returns the underlying *http.Client.
func (c *Client) HTTPClient() *http.Client { return c.httpClient }

// graphErrorPayload mirrors the standard Meta error envelope.
type graphErrorPayload struct {
	Error struct {
		Message      string `json:"message"`
		Type         string `json:"type"`
		Code         int    `json:"code"`
		ErrorSubcode int    `json:"error_subcode"`
		FBTraceID    string `json:"fbtrace_id"`
	} `json:"error"`
}

// doGet executes an authenticated GET against the Graph API and decodes JSON
// into out. The access_token is appended automatically from the Client.
func (c *Client) doGet(ctx context.Context, path string, query url.Values, out any) error {
	if c.accessToken == "" {
		return errors.New("threads: access token is required for this endpoint")
	}
	if query == nil {
		query = url.Values{}
	}
	query.Set("access_token", c.accessToken)

	fullURL := c.baseURL + "/" + strings.TrimLeft(path, "/") + "?" + query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return fmt.Errorf("threads: build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	return c.do(req, out)
}

// doPostForm executes an authenticated POST that submits an
// application/x-www-form-urlencoded body. Used by container creation and
// publish endpoints, which expect form-encoded params.
func (c *Client) doPostForm(ctx context.Context, path string, form url.Values, out any) error {
	if c.accessToken == "" {
		return errors.New("threads: access token is required for this endpoint")
	}
	if form == nil {
		form = url.Values{}
	}
	form.Set("access_token", c.accessToken)

	fullURL := c.baseURL + "/" + strings.TrimLeft(path, "/")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("threads: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	return c.do(req, out)
}

// doRefresh executes the unauthenticated refresh-token endpoint. This call
// uses application credentials and does not consume the user token.
func (c *Client) doRefresh(ctx context.Context, query url.Values, out any) error {
	fullURL := c.baseURL + "/refresh_access_token?" + query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return fmt.Errorf("threads: build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	return c.do(req, out)
}

// do is the common request execution + error-handling helper.
func (c *Client) do(req *http.Request, out any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("threads: %s %s: %w", req.Method, req.URL.Path, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("threads: read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var ep graphErrorPayload
		_ = json.Unmarshal(raw, &ep)
		return &Error{
			StatusCode: resp.StatusCode,
			Type:       ep.Error.Type,
			Code:       ep.Error.Code,
			Subcode:    ep.Error.ErrorSubcode,
			Message:    ep.Error.Message,
			FBTraceID:  ep.Error.FBTraceID,
			Body:       raw,
		}
	}

	if out == nil {
		return nil
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return &Error{
			StatusCode: resp.StatusCode,
			Type:       "invalid_response",
			Message:    fmt.Sprintf("Threads returned non-JSON: %v", err),
			Body:       raw,
		}
	}
	return nil
}

