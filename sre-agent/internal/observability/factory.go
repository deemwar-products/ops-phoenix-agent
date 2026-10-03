package observability

import (
	"fmt"
	"strings"

	"github.com/deemwar-products/ops-phoenix-agent/sre-agent/internal/config"
	"github.com/deemwar-products/ops-phoenix-agent/sre-agent/internal/credentials"
)

// NewAdapter creates the appropriate adapter from config and injects
// the resolved observability token from the credentials layer
// (env → keychain). The token field on the config has yaml:"-",
// so it never persists — it's populated fresh each call.
func NewAdapter(cfg *config.Config) (Adapter, error) {
	// Resolve the observability token at runtime. If no token is found
	// anywhere, we still hand back an adapter — QueryErrors will fail
	// with a clear 401 so the user gets a useful error message.
	tok, _ := credentials.GetObservabilityToken(cfg.Observability.Type)
	token := tok.Value

	// The container/host/error filters the user configured at setup. Without
	// these the adapters fall back to a label-only filter, which silently
	// returns nothing for logs that carry level inside the message body.
	filters := FilterConfig{
		ContainerPatterns: cfg.Observability.ContainerPatterns,
		HostPatterns:      cfg.Observability.HostPatterns,
		ErrorPattern:      cfg.Observability.ErrorPattern,
	}

	switch cfg.Observability.Type {
	case ProviderLokiDirect:
		a := NewLokiDirectAdapter(
			cfg.Observability.URL,
			cfg.Observability.OrgID,
			cfg.Observability.Tenant,
			"",
			token, // basic-auth password = API token for Loki
		)
		a.Filters = filters
		return a, nil
	case ProviderGrafanaCloud:
		stack := extractStackSlug(cfg.Observability.URL)
		a := NewGrafanaCloudAdapter(
			stack,
			token, // Bearer token for Grafana Cloud
			cfg.Observability.OrgID,
			cfg.Observability.Tenant,
		)
		a.Filters = filters
		return a, nil
	case ProviderGrafanaSelfHosted:
		a := NewGrafanaSelfHostedAdapter(
			cfg.Observability.URL,
			token, // Bearer token for self-hosted Grafana
			cfg.Observability.OrgID,
			cfg.Observability.Tenant,
		)
		a.Filters = filters
		return a, nil
	default:
		return nil, fmt.Errorf("unknown observability provider: %q", cfg.Observability.Type)
	}
}

// extractStackSlug extracts "myorg" from "https://myorg.grafana.net" or "myorg" alone.
func extractStackSlug(s string) string {
	if idx := strings.Index(s, "://"); idx >= 0 {
		rest := s[idx+3:] // "myorg.grafana.net"
		if dot := strings.Index(rest, "."); dot >= 0 {
			return rest[:dot]
		}
		return rest
	}
	return s
}
