package observability

import (
	"context"
	"time"
)

// Adapter is the interface for querying error logs from an observability backend.
type Adapter interface {
	// QueryErrors returns error log lines from the given time window.
	// timeWindow examples: "30m", "2h", "1d"
	QueryErrors(ctx context.Context, timeWindow string) ([]ErrorLog, error)

	// TestConnection checks if the backend is reachable.
	TestConnection(ctx context.Context) error
}

// ErrorLog represents one error log entry.
type ErrorLog struct {
	Timestamp time.Time
	Container string
	Message   string
	Labels    map[string]string
	Raw       string
}

// Provider type constants.
const (
	ProviderGrafanaCloud       = "grafana_cloud"
	ProviderLokiDirect         = "loki_direct"
	ProviderGrafanaSelfHosted  = "grafana_self_hosted"
)

// defaultErrorPattern is the fallback when a config carries no error_pattern.
// It is a LABEL filter, so it only matches logs whose "level" is a Loki label.
// Logs that carry level inside the message (e.g. JSON) need a line filter in
// the config instead, e.g. error_pattern: 'error|fatal|panic'.
const defaultErrorPattern = `level=~"(?i)error|fatal|panic"`

// orDefault returns the configured filters, filling in the legacy error
// pattern when the config left error_pattern empty.
func (f FilterConfig) orDefault() FilterConfig {
	if f.ErrorPattern == "" {
		f.ErrorPattern = defaultErrorPattern
	}
	return f
}
