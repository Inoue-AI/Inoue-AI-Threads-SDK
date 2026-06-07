package threads

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestCreateContainer_Text(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me/threads" {
			t.Errorf("path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("method: %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		form, _ := url.ParseQuery(string(body))
		if form.Get("media_type") != "TEXT" {
			t.Errorf("media_type: %s", form.Get("media_type"))
		}
		if form.Get("text") != "hello" {
			t.Errorf("text: %s", form.Get("text"))
		}
		_, _ = io.WriteString(w, `{"id":"CONTAINER1"}`)
	})
	out, err := c.CreateContainer(context.Background(), CreateContainerParams{
		MediaType: MediaTypeText,
		Text:      "hello",
	})
	if err != nil {
		t.Fatalf("CreateContainer: %v", err)
	}
	if out.ID != "CONTAINER1" {
		t.Fatalf("unexpected id: %s", out.ID)
	}
}

func TestCreateContainer_RequiresMediaType(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("server must not be called")
	})
	if _, err := c.CreateContainer(context.Background(), CreateContainerParams{Text: "x"}); err == nil {
		t.Fatal("expected error for missing MediaType")
	}
}

func TestPublish(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me/threads_publish" {
			t.Errorf("path: %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		form, _ := url.ParseQuery(string(body))
		if form.Get("creation_id") != "CONTAINER1" {
			t.Errorf("creation_id: %s", form.Get("creation_id"))
		}
		_, _ = io.WriteString(w, `{"id":"PUBLISHED1"}`)
	})
	out, err := c.Publish(context.Background(), "me", "CONTAINER1")
	if err != nil {
		t.Fatalf("Publish: %v", err)
	}
	if out.ID != "PUBLISHED1" {
		t.Fatalf("unexpected id: %s", out.ID)
	}
}

func TestPostText_FullFlow(t *testing.T) {
	calls := 0
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		switch r.URL.Path {
		case "/me/threads":
			_, _ = io.WriteString(w, `{"id":"C1"}`)
		case "/me/threads_publish":
			_, _ = io.WriteString(w, `{"id":"P1"}`)
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	})
	out, err := c.PostText(context.Background(), "me", "hi there")
	if err != nil {
		t.Fatalf("PostText: %v", err)
	}
	if out.ID != "P1" {
		t.Fatalf("unexpected id: %s", out.ID)
	}
	if calls != 2 {
		t.Fatalf("expected create+publish (2 calls), got %d", calls)
	}
}

func TestPostText_RequiresText(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("server must not be called")
	})
	if _, err := c.PostText(context.Background(), "me", ""); err == nil {
		t.Fatal("expected error for empty text")
	}
}

func TestDeletePost(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/thread_99" {
			t.Errorf("path: %s", r.URL.Path)
		}
		if r.Method != http.MethodDelete {
			t.Errorf("method: %s", r.Method)
		}
		_, _ = io.WriteString(w, `{"success":true}`)
	})
	ok, err := c.DeletePost(context.Background(), "thread_99")
	if err != nil {
		t.Fatalf("DeletePost: %v", err)
	}
	if !ok {
		t.Fatal("expected success=true")
	}
}

func TestCreateImagePost(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me/threads" {
			t.Errorf("path: %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		form, _ := url.ParseQuery(string(body))
		if form.Get("media_type") != "IMAGE" {
			t.Errorf("media_type: %s", form.Get("media_type"))
		}
		if form.Get("image_url") != "https://x/i.jpg" {
			t.Errorf("image_url: %s", form.Get("image_url"))
		}
		if form.Get("text") != "cap" {
			t.Errorf("text: %s", form.Get("text"))
		}
		_, _ = io.WriteString(w, `{"id":"container_2"}`)
	})
	out, err := c.CreateImagePost(context.Background(), CreateImagePostParams{
		ImageURL: "https://x/i.jpg",
		Text:     "cap",
	})
	if err != nil {
		t.Fatalf("CreateImagePost: %v", err)
	}
	if out.ID != "container_2" {
		t.Fatalf("unexpected id: %s", out.ID)
	}
}

func TestCreateImagePost_RequiresURL(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("server must not be called")
	})
	if _, err := c.CreateImagePost(context.Background(), CreateImagePostParams{}); err == nil {
		t.Fatal("expected error for missing ImageURL")
	}
}

func TestCreateVideoPost(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		form, _ := url.ParseQuery(string(body))
		if form.Get("media_type") != "VIDEO" {
			t.Errorf("media_type: %s", form.Get("media_type"))
		}
		if form.Get("video_url") != "https://x/v.mp4" {
			t.Errorf("video_url: %s", form.Get("video_url"))
		}
		_, _ = io.WriteString(w, `{"id":"container_3"}`)
	})
	out, err := c.CreateVideoPost(context.Background(), CreateVideoPostParams{VideoURL: "https://x/v.mp4"})
	if err != nil {
		t.Fatalf("CreateVideoPost: %v", err)
	}
	if out.ID != "container_3" {
		t.Fatalf("unexpected id: %s", out.ID)
	}
}

func TestCreateTextPost(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		form, _ := url.ParseQuery(string(body))
		if form.Get("media_type") != "TEXT" {
			t.Errorf("media_type: %s", form.Get("media_type"))
		}
		if form.Get("link_attachment") != "https://x/link" {
			t.Errorf("link_attachment: %s", form.Get("link_attachment"))
		}
		_, _ = io.WriteString(w, `{"id":"container_1"}`)
	})
	out, err := c.CreateTextPost(context.Background(), CreateTextPostParams{
		Text:           "Hello Threads!",
		LinkAttachment: "https://x/link",
	})
	if err != nil {
		t.Fatalf("CreateTextPost: %v", err)
	}
	if out.ID != "container_1" {
		t.Fatalf("unexpected id: %s", out.ID)
	}
}

func TestCreateCarouselPost(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		form, _ := url.ParseQuery(string(body))
		if form.Get("media_type") != "CAROUSEL" {
			t.Errorf("media_type: %s", form.Get("media_type"))
		}
		if form.Get("children") != "child_1,child_2,child_3" {
			t.Errorf("children: %s", form.Get("children"))
		}
		_, _ = io.WriteString(w, `{"id":"carousel_container"}`)
	})
	out, err := c.CreateCarouselPost(context.Background(), CreateCarouselPostParams{
		Children: []string{"child_1", "child_2", "child_3"},
	})
	if err != nil {
		t.Fatalf("CreateCarouselPost: %v", err)
	}
	if out.ID != "carousel_container" {
		t.Fatalf("unexpected id: %s", out.ID)
	}
}

func TestCreateCarouselPost_TooFew(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("server must not be called")
	})
	_, err := c.CreateCarouselPost(context.Background(), CreateCarouselPostParams{Children: []string{"one"}})
	if err == nil || !strings.Contains(err.Error(), "at least 2") {
		t.Fatalf("expected at-least-2 error, got %v", err)
	}
}

func TestCreateCarouselPost_TooMany(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("server must not be called")
	})
	children := make([]string, 21)
	for i := range children {
		children[i] = "i"
	}
	_, err := c.CreateCarouselPost(context.Background(), CreateCarouselPostParams{Children: children})
	if err == nil || !strings.Contains(err.Error(), "at most 20") {
		t.Fatalf("expected at-most-20 error, got %v", err)
	}
}

func TestGetContainerStatus(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/container_1" {
			t.Errorf("path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("fields"); got != "id,status,error_message" {
			t.Errorf("fields: %s", got)
		}
		_, _ = io.WriteString(w, `{"id":"container_1","status":"FINISHED"}`)
	})
	status, err := c.GetContainerStatus(context.Background(), "container_1")
	if err != nil {
		t.Fatalf("GetContainerStatus: %v", err)
	}
	if status.Status != ContainerStatusFinished {
		t.Fatalf("unexpected status: %s", status.Status)
	}
	if status.ErrorMessage != "" {
		t.Fatalf("expected empty error_message, got %q", status.ErrorMessage)
	}
}

func TestWaitForContainer_Finished(t *testing.T) {
	calls := 0
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			_, _ = io.WriteString(w, `{"id":"c1","status":"IN_PROGRESS"}`)
			return
		}
		_, _ = io.WriteString(w, `{"id":"c1","status":"FINISHED"}`)
	})
	status, err := c.WaitForContainer(context.Background(), "c1", WaitForContainerParams{
		PollInterval: 10 * time.Millisecond,
		Timeout:      2 * time.Second,
	})
	if err != nil {
		t.Fatalf("WaitForContainer: %v", err)
	}
	if status.Status != ContainerStatusFinished {
		t.Fatalf("unexpected status: %s", status.Status)
	}
	if calls < 2 {
		t.Fatalf("expected at least 2 polls, got %d", calls)
	}
}

func TestWaitForContainer_Errored(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"id":"c2","status":"ERRORED","error_message":"Bad format"}`)
	})
	_, err := c.WaitForContainer(context.Background(), "c2", WaitForContainerParams{
		PollInterval: 10 * time.Millisecond,
	})
	if err == nil || !strings.Contains(err.Error(), "Bad format") {
		t.Fatalf("expected errored container failure, got %v", err)
	}
}

func TestWaitForContainer_Timeout(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"id":"c3","status":"IN_PROGRESS"}`)
	})
	_, err := c.WaitForContainer(context.Background(), "c3", WaitForContainerParams{
		PollInterval: 50 * time.Millisecond,
		Timeout:      20 * time.Millisecond,
	})
	if err == nil || !strings.Contains(err.Error(), "did not finish") {
		t.Fatalf("expected timeout error, got %v", err)
	}
}

func TestWaitForContainer_CtxCancel(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"id":"c4","status":"IN_PROGRESS"}`)
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := c.WaitForContainer(ctx, "c4", WaitForContainerParams{
		PollInterval: 1 * time.Second,
		Timeout:      10 * time.Second,
	})
	if err == nil {
		t.Fatal("expected ctx cancellation error")
	}
}

func TestPostImage_FullFlow(t *testing.T) {
	var paths []string
	statusCalls := 0
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		switch {
		case r.URL.Path == "/me/threads" && r.Method == http.MethodPost:
			_, _ = io.WriteString(w, `{"id":"C_IMG"}`)
		case r.URL.Path == "/C_IMG":
			statusCalls++
			_, _ = io.WriteString(w, `{"id":"C_IMG","status":"FINISHED"}`)
		case r.URL.Path == "/me/threads_publish":
			_, _ = io.WriteString(w, `{"id":"PUB_IMG"}`)
		default:
			t.Errorf("unexpected path/method: %s %s", r.Method, r.URL.Path)
		}
	})
	out, err := c.PostImage(context.Background(), PostImageParams{
		ImageURL:     "https://x/i.jpg",
		PollInterval: 10 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("PostImage: %v", err)
	}
	if out.ID != "PUB_IMG" {
		t.Fatalf("unexpected id: %s", out.ID)
	}
	if statusCalls < 1 {
		t.Fatal("expected at least one status poll")
	}
}

func TestPostVideo_FullFlow(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/me/threads" && r.Method == http.MethodPost:
			_, _ = io.WriteString(w, `{"id":"C_VID"}`)
		case r.URL.Path == "/C_VID":
			_, _ = io.WriteString(w, `{"id":"C_VID","status":"FINISHED"}`)
		case r.URL.Path == "/me/threads_publish":
			_, _ = io.WriteString(w, `{"id":"PUB_VID"}`)
		default:
			t.Errorf("unexpected path/method: %s %s", r.Method, r.URL.Path)
		}
	})
	out, err := c.PostVideo(context.Background(), PostVideoParams{
		VideoURL:     "https://x/v.mp4",
		PollInterval: 10 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("PostVideo: %v", err)
	}
	if out.ID != "PUB_VID" {
		t.Fatalf("unexpected id: %s", out.ID)
	}
}

func TestPostCarousel_FullFlow(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/me/threads" && r.Method == http.MethodPost:
			body, _ := io.ReadAll(r.Body)
			form, _ := url.ParseQuery(string(body))
			if form.Get("media_type") != "CAROUSEL" {
				t.Errorf("media_type: %s", form.Get("media_type"))
			}
			_, _ = io.WriteString(w, `{"id":"C_CAR"}`)
		case r.URL.Path == "/C_CAR":
			_, _ = io.WriteString(w, `{"id":"C_CAR","status":"FINISHED"}`)
		case r.URL.Path == "/me/threads_publish":
			_, _ = io.WriteString(w, `{"id":"PUB_CAR"}`)
		default:
			t.Errorf("unexpected path/method: %s %s", r.Method, r.URL.Path)
		}
	})
	out, err := c.PostCarousel(context.Background(), PostCarouselParams{
		Children:     []string{"a", "b"},
		PollInterval: 10 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("PostCarousel: %v", err)
	}
	if out.ID != "PUB_CAR" {
		t.Fatalf("unexpected id: %s", out.ID)
	}
}

func TestGetPublishingLimit(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me/threads_publishing_limit" {
			t.Errorf("path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("fields"); !strings.Contains(got, "quota_usage") {
			t.Errorf("fields: %s", got)
		}
		_, _ = io.WriteString(w, `{"data":[{"quota_usage":5,"config":{"quota_total":250,"quota_duration":86400}}]}`)
	})
	limit, err := c.GetPublishingLimit(context.Background())
	if err != nil {
		t.Fatalf("GetPublishingLimit: %v", err)
	}
	if limit.QuotaUsage != 5 {
		t.Fatalf("quota_usage: %d", limit.QuotaUsage)
	}
	if limit.Config == nil || limit.Config.QuotaTotal != 250 {
		t.Fatalf("unexpected config: %+v", limit.Config)
	}
}

func TestGetPublishingLimit_Empty(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"data":[]}`)
	})
	limit, err := c.GetPublishingLimit(context.Background())
	if err != nil {
		t.Fatalf("GetPublishingLimit: %v", err)
	}
	if limit.QuotaUsage != 0 || limit.Config != nil {
		t.Fatalf("expected zero-valued limit, got %+v", limit)
	}
}
