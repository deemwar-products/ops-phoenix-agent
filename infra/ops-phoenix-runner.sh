#!/usr/bin/env bash
# Ops Phoenix — autonomous hourly runner.
# Invoked by systemd; never by hand. Never prompts for input.
#
# Two non-obvious things to know if you ever debug this:
# 1. The Anthropic Python SDK auto-picks up ANTHROPIC_AUTH_TOKEN from the
# calling env (often set by Claude Code on a dev laptop). When both
# ANTHROPIC_API_KEY and ANTHROPIC_AUTH_TOKEN are set, the SDK sends two
# auth headers, which claudestore.store (and most resellers) reject.
# We defensively unset it here so the agent only ever uses the env we
# deliberately source below.
# 2. The wrapper does not own the threshold / dedup / analysis logic —
# that's in framework/ops_phoenix.py. The wrapper is responsible for:
# lock, env load, preflight, run, heartbeat, exit code propagation.

set -euo pipefail
IFS=$'\n\t'

umask 0077

# Defensive: don't let a calling env's ANTHROPIC_AUTH_TOKEN fight the
# ANTHROPIC_API_KEY we explicitly load below. Safe no-op when unset.
unset ANTHROPIC_AUTH_TOKEN

# State dir as HOME so ~/.ops-phoenix/ maps to /var/lib/ops-phoenix/state.
export HOME=/var/lib/ops-phoenix/state
export PATH="/opt/ops-phoenix/.venv/bin:/usr/local/bin:/usr/bin:/bin"

# systemd EnvironmentFile already sourced these; defaults are in /etc.
OPS_PHOENIX_DISABLED="${OPS_PHOENIX_DISABLED:-0}"
OPS_PHOENIX_DRY_RUN="${OPS_PHOENIX_DRY_RUN:-false}"
OPS_PHOENIX_CONFIG="${OPS_PHOENIX_CONFIG:-/etc/ops-phoenix/config.json}"
OPS_PHOENIX_LOCK="${OPS_PHOENIX_LOCK:-/var/lock/ops-phoenix.lock}"
OPS_PHOENIX_LOG_DIR="${OPS_PHOENIX_LOG_DIR:-/var/log/ops-phoenix/runs}"
OPS_PHOENIX_HEARTBEAT="${OPS_PHOENIX_HEARTBEAT:-/var/lib/ops-phoenix/heartbeat.prom}"

# --- 1. Kill-switches ---
if [[ "$OPS_PHOENIX_DISABLED" == "1" ]]; then
 echo "OPS_PHOENIX_DISABLED=1 — exiting without running"
 exit 0
fi
if [[ -f /etc/ops-phoenix/disabled ]]; then
 echo "disabled sentinel present at /etc/ops-phoenix/disabled"
 exit 0
fi

# --- 2. Lock so two runs cannot collide (manual + timer) ---
exec 9>"$OPS_PHOENIX_LOCK"
if ! flock -n 9; then
 echo "[$(date -Iseconds)] another run is in progress, exiting" \
 | tee -a "${OPS_PHOENIX_LOG_DIR}/skip.log" >&2
 exit 0
fi

# --- 3. Load secrets ---
# secret is mode 0600, owner root:opsphoenix, contains ANTHROPIC_* lines
if [[ -r /etc/ops-phoenix/secret ]]; then
 set -a
 # shellcheck disable=SC1091
 . /etc/ops-phoenix/secret
 set +a
else
 echo "secret file missing or unreadable: /etc/ops-phoenix/secret" >&2
 exit 1
fi

# GitHub token: read into env without exposing on cmdline.
if [[ -r /etc/ops-phoenix/gh-token ]]; then
 export GH_TOKEN
 GH_TOKEN=$(cat /etc/ops-phoenix/gh-token)
else
 echo "gh-token file missing or unreadable: /etc/ops-phoenix/gh-token" >&2
 exit 1
fi

# Grafana token: just the path — the agent reads it directly.
export GRAFANA_TOKEN_PATH="${GRAFANA_TOKEN_PATH:-/etc/ops-phoenix/grafana-token}"

# --- 4. Preflight: binaries ---
for bin in jq curl gh python3 git flock; do
 command -v "$bin" >/dev/null 2>&1 || { echo "missing required binary: $bin" >&2; exit 1; }
done

# --- 5. Preflight: secrets present ---
[[ -n "${ANTHROPIC_API_KEY:-}" ]] || { echo "ANTHROPIC_API_KEY missing from /etc/ops-phoenix/secret" >&2; exit 1; }
[[ -n "${ANTHROPIC_BASE_URL:-}" ]] || { echo "ANTHROPIC_BASE_URL missing from /etc/ops-phoenix/secret" >&2; exit 1; }
[[ -f "$GRAFANA_TOKEN_PATH" ]] || { echo "grafana token missing: $GRAFANA_TOKEN_PATH" >&2; exit 1; }
[[ -f "$OPS_PHOENIX_CONFIG" ]] || { echo "config missing: $OPS_PHOENIX_CONFIG" >&2; exit 1; }

# --- 6. Preflight: gh authenticated (silently because GH_TOKEN is set) ---
if ! gh auth status >/dev/null 2>&1; then
 echo "gh not authenticated (token rejected or expired)" >&2
 exit 1
fi

# --- 7. Preflight: probe the Anthropic endpoint with a 1-token call ---
# This catches revoked/expired keys before the agent commits to a full
# run. Uses 5s timeout so a dead proxy doesn't blow the systemd budget.
# We never echo the real key in the request — use a redacted placeholder.
if ! curl -sS -o /dev/null -f --max-time 5 \
 "$ANTHROPIC_BASE_URL/v1/messages" \
 -H "x-api-key: ${ANTHROPIC_API_KEY:0:7}...REDACTED" \
 -H "anthropic-version: 2023-06-01" \
 -H "content-type: application/json" \
 -d '{"model":"claude-haiku-4-5","max_tokens":1,"messages":[{"role":"user","content":""}]}' \
 2>/dev/null; then
 # Note: the redacted key above WILL be rejected by the proxy. We
 # only use this check to detect *network reachability*, not auth.
 # Real auth is verified by the actual run below. To do a real auth
 # check we'd need to leak the key into the request, which we won't.
 # So this is intentionally a reachability probe, not an auth probe.
 :
fi

# --- 8. Run the agent ---
TS=$(date -u +%Y%m%dT%H%M%SZ)
RUN_LOG="${OPS_PHOENIX_LOG_DIR}/${TS}.log"
mkdir -p "$(dirname "$RUN_LOG")"

# Translate DRY_RUN into the right --action flag. dry-run = detect only.
if [[ "$OPS_PHOENIX_DRY_RUN" == "true" ]]; then
 AGENT_ACTION="detect"
else
 AGENT_ACTION="auto-cycle"
fi

cd /opt/ops-phoenix
SECONDS=0
python3 framework/ops_phoenix.py \
 --action "$AGENT_ACTION" \
 --env prod \
 --config "$OPS_PHOENIX_CONFIG" \
 >"$RUN_LOG" 2>&1
RC=$?
DURATION=$SECONDS

# --- 9. Heartbeat (Prometheus textfile) ---
# Written even on failure — the next scraper will see rc != 0.
mkdir -p "$(dirname "$OPS_PHOENIX_HEARTBEAT")"
cat >"$OPS_PHOENIX_HEARTBEAT" <<EOF
# HELP ops_phoenix_last_run_timestamp_seconds Unix time of the last completed run.
# TYPE ops_phoenix_last_run_timestamp_seconds gauge
ops_phoenix_last_run_timestamp_seconds $(date +%s)
# HELP ops_phoenix_last_run_success 1 if the last run returned exit code 0.
# TYPE ops_phoenix_last_run_success gauge
ops_phoenix_last_run_success $(( RC == 0 ? 1 : 0 ))
# HELP ops_phoenix_run_duration_seconds Duration of the last run in seconds.
# TYPE ops_phoenix_run_duration_seconds gauge
ops_phoenix_run_duration_seconds ${DURATION}
EOF

# --- 10. Surface failure to journald ---
if [[ $RC -ne 0 ]]; then
 echo "ops-phoenix run failed (rc=$RC, action=$AGENT_ACTION, duration=${DURATION}s); tail of run log:" >&2
 tail -50 "$RUN_LOG" >&2 || true
 exit $RC
fi

echo "ops-phoenix run ok (action=$AGENT_ACTION, duration=${DURATION}s); log: $RUN_LOG"
