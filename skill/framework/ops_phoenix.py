#!/usr/bin/env python3
"""
Ops Phoenix - Main Agent

Comprehensive ops agent with configurable adapters.
Will ask configuration questions if setup is incomplete.

Usage:
python3 ops_phoenix.py # Interactive (asks if incomplete)
python3 ops_phoenix.py --setup # Run configuration wizard
python3 ops_phoenix.py --dry-run # Detect errors only
python3 ops_phoenix.py --env prod --full-cycle # Full cycle
python3 ops_phoenix.py --config # Show current config
"""

import json
import os
import subprocess
import sys
import time
import re
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Optional

# Add framework to path
sys.path.insert(0, str(Path(__file__).parent))

from adapters.observability import create_adapter
from history import HistoryTracker

# Config file location
CONFIG_DIR = Path.home() / ".ops-phoenix"
CONFIG_FILE = CONFIG_DIR / "config.json"

# Defaults
TIME_WINDOW = "1h"
ENV_MODE = "dev"
FULL_CYCLE = False
DRY_RUN = False


class OpsPhoenix:
  def __init__(self, config_path=None):
    self.config = {}
    self.adapter = None
    self.api_key = None
    self.api_url = None
    self.issue_num = None
    self.pr_num = None
    self.workflow_run_id = None
    self.history = HistoryTracker()
    self._load_env()
    # config_path arg > OPS_PHOENIX_CONFIG env > default ~/.ops-phoenix/config.json
    env_path = os.environ.get("OPS_PHOENIX_CONFIG")
    effective = config_path or (Path(env_path) if env_path else None)
    self.load_config(effective)

  def log(self, message, level="INFO"):
    ts = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    print(f"[{ts}] [{level}] {message}")

  def load_config(self, config_path=None) -> Dict:
    path = config_path or CONFIG_FILE
    if path.exists():
      self.config = json.loads(path.read_text())
      return self.config
    return self.config

  def is_configured(self) -> bool:
    required = ["github.repo", "github.token"]
    for field in required:
      parts = field.split(".")
      val = self.config
      for p in parts:
        val = val.get(p, {}) if isinstance(val, dict) else None
        if val is None:
          return False
          return True

  def needs_setup(self):
    return not self.is_configured()

  def run_setup_non_interactive(self):
    grafana_url = os.environ.get("GRAFANA_URL", "")
    loki_url = os.environ.get("LOKI_URL", "")
    grafana_token = os.environ.get("GRAFANA_TOKEN", "")
    github_repo = os.environ.get("GITHUB_REPO", "")
    github_token = os.environ.get("GH_TOKEN", "")
    anthropic_key = os.environ.get("ANTHROPIC_API_KEY", "")

    if not any([grafana_url, loki_url, github_repo]):
      return False

      config = {
      "github": {
      "repo": github_repo or "deemwar-products/reqsume",
      "token": github_token or "",
      },
      "observability": {},
      }

      if grafana_url:
        config["observability"]["grafana"] = {
        "url": grafana_url,
        "token": grafana_token,
        }
        if loki_url:
          config["observability"]["loki"] = {"url": loki_url}
          if anthropic_key:
            config["anthropic"] = {
            "api_key": anthropic_key,
            "api_url": os.environ.get("ANTHROPIC_API_URL", "https://api.anthropic.com"),
            "model": os.environ.get("ANTHROPIC_MODEL", "claude-haiku-4-5"),
            }

            self.config = config
            self.save_config()
            self.log("Configuration loaded from environment variables")
            return True

  def save_config(self):
    CONFIG_DIR.mkdir(parents=True, exist_ok=True)
    CONFIG_FILE.write_text(json.dumps(self.config, indent=2))
    self.log(f"Configuration saved to {CONFIG_FILE}")

  def run_setup(self):
    if not sys.stdin.isatty():
        self.log(
        "refusing to run interactive setup: stdin is not a TTY. "
        "Use --config <path> or set OPS_PHOENIX_CONFIG for non-interactive runs."
        )
        raise SystemExit(2)

    print("\n--- Ops Phoenix Setup ---\n")

    github_repo = input("GitHub repo (e.g. deemwar-products/reqsume): ").strip()
    github_token = input("GitHub token (PAT with repo scope): ").strip()

    print("\nObservability (leave blank to skip)")
    grafana_url = input("Grafana URL: ").strip()
    grafana_token = input("Grafana API token: ").strip()
    loki_url = input("Loki URL: ").strip()

    print("\nAI Analysis (Anthropic Claude)")
    anthropic_key = input("Anthropic API key (blank to skip): ").strip()

    self.config = {
    "github": {
    "repo": github_repo,
    "token": github_token,
    },
    "observability": {},
    }

    if grafana_url:
      self.config["observability"]["grafana"] = {
      "url": grafana_url,
      "token": grafana_token,
      }
      if loki_url:
        self.config["observability"]["loki"] = {"url": loki_url}
        if anthropic_key:
          self.config["anthropic"] = {
          "api_key": anthropic_key,
          "api_url": os.environ.get("ANTHROPIC_API_URL", "https://api.anthropic.com"),
          "model": os.environ.get("ANTHROPIC_MODEL", "claude-haiku-4-5"),
          }

          self.save_config()
          print("\nSetup complete!")

  def test_connections(self) -> bool:
    self.log("Testing connections...")
    ok = True

    if self.config.get("github"):
      result = subprocess.run(
      ["gh", "auth", "status"],
      capture_output=True,
      text=True,
      env={**os.environ, "GH_TOKEN": self.config["github"]["token"]},
      )
      if result.returncode == 0:
        self.log("GitHub connection OK")
      else:
        self.log("GitHub connection failed", "ERROR")
        ok = False

        obs = self.config.get("observability", {})
        if obs:
          self.adapter = create_adapter(obs)
          if self.adapter and self.adapter.test():
            self.log("Observability connection OK")
          else:
            self.log("Observability connection failed", "ERROR")
            ok = False

            return ok

  def detect_errors(self, time_window: str = "1h") -> List[str]:
    self.log(f"\n=== STEP 1: DETECT ERRORS (window: {time_window}) ===")

    if not self.adapter:
      self.log("No observability adapter configured", "WARN")
      return []

      errors = self.adapter.get_errors(time_window)
      self.log(f"Detected {len(errors)} errors")

      return errors

  def count_error_types(self, errors: List[str]) -> Dict[str, int]:
    counts = {}
    for err in errors:
      match = re.search(r"([A-Z]{3,}|\d{3})", err)
      if match:
        key = match.group(1)
      else:
        key = "unknown"
        counts[key] = counts.get(key, 0) + 1
        return counts

  def analyze_errors(self, errors: List[str]) -> tuple:
    self.log(f"\n=== STEP 2: ANALYZE ERRORS ===")

    if not errors:
      return None, {}

      if not self.api_key:
        self.log("ANTHROPIC_API_KEY not set, skipping analysis", "WARN")
        return None, self.count_error_types(errors)

        error_types = self.count_error_types(errors)
        error_summary = "\n".join([f"- {k}: {v} occurrences" for k, v in error_types.items()])

        prompt = f"""You are an SRE analyzing production errors.
        
        ## Error Summary
        {error_summary}
        
        Provide:
        1. ROOT_CAUSE: One sentence explaining root cause
        2. AFFECTED: What functionality is affected
        3. SEVERITY: critical/high/medium/low
        4. FIX_SUGGESTION: How to fix this
        5. FILES_AFFECTED: Which files need changes
        
        Format:
        ROOT_CAUSE: ...
        AFFECTED: ...
        SEVERITY: ...
        FIX_SUGGESTION: ...
        FILES_AFFECTED: ..."""

        result = subprocess.run(
        [
        "curl",
        "-sS",
        "-X",
        "POST",
        f"{self.api_url}/v1/messages",
        "-H",
        f"x-api-key: {self.api_key}",
        "-H",
        "anthropic-version: 2023-06-01",
        "-H",
        "content-type: application/json",
        "--data",
        json.dumps(
        {
        "model": self.config.get("anthropic", {}).get("model", "claude-haiku-4-5"),
        "max_tokens": 1024,
        "messages": [{"role": "user", "content": prompt}],
        }
        ),
        ],
        capture_output=True,
        text=True,
        )

        if result.returncode == 0:
          try:
            response = json.loads(result.stdout)
            content = response.get("content", [])
            if content:
              analysis = self.parse_analysis(content[0].get("text", ""))
              self.log(f"Analysis complete: {analysis.get('root_cause', 'N/A')}")
              return analysis, error_types
          except (json.JSONDecodeError, KeyError) as e:
            self.log(f"Failed to parse analysis: {e}", "ERROR")

            return None, error_types

  def parse_analysis(self, text: str) -> Dict:
    analysis = {}
    for line in text.split("\n"):
      if line.startswith("ROOT_CAUSE:"):
        analysis["root_cause"] = line.split(":", 1)[1].strip()
      elif line.startswith("AFFECTED:"):
        analysis["affected"] = line.split(":", 1)[1].strip()
      elif line.startswith("SEVERITY:"):
        analysis["severity"] = line.split(":", 1)[1].strip()
      elif line.startswith("FIX_SUGGESTION:"):
        analysis["fix_suggestion"] = line.split(":", 1)[1].strip()
      elif line.startswith("FILES_AFFECTED:"):
        analysis["files_affected"] = line.split(":", 1)[1].strip()
        return analysis

  def create_issue(self, errors: List[str], analysis: Dict, error_types: Dict) -> Optional[str]:
    self.log(f"\n=== STEP 3: CREATE ISSUE ===")

    if not errors:
      return None

      error_summary = "\n".join([f"- {k}: {v}" for k, v in error_types.items()])
      issue_title = f"[ops-alert] {list(error_types.keys())[0] if error_types else 'Errors detected'}"

      issue_body = f"""## Ops Phoenix Alert
      
      **Time:** {datetime.now().isoformat()}
      **Mode:** {ENV_MODE.upper()}
      **Window:** {TIME_WINDOW}
      
      ### Error Summary
      {error_summary}
      
      ### Analysis
      - **Root Cause:** {analysis.get('root_cause', 'N/A') if analysis else 'N/A'}
      - **Affected:** {analysis.get('affected', 'N/A') if analysis else 'N/A'}
      - **Severity:** {analysis.get('severity', 'medium') if analysis else 'medium'}
      
      ### Fix Suggestion
      {analysis.get('fix_suggestion', 'N/A') if analysis else 'N/A'}
      
      ### Files Affected
      {analysis.get('files_affected', 'N/A') if analysis else 'N/A'}
      
      ---
      *Generated by Ops Phoenix*
      """

      cmd = ["gh", "issue", "create", "--title", issue_title, "--body", issue_body, "--label", "sre-alert"]
      if self.config.get("github", {}).get("repo"):
        cmd.extend(["--repo", self.config["github"]["repo"]])

        result = subprocess.run(cmd, capture_output=True, text=True)

        if result.returncode == 0:
          issue_url = result.stdout.strip()
          match = re.search(r"/issues/(\d+)", issue_url)
          self.issue_num = match.group(1) if match else None
          self.log(f"Issue created: {issue_url}")
          return self.issue_num

          self.log(f"Issue creation failed: {result.stderr}", "ERROR")
          return None

  def create_pr(self, analysis: Dict, error_types: Dict) -> Optional[str]:
    """Step 5: Create PR with fix. Option A: code-aware fix with safety interlock."""
    self.log(f"\n=== STEP 5: CREATE PR (code-aware) ===")

    # 5a. Ensure work repo exists
    try:
      work_repo = self.ensure_work_repo()
    except Exception as e:
      self.log(f"Work repo setup failed: {e}", "ERROR")
      return self._create_stub_pr(analysis, error_types)

      # 5b. Sync to main, branch off
      timestamp = datetime.now().strftime("%Y%m%d%H%M%S")
      branch_name = f"ops-phoenix/fix-{timestamp}"
      self._run(["git", "checkout", "main"], cwd=work_repo)
      self._run(["git", "pull", "--ff-only", "origin", "main"], cwd=work_repo)
      self._run(["git", "checkout", "-b", branch_name], cwd=work_repo)
      self.log(f"Created branch: {branch_name}")

      # 5c. Pull relevant source snippets from stack traces
      recent_errors = self.history.get_recent_errors() if hasattr(self, "history") else []
      relevant_source = self.collect_relevant_source(recent_errors, work_repo)

      # 5d. Ask Claude for a unified diff
      diff_text = self.generate_diff(analysis, error_types, relevant_source)
      if not diff_text:
        self.log("Claude did not return a usable diff; falling back to stub PR", "WARN")
        self._run(["git", "checkout", "--", "."], cwd=work_repo)
        return self._create_stub_pr(analysis, error_types, work_repo, branch_name)

        # 5e. Apply the diff (check first, then real)
        ok, msg = self.apply_diff(diff_text, work_repo)
        if not ok:
          self.log(f"Diff does not apply cleanly: {msg}", "WARN")
          self._run(["git", "checkout", "--", "."], cwd=work_repo)
          return self._create_stub_pr(analysis, error_types, work_repo, branch_name, diff_text)

          # 5f. Run scoped tests BEFORE we push or merge. Safety interlock.
          changed = self._changed_files(work_repo)
          tests_ok, tests_msg = self.run_affected_tests(changed, work_repo)
          self.log(f"Tests: {'PASS' if tests_ok else 'FAIL'} - {tests_msg}")

          # 5g. Commit + push
          error_summary = list(error_types.keys())[0] if error_types else "Unknown"
          self._run(["git", "add", "-A"], cwd=work_repo)
          commit = subprocess.run(
          ["git", "commit", "-m", f"fix(ops-phoenix): {error_summary}\n\nGenerated by Ops Phoenix"],
          capture_output=True,
          text=True,
          cwd=work_repo,
          )
          if commit.returncode != 0:
            self.log(f"Commit failed: {commit.stderr}", "ERROR")
            return None
            self._run(["git", "push", "-u", "origin", branch_name], cwd=work_repo, check=False)

            # 5h. Open PR
            pr_title = f"fix(ops-phoenix): {error_summary}"
            pr_body = (
            f"## Ops Phoenix Fix (code-aware)\n\n"
            f"**Issue:** #{self.issue_num or 'N/A'}\n"
            f"**Mode:** {ENV_MODE}\n"
            f"**Tests:** {'PASS' if tests_ok else 'FAIL - human review required'}\n"
            f"**Files changed:** {len(changed)}\n\n"
            f"### Root Cause\n{analysis.get('root_cause', 'N/A') if analysis else 'N/A'}\n\n"
            f"### Diff summary\n```diff\n{diff_text[:2000]}\n```\n\n"
            "---\n*Generated by Ops Phoenix*\n"
            )
            result = subprocess.run(
            ["gh", "pr", "create", "--title", pr_title, "--body", pr_body, "--label", "sre-alert"],
            capture_output=True,
            text=True,
            )
            if result.returncode != 0:
              self.log(f"PR creation failed: {result.stderr}", "ERROR")
              return None

              pr_url = result.stdout.strip()
              match = re.search(r"/pull/(\d+)", pr_url)
              self.pr_num = match.group(1) if match else None
              self.log(f"PR created: {pr_url}")

              # 5i. Auto-merge ONLY if tests passed AND config allows it.
              if self.config.get("deployment", {}).get("auto_merge", False) and tests_ok:
                self.log("Tests passed and auto_merge enabled - merging")
                merge = subprocess.run(
                ["gh", "pr", "merge", self.pr_num, "--squash", "--delete-branch"],
                capture_output=True,
                text=True,
                )
                if merge.returncode == 0:
                  self.log(f"PR #{self.pr_num} auto-merged")
                else:
                  self.log(f"Auto-merge failed: {merge.stderr}", "ERROR")
              else:
                  reason = "tests failed" if not tests_ok else "auto_merge disabled"
                  self.log(f"PR opened for human review ({reason})")
                  return self.pr_num

  def ensure_work_repo(self) -> Path:
    """Clone the reqsume repo to a stable work dir; reuse on subsequent runs."""
    work_dir = Path(
    self.config.get("ops_phoenix", {}).get("work_repo_path", "/opt/ops-phoenix/work/repo")
    )
    repo_url = f"https://github.com/{self.config['github']['repo']}.git"
    token = os.environ.get("GH_TOKEN", "")
    authed_url = repo_url if not token else repo_url.replace("https://", f"https://x-access-token:{token}@")
    work_dir.parent.mkdir(parents=True, exist_ok=True)
    if not (work_dir / ".git").exists():
      self.log(f"Cloning {self.config['github']['repo']} into {work_dir}")
      r = subprocess.run(
      ["git", "clone", "--depth", "50", authed_url, str(work_dir)],
      capture_output=True,
      text=True,
      )
      if r.returncode != 0:
        raise RuntimeError(f"git clone failed: {r.stderr}")
        return work_dir

  def collect_relevant_source(self, errors: List[Dict], work_repo: Path) -> str:
    """Read source files mentioned in error stack traces. Cap at 20 files / 200KB."""
    if not errors:
      return ""
      path_re = re.compile(r"((?:apps|framework|infra)/\S+\.(?:go|ts|tsx|js|py)):(\d+)")
      paths = set()
      for err in errors:
        raw = err.get("raw", "") if isinstance(err, dict) else str(err)
        for m in path_re.finditer(raw):
          paths.add((m.group(1), int(m.group(2))))
          if not paths:
            return ""
            chunks, total = [], 0
            for path, line in sorted(paths)[:20]:
              full = work_repo / path
              if not full.exists() or not full.is_file():
                continue
                try:
                  text = full.read_text(errors="replace")
                except Exception:
                  continue
                  lines = text.splitlines()
                  start = max(0, line - 25)
                  end = min(len(lines), line + 25)
                  snippet = "\n".join(lines[start:end])
                  chunks.append(f"=== {path}:{line} ===\n{snippet}")
                  total += len(snippet)
                  if total > 200_000:
                    break
                    return "\n\n".join(chunks)

  def generate_diff(self, analysis: Dict, error_types: Dict, relevant_source: str) -> Optional[str]:
    """POST to Claude and parse a unified diff out of the response."""
    if not self.api_key:
      return None
      error_summary = "\n".join([f"- {k}: {v}" for k, v in (error_types or {}).items()])
      anthropic_cfg = self.config.get("anthropic", {})
      model = anthropic_cfg.get("model", "claude-haiku-4-5")
      max_tokens = anthropic_cfg.get("max_tokens_for_diff", 4096)
      timeout = str(anthropic_cfg.get("request_timeout_seconds", 60))

      system = (
      "You are an SRE fixing production errors in the reqsume codebase. "
      "You will be given error logs and the relevant source files. "
      "Output ONLY a unified diff that fixes the root cause. "
      "No prose, no markdown fences wrapping the diff - just the raw diff. "
      "If you cannot produce a confident fix, output the single line: NO_DIFF"
      )
      user = (
      f"## Error summary\n{error_summary}\n\n"
      f"## Root cause analysis\n{analysis.get('root_cause', 'N/A') if analysis else 'N/A'}\n\n"
      f"## Relevant source files\n{relevant_source if relevant_source else '(no source paths found in stack traces)'}\n\n"
      "Generate a unified diff (--- a/path, +++ b/path, @@ hunks @@) that fixes the root cause.\n"
      "If the fix is risky or unclear, output NO_DIFF on a single line and nothing else.\n"
      )
      body = json.dumps({"model": model, "max_tokens": max_tokens, "system": system, "messages": [{"role": "user", "content": user}]})
      result = subprocess.run(
      [
      "curl",
      "-sS",
      "-X",
      "POST",
      f"{self.api_url}/v1/messages",
      "-H",
      f"x-api-key: {self.api_key}",
      "-H",
      "anthropic-version: 2023-06-01",
      "-H",
      "content-type: application/json",
      "--max-time",
      timeout,
      "--data",
      body,
      ],
      capture_output=True,
      text=True,
      )
      if result.returncode != 0:
        self.log(f"Claude call failed: {result.stderr}", "ERROR")
        return None
        try:
          resp = json.loads(result.stdout)
        except json.JSONDecodeError:
          self.log(f"Claude returned non-JSON: {result.stdout[:200]}", "ERROR")
          return None
          content = resp.get("content", [])
          if not content:
            return None
            text = content[0].get("text", "").strip()
            if "NO_DIFF" in text and len(text) < 50:
              return None
              text = re.sub(r"^```diff\s*", "", text)
              text = re.sub(r"```\s*$", "", text)
              return text

  def apply_diff(self, diff_text: str, work_repo: Path) -> tuple:
    """Verify the diff applies cleanly, then apply for real. Returns (ok, msg)."""
    proc = subprocess.run(
    ["git", "apply", "--check", "-"],
    input=diff_text,
    capture_output=True,
    text=True,
    cwd=work_repo,
    )
    if proc.returncode != 0:
      return False, proc.stderr.strip()
      proc = subprocess.run(
      ["git", "apply", "-"],
      input=diff_text,
      capture_output=True,
      text=True,
      cwd=work_repo,
      )
      if proc.returncode != 0:
        return False, f"apply failed: {proc.stderr.strip()}"
        return True, "applied"

  def run_affected_tests(self, changed_paths: List[str], work_repo: Path) -> tuple:
    """Run go test / bun test scoped to the changed files."""
    if not changed_paths:
      return True, "no files changed"
      go_dirs, ts_dirs = set(), set()
      for p in changed_paths:
        ps = str(p)
        if ps.endswith(".go"):
          go_dirs.add("apps/api")
        elif ps.endswith((".ts", ".tsx", ".js", ".jsx")):
          if ps.startswith("apps/ui/"):
            ts_dirs.add("apps/ui")
          elif ps.startswith("apps/home/"):
            ts_dirs.add("apps/home")
          elif ps.startswith("apps/extension/"):
            ts_dirs.add("apps/extension")
            results = []
            all_ok = True
            for d in sorted(go_dirs):
              ok, msg = self._run_go_tests(d, work_repo)
              results.append(f"go[{d}]: {'OK' if ok else 'FAIL'} {msg}")
              all_ok = all_ok and ok
              for d in sorted(ts_dirs):
                ok, msg = self._run_ts_tests(d, work_repo)
                results.append(f"ts[{d}]: {'OK' if ok else 'FAIL'} {msg}")
                all_ok = all_ok and ok
                return all_ok, "; ".join(results) or "no testable files"

  def _run_go_tests(self, rel_dir: str, work_repo: Path) -> tuple:
    cwd = work_repo / rel_dir
    if not (cwd / "go.mod").exists():
      return True, f"no go.mod in {rel_dir}"
      r = subprocess.run(
      ["go", "test", "./...", "-count=1", "-timeout", "120s"],
      capture_output=True,
      text=True,
      cwd=cwd,
      )
      if r.returncode == 0:
        return True, "all green"
        tail = "\n".join(r.stdout.splitlines()[-5:])
        return False, f"rc={r.returncode}: {tail[:300]}"

  def _run_ts_tests(self, rel_dir: str, work_repo: Path) -> tuple:
    cwd = work_repo / rel_dir
    if not (cwd / "package.json").exists():
      return True, f"no package.json in {rel_dir}"
      r = subprocess.run(["bun", "test", "--bail", "1"], capture_output=True, text=True, cwd=cwd)
      if r.returncode == 0:
        return True, "all green"
        tail = "\n".join(r.stdout.splitlines()[-5:])
        return False, f"rc={r.returncode}: {tail[:300]}"

  def _run(self, cmd: list, cwd: Path = None, check: bool = True) -> subprocess.CompletedProcess:
    return subprocess.run(cmd, capture_output=True, text=True, cwd=cwd, check=check)

  def _changed_files(self, work_repo: Path) -> List[str]:
    r = subprocess.run(
    ["git", "diff", "--name-only", "main...HEAD"],
    capture_output=True,
    text=True,
    cwd=work_repo,
    )
    return [p for p in r.stdout.splitlines() if p]

  def _create_stub_pr(
    self,
    analysis: Dict,
    error_types: Dict,
    work_repo: Path = None,
    branch_name: str = None,
    diff_text: str = None,
    ) -> Optional[str]:
      """Fallback: PR with the analysis Markdown stub (no code change)."""
      error_summary = list(error_types.keys())[0] if error_types else "Unknown"
      if work_repo is None or branch_name is None:
        work_repo = Path(".")
        timestamp = datetime.now().strftime("%Y%m%d%H%M%S")
        branch_name = f"ops-phoenix/fix-{timestamp}"
        self._run(["git", "checkout", "main"], cwd=work_repo)
        self._run(["git", "pull", "origin", "main"], cwd=work_repo)
        self._run(["git", "checkout", "-b", branch_name], cwd=work_repo)

        diff_section = (
        f"\n## Attempted diff (did not apply cleanly)\n```diff\n{diff_text[:2000]}\n```\n"
        if diff_text
        else ""
        )
        fix_content = (
        f"# Ops Phoenix Fix (analysis only)\n\n"
        f"Generated: {datetime.now().isoformat()}\n"
        f"Issue: #{self.issue_num or 'N/A'}\n"
        f"Root Cause: {analysis.get('root_cause', 'N/A') if analysis else 'N/A'}\n"
        f"Severity: {analysis.get('severity', 'medium') if analysis else 'medium'}\n"
        f"{diff_section}\n"
        f"## Notes\n\n"
        "This PR contains analysis only. The code-aware fix generator did not produce "
        "a diff that applied cleanly, so the agent fell back to opening a doc-only PR "
        "for human review.\n"
        )
        fix_file = work_repo / "ops-phoenix-fix.md"
        fix_file.write_text(fix_content)
        self._run(["git", "add", "ops-phoenix-fix.md"], cwd=work_repo)
        commit = subprocess.run(
        ["git", "commit", "-m", f"docs(ops-phoenix): {error_summary} (analysis only)"],
        capture_output=True,
        text=True,
        cwd=work_repo,
        )
        if commit.returncode != 0:
          self.log(f"Stub commit failed: {commit.stderr}", "ERROR")
          return None
          self._run(["git", "push", "-u", "origin", branch_name], cwd=work_repo, check=False)
          pr_title = f"docs(ops-phoenix): {error_summary} (analysis only)"
          pr_body = (
          f"## Ops Phoenix - analysis only (no code change)\n\n"
          f"**Issue:** #{self.issue_num or 'N/A'}\n"
          f"**Mode:** {ENV_MODE}\n\n"
          f"### Root Cause\n{analysis.get('root_cause', 'N/A') if analysis else 'N/A'}\n\n"
          "---\n*Generated by Ops Phoenix*\n"
          )
          result = subprocess.run(
          ["gh", "pr", "create", "--title", pr_title, "--body", pr_body, "--label", "sre-alert"],
          capture_output=True,
          text=True,
          )
          if result.returncode == 0:
            pr_url = result.stdout.strip()
            m = re.search(r"/pull/(\d+)", pr_url)
            self.pr_num = m.group(1) if m else None
            self.log(f"Stub PR created: {pr_url}")
            return self.pr_num
            self.log(f"Stub PR creation failed: {result.stderr}", "ERROR")
            return None

  def deploy(self) -> Optional[str]:
    """Step 6: Merge and deploy"""
    self.log(f"\n=== STEP 6: DEPLOY ===")

    if ENV_MODE == "dev":
      self.log("DEV MODE: Skipping auto-merge")
      self.log(f"To merge manually: gh pr merge {self.pr_num}")
      return None

      self.log("Merging PR...")
      result = subprocess.run(
      ["gh", "pr", "merge", self.pr_num, "--squash", "--delete-branch"],
      capture_output=True,
      text=True,
      )

      if result.returncode != 0:
        self.log(f"Merge failed: {result.stderr}", "ERROR")
        return None

        self.log("PR merged")

        workflow_name = self.config.get("cicd", {}).get("workflow_name", "Deploy to Production")
        self.log(f"Triggering workflow: {workflow_name}")

        result = subprocess.run(["gh", "workflow", "run", workflow_name], capture_output=True, text=True)

        if result.returncode != 0:
          self.log(f"Workflow trigger failed: {result.stderr}", "ERROR")
          return None

          run_url = result.stdout.strip()
          self.workflow_run_id = run_url
          self.log(f"Workflow triggered: {run_url}")

          return run_url

  def monitor_workflow(self, max_wait: int = 300) -> str:
    """Step 7: Monitor workflow"""
    self.log(f"\n=== STEP 7: MONITOR WORKFLOW ===")

    if not self.workflow_run_id:
      self.log("No workflow to monitor")
      return "unknown"

      start_time = time.time()
      check_interval = 15

      while time.time() - start_time < max_wait:
        result = subprocess.run(
        ["gh", "run", "view", self.workflow_run_id, "--json", "status,conclusion,name"],
        capture_output=True,
        text=True,
        )

        if result.returncode == 0:
          data = json.loads(result.stdout)
          status = data.get("status", "")
          conclusion = data.get("conclusion", "")

          elapsed = int(time.time() - start_time)
          self.log(f"Status: {status}/{conclusion or 'running'} (elapsed: {elapsed}s)")

          if status == "completed":
            if conclusion == "success":
              self.log("Workflow succeeded!", "INFO")
              return "success"
            else:
              self.log(f"Workflow failed: {conclusion}", "ERROR")
              return f"failed: {conclusion}"

              time.sleep(check_interval)

              self.log("Workflow monitoring timed out", "WARN")
              return "timeout"

  def run_full_cycle(self):
    """Run complete Ops Phoenix cycle"""
    self.log("=== OPS PHOENIX FULL CYCLE ===")
    self.log(f"Mode: {ENV_MODE.upper()}")
    self.log(f"Config: {self.config.get('github', {}).get('repo', 'N/A')}")

    self.history.start_run(mode=ENV_MODE, time_window=TIME_WINDOW)

    if not self.test_connections():
      self.log("Connection tests failed. Run --setup to configure.", "ERROR")
      self.history.complete_run(status="failed", error_message="Connection failed")
      return {"status": "failed", "reason": "connection_failed"}

      errors = self.detect_errors(TIME_WINDOW)

      if not errors:
        self.log("\nNo errors found - all systems healthy!")
        self.history.add_errors([], {})
        self.history.complete_run(status="healthy")
        return {"status": "healthy", "errors": 0}

        error_types = self.count_error_types(errors)
        self.history.add_errors(errors, error_types)

        analysis, error_types = self.analyze_errors(errors)

        issue_num = self.create_issue(errors, analysis, error_types)
        if issue_num:
          self.history.add_issue(
          int(issue_num),
          f"[ops-alert] {list(error_types.keys())[0] if error_types else 'errors'}",
          f"https://github.com/{self.config.get('github', {}).get('repo')}/issues/{issue_num}",
          )

          pr_num = self.create_pr(analysis, error_types)
          if pr_num:
            self.history.add_pr(
            int(pr_num),
            f"fix: {list(error_types.keys())[0] if error_types else 'errors'}",
            f"https://github.com/{self.config.get('github', {}).get('repo')}/pull/{pr_num}",
            )

            if not pr_num:
              self.history.complete_run(status="failed", error_message="PR creation failed")
              return {"status": "failed", "step": "PR creation"}

              if FULL_CYCLE:
                run_id = self.deploy()
                if run_id:
                  self.history.add_deploy(
                  run_id, "triggered", self.config.get("cicd", {}).get("workflow_name", "unknown")
                  )

                  result = self.monitor_workflow()

                  if run_id:
                    if result == "success":
                      self.history.add_deploy(
                      run_id, "success", self.config.get("cicd", {}).get("workflow_name", "unknown")
                      )
                    else:
                      self.history.add_deploy(
                      run_id,
                      f"failed: {result}",
                      self.config.get("cicd", {}).get("workflow_name", "unknown"),
                      )

                      self.history.complete_run(status=result)
                      return {
                      "status": result,
                      "issue": self.issue_num,
                      "pr": self.pr_num,
                      "workflow": self.workflow_run_id,
                      }

                      self.history.complete_run(status="complete")
                      return {"status": "complete", "issue": self.issue_num, "pr": self.pr_num}

  def run_auto_cycle(self):
    """Autonomous-runner entry point (called by --action auto-cycle).
    
    Detects errors, then dispatches based on what the threshold manager
    decides (full cycle / pr_only / issue_only / ignore). The full
    ThresholdManager integration lives in a follow-up PR; this initial
    version falls through to run_full_cycle() for any non-empty error
    set, matching the existing --full-cycle flow.
    """
    self.log("=== OPS PHOENIX AUTO-CYCLE ===")
    self.history.start_run(mode=ENV_MODE, time_window=TIME_WINDOW)
    
    if not self.test_connections():
      self.log("Connection tests failed.", "ERROR")
      self.history.complete_run(status="failed", error_message="connection_failed")
      return {"status": "failed", "reason": "connection_failed"}
    
    errors = self.detect_errors(TIME_WINDOW)
    if not errors:
      self.log("No errors found - all systems healthy.")
      self.history.complete_run(status="healthy")
      return {"status": "healthy"}
    
    # TODO(spec 055.6): call ThresholdManager.should_create_pr() and
    # dispatch on action_level. For now default to full cycle.
    self.log("Threshold manager not yet wired; running full cycle.")
    self.history.complete_run(status="complete", error_message="threshold_mgr_pending")
    return self.run_full_cycle()
    
  def _load_env(self):
    if os.environ.get("ANTHROPIC_API_KEY"):
      self.api_key = os.environ.get("ANTHROPIC_API_KEY")
      self.api_url = os.environ.get("ANTHROPIC_API_URL", "https://api.anthropic.com")


  def main():
    global TIME_WINDOW, ENV_MODE, FULL_CYCLE, DRY_RUN

    agent = OpsPhoenix()

    args = sys.argv[1:]
    i = 0
    while i < len(args):
      arg = args[i]
      if arg == "--setup":
        agent.run_setup()
        return
      elif arg == "--history":
        agent.history.print_summary()
        return
      elif arg == "--history-export":
        path = Path.home() / ".ops-phoenix" / "history.md"
        agent.history.export_markdown(path)
        print(f"Exported to: {path}")
        return
      elif arg == "--dry-run":
        DRY_RUN = True
      elif arg == "--full-cycle":
        FULL_CYCLE = True
      elif arg == "--env" and i + 1 < len(args):
        ENV_MODE = args[i + 1]
        i += 1
      elif arg == "--window" and i + 1 < len(args):
        TIME_WINDOW = args[i + 1]
        i += 1
      elif arg == "--config" and i + 1 < len(args):
        agent.load_config(Path(args[i + 1]))
        print(json.dumps(agent.config, indent=2))
        return
      elif arg == "--action" and i + 1 < len(args):
        action = args[i + 1]
        i += 1
        if action == "auto-cycle":
          agent.run_auto_cycle()
          return
      elif arg == "--help":
        print(__doc__)
        return
        i += 1

        if agent.needs_setup():
          print("\n" + "=" * 60)
          print(" OPS PHOENIX - FIRST TIME SETUP")
          print("=" * 60)

          if os.environ.get("GRAFANA_URL") or os.environ.get("LOKI_URL") or os.environ.get("GITHUB_REPO"):
            print("\nDetected environment variables. Setting up automatically...\n")
            agent.run_setup_non_interactive()
          else:
            print("\nConfiguration is incomplete. Running setup wizard...\n")
            agent.run_setup()
            print("\nSetup complete! Run again to start monitoring.")
            return

            if DRY_RUN:
              agent.test_connections()
              agent.detect_errors(TIME_WINDOW)
            else:
              result = agent.run_full_cycle()
              print(f"\n=== RESULT ===")
              print(f"Status: {result.get('status', 'unknown')}")
              if result.get("issue"):
                print(f"Issue: #{result['issue']}")
                if result.get("pr"):
                  print(f"PR: #{result['pr']}")


                  if __name__ == "__main__":
                    main()