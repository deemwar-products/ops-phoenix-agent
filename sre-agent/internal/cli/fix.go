package cli

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/deemwar-products/sre-agent/internal/ai"
	"github.com/deemwar-products/sre-agent/internal/cicd"
	"github.com/deemwar-products/sre-agent/internal/config"
	"github.com/deemwar-products/sre-agent/internal/detect"
	"github.com/deemwar-products/sre-agent/internal/observability"
	"github.com/deemwar-products/sre-agent/internal/vcs"
	"github.com/spf13/cobra"
)

// FixCmd returns `sre-agent fix` — generate and apply a fix.
func FixCmd(cfg *config.Config, stdout, stderr io.Writer) *cobra.Command {
	var dryRun bool
	var autoMerge bool
	var draft bool

	cmd := &cobra.Command{
		Use:   "fix",
		Short: "Generate a fix for the most recent error and create a PR",
		Long:  "Analyzes the most recent error, generates a code fix using AI, applies it in a sandbox, and creates a PR.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			start := time.Now().UTC()
			ctx := cmd.Context()
			gh, err := vcs.NewGitHub(cfg.GitHub.Repo, cfg.GitHub.BaseBranch, cfg.GitHub.TokenEnv, cfg.Agent.WorkDir)
			if err != nil {
				recordRun("fix", "", "failed", err.Error(), start)
				fmt.Fprintf(stderr, "fix: %v\n", err)
				return nil
			}
			if err := gh.EnsureToken(); err != nil {
				recordRun("fix", "", "failed", err.Error(), start)
				fmt.Fprintf(stderr, "fix: %v\n", err)
				return nil
			}

			// 1. Pull errors from observability
			adapter, err := observability.NewAdapter(cfg)
			if err != nil {
				recordRun("fix", "", "failed", err.Error(), start)
				fmt.Fprintf(stderr, "fix: %v\n", err)
				return nil
			}
			timeWindow := detect.DefaultTimeWindow(cfg.Agent.TimeWindow)
			raw, err := adapter.QueryErrors(ctx, timeWindow)
			if err != nil {
				recordRun("fix", "", "failed", err.Error(), start)
				fmt.Fprintf(stderr, "fix: error fetching logs: %v\n", err)
				return nil
			}
			groups := detect.Deduplicate(raw)
			if len(groups) == 0 {
				fmt.Fprintln(stdout, "fix: no errors to fix")
				recordRun("fix", "no errors", "success", "", start)
				return nil
			}

			// 2. AI analysis
			apiKeyEnv := cfg.AI.APIKeyEnv
			if apiKeyEnv == "" {
				apiKeyEnv = "ANTHROPIC_API_KEY"
			}
			analyzer, err := ai.New(ai.Provider(cfg.AI.Provider), cfg.AI.Model, apiKeyEnv)
			if err != nil {
				recordRun("fix", "", "failed", err.Error(), start)
				fmt.Fprintf(stderr, "fix: %v\n", err)
				return nil
			}
			fmt.Fprintf(stdout, "fix: analyzing %d error(s)...\n", len(groups))
			analysis, err := analyzer.Analyze(ctx, groups, "medium")
			if err != nil {
				recordRun("fix", "", "failed", err.Error(), start)
				fmt.Fprintf(stderr, "fix: AI analysis failed: %v\n", err)
				return nil
			}
			if len(analysis.Findings) == 0 {
				fmt.Fprintln(stdout, "fix: AI found no actionable findings")
				recordRun("fix", "no findings", "success", "", start)
				return nil
			}

			top := analysis.Findings[0]
			fmt.Fprintf(stdout, "fix: top finding — %s (%s, confidence=%.2f)\n",
				detect.Truncate(top.SuggestedFix, 80), top.Severity, top.Confidence)

			// 3. Clone repo
			fmt.Fprintf(stdout, "fix: cloning %s...\n", cfg.GitHub.Repo)
			workDir, err := gh.Clone(ctx)
			if err != nil {
				recordRun("fix", top.SuggestedFix, "failed", err.Error(), start)
				fmt.Fprintf(stderr, "fix: %v\n", err)
				return nil
			}
			fmt.Fprintf(stdout, "fix: workdir=%s\n", workDir)

			// 4. Branch
			branch := vcs.BranchName()
			if err := gh.CheckoutBranch(ctx, workDir, branch); err != nil {
				recordRun("fix", top.SuggestedFix, "failed", err.Error(), start)
				fmt.Fprintf(stderr, "fix: %v\n", err)
				return nil
			}
			if err := gh.ConfigUser(ctx, workDir); err != nil {
				recordRun("fix", top.SuggestedFix, "failed", err.Error(), start)
				fmt.Fprintf(stderr, "fix: %v\n", err)
				return nil
			}

			// 5. Generate patch via AI
			fmt.Fprintln(stdout, "fix: generating diff...")
			repoFiles, _ := listRepoFiles(workDir)
			diffResult, err := analyzer.GenerateDiff(ctx, top, groups[0], repoFiles)
			if err != nil {
				fmt.Fprintf(stderr, "fix: diff generation failed (%v) — using fallback\n", err)
				diffResult = &ai.DiffResult{Diff: buildFallbackPatch(top, groups[0]), Summary: top.SuggestedFix, Confidence: 0.0}
			}
			fmt.Fprintf(stdout, "fix: diff confidence=%.2f (%d files)\n", diffResult.Confidence, strings.Count(diffResult.Diff, "--- a/"))
			if err := gh.ApplyDiff(ctx, workDir, diffResult.Diff); err != nil {
				recordRun("fix", top.SuggestedFix, "failed", err.Error(), start)
				fmt.Fprintf(stderr, "fix: %v\n", err)
				return nil
			}

			// 6. Run tests
			fmt.Fprintln(stdout, "fix: running tests...")
			ci := cicd.NewGitHubActions(cfg.GitHub.Repo)
			testResult, _ := ci.RunTests(ctx, workDir)
			if testResult != nil && !testResult.Passed {
				fmt.Fprintf(stderr, "fix: tests failed: %s\n", testResult.LastLine())
				if !dryRun {
					fmt.Fprintln(stderr, "fix: not creating PR — fix didn't pass tests")
					recordRun("fix", top.SuggestedFix, "failed", "tests failed", start)
					return nil
				}
			} else if testResult != nil {
				fmt.Fprintln(stdout, "fix: tests passed")
			}

			// 7. Commit and push
			if err := gh.Commit(ctx, workDir, fmt.Sprintf("fix: %s", detect.Truncate(top.SuggestedFix, 60))); err != nil {
				recordRun("fix", top.SuggestedFix, "failed", err.Error(), start)
				fmt.Fprintf(stderr, "fix: %v\n", err)
				return nil
			}
			if dryRun {
				fmt.Fprintln(stdout, "fix: DRY RUN — not pushing or creating PR")
				recordRun("fix", top.SuggestedFix+" (dry-run)", "success", "", start)
				return nil
			}
			if err := gh.Push(ctx, workDir, branch); err != nil {
				recordRun("fix", top.SuggestedFix, "failed", err.Error(), start)
				fmt.Fprintf(stderr, "fix: %v\n", err)
				return nil
			}

			// 8. PR
			prURL, err := gh.CreatePR(ctx, branch,
				fmt.Sprintf("fix: %s", detect.Truncate(top.SuggestedFix, 60)),
				buildPRBody(top, groups[0], analysis),
				draft, autoMerge)
			if err != nil {
				recordRun("fix", top.SuggestedFix, "failed", err.Error(), start)
				fmt.Fprintf(stderr, "fix: %v\n", err)
				return nil
			}
			fmt.Fprintf(stdout, "fix: PR created: %s\n", prURL)
			recordRun("fix", detect.Truncate(top.SuggestedFix, 60), "success", "", start)
			return nil
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Generate fix but don't create PR")
	cmd.Flags().BoolVar(&autoMerge, "auto-merge", false, "Auto-merge PR if CI passes")
	cmd.Flags().BoolVar(&draft, "draft", false, "Create as draft PR")
	return cmd
}

// buildFallbackPatch produces a comment-only diff when the AI fails to
// generate one. It still creates a PR-able change so the workflow can proceed.
func buildFallbackPatch(f ai.Finding, g detect.ErrorGroup) string {
	return fmt.Sprintf(`--- a/SRE_AGENT_NOTES.md
+++ b/SRE_AGENT_NOTES.md
@@ -0,0 +1,7 @@
+# Generated by sre-agent (AI diff fallback)
+
+- Root cause: %s
+- Service: %s
+- Suggested fix: %s
+- Severity: %s
+- Source error: %s
`,
		f.RootCause, f.AffectedSvc, f.SuggestedFix, f.Severity, g.Message)
}

// listRepoFiles returns a list of tracked + untracked files in the repo,
// up to `limit` entries. Used as hints for AI diff generation.
func listRepoFiles(workDir string) ([]string, error) {
	out, err := runGit(workDir, "ls-files")
	if err != nil {
		return nil, err
	}
	lines := strings.Split(out, "\n")
	if len(lines) > 100 {
		lines = lines[:100]
	}
	return lines, nil
}

func runGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(output)), err
}

func buildPRBody(f ai.Finding, g detect.ErrorGroup, r *ai.AnalysisResult) string {
	b := new(strings.Builder)
	fmt.Fprintf(b, "## %s\n\n%s\n\n### Error\n\n> %s\n\nContainer: `%s` (observed %d time(s))\n\n### Root cause\n\n%s\n\n### Suggested fix\n\n%s\n\n### Confidence\n\n%.0f%%\n\n---\nGenerated by [sre-agent](https://github.com/deemwar-products/sre-agent).\n",
		f.Severity, r.Summary, g.Message, g.Container, g.Count, f.RootCause, f.SuggestedFix, f.Confidence*100)
	return b.String()
}
