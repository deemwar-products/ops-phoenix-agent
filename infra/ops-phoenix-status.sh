#!/usr/bin/env bash
# Ops Phoenix — status read-out for operators.
#
# Prints the last 5 runs (from /var/log/ops-phoenix/runs/*.log),
# the last heartbeat (Prometheus textfile), and the current rate-limit
# state (from the agent's state file).
#
# Read-only. Never invokes the runner. Never blocks on anything.
# Safe to run over SSH as any user with read access to those paths.

set -euo pipefail

LOG_DIR="/var/log/ops-phoenix/runs"
HEARTBEAT="${OPS_PHOENIX_HEARTBEAT:-/var/lib/ops-phoenix/heartbeat.prom}"
STATE_FILE="${OPS_PHOENIX_STATE:-/var/lib/ops-phoenix/state/.ops-phoenix/threshold_state.json}"
RUNS_TO_SHOW=5

print_section() {
 printf '\n=== %s ===\n' "$1"
}

# 1. Last 5 runs (most recent first).
print_section "Last ${RUNS_TO_SHOW} runs"
if [[ -d "$LOG_DIR" ]]; then
 # ls -1t gives newest first; -1 | head -n 5 gives us the last 5 timestamps.
 while IFS= read -r log; do
 [[ -z "$log" ]] && continue
 # Status: rc comes from a sidecar file the runner writes as ${log}.rc; if
 # absent (e.g. pre-spec deployments), fall back to grepping the log for
 # the last "ERROR" or "Traceback" line.
 rc_file="${log}.rc"
 if [[ -f "$rc_file" ]]; then
 rc=$(cat "$rc_file")
 if [[ "$rc" == "0" ]]; then status="success"; else status="failure(rc=$rc)"; fi
 else
 if tail -5 "$log" 2>/dev/null | grep -qE "Traceback|ERROR |FATAL "; then
 status="failure(unparsed)"
 else
 status="success"
 fi
 fi
 # Duration: parse the ISO timestamp from the filename.
 ts=$(basename "$log" .log)
 printf ' %s status=%s\n' "$ts" "$status"
 done < <(ls -1t "$LOG_DIR"/*.log 2>/dev/null | head -n "$RUNS_TO_SHOW")
else
 echo " (log dir $LOG_DIR not present)"
fi

# 2. Heartbeat.
print_section "Heartbeat ($HEARTBEAT)"
if [[ -f "$HEARTBEAT" ]]; then
 # Pull each metric line. The textfile format is:
 # # HELP foo ...
 # # TYPE foo gauge
 # foo <value>
 last_ts=$(grep -E "^ops_phoenix_last_run_timestamp_seconds " "$HEARTBEAT" | awk '{print $2}' || true)
 last_ok=$(grep -E "^ops_phoenix_last_run_success " "$HEARTBEAT" | awk '{print $2}' || true)
 last_dur=$(grep -E "^ops_phoenix_run_duration_seconds " "$HEARTBEAT" | awk '{print $2}' || true)
 if [[ -n "$last_ts" ]]; then
 age=$(( $(date +%s) - last_ts ))
 printf ' last_run=%s (age=%ss)\n' "$(date -u -r "$last_ts" -Iseconds 2>/dev/null || echo "$last_ts")" "$age"
 fi
 [[ -n "$last_ok" ]] && printf ' last_run_success=%s\n' "$last_ok"
 [[ -n "$last_dur" ]] && printf ' last_run_duration=%ss\n' "$last_dur"
else
 echo " (no heartbeat file at $HEARTBEAT — has the runner ever run?)"
fi

# 3. Rate-limit state (from the agent's threshold manager).
print_section "Rate limit"
if [[ -f "$STATE_FILE" ]]; then
 python3 - "$STATE_FILE" <<'PY' 2>/dev/null || echo " (could not parse $STATE_FILE)"
import json, sys
from datetime import datetime, timedelta
p = sys.argv[1]
try:
 with open(p) as f:
 s = json.load(f)
except Exception as e:
 print(f" parse error: {e}")
 sys.exit(0)
recent = s.get("recent_runs", [])
last_action = s.get("last_action")
cutoff = datetime.now() - timedelta(hours=1)
recent_hr = 0
for ts in recent:
 try:
 if datetime.fromisoformat(ts) > cutoff:
 recent_hr += 1
 except Exception:
 pass
print(f" runs in last hour: {recent_hr}")
if last_action:
 try:
 la = datetime.fromisoformat(last_action)
 ago = datetime.now() - la
 print(f" last action: {last_action} ({int(ago.total_seconds())}s ago)")
 except Exception:
 print(f" last action: {last_action}")
PY
else
 echo " (no state file at $STATE_FILE)"
fi
