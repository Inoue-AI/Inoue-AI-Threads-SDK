package threads

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestGetMediaInsights(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/post_1/insights" {
			t.Errorf("path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("metric"); !strings.Contains(got, "views") {
			t.Errorf("metric: %s", got)
		}
		_, _ = io.WriteString(w, `{"data":[
			{"name":"views","period":"lifetime","values":[{"value":1500}],"title":"Views","id":"post_1/insights/views/lifetime"},
			{"name":"likes","period":"lifetime","values":[{"value":42}],"title":"Likes","id":"post_1/insights/likes/lifetime"}
		]}`)
	})
	resp, err := c.GetMediaInsights(context.Background(), "post_1", nil)
	if err != nil {
		t.Fatalf("GetMediaInsights: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 metrics, got %d", len(resp.Data))
	}
	if resp.Data[0].Name != "views" {
		t.Fatalf("name: %s", resp.Data[0].Name)
	}
	if v := resp.Data[0].Values[0]; !v.IsInt || v.Int != 1500 {
		t.Fatalf("unexpected views value: %+v", v)
	}
	if v := resp.Data[1].Values[0]; v.Int != 42 {
		t.Fatalf("unexpected likes value: %+v", v)
	}
}

func TestGetMediaInsights_RequiresID(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("server must not be called")
	})
	if _, err := c.GetMediaInsights(context.Background(), "", nil); err == nil {
		t.Fatal("expected error for empty mediaID")
	}
}

func TestGetMediaInsights_SpecificMetric(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("metric"); got != "likes" {
			t.Errorf("metric: %s", got)
		}
		_, _ = io.WriteString(w, `{"data":[{"name":"likes","period":"lifetime","values":[{"value":99}]}]}`)
	})
	resp, err := c.GetMediaInsights(context.Background(), "post_2", []MediaMetric{MediaMetricLikes})
	if err != nil {
		t.Fatalf("GetMediaInsights: %v", err)
	}
	if resp.Data[0].Values[0].Int != 99 {
		t.Fatalf("unexpected value: %+v", resp.Data[0].Values[0])
	}
}

func TestGetAccountInsights(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me/threads_insights" {
			t.Errorf("path: %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("period") != "day" {
			t.Errorf("period: %s", q.Get("period"))
		}
		if q.Get("since") != "2024-06-01" || q.Get("until") != "2024-06-03" {
			t.Errorf("since/until: %s %s", q.Get("since"), q.Get("until"))
		}
		// Default metric set is the five core engagement metrics.
		if got := q.Get("metric"); !strings.Contains(got, "views") || !strings.Contains(got, "quotes") {
			t.Errorf("metric: %s", got)
		}
		_, _ = io.WriteString(w, `{"data":[{"name":"views","period":"day","values":[
			{"value":200,"end_time":"2024-06-01T07:00:00+0000"},
			{"value":350,"end_time":"2024-06-02T07:00:00+0000"}
		]}]}`)
	})
	resp, err := c.GetAccountInsights(context.Background(), AccountInsightsParams{
		Period: InsightPeriodDay,
		Since:  "2024-06-01",
		Until:  "2024-06-03",
	})
	if err != nil {
		t.Fatalf("GetAccountInsights: %v", err)
	}
	if len(resp.Data) != 1 || len(resp.Data[0].Values) != 2 {
		t.Fatalf("unexpected data: %+v", resp.Data)
	}
	if resp.Data[0].Values[0].EndTime != "2024-06-01T07:00:00+0000" {
		t.Fatalf("end_time: %s", resp.Data[0].Values[0].EndTime)
	}
}

func TestGetAccountInsights_DefaultPeriod(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("period"); got != "day" {
			t.Errorf("expected default period day, got %s", got)
		}
		_, _ = io.WriteString(w, `{"data":[]}`)
	})
	if _, err := c.GetAccountInsights(context.Background(), AccountInsightsParams{}); err != nil {
		t.Fatalf("GetAccountInsights: %v", err)
	}
}

func TestGetAccountInsights_FollowerDemographics(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("metric") != "follower_demographics" {
			t.Errorf("metric: %s", q.Get("metric"))
		}
		if q.Get("breakdown") != "country" {
			t.Errorf("breakdown: %s", q.Get("breakdown"))
		}
		_, _ = io.WriteString(w, `{"data":[{"name":"follower_demographics","period":"lifetime","values":[
			{"value":{"US":500,"UK":200,"JP":150}}
		]}]}`)
	})
	resp, err := c.GetAccountInsights(context.Background(), AccountInsightsParams{
		Metrics:   []AccountMetric{AccountMetricFollowerDemographics},
		Period:    InsightPeriodLifetime,
		Breakdown: DemographicBreakdownCountry,
	})
	if err != nil {
		t.Fatalf("GetAccountInsights: %v", err)
	}
	val := resp.Data[0].Values[0]
	if val.IsInt {
		t.Fatalf("expected demographic breakdown, got int form: %+v", val)
	}
	if val.Breakdown["US"] != 500 || val.Breakdown["UK"] != 200 {
		t.Fatalf("unexpected breakdown: %+v", val.Breakdown)
	}
}

func TestGetMetric(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("metric"); got != "likes" {
			t.Errorf("metric: %s", got)
		}
		_, _ = io.WriteString(w, `{"data":[{"name":"likes","period":"lifetime","values":[{"value":99}]}]}`)
	})
	metric, err := c.GetMetric(context.Background(), "post_2", MediaMetricLikes)
	if err != nil {
		t.Fatalf("GetMetric: %v", err)
	}
	if metric == nil || metric.Name != "likes" {
		t.Fatalf("unexpected metric: %+v", metric)
	}
	if metric.Values[0].Int != 99 {
		t.Fatalf("unexpected value: %+v", metric.Values[0])
	}
}

func TestGetMetric_Empty(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"data":[]}`)
	})
	metric, err := c.GetMetric(context.Background(), "post_3", MediaMetricShares)
	if err != nil {
		t.Fatalf("GetMetric: %v", err)
	}
	if metric != nil {
		t.Fatalf("expected nil metric, got %+v", metric)
	}
}
