package observability

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// LokiDirectAdapter queries Loki directly via its HTTP API.
type LokiDirectAdapter struct {
	URL       string
	OrgID     string
	Tenant    string
	BasicAuth *BasicAuth
}

// BasicAuth holds optional basic auth credentials.
type BasicAuth struct {
	User string
	Pass string
}

// NewLokiDirectAdapter creates a new Loki direct adapter.
func NewLokiDirectAdapter(rawURL, orgID, tenant, user, pass string) *LokiDirectAdapter {
	a := &LokiDirectAdapter{
		URL:    strings.TrimSuffix(rawURL, "/"),
		OrgID:  orgID,
		Tenant: tenant,
	}
	if user != "" || pass != "" {
		a.BasicAuth = &BasicAuth{User: user, Pass: pass}
	}
	return a
}

// QueryErrors queries Loki's query_range endpoint for error logs.
func (a *LokiDirectAdapter) QueryErrors(ctx context.Context, timeWindow string) ([]ErrorLog, error) {
	end := time.Now().UnixNano()
	startOffset, err := parseDuration(timeWindow)
	if err != nil {
		return nil, fmt.Errorf("parse time window %q: %w", timeWindow, err)
	}
	start := time.Now().Add(-startOffset).UnixNano()

	query := BuildLogQL(FilterConfig{
		ErrorPattern: `level=~"(?i)error|fatal|panic"`,
	})

	u := url.Values{}
	u.Set("query", query)
	u.Set("start", fmt.Sprintf("%d", start))
	u.Set("end", fmt.Sprintf("%d", end))
	u.Set("step", "60")

	fullURL := fmt.Sprintf("%s/loki/api/v1/query_range?%s", a.URL, u.Encode())

	data, err := a.get(ctx, fullURL)
	if err != nil {
		return nil, fmt.Errorf("loki query: %w", err)
	}

	return ParseLokiResponse(data)
}

// TestConnection checks if Loki is reachable.
func (a *LokiDirectAdapter) TestConnection(ctx context.Context) error {
	u := fmt.Sprintf("%s/ready", a.URL)
	_, err := a.get(ctx, u)
	return err
}

// get performs a GET request with auth headers.
func (a *LokiDirectAdapter) get(ctx context.Context, rawURL string) ([]byte, error) {
	headers := map[string]string{}
	if a.Tenant != "" {
		headers["X-Scope-OrgID"] = a.Tenant
	}
	if a.BasicAuth != nil {
		creds := base64.StdEncoding.EncodeToString(
			[]byte(a.BasicAuth.User + ":" + a.BasicAuth.Pass))
		headers["Authorization"] = "Basic " + creds
	}
	return doGET(ctx, rawURL, headers)
}

// parseDuration handles Loki time windows like "30m", "2h", "1d", "5m".
func parseDuration(s string) (time.Duration, error) {
	d, err := time.ParseDuration(s)
	if err == nil {
		return d, nil
	}
	// Try bare minutes
	var mins int
	if _, err := fmt.Sscanf(s, "%dm", &mins); err == nil {
		return time.Duration(mins) * time.Minute, nil
	}
	var hours int
	if _, err := fmt.Sscanf(s, "%dh", &hours); err == nil {
		return time.Duration(hours) * time.Hour, nil
	}
	return 0, fmt.Errorf("unknown duration format: %s", s)
}
