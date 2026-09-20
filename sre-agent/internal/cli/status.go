package cli

import (
	"fmt"
	"io"
	"time"

	"github.com/deemwar-products/sre-agent/internal/config"
	"github.com/deemwar-products/sre-agent/internal/credentials"
	"github.com/deemwar-products/sre-agent/internal/history"
	"github.com/deemwar-products/sre-agent/internal/observability"
	"github.com/spf13/cobra"
)

// StatusCmd returns `sre-agent status` — show agent health.
func StatusCmd(cfg *config.Config, stdout, stderr io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show agent status and recent run history",
		Long:  "Displays configured backends, last run, recent incidents.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsConfigured() {
				fmt.Fprintf(stdout, "not configured — run 'sre-agent setup' first\n")
				return nil
			}

			// Test backend connectivity
			adapter, err := observability.NewAdapter(cfg)
			if err != nil {
				fmt.Fprintf(stderr, "status: error creating adapter: %v\n", err)
			} else if err := adapter.TestConnection(cmd.Context()); err != nil {
				fmt.Fprintf(stderr, "status: backend unreachable: %v\n", err)
			} else {
				fmt.Fprintf(stdout, "status: backend reachable\n")
			}

			fmt.Fprintf(stdout, "\nsre-agent status\n")
			fmt.Fprintf(stdout, "───────────────\n\n")
			fmt.Fprintf(stdout, "Observability:  %s @ %s\n", cfg.Observability.Type, cfg.Observability.URL)
			fmt.Fprintf(stdout, "GitHub:         %s (base: %s)\n", cfg.GitHub.Repo, cfg.GitHub.BaseBranch)
			fmt.Fprintf(stdout, "AI:             %s / %s\n", cfg.AI.Provider, cfg.AI.Model)
			fmt.Fprintf(stdout, "Mode:           %s (dry_run=%v)\n", cfg.Agent.Mode, cfg.Agent.DryRun)

			// Show credential status (masked, no values)
			fmt.Fprintf(stdout, "\nCredentials:\n")
			creds := credentials.Status()
			for _, name := range []string{"github", "anthropic", "grafana"} {
				if c, ok := creds[name]; ok && c.Value != "" {
					fmt.Fprintf(stdout, "  %-12s ✓ %s\n", name, c.Source)
				} else {
					fmt.Fprintf(stdout, "  %-12s ✗ not found\n", name)
				}
			}

			// Show recent run history
			runs, err := history.Recent(10)
			if err != nil {
				fmt.Fprintf(stderr, "\n(run history unavailable: %v)\n", err)
				return nil
			}

			fmt.Fprintf(stdout, "\nRecent runs:\n")
			if len(runs) == 0 {
				fmt.Fprintf(stdout, "  (no runs yet)\n")
				return nil
			}
			for i := len(runs) - 1; i >= 0; i-- {
				r := runs[i]
				icon := "✓"
				if r.Status != "success" {
					icon = "✗"
				}
				fmt.Fprintf(stdout, "  %s %-10s %-7s %s",
					icon,
					r.Command,
					r.Status,
					relativeTime(r.EndedAt))
				if r.Summary != "" {
					fmt.Fprintf(stdout, "  %s", truncate(r.Summary, 50))
				}
				fmt.Fprintln(stdout)
			}
			return nil
		},
	}
	return cmd
}

func relativeTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		mins := int(d.Minutes())
		return fmt.Sprintf("%dm ago", mins)
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return t.Format("Jan 02")
	}
}
