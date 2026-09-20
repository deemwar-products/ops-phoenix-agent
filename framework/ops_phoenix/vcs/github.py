"""GitHub VCS integration."""

import os
import subprocess
from typing import Any, Dict, List, Optional


class GitHubVCS:
    """GitHub VCS adapter for issue and PR management."""

    def __init__(self, repo: str, token_env_var: str = "GITHUB_TOKEN",
                 labels: Optional[List[str]] = None):
        self.repo = repo
        self.token = os.environ.get(token_env_var, "")
        self.labels = labels or ["ops-alert"]

    def _gh(self, args: List[str]) -> Dict[str, Any]:
        """Run a gh CLI command and return parsed JSON."""
        cmd = ["gh"] + args
        env = dict(os.environ)
        if self.token:
            env["GITHUB_TOKEN"] = self.token
        r = subprocess.run(cmd, capture_output=True, text=True, env=env)
        if r.returncode != 0:
            raise RuntimeError(f"gh failed: {r.stderr[:300]}")
        return r.stdout.strip()

    def test_auth(self) -> bool:
        """Check GitHub auth."""
        r = subprocess.run(
            ["gh", "auth", "status"], capture_output=True, text=True
        )
        return r.returncode == 0

    def recent_issues(self, hours: int = 48) -> List[Dict]:
        """Fetch recent issues."""
        try:
            cmd = [
                "gh", "issue", "list", "--repo", self.repo,
                "--state", "all", "--limit", "50",
                "--json", "number,title,createdAt",
            ]
            output = self._gh(cmd)
            if not output:
                return []
            import json
            return json.loads(output)
        except Exception:
            return []

    def create_issue(self, title: str, body: str, labels: Optional[List[str]] = None) -> Optional[int]:
        """Create a GitHub issue."""
        cmd = [
            "gh", "issue", "create",
            "--repo", self.repo,
            "--title", title,
            "--body", body,
            "--label", ",".join(labels or self.labels),
        ]
        try:
            output = self._gh(cmd)
            # gh returns the issue URL, extract number
            url = output if output.startswith("http") else ""
            import re
            m = re.search(r"/issues/(\d+)", url)
            return int(m.group(1)) if m else None
        except Exception as e:
            print(f"Failed to create issue: {e}")
            return None

    def create_pr(self, title: str, body: str, base_branch: str = "main",
                  head_branch: str = "") -> Optional[int]:
        """Create a PR."""
        cmd = [
            "gh", "pr", "create",
            "--repo", self.repo,
            "--title", title,
            "--body", body,
            "--base", base_branch,
        ]
        if head_branch:
            cmd += ["--head", head_branch]
        try:
            output = self._gh(cmd)
            import re
            m = re.search(r"/pull/(\d+)", output)
            return int(m.group(1)) if m else None
        except Exception as e:
            print(f"Failed to create PR: {e}")
            return None

    def merge_pr(self, pr_num: int, method: str = "merge") -> bool:
        """Merge a PR."""
        cmd = ["gh", "pr", "merge", str(pr_num), "--repo", self.repo, "--" + method, "--yes"]
        try:
            self._gh(cmd)
            return True
        except Exception as e:
            print(f"Failed to merge PR #{pr_num}: {e}")
            return False

    def pr_status(self, pr_num: int) -> Dict[str, Any]:
        """Get PR status."""
        cmd = [
            "gh", "pr", "view", str(pr_num),
            "--repo", self.repo,
            "--json", "state,mergeStateStatus,merged",
        ]
        try:
            import json
            output = self._gh(cmd)
            return json.loads(output)
        except Exception:
            return {}
