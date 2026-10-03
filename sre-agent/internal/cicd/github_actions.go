package cicd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	// Run `go test ./...` directly. We don't use `task` because a missing
	// task exits 0 on some setups, which reads as "passed" when nothing ran.
	// The Go module may live in a subdirectory (e.g. apps/api), so find the
	// directory that actually contains go.mod and run the tests there.
	dir := workDir
	if _, err := os.Stat(filepath.Join(workDir, "go.mod")); err != nil {
		for _, sub := range []string{"apps/api", "api", "server", "backend"} {
			if _, err := os.Stat(filepath.Join(workDir, sub, "go.mod")); err == nil {
				dir = filepath.Join(workDir, sub)
				break
			}
		}
	}
	result, err := runCommand(ctx, dir, "go", "test", "./...")
	if os.Getenv("SRE_AGENT_DEBUG") != "" {
		fmt.Fprintf(os.Stderr, "fix: RunTests dir=%s err=%v passed=%v\n", dir, err, result != nil && result.Passed)
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
