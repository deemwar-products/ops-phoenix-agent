package vcs

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// GitHub operations via gh CLI (handles auth, PR creation, etc.).
type GitHub struct {
	Repo         string
	BaseBranch   string
	TokenEnv     string
	WorkDir      string
	ghPath       string
}

func NewGitHub(repo, baseBranch, tokenEnv, workDir string) (*GitHub, error) {
	gh, err := exec.LookPath("gh")
	if err != nil {
		return nil, fmt.Errorf("gh CLI not found — install from https://cli.github.com: %w", err)
	}
	if repo == "" {
		return nil, fmt.Errorf("github repo is empty — set it in config")
	}
	if baseBranch == "" {
		baseBranch = "main"
	}
	return &GitHub{
		Repo:       repo,
		BaseBranch: baseBranch,
		TokenEnv:   tokenEnv,
		WorkDir:    workDir,
		ghPath:     gh,
	}, nil
}

// EnsureToken verifies we have a GitHub token available.
func (g *GitHub) EnsureToken() error {
	cmd := exec.Command(g.ghPath, "auth", "status")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gh auth status failed — run `gh auth login`: %w", err)
	}
	return nil
}

// Clone clones the repo into a fresh work directory.
func (g *GitHub) Clone(ctx context.Context) (string, error) {
	workDir := g.WorkDir
	if workDir == "" {
		workDir = filepath.Join(os.TempDir(), fmt.Sprintf("sre-agent-%d", time.Now().Unix()))
	}

	if err := os.MkdirAll(workDir, 0o700); err != nil {
		return "", fmt.Errorf("create work dir: %w", err)
	}

	cloneURL := g.Repo
	if !strings.HasPrefix(cloneURL, "https://") && !strings.HasPrefix(cloneURL, "git@") {
		cloneURL = "https://github.com/" + cloneURL + ".git"
	}

	cmd := exec.CommandContext(ctx, g.ghPath, "repo", "clone", g.Repo, workDir, "--", "--branch", g.BaseBranch, "--single-branch")
	cmd.Dir = os.TempDir()
	err := cmd.Run()
	if err != nil {
		// Fall back to git clone if gh repo clone fails
		cmd = exec.CommandContext(ctx, "git", "clone", "--branch", g.BaseBranch, "--depth", "1", cloneURL, workDir)
		cmd.Dir = os.TempDir()
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("clone %s failed: %w", g.Repo, err)
		}
	}

	return workDir, nil
}

// BranchName returns a unique branch name for the fix.
func BranchName() string {
	return fmt.Sprintf("sre-agent/fix-%d", time.Now().UnixNano())
}

// CheckoutBranch creates and checks out a new branch.
func (g *GitHub) CheckoutBranch(ctx context.Context, workDir, branch string) error {
	cmd := exec.CommandContext(ctx, "git", "checkout", "-b", branch)
	cmd.Dir = workDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("create branch %s: %w\n%s", branch, err, string(out))
	}
	return nil
}

// ApplyDiff writes the patch to a temp file and applies it with git apply.
func (g *GitHub) ApplyDiff(ctx context.Context, workDir, diff string) error {
	patchPath := filepath.Join(workDir, ".sre-agent-fix.patch")
	if err := os.WriteFile(patchPath, []byte(diff), 0o600); err != nil {
		return fmt.Errorf("write patch: %w", err)
	}
	cmd := exec.CommandContext(ctx, "git", "apply", "--whitespace=nowarn", patchPath)
	cmd.Dir = workDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("apply diff: %w\n%s", err, string(out))
	}
	return os.Remove(patchPath)
}

// ConfigUser sets the git user for commits.
func (g *GitHub) ConfigUser(ctx context.Context, workDir string) error {
	for _, kv := range [][2]string{
		{"user.email", "sre-agent@users.noreply.github.com"},
		{"user.name", "SRE Agent"},
	} {
		cmd := exec.CommandContext(ctx, "git", "config", kv[0], kv[1])
		cmd.Dir = workDir
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("git config %s: %w", kv[0], err)
		}
	}
	return nil
}

// Commit stages all changes and commits them.
func (g *GitHub) Commit(ctx context.Context, workDir, message string) error {
	cmd := exec.CommandContext(ctx, "git", "add", ".")
	cmd.Dir = workDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git add: %w\n%s", err, string(out))
	}
	cmd = exec.CommandContext(ctx, "git", "commit", "-m", message, "--no-verify")
	cmd.Dir = workDir
	cmd.Env = append(os.Environ(), "GIT_EDITOR=true")
	if out, err := cmd.CombinedOutput(); err != nil {
		if strings.Contains(string(out), "nothing to commit") {
			return fmt.Errorf("no changes to commit — the diff may not have applied cleanly")
		}
		return fmt.Errorf("git commit: %w\n%s", err, string(out))
	}
	return nil
}

// Push pushes the branch to remote.
func (g *GitHub) Push(ctx context.Context, workDir, branch string) error {
	cmd := exec.CommandContext(ctx, "git", "push", "origin", branch)
	cmd.Dir = workDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git push: %w\n%s", err, string(out))
	}
	return nil
}

// CreatePR creates a PR using gh.
func (g *GitHub) CreatePR(ctx context.Context, branch, title, body string, draft, autoMerge bool) (string, error) {
	args := []string{"pr", "create", "--repo", g.Repo, "--base", g.BaseBranch, "--head", branch, "--title", title, "--body", body}
	if draft {
		args = append(args, "--draft")
	}
	cmd := exec.CommandContext(ctx, g.ghPath, args...)
	cmd.Dir = g.WorkDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("gh pr create: %w\n%s", err, string(out))
	}
	prURL := strings.TrimSpace(string(out))

	if autoMerge {
		_ = g.EnableAutoMerge(ctx, branch)
	}
	return prURL, nil
}

// EnableAutoMerge sets auto-merge on the PR for the given branch.
func (g *GitHub) EnableAutoMerge(ctx context.Context, branch string) error {
	prURL, err := g.lookupPRURL(ctx, branch)
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, g.ghPath, "pr", "merge", prURL, "--auto", "--squash")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("enable auto-merge: %w\n%s", err, string(out))
	}
	return nil
}

func (g *GitHub) lookupPRURL(ctx context.Context, branch string) (string, error) {
	cmd := exec.CommandContext(ctx, g.ghPath, "pr", "view", branch, "--repo", g.Repo, "--json", "url", "--jq", ".url")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("lookup PR: %w\n%s", err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
}
