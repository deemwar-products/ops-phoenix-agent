package observability

import (
	"fmt"
	"strings"
)

// FilterConfig holds the customer-specific filters from setup.
type FilterConfig struct {
	ContainerPatterns []string // e.g. ["myapp-*", "api-*"]
	HostPatterns      []string // e.g. ["prod-*"]
	ErrorPattern      string   // LogQL fragment, e.g. `level=~"(?i)error|fatal"`
}

// BuildLogQL constructs a LogQL query from the filters.
// Examples:
//
//   {container=~"myapp-.*|api-.*"} | severity=~"(?i)error|fatal"
//   {container=~"myapp-.*"} |= "(?i)(error|panic|fatal)"
func BuildLogQL(fc FilterConfig) string {
	parts := []string{}

	// Build the label selector
	if len(fc.ContainerPatterns) > 0 {
		patterns := globToRegex(fc.ContainerPatterns)
		parts = append(parts, fmt.Sprintf(`container=~"%s"`, strings.Join(patterns, "|")))
	}
	if len(fc.HostPatterns) > 0 {
		patterns := globToRegex(fc.HostPatterns)
		parts = append(parts, fmt.Sprintf(`host=~"%s"`, strings.Join(patterns, "|")))
	}

	label := "{"
	label += strings.Join(parts, ",")
	label += "}"

	// Add error pattern (either a line filter or label filter)
	if fc.ErrorPattern != "" {
		// If it contains a label selector like "level=~...", append to label
		// If it's a bare word, use as a regex line filter
		if strings.Contains(fc.ErrorPattern, "=") {
			// strip braces, append to existing
			inner := strings.Trim(fc.ErrorPattern, "{}")
			label = label[:len(label)-1] + "," + inner + "}"
		} else {
			label += fmt.Sprintf(` |= "(?i)(%s)"`, fc.ErrorPattern)
		}
	}

	return label
}

// globToRegex converts shell-style globs to regex patterns.
//   "myapp-*"  → "myapp-.*"
//   "api-v[12]" → "api-v[12]"
func globToRegex(patterns []string) []string {
	out := make([]string, len(patterns))
	for i, p := range patterns {
		out[i] = strings.ReplaceAll(p, "*", ".*")
	}
	return out
}
