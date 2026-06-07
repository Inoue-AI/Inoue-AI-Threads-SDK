package threads

import (
	"context"
	"errors"
	"net/url"
	"strings"
)

// SearchMediaType is the optional media-type filter for keyword search. It is a
// distinct domain from publishing's MediaType (the keyword-search API accepts
// only TEXT/IMAGE/VIDEO), mirroring the Python SDK's separate SearchMediaType
// enum.
type SearchMediaType string

// Recognised keyword-search media-type filters.
const (
	SearchMediaTypeText  SearchMediaType = "TEXT"
	SearchMediaTypeImage SearchMediaType = "IMAGE"
	SearchMediaTypeVideo SearchMediaType = "VIDEO"
)

// SearchResult is the cursor-paginated result of a keyword search. Each element
// is a Post (a Threads media object).
type SearchResult struct {
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

// SearchParams configures Search.
type SearchParams struct {
	// Query is the search term. Prefix with "#" to search by topic tag.
	Query string
	// MediaType optionally filters results by media type.
	MediaType SearchMediaType
	// Fields are the thread fields to include. Defaults to the full media field
	// set when empty.
	Fields []string
	// Since / Until are Unix timestamps bounding the date range (optional).
	Since string
	Until string
}

// Search finds public threads matching a keyword or "#topic". Mirrors the Python
// SDK's SearchAPI.search, including the optional media-type filter and date
// bounds. Requires the threads_keyword_search scope.
func (c *Client) Search(ctx context.Context, p SearchParams) (*SearchResult, error) {
	if p.Query == "" {
		return nil, errors.New("threads: Search requires a non-empty Query")
	}
	q := url.Values{}
	q.Set("q", p.Query)
	q.Set("fields", strings.Join(mediaFieldsOrDefault(p.Fields), ","))
	if p.MediaType != "" {
		q.Set("media_type", string(p.MediaType))
	}
	if p.Since != "" {
		q.Set("since", p.Since)
	}
	if p.Until != "" {
		q.Set("until", p.Until)
	}

	out := &SearchResult{}
	if err := c.doGet(ctx, "keyword_search", q, out); err != nil {
		return nil, err
	}
	return out, nil
}
