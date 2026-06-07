package threads

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
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

// Carousel item-count bounds, mirroring the Python SDK config
// (CAROUSEL_MIN_ITEMS / CAROUSEL_MAX_ITEMS).
const (
	carouselMinItems = 2
	carouselMaxItems = 20
)

// Container-polling defaults, mirroring the Python SDK config
// (DEFAULT_POLL_INTERVAL / DEFAULT_POLL_TIMEOUT).
const (
	// DefaultPollInterval is the delay between container-status polls used by
	// WaitForContainer when CreatePollParams.PollInterval is zero.
	DefaultPollInterval = 5 * time.Second
	// DefaultPollTimeout is the maximum time WaitForContainer waits for a
	// container to reach a terminal status when CreatePollParams.Timeout is zero.
	DefaultPollTimeout = 300 * time.Second
)

// ContainerStatus is the processing status of a media container.
type ContainerStatus string

// Recognised container processing statuses.
const (
	ContainerStatusInProgress ContainerStatus = "IN_PROGRESS"
	ContainerStatusFinished   ContainerStatus = "FINISHED"
	ContainerStatusErrored    ContainerStatus = "ERRORED"
	ContainerStatusExpired    ContainerStatus = "EXPIRED"
	ContainerStatusPublished  ContainerStatus = "PUBLISHED"
)

// isTerminal reports whether the status will not change with further polling.
func (s ContainerStatus) isTerminal() bool {
	switch s {
	case ContainerStatusFinished, ContainerStatusErrored, ContainerStatusExpired, ContainerStatusPublished:
		return true
	default:
		return false
	}
}

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

// CreateTextPostParams configures CreateTextPost.
type CreateTextPostParams struct {
	UserID         string // Defaults to "me".
	Text           string // Required post body.
	ReplyToID      string // Optional id of the thread being replied to.
	ReplyControl   ReplyControl
	LinkAttachment string // Optional link-preview URL.
}

// CreateTextPost creates a text-only container without publishing it. Mirrors
// the Python SDK's PublishingAPI.create_text_post.
func (c *Client) CreateTextPost(ctx context.Context, p CreateTextPostParams) (*PublishResult, error) {
	return c.CreateContainer(ctx, CreateContainerParams{
		UserID:         p.UserID,
		MediaType:      MediaTypeText,
		Text:           p.Text,
		ReplyToID:      p.ReplyToID,
		ReplyControl:   p.ReplyControl,
		LinkAttachment: p.LinkAttachment,
	})
}

// CreateImagePostParams configures CreateImagePost.
type CreateImagePostParams struct {
	UserID         string // Defaults to "me".
	ImageURL       string // Required, publicly reachable image URL.
	Text           string // Optional caption.
	ReplyToID      string
	ReplyControl   ReplyControl
	IsCarouselItem bool // Set true when building a carousel child.
}

// CreateImagePost creates an image container without publishing it. Mirrors the
// Python SDK's PublishingAPI.create_image_post.
func (c *Client) CreateImagePost(ctx context.Context, p CreateImagePostParams) (*PublishResult, error) {
	if p.ImageURL == "" {
		return nil, errors.New("threads: CreateImagePost requires ImageURL")
	}
	return c.CreateContainer(ctx, CreateContainerParams{
		UserID:         p.UserID,
		MediaType:      MediaTypeImage,
		ImageURL:       p.ImageURL,
		Text:           p.Text,
		ReplyToID:      p.ReplyToID,
		ReplyControl:   p.ReplyControl,
		IsCarouselItem: p.IsCarouselItem,
	})
}

// CreateVideoPostParams configures CreateVideoPost.
type CreateVideoPostParams struct {
	UserID         string // Defaults to "me".
	VideoURL       string // Required, publicly reachable video URL.
	Text           string // Optional caption.
	ReplyToID      string
	ReplyControl   ReplyControl
	IsCarouselItem bool // Set true when building a carousel child.
}

// CreateVideoPost creates a video container without publishing it. Mirrors the
// Python SDK's PublishingAPI.create_video_post.
func (c *Client) CreateVideoPost(ctx context.Context, p CreateVideoPostParams) (*PublishResult, error) {
	if p.VideoURL == "" {
		return nil, errors.New("threads: CreateVideoPost requires VideoURL")
	}
	return c.CreateContainer(ctx, CreateContainerParams{
		UserID:         p.UserID,
		MediaType:      MediaTypeVideo,
		VideoURL:       p.VideoURL,
		Text:           p.Text,
		ReplyToID:      p.ReplyToID,
		ReplyControl:   p.ReplyControl,
		IsCarouselItem: p.IsCarouselItem,
	})
}

// CreateCarouselPostParams configures CreateCarouselPost.
type CreateCarouselPostParams struct {
	UserID       string   // Defaults to "me".
	Children     []string // 2-20 item container ids (each is_carousel_item=true).
	Text         string   // Optional caption.
	ReplyToID    string
	ReplyControl ReplyControl
}

// CreateCarouselPost creates a carousel container from previously created item
// containers, without publishing it. Mirrors the Python SDK's
// PublishingAPI.create_carousel_post, including the 2-20 item-count validation.
func (c *Client) CreateCarouselPost(ctx context.Context, p CreateCarouselPostParams) (*PublishResult, error) {
	if len(p.Children) < carouselMinItems {
		return nil, fmt.Errorf("threads: a carousel requires at least %d items", carouselMinItems)
	}
	if len(p.Children) > carouselMaxItems {
		return nil, fmt.Errorf("threads: a carousel allows at most %d items", carouselMaxItems)
	}
	return c.CreateContainer(ctx, CreateContainerParams{
		UserID:       p.UserID,
		MediaType:    MediaTypeCarousel,
		Children:     p.Children,
		Text:         p.Text,
		ReplyToID:    p.ReplyToID,
		ReplyControl: p.ReplyControl,
	})
}

// ContainerStatusResult is the processing status of a media container, as
// returned by GET /{container-id}?fields=id,status,error_message.
type ContainerStatusResult struct {
	ID           string          `json:"id"`
	Status       ContainerStatus `json:"status,omitempty"`
	ErrorMessage string          `json:"error_message,omitempty"`
}

// GetContainerStatus checks the processing status of a media container. Mirrors
// the Python SDK's PublishingAPI.get_container_status.
func (c *Client) GetContainerStatus(ctx context.Context, containerID string) (*ContainerStatusResult, error) {
	if containerID == "" {
		return nil, errors.New("threads: GetContainerStatus requires a containerID")
	}
	q := url.Values{}
	q.Set("fields", "id,status,error_message")

	out := &ContainerStatusResult{}
	if err := c.doGet(ctx, containerID, q, out); err != nil {
		return nil, err
	}
	return out, nil
}

// WaitForContainerParams tunes WaitForContainer's polling behaviour. Zero
// values fall back to DefaultPollInterval / DefaultPollTimeout, matching the
// Python SDK defaults (5s interval, 300s timeout).
type WaitForContainerParams struct {
	PollInterval time.Duration
	Timeout      time.Duration
}

// WaitForContainer polls a container until it reaches a terminal status,
// mirroring the Python SDK's PublishingAPI.wait_for_container. It returns an
// error if the container errors out or the timeout is exceeded. Polling honours
// ctx cancellation between attempts (no unconditional sleep).
func (c *Client) WaitForContainer(ctx context.Context, containerID string, p WaitForContainerParams) (*ContainerStatusResult, error) {
	if containerID == "" {
		return nil, errors.New("threads: WaitForContainer requires a containerID")
	}
	interval := p.PollInterval
	if interval <= 0 {
		interval = DefaultPollInterval
	}
	timeout := p.Timeout
	if timeout <= 0 {
		timeout = DefaultPollTimeout
	}

	deadline := time.Now().Add(timeout)
	for {
		status, err := c.GetContainerStatus(ctx, containerID)
		if err != nil {
			return nil, err
		}
		if status.Status.isTerminal() {
			if status.Status == ContainerStatusErrored {
				return nil, fmt.Errorf("threads: container %s failed: %s", containerID, status.ErrorMessage)
			}
			return status, nil
		}
		if time.Now().Add(interval).After(deadline) {
			return nil, fmt.Errorf("threads: container %s did not finish within %s", containerID, timeout)
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(interval):
		}
	}
}

// PostImageParams configures PostImage.
type PostImageParams struct {
	UserID       string // Defaults to "me".
	ImageURL     string // Required, publicly reachable image URL.
	Text         string // Optional caption.
	ReplyToID    string
	ReplyControl ReplyControl
	// PollInterval / Timeout tune the processing wait. Zero uses the defaults.
	PollInterval time.Duration
	Timeout      time.Duration
}

// PostImage creates an image container, waits for processing to finish, and
// publishes it. Mirrors the Python SDK's PublishingAPI.post_image.
func (c *Client) PostImage(ctx context.Context, p PostImageParams) (*PublishResult, error) {
	container, err := c.CreateImagePost(ctx, CreateImagePostParams{
		UserID:       p.UserID,
		ImageURL:     p.ImageURL,
		Text:         p.Text,
		ReplyToID:    p.ReplyToID,
		ReplyControl: p.ReplyControl,
	})
	if err != nil {
		return nil, err
	}
	if _, err := c.WaitForContainer(ctx, container.ID, WaitForContainerParams{
		PollInterval: p.PollInterval,
		Timeout:      p.Timeout,
	}); err != nil {
		return nil, err
	}
	return c.Publish(ctx, p.UserID, container.ID)
}

// PostVideoParams configures PostVideo.
type PostVideoParams struct {
	UserID       string // Defaults to "me".
	VideoURL     string // Required, publicly reachable video URL.
	Text         string // Optional caption.
	ReplyToID    string
	ReplyControl ReplyControl
	// PollInterval / Timeout tune the processing wait. Zero uses the defaults.
	PollInterval time.Duration
	Timeout      time.Duration
}

// PostVideo creates a video container, waits for processing to finish, and
// publishes it. Mirrors the Python SDK's PublishingAPI.post_video.
func (c *Client) PostVideo(ctx context.Context, p PostVideoParams) (*PublishResult, error) {
	container, err := c.CreateVideoPost(ctx, CreateVideoPostParams{
		UserID:       p.UserID,
		VideoURL:     p.VideoURL,
		Text:         p.Text,
		ReplyToID:    p.ReplyToID,
		ReplyControl: p.ReplyControl,
	})
	if err != nil {
		return nil, err
	}
	if _, err := c.WaitForContainer(ctx, container.ID, WaitForContainerParams{
		PollInterval: p.PollInterval,
		Timeout:      p.Timeout,
	}); err != nil {
		return nil, err
	}
	return c.Publish(ctx, p.UserID, container.ID)
}

// PostCarouselParams configures PostCarousel.
type PostCarouselParams struct {
	UserID       string   // Defaults to "me".
	Children     []string // 2-20 item container ids (each is_carousel_item=true).
	Text         string   // Optional caption.
	ReplyToID    string
	ReplyControl ReplyControl
	// PollInterval / Timeout tune the processing wait. Zero uses the defaults.
	PollInterval time.Duration
	Timeout      time.Duration
}

// PostCarousel creates a carousel container from previously created item
// containers, waits for it to finish processing, and publishes it. Mirrors the
// Python SDK's PublishingAPI.post_carousel.
func (c *Client) PostCarousel(ctx context.Context, p PostCarouselParams) (*PublishResult, error) {
	container, err := c.CreateCarouselPost(ctx, CreateCarouselPostParams{
		UserID:       p.UserID,
		Children:     p.Children,
		Text:         p.Text,
		ReplyToID:    p.ReplyToID,
		ReplyControl: p.ReplyControl,
	})
	if err != nil {
		return nil, err
	}
	if _, err := c.WaitForContainer(ctx, container.ID, WaitForContainerParams{
		PollInterval: p.PollInterval,
		Timeout:      p.Timeout,
	}); err != nil {
		return nil, err
	}
	return c.Publish(ctx, p.UserID, container.ID)
}

// PublishingLimitConfig is the rate-limit configuration block returned inside a
// PublishingLimit.
type PublishingLimitConfig struct {
	QuotaTotal    int `json:"quota_total,omitempty"`
	QuotaDuration int `json:"quota_duration,omitempty"`
}

// PublishingLimit is the current publishing quota usage and configuration, as
// returned by GET /me/threads_publishing_limit.
type PublishingLimit struct {
	QuotaUsage      int                    `json:"quota_usage,omitempty"`
	Config          *PublishingLimitConfig `json:"config,omitempty"`
	ReplyQuotaUsage int                    `json:"reply_quota_usage,omitempty"`
	ReplyConfig     *PublishingLimitConfig `json:"reply_config,omitempty"`
}

// publishingLimitResult wraps the {"data":[...]} envelope the limit endpoint
// returns.
type publishingLimitResult struct {
	Data []PublishingLimit `json:"data"`
}

// GetPublishingLimit checks the current publishing rate-limit quota. Mirrors the
// Python SDK's PublishingAPI.get_publishing_limit. When the API returns an empty
// data array a zero-valued PublishingLimit is returned (matching Python).
func (c *Client) GetPublishingLimit(ctx context.Context) (*PublishingLimit, error) {
	q := url.Values{}
	q.Set("fields", "quota_usage,config,reply_quota_usage,reply_config")

	out := &publishingLimitResult{}
	if err := c.doGet(ctx, "me/threads_publishing_limit", q, out); err != nil {
		return nil, err
	}
	if len(out.Data) > 0 {
		limit := out.Data[0]
		return &limit, nil
	}
	return &PublishingLimit{}, nil
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
