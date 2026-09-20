"""Run record data model for Ops Phoenix history tracking."""

from dataclasses import dataclass, field
from datetime import datetime
from typing import Dict, List, Optional


@dataclass
class RunRecord:
 """Snapshot of a single Ops Phoenix run."""
 timestamp: str
 mode: str
 time_window: str
 status: str
 errors_found: int
 error_types: Dict[str, int] = field(default_factory=dict)
 issues_created: List[dict] = field(default_factory=list)
 prs_created: List[dict] = field(default_factory=list)
 prs_merged: List[dict] = field(default_factory=list)
 deploys_triggered: List[dict] = field(default_factory=list)
 deploys_succeeded: List[dict] = field(default_factory=list)
 deploys_failed: List[dict] = field(default_factory=list)
 auto_fixes: List[dict] = field(default_factory=list)
 manual_issues: List[dict] = field(default_factory=list)
 duration_seconds: Optional[float] = None
 error_message: Optional[str] = None
