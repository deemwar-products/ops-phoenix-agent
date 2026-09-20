"""Threshold engine for Ops Phoenix."""

from typing import Dict, List, Optional

class ThresholdConfig:
 """Configuration for error thresholds."""
 
 def __init__(self):
  self.min_errors_to_act = 1
  self.min_unique_types = 1
  self.issue_cooldown_minutes = 60
  self.pr_cooldown_minutes = 120
  self.max_retries_per_cycle = 2
  self.auto_fix_enabled = False
  self.severity_thresholds = {
   "critical": 1,
   "high": 3,
   "medium": 5,
   "low": 10,
  }
 
class DecisionEngine:
 """Decides whether to act on detected errors."""
 
