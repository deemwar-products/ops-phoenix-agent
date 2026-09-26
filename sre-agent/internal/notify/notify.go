// Package notify posts agent run results to chat backends (Teams, ...).
package notify

import (
	"context"
	"time"

	"github.com/deemwar-products/sre-agent/internal/detect"
)

// Status describes how an agent run ended.
type Status string

const (
	// StatusSuccess — run completed, no errors found.
	StatusSuccess Status = "success"
	// StatusErrors — run completed, errors were found.
	StatusErrors Status = "errors"
	// StatusFailed — the run itself failed (backend down, auth, ...).
	StatusFailed Status = "failed"
)

// Result is the backend-agnostic summary of one agent run.
type Result struct {
	Command     string              // e.g. "detect"
	Status      Status              // outcome of the run
	Environment string              // e.g. "reqsume-prod"
	TimeWindow  string              // e.g. "1h"
	Duration    time.Duration       // wall time of the run
	Groups      []detect.ErrorGroup // deduplicated errors (may be empty)
	LogsURL     string              // deep link shown as a button (optional)
	Err         string              // failure detail when Status == StatusFailed
}

// Notifier posts a run Result to a chat backend.
type Notifier interface {
	Name() string
	Notify(ctx context.Context, r Result) error
}
