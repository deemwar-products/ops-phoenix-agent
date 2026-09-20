package cicd

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// GitHubActions manages CI workflows for GitHub repos.
type GitHubActions struct {
	Repo string
}

func NewGitHubActions(repo string) *GitHubActions {
	return &GitHubActions{Repo: repo}
}

// TriggerWorkflow starts a GitHub Actions workflow run and returns its URL.
func (g *GitHubActions) TriggerWorkflow(ctx context.Context, workflowName, branch string, inputs map[string]string) (string, error) {
	ghPath, err := exec.LookPath("gh")
	if err != nil {
		return "", fmt.Errorf("gh CLI not found: %w", err)
	}

	args := []string{"workflow", "run", workflowName, "--repo", g.Repo, "--ref", branch}
	for k, v := range inputs {
		args = append(args, "-f", k+"="+v)
	}
	cmd := exec.CommandContext(ctx, ghPath, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("trigger workflow %s: %w\n%s", workflowName, err, string(out))
	}
	return fmt.Sprintf("https://github.com/%s/actions", g.Repo), nil
}

// RunTests executes the project's test suite in the given directory.
// It tries Taskfile first, then falls back to go test / bun test.
func (g *GitHubActions) RunTests(ctx context.Context, workDir string) (*TestResult, error) {
	// Try task first (the project uses Taskfile)
	result, err := runCommand(ctx, workDir, "task", "test:all")
	if err == nil {
		return result, nil
	}

	// Fall back to go test for Go projects
	result, err = runCommand(ctx, workDir, "go", "test", "./...")
	if err == nil {
		return result, nil
	}

	return result, err
}

// TestResult summarizes a test run.
type TestResult struct {
	Passed  bool
	Output  string
	FailLog string
}

func runCommand(ctx context.Context, dir, name string, args ...string) (*TestResult, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		return &TestResult{Passed: false, Output: out.String(), FailLog: out.String()}, err
	}
	return &TestResult{Passed: true, Output: out.String()}, nil
}

// LastLine returns the last non-empty line of output.
func (t *TestResult) LastLine() string {
	lines := strings.Split(strings.TrimSpace(t.Output), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) != "" {
			return strings.TrimSpace(lines[i])
		}
	}
	return ""
}
