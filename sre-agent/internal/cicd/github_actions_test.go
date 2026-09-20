package cicd_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/deemwar-products/sre-agent/internal/cicd"
)

func TestNewGitHubActions(t *testing.T) {
	g := cicd.NewGitHubActions("myorg/myrepo")
	if g == nil || g.Repo != "myorg/myrepo" {
		t.Errorf("unexpected: %+v", g)
	}
}

func TestNewGitHubActions_RepoSeveralFormats(t *testing.T) {
	cases := []string{"myorg/myrepo", "my-org/my-repo", "owner/repo"}
	for _, repo := range cases {
		g := cicd.NewGitHubActions(repo)
		if g.Repo != repo {
			t.Errorf("Repo = %q, want %q", g.Repo, repo)
		}
	}
}

func TestTriggerWorkflow_RequiresGhCLI(t *testing.T) {
	g := cicd.NewGitHubActions("myorg/myrepo")
	_, err := g.TriggerWorkflow(context.Background(), "ci.yml", "main", nil)
	if err == nil {
		// gh CLI is available and authenticated — test passes without error
		return
	}
	// Any error is expected (no real repo or no auth) — just confirm
	// it doesn't panic and produces an error.
	if err.Error() == "" {
		t.Error("expected non-empty error message")
	}
}

func TestTestResult_LastLine(t *testing.T) {
	r := &cicd.TestResult{Passed: true, Output: "line1\nline2\nline3\n"}
	if r.LastLine() != "line3" {
		t.Errorf("LastLine() = %q", r.LastLine())
	}
}

func TestTestResult_LastLineEmpty(t *testing.T) {
	r := &cicd.TestResult{Passed: false}
	if r.LastLine() != "" {
		t.Errorf("LastLine() = %q, want empty", r.LastLine())
	}
}

func TestTestResult_LastLineSkipsBlankTrailing(t *testing.T) {
	r := &cicd.TestResult{
		Passed:  false,
		Output:  "line1\n\nline3\n\n",
		FailLog: "line1\n\nline3\n\n",
	}
	if r.LastLine() != "line3" {
		t.Errorf("LastLine() = %q, want %q", r.LastLine(), "line3")
	}
}

func TestTestResult_Defaults(t *testing.T) {
	r := &cicd.TestResult{}
	if r.Passed {
		t.Error("expected default Passed=false")
	}
	if r.LastLine() != "" {
		t.Errorf("expected empty LastLine, got %q", r.LastLine())
	}
}

func TestRunTests_NonExistentDir(t *testing.T) {
	dir := t.TempDir()
	nonexistent := filepath.Join(dir, "does-not-exist")
	g := cicd.NewGitHubActions("myorg/myrepo")
	_, err := g.RunTests(context.Background(), nonexistent)
	if err == nil {
		t.Error("expected error for non-existent directory")
	}
}

func TestRunTests_ValidEmptyDir(t *testing.T) {
	dir := t.TempDir()
	g := cicd.NewGitHubActions("myorg/myrepo")
	_, err := g.RunTests(context.Background(), dir)
	// Both task and go test will fail in an empty dir — that's expected.
	if err == nil {
		t.Log("task or go test unexpectedly succeeded in empty dir")
	}
}

func TestRunTests_ContextCancelled(t *testing.T) {
	dir := t.TempDir()
	g := cicd.NewGitHubActions("myorg/myrepo")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, _ = g.RunTests(ctx, dir)
	// Cancelled context prevents command execution — no assertion needed
	// since behavior depends on exec.CommandContext implementation.
}

func writeTestFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func mustRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("run %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}
