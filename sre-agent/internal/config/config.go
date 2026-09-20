package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ErrNotConfigured is returned when no config file exists yet.
var ErrNotConfigured = errors.New("config: not configured — run 'sre-agent setup'")

// ConfigDir returns the directory where config is stored.
func ConfigDir() string {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, _ := os.UserHomeDir()
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "sre-agent")
}

// ConfigPath returns the full path to the config file.
func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.yaml")
}

// Config holds all customer-specific settings. Token fields are never persisted.
type Config struct {
	Observability ObservabilityConfig `yaml:"observability"`
	GitHub        GitHubConfig        `yaml:"github"`
	AI            AIConfig            `yaml:"ai"`
	CICD          CICDConfig          `yaml:"cicd"`
	Agent         AgentConfig         `yaml:"agent"`
}

type ObservabilityConfig struct {
	Type              string   `yaml:"type"`
	URL               string   `yaml:"url"`
	Token             string   `yaml:"-"` // never persisted — resolved at runtime from credentials
	Tenant            string   `yaml:"tenant,omitempty"`
	OrgID             string   `yaml:"org_id,omitempty"`
	ContainerPatterns []string `yaml:"container_patterns,omitempty"`
	HostPatterns      []string `yaml:"host_patterns,omitempty"`
	ErrorPattern      string   `yaml:"error_pattern,omitempty"`
}

type GitHubConfig struct {
	Repo         string   `yaml:"repo"`
	TokenEnv     string   `yaml:"token_env,omitempty"` // env var name or keychain key
	BaseBranch   string   `yaml:"base_branch,omitempty"`
	Labels       []string `yaml:"labels,omitempty"`
	WorkRepoPath string   `yaml:"work_repo_path,omitempty"`
	AutoMerge    bool     `yaml:"auto_merge,omitempty"`
}

type AIConfig struct {
	Provider  string `yaml:"provider"`
	APIKeyEnv string `yaml:"api_key_env,omitempty"` // env var name or keychain key
	Model     string `yaml:"model"`
	MaxTokens int    `yaml:"max_tokens,omitempty"`
}

type CICDConfig struct {
	Type         string `yaml:"type,omitempty"`
	WorkflowName string `yaml:"workflow_name,omitempty"`
}

type AgentConfig struct {
	WorkDir    string `yaml:"work_dir,omitempty"`
	LogDir     string `yaml:"log_dir,omitempty"`
	Mode       string `yaml:"mode,omitempty"`
	DryRun     bool   `yaml:"dry_run,omitempty"`
	TimeWindow string `yaml:"time_window,omitempty"`
}

// Save writes the config to disk. Token fields are stripped before serialization.
func Save(cfg *Config) error {
	dir := ConfigDir()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	path := ConfigPath()
	return os.WriteFile(path, data, 0o600)
}

// Load reads the config from disk.
func Load() (*Config, error) {
	path := ConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrNotConfigured
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &cfg, nil
}

// MustLoad returns the config or panics (used at startup).
func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		if errors.Is(err, ErrNotConfigured) {
			return &Config{} // allow setup/auth to work without full config
		}
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	return cfg
}

// IsConfigured returns true if a config file exists.
func IsConfigured() bool {
	_, err := os.Stat(ConfigPath())
	return err == nil
}
