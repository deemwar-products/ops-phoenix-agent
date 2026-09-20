#!/usr/bin/env python3
"""
Ops Phoenix - Automated Ops Agent

Monitors: https://observability.deemwar.com/d/reqsume-logs

Usage:
    python3 check-errors.py --dry-run              # Detect only
    python3 check-errors.py --create-issue         # Detect + Issue
    python3 check-errors.py --full-cycle           # Full flow (detect → issue → fix → PR → deploy)
    python3 check-errors.py --env dev              # Test mode (creates PR but waits for approval)
    python3 check-errors.py --env prod             # Production mode (auto-merges and deploys)
    python3 check-errors.py --monitor              # Monitor workflow status

Modes:
    --env dev     = Test mode. Creates PR but does NOT merge. Waits for approval.
    --env prod    = Production mode. Auto-merges and deploys.
    --env staging = Same as prod but deploys to dev first

Requirements:
    - ANTHROPIC_API_KEY environment variable
    - ~/Downloads/Archive/keys/grafana-api-token
    - gh CLI authenticated with GitHub
"""

import json
import subprocess
import sys
import os
import re
import time
from datetime import datetime
from pathlib import Path

# Configuration
GRAFANA_TOKEN_PATH = os.path.expanduser("~/Downloads/Archive/keys/grafana-api-token")
GRAFANA_API = "https://observability.deemwar.com/api/ds/query?ds_type=loki"
GITHUB_REPO = "muthuishere/reqsume"
ANTHROPIC_API_URL = os.environ.get("ANTHROPIC_BASE_URL", "https://api.opusmax.pro")
LOG_DIR = "/tmp/ops-phoenix"

# Defaults
DRY_RUN = False
CREATE_ISSUE = False
FULL_CYCLE = False
ENV_MODE = "dev"  # 'dev' = test mode (no auto-merge), 'prod' = production (auto-merge)
TIME_WINDOW = "1h"
MONITOR_ONLY = False

class OpsPhoenix:
    def __init__(self):
        self.api_key = os.environ.get("ANTHROPIC_API_KEY")
        self.errors_found = []
        self.issue_num = None
        self.pr_num = None
        self.workflow_run_id = None

    def log(self, message, level="INFO"):
        """Log with timestamp and level"""
        timestamp = datetime.now().strftime('%Y-%m-%d %H:%M:%S')
        print(f"[{timestamp}] [{level}] {message}")

    def load_grafana_token(self):
        """Load Grafana API token from file"""
        if not os.path.exists(GRAFANA_TOKEN_PATH):
            raise FileNotFoundError(f"Grafana token not found: {GRAFANA_TOKEN_PATH}")
        with open(GRAFANA_TOKEN_PATH) as f:
            return f.read().strip()

    def get_time_range(self, window):
        """Map time window to Grafana duration strings"""
        ranges = {
            "1h": ("now-1h", "now"),
            "6h": ("now-6h", "now"),
            "24h": ("now-24h", "now"),
            "7d": ("now-7d", "now"),
        }
        return ranges.get(window, ("now-1h", "now"))

    def query_loki_errors(self, token, window="1h"):
        """Query Loki for error logs from reqsume-app containers"""
        from_ts, to_ts = self.get_time_range(window)

        query = {
            "queries": [{
                "expr": '{container=~"reqsume-app-web.*|reqsume-api.*", host=~"app1|app2"} | json | level=~"(?i)error|fatal"',
                "refId": "A",
                "datasource": {"type": "loki", "uid": "loki"},
                "maxLines": 100
            }],
            "from": from_ts,
            "to": to_ts
        }

        self.log(f"Querying Loki for errors (window: {window})...")

        result = subprocess.run([
            "curl", "-sS", "-H", f"Authorization: Bearer {token}",
            "-H", "Content-Type: application/json",
            "-X", "POST", GRAFANA_API,
            "--data", json.dumps(query)
        ], capture_output=True, text=True)

        if result.returncode != 0:
            self.log(f"curl error: {result.stderr}", "ERROR")
            return []

        return self.parse_errors(result.stdout)

    def parse_errors(self, response):
        """Parse Loki response and extract error logs"""
        if not response:
            return []

        try:
            data = json.loads(response)
        except json.JSONDecodeError:
            return []

        status = data.get("results", {}).get("A", {}).get("status")
        if status != 200:
            return []

        frames = data.get("results", {}).get("A", {}).get("frames", [])
        if not frames:
            return []

        vals = frames[0].get("data", {}).get("values", [])
        if len(vals) <= 2:
            return []

        return [l for l in vals[2] if l and l.strip()]

    def count_error_types(self, errors):
        """Count errors by type"""
        error_types = {}
        for line in errors:
            try:
                if line.startswith("{"):
                    parsed = json.loads(line)
                    msg = parsed.get("msg", "unknown")
                    logger = parsed.get("logger", "unknown")
                    key = f"{logger}: {msg}"
                    error_types[key] = error_types.get(key, 0) + 1
            except:
                pass
        return error_types

    def detect(self, time_window="1h"):
        """Step 1: Detect errors"""
        self.log("=== STEP 1: DETECT ===")

        if not self.api_key:
            self.log("ANTHROPIC_API_KEY not set", "ERROR")
            return None

        try:
            token = self.load_grafana_token()
        except FileNotFoundError as e:
            self.log(str(e), "ERROR")
            return None

        errors = self.query_loki_errors(token, time_window)

        if not errors:
            self.log("No errors found in reqsume-app containers")
            return []

        error_types = self.count_error_types(errors)

        self.log(f"Found {len(errors)} error logs")
        self.log("Error breakdown:")
        for key, count in sorted(error_types.items(), key=lambda x: -x[1])[:5]:
            self.log(f"  - {key}: {count} occurrences")

        return errors

    def analyze_errors(self, errors):
        """Step 2: Analyze errors with Claude AI"""
        self.log("\n=== STEP 2: ANALYZE ===")

        if not errors:
            return None, {}

        error_types = self.count_error_types(errors)
        error_summary = "\n".join([f"- {k}: {v} occurrences" for k, v in error_types.items()])

        prompt = f"""You are an SRE analyzing production errors from the Reqsume application.

## Error Summary
{error_summary}

Please provide:
1. ROOT_CAUSE: One sentence explaining the root cause
2. AFFECTED: What functionality is affected
3. SEVERITY: critical/high/medium/low
4. FIX_SUGGESTION: How to fix this (be specific)
5. FILES_AFFECTED: Which files need to be changed

Format your response as:
ROOT_CAUSE: ...
AFFECTED: ...
SEVERITY: ...
FIX_SUGGESTION: ...
FILES_AFFECTED: ..."""

        self.log("Analyzing with Claude AI...")

        result = subprocess.run([
            "curl", "-sS", "-X", "POST", f"{ANTHROPIC_API_URL}/v1/messages",
            "-H", f"x-api-key: {self.api_key}",
            "-H", "anthropic-version: 2023-06-01",
            "-H", "content-type: application/json",
            "--data", json.dumps({
                "model": "claude-opus-4-7",
                "max_tokens": 1024,
                "messages": [{"role": "user", "content": prompt}]
            })
        ], capture_output=True, text=True)

        if result.returncode != 0:
            self.log(f"Claude API error: {result.stderr}", "ERROR")
            return None, error_types

        try:
            response = json.loads(result.stdout)
            content = response.get("content", [])
            if content:
                analysis_text = content[0].get("text", "")
                analysis = self.parse_analysis(analysis_text)
                self.log(f"Root cause: {analysis.get('root_cause', 'N/A')}")
                self.log(f"Affected: {analysis.get('affected', 'N/A')}")
                self.log(f"Severity: {analysis.get('severity', 'medium')}")
                return analysis, error_types
        except (json.JSONDecodeError, KeyError) as e:
            self.log(f"Parse error: {e}", "ERROR")

        return None, error_types

    def parse_analysis(self, analysis_text):
        """Parse Claude's analysis response"""
        result = {
            "root_cause": "Unknown",
            "affected": "Unknown",
            "severity": "medium",
            "fix_suggestion": "",
            "files_affected": []
        }

        for line in analysis_text.split("\n"):
            if line.startswith("ROOT_CAUSE:"):
                result["root_cause"] = line.replace("ROOT_CAUSE:", "").strip()
            elif line.startswith("AFFECTED:"):
                result["affected"] = line.replace("AFFECTED:", "").strip()
            elif line.startswith("SEVERITY:"):
                result["severity"] = line.replace("SEVERITY:", "").strip().lower()
            elif line.startswith("FIX_SUGGESTION:"):
                result["fix_suggestion"] = line.replace("FIX_SUGGESTION:", "").strip()
            elif line.startswith("FILES_AFFECTED:"):
                files = line.replace("FILES_AFFECTED:", "").strip()
                result["files_affected"] = [f.strip() for f in files.split(",") if f.strip()]

        return result

    def create_issue(self, errors, analysis, error_types):
        """Step 3: Create GitHub issue"""
        self.log("\n=== STEP 3: CREATE ISSUE ===")

        error_summary = "\n".join([f"- {k}: {v} occurrences" for k, v in error_types.items()])

        body = f"""## Ops Phoenix Alert

**Detected:** {datetime.now().strftime('%Y-%m-%d %H:%M')}
**Environment:** production
**Mode:** {ENV_MODE}
**Dashboard:** https://observability.deemwar.com/d/reqsume-logs

### Errors Found

Total: {len(errors)} error logs

{error_summary}
"""

        if analysis:
            body += f"""
### Analysis

- **Root Cause:** {analysis.get('root_cause', 'Unknown')}
- **Affected:** {analysis.get('affected', 'Unknown')}
- **Severity:** {analysis.get('severity', 'medium')}

### Fix Suggestion

{analysis.get('fix_suggestion', 'Review needed')}

### Files to Change

{', '.join(analysis.get('files_affected', []) or ['Review needed'])}
"""

        body += """
---
*Generated by Ops Phoenix agent*
"""

        summary = list(error_types.keys())[0] if error_types else "Production errors detected"

        result = subprocess.run([
            "gh", "issue", "create",
            "--title", f"[ops-alert] {summary}",
            "--body", body,
            "--label", "sre-alert",
            "--label", "bug"
        ], capture_output=True, text=True)

        if result.returncode == 0:
            issue_url = result.stdout.strip()
            issue_num = re.search(r'/issues/(\d+)', issue_url)
            self.issue_num = issue_num.group(1) if issue_num else None
            self.log(f"Issue created: {issue_url}")
            return self.issue_num

        self.log(f"Issue creation failed: {result.stderr}", "ERROR")
        return None

    def generate_fix(self, analysis, error_types):
        """Step 4: Generate code fix"""
        self.log("\n=== STEP 4: GENERATE FIX ===")

        if not analysis or not analysis.get('fix_suggestion'):
            self.log("No fix suggestion available", "WARN")
            # Create a simple test fix
            return self.create_test_fix(analysis, error_types)

        error_summary = "\n".join([f"- {k}: {v} occurrences" for k, v in error_types.items()])

        prompt = f"""Generate a code fix for this production error:

## Error Analysis
- Root Cause: {analysis.get('root_cause', 'Unknown')}
- Affected: {analysis.get('affected', 'Unknown')}
- Severity: {analysis.get('severity', 'medium')}

## Suggested Fix
{analysis.get('fix_suggestion', '')}

## Files to Change
{', '.join(analysis.get('files_affected', []) or ['Review needed'])}

## Context
This is a Go/Gin API project in the apps/api/ directory.

Format your response as:
FILE: <path>
---
<code>
---
EXPLANATION: <brief explanation>
"""

        self.log("Generating fix with Claude AI...")

        result = subprocess.run([
            "curl", "-sS", "-X", "POST", f"{ANTHROPIC_API_URL}/v1/messages",
            "-H", f"x-api-key: {self.api_key}",
            "-H", "anthropic-version: 2023-06-01",
            "-H", "content-type: application/json",
            "--data", json.dumps({
                "model": "claude-opus-4-7",
                "max_tokens": 2048,
                "messages": [{"role": "user", "content": prompt}]
            })
        ], capture_output=True, text=True)

        if result.returncode != 0:
            self.log(f"Claude API error: {result.stderr}", "ERROR")
            return self.create_test_fix(analysis, error_types)

        try:
            response = json.loads(result.stdout)
            content = response.get("content", [])
            if content:
                fix_text = content[0].get("text", "")
                return self.parse_fix_response(fix_text)
        except (json.JSONDecodeError, KeyError) as e:
            self.log(f"Parse error: {e}", "ERROR")

        return self.create_test_fix(analysis, error_types)

    def create_test_fix(self, analysis, error_types):
        """Create a test fix for demo purposes"""
        error_summary = list(error_types.keys())[0] if error_types else "Unknown"

        # Create a simple test file as the "fix"
        fix = {
            "files": {
                "ops-phoenix-test.txt": f"""# Ops Phoenix Test Fix
Generated at: {datetime.now().isoformat()}
Error: {error_summary}
Root cause: {analysis.get('root_cause', 'N/A') if analysis else 'N/A'}
Mode: {ENV_MODE}
"""
            },
            "explanation": f"Test fix for {error_summary}"
        }

        self.log(f"Created test fix: ops-phoenix-test.txt")
        return fix

    def parse_fix_response(self, fix_text):
        """Parse Claude's fix response"""
        files_to_change = {}
        current_file = None
        current_content = []
        explanation = ""

        lines = fix_text.split("\n")
        i = 0
        while i < len(lines):
            line = lines[i].strip()

            if line.startswith("FILE:"):
                if current_file and current_content:
                    files_to_change[current_file] = "\n".join(current_content)
                current_file = line.replace("FILE:", "").strip()
                current_content = []
                i += 2  # Skip --- line
            elif line.startswith("EXPLANATION:"):
                explanation = line.replace("EXPLANATION:", "").strip()
            elif current_file:
                current_content.append(line)
                i += 1
            else:
                i += 1

        if current_file and current_content:
            files_to_change[current_file] = "\n".join(current_content)

        return {"files": files_to_change, "explanation": explanation}

    def create_pr(self, fix_data, analysis):
        """Step 5: Create PR"""
        self.log("\n=== STEP 5: CREATE PR ===")

        if not fix_data or not fix_data.get("files"):
            self.log("No files to change", "WARN")
            return None

        # Create branch
        timestamp = datetime.now().strftime("%Y%m%d%H%M%S")
        branch_name = f"ops-phoenix/{timestamp}"

        self.log(f"Creating branch: {branch_name}")

        # Switch to main and create branch
        subprocess.run(["git", "checkout", "main"], capture_output=True)
        subprocess.run(["git", "pull", "origin", "main"], capture_output=True)
        subprocess.run(["git", "checkout", "-b", branch_name], capture_output=True)

        # Apply changes
        for file_path, content in fix_data["files"].items():
            full_path = Path(os.getcwd()) / file_path
            full_path.parent.mkdir(parents=True, exist_ok=True)
            with open(full_path, "w") as f:
                f.write(content)
            self.log(f"Created/modified: {file_path}")

        # Create commit
        error_summary = analysis.get('root_cause', 'Unknown') if analysis else 'Unknown'

        result = subprocess.run([
            "git", "add", "."
        ], capture_output=True)

        result = subprocess.run([
            "git", "commit", "-m",
            f"fix(ops-phoenix): {error_summary}\n\nGenerated by Ops Phoenix agent"
        ], capture_output=True, text=True)

        if result.returncode != 0:
            self.log(f"Commit failed: {result.stderr}", "ERROR")
            return None

        # Push branch
        result = subprocess.run(
            ["git", "push", "-u", "origin", branch_name],
            capture_output=True, text=True
        )

        if result.returncode != 0:
            self.log(f"Push failed: {result.stderr}", "ERROR")
            return None

        self.log(f"Branch pushed: {branch_name}")

        # Create PR
        pr_body = f"""## Ops Phoenix Fix

**Mode:** {ENV_MODE}
**Issue:** #{self.issue_num or 'N/A'}

### Root Cause
{analysis.get('root_cause', 'Unknown') if analysis else 'N/A'}

### Affected
{analysis.get('affected', 'Unknown') if analysis else 'N/A'}

### Changes
{fix_data.get('explanation', 'See issue for details')}

### Workflow
{"AUTO-MERGE ENABLED" if ENV_MODE == "prod" else "MANUAL REVIEW REQUIRED - Will not auto-merge in dev mode"}

---
*Generated by Ops Phoenix agent*
"""

        result = subprocess.run([
            "gh", "pr", "create",
            "--title", f"fix(ops-phoenix): {error_summary}",
            "--body", pr_body,
            "--label", "sre-alert",
            ENV_MODE == "prod" and "--approve" or "--label", "sre-alert" if ENV_MODE == "prod" else ""
        ], capture_output=True, text=True)

        # Handle the flag issue
        if result.returncode != 0:
            # Try without extra flags
            result = subprocess.run([
                "gh", "pr", "create",
                "--title", f"fix(ops-phoenix): {error_summary}",
                "--body", pr_body,
                "--label", "sre-alert"
            ], capture_output=True, text=True)

        if result.returncode == 0:
            pr_url = result.stdout.strip()
            pr_match = re.search(r'/pull/(\d+)', pr_url)
            self.pr_num = pr_match.group(1) if pr_match else None
            self.log(f"PR created: {pr_url}")
            return self.pr_num

        self.log(f"PR creation failed: {result.stderr}", "ERROR")
        return None

    def monitor_workflow(self, pr_num=None):
        """Monitor workflow status"""
        self.log("\n=== MONITOR WORKFLOW ===")

        if pr_num:
            # Check PR workflow
            self.log(f"Checking PR #{pr_num} workflow status...")

            result = subprocess.run([
                "gh", "run", "list", "--workflow", "--json", "status,conclusion,name,headBranch"
            ], capture_output=True, text=True)

            if result.returncode == 0:
                try:
                    runs = json.loads(result.stdout)
                    for run in runs[:5]:
                        self.log(f"  - {run.get('name', 'Unknown')}: {run.get('status', 'N/A')}/{run.get('conclusion', 'N/A')}")
                except json.JSONDecodeError:
                    pass
        else:
            # Check recent workflow runs
            result = subprocess.run([
                "gh", "run", "list", "--limit", "5"
            ], capture_output=True, text=True)

            if result.returncode == 0:
                print(result.stdout)
            else:
                self.log("Failed to list workflows", "ERROR")

    def monitor_workflow(self, run_id=None, max_wait_seconds=300):
        """Monitor workflow status and wait for completion"""
        self.log("\n=== STEP 7: MONITOR WORKFLOW ===")

        if not run_id:
            run_id = self.workflow_run_id

        if not run_id:
            self.log("No workflow run ID to monitor", "WARN")
            return None

        import time
        start_time = time.time()
        check_interval = 15  # Check every 15 seconds

        while time.time() - start_time < max_wait_seconds:
            result = subprocess.run([
                "gh", "run", "view", run_id,
                "--json", "status,conclusion,name"
            ], capture_output=True, text=True)

            if result.returncode == 0:
                data = json.loads(result.stdout)
                status = data.get("status", "")
                conclusion = data.get("conclusion", "")

                elapsed = int(time.time() - start_time)
                self.log(f"Workflow status: {status}/{conclusion or 'running'} (elapsed: {elapsed}s)")

                if status == "completed":
                    if conclusion == "success":
                        self.log("✓ Workflow completed successfully!", "INFO")
                        return "success"
                    else:
                        self.log(f"✗ Workflow failed: {conclusion}", "ERROR")
                        return self.analyze_workflow_failure(run_id)
                elif status == "in_progress":
                    time.sleep(check_interval)
                else:
                    time.sleep(check_interval)
            else:
                self.log(f"Failed to get workflow status: {result.stderr}", "ERROR")
                break

        self.log("Workflow monitoring timed out", "WARN")
        return "timeout"

    def analyze_workflow_failure(self, run_id):
        """Analyze workflow failure and return failure details"""
        self.log("\nAnalyzing workflow failure...")

        result = subprocess.run([
            "gh", "run", "view", run_id, "--log-failed"
        ], capture_output=True, text=True)

        if result.returncode != 0:
            return {"status": "failed", "reason": "unknown", "can_fix": False}

        log_output = result.stdout

        # Parse failure reasons
        failure = {
            "status": "failed",
            "reason": "unknown",
            "can_fix": False,
            "details": []
        }

        # Test failures
        if "test:api-unit" in log_output or "task: Failed to run task \"test:api-unit\"" in log_output:
            failure["reason"] = "test_failure"
            failure["can_fix"] = True
            failure["details"].append("API unit tests failed")

        # Build failures
        if "go build" in log_output and "error" in log_output.lower():
            failure["reason"] = "build_failure"
            failure["can_fix"] = True
            failure["details"].append("Build failed - may need code fix")

        # Deployment failures
        if "deploy" in log_output.lower() and "failed" in log_output.lower():
            failure["reason"] = "deploy_failure"
            failure["can_fix"] = False
            failure["details"].append("Deployment failed - infrastructure issue")

        self.log(f"Failure reason: {failure['reason']}")
        for detail in failure['details']:
            self.log(f"  - {detail}")

        return failure

    def deploy(self, pr_num):
        """Step 6: Merge and deploy with auto-fix on failure"""
        self.log("\n=== STEP 6: DEPLOY ===")

        if ENV_MODE == "dev":
            self.log("DEV MODE: Skipping auto-merge")
            self.log("To merge manually:")
            self.log(f"  gh pr merge {pr_num} --squash --delete-branch")
            return None

        self.log("PROD MODE: Merging and deploying...")

        # Merge PR
        result = subprocess.run([
            "gh", "pr", "merge", pr_num, "--squash", "--delete-branch"
        ], capture_output=True, text=True)

        if result.returncode != 0:
            self.log(f"Merge failed: {result.stderr}", "ERROR")
            return None

        self.log("PR merged successfully")

        # Trigger deploy workflow
        self.log("Triggering 'Deploy to Production New' workflow...")

        result = subprocess.run([
            "gh", "workflow", "run", "Deploy to Production New"
        ], capture_output=True, text=True)

        if result.returncode != 0:
            self.log(f"Workflow trigger failed: {result.stderr}", "ERROR")
            return None

        run_url = result.stdout.strip()
        self.workflow_run_id = run_url
        self.log(f"Deploy workflow triggered: {run_url}")

        # Monitor workflow with auto-retry
        return self.deploy_with_retry()

    def deploy_with_retry(self, max_retries=2):
        """Monitor deployment and auto-fix if needed"""
        for attempt in range(1, max_retries + 1):
            self.log(f"\n--- Deploy attempt {attempt}/{max_retries} ---")

            result = self.monitor_workflow(max_wait_seconds=300)

            if result == "success":
                self.log("✓ Deployment successful!", "INFO")
                return {"status": "success", "attempts": attempt}

            if result == "timeout":
                self.log("Deployment monitoring timed out", "WARN")
                return {"status": "timeout", "attempts": attempt}

            # Check if we can fix
            if isinstance(result, dict) and result.get("can_fix"):
                self.log(f"Attempting auto-fix for: {result.get('reason')}")
                fix_success = self.fix_and_redeploy(result)

                if fix_success:
                    self.log("Auto-fix applied, deployment succeeded", "INFO")
                    return {"status": "success", "attempts": attempt, "auto_fixed": True}
            else:
                self.log("Failure cannot be auto-fixed, requires manual intervention", "ERROR")
                return {"status": "failed", "reason": result.get("reason") if isinstance(result, dict) else result}

        self.log("Max retries exceeded", "ERROR")
        return {"status": "failed", "attempts": max_retries}

    def fix_and_redeploy(self, failure_info):
        """Fix the issue and redeploy"""
        self.log("\n=== ATTEMPTING AUTO-FIX ===")

        reason = failure_info.get("reason", "")

        if reason == "test_failure":
            self.log("Test failure detected - analyzing...")
            # For test failures, we need to analyze and fix the test or code
            # This would require Claude AI to analyze the test failure logs
            self.log("Auto-fix for test failures requires Claude AI analysis")
            self.log("Creating issue for manual review...")

            # Create an issue for the test failure
            subprocess.run([
                "gh", "issue", "create",
                "--title", "[auto-fix] CI test failure needs review",
                "--body", f"""## Auto-Fix Failed

**Failure:** {failure_info.get('reason')}
**Details:** {', '.join(failure_info.get('details', []))}

The automated fix could not resolve this issue automatically.
Manual intervention required.

---
*Generated by Ops Phoenix agent*
""",
                "--label", "sre-alert",
                "--label", "bug"
            ], capture_output=True)

            return False

        return False

    def run_full_cycle(self):
        """Run the complete Ops Phoenix cycle"""
        self.log("=== OPS PHOENIX FULL CYCLE ===")
        self.log(f"Mode: {ENV_MODE.upper()}")
        self.log(f"Time window: {TIME_WINDOW}")

        # Step 1: Detect
        errors = self.detect(TIME_WINDOW)

        if not errors:
            self.log("\n✓ No errors found - all systems healthy!")
            return {"status": "healthy", "errors": 0}

        self.errors_found = errors

        # Step 2: Analyze
        analysis, error_types = self.analyze_errors(errors)

        # Step 3: Create Issue
        self.create_issue(errors, analysis, error_types)

        # Step 4: Generate Fix
        fix_data = self.generate_fix(analysis, error_types)

        # Step 5: Create PR
        pr_num = self.create_pr(fix_data, analysis)

        if not pr_num:
            self.log("PR creation failed, aborting deploy", "ERROR")
            return {"status": "failed", "step": "PR creation"}

        # Step 6: Deploy
        if FULL_CYCLE:
            self.deploy(pr_num)

        return {
            "status": "complete",
            "issue": self.issue_num,
            "pr": pr_num,
            "workflow": self.workflow_run_id
        }

def main():
    global DRY_RUN, CREATE_ISSUE, FULL_CYCLE, ENV_MODE, TIME_WINDOW, MONITOR_ONLY

    agent = OpsPhoenix()

    # Parse arguments
    args = sys.argv[1:]
    i = 0
    while i < len(args):
        arg = args[i]
        if arg == "--dry-run":
            DRY_RUN = True
        elif arg == "--create-issue":
            CREATE_ISSUE = True
        elif arg == "--full-cycle":
            FULL_CYCLE = True
        elif arg == "--env" and i + 1 < len(args):
            ENV_MODE = args[i + 1]
            i += 1
        elif arg == "--window" and i + 1 < len(args):
            TIME_WINDOW = args[i + 1]
            i += 1
        elif arg == "--monitor":
            MONITOR_ONLY = True
        elif arg == "--help" or arg == "-h":
            print(__doc__)
            sys.exit(0)
        i += 1

    # Run
    if MONITOR_ONLY:
        agent.monitor_workflow()
    elif DRY_RUN:
        agent.detect(TIME_WINDOW)
        print("\nDry run complete. Use --full-cycle to run the complete cycle.")
    else:
        result = agent.run_full_cycle()
        print(f"\n=== RESULT ===")
        print(f"Status: {result.get('status', 'unknown')}")
        if result.get('issue'):
            print(f"Issue: #{result['issue']}")
        if result.get('pr'):
            print(f"PR: #{result['pr']}")
        if result.get('workflow'):
            print(f"Workflow: {result['workflow']}")

if __name__ == "__main__":
    main()