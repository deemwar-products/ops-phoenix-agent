package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/deemwar-products/sre-agent/internal/ai"
	"github.com/deemwar-products/sre-agent/internal/config"
	"github.com/deemwar-products/sre-agent/internal/detect"
	"github.com/deemwar-products/sre-agent/internal/observability"
	"github.com/spf13/cobra"
)

// AnalyzeCmd returns `sre-agent analyze` — analyze errors with AI.
func AnalyzeCmd(cfg *config.Config, stdout, stderr io.Writer) *cobra.Command {
	var contextLevel string
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Analyze detected errors using AI",
		Long:  "Feeds error logs to the AI model with code context to identify root cause and propose fixes.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			start := time.Now().UTC()
			adapter, err := observability.NewAdapter(cfg)
			if err != nil {
				recordRun("analyze", "", "failed", err.Error(), start)
				return fmt.Errorf("create observability adapter: %w", err)
			}

			timeWindow := detect.DefaultTimeWindow(cfg.Agent.TimeWindow)
			raw, err := adapter.QueryErrors(cmd.Context(), timeWindow)
			if err != nil {
				recordRun("analyze", "", "failed", err.Error(), start)
				fmt.Fprintf(stderr, "analyze: error fetching logs: %v\n", err)
				return nil
			}

			groups := detect.Deduplicate(raw)
			fmt.Fprintf(stdout, "analyze: found %d unique error(s) (window=%s)\n", len(groups), timeWindow)

			if len(groups) == 0 {
				recordRun("analyze", "no errors", "success", "", start)
				fmt.Fprintln(stdout, "analyze: no errors to analyze")
				return nil
			}

			// Resolve AI key and create analyzer
			apiKeyEnv := cfg.AI.APIKeyEnv
			if apiKeyEnv == "" {
				apiKeyEnv = "ANTHROPIC_API_KEY"
			}
			analyzer, err := ai.New(ai.Provider(cfg.AI.Provider), cfg.AI.Model, apiKeyEnv)
			if err != nil {
				recordRun("analyze", "", "failed", err.Error(), start)
				fmt.Fprintf(stderr, "analyze: %v\n", err)
				return nil
			}

			fmt.Fprintf(stdout, "analyze: querying %s (%s, context=%s)...\n", cfg.AI.Provider, cfg.AI.Model, contextLevel)
			result, err := analyzer.Analyze(cmd.Context(), groups, contextLevel)
			if err != nil {
				recordRun("analyze", "", "failed", err.Error(), start)
				fmt.Fprintf(stderr, "analyze: AI call failed: %v\n", err)
				return nil
			}

			// Render output
			of, err := detect.ParseFormat(outputFormat)
			if err != nil {
				of = detect.FormatText
			}
			switch of {
			case detect.FormatJSON:
				data, _ := json.MarshalIndent(result, "", "  ")
				fmt.Fprintln(stdout, string(data))
			default:
				renderAnalysisText(stdout, result)
			}
			recordRun("analyze",
				fmt.Sprintf("%d findings", len(result.Findings)),
				"success", "", start)
			return nil
		},
	}
	cmd.Flags().StringVar(&contextLevel, "context", "medium", "Code context depth: minimal, medium, full")
	cmd.Flags().StringVar(&outputFormat, "format", "text", "Output format: text, json")
	return cmd
}

func renderAnalysisText(w io.Writer, result *ai.AnalysisResult) {
	fmt.Fprintf(w, "\nSummary: %s\n\n", result.Summary)
	if len(result.Findings) == 0 {
		fmt.Fprintln(w, "No actionable findings.")
		return
	}
	for i, f := range result.Findings {
		fmt.Fprintf(w, "[%d] %s (severity=%s confidence=%.2f)\n", i+1, f.Severity, f.Severity, f.Confidence)
		fmt.Fprintf(w, "    Error:   %s\n", truncate(f.Error, 100))
		fmt.Fprintf(w, "    Cause:   %s\n", truncate(f.RootCause, 200))
		fmt.Fprintf(w, "    Service: %s\n", f.AffectedSvc)
		fmt.Fprintf(w, "    Fix:     %s\n", truncate(f.SuggestedFix, 300))
		fmt.Fprintln(w)
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
