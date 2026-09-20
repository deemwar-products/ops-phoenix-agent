"""Microsoft Teams notification via Power Automate flow."""

import json
import subprocess
from typing import Any, Dict, List, Optional

# Power Automate expects this shape so the flow can parse title, body,
# and severity from the JSON body and forward to Teams via
# "Post message in a chat or channel".
PAYLOAD_TEMPLATE = {
 "title": "Ops Phoenix Alert",
 "message": "",
 "severity": "info",
 "details": [],
 "actions": [],
}


def send_teams_message(webhook_url: str, payload: Dict[str, Any]) -> bool:
 """POST a JSON payload to the Power Automate flow endpoint.

 The flow's "Post message in a chat or channel" action should be
 configured with:
 - Team / Channel: your target (set once in the flow UI)
 - Message: the value from the payload's 'message' field
 - (optional) Subject: the value from 'title'

 Expected Power Automate flow:
 Trigger: When a HTTP request is received
 Action: Post message in a chat or channel (Microsoft Teams)
 -> Channel: <your team / channel>
 -> Message: @{triggerBody()?['message']}
 -> Subject: @{triggerBody()?['title']} (optional)
 """
 result = subprocess.run(
  [
   "curl", "-sS", "-X", "POST",
   "-H", "Content-Type: application/json",
   "--data", json.dumps(payload, ensure_ascii=False),
   webhook_url,
  ],
  capture_output=True,
  text=True,
  timeout=15,
  )
 
 if result.returncode != 0:
  print(f"Power Automate request failed: {result.stderr[:200]}")
  return False
 
 # Power Automate returns 202 Accepted on success
 response = result.stdout.strip()
 if response in ("", "202", "202"):
  return True
 
 print(f"Power Automate unexpected response: {response[:200]}")
 return True


def build_alert_payload(
 title: str,
 message: str,
 severity: str = "info",
 details: Optional[List[Dict[str, str]]] = None,
 actions: Optional[List[str]] = None,
) -> Dict[str, Any]:
 """Build a payload for Ops Phoenix -> Power Automate -> Teams."""
 payload = dict(PAYLOAD_TEMPLATE)
 payload["title"] = title
 payload["message"] = message
 payload["severity"] = severity
 
 if details:
  payload["details"] = details
 
 if actions:
  payload["actions"] = actions
 
 return payload


# ---- Specific card builders ----

SEVERITY_COLORS = {
 "critical": "FF0000",
 "high": "FF6D00",
 "medium": "FFB900",
 "low": "FFE600",
 "info": "0076D7",
}


def errors_detected_card(
 error_types: Dict[str, int], time_window: str, analysis: Optional[Dict] = None
) -> Dict[str, Any]:
 """Payload for the initial error detection step."""
 total = sum(error_types.values())
 unique = len(error_types)
 top = sorted(error_types.items(), key=lambda x: -x[1])[:5]
 error_lines = "\n".join(f"- {k}: {v}x" for k, v in top)
 
 severity = "info"
 title = "Ops Phoenix: Errors Detected"
 
 if analysis:
  sev = analysis.get("severity", "info").lower()
  severity = sev if sev in SEVERITY_COLORS else "info"
  title = (
   f"Ops Phoenix: {analysis.get('severity', 'Error').upper()} - "
   f"{analysis.get('root_cause', 'Unknown')[:60]}"
  )
 
 message = (
  f"Found **{total} errors** ({unique} unique) in the last **{time_window}**.\n\n"
  f"{error_lines}"
 )
 
 details = [
  {"name": "Total errors", "value": str(total)},
  {"name": "Unique signatures", "value": str(unique)},
  {"name": "Time window", "value": time_window},
 ]
 
 if analysis:
  details.extend(
   [
    {"name": "Severity", "value": analysis.get("severity", "N/A")},
    {"name": "Root cause", "value": analysis.get("root_cause", "N/A")[:200]},
    {"name": "Fix suggestion", "value": analysis.get("fix_suggestion", "N/A")[:200]},
   ]
  )
 
 detail_lines = "\n".join(f"**{d['name']}:** {d['value']}" for d in details)
 message = f"{message}\n\n{detail_lines}"
 
 return build_alert_payload(title=title, message=message, severity=severity, details=details)


def issue_created_card(
 issue_num: int, issue_url: str, error_types: Dict[str, int]
) -> Dict[str, Any]:
 """Payload for GitHub issue creation."""
 top = sorted(error_types.items(), key=lambda x: -x[1])[0][0][:80]
 message = f"Issue **#{issue_num}** created for: *{top}*\n[View Issue]({issue_url})"
 
 details = [
  {"name": "Issue", "value": f"#{issue_num}"},
  {"name": "Primary error", "value": top},
  {"name": "Status", "value": "Awaiting triage"},
  {"name": "URL", "value": issue_url},
 ]
 
 return build_alert_payload(
  title=f"Issue #{issue_num} Created",
  message=message,
  severity="info",
  details=details,
 )


def pr_created_card(
 pr_num: int, pr_url: str, tests_ok: bool, analysis: Optional[Dict] = None
) -> Dict[str, Any]:
 """Payload for PR creation."""
 root_cause = (analysis or {}).get("root_cause", "Unknown")[:80]
 test_status = "PASS" if tests_ok else "FAIL - manual review needed"
 severity = "info" if tests_ok else "high"
 
 message = (
  f"PR **#{pr_num}** opened to fix: *{root_cause}*\n"
  f"Tests: **{test_status}**\n"
  f"[Review PR]({pr_url})"
 )
 
 details = [
  {"name": "PR", "value": f"#{pr_num}"},
  {"name": "Root cause", "value": root_cause},
  {"name": "Tests", "value": test_status},
  {"name": "URL", "value": pr_url},
 ]
 
 return build_alert_payload(
  title=f"PR #{pr_num}: Auto-Fix Proposed",
  message=message,
  severity=severity,
  details=details,
 )


def deploy_result_card(
 run_id: str, result: str, pr_num: Optional[int] = None
) -> Dict[str, Any]:
 """Payload for deploy monitoring result."""
 status_map = {
  "success": ("Deployed Successfully", "info"),
  "failed": ("Deploy Failed", "critical"),
  "timeout": ("Deploy Monitoring Timed Out", "high"),
 }
 title, severity = status_map.get(result, ("Deploy Complete", "info"))
 
 message = f"Deployment workflow `{run_id}` finished with status: **{result.upper()}**"
 if pr_num:
  message += f"\nMerged from PR **#{pr_num}**"
 
 details = [
  {"name": "Workflow run", "value": run_id},
  {"name": "Result", "value": result.upper()},
  {"name": "PR", "value": f"#{pr_num}" if pr_num else "N/A"},
 ]
 
 return build_alert_payload(
  title=title, message=message, severity=severity, details=details
 )
