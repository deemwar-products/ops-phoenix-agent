package vcs_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/deemwar-products/sre-agent/internal/vcs"
)

// setupTestRepo creates a temp dir with a git repo that has a single commit on main.
func setupTestRepo(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "sre-agent-vcs-test-*")
	if err != nil {
		t.Fatal(err)
	}
	mustGit(t, dir, "init", "-b", "main")
	mustGit(t, dir, "config", "--local", "user.email", "test@test.com")
	mustGit(t, dir, "config", "--local", "user.name", "Test")

	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("hello\n"), 0o644); err != nil {
		os.RemoveAll(dir)
		t.Fatal(err)
	}
	mustGit(t, dir, "add", "README.md")
	mustGit(t, dir, "commit", "-m", "init")
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

func mustGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out))
}

func TestNewGitHub_EmptyRepo(t *testing.T) {
	_, err := vcs.NewGitHub("", "main", "GITHUB_TOKEN", "")
	if err == nil {
		t.Fatal("expected error for empty repo")
	}
	if !strings.Contains(err.Error(), "repo is empty") {
		t.Errorf("expected 'repo is empty' error, got: %v", err)
	}
}

func TestNewGitHub_ValidInputs(t *testing.T) {
	g, err := vcs.NewGitHub("myorg/myrepo", "develop", "MY_TOKEN", "/tmp/work")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.Repo != "myorg/myrepo" {
		t.Errorf("Repo = %q, want %q", g.Repo, "myorg/myrepo")
	}
	if g.BaseBranch != "develop" {
		t.Errorf("BaseBranch = %q, want %q", g.BaseBranch, "develop")
	}
}

func TestNewGitHub_DefaultsBaseBranch(t *testing.T) {
	g, err := vcs.NewGitHub("myorg/myrepo", "", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.BaseBranch != "main" {
		t.Errorf("expected default base branch 'main', got %q", g.BaseBranch)
	}
}

func TestBranchName_Unique(t *testing.T) {
	b1 := vcs.BranchName()
	b2 := vcs.BranchName()
	if b1 == b2 {
		t.Error("BranchName should return unique values")
	}
	if !strings.HasPrefix(b1, "sre-agent/fix-") {
		t.Errorf("expected branch prefix 'sre-agent/fix-', got %q", b1)
	}
}

func TestCheckoutBranch(t *testing.T) {
	dir := setupTestRepo(t)
	g, err := vcs.NewGitHub("myorg/myrepo", "main", "GITHUB_TOKEN", dir)
	if err != nil {
		t.Fatal(err)
	}
	branch := "sre-agent/test-branch"
	ctx := context.Background()
	if err := g.CheckoutBranch(ctx, dir, branch); err != nil {
		t.Fatalf("CheckoutBranch failed: %v", err)
	}
	current := mustGit(t, dir, "rev-parse", "--abbrev-ref", "HEAD")
	if current != branch {
		t.Errorf("current branch = %q, want %q", current, branch)
	}
}

func TestCheckoutBranch_Duplicate(t *testing.T) {
	dir := setupTestRepo(t)
	g, err := vcs.NewGitHub("myorg/myrepo", "main", "GITHUB_TOKEN", dir)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := g.CheckoutBranch(ctx, dir, "dup-branch"); err != nil {
		t.Fatalf("first checkout: %v", err)
	}
	if err := g.CheckoutBranch(ctx, dir, "dup-branch"); err == nil {
		t.Fatal("expected error when checking out duplicate branch")
	}
}

func TestConfigUser(t *testing.T) {
	dir := setupTestRepo(t)
	g, err := vcs.NewGitHub("myorg/myrepo", "main", "GITHUB_TOKEN", dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := g.ConfigUser(context.Background(), dir); err != nil {
		t.Fatalf("ConfigUser: %v", err)
	}
	email := mustGit(t, dir, "config", "user.email")
	if email != "sre-agent@users.noreply.github.com" {
		t.Errorf("user.email = %q", email)
	}
	name := mustGit(t, dir, "config", "user.name")
	if name != "SRE Agent" {
		t.Errorf("user.name = %q", name)
	}
}

func TestApplyDiff(t *testing.T) {
	dir := setupTestRepo(t)
	g, err := vcs.NewGitHub("myorg/myrepo", "main", "GITHUB_TOKEN", dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := g.CheckoutBranch(context.Background(), dir, "patch-test"); err != nil {
		t.Fatal(err)
	}
	diff := "--- /dev/null\n+++ b/NEW_FILE.txt\n@@ -0,0 +1 @@\n+hello from patch\n"
	if err := g.ApplyDiff(context.Background(), dir, diff); err != nil {
		t.Fatalf("ApplyDiff: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "NEW_FILE.txt"))
	if err != nil {
		t.Fatalf("NEW_FILE.txt not created: %v", err)
	}
	if string(data) != "hello from patch\n" {
		t.Errorf("file content = %q, want %q", string(data), "hello from patch\n")
	}
	patchPath := filepath.Join(dir, ".sre-agent-fix.patch")
	if _, err := os.Stat(patchPath); !os.IsNotExist(err) {
		t.Error("patch file was not cleaned up")
	}
}

func TestCommit(t *testing.T) {
	dir := setupTestRepo(t)
	g, err := vcs.NewGitHub("myorg/myrepo", "main", "GITHUB_TOKEN", dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := g.CheckoutBranch(context.Background(), dir, "commit-test"); err != nil {
		t.Fatal(err)
	}
	if err := g.ConfigUser(context.Background(), dir); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "touched.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := g.Commit(context.Background(), dir, "test: touch file"); err != nil {
		t.Fatalf("Commit: %v", err)
	}
	msg := mustGit(t, dir, "log", "-1", "--format=%s")
	if msg != "test: touch file" {
		t.Errorf("commit message = %q", msg)
	}
}

func TestCommit_NoChanges(t *testing.T) {
	dir := setupTestRepo(t)
	g, err := vcs.NewGitHub("myorg/myrepo", "main", "GITHUB_TOKEN", dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := g.CheckoutBranch(context.Background(), dir, "nochange-test"); err != nil {
		t.Fatal(err)
	}
	if err := g.ConfigUser(context.Background(), dir); err != nil {
		t.Fatal(err)
	}
	err = g.Commit(context.Background(), dir, "empty commit")
	if err == nil {
		t.Fatal("expected error for empty commit")
	}
	if !strings.Contains(err.Error(), "no changes to commit") {
		t.Errorf("expected 'no changes to commit' error, got: %v", err)
	}
}

func TestPush_NoRemote(t *testing.T) {
	dir := setupTestRepo(t)
	g, err := vcs.NewGitHub("myorg/myrepo", "main", "GITHUB_TOKEN", dir)
	if err != nil {
		t.Fatal(err)
	}
	err = g.Push(context.Background(), dir, "some-branch")
	if err == nil {
		t.Fatal("expected push to fail with no remote")
	}
}
