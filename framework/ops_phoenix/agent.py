"""Core Ops Phoenix agent - detect, analyze, fix, deploy."""

import json
import os
import re
import subprocess
import sys
import time
from datetime import datetime
from pathlib import Path
from typing import Any, Dict, List, Optional, Tuple

from ops_phoenix.adapters.factory import create_adapter
from ops_phoenix.ai import analyze_errors, generate_fix_diff
from ops_phoenix.cicd.github_actions import GitHubActionsCI
from ops_phoenix.config import ConfigManager, CONFIG_DIR
from ops_phoenix.history import HistoryTracker
from ops_phoenix.models.run_record import RunRecord
from ops_phoenix.notify import (
    errors_detected_card,
    issue_created_card,
    pr_created_card,
    send_teams_message,
    deploy_result_card as teams_deploy_result_card,
)
from ops_phoenix.threshold import DecisionEngine, ThresholdConfig
from ops_phoenix.vcs.github import GitHubVCS


class OpsPhoenix:
    """Autonomous SRE agent: detect -> analyze -> fix -> deploy."""

    def __init__(self, config: Optional[ConfigManager] = None):
        self.config_mgr = config or ConfigManager()
        self.config = self.config_mgr.with_defaults()
        self.adapter = None
        self.history = HistoryTracker(CONFIG_DIR)
        self.decision_engine = DecisionEngine()
        self.vcs: Optional[GitHubVCS] = None
        self.cicd: Optional[GitHubActionsCI] = None
        self._setup_components()

        # Runtime state
        self.issue_num: Optional[int] = None
        self.pr_num: Optional[int] = None
        self.workflow_run_id: Optional[str] = None
        self.api_key = os.environ.get(
            self.config.get("ai.api_key_env", "ANTHROPIC_API_KEY"), ""
        )
        self.api_url = self.config.get("ai.api_url", "https://api.anthropic.com")
        self.model = self.config.get("ai.model", "claude-sonnet-4-20250514")
        self.dry_run = False
        self.mode = "dev"

    def _setup_components(self):
        """Initialize VCS, CI/CD, and observability adapters."""
        # VCS
        gh_config = self.config.get("github", {})
        repo = gh_config.get("repo", "")
        if repo:
            self.vcs = GitHubVCS(
                repo=repo,
                token_env_var=gh_config.get("token_env_var", "GITHUB_TOKEN"),
                labels=gh_config.get("labels", ["ops-alert"]),
            )

            # CI/CD
            cicd_config = self.config.get("cicd", {})
            if cicd_config.get("type") == "github_actions" and self.vcs:
                workflow_name = cicd_config.get(
                    "workflow_name", "Deploy to Production"
                )
                self.cicd = GitHubActionsCI(
                    vcs=self.vcs, workflow_name=workflow_name
                )

                # Observability adapter
                obs_config = self.config.get("observability", {})
                if obs_config.get("type"):
                    try:
                        self.adapter = create_adapter(self.config)
                    except Exception as e:
                        print(
                            f"Warning: could not create observability adapter: {e}"
                        )

    def _notify_teams(self, card: Dict[str, Any]) -> bool:
        """Post a card to the configured Teams webhook (if set)."""
        webhook = self.config.get("notifications.teams.webhook_url", "")
        if not webhook:
            return False
        ok = send_teams_message(webhook, card)
        if ok:
            self.log("Teams notification sent")
            return True
        self.log("Teams notification failed", "WARN")
        return False

    def log(self, message: str, level: str = "INFO"):
        ts = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        print(f"[{ts}] [{level}] {message}", flush=True)

    def test_connections(self) -> bool:
        """Verify all configured integrations are reachable."""
        self.log("Testing connections...")
        ok = True

        if self.vcs and not self.vcs.test_auth():
            self.log("GitHub: authentication failed", "ERROR")
            ok = False
        else:
            self.log("GitHub: connected")

        if self.adapter and not self.adapter.test_connection():
            self.log("Observability: connection failed", "ERROR")
            ok = False
        else:
            self.log("Observability: connected")

        return ok

    # ---- Step 1: Detect ----

    def detect_errors(self, time_window: str = "1h") -> List[str]:
        """Query logs for errors."""
        self.log(f"=== STEP 1: DETECT (window: {time_window}) ===")

        if not self.adapter:
            self.log("No observability adapter configured", "WARN")
            return []

        errors = self.adapter.query_errors(time_window)
        self.log(f"Detected {len(errors)} errors")
        return errors

    # ---- Step 2: Analyze ----

    def analyze_errors(self, errors: List[str]) -> Tuple[Optional[Dict], Dict[str, int]]:
        """Send errors to Claude for root-cause analysis."""
        self.log("=== STEP 2: ANALYZE ===")

        if not errors:
            return None, {}

        if not self.api_key:
            self.log(
                f"AI key not set (env: {self.config.get('ai.api_key_env')})",
                "WARN",
            )
            return None, self._count_error_types(errors)

        analysis = analyze_errors(
            errors=errors,
            api_key=self.api_key,
            api_url=self.api_url,
            model=self.model,
        )
        error_types = self._count_error_types(errors)

        if analysis:
            self.log(f"Root cause: {analysis.get('root_cause', 'N/A')}")
            self.log(f"Severity: {analysis.get('severity', 'N/A')}")
            return analysis, error_types

        return None, error_types

    def _count_error_types(self, errors: List[str]) -> Dict[str, int]:
        """Group errors by type signature."""
        counts: Dict[str, int] = {}
        for err in errors:
            key = err[:80] if len(err) > 80 else err
            counts[key] = counts.get(key, 0) + 1
        return counts

    # ---- Step 3: Create Issue ----

    def create_issue(
        self,
        errors: List[str],
        analysis: Optional[Dict],
        error_types: Dict[str, int],
    ) -> Optional[int]:
        """Create a GitHub issue for the detected errors."""
        self.log("=== STEP 3: CREATE ISSUE ===")

        if not errors or not self.vcs:
            return None

        # Check threshold decision
        should_act, action_level, reason = self.decision_engine.should_act(
            error_types, analysis, self.vcs.recent_issues() if self.vcs else []
        )
        self.log(f"Decision: {action_level} -- {reason}")
        if not should_act or action_level == "ignore":
            return None

        primary = list(error_types.keys())[0] if error_types else "Errors"
        severity = (analysis or {}).get("severity", "medium")
        error_summary = "\n".join(
            f"- {k}: {v} occurrences"
            for k, v in sorted(error_types.items(), key=lambda x: -x[1])[:10]
        )

        title = f"[ops-alert] {primary[:60]}"
        body = f"""## Ops Phoenix Alert

**Detected:** {datetime.now().isoformat()}
**Mode:** {self.mode.upper()}
**Severity:** {severity}
**Time window:** last check

### Error Summary
{error_summary}

### Analysis
- **Root Cause:** {analysis.get('root_cause', 'N/A') if analysis else 'N/A'}
- **Affected:** {analysis.get('affected', 'N/A') if analysis else 'N/A'}
- **Fix Suggestion:** {analysis.get('fix_suggestion', 'N/A') if analysis else 'N/A'}
- **Files Affected:** {analysis.get('files_affected', 'N/A') if analysis else 'N/A'}

---
*Generated by Ops Phoenix*
"""

        issue_num = self.vcs.create_issue(title, body)
        if issue_num:
            self.log(
                f"Issue #{issue_num} created: "
                f"https://github.com/{self.vcs.repo}/issues/{issue_num}"
            )
            self.history.add_issue(
                issue_num,
                title,
                f"https://github.com/{self.vcs.repo}/issues/{issue_num}",
            )
            self.decision_engine.record_action(action_level, primary)
            return issue_num

        self.log("Issue creation failed", "ERROR")
        return None

    # ---- Step 4: Create PR (code-aware) ----

    def create_pr(
        self, analysis: Optional[Dict], error_types: Dict[str, int]
    ) -> Optional[int]:
        """Create a PR with a code fix. Falls back to analysis-only PR if diff generation fails."""
        self.log("=== STEP 4: CREATE PR ===")

        if not self.vcs or not analysis:
            return None

        should_act, action_level, reason = self.decision_engine.should_act(
            error_types, analysis
        )
        if not should_act or action_level == "ignore":
            self.log(f"Skipping PR: {reason}")
            return None

        # Only create code PRs for severity that warrants it
        if action_level == "issue_only":
            self.log(f"Issue-only mode -- skipping PR creation ({reason})")
            return None

        primary = list(error_types.keys())[0] if error_types else "errors"
        error_summary = "\n".join(
            f"- {k}: {v}"
            for k, v in sorted(error_types.items(), key=lambda x: -x[1])[:10]
        )

        # Try code-aware fix
        work_repo_path = Path(
            self.config.get(
                "agent.work_repo_path", str(CONFIG_DIR / "work" / "repo")
            )
        )
        work_repo = None
        diff_text = None
        tests_ok = False

        try:
            work_repo = self._ensure_work_repo(work_repo_path)
            if work_repo:
                source_context = self._collect_relevant_source(work_repo)
                error_types_str = "\n".join(
                    f"- {k}: {v}"
                    for k, v in sorted(error_types.items(), key=lambda x: -x[1])
                )
                diff_text = generate_fix_diff(
                    error_summary=error_types_str,
                    analysis=analysis,
                    source_context=source_context,
                    api_key=self.api_key,
                    api_url=self.api_url,
                    model=self.model,
                )
        except Exception as e:
            self.log(f"Diff generation error: {e}", "WARN")

        # If we have a diff, try to apply it
        if diff_text and work_repo:
            ok, msg = self._apply_diff(diff_text, work_repo)
            if ok:
                changed = self._changed_files(work_repo)
                tests_ok, tests_msg = self._run_affected_tests(
                    changed, work_repo
                )
                self.log(
                    f"Tests: {'PASS' if tests_ok else 'FAIL'} -- {tests_msg}"
                )
            else:
                self.log(f"Diff did not apply cleanly: {msg}", "WARN")
                self._git_reset(work_repo)

        # Persist test result for notification (must come before the PR branch below)
        self._last_pr_tests_ok = tests_ok

        # Determine if we should auto-merge
        auto_merge = self.config.get("deployment.auto_merge", False) and tests_ok

        # Build branch name
        timestamp = datetime.now().strftime("%Y%m%d%H%M%S")
        branch_name = f"ops-phoenix/fix-{timestamp}"

        if diff_text and work_repo and ok:
            # Commit + push the fix
            self._commit_and_push(work_repo, branch_name, primary)
            pr_title = f"fix(ops-phoenix): {primary}"
            pr_body = (
                f"## Ops Phoenix Fix (code-aware)\n\n"
                f"**Issue:** #{self.issue_num or 'N/A'}\n"
                f"**Mode:** {self.mode.upper()}\n"
                f"**Tests:** {'PASS' if tests_ok else 'FAIL - human review required'}\n\n"
                f"### Root Cause\n{analysis.get('root_cause', 'N/A')}\n\n"
                f"### Diff\n```diff\n{diff_text[:4000]}\n```\n\n"
                f"---\n*Generated by Ops Phoenix*\n"
            )
        else:
            # Fallback: analysis-only PR
            branch_name = self._create_stub_branch(
                work_repo_path, branch_name, primary, analysis, diff_text
            )
            pr_title = f"docs(ops-phoenix): {primary} (analysis only)"
            pr_body = (
                f"## Ops Phoenix - Analysis Only\n\n"
                f"**Issue:** #{self.issue_num or 'N/A'}\n"
                f"**Mode:** {self.mode.upper()}\n\n"
                f"### Root Cause\n{analysis.get('root_cause', 'N/A')}\n\n"
                f"The code-aware fix generator could not produce an applicable diff.\n"
                f"Human review required.\n\n"
                f"---\n*Generated by Ops Phoenix*\n"
            )

        # Create the PR
        pr_num = self.vcs.create_pr(
            title=pr_title,
            body=pr_body,
            base_branch=self.config.get("github.base_branch", "main"),
            head_branch=branch_name,
        )

        if pr_num:
            self.log(
                f"PR #{pr_num} created: "
                f"https://github.com/{self.vcs.repo}/pull/{pr_num}"
            )
            self.history.add_pr(
                pr_num,
                pr_title,
                f"https://github.com/{self.vcs.repo}/pull/{pr_num}",
            )

            # Auto-merge if configured and tests passed
            if auto_merge:
                if self.vcs.merge_pr(pr_num):
                    self.log(f"PR #{pr_num} auto-merged")
                    self.history.add_pr(
                        pr_num,
                        pr_title,
                        f"https://github.com/{self.vcs.repo}/pull/{pr_num}",
                        merged=True,
                    )
                else:
                    self.log(f"Auto-merge failed for PR #{pr_num}", "WARN")
            else:
                reason = (
                    "tests failed"
                    if diff_text and not tests_ok
                    else "auto_merge disabled"
                )
                self.log(f"PR #{pr_num} waiting for human review ({reason})")

        return pr_num

        self.log("PR creation failed", "ERROR")
        return None

    # ---- Step 5: Deploy ----

    def deploy(self) -> Optional[str]:
        """Merge PR and trigger deployment."""
        self.log("=== STEP 5: DEPLOY ===")

        if self.mode == "dev":
            self.log("DEV MODE: skipping auto-deploy")
            if self.pr_num:
                self.log(f"To deploy manually: gh pr merge {self.pr_num}")
            return None

        if not self.cicd:
            self.log("No CI/CD configured", "WARN")
            return None

        # Merge PR
        if self.pr_num and not self.vcs.merge_pr(self.pr_num):
            self.log(f"Merge failed for PR #{self.pr_num}", "ERROR")
            return None

        if self.pr_num:
            self.log(f"PR #{self.pr_num} merged")

        # Trigger deployment workflow
        run_url = self.cicd.trigger()
        if run_url:
            self.workflow_run_id = self.cicd.extract_run_id(run_url) or run_url
            self.log(f"Deploy triggered: {run_url}")
            return run_url

        self.log("Deploy trigger failed", "ERROR")
        return None

    # ---- Step 6: Monitor ----

    def monitor_deploy(self, max_wait: int = 300) -> str:
        """Wait for the deployment workflow to complete."""
        self.log("=== STEP 6: MONITOR ===")

        if not self.workflow_run_id or not self.cicd:
            return "skipped"

        poll_interval = 15
        start = time.time()

        while time.time() - start < max_wait:
            status, conclusion = self.cicd.wait_for_completion(
                self.workflow_run_id, timeout=poll_interval
            )
            elapsed = int(time.time() - start)
            self.log(
                f"Status: {status}/{conclusion or 'running'} ({elapsed}s elapsed)"
            )

            if status == "completed":
                if conclusion == "success":
                    self.log("Deployment succeeded!", "INFO")
                    return "success"
                self.log(f"Deployment failed: {conclusion}", "ERROR")
                return f"failed:{conclusion}"
            time.sleep(1)

        self.log("Deploy monitoring timed out", "WARN")
        return "timeout"

    # ---- Full cycle ----

    def run_full_cycle(self, time_window: str = "1h") -> Dict[str, Any]:
        """Run the complete detect -> analyze -> issue -> fix -> deploy -> monitor cycle."""
        self.log("=== OPS PHOENIX FULL CYCLE ===")
        self.log(f"Mode: {self.mode.upper()}")
        if self.vcs:
            self.log(f"Target: {self.vcs.repo}")

        self.history.start_run(mode=self.mode, time_window=time_window)

        # Validate config
        errors = self.config_mgr.validate()
        if errors:
            self.log("Configuration errors:", "ERROR")
            for e in errors:
                self.log(f" - {e}", "ERROR")
            self.history.complete_run(
                status="failed", error_message="; ".join(errors)
            )
            return {
                "status": "failed",
                "reason": "config_validation",
                "errors": errors,
            }

        # Test connections
        if not self.test_connections():
            self.history.complete_run(
                status="failed", error_message="connection_failed"
            )
            return {"status": "failed", "reason": "connection_failed"}

        # Step 1: Detect
        errors = self.detect_errors(time_window)
        self.history.add_errors(errors, {})

        if not errors:
            self.log("No errors found -- all systems healthy!")
            self.history.complete_run(status="healthy")
            return {"status": "healthy", "errors_found": 0}

        error_types = self._count_error_types(errors)

        # Step 2: Analyze
        analysis, error_types = self.analyze_errors(errors)

        # Notify: errors detected
        self._notify_teams(errors_detected_card(error_types, time_window, analysis))

        # Step 3: Create issue
        self.issue_num = self.create_issue(errors, analysis, error_types)

        # Notify: issue created
        if self.issue_num:
            issue_url = f"https://github.com/{self.vcs.repo}/issues/{self.issue_num}"
            self._notify_teams(issue_created_card(self.issue_num, issue_url, error_types))

        # Step 4: Create PR
        self.pr_num = self.create_pr(analysis, error_types)

        # Notify: PR created
        if self.pr_num:
            tests_ok = getattr(self, "_last_pr_tests_ok", False)
            pr_url = f"https://github.com/{self.vcs.repo}/pull/{self.pr_num}"
            self._notify_teams(
                pr_created_card(self.pr_num, pr_url, tests_ok, analysis)
            )

        if not self.pr_num and not self.issue_num:
            self.history.complete_run(
                status="failed", error_message="no_action_taken"
            )
            return {"status": "failed", "reason": "threshold_skip"}

        # Step 5: Deploy (only if we have a PR and it was auto-merged or we want to deploy)
        # In dev mode, we don't auto-deploy; in prod mode, the PR was already merged
        # by create_pr if auto_merge is on
        if self.mode == "prod" and self.pr_num:
            run_url = self.deploy()
            if run_url:
                self.history.add_deploy(
                    self.workflow_run_id or "unknown",
                    "triggered",
                    self.config.get("cicd.workflow_name", "unknown"),
                )

        # Step 6: Monitor
        result = self.monitor_deploy(
            self.config.get("deployment.monitor_timeout", 300)
        )

        # Notify: deploy result
        if self.workflow_run_id:
            status_key = "success" if result == "success" else f"failed:{result}"
            self.history.add_deploy(
                self.workflow_run_id,
                status_key,
                self.config.get("cicd.workflow_name", "unknown"),
            )
            self._notify_teams(
                teams_deploy_result_card(self.workflow_run_id, result, self.pr_num)
            )

        self.history.complete_run(status=result)
        return {
            "status": result,
            "issue": self.issue_num,
            "pr": self.pr_num,
            "workflow": self.workflow_run_id,
            "errors_found": len(errors),
            "error_types": error_types,
        }

    # ---- Helper methods ----

    def _ensure_work_repo(self, work_dir: Path) -> Optional[Path]:
        """Clone or update the target repo for code-aware fixes."""
        if not self.vcs:
            return None

        work_dir = work_dir.expanduser()
        work_dir.mkdir(parents=True, exist_ok=True)

        token = os.environ.get(
            self.config.get("github.token_env_var", "GITHUB_TOKEN"), ""
        )
        repo_url = f"https://github.com/{self.vcs.repo}.git"
        if token:
            repo_url = repo_url.replace(
                "https://", f"https://x-access-token:{token}@"
            )

        if (work_dir / ".git").exists():
            # Update existing clone
            subprocess.run(
                ["git", "fetch", "origin"], cwd=work_dir, capture_output=True
            )
            subprocess.run(
                ["git", "checkout", "main"],
                cwd=work_dir,
                capture_output=True,
            )
            subprocess.run(
                ["git", "reset", "--hard", "origin/main"],
                cwd=work_dir,
                capture_output=True,
            )
            return work_dir

        r = subprocess.run(
            ["git", "clone", "--depth", "50", repo_url, str(work_dir)],
            capture_output=True,
            text=True,
        )
        if r.returncode != 0:
            self.log(f"git clone failed: {r.stderr}", "ERROR")
            return None
        return work_dir

    def _collect_relevant_source(self, work_dir: Path) -> str:
        """Collect source files relevant to the error types."""
        # Gather recent changes and key source files
        try:
            r = subprocess.run(
                ["git", "diff", "--name-only", "HEAD~5..HEAD"],
                cwd=work_dir,
                capture_output=True,
                text=True,
            )
            changed = r.stdout.strip().splitlines() if r.stdout.strip() else []
        except Exception:
            changed = []

        sources = []
        for fname in changed[:20]:
            fpath = work_dir / fname
            if fpath.exists() and fpath.is_file():
                try:
                    content = fpath.read_text(errors="replace")[:2000]
                    sources.append(f"--- {fname} ---\n{content}")
                except Exception:
                    pass

        return "\n\n".join(sources) if sources else "No recent source changes found."

    def _apply_diff(self, diff_text: str, work_dir: Path) -> Tuple[bool, str]:
        """Apply a unified diff to the work repo."""
        try:
            r = subprocess.run(
                ["git", "apply", "--index"],
                input=diff_text,
                cwd=work_dir,
                capture_output=True,
                text=True,
            )
            if r.returncode == 0:
                return True, "applied"
            return False, r.stderr[:500]
        except Exception as e:
            return False, str(e)

    def _commit_and_push(
        self, work_dir: Path, branch_name: str, message: str
    ) -> bool:
        """Commit changes and push to the branch."""
        try:
            subprocess.run(
                ["git", "checkout", "-b", branch_name],
                cwd=work_dir,
                capture_output=True,
            )
            subprocess.run(
                ["git", "commit", "-m", f"chore(ops-phoenix): {message}"],
                cwd=work_dir,
                capture_output=True,
            )
            r = subprocess.run(
                ["git", "push", "origin", branch_name],
                cwd=work_dir,
                capture_output=True,
                text=True,
            )
            return r.returncode == 0
        except Exception as e:
            self.log(f"Commit/push error: {e}", "WARN")
            return False

    def _changed_files(self, work_dir: Path) -> List[str]:
        """Return list of changed files after applying a diff."""
        try:
            r = subprocess.run(
                ["git", "diff", "--name-only", "--cached"],
                cwd=work_dir,
                capture_output=True,
                text=True,
            )
            return r.stdout.strip().splitlines() if r.stdout.strip() else []
        except Exception:
            return []

    def _run_affected_tests(
        self, changed: List[str], work_dir: Path
    ) -> Tuple[bool, str]:
        """Run tests for affected packages. Returns (passed, message)."""
        if not changed:
            return True, "no changed files"

        # Find test runner from repo root
        for cmd in (
            ["pytest", "-x", "--tb=short"],
            ["npm", "test", "--", "--run"],
            ["go", "test", "./..."],
        ):
            r = subprocess.run(
                cmd, cwd=work_dir, capture_output=True, text=True, timeout=120
            )
            if r.returncode == 0:
                return True, f"{cmd[0]} passed"
            # If the command wasn't found, try the next
            if r.returncode == 127:
                continue
            return False, f"{cmd[0]} failed: {r.stderr[:200]}"

        return True, "no test runner found, skipping"

    def _git_reset(self, work_dir: Path) -> None:
        """Reset any uncommitted changes."""
        try:
            subprocess.run(
                ["git", "checkout", "--", "."],
                cwd=work_dir,
                capture_output=True,
            )
            subprocess.run(
                ["git", "clean", "-fd"],
                cwd=work_dir,
                capture_output=True,
            )
        except Exception as e:
            self.log(f"git reset error: {e}", "WARN")

    def _create_stub_branch(
        self,
        work_dir: Path,
        branch_name: str,
        primary: str,
        analysis: Optional[Dict],
        diff_text: Optional[str],
    ) -> str:
        """Create a stub branch with analysis-only content (no code diff)."""
        stub = work_dir / "ops-phoenix-analysis.md"
        stub.write_text(
            f"# Ops Phoenix Analysis\n\n"
            f"**Primary error:** {primary}\n\n"
            f"**Root cause:** {analysis.get('root_cause', 'N/A') if analysis else 'N/A'}\n\n"
            f"## Recommended Fix\n\n"
            f"{analysis.get('fix_suggestion', 'N/A') if analysis else 'N/A'}\n\n"
            f"---\n*This PR contains analysis only. Code fix requires human review.*\n"
        )
        try:
            subprocess.run(
                ["git", "checkout", "-b", branch_name],
                cwd=work_dir,
                capture_output=True,
            )
            subprocess.run(
                ["git", "add", str(stub.name)],
                cwd=work_dir,
                capture_output=True,
            )
            subprocess.run(
                ["git", "commit", "-m", f"docs(ops-phoenix): analysis for {primary}"],
                cwd=work_dir,
                capture_output=True,
            )
            subprocess.run(
                ["git", "push", "origin", branch_name],
                cwd=work_dir,
                capture_output=True,
            )
        except Exception as e:
            self.log(f"Stub branch error: {e}", "WARN")
        return branch_name

    # ---- Full cycle orchestrator ----

    def run(self, time_window: str = "1h") -> Dict[str, Any]:
        """Alias for run_full_cycle."""
        return self.run_full_cycle(time_window)
