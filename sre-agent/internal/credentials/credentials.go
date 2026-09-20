package credentials

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/deemwar-products/sre-agent/internal/store"
)

// Source indicates where a credential came from.
type Source string

const (
	SourceKeychain Source = "keychain"
	SourceEnv      Source = "env"
	SourceGhCLI    Source = "gh_cli"
	SourceFile     Source = "file"
	SourceNone     Source = "none"
)

// Credential represents one secret with its source.
type Credential struct {
	Name  string
	Value string
	Source Source
}

// ErrNotFound is the standard "missing credential" sentinel.
var ErrNotFound = store.ErrKeyNotFound

// Store delegates to the OS keychain store (same interface).
type Store = store.Store

// DefaultStore returns the keychain-backed store (or noop if unavailable).
func DefaultStore() Store {
	s, err := store.New()
	if err != nil {
		return &store.StoreFuncs{
			GetFn: func(key string) (string, error) { return "", ErrNotFound },
			PutFn: func(key, value string) error { return store.ErrStoreLocked },
			DeleteFn: func(key string) error { return store.ErrStoreLocked },
		}
	}
	return s
}

// --- Token resolvers (try multiple sources, return first match) ---

// GetGitHubToken checks env → gh CLI → keychain.
func GetGitHubToken() (Credential, error) {
	// 1. Env
	if v := os.Getenv("GITHUB_TOKEN"); v != "" {
		return Credential{Name: "github_token", Value: v, Source: SourceEnv}, nil
	}
	if v := os.Getenv("GH_TOKEN"); v != "" {
		return Credential{Name: "github_token", Value: v, Source: SourceEnv}, nil
	}

	// 2. gh CLI
	if v, err := getGhCLIToken(); err == nil && v != "" {
		return Credential{Name: "github_token", Value: v, Source: SourceGhCLI}, nil
	}

	// 3. Keychain
	store := DefaultStore()
	if v, err := store.Get("github_token"); err == nil && v != "" {
		return Credential{Name: "github_token", Value: v, Source: SourceKeychain}, nil
	}

	return Credential{Name: "github_token"}, fmt.Errorf("no GitHub token — set GITHUB_TOKEN or run `gh auth login`")
}

// GetAIKey checks env → keychain for the given env var name.
func GetAIKey(envName string) (Credential, error) {
	if envName == "" {
		envName = "ANTHROPIC_API_KEY"
	}
	if v := os.Getenv(envName); v != "" {
		return Credential{Name: envName, Value: v, Source: SourceEnv}, nil
	}
	s := DefaultStore()
	if v, err := s.Get(envName); err == nil && v != "" {
		return Credential{Name: envName, Value: v, Source: SourceKeychain}, nil
	}
	return Credential{Name: envName}, fmt.Errorf("no %s — set %s or `sre-agent auth set %s <token>`", envName, envName, envName)
}

// GetObservabilityToken checks env → keychain.
func GetObservabilityToken(provider string) (Credential, error) {
	envCandidates := []string{"GRAFANA_API_TOKEN", "LOKI_TOKEN", "OBSERVABILITY_TOKEN"}
	for _, env := range envCandidates {
		if v := os.Getenv(env); v != "" {
			return Credential{Name: env, Value: v, Source: SourceEnv}, nil
		}
	}
	keyCandidates := []string{"grafana_token", "observability_token"}
	s := DefaultStore()
	for _, key := range keyCandidates {
		if v, err := s.Get(key); err == nil && v != "" {
			return Credential{Name: key, Value: v, Source: SourceKeychain}, nil
		}
	}
	return Credential{Source: SourceNone}, fmt.Errorf("no %s token — set GRAFANA_API_TOKEN or `sre-agent auth set grafana_token <token>`", provider)
}

// --- Status (non-secret) ---

// Status returns credential availability without revealing values.
func Status() map[string]Credential {
	c := map[string]Credential{"github": {}, "anthropic": {}, "grafana": {}}
	if tok, err := GetGitHubToken(); err == nil {
		c["github"] = tok
	}
	if tok, err := GetAIKey("ANTHROPIC_API_KEY"); err == nil {
		c["anthropic"] = tok
	}
	if tok, err := GetObservabilityToken("grafana"); err == nil {
		c["grafana"] = tok
	}
	return c
}

// --- Utilities ---

// MaskToken returns first 4 chars + "****".
func MaskToken(tok string) string {
	if len(tok) <= 4 {
		return "****"
	}
	return tok[:4] + "****"
}

// --- Private ---

func getGhCLIToken() (string, error) {
	ghPath, err := exec.LookPath("gh")
	if err != nil {
		return "", err
	}
	out, err := exec.Command(ghPath, "auth", "token").Output()
	if err != nil {
		return "", err
	}
	tok := strings.TrimSpace(string(out))
	if tok == "" {
		return "", fmt.Errorf("gh CLI not authenticated")
	}
	return tok, nil
}
