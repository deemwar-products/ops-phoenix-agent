"""GitHub Actions CI integration."""

import os
import time
from typing import Optional


class GitHubActionsCI:
    """Trigger and monitor GitHub Actions workflows."""

    def __init__(self, vcs, workflow_name: str = "Deploy to Production"):
        self.vcs = vcs
        self.workflow_name = workflow_name

    def trigger(self, ref: str = "main") -> Optional[str]:
        """Trigger the deployment workflow."""
        import subprocess
        repo = self.vcs.repo
        cmd = [
            "gh", "workflow", "run", self.workflow_name,
            "--repo", repo,
            "--ref", ref,
        ]
        env = dict(os.environ)
        if self.vcs.token:
            env["GITHUB_TOKEN"] = self.vcs.token
        r = subprocess.run(cmd, capture_output=True, text=True, env=env)
        if r.returncode != 0:
            print(f"Failed to trigger workflow: {r.stderr}")
            return None
        return r.stdout.strip()

    def wait_for_completion(self, run_id: str, timeout: int = 300) -> tuple:
        """Poll workflow run until complete."""
        import subprocess
        start = time.time()
        while time.time() - start < timeout:
            status, conclusion = self._get_status(run_id)
            if status == "completed":
                return status, conclusion
            time.sleep(10)
        return "timeout", None

    def _get_status(self, run_id: str) -> tuple:
        """Get current workflow run status."""
        import subprocess, json
        cmd = [
            "gh", "run", "view", run_id,
            "--json", "status,conclusion",
        ]
        env = dict(os.environ)
        if self.vcs.token:
            env["GITHUB_TOKEN"] = self.vcs.token
        r = subprocess.run(cmd, capture_output=True, text=True, env=env)
        if r.returncode != 0:
            return "unknown", None
        try:
            data = json.loads(r.stdout)
            return data.get("status", "unknown"), data.get("conclusion")
        except Exception:
            return "unknown", None

    def extract_run_id(self, url: str) -> Optional[str]:
        """Extract run ID from workflow run URL."""
        import re
        m = re.search(r"/actions/runs/(\d+)", url)
        return m.group(1) if m else None
