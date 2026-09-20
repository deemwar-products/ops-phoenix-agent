package cli

import (
	"fmt"
	"io"
	"time"

	"github.com/deemwar-products/sre-agent/internal/config"
	"github.com/deemwar-products/sre-agent/internal/detect"
	"github.com/deemwar-products/sre-agent/internal/observability"
	"github.com/spf13/cobra"
)

// DetectCmd returns `sre-agent detect` — query logs for errors.
func DetectCmd(cfg *config.Config, stdout, stderr io.Writer) *cobra.Command {
	var timeWindow string
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "detect [time-window]",
		Short: "Query logs for errors in the given time window",
		Long:  "Queries your observability backend for errors in the last N minutes/hours, deduplicates, and prints a summary.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			start := time.Now().UTC()
			if len(args) > 0 {
				timeWindow = args[0]
			}
			if timeWindow == "" {
				timeWindow = detect.DefaultTimeWindow(cfg.Agent.TimeWindow)
			}
			if err := detect.ValidateTimeWindow(timeWindow); err != nil {
				recordRun("detect", "", "failed", err.Error(), start)
				return err
			}

			of, err := detect.ParseFormat(outputFormat)
			if err != nil {
				recordRun("detect", "", "failed", err.Error(), start)
				return err
			}

			adapter, err := observability.NewAdapter(cfg)
			if err != nil {
				recordRun("detect", "", "failed", err.Error(), start)
				return fmt.Errorf("create observability adapter: %w", err)
			}

			raw, err := adapter.QueryErrors(cmd.Context(), timeWindow)
			if err != nil {
				recordRun("detect", "", "failed", err.Error(), start)
				fmt.Fprintf(stderr, "detect: backend error: %v\n", err)
				return nil
			}

			groups := detect.Deduplicate(raw)
			fmt.Fprintln(stdout, detect.Render(groups, of, timeWindow))
			summary := fmt.Sprintf("%d error(s) in %s", len(groups), timeWindow)
			status := "success"
			if len(groups) > 0 {
				status = "partial"
			}
			recordRun("detect", summary, status, "", start)
			return nil
		},
	}
	cmd.Flags().StringVar(&timeWindow, "time-window", "", "Time window (e.g. 30m, 2h, 1d) — overrides config default")
	cmd.Flags().StringVar(&outputFormat, "format", "text", "Output format: text, json, markdown")
	return cmd
}
