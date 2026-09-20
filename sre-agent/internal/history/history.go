package history

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	MaxRecentRuns = 20
)

// Run records one execution of a command.
type Run struct {
	Command   string    `yaml:"command"`
	StartedAt time.Time `yaml:"started_at"`
	EndedAt   time.Time `yaml:"ended_at"`
	Duration  string    `yaml:"duration"`
	Status    string    `yaml:"status"` // success | failed | partial
	Summary   string    `yaml:"summary,omitempty"`
	Error     string    `yaml:"error,omitempty"`
}

// RunLog is the persisted history file.
type RunLog struct {
	Runs []Run `yaml:"runs"`
}

// Path returns where run history is stored.
func Path() string {
	return filepath.Join(configDir(), "runs.yaml")
}

// Reset removes the run history file.
func Reset() error {
	path := Path()
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// Append adds a run to the log, trimming to MaxRecentRuns.
func Append(r Run) error {
	log, err := Load()
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	log.Runs = append(log.Runs, r)
	if len(log.Runs) > MaxRecentRuns {
		log.Runs = log.Runs[len(log.Runs)-MaxRecentRuns:]
	}
	return Save(log)
}

// Load reads the run history from disk.
func Load() (*RunLog, error) {
	data, err := os.ReadFile(Path())
	if err != nil {
		if os.IsNotExist(err) {
			return &RunLog{}, nil
		}
		return nil, fmt.Errorf("read run history: %w", err)
	}
	var log RunLog
	if err := yaml.Unmarshal(data, &log); err != nil {
		return nil, fmt.Errorf("parse run history: %w", err)
	}
	if log.Runs == nil {
		log.Runs = []Run{}
	}
	return &log, nil
}

// Save writes the run history to disk.
func Save(log *RunLog) error {
	dir := configDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create history dir: %w", err)
	}
	data, err := yaml.Marshal(log)
	if err != nil {
		return err
	}
	path := Path()
	return os.WriteFile(path, data, 0600)
}

// Recent returns the last N runs.
func Recent(n int) ([]Run, error) {
	log, err := Load()
	if err != nil {
		return nil, err
	}
	if n <= 0 || n > len(log.Runs) {
		return log.Runs, nil
	}
	return log.Runs[len(log.Runs)-n:], nil
}

func configDir() string {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base != "" {
		return filepath.Join(base, "sre-agent")
	}
	home, _ := os.UserHomeDir()
	if home != "" {
		return filepath.Join(home, ".config", "sre-agent")
	}
	// Fallback: relative path (should only happen in unusual envs)
	return filepath.Join(".sre-agent")
}
