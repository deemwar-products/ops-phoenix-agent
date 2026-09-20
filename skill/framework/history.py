#!/usr/bin/env python3
"""
Ops Phoenix - History Tracker

Tracks all runs, errors detected, issues created, PRs merged, and fixes applied.
Stores in ~/.ops-phoenix/history.json
"""

import json
import os
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Optional
from dataclasses import dataclass, asdict, field

# Config file location
OPS_PHOENIX_DIR = Path.home() / ".ops-phoenix"
HISTORY_FILE = OPS_PHOENIX_DIR / "history.json"
MEMORY_FILE = OPS_PHOENIX_DIR / "memory.md"

@dataclass
class RunRecord:
    """Record of a single Ops Phoenix run"""
    timestamp: str
    mode: str  # dev, prod
    time_window: str  # 1h, 6h, 24h, 7d
    status: str  # healthy, complete, failed
    errors_found: int
    error_types: Dict[str, int] = field(default_factory=dict)
    issues_created: List[Dict] = field(default_factory=list)
    prs_created: List[Dict] = field(default_factory=list)
    prs_merged: List[Dict] = field(default_factory=list)
    deploys_triggered: List[Dict] = field(default_factory=list)
    deploys_succeeded: List[Dict] = field(default_factory=list)
    deploys_failed: List[Dict] = field(default_factory=list)
    auto_fixes: List[Dict] = field(default_factory=list)
    manual_issues: List[Dict] = field(default_factory=list)
    duration_seconds: Optional[float] = None
    error_message: Optional[str] = None

class HistoryTracker:
    """Tracks Ops Phoenix history"""

    def __init__(self):
        self.runs: List[RunRecord] = []
        self.current_run: Optional[RunRecord] = None
        self.load()

    def load(self):
        """Load history from file"""
        if HISTORY_FILE.exists():
            try:
                with open(HISTORY_FILE) as f:
                    data = json.load(f)
                    self.runs = [RunRecord(**r) for r in data.get("runs", [])]
            except (json.JSONDecodeError, TypeError):
                self.runs = []

    def save(self):
        """Save history to file"""
        OPS_PHOENIX_DIR.mkdir(parents=True, exist_ok=True)
        data = {
            "last_updated": datetime.now().isoformat(),
            "runs": [asdict(r) for r in self.runs]
        }
        with open(HISTORY_FILE, 'w') as f:
            json.dump(data, f, indent=2)
        self.export_state_markdown(MEMORY_FILE)

    def start_run(self, mode: str = "dev", time_window: str = "1h"):
        """Start a new run"""
        self.current_run = RunRecord(
            timestamp=datetime.now().isoformat(),
            mode=mode,
            time_window=time_window,
            status="running",
            errors_found=0
        )

    def add_errors(self, errors: List[str], error_types: Dict[str, int]):
        """Record detected errors"""
        if self.current_run:
            self.current_run.errors_found = len(errors)
            self.current_run.error_types = error_types

    def add_issue(self, issue_num: int, title: str, url: str):
        """Record created issue"""
        if self.current_run:
            self.current_run.issues_created.append({
                "number": issue_num,
                "title": title,
                "url": url,
                "created_at": datetime.now().isoformat()
            })

    def add_pr(self, pr_num: int, title: str, url: str, merged: bool = False):
        """Record created PR"""
        if self.current_run:
            pr_record = {
                "number": pr_num,
                "title": title,
                "url": url,
                "created_at": datetime.now().isoformat(),
                "merged": merged
            }
            self.current_run.prs_created.append(pr_record)
            if merged:
                self.current_run.prs_merged.append(pr_record)

    def add_deploy(self, run_id: str, status: str, workflow: str):
        """Record deployment"""
        if self.current_run:
            deploy_record = {
                "run_id": run_id,
                "workflow": workflow,
                "status": status,
                "triggered_at": datetime.now().isoformat()
            }
            self.current_run.deploys_triggered.append(deploy_record)
            if status == "success":
                self.current_run.deploys_succeeded.append(deploy_record)
            else:
                self.current_run.deploys_failed.append(deploy_record)

    def add_auto_fix(self, error_type: str, fix_applied: str, success: bool):
        """Record auto-fix attempt"""
        if self.current_run:
            self.current_run.auto_fixes.append({
                "error_type": error_type,
                "fix_applied": fix_applied,
                "success": success,
                "attempted_at": datetime.now().isoformat()
            })

    def add_manual_issue(self, issue_num: int, title: str, reason: str):
        """Record manual issue created (when auto-fix fails)"""
        if self.current_run:
            self.current_run.manual_issues.append({
                "number": issue_num,
                "title": title,
                "reason": reason,
                "created_at": datetime.now().isoformat()
            })

    def complete_run(self, status: str = "complete", error_message: Optional[str] = None):
        """Complete the current run"""
        if self.current_run:
            self.current_run.status = status
            self.current_run.error_message = error_message

            # Calculate duration if we tracked start time
            start = datetime.fromisoformat(self.current_run.timestamp)
            self.current_run.duration_seconds = (datetime.now() - start).total_seconds()

            self.runs.append(self.current_run)
            self.current_run = None
            self.save()

    def get_stats(self) -> Dict:
        """Get statistics"""
        return {
            "total_runs": len(self.runs),
            "healthy_runs": len([r for r in self.runs if r.status == "healthy"]),
            "complete_runs": len([r for r in self.runs if r.status == "complete"]),
            "failed_runs": len([r for r in self.runs if r.status == "failed"]),
            "total_issues": sum(len(r.issues_created) for r in self.runs),
            "total_prs": sum(len(r.prs_created) for r in self.runs),
            "total_prs_merged": sum(len(r.prs_merged) for r in self.runs),
            "total_deploys": sum(len(r.deploys_triggered) for r in self.runs),
            "deploys_succeeded": sum(len(r.deploys_succeeded) for r in self.runs),
            "deploys_failed": sum(len(r.deploys_failed) for r in self.runs),
            "total_auto_fixes": len([f for r in self.runs for f in r.auto_fixes]),
            "auto_fixes_succeeded": len([f for r in self.runs for f in r.auto_fixes if f.get("success")]),
            "total_manual_issues": len([i for r in self.runs for i in r.manual_issues]),
        }

    def get_recent_runs(self, limit: int = 10) -> List[RunRecord]:
        """Get most recent runs"""
        return sorted(self.runs, key=lambda r: r.timestamp, reverse=True)[:limit]

    def get_fixes_summary(self) -> List[Dict]:
        """Get summary of all fixes applied"""
        fixes = []
        for run in self.runs:
            for pr in run.prs_merged:
                fixes.append({
                    "timestamp": run.timestamp,
                    "pr_number": pr.get("number"),
                    "pr_title": pr.get("title"),
                    "pr_url": pr.get("url"),
                    "errors_fixed": list(run.error_types.keys()) if run.error_types else [],
                    "run_mode": run.mode
                })
        return fixes

    def print_summary(self):
        """Print a summary of all runs"""
        stats = self.get_stats()

        print("\n" + "="*60)
        print("  OPS PHOENIX HISTORY SUMMARY")
        print("="*60)

        print(f"\n📊 Overall Statistics:")
        print(f"   Total runs: {stats['total_runs']}")
        print(f"   ✅ Healthy: {stats['healthy_runs']}")
        print(f"   🔧 Complete: {stats['complete_runs']}")
        print(f"   ❌ Failed: {stats['failed_runs']}")

        print(f"\n📝 Issues & PRs:")
        print(f"   Issues created: {stats['total_issues']}")
        print(f"   PRs created: {stats['total_prs']}")
        print(f"   PRs merged: {stats['total_prs_merged']}")

        print(f"\n🚀 Deployments:")
        print(f"   Deploys triggered: {stats['total_deploys']}")
        print(f"   ✅ Succeeded: {stats['deploys_succeeded']}")
        print(f"   ❌ Failed: {stats['deploys_failed']}")

        print(f"\n🔧 Auto-Fixes:")
        print(f"   Total attempts: {stats['total_auto_fixes']}")
        print(f"   ✅ Succeeded: {stats['auto_fixes_succeeded']}")

        print(f"\n📋 Manual Issues:")
        print(f"   Total created: {stats['total_manual_issues']}")

        print("\n" + "-"*60)
        print("  RECENT FIXES")
        print("-"*60)

        fixes = self.get_fixes_summary()[:5]
        if fixes:
            for fix in fixes:
                date = fix['timestamp'][:10]
                print(f"\n[{date}] PR #{fix['pr_number']}: {fix['pr_title']}")
                print(f"   Mode: {fix['run_mode']}")
                if fix['errors_fixed']:
                    for err in fix['errors_fixed'][:3]:
                        print(f"   Fixed: {err}")
        else:
            print("\n   No fixes recorded yet.")

        print("\n" + "="*60)

    def export_markdown(self, filepath: Optional[Path] = None) -> str:
        """Export history as markdown report"""
        stats = self.get_stats()
        fixes = self.get_fixes_summary()

        md = f"""# Ops Phoenix - Run History

## Summary

| Metric | Count |
|--------|-------|
| Total runs | {stats['total_runs']} |
| Healthy runs | {stats['healthy_runs']} |
| Complete runs | {stats['complete_runs']} |
| Failed runs | {stats['failed_runs']} |
| Issues created | {stats['total_issues']} |
| PRs created | {stats['total_prs']} |
| PRs merged | {stats['total_prs_merged']} |
| Deploys triggered | {stats['total_deploys']} |
| Deploys succeeded | {stats['deploys_succeeded']} |
| Deploys failed | {stats['deploys_failed']} |
| Auto-fixes | {stats['total_auto_fixes']} |
| Auto-fixes succeeded | {stats['auto_fixes_succeeded']} |
| Manual issues | {stats['total_manual_issues']} |

## Fixes Applied

"""
        if fixes:
            for fix in fixes:
                md += f"""### PR #{fix['pr_number']}: {fix['pr_title']}

- **Date:** {fix['timestamp'][:10]}
- **Mode:** {fix['run_mode']}
- **URL:** {fix['pr_url']}
- **Errors Fixed:**

"""
                for err in fix['errors_fixed']:
                    md += f"  - {err}\n"
                md += "\n"
        else:
            md += "No fixes recorded yet.\n"

        md += """
## Recent Runs

| Date | Mode | Status | Errors | Issues | PRs | Deploy |
|------|------|--------|--------|--------|-----|--------|
"""
        for run in self.get_recent_runs(limit=20):
            md += f"| {run.timestamp[:10]} | {run.mode} | {run.status} | {run.errors_found} | {len(run.issues_created)} | {len(run.prs_created)} | {len(run.deploys_triggered)} |\n"

        md += f"""

---
*Generated by Ops Phoenix on {datetime.now().isoformat()}*
"""

        if filepath:
            filepath.parent.mkdir(parents=True, exist_ok=True)
            with open(filepath, 'w') as f:
                f.write(md)

        return md

    def _run_state(self, run: RunRecord) -> str:
        """Infer lifecycle state from a run record"""
        if run.status == "healthy":
            return "healthy"
        if run.deploys_succeeded or run.status == "success":
            return "deployed"
        if run.prs_created:
            return "pr_open"
        if run.issues_created:
            return "issue_open"
        if run.errors_found > 0:
            return "detected"
        if isinstance(run.status, str) and run.status.startswith("failed"):
            return "failed"
        return run.status or "unknown"

    def _error_state_summary(self) -> List[Dict]:
        """Aggregate state for each detected error type across all runs"""
        summary: Dict[str, Dict] = {}
        runs = sorted(self.runs, key=lambda r: r.timestamp)

        for run in runs:
            if not run.error_types:
                continue

            for error_type, count in run.error_types.items():
                if error_type not in summary:
                    summary[error_type] = {
                        "error_type": error_type,
                        "total_occurrences": 0,
                        "runs_seen": 0,
                        "first_seen": run.timestamp,
                        "last_seen": run.timestamp,
                        "state": "detected",
                        "latest_issue": None,
                        "latest_pr": None,
                        "latest_deploy": None,
                    }

                item = summary[error_type]
                item["total_occurrences"] += int(count)
                item["runs_seen"] += 1
                item["last_seen"] = run.timestamp
                item["state"] = self._run_state(run)

                if run.issues_created:
                    issue = run.issues_created[-1]
                    item["latest_issue"] = issue.get("number")

                if run.prs_created:
                    pr = run.prs_created[-1]
                    item["latest_pr"] = pr.get("number")

                if run.deploys_triggered:
                    deploy = run.deploys_triggered[-1]
                    item["latest_deploy"] = deploy.get("run_id")

        items = list(summary.values())
        items.sort(
            key=lambda i: (i["runs_seen"], i["total_occurrences"], i["last_seen"]),
            reverse=True,
        )
        return items

    def export_state_markdown(self, filepath: Optional[Path] = None) -> str:
        """Export current state snapshot for agent memory tracking"""
        stats = self.get_stats()
        error_states = self._error_state_summary()
        last_run = self.get_recent_runs(limit=1)
        last_run = last_run[0] if last_run else None

        md = f"""# Ops Phoenix Memory State

Last updated: {datetime.now().isoformat()}

## Current Cycle State

| Field | Value |
|-------|-------|
| Total runs | {stats['total_runs']} |
| Last run status | {last_run.status if last_run else 'none'} |
| Last run timestamp | {last_run.timestamp if last_run else 'none'} |
| Total issues | {stats['total_issues']} |
| Total PRs | {stats['total_prs']} |
| Deploy success/fail | {stats['deploys_succeeded']}/{stats['deploys_failed']} |

## Repeated Issue States

| Error type | State | Runs seen | Total occurrences | Last seen | Issue | PR | Deploy run |
|------------|-------|-----------|-------------------|-----------|-------|----|------------|
"""
        if error_states:
            for item in error_states:
                md += (
                    f"| {item['error_type']} | {item['state']} | {item['runs_seen']} | "
                    f"{item['total_occurrences']} | {item['last_seen'][:19]} | "
                    f"{item['latest_issue'] or '-'} | {item['latest_pr'] or '-'} | "
                    f"{item['latest_deploy'] or '-'} |\n"
                )
        else:
            md += "| none | healthy | 0 | 0 | - | - | - | - |\n"

        md += """
## State Policy

- New error signature: create issue + PR, then monitor deploy state.
- Repeated error signature: update this state table instead of opening duplicate tracking records.
- Deploy failure: state moves to `failed` until next successful deploy.
"""

        if filepath:
            filepath.parent.mkdir(parents=True, exist_ok=True)
            with open(filepath, 'w') as f:
                f.write(md)

        return md


def main():
    """CLI for history tracker"""
    tracker = HistoryTracker()

    import sys
    if len(sys.argv) > 1:
        if sys.argv[1] == "--stats":
            tracker.print_summary()
        elif sys.argv[1] == "--export":
            path = Path(sys.argv[2]) if len(sys.argv) > 2 else OPS_PHOENIX_DIR / "history.md"
            content = tracker.export_markdown(path)
            print(f"Exported to: {path}")
        elif sys.argv[1] == "--fixes":
            fixes = tracker.get_fixes_summary()
            for fix in fixes:
                print(f"[{fix['timestamp'][:10]}] PR #{fix['pr_number']}: {fix['pr_title']}")
        else:
            print("Usage: history.py [--stats|--export [path]|--fixes]")
    else:
        tracker.print_summary()


if __name__ == "__main__":
    main()
