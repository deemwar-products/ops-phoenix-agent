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
			// strip braces, append to existing selector(s) — or start fresh
			inner := strings.Trim(fc.ErrorPattern, "{}")
			if label == "{}" {
				label = "{" + inner + "}"
			} else {
				label = label[:len(label)-1] + "," + inner + "}"
			}
		} else {
			// Case-insensitive line filter.
			//
			// Two traps here, both of which silently return zero rows:
			//  1. This is RE2, and a line filter is a *literal* string, not a
			//     regex. `|= "(?i)foo"` looks for the literal text "(?i)foo"
			//     and `|= "[Hh]ealthz"` looks for the literal "[Hh]...".
			//  2. The case-insensitive form of a literal filter is `|~`
			//     (case-insensitive =), not `|=`. `|~` takes the regex.
			//
			// So: `|~ "(?i)..."` for a regex, `|~ "(?i:...)"` also works, and
			// `|= "..."` for an already-lowercase literal.
			//
			// Quote and escape the caller's pattern for the LogQL string
			// literal — an error_pattern like "level":"ERROR" contains double
			// quotes, and an unescaped quote ends the literal and turns the
			// rest of the query into a syntax error.
			label += fmt.Sprintf(` |~ "(?i)%s"`, escapeLogQL(fc.ErrorPattern))
		}
	}

	return label
}

// escapeLogQL escapes a caller-supplied pattern for use inside a LogQL string
// literal. Backslashes and double quotes must be escaped; everything else
// (braces, pipes, parens) is regex syntax we want to keep.
func escapeLogQL(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	return strings.ReplaceAll(s, `"`, `\"`)
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
