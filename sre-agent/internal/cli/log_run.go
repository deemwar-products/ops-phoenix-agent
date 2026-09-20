package cli

import (
	"time"

	"github.com/deemwar-products/sre-agent/internal/history"
)

// recordRun saves a run to the history log. Errors are silently ignored —
// logging failure should never crash the user-facing command.
func recordRun(commandName, summary, status, errorMsg string, startedAt time.Time) {
	r := history.Run{
		Command:   commandName,
		StartedAt: startedAt,
		EndedAt:   time.Now().UTC(),
		Duration:  time.Since(startedAt).Round(time.Millisecond).String(),
		Status:    status,
		Summary:   summary,
		Error:     errorMsg,
	}
	_ = history.Append(r)
}
