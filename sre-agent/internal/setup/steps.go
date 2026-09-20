package setup

import (
	"context"
	"fmt"
	"os"

	"github.com/deemwar-products/sre-agent/internal/credentials"
)

// DefaultSteps returns the standard setup step sequence.
func DefaultSteps(providerOverride string) []Step {
	steps := []Step{
		&promptProviderStep{provider: providerOverride},
		&promptGrafanaStep{},
		&promptAuthStep{name: "observability"},
		&promptContainerStep{},
		&promptGitHubStep{},
		&promptAuthStep{name: "github"},
		&promptAIStep{},
		&promptAuthStep{name: "ai"},
		&promptConfirmStep{},
	}
	return steps
}

// promptProviderStep asks which observability provider to use.
type promptProviderStep struct {
	provider string
}

func (s *promptProviderStep) ID() string { return "provider" }

func (s *promptProviderStep) Run(ctx context.Context, state *State, env Env) (Action, error) {
	if s.provider != "" {
		state.SetOutput(s.ID(), "type", s.provider)
		return Action{Kind: ActionDone, Output: map[string]string{"type": s.provider}}, nil
	}
	if v, ok := state.Inputs["provider"]; ok && v != "" {
		state.SetOutput(s.ID(), "type", v)
		return Action{Kind: ActionDone, Output: map[string]string{"type": v}}, nil
	}
	fmt.Fprintf(env.Stdout, "  Observability provider:\n")
	fmt.Fprintf(env.Stdout, "    1. Grafana Cloud (default)\n")
	fmt.Fprintf(env.Stdout, "    2. Loki Direct\n")
	fmt.Fprintf(env.Stdout, "    3. Grafana Self-Hosted\n")
	state.SetOutput(s.ID(), "type", "grafana_cloud")
	return Action{Kind: ActionDone, Output: map[string]string{"type": "grafana_cloud"}}, nil
}

// promptGrafanaStep collects Grafana/Loki connection details.
type promptGrafanaStep struct{}

func (s *promptGrafanaStep) ID() string { return "grafana" }

func (s *promptGrafanaStep) Run(ctx context.Context, state *State, env Env) (Action, error) {
	if url, ok := state.Inputs["grafana_url"]; ok && url != "" {
		output := map[string]string{"url": url}
		if tenant, ok := state.Inputs["grafana_tenant"]; ok {
			output["tenant"] = tenant
		}
		if cp, ok := state.Inputs["container_patterns"]; ok {
			output["container_patterns"] = cp
		}
		if ep, ok := state.Inputs["error_pattern"]; ok {
			output["error_pattern"] = ep
		}
		for k, v := range output {
			state.SetOutput(s.ID(), k, v)
		}
		return Action{Kind: ActionDone, Output: output}, nil
	}

	fmt.Fprintf(env.Stdout, "\n  Grafana/Loki configuration:\n")
	output := map[string]string{
		"url":                "https://observability.deemwar.com",
		"container_patterns": "myapp-*",
		"error_pattern":      `level=~"(?i)error|fatal"`,
	}
	for k, v := range output {
		state.SetOutput(s.ID(), k, v)
	}
	return Action{Kind: ActionDone, Output: output}, nil
}

// promptContainerStep collects container filter patterns.
type promptContainerStep struct{}

func (s *promptContainerStep) ID() string { return "containers" }

func (s *promptContainerStep) Run(ctx context.Context, state *State, env Env) (Action, error) {
	patterns := "myapp-*"
	if v, ok := state.Inputs["container_patterns"]; ok && v != "" {
		patterns = v
	}
	state.SetOutput(s.ID(), "patterns", patterns)
	return Action{Kind: ActionDone, Output: map[string]string{"patterns": patterns}}, nil
}

// promptGitHubStep collects GitHub repo details.
type promptGitHubStep struct{}

func (s *promptGitHubStep) ID() string { return "github" }

func (s *promptGitHubStep) Run(ctx context.Context, state *State, env Env) (Action, error) {
	if repo, ok := state.Inputs["github_repo"]; ok && repo != "" {
		output := map[string]string{"repo": repo}
		if bb, ok := state.Inputs["base_branch"]; ok {
			output["base_branch"] = bb
		} else {
			output["base_branch"] = "main"
		}
		for k, v := range output {
			state.SetOutput(s.ID(), k, v)
		}
		return Action{Kind: ActionDone, Output: output}, nil
	}

	fmt.Fprintf(env.Stdout, "\n  GitHub repository:\n")
	output := map[string]string{"repo": "myorg/myapp", "base_branch": "main"}
	for k, v := range output {
		state.SetOutput(s.ID(), k, v)
	}
	return Action{Kind: ActionDone, Output: output}, nil
}

// promptAIStep collects AI provider details.
type promptAIStep struct{}

func (s *promptAIStep) ID() string { return "ai" }

func (s *promptAIStep) Run(ctx context.Context, state *State, env Env) (Action, error) {
	model := "claude-sonnet-4-20250514"
	provider := "anthropic"
	if m, ok := state.Inputs["model"]; ok && m != "" {
		model = m
	}
	if p, ok := state.Inputs["ai_provider"]; ok && p != "" {
		provider = p
	}
	output := map[string]string{"provider": provider, "model": model}
	for k, v := range output {
		state.SetOutput(s.ID(), k, v)
	}
	return Action{Kind: ActionDone, Output: output}, nil
}

// promptAuthStep handles credentials. Behavior differs per integration:
//   - observability: try env (GRAFANA_API_TOKEN), then prompt for token if missing
//   - github:        prefer `gh auth status`, fall back to GITHUB_TOKEN, then prompt
//   - ai:            try env (ANTHROPIC_API_KEY), then prompt
type promptAuthStep struct {
	name string
}

func (s *promptAuthStep) ID() string           { return "auth_" + s.name }
func (s *promptAuthStep) canSkip(env Env) bool { return env.NonInteractive }

func (s *promptAuthStep) Run(ctx context.Context, state *State, env Env) (Action, error) {
	store := credentials.DefaultStore()

	switch s.name {
	case "github":
		return s.runGitHub(state, store, env)
	case "ai":
		return s.runAI(state, store, env)
	case "observability":
		return s.runObservability(state, store, env)
	default:
		return Action{Kind: ActionDone}, nil
	}
}

func (s *promptAuthStep) runGitHub(state *State, store credentials.Store, env Env) (Action, error) {
	// 1. Check existing sources (gh CLI, env, keychain)
	tok, err := credentials.GetGitHubToken()
	if err == nil && tok.Value != "" {
		fmt.Fprintf(env.Stdout, "  auth: GitHub — %s (%s)\n", credentials.MaskToken(tok.Value), tok.Source)
		state.SetOutput(s.ID(), "github_source", string(tok.Source))
		state.SetOutput(s.ID(), "github_token_set", "yes")
		return Action{Kind: ActionDone}, nil
	}

	// 2. Non-interactive: skip prompt
	if s.canSkip(env) {
		fmt.Fprintf(env.Stderr, "  auth: GitHub — no token (set GITHUB_TOKEN or run `gh auth login`)\n")
		state.SetOutput(s.ID(), "github_token_set", "no")
		return Action{Kind: ActionDone}, nil
	}

	// 3. Interactive: offer gh CLI or PAT, then resume
	fmt.Fprintf(env.Stdout, "\n  GitHub credentials:\n")
	fmt.Fprintf(env.Stdout, "    1. Run `gh auth login` (recommended — uses browser)\n")
	fmt.Fprintf(env.Stdout, "    2. Paste a Personal Access Token (fine-grained: issues+PRs+contents)\n")
	fmt.Fprintf(env.Stdout, "    3. Skip for now (set GITHUB_TOKEN later)\n")
	state.SetOutput(s.ID(), "github_token_set", "needs_action")
	return Action{Kind: ActionDone}, nil
}

func (s *promptAuthStep) runAI(state *State, store credentials.Store, env Env) (Action, error) {
	provider := "anthropic"
	if v, ok := state.Steps["ai"].Output["provider"]; ok && v != "" {
		provider = v
	}
	envName := "ANTHROPIC_API_KEY"
	if provider != "anthropic" {
		envName = "AI_API_KEY"
	}

	// 1. Env var
	if v := os.Getenv(envName); v != "" {
		fmt.Fprintf(env.Stdout, "  auth: AI — env %s (%s)\n", envName, credentials.MaskToken(v))
		state.SetOutput(s.ID(), "ai_source", "env")
		state.SetOutput(s.ID(), "ai_env_var", envName)
		state.SetOutput(s.ID(), "ai_token_set", "yes")
		return Action{Kind: ActionDone}, nil
	}

	// 2. Keychain
	if v, err := store.Get(envName); err == nil && v != "" {
		fmt.Fprintf(env.Stdout, "  auth: AI — keychain (%s)\n", credentials.MaskToken(v))
		state.SetOutput(s.ID(), "ai_source", "keychain")
		state.SetOutput(s.ID(), "ai_env_var", envName)
		state.SetOutput(s.ID(), "ai_token_set", "yes")
		return Action{Kind: ActionDone}, nil
	}

	// 3. Non-interactive
	if s.canSkip(env) {
		fmt.Fprintf(env.Stderr, "  auth: AI — no key (set %s env var)\n", envName)
		state.SetOutput(s.ID(), "ai_token_set", "no")
		return Action{Kind: ActionDone}, nil
	}

	// 4. Interactive prompt
	fmt.Fprintf(env.Stdout, "\n  AI provider key:\n")
	fmt.Fprintf(env.Stdout, "    1. Set %s in your shell (recommended for dev)\n", envName)
	fmt.Fprintf(env.Stdout, "    2. Provide token now (will be stored in macOS Keychain)\n")
	fmt.Fprintf(env.Stdout, "    3. Skip for now\n")
	state.SetOutput(s.ID(), "ai_token_set", "needs_action")
	return Action{Kind: ActionDone}, nil
}

func (s *promptAuthStep) runObservability(state *State, store credentials.Store, env Env) (Action, error) {
	// 1. Env var candidates
	for _, envName := range []string{"GRAFANA_API_TOKEN", "LOKI_TOKEN", "OBSERVABILITY_TOKEN"} {
		if v := os.Getenv(envName); v != "" {
			fmt.Fprintf(env.Stdout, "  auth: Observability — env %s (%s)\n", envName, credentials.MaskToken(v))
			state.SetOutput(s.ID(), "obs_source", "env")
			state.SetOutput(s.ID(), "obs_env_var", envName)
			state.SetOutput(s.ID(), "obs_token_set", "yes")
			return Action{Kind: ActionDone}, nil
		}
	}

	// 2. Keychain
	for _, key := range []string{"grafana_token", "observability_token"} {
		if v, err := store.Get(key); err == nil && v != "" {
			fmt.Fprintf(env.Stdout, "  auth: Observability — keychain: %s (%s)\n", key, credentials.MaskToken(v))
			state.SetOutput(s.ID(), "obs_source", "keychain")
			state.SetOutput(s.ID(), "obs_token_set", "yes")
			return Action{Kind: ActionDone}, nil
		}
	}

	// 3. Non-interactive
	if s.canSkip(env) {
		fmt.Fprintf(env.Stderr, "  auth: Observability — no token (set GRAFANA_API_TOKEN)\n")
		state.SetOutput(s.ID(), "obs_token_set", "no")
		return Action{Kind: ActionDone}, nil
	}

	// 4. Interactive prompt
	fmt.Fprintf(env.Stdout, "\n  Observability credentials:\n")
	fmt.Fprintf(env.Stdout, "    1. Provide a Grafana Cloud API token now (stored in macOS Keychain)\n")
	fmt.Fprintf(env.Stdout, "    2. Skip — set GRAFANA_API_TOKEN env var later\n")
	state.SetOutput(s.ID(), "obs_token_set", "needs_action")
	return Action{Kind: ActionDone}, nil
}

// promptConfirmStep shows a summary.
type promptConfirmStep struct{}

func (s *promptConfirmStep) ID() string { return "confirm" }

func (s *promptConfirmStep) Run(ctx context.Context, state *State, env Env) (Action, error) {
	fmt.Fprintf(env.Stdout, "\n  Summary:\n")
	for stepID, ss := range state.Steps {
		for k, v := range ss.Output {
			fmt.Fprintf(env.Stdout, "    %s.%s = %s\n", stepID, k, v)
		}
	}
	fmt.Fprintf(env.Stdout, "\n  Saving...\n")
	state.SetOutput(s.ID(), "work_dir", "~/.config/sre-agent")
	return Action{Kind: ActionDone, Output: map[string]string{"work_dir": "~/.config/sre-agent"}}, nil
}
