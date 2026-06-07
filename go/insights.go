package threads

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
)

// MediaMetric is a metric available for an individual media post.
type MediaMetric string

// Recognised media-level metrics.
const (
	MediaMetricViews   MediaMetric = "views"
	MediaMetricLikes   MediaMetric = "likes"
	MediaMetricReplies MediaMetric = "replies"
	MediaMetricReposts MediaMetric = "reposts"
	MediaMetricQuotes  MediaMetric = "quotes"
	MediaMetricShares  MediaMetric = "shares"
)

// AccountMetric is a metric available for account-level insights.
type AccountMetric string

// Recognised account-level metrics.
const (
	AccountMetricViews                AccountMetric = "views"
	AccountMetricLikes                AccountMetric = "likes"
	AccountMetricReplies              AccountMetric = "replies"
	AccountMetricReposts              AccountMetric = "reposts"
	AccountMetricQuotes               AccountMetric = "quotes"
	AccountMetricShares               AccountMetric = "shares"
	AccountMetricFollowersCount       AccountMetric = "followers_count"
	AccountMetricFollowerDemographics AccountMetric = "follower_demographics"
)

// InsightPeriod is the aggregation period for account-level insights.
type InsightPeriod string

// Recognised aggregation periods.
const (
	InsightPeriodDay      InsightPeriod = "day"
	InsightPeriodWeek     InsightPeriod = "week"
	InsightPeriodDays28   InsightPeriod = "days_28"
	InsightPeriodLifetime InsightPeriod = "lifetime"
)

// DemographicBreakdown is a dimension for follower demographic breakdowns.
type DemographicBreakdown string

// Recognised demographic breakdown dimensions.
const (
	DemographicBreakdownCountry DemographicBreakdown = "country"
	DemographicBreakdownCity    DemographicBreakdown = "city"
	DemographicBreakdownAge     DemographicBreakdown = "age"
	DemographicBreakdownGender  DemographicBreakdown = "gender"
)

// defaultAccountMetrics is the five core engagement metrics requested when no
// explicit metric set is supplied, mirroring the Python SDK default.
var defaultAccountMetrics = []AccountMetric{
	AccountMetricViews,
	AccountMetricLikes,
	AccountMetricReplies,
	AccountMetricReposts,
	AccountMetricQuotes,
}

// allMediaMetrics enumerates every media metric, requested when no explicit
// metric set is supplied to GetMediaInsights, mirroring the Python SDK default
// (list(MediaMetric)).
var allMediaMetrics = []MediaMetric{
	MediaMetricViews,
	MediaMetricLikes,
	MediaMetricReplies,
	MediaMetricReposts,
	MediaMetricQuotes,
	MediaMetricShares,
}

// InsightValue is a single insight data-point. The Graph API encodes "value" as
// either a scalar integer (most metrics) or an object keyed by demographic
// dimension (follower_demographics). Both forms are decoded here without
// resorting to an opaque map: Int holds the scalar form and Breakdown holds the
// demographic form. IsInt reports which form was present.
type InsightValue struct {
	// Int is the scalar value, populated when the API returned an integer.
	Int int64
	// Breakdown is the demographic map (e.g. {"US": 500}), populated when the
	// API returned an object.
	Breakdown map[string]int64
	// IsInt reports whether the scalar (Int) form was present rather than the
	// Breakdown form.
	IsInt bool
	// EndTime is the ISO-8601 timestamp bounding this data-point, when present.
	EndTime string
}

// insightValueWire is the on-the-wire shape used to decode an InsightValue.
type insightValueWire struct {
	Value   json.RawMessage `json:"value"`
	EndTime string          `json:"end_time"`
}

// UnmarshalJSON decodes the int-or-object "value" union returned by the Graph
// API. A missing or null value leaves both Int and Breakdown at their zero
// values with IsInt false.
func (v *InsightValue) UnmarshalJSON(data []byte) error {
	var wire insightValueWire
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}
	v.EndTime = wire.EndTime
	v.Int = 0
	v.Breakdown = nil
	v.IsInt = false

	raw := strings.TrimSpace(string(wire.Value))
	if raw == "" || raw == "null" {
		return nil
	}
	if raw[0] == '{' {
		breakdown := make(map[string]int64)
		if err := json.Unmarshal(wire.Value, &breakdown); err != nil {
			return err
		}
		v.Breakdown = breakdown
		return nil
	}
	var n int64
	if err := json.Unmarshal(wire.Value, &n); err != nil {
		return err
	}
	v.Int = n
	v.IsInt = true
	return nil
}

// InsightData is a single insight metric together with its values.
type InsightData struct {
	Name        string         `json:"name"`
	Period      string         `json:"period"`
	Values      []InsightValue `json:"values"`
	Title       string         `json:"title,omitempty"`
	Description string         `json:"description,omitempty"`
	ID          string         `json:"id,omitempty"`
	TotalValue  *InsightValue  `json:"total_value,omitempty"`
}

// InsightsResult is the response from an insights endpoint.
type InsightsResult struct {
	Data []InsightData `json:"data"`
}

// GetMediaInsights retrieves insights for a specific thread / media object.
// When metrics is empty, all available media metrics are requested. Mirrors the
// Python SDK's InsightsAPI.get_media_insights.
func (c *Client) GetMediaInsights(ctx context.Context, mediaID string, metrics []MediaMetric) (*InsightsResult, error) {
	if mediaID == "" {
		return nil, errors.New("threads: GetMediaInsights requires a mediaID")
	}
	if len(metrics) == 0 {
		metrics = allMediaMetrics
	}
	names := make([]string, len(metrics))
	for i, m := range metrics {
		names[i] = string(m)
	}
	q := url.Values{}
	q.Set("metric", strings.Join(names, ","))

	out := &InsightsResult{}
	if err := c.doGet(ctx, mediaID+"/insights", q, out); err != nil {
		return nil, err
	}
	return out, nil
}

// AccountInsightsParams configures GetAccountInsights.
type AccountInsightsParams struct {
	// Metrics to fetch. Defaults to the five core engagement metrics
	// (views, likes, replies, reposts, quotes) when empty.
	Metrics []AccountMetric
	// Period is the aggregation window. Defaults to InsightPeriodDay when empty.
	Period InsightPeriod
	// Since / Until are ISO-8601 date strings bounding the range (optional).
	Since string
	Until string
	// Breakdown is a demographic dimension (optional; requires >=100 followers).
	Breakdown DemographicBreakdown
}

// GetAccountInsights retrieves account-level insights with an optional date
// range and demographic breakdown. Mirrors the Python SDK's
// InsightsAPI.get_account_insights.
func (c *Client) GetAccountInsights(ctx context.Context, p AccountInsightsParams) (*InsightsResult, error) {
	metrics := p.Metrics
	if len(metrics) == 0 {
		metrics = defaultAccountMetrics
	}
	period := p.Period
	if period == "" {
		period = InsightPeriodDay
	}
	names := make([]string, len(metrics))
	for i, m := range metrics {
		names[i] = string(m)
	}
	q := url.Values{}
	q.Set("metric", strings.Join(names, ","))
	q.Set("period", string(period))
	if p.Since != "" {
		q.Set("since", p.Since)
	}
	if p.Until != "" {
		q.Set("until", p.Until)
	}
	if p.Breakdown != "" {
		q.Set("breakdown", string(p.Breakdown))
	}

	out := &InsightsResult{}
	if err := c.doGet(ctx, "me/threads_insights", q, out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetMetric is a convenience that retrieves a single media metric. It returns
// (nil, nil) when the API yields no data for the metric, mirroring the Python
// SDK's InsightsAPI.get_metric returning None.
func (c *Client) GetMetric(ctx context.Context, mediaID string, metric MediaMetric) (*InsightData, error) {
	resp, err := c.GetMediaInsights(ctx, mediaID, []MediaMetric{metric})
	if err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, nil
	}
	first := resp.Data[0]
	return &first, nil
}
