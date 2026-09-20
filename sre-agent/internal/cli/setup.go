package cli

import (
	"errors"
	"fmt"
	"io"

	"github.com/deemwar-products/sre-agent/internal/config"
	"github.com/deemwar-products/sre-agent/internal/setup"
	"github.com/spf13/cobra"
)

// NewSetupCommand returns the `sre-agent setup` command tree.
func NewSetupCommand(_ *config.Config, stdout, stderr io.Writer) *cobra.Command {
	var (
		label    string
		provider string
		resume   bool
	)

	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Configure the agent for your environment",
		Long:  "Runs a setup wizard to configure observability, GitHub, and AI settings.",
	}

	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run the setup wizard",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetup(cmd, label, provider, resume)
		},
	}
	runCmd.Flags().StringVar(&label, "label", "default", "Account label (e.g. 'prod')")
	runCmd.Flags().StringVar(&provider, "provider", "", "Provider type (grafana_cloud, loki_direct, grafana_self_hosted)")
	runCmd.Flags().BoolVar(&resume, "resume", false, "Resume from a previously suspended setup")
	cmd.AddCommand(runCmd)

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show setup progress",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			st, err := setup.Load(provider, label)
			if err != nil {
				if errors.Is(err, setup.ErrNoState) {
					fmt.Fprintln(cmd.OutOrStdout(), "no setup in progress")
					return nil
				}
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "state file: %s\n", setup.StatePath(provider, label))
			fmt.Fprintf(cmd.OutOrStdout(), "started: %s\n", st.StartedAt.Format("2006-01-02 15:04:05"))
			for id, ss := range st.Steps {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s: %s", id, ss.Status)
				if ss.URL != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "  url=%s", ss.URL)
				}
				fmt.Fprintln(cmd.OutOrStdout())
			}
			return nil
		},
	}
	cmd.AddCommand(statusCmd)

	resetCmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset setup state",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := setup.Reset(provider, label); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "setup state cleared")
			return nil
		},
	}
	resetCmd.Flags().StringVar(&label, "label", "default", "Account label")
	resetCmd.Flags().StringVar(&provider, "provider", "", "Provider type")
	cmd.AddCommand(resetCmd)

	return cmd
}

func runSetup(cmd *cobra.Command, label, provider string, resume bool) error {
	stdout := cmd.OutOrStdout()
	stderr := cmd.OutOrStderr()

	var st *setup.State
	var err error
	if resume {
		st, err = setup.Load(provider, label)
		if err != nil {
			return fmt.Errorf("resume: no saved state for %s/%s: %w", provider, label, err)
		}
	} else {
		st = setup.New(provider, label)
	}

	steps := setup.DefaultSteps(provider)

	r := &setup.Runner{}
	em := setup.NewTextEmitter(stdout)

	env := setup.Env{
		Stdout: stdout,
		Stderr: stderr,
		Label:  label,
	}

	if err := r.Run(cmd.Context(), st, steps, em, env); err != nil {
		if errors.Is(err, setup.ErrSuspended) {
			fmt.Fprintln(stderr, "Setup suspended. Resume with: sre-agent setup run --resume --label="+label)
			return &CLIError{Code: 75, Msg: "setup suspended"}
		}
		return err
	}

	if err := persistConfig(st); err != nil {
		return err
	}

	_ = setup.Reset(provider, label)

	fmt.Fprintln(stdout, "Setup complete! Config written to", config.ConfigPath())
	fmt.Fprintln(stdout, "")
	fmt.Fprintln(stdout, "Next steps:")
	fmt.Fprintln(stdout, "  sre-agent detect          # check for errors")
	fmt.Fprintln(stdout, "  sre-agent status          # show agent health")
	return nil
}

func persistConfig(st *setup.State) error {
	cfg := &config.Config{}

	for stepID, ss := range st.Steps {
		for k, v := range ss.Output {
			switch stepID {
			case "provider":
				if k == "type" {
					cfg.Observability.Type = v
				}
			case "grafana":
				switch k {
				case "url":
					cfg.Observability.URL = v
				case "tenant":
					cfg.Observability.Tenant = v
				case "container_patterns":
					cfg.Observability.ContainerPatterns = []string{v}
				case "error_pattern":
					cfg.Observability.ErrorPattern = v
				}
			case "github":
				switch k {
				case "repo":
					cfg.GitHub.Repo = v
				case "base_branch":
					cfg.GitHub.BaseBranch = v
				}
			case "ai":
				switch k {
				case "model":
					cfg.AI.Model = v
				case "provider":
					cfg.AI.Provider = v
				}
			case "confirm":
				if k == "work_dir" {
					cfg.Agent.WorkDir = v
				}
			}
		}
	}

	cfg.Agent.Mode = "dev"
	cfg.Agent.TimeWindow = "1h"
	cfg.AI.MaxTokens = 1024

	return config.Save(cfg)
}
