package setup

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// StateDir returns the setup state directory.
func StateDir() string {
	dir, _ := os.UserHomeDir()
	return filepath.Join(dir, ".config", "sre-agent", "setup-state")
}

// StatePath returns the path for a specific provider+label state file.
func StatePath(provider, label string) string {
	if provider == "" {
		provider = "unknown"
	}
	return filepath.Join(StateDir(), fmt.Sprintf("%s-%s.yaml", provider, label))
}

// ErrNoState is returned when no state file exists for a provider+label.
var ErrNoState = errors.New("setup: no state found")

// ErrSuspended is returned when the setup is waiting for human input.
var ErrSuspended = errors.New("setup: suspended awaiting human")

// Status constants.
const (
	StatusPending    = "pending"
	StatusInProgress = "in_progress"
	StatusDone       = "done"
	StatusFailed     = "failed"
)

// StepState tracks the result of one step.
type StepState struct {
	Status  string            `yaml:"status"`
	URL     string            `yaml:"url,omitempty"`
	Message string            `yaml:"message,omitempty"`
	Fields  []string          `yaml:"fields,omitempty"`
	Output  map[string]string `yaml:"output,omitempty"`
}

// State holds the full setup session.
type State struct {
	Provider  string               `yaml:"provider"`
	Label     string               `yaml:"label"`
	Steps     map[string]StepState `yaml:"steps"`
	Inputs    map[string]string    `yaml:"inputs,omitempty"`
	StartedAt time.Time            `yaml:"started_at"`
}

// Fields on StepState we serialize with yaml — defined above.

// New creates a new empty state.
func New(provider, label string) *State {
	return &State{
		Provider:  provider,
		Label:     label,
		Steps:     make(map[string]StepState),
		Inputs:    make(map[string]string),
		StartedAt: time.Now().UTC(),
	}
}

// Save writes state to disk.
func (s *State) Save() error {
	dir := StateDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create setup state dir: %w", err)
	}
	path := StatePath(s.Provider, s.Label)
	data, err := yaml.Marshal(s)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// Load reads state from disk.
func Load(provider, label string) (*State, error) {
	path := StatePath(provider, label)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrNoState
		}
		return nil, fmt.Errorf("read setup state: %w", err)
	}
	var st State
	if err := yaml.Unmarshal(data, &st); err != nil {
		return nil, fmt.Errorf("parse setup state: %w", err)
	}
	if st.Steps == nil {
		st.Steps = make(map[string]StepState)
	}
	if st.Inputs == nil {
		st.Inputs = make(map[string]string)
	}
	return &st, nil
}

// Reset removes the state file.
func Reset(provider, label string) error {
	path := StatePath(provider, label)
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

// SetOutput stores output from a completed step.
func (s *State) SetOutput(stepID, key, value string) {
	ss := s.Steps[stepID]
	if ss.Output == nil {
		ss.Output = make(map[string]string)
	}
	ss.Output[key] = value
	ss.Status = StatusDone
	s.Steps[stepID] = ss
}

// Step is one unit of setup work.
type Step interface {
	ID() string
	Run(ctx context.Context, state *State, env Env) (Action, error)
}

// ActionKind classifies what the step wants next.
type ActionKind int

const (
	ActionDone ActionKind = iota
	ActionAwaitHuman
	ActionAwaitInput
)

// Action is what a step returns.
type Action struct {
	Kind    ActionKind
	URL     string
	Message string
	Fields  []string
	Output  map[string]string
}

// Env carries dependencies into steps.
type Env struct {
	Stdout         io.Writer
	Stderr         io.Writer
	Label          string
	NonInteractive bool // true in CI/automation — suppresses interactive prompts
}

// Runner executes steps.
type Runner struct{}

func (r *Runner) Run(ctx context.Context, state *State, steps []Step, em Emitter, env Env) error {
	for _, step := range steps {
		if state.Steps[step.ID()].Status == StatusDone {
			em.StepSkipped(step.ID())
			continue
		}

		em.StepStarted(step.ID())

		action, err := step.Run(ctx, state, env)
		if err != nil {
			state.Steps[step.ID()] = StepState{Status: StatusFailed}
			_ = state.Save()
			return fmt.Errorf("step %s: %w", step.ID(), err)
		}

		switch action.Kind {
		case ActionDone:
			for k, v := range action.Output {
				state.SetOutput(step.ID(), k, v)
			}
			em.StepDone(step.ID(), action.Output)
			_ = state.Save()

		case ActionAwaitHuman:
			state.Steps[step.ID()] = StepState{
				Status:  StatusInProgress,
				URL:     action.URL,
				Message: action.Message,
			}
			_ = state.Save()
			em.StepAwaitingHuman(step.ID(), action.URL, action.Message)
			return ErrSuspended

		case ActionAwaitInput:
			state.Steps[step.ID()] = StepState{
				Status:  StatusInProgress,
				Message: action.Message,
				Fields:  action.Fields,
			}
			_ = state.Save()
			em.StepAwaitingInput(step.ID(), action.Fields, action.Message)
			return ErrSuspended
		}
	}
	return nil
}
