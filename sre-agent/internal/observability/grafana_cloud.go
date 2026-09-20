package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// GrafanaCloudAdapter queries Loki through Grafana Cloud's datasource API.
type GrafanaCloudAdapter struct {
	URL       string
	StackSlug string
	APIToken  string
	OrgID     string
	Tenant    string
}

// NewGrafanaCloudAdapter creates a new Grafana Cloud adapter.
// stackSlug: e.g. "myorg" → queries https://myorg.grafana.net/api/ds/query
func NewGrafanaCloudAdapter(stackSlug, apiToken, orgID, tenant string) *GrafanaCloudAdapter {
	base := fmt.Sprintf("https://%s.grafana.net/api/ds/query", stackSlug)
	return &GrafanaCloudAdapter{
		URL:       base,
		StackSlug: stackSlug,
		APIToken:  apiToken,
		OrgID:     orgID,
		Tenant:    tenant,
	}
}

// QueryErrors sends a LogQL query to Grafana Cloud's Loki datasource.
func (a *GrafanaCloudAdapter) QueryErrors(ctx context.Context, timeWindow string) ([]ErrorLog, error) {
	startOffset, err := parseDuration(timeWindow)
	if err != nil {
		return nil, fmt.Errorf("parse time window %q: %w", timeWindow, err)
	}

	end := time.Now().UnixMilli()
	start := time.Now().Add(-startOffset).UnixMilli()

	query := BuildLogQL(FilterConfig{
		ErrorPattern: `level=~"(?i)error|fatal|panic"`,
	})

	body := map[string]any{
		"queries": []map[string]any{
			{
				"refId":      "A",
				"expr":       query,
				"queryType":  "range",
				"maxLines":   500,
				"datasource": map[string]string{"type": "loki", "uid": a.OrgID},
			},
		},
		"from": start,
		"to":   end,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal query body: %w", err)
	}

	headers := map[string]string{
		"Authorization": "Bearer " + a.APIToken,
	}

	resp, err := doPOST(ctx, a.URL, bodyBytes, headers)
	if err != nil {
		return nil, fmt.Errorf("grafana query: %w", err)
	}

	return parseGrafanaResponse(resp)
}

// TestConnection checks Grafana Cloud availability.
func (a *GrafanaCloudAdapter) TestConnection(ctx context.Context) error {
	u := fmt.Sprintf("https://%s.grafana.net/api/health", a.StackSlug)
	headers := map[string]string{}
	if a.APIToken != "" {
		headers["Authorization"] = "Bearer " + a.APIToken
	}
	data, err := doGET(ctx, u, headers)
	if err != nil {
		return err
	}
	_ = data
	return nil
}

// parseGrafanaResponse unwraps Grafana's envelope response from a datasource query.
// Grafana wraps Loki results as: results.A.frames[0].data.values[] where each
// value is [timestamp_ms, log_line].
func parseGrafanaResponse(data []byte) ([]ErrorLog, error) {
	var env struct {
		Results map[string]struct {
			Frames []struct {
				Data struct {
					Values [][]string `json:"values"`
				} `json:"data"`
			} `json:"frames"`
		} `json:"results"`
	}
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("unmarshal grafana response: %w", err)
	}

	var out []ErrorLog
	for _, r := range env.Results {
		for _, frame := range r.Frames {
			for _, pair := range frame.Data.Values {
				if len(pair) != 2 {
					continue
				}
				ts := parseGrafanaTimestamp(pair[0])
				out = append(out, ErrorLog{
					Timestamp: ts,
					Message:   pair[1],
					Raw:       pair[1],
					Labels:    map[string]string{},
				})
			}
		}
	}
	return out, nil
}

// parseGrafanaTimestamp handles millisecond timestamps from Grafana.
// 13 digits = ms since epoch, anything else = ns since epoch.
func parseGrafanaTimestamp(s string) time.Time {
	n, err := parseInt64(s)
	if err != nil {
		return time.Time{}
	}
	if len(strings.TrimSpace(s)) <= 13 {
		return time.UnixMilli(n)
	}
	return time.Unix(0, n)
}

func parseInt64(s string) (int64, error) {
	var n int64
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

// IsStackSlug returns true if s looks like a bare stack slug (no dots, no scheme).
func IsStackSlug(s string) bool {
	return !strings.Contains(s, ".") && !strings.Contains(s, "/") && s != ""
}
