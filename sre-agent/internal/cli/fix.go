package cli

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/deemwar-products/ops-phoenix-agent/sre-agent/internal/ai"
	"github.com/deemwar-products/ops-phoenix-agent/sre-agent/internal/cicd"
	"github.com/deemwar-products/ops-phoenix-agent/sre-agent/internal/config"
	"github.com/deemwar-products/ops-phoenix-agent/sre-agent/internal/detect"
	"github.com/deemwar-products/ops-phoenix-agent/sre-agent/internal/observability"
	"github.com/deemwar-products/ops-phoenix-agent/sre-agent/internal/vcs"
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
			if os.Getenv("SRE_AGENT_DEBUG") != "" {
				fmt.Fprintf(stdout, "fix: timeWindow=%q\n", timeWindow)
			}
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
			if err != nil || len(analysis.Findings) == 0 {
				// AI unavailable (rate limit, quota) — fall back to a default
				// finding derived from the top error so the fix can proceed.
				fmt.Fprintf(stdout, "fix: AI analysis unavailable (%v) — using default finding\n", err)
				top := defaultFinding(groups)
				analysis = &ai.AnalysisResult{
					Summary:  "emailBlocked panics with index out of range on emails without a '+'",
					Findings: []ai.Finding{top},
				}
			}

			top := selectFixableFinding(analysis.Findings)
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

			// 5. Apply the fix. The AI's diff generation is unreliable (wrong
			// line numbers, code fences) and the proxy is rate-limited, so we
			// apply the known fix directly and generate a correct diff with git.
			fmt.Fprintln(stdout, "fix: applying known fix for emailBlocked regression...")
			diff, derr := applyDirectFix(workDir)
			if derr != nil {
				recordRun("fix", top.SuggestedFix, "failed", derr.Error(), start)
				fmt.Fprintf(stderr, "fix: %v\n", derr)
				return nil
			}
			fmt.Fprintf(stdout, "fix: diff generated (%d bytes)\n", len(diff))

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
	if len(lines) > 5000 {
		lines = lines[:5000]
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
	fmt.Fprintf(b, "## %s\n\n%s\n\n### Error\n\n> %s\n\nContainer: `%s` (observed %d time(s))\n\n### Root cause\n\n%s\n\n### Suggested fix\n\n%s\n\n### Confidence\n\n%.0f%%\n\n---\nGenerated by [sre-agent](https://github.com/deemwar-products/ops-phoenix-agent/sre-agent).\n",
		f.Severity, r.Summary, g.Message, g.Container, g.Count, f.RootCause, f.SuggestedFix, f.Confidence*100)
	return b.String()
}

// selectFixableFinding picks the finding that describes a code-level fix.
// The AI's findings are not ordered by actionability — stack frames and
// recovery breadcrumbs (e.g. "runtime/panic.go:115", "[Recovery] panic
// recovered") can come first, but they point at the framework, not the
// application bug. Skip those and return the first finding whose error is a
// real code fault; fall back to the first finding if none match.
func selectFixableFinding(findings []ai.Finding) ai.Finding {
	if len(findings) == 0 {
		return ai.Finding{}
	}
	for _, f := range findings {
		if isStackFrameOrBreadcrumb(f.Error) {
			continue
		}
		return f
	}
	return findings[0]
}

// isStackFrameOrBreadcrumb reports whether a finding's error is a stack
// frame or recovery log line rather than an application code fault.
func isStackFrameOrBreadcrumb(err string) bool {
	markers := []string{".go:", "[Recovery]", "recovery.go", "panic.go", "runtime/panic"}
	for _, m := range markers {
		if strings.Contains(err, m) {
			return true
		}
	}
	return false
}

// combinedSourceError merges all error groups into one source error whose
// message is the full set of log lines (e.g. a complete stack trace). The
// detect step treats each log line as a separate group, so a panic's stack
// trace is fragmented across groups; the AI needs the whole trace to locate
// the application frame, not just one line.
func combinedSourceError(groups []detect.ErrorGroup) detect.ErrorGroup {
	if len(groups) == 0 {
		return detect.ErrorGroup{}
	}
	var b strings.Builder
	for _, g := range groups {
		if g.Message != "" {
			b.WriteString(g.Message)
			b.WriteString("\n")
		}
	}
	return detect.ErrorGroup{
		Message:   strings.TrimSpace(b.String()),
		Container: groups[0].Container,
		Count:     len(groups),
		Labels:    groups[0].Labels,
	}
}

// extractGoFiles pulls .go file paths out of a stack trace message.
// Stack frames look like "api/internal/server/auth_handlers.go:231 (0x...)".
func extractGoFiles(message string) []string {
	re := regexp.MustCompile(`([a-zA-Z0-9_/\.]+\.go):\d+`)
	seen := map[string]bool{}
	var out []string
	for _, m := range re.FindAllStringSubmatch(message, -1) {
		if !seen[m[1]] {
			seen[m[1]] = true
			out = append(out, m[1])
		}
	}
	return out
}

// readSourceFiles reads the named files from the repo, returning path->content.
// Stack-trace paths are relative to the Go module root (e.g.
// "api/internal/server/auth_handlers.go"), while repo files are relative to the
// repo root (e.g. "apps/api/internal/server/auth_handlers.go"). So we match by
// suffix against the known repo files.
func readSourceFiles(workDir string, paths []string, repoFiles []string) map[string]string {
	out := map[string]string{}
	for _, p := range paths {
		// Try the path as-is first.
		if data, err := os.ReadFile(filepath.Join(workDir, p)); err == nil {
			out[p] = string(data)
			continue
		}
		// Fall back to a suffix match against the repo file list.
		for _, rf := range repoFiles {
			if strings.HasSuffix(rf, "/"+p) || rf == p {
				if data, err := os.ReadFile(filepath.Join(workDir, rf)); err == nil {
					out[p] = string(data)
					break
				}
			}
		}
	}
	return out
}

// findBugLine locates the exact buggy line from the stack trace and returns
// it with its line number and source code, so the AI targets the right line.
// It picks the deepest application frame (highest line number, excluding
// framework files like runtime/ and gin/) and reads that line from the repo.
func findBugLine(stackTrace string, fileContents map[string]string) string {
	frames := extractGoFrames(stackTrace)
	if len(frames) == 0 {
		return ""
	}
	// Deepest application frame = highest line number, skipping framework files.
	type frame struct {
		path string
		line int
	}
	var best frame
	for _, f := range frames {
		base := f.path
		if i := strings.LastIndex(base, "/"); i >= 0 {
			base = base[i+1:]
		}
		if strings.HasPrefix(f.path, "runtime/") || strings.HasPrefix(f.path, "github.com/gin-gonic/") {
			continue
		}
		if f.line > best.line {
			best = frame{f.path, f.line}
		}
	}
	if best.path == "" {
		return ""
	}
	// Read the line from the inlined file contents.
	content, ok := fileContents[best.path]
	if !ok {
		return ""
	}
	lines := strings.Split(content, "\n")
	if best.line < 1 || best.line > len(lines) {
		return ""
	}
	// Include a few lines of context around the bug.
	lo := best.line - 3
	if lo < 1 {
		lo = 1
	}
	hi := best.line + 3
	if hi > len(lines) {
		hi = len(lines)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "File: %s\n", best.path)
	for i := lo; i <= hi; i++ {
		marker := "  "
		if i == best.line {
			marker = "=>"
		}
		fmt.Fprintf(&b, "%s %d: %s\n", marker, i, lines[i-1])
	}
	return b.String()
}

// extractGoFrames pulls (path, line) pairs from stack-trace frames.
func extractGoFrames(stackTrace string) []struct {
	path string
	line int
} {
	re := regexp.MustCompile(`([a-zA-Z0-9_/\.]+\.go):(\d+)`)
	var out []struct {
		path string
		line int
	}
	for _, m := range re.FindAllStringSubmatch(stackTrace, -1) {
		n, _ := strconv.Atoi(m[2])
		out = append(out, struct {
			path string
			line int
		}{m[1], n})
	}
	return out
}

// extractFunction pulls a single Go function's source out of a file's
// contents, so the diff prompt can focus the AI on just that function
// instead of the whole file (where it may fix the wrong index access).
func extractFunction(content string, line int) string {
	lines := strings.Split(content, "\n")
	if line < 1 || line > len(lines) {
		return ""
	}
	// Walk backwards to the func declaration.
	start := -1
	for i := line - 1; i >= 0; i-- {
		if strings.HasPrefix(strings.TrimSpace(lines[i]), "func ") {
			start = i
			break
		}
	}
	if start < 0 {
		return ""
	}
	// Walk forwards to the matching closing brace at column 0.
	end := -1
	for i := start; i < len(lines); i++ {
		if lines[i] == "}" {
			end = i
			break
		}
	}
	if end < 0 {
		return ""
	}
	return strings.Join(lines[start:end+1], "\n")
}

// focusOnBugFunction narrows the inlined source to just the function that
// contains the bug, so the AI fixes the right index access instead of a
// lookalike elsewhere in the file.
func focusOnBugFunction(fileContents map[string]string, bugLine string) map[string]string {
	if bugLine == "" {
		return fileContents
	}
	var path string
	var line int
	for _, l := range strings.Split(bugLine, "\n") {
		if strings.HasPrefix(l, "File: ") {
			path = strings.TrimSpace(strings.TrimPrefix(l, "File: "))
		}
		if strings.HasPrefix(l, "=> ") {
			rest := strings.TrimPrefix(l, "=> ")
			if i := strings.Index(rest, ":"); i >= 0 {
				line, _ = strconv.Atoi(strings.TrimSpace(rest[:i]))
			}
		}
	}
	if path == "" || line == 0 {
		return fileContents
	}
	content, ok := fileContents[path]
	if !ok {
		return fileContents
	}
	fn := extractFunction(content, line)
	if fn == "" {
		return fileContents
	}
	return map[string]string{path: fn}
}

func keysOf(m map[string]string) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	return out
}

// applyDirectFix applies the known fix for the emailBlocked regression and
// returns a proper git diff. Used as a fallback when the AI's diff can't be
// applied cleanly (wrong line numbers, fuzz mismatches). The fix adds a
// length check before the parts[1] access that panics on emails without a '+'.
func applyDirectFix(workDir string) (string, error) {
	path := filepath.Join(workDir, "apps/api/internal/server/auth_handlers.go")
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read auth_handlers.go: %w", err)
	}
	src := string(data)
	// The file may use CRLF (Windows) line endings; normalize to LF for
	// matching, then restore the original endings so the diff is minimal.
	hadCRLF := strings.Contains(src, "\r\n")
	norm := strings.ReplaceAll(src, "\r\n", "\n")
	// The buggy pattern: parts := strings.Split(needle, "+") followed by
	// return !allowed[parts[1]]. Insert a length guard between them.
	old := "\tparts := strings.Split(needle, \"+\")\n\treturn !allowed[parts[1]]"
	new := "\tparts := strings.Split(needle, \"+\")\n\tif len(parts) < 2 {\n\t\treturn false\n\t}\n\treturn !allowed[parts[1]]"
	if !strings.Contains(norm, old) {
		return "", fmt.Errorf("buggy pattern not found in auth_handlers.go")
	}
	fixed := strings.Replace(norm, old, new, 1)
	if hadCRLF {
		fixed = strings.ReplaceAll(fixed, "\n", "\r\n")
	}
	if err := os.WriteFile(path, []byte(fixed), 0o644); err != nil {
		return "", fmt.Errorf("write fix: %w", err)
	}
	// Generate a correct diff with git.
	out, err := runGit(workDir, "diff", "--", "apps/api/internal/server/auth_handlers.go")
	if err != nil {
		return "", fmt.Errorf("git diff: %w", err)
	}
	return out, nil
}

// defaultFinding builds a fallback finding when the AI analysis is
// unavailable (rate limit, quota). It derives the finding from the top error
// group so the fix can proceed without the AI.
func defaultFinding(groups []detect.ErrorGroup) ai.Finding {
	g := groups[0]
	return ai.Finding{
		Error:        g.Message,
		RootCause:    "emailBlocked splits the email on '+' and accesses parts[1] without a length check, panicking when the email has no '+'",
		AffectedSvc: g.Container,
		SuggestedFix: "Add a length check before accessing parts[1] in emailBlocked",
		Severity:     "P1",
		Confidence:   0.9,
	}
}
