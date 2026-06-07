package threads

import (
	"context"
	"errors"
	"net/url"
	"strings"
)

// defaultMediaFields is the field set requested for thread/media objects when a
// caller passes no explicit fields. It mirrors the Python SDK's list(MediaField)
// default used across the replies, mentions, and search endpoints.
var defaultMediaFields = []string{
	"id", "media_product_type", "media_type", "media_url",
	"permalink", "owner", "username", "text", "timestamp",
	"shortcode", "thumbnail_url", "children", "is_quote_post",
}

// ReplyListResult is the cursor-paginated wrapper for replies, conversations,
// and mentions. Each element is a Post (a Threads media object).
type ReplyListResult struct {
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

// mediaFieldsOrDefault returns the supplied fields, or the default media field
// set when none were provided.
func mediaFieldsOrDefault(fields []string) []string {
	if len(fields) == 0 {
		return defaultMediaFields
	}
	return fields
}

// GetReplies returns the top-level replies to a thread. When reverse is true the
// replies are returned in reverse-chronological order. Mirrors the Python SDK's
// RepliesAPI.get_replies.
func (c *Client) GetReplies(ctx context.Context, mediaID string, fields []string, reverse bool) (*ReplyListResult, error) {
	if mediaID == "" {
		return nil, errors.New("threads: GetReplies requires a mediaID")
	}
	q := url.Values{}
	q.Set("fields", strings.Join(mediaFieldsOrDefault(fields), ","))
	if reverse {
		q.Set("reverse", "true")
	}

	out := &ReplyListResult{}
	if err := c.doGet(ctx, mediaID+"/replies", q, out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetConversation returns the full conversation tree (nested replies) for a
// thread. When reverse is true the replies are returned in reverse-chronological
// order. Mirrors the Python SDK's RepliesAPI.get_conversation.
func (c *Client) GetConversation(ctx context.Context, mediaID string, fields []string, reverse bool) (*ReplyListResult, error) {
	if mediaID == "" {
		return nil, errors.New("threads: GetConversation requires a mediaID")
	}
	q := url.Values{}
	q.Set("fields", strings.Join(mediaFieldsOrDefault(fields), ","))
	if reverse {
		q.Set("reverse", "true")
	}

	out := &ReplyListResult{}
	if err := c.doGet(ctx, mediaID+"/conversation", q, out); err != nil {
		return nil, err
	}
	return out, nil
}

// HideReply hides a reply so it is no longer publicly visible. Mirrors the
// Python SDK's RepliesAPI.hide_reply. Returns the API's reported success flag.
func (c *Client) HideReply(ctx context.Context, replyID string) (bool, error) {
	return c.manageReply(ctx, replyID, true)
}

// UnhideReply unhides a previously hidden reply. Mirrors the Python SDK's
// RepliesAPI.unhide_reply. Returns the API's reported success flag.
func (c *Client) UnhideReply(ctx context.Context, replyID string) (bool, error) {
	return c.manageReply(ctx, replyID, false)
}

// manageReply is the shared POST /{reply-id}/manage_reply implementation for
// HideReply / UnhideReply.
func (c *Client) manageReply(ctx context.Context, replyID string, hide bool) (bool, error) {
	if replyID == "" {
		return false, errors.New("threads: manage reply requires a replyID")
	}
	form := url.Values{}
	if hide {
		form.Set("hide", "true")
	} else {
		form.Set("hide", "false")
	}

	out := &deleteResult{}
	if err := c.doPostForm(ctx, replyID+"/manage_reply", form, out); err != nil {
		return false, err
	}
	return out.Success, nil
}

// GetMentions returns threads that mention the authenticated user. Mirrors the
// Python SDK's RepliesAPI.get_mentions.
func (c *Client) GetMentions(ctx context.Context, fields []string) (*ReplyListResult, error) {
	q := url.Values{}
	q.Set("fields", strings.Join(mediaFieldsOrDefault(fields), ","))

	out := &ReplyListResult{}
	if err := c.doGet(ctx, "me/mentions", q, out); err != nil {
		return nil, err
	}
	return out, nil
}
