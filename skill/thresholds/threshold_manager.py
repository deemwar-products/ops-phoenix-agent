#!/usr/bin/env python3
"""
Threshold Manager for SRE Phoenix Agent

Implements intelligent thresholding to decide when to create PRs vs. just log alerts.

Usage:
 from threshold_manager import ThresholdManager
 tm = ThresholdManager()
 should_act, action_level, reason = tm.should_create_pr(error_types, analysis, recent_issues)
"""

import json
import subprocess
import os
from datetime import datetime, timedelta
from typing import Dict, List, Optional, Tuple, Any
from pathlib import Path


class ThresholdConfig:
    """Configuration for threshold rules"""

    def __init__(self, config_path: Optional[Path] = None):
        self._load_config(config_path)

        # Error count thresholds
        self.min_error_count = 3  # Minimum errors to trigger action
        self.critical_count = 1  # Critical errors trigger at 1

        # Severity thresholds
        self.auto_fix_severity = ["critical"]  # Auto-deploy
        self.pr_only_severity = ["high", "medium"]  # Create PR for review
        self.issue_only_severity = ["low"]  # Create issue only
        self.ignore_severity = []  # Log only

        # Deduplication
        self.dedup_window_hours = 48  # Don't duplicate within 48h

        # Rate limiting
        self.max_runs_per_hour = 4  # Max runs per hour
        self.rate_cooldown_minutes = 15  # Cooldown between runs

        # Error type filters
        self.ignore_error_patterns = [
            "heartbeat",
            "health_check",
            "ping",
            "liveness",
            "readiness",
        ]
        self.always_alert_patterns = [
            "panic",
            "fatal",
            "crash",
            "oom",
            "out_of_memory",
        ]

    def _load_config(self, config_path: Optional[Path]):
        """Load thresholds from config file if exists"""
        if config_path is not None and config_path.exists():
            try:
                with open(config_path) as f:
                    data = json.load(f)
                thresholds = data.get("thresholds", {})
                for key, value in thresholds.items():
                    if hasattr(self, key):
                        setattr(self, key, value)
            except Exception:
                pass


class ThresholdManager:
    """
    Manages thresholds for SRE Phoenix PR creation decisions.

    Decision hierarchy:
    1. Check if error matches "always alert" patterns (panic, fatal, etc.) - always act
    2. Check deduplication - don't create duplicate issues
    3. Check error count threshold - minimum errors required
    4. Check severity threshold - determines action level
    5. Check rate limiting - avoid alert fatigue
    """

    def __init__(self, config: Optional[ThresholdConfig] = None):
        self.config = config or ThresholdConfig()

        # Track recent runs for rate limiting
        self.state_file = Path.home() / ".ops-phoenix" / "threshold_state.json"
        self._load_state()

    def _load_state(self):
        """Load rate limiting state"""
        if self.state_file.exists():
            try:
                with open(self.state_file) as f:
                    self.state = json.load(f)
            except Exception:
                self.state = {"recent_runs": [], "last_action": None}
        else:
            self.state = {"recent_runs": [], "last_action": None}

    def _save_state(self):
        """Save rate limiting state"""
        self.state_file.parent.mkdir(parents=True, exist_ok=True)
        with open(self.state_file, "w") as f:
            json.dump(self.state, f, indent=2)

    def _clean_old_runs(self):
        """Remove runs older than 1 hour"""
        cutoff = datetime.now() - timedelta(hours=1)
        self.state["recent_runs"] = [
            ts for ts in self.state["recent_runs"]
            if datetime.fromisoformat(ts) > cutoff
        ]

    def matches_pattern(self, error_type: str, patterns: List[str]) -> bool:
        """Check if error matches any of the patterns (case-insensitive)"""
        error_lower = error_type.lower()
        return any(p.lower() in error_lower for p in patterns)

    def check_always_alert(self, error_types: Dict[str, int]) -> Tuple[bool, str]:
        """Check if any error matches always-alert patterns"""
        for error_type in error_types.keys():
            for pattern in self.config.always_alert_patterns:
                if pattern.lower() in error_type.lower():
                    return True, f"Always-alert pattern matched: {pattern}"
        return False, ""

    def check_deduplication(
        self, error_types: Dict[str, int], recent_issues: List[Dict]
    ) -> Tuple[bool, Optional[int]]:
        """
        Check if similar issue exists in deduplication window.
        Returns (is_duplicate, existing_issue_number)
        """
        if not recent_issues:
            return False, None

        primary_error = list(error_types.keys())[0] if error_types else ""

        for issue in recent_issues:
            try:
                created_str = issue.get("created_at", issue.get("created", ""))
                if not created_str:
                    continue

                # Parse date (handle various formats)
                created = self._parse_date(created_str)
                if not created:
                    continue

                # Check if within deduplication window
                if (
                    datetime.now() - created
                    > timedelta(hours=self.config.dedup_window_hours)
                ):
                    continue

                # Check if similar title
                title = issue.get("title", "").lower()
                if primary_error.lower() in title or any(
                    p.lower() in title for p in primary_error.lower().split(":")
                ):
                    return True, issue.get("number")

            except Exception:
                continue

        return False, None

    def _parse_date(self, date_str: str) -> Optional[datetime]:
        """Parse various date formats"""
        formats = [
            "%Y-%m-%dT%H:%M:%SZ",
            "%Y-%m-%dT%H:%M:%S",
            "%Y-%m-%d %H:%M:%S",
            "%Y-%m-%d",
        ]
        for fmt in formats:
            try:
                return datetime.strptime(date_str.replace("Z", ""), fmt)
            except ValueError:
                continue
        return None

    def check_error_count(self, error_types: Dict[str, int]) -> Tuple[bool, int]:
        """Check if error count meets minimum threshold"""
        if not error_types:
            return False, 0

        max_count = max(error_types.values())

        # Critical errors trigger at 1
        for error_type in error_types.keys():
            if self.matches_pattern(
                error_type, self.config.always_alert_patterns
            ):
                return True, max_count

        if max_count >= self.config.min_error_count:
            return True, max_count

        return False, max_count

    def get_severity_action(self, severity: str) -> str:
        """Determine action level based on severity"""
        severity = severity.lower()

        if severity in self.config.auto_fix_severity:
            return "full_cycle"  # Auto-deploy
        elif severity in self.config.pr_only_severity:
            return "pr_only"  # Create PR for review
        elif severity in self.config.issue_only_severity:
            return "issue_only"  # Create issue only
        elif severity in self.config.ignore_severity:
            return "ignore"  # Log only

        # Default based on severity level
        severity_order = {"critical": 4, "high": 3, "medium": 2, "low": 1}
        level = severity_order.get(severity, 2)

        if level >= 3:  # high or critical
            return "pr_only"
        elif level >= 2:  # medium
            return "issue_only"
        else:
            return "ignore"

    def check_rate_limit(self) -> Tuple[bool, str]:
        """Check if we're within rate limits"""
        self._clean_old_runs()

        recent_count = len(self.state["recent_runs"])
        if recent_count >= self.config.max_runs_per_hour:
            return (
                False,
                f"Rate limit exceeded: {recent_count}/{self.config.max_runs_per_hour} runs/hour",
            )

        # Check cooldown
        if self.state.get("last_action"):
            last_action = self._parse_date(self.state["last_action"])
            if last_action:
                cooldown = timedelta(minutes=self.config.rate_cooldown_minutes)
                if datetime.now() - last_action < cooldown:
                    return (
                        False,
                        f"Cooldown active: wait {self.config.rate_cooldown_minutes} minutes between actions",
                    )

        return True, ""

    def should_create_pr(
        self,
        error_types: Dict[str, int],
        analysis: Optional[Dict[str, Any]] = None,
        recent_issues: Optional[List[Dict]] = None,
    ) -> Tuple[bool, str, str]:
        """
        Main decision function for PR creation.

        Args:
        error_types: Dict of {error_type: count}
        analysis: Optional Claude AI analysis result
        recent_issues: List of recent issues from GitHub

        Returns:
        (should_act, action_level, reason)
        action_level: 'full_cycle', 'pr_only', 'issue_only', 'ignore'
        """

        # Default values
        analysis = analysis or {}
        recent_issues = recent_issues or []

        # Step 1: Check always-alert patterns
        is_always_alert, alert_reason = self.check_always_alert(error_types)
        if is_always_alert:
            return True, "full_cycle", alert_reason

        # Step 2: Check deduplication
        is_dup, existing_issue = self.check_deduplication(error_types, recent_issues)
        if is_dup:
            return True, "update_existing", f"Similar issue #{existing_issue} exists"

        # Step 3: Check error count
        count_ok, max_count = self.check_error_count(error_types)
        if not count_ok:
            return (
                False,
                "ignore",
                f"Error count {max_count} below threshold {self.config.min_error_count}",
            )

        # Step 4: Check severity
        severity = analysis.get("severity", "medium")
        action_level = self.get_severity_action(severity)

        if action_level == "ignore":
            return False, "ignore", f"Severity {severity} below threshold"

        # Step 5: Check rate limit
        rate_ok, rate_reason = self.check_rate_limit()
        if not rate_ok:
            return True, "log_only", f"Rate limited: {rate_reason}"

        # All checks passed
        severity_reasons = {
            "full_cycle": f"Critical error - auto-deploy (severity: {severity})",
            "pr_only": f"Medium/high severity - create PR for review (severity: {severity})",
            "issue_only": f"Low severity - create issue only (severity: {severity})",
        }
        return (
            True,
            action_level,
            severity_reasons.get(action_level, f"Severity: {severity}"),
        )

    def record_action(self, action_level: str, error_signature: str):
        """Record an action for rate limiting"""
        now = datetime.now()
        self.state["recent_runs"].append(now.isoformat())
        self.state["last_action"] = now.isoformat()
        self.state["last_error_signature"] = error_signature
        self._save_state()

    def get_action_summary(self, should_act: bool, action_level: str, reason: str) -> str:
        """Get human-readable summary of decision"""
        if not should_act:
            return f"IGNORE: {reason}"

        summaries = {
            "full_cycle": f"ACT: {reason}",
            "pr_only": f"ACT: {reason}",
            "issue_only": f"ACT: {reason}",
            "update_existing": f"ACT: {reason}",
            "log_only": f"LOG ONLY: {reason}",
            "ignore": f"IGNORE: {reason}",
        }
        return summaries.get(action_level, f"UNKNOWN: {reason}")


def get_recent_github_issues(repo: str, hours: int = 48) -> List[Dict]:
    """Fetch recent GitHub issues for deduplication check"""
    try:
        result = subprocess.run(
            [
                "gh",
                "issue",
                "list",
                "--label",
                "sre-alert",
                "--state",
                "all",
                "--limit",
                "50",
                "--json",
                "number,title,createdAt,labels",
            ],
            capture_output=True,
            text=True,
            check=True,
        )

        issues = json.loads(result.stdout)

        # Filter to recent issues
        cutoff = datetime.now() - timedelta(hours=hours)
        recent = []
        for issue in issues:
            created = issue.get("createdAt", "")
            if created:
                parsed = datetime.fromisoformat(created.replace("Z", "+00:00"))
                if parsed.replace(tzinfo=None) > cutoff:
                    recent.append(issue)

        return recent
    except Exception:
        return []


def demo():
    """Demo the threshold manager"""
    tm = ThresholdManager()

    # Test cases
    test_cases = [
        {
            "name": "Single non-critical error",
            "error_types": {"api: timeout": 1},
            "analysis": {"severity": "low"},
        },
        {
            "name": "Multiple medium errors",
            "error_types": {"api: validation_failed": 5},
            "analysis": {"severity": "medium"},
        },
        {
            "name": "Critical panic error",
            "error_types": {"api: panic: connection lost": 1},
            "analysis": {"severity": "critical"},
        },
        {
            "name": "High severity with count",
            "error_types": {"database: connection_refused": 3},
            "analysis": {"severity": "high"},
        },
    ]

    print("=" * 70)
    print("SRE Phoenix Threshold Manager Demo")
    print("=" * 70)
    print()

    for tc in test_cases:
        print(f"Test: {tc['name']}")
        print(f" Error types: {tc['error_types']}")
        print(f" Severity: {tc['analysis'].get('severity')}")

        should_act, action, reason = tm.should_create_pr(
            tc["error_types"], tc["analysis"]
        )

        print(f" Decision: {tm.get_action_summary(should_act, action, reason)}")
        print()

    print("=" * 70)
    print("Threshold Configuration")
    print("=" * 70)
    print(f" Min error count: {tm.config.min_error_count}")
    print(f" Auto-fix severity: {tm.config.auto_fix_severity}")
    print(f" PR-only severity: {tm.config.pr_only_severity}")
    print(f" Issue-only severity: {tm.config.issue_only_severity}")
    print(f" Dedup window: {tm.config.dedup_window_hours}h")
    print(f" Max runs/hour: {tm.config.max_runs_per_hour}")


if __name__ == "__main__":
    demo()
