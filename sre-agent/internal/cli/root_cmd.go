package cli

import (
	"fmt"
	"io"
	"errors"

	"github.com/deemwar-products/sre-agent/internal/config"
	"github.com/spf13/cobra"
)

// RootCommand builds the top-level `sre-agent` command tree.
func RootCommand(stdout, stderr io.Writer) *cobra.Command {
	cfg := config.MustLoad()

	root := &cobra.Command{
		Use:   "sre-agent",
		Short: "Autonomous SRE agent — detect, analyze, fix, deploy",
		Long:  "sre-agent observes your infrastructure, diagnoses errors with AI, and auto-fixes them.",
	}

	root.PersistentFlags().StringP("config", "c", "", "Config directory (default: ~/.config/sre-agent)")
	root.PersistentFlags().Bool("verbose", false, "Enable debug logging")

	root.AddCommand(DetectCmd(cfg, stdout, stderr))
	root.AddCommand(AnalyzeCmd(cfg, stdout, stderr))
	root.AddCommand(FixCmd(cfg, stdout, stderr))
	root.AddCommand(StatusCmd(cfg, stdout, stderr))
	root.AddCommand(NewSetupCommand(cfg, stdout, stderr))
	root.AddCommand(NewAuthCommand(stdout, stderr))

	return root
}

// Run executes the CLI from main.go.
func Run(args []string, stdout, stderr io.Writer) int {
	cmd := RootCommand(stdout, stderr)
	cmd.SetArgs(args)
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	if err := cmd.Execute(); err != nil {
		var cliErr *CLIError
		if errors.As(err, &cliErr) {
			return cliErr.Code
		}
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	return 0
}
