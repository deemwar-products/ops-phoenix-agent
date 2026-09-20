"""History tracker - persistent run records for Ops Phoenix."""

import json
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Optional

from ops_phoenix.models import RunRecord

class HistoryTracker:
 """Tracks Ops Phoenix runs, errors, issues, PRs, and deploys."""
 
 def __init__(self, data_dir: Path):
  self.data_dir = data_dir
  self.data_dir.mkdir(parents=True, exist_ok=True)
  self.history_file = self.data_dir / 'history.json'
  self.memory_file = self.data_dir / 'memory.md'
  self.runs: List[RunRecord] = []
  self.current_run: Optional[RunRecord] = None
  self.load()
 
 def load(self):
  """Load history from disk."""
  if self.history_file.exists():
   try:
    data = json.loads(self.history_file.read_text())
    self.runs = [RunRecord(**r) for r in data.get('runs', [])]
   except (json.JSONDecodeError, TypeError):
    self.runs = []
 
 def save(self):
  """Persist history to disk."""
  data = {
   "last_updated": datetime.now().isoformat(),
   "runs": [self._asdict(r) for r in self.runs],
  }
  self.history_file.write_text(json.dumps(data, indent=2))
  self.export_state_markdown()
 
 @staticmethod
 def _asdict(r):
  """Convert RunRecord to dict."""
  return {
   "timestamp": r.timestamp,
   "mode": r.mode,
   "time_window": r.time_window,
   "status": r.status,
   "errors_found": r.errors_found,
   "error_types": r.error_types,
   "issues_created": r.issues_created,
   "prs_created": r.prs_created,
   "prs_merged": r.prs_merged,
   "deploys_triggered": r.deploys_triggered,
   "deploys_succeeded": r.deploys_succeeded,
   "deploys_failed": r.deploys_failed,
   "auto_fixes": r.auto_fixes,
   "manual_issues": r.manual_issues,
   "duration_seconds": r.duration_seconds,
   "error_message": r.error_message,
   "issue_num": r.issue_num,
   "pr_num": r.pr_num,
   "workflow_run_id": r.workflow_run_id,
  }
 
 def start_run(self, mode='dev', time_window='1h'):
  """Begin a new run record."""
  self.current_run = RunRecord(
   timestamp=datetime.now().isoformat(),
   mode=mode,
   time_window=time_window,
  )
 
 def add_errors(self, errors, error_types):
  """Record errors found during detection."""
  if self.current_run:
   self.current_run.errors_found = len(errors)
   self.current_run.error_types = error_types
 
 def add_issue(self, issue_num, title, url):
  """Record a created issue."""
  if self.current_run:
   self.current_run.issues_created += 1
   self.current_run.issue_num = issue_num
 
 def add_pr(self, pr_num, title, url, merged=False):
  """Record a created PR."""
  if self.current_run:
   self.current_run.prs_created += 1
   self.current_run.pr_num = pr_num
   if merged:
    self.current_run.prs_merged += 1
   if 'auto_fixes' not in self.current_run.error_types:
    self.current_run.auto_fixes = 0
   self.current_run.auto_fixes += 1
 
 def add_deploy(self, workflow_run_id, status, workflow_name='unknown'):
  """Record a deployment."""
  if self.current_run:
   self.current_run.deploys_triggered += 1
   self.current_run.workflow_run_id = workflow_run_id
   if 'success' in status:
    self.current_run.deploys_succeeded += 1
   elif 'fail' in status or 'timeout' in status:
    self.current_run.deploys_failed += 1
 
 def complete_run(self, status, error_message=''):
  """Finalize the current run record."""
  if self.current_run:
   self.current_run.status = status
   if error_message:
    self.current_run.error_message = error_message
   self.current_run.duration_seconds = (datetime.now() - datetime.fromisoformat(self.current_run.timestamp)).total_seconds()
   self.runs.append(self.current_run)
   self.current_run = None
   self.save()
 
 def export_state_markdown(self):
  """Export recent state as markdown for debugging."""
  lines = []
  lines.append('# Ops Phoenix History')
  lines.append('')
  lines.append(f'Last updated: {datetime.now().isoformat()}')
  lines.append('')
  for run in self.runs[-10:]:
   lines.append(f"- **{run.timestamp}** [{run.mode.upper()}] {run.status}")
   lines.append(f" Errors: {run.errors_found}, Issues: {run.issues_created}, PRs: {run.prs_created}")
  self.memory_file.write_text(chr(10).join(lines))
 
 def recent_issues(self, limit=10):
  """Return list of recent issue titles to prevent duplicates."""
  result = []
  for run in reversed(self.runs):
   if run.issue_num:
    result.append(f"#{run.issue_num}: {run.error_types}")
   if len(result) >= limit:
    break
  return result
