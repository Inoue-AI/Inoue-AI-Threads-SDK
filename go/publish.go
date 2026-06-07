package threads

import (
	"context"
	"errors"
	"net/url"
	"strings"
)

// MediaType is the container type used when creating a post.
type MediaType string

// Recognised container media types.
const (
	MediaTypeText     MediaType = "TEXT"
	MediaTypeImage    MediaType = "IMAGE"
	MediaTypeVideo    MediaType = "VIDEO"
	MediaTypeCarousel MediaType = "CAROUSEL"
)

// ReplyControl controls who may reply to a thread.
type ReplyControl string

// Recognised reply-control values.
const (
	ReplyControlEveryone       ReplyControl = "everyone"
	ReplyControlAccountsFollow ReplyControl = "accounts_you_follow"
	ReplyControlMentionedOnly  ReplyControl = "mentioned_only"
)

// PublishResult is returned by container creation and publish. Both
// endpoints return a single object id.
type PublishResult struct {
	ID string `json:"id"`
}

// CreateContainerParams configures CreateContainer. Set only the fields that
// apply to the chosen MediaType (e.g. Text for TEXT, ImageURL for IMAGE).
type CreateContainerParams struct {
	UserID         string // Defaults to "me".
	MediaType      MediaType
	Text           string
	ImageURL       string
	VideoURL       string
	Children       []string // Carousel item container ids (2-20).
	IsCarouselItem bool
	ReplyToID      string
	ReplyControl   ReplyControl
	LinkAttachment string
}

// CreateContainer creates a media container (the first step of publishing).
// Mirrors the Python SDK's PublishingAPI.create_container. The returned id is
// passed to Publish once any media processing has finished.
func (c *Client) CreateContainer(ctx context.Context, p CreateContainerParams) (*PublishResult, error) {
	if p.MediaType == "" {
		return nil, errors.New("threads: CreateContainer requires MediaType")
	}
	userID := p.UserID
	if userID == "" {
		userID = "me"
	}
	form := url.Values{}
	form.Set("media_type", string(p.MediaType))
	if p.Text != "" {
		form.Set("text", p.Text)
	}
	if p.ImageURL != "" {
		form.Set("image_url", p.ImageURL)
	}
	if p.VideoURL != "" {
		form.Set("video_url", p.VideoURL)
	}
	if len(p.Children) > 0 {
		form.Set("children", strings.Join(p.Children, ","))
	}
	if p.IsCarouselItem {
		form.Set("is_carousel_item", "true")
	}
	if p.ReplyToID != "" {
		form.Set("reply_to_id", p.ReplyToID)
	}
	if p.ReplyControl != "" {
		form.Set("reply_control", string(p.ReplyControl))
	}
	if p.LinkAttachment != "" {
		form.Set("link_attachment", p.LinkAttachment)
	}

	out := &PublishResult{}
	if err := c.doPostForm(ctx, userID+"/threads", form, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Publish publishes a previously created container. Mirrors the Python SDK's
// PublishingAPI.publish.
func (c *Client) Publish(ctx context.Context, userID, creationID string) (*PublishResult, error) {
	if creationID == "" {
		return nil, errors.New("threads: Publish requires a creationID")
	}
	if userID == "" {
		userID = "me"
	}
	form := url.Values{}
	form.Set("creation_id", creationID)

	out := &PublishResult{}
	if err := c.doPostForm(ctx, userID+"/threads_publish", form, out); err != nil {
		return nil, err
	}
	return out, nil
}

// PostText is a convenience that creates a text container and immediately
// publishes it (text posts require no media processing). Mirrors the Python
// SDK's PublishingAPI.post_text.
func (c *Client) PostText(ctx context.Context, userID, text string) (*PublishResult, error) {
	if text == "" {
		return nil, errors.New("threads: PostText requires non-empty text")
	}
	container, err := c.CreateContainer(ctx, CreateContainerParams{
		UserID:    userID,
		MediaType: MediaTypeText,
		Text:      text,
	})
	if err != nil {
		return nil, err
	}
	return c.Publish(ctx, userID, container.ID)
}

// deleteResult decodes the {"success": bool} body returned by DELETE.
type deleteResult struct {
	Success bool `json:"success"`
}

// DeletePost deletes a published thread. Mirrors the Python SDK's
// PublishingAPI.delete_post. Requires the threads_delete scope.
func (c *Client) DeletePost(ctx context.Context, threadID string) (bool, error) {
	if threadID == "" {
		return false, errors.New("threads: DeletePost requires a threadID")
	}
	out := &deleteResult{}
	if err := c.doDelete(ctx, threadID, out); err != nil {
		return false, err
	}
	return out.Success, nil
}
