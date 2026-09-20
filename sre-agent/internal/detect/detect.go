package detect

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/deemwar-products/sre-agent/internal/observability"
)

// ErrorGroup is one unique error after deduplication.
type ErrorGroup struct {
	Fingerprint string
	Message     string
	Container   string
	Count       int
	FirstSeen   time.Time
	LastSeen    time.Time
	Labels      map[string]string
	SampleRaw   string
}

// OutputFormat controls how results are rendered.
type OutputFormat string

const (
	FormatText     OutputFormat = "text"
	FormatJSON     OutputFormat = "json"
	FormatMarkdown OutputFormat = "markdown"
)

// ParseFormat accepts "text", "json", or "markdown".
func ParseFormat(s string) (OutputFormat, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "text", "":
		return FormatText, nil
	case "json":
		return FormatJSON, nil
	case "markdown", "md":
		return FormatMarkdown, nil
	default:
		return "", fmt.Errorf("unknown format %q — use text, json, or markdown", s)
	}
}

// ValidateTimeWindow accepts durations like "30m", "2h", "1d", "5m".
func ValidateTimeWindow(s string) error {
	if s == "" {
		return fmt.Errorf("time window is empty")
	}
	// time.ParseDuration handles Go-style: 30m, 2h, 1d, 5m, 10s
	d, err := time.ParseDuration(s)
	if err == nil && d > 0 {
		return nil
	}
	// Bare minutes (e.g. "30") — treat as minutes
	var n int
	if _, err := fmt.Sscanf(s, "%d", &n); err == nil && n > 0 {
		return nil
	}
	return fmt.Errorf("invalid time window %q — use formats like 30m, 2h, 1d", s)
}

// DefaultTimeWindow returns the configured default or "1h".
func DefaultTimeWindow(cfg string) string {
	if cfg == "" {
		return "1h"
	}
	return cfg
}

// ---------------------------------------------------------------------------
// Deduplication
// ---------------------------------------------------------------------------

// fingerprint returns a stable key for grouping the same error together.
// We lowercase the message, collapse whitespace, and take the first 120
// characters — enough to distinguish different errors while collapsing
// near-duplicates (timestamps, IDs, memory addresses).
func fingerprint(msg string) string {
	s := strings.ToLower(strings.TrimSpace(msg))
	// Collapse whitespace
	var b strings.Builder
	b.Grow(len(s))
	prevSpace := false
	for _, r := range s {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			if !prevSpace {
				b.WriteRune(' ')
				prevSpace = true
			}
		} else {
			b.WriteRune(r)
			prevSpace = false
		}
	}
	normalized := b.String()
	if len(normalized) > 120 {
		return normalized[:120]
	}
	return normalized
}

// Deduplicate groups raw error logs into ErrorGroups.
// Each group contains all occurrences of the same fingerprint, sorted
// by most-recently seen first.
func Deduplicate(logs []observability.ErrorLog) []ErrorGroup {
	type keyed struct {
		key string
		log observability.ErrorLog
	}
	keys := make([]keyed, len(logs))
	for i, l := range logs {
		keys[i] = keyed{fingerprint(l.Message), l}
	}

	groups := make(map[string]*ErrorGroup)
	for _, k := range keys {
		g, exists := groups[k.key]
		if !exists {
			g = &ErrorGroup{
				Fingerprint: k.key,
				Message:     k.log.Message,
				Container:   k.log.Container,
				Labels:      k.log.Labels,
				SampleRaw:   k.log.Raw,
			}
			groups[k.key] = g
		}
		g.Count++
		if k.log.Timestamp.Before(g.FirstSeen) || g.FirstSeen.IsZero() {
			g.FirstSeen = k.log.Timestamp
		}
		if k.log.Timestamp.After(g.LastSeen) {
			g.LastSeen = k.log.Timestamp
			// Update message to the most recent variant (might have slightly
			// different wording like a changing error code)
			g.Message = k.log.Message
			g.Container = k.log.Container
			g.SampleRaw = k.log.Raw
		}
	}

	out := make([]ErrorGroup, 0, len(groups))
	for _, g := range groups {
		out = append(out, *g)
	}
	// Sort: most recently seen first, then highest count
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j].LastSeen.After(out[i].LastSeen) ||
				(out[j].LastSeen.Equal(out[i].LastSeen) && out[j].Count > out[i].Count) {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// Output rendering
// ---------------------------------------------------------------------------

// Render returns the error groups formatted for display.
func Render(groups []ErrorGroup, format OutputFormat, window string) string {
	switch format {
	case FormatJSON:
		return renderJSON(groups)
	case FormatMarkdown:
		return renderMarkdown(groups, window)
	default:
		return renderText(groups, window)
	}
}

func renderText(groups []ErrorGroup, window string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Found %d unique error(s) in the last %s\n\n", len(groups), window))
	for _, g := range groups {
		b.WriteString(fmt.Sprintf("  [%s] (x%d) %s\n", g.Container, g.Count, Truncate(g.Message, 120)))
		b.WriteString(fmt.Sprintf("    first: %s  last: %s\n", g.FirstSeen.Format("15:04:05"), g.LastSeen.Format("15:04:05")))
		if g.Count > 1 {
			b.WriteString(fmt.Sprintf("    repeated %d times\n", g.Count-1))
		}
		b.WriteString("\n")
	}
	return b.String()
}

func renderMarkdown(groups []ErrorGroup, window string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("# Error Report — last %s\n\n", window))
	b.WriteString(fmt.Sprintf("**%d unique error(s) detected**\n\n", len(groups)))

	b.WriteString("| # | Container | Count | Last Seen | Message |\n")
	b.WriteString("|---|-----------|-------|-----------|---------|\n")
	for i, g := range groups {
		msg := strings.ReplaceAll(Truncate(g.Message, 80), "|", "\\|")
		b.WriteString(fmt.Sprintf("| %d | `%s` | %d | %s | %s |\n",
			i+1, g.Container, g.Count, g.LastSeen.Format("15:04:05"), msg))
	}
	b.WriteString("\n")
	return b.String()
}

func renderJSON(groups []ErrorGroup) string {
	data, _ := json.MarshalIndent(groups, "", "  ")
	return string(data)
}

// Summary returns a one-line human summary (used by the run loop for decisions).
func Summary(groups []ErrorGroup) string {
	if len(groups) == 0 {
		return "no errors"
	}
	return fmt.Sprintf("%d unique error(s), top: %s (%dx in %s)",
		len(groups),
		Truncate(groups[0].Message, 60),
		groups[0].Count,
		groups[0].Container)
}

// Truncate shortens s to max runes, adding "..." if it was longer.
func Truncate(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
