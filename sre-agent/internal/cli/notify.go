package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/deemwar-products/sre-agent/internal/config"
	"github.com/deemwar-products/sre-agent/internal/credentials"
	"github.com/deemwar-products/sre-agent/internal/notify"
)

// maybeNotify dispatches a run result to the configured chat backend.
// Flag semantics: "" = config default, "off" = never, otherwise a backend name.
// Notification failures are warnings only — they never fail the run itself.
func maybeNotify(ctx context.Context, cfg *config.Config, flag string, r notify.Result, stderr io.Writer) {
	backend := flag
	if backend == "" && cfg.Notify.Teams.Enabled {
		backend = "teams"
	}
	if backend == "" || backend == "off" {
		return
	}
	if backend != "teams" {
		fmt.Fprintf(stderr, "notify: unknown backend %q\n", backend)
		return
	}
	if r.Status == notify.StatusSuccess && !cfg.Notify.Teams.OnSuccess {
		return
	}

	wh, err := credentials.GetTeamsWebhookURL()
	if err != nil {
		fmt.Fprintf(stderr, "notify: %v\n", err)
		return
	}

	if r.Environment == "" {
		r.Environment = cfg.Notify.Teams.Environment
	}
	if r.LogsURL == "" {
		r.LogsURL = cfg.Notify.Teams.LogsURL
	}

	n := notify.NewTeamsNotifier(wh.Value, cfg.Notify.Teams.MaxErrorsShown)
	if err := n.Notify(ctx, r); err != nil {
		fmt.Fprintf(stderr, "notify: teams: %v\n", err)
		return
	}
	fmt.Fprintf(stderr, "notify: teams ✓ posted\n")
}
