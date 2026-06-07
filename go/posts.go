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
