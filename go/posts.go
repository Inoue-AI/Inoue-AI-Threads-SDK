package threads

import (
	"context"
	"errors"
	"net/url"
	"strconv"
	"strings"
)

// Post mirrors the Threads media object fields the Inoue backend consumes.
// Fields are populated only when requested via the `fields` argument.
type Post struct {
	ID               string `json:"id,omitempty"`
	MediaProductType string `json:"media_product_type,omitempty"`
	MediaType        string `json:"media_type,omitempty"`
	MediaURL         string `json:"media_url,omitempty"`
	Permalink        string `json:"permalink,omitempty"`
	Owner            struct {
		ID string `json:"id,omitempty"`
	} `json:"owner,omitempty"`
	Username    string `json:"username,omitempty"`
	Text        string `json:"text,omitempty"`
	Timestamp   string `json:"timestamp,omitempty"`
	ShortcodeID string `json:"shortcode,omitempty"`
	IsQuotePost bool   `json:"is_quote_post,omitempty"`
}

// PostListResult is the cursor-paginated wrapper for /<user_id>/threads.
type PostListResult struct {
	Data   []Post `json:"data"`
	Paging struct {
		Cursors struct {
			Before string `json:"before"`
			After  string `json:"after"`
		} `json:"cursors"`
		Next     string `json:"next,omitempty"`
		Previous string `json:"previous,omitempty"`
	} `json:"paging"`
}

// ListPostsParams configures ListPosts.
type ListPostsParams struct {
	UserID string   // Defaults to "me".
	Fields []string // Defaults to a sensible set when empty.
	Limit  int      // Defaults to 25 when zero.
	Since  string   // Optional ISO-8601 timestamp.
	Until  string   // Optional ISO-8601 timestamp.
	After  string   // Optional pagination cursor.
	Before string   // Optional pagination cursor.
}

// ListPosts returns one page of posts for the user.
func (c *Client) ListPosts(ctx context.Context, p ListPostsParams) (*PostListResult, error) {
	userID := p.UserID
	if userID == "" {
		userID = "me"
	}
	fields := p.Fields
	if len(fields) == 0 {
		fields = []string{
			"id", "media_product_type", "media_type", "media_url",
			"permalink", "owner", "username", "text", "timestamp",
			"shortcode", "is_quote_post",
		}
	}
	limit := p.Limit
	if limit <= 0 {
		limit = 25
	}

	q := url.Values{}
	q.Set("fields", strings.Join(fields, ","))
	q.Set("limit", strconv.Itoa(limit))
	if p.Since != "" {
		q.Set("since", p.Since)
	}
	if p.Until != "" {
		q.Set("until", p.Until)
	}
	if p.After != "" {
		q.Set("after", p.After)
	}
	if p.Before != "" {
		q.Set("before", p.Before)
	}

	out := &PostListResult{}
	if err := c.doGet(ctx, userID+"/threads", q, out); err != nil {
		return nil, err
	}
	return out, nil
}

// IterPostsParams configures IterPosts. It mirrors the page-level inputs of
// ListPostsParams but omits the cursor fields (After/Before) because IterPosts
// drives pagination itself.
type IterPostsParams struct {
	UserID string   // Defaults to "me".
	Fields []string // Defaults to a sensible set when empty.
	Limit  int      // Per-page size. Defaults to 25 when zero.
	Since  string   // Optional ISO-8601 timestamp.
	Until  string   // Optional ISO-8601 timestamp.
	// MaxPages caps how many pages are fetched (0 = no cap; follow cursors
	// until the API stops returning an "after" cursor). A cap guards against
	// runaway iteration and unbounded memory growth.
	MaxPages int
}

// IterPosts returns all posts for a user across pages, transparently following
// the cursor-based "after" links. It is the Go equivalent of the Python SDK's
// async generator MediaAPI.iter_threads: where Python yields lazily, this Go
// helper accumulates into a slice (the repo's established convention for
// convenience iterators). Pagination stops when the API returns no further
// "after" cursor, when MaxPages is reached, or when ctx is cancelled.
func (c *Client) IterPosts(ctx context.Context, p IterPostsParams) ([]Post, error) {
	all := make([]Post, 0)
	after := ""
	pages := 0
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		page, err := c.ListPosts(ctx, ListPostsParams{
			UserID: p.UserID,
			Fields: p.Fields,
			Limit:  p.Limit,
			Since:  p.Since,
			Until:  p.Until,
			After:  after,
		})
		if err != nil {
			return nil, err
		}
		all = append(all, page.Data...)
		pages++

		next := page.Paging.Cursors.After
		if next == "" {
			break
		}
		if p.MaxPages > 0 && pages >= p.MaxPages {
			break
		}
		after = next
	}
	return all, nil
}

// GetPost fetches a single post by media ID.
func (c *Client) GetPost(ctx context.Context, mediaID string, fields []string) (*Post, error) {
	if mediaID == "" {
		return nil, errors.New("threads: GetPost requires a non-empty mediaID")
	}
	if len(fields) == 0 {
		fields = []string{
			"id", "media_product_type", "media_type", "media_url",
			"permalink", "owner", "username", "text", "timestamp",
			"shortcode", "is_quote_post",
		}
	}
	q := url.Values{}
	q.Set("fields", strings.Join(fields, ","))

	out := &Post{}
	if err := c.doGet(ctx, mediaID, q, out); err != nil {
		return nil, err
	}
	return out, nil
}
