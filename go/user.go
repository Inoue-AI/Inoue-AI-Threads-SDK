package threads

import (
	"context"
	"net/url"
	"strings"
)

// User mirrors the subset of Threads user fields the Inoue backend consumes.
// Fields are populated only when requested via the `fields` argument.
type User struct {
	ID                    string `json:"id,omitempty"`
	Username              string `json:"username,omitempty"`
	Name                  string `json:"name,omitempty"`
	ThreadsProfilePicture string `json:"threads_profile_picture_url,omitempty"`
	ThreadsBiography      string `json:"threads_biography,omitempty"`
}

// GetUser fetches the profile for the given user (or "me" for the
// authenticated user). When fields is empty a sensible default set is used.
func (c *Client) GetUser(ctx context.Context, userID string, fields []string) (*User, error) {
	if userID == "" {
		userID = "me"
	}
	if len(fields) == 0 {
		fields = []string{
			"id",
			"username",
			"name",
			"threads_profile_picture_url",
			"threads_biography",
		}
	}
	q := url.Values{}
	q.Set("fields", strings.Join(fields, ","))

	out := &User{}
	if err := c.doGet(ctx, userID, q, out); err != nil {
		return nil, err
	}
	return out, nil
}
