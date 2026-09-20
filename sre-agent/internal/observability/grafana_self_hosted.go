package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// GrafanaSelfHostedAdapter queries a self-hosted Grafana instance's datasource proxy.
type GrafanaSelfHostedAdapter struct {
	URL           string
	APIToken      string
	DatasourceUID string
	Tenant        string
}

// NewGrafanaSelfHostedAdapter creates a new self-hosted Grafana adapter.
func NewGrafanaSelfHostedAdapter(url, apiToken, datasourceUID, tenant string) *GrafanaSelfHostedAdapter {
	return &GrafanaSelfHostedAdapter{
		URL:           strings.TrimSuffix(url, "/"),
		APIToken:      apiToken,
		DatasourceUID: datasourceUID,
		Tenant:        tenant,
	}
}

// QueryErrors queries the self-hosted Grafana's Loki datasource.
func (a *GrafanaSelfHostedAdapter) QueryErrors(ctx context.Context, timeWindow string) ([]ErrorLog, error) {
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
				"datasource": map[string]string{"type": "loki", "uid": a.DatasourceUID},
			},
		},
		"from": start,
		"to":   end,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal query body: %w", err)
	}

	fullURL := fmt.Sprintf("%s/api/ds/query", a.URL)
	headers := map[string]string{
		"Authorization": "Bearer " + a.APIToken,
	}

	resp, err := doPOST(ctx, fullURL, bodyBytes, headers)
	if err != nil {
		return nil, fmt.Errorf("grafana self-hosted query: %w", err)
	}

	return parseGrafanaResponse(resp)
}

// TestConnection checks Grafana availability.
func (a *GrafanaSelfHostedAdapter) TestConnection(ctx context.Context) error {
	u := fmt.Sprintf("%s/api/health", a.URL)
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
