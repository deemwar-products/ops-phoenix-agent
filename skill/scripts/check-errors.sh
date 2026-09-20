#!/usr/bin/env bash
# check-errors.sh — SRE Phoenix agent script for cron jobs
# Usage: ./check-errors.sh [--create-issue] [--dry-run]
# Requirements: curl, jq, gh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_DIR="/tmp/sre-phoenix"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
CREATE_ISSUE=false
DRY_RUN=false
TIME_WINDOW="1h"

# Default paths (can override with env vars)
GRAFANA_TOKEN_PATH="${GRAFANA_TOKEN_PATH:-${HOME}/Downloads/Archive/keys/grafana-api-token}"

# Parse flags
while [[ $# -gt 0 ]]; do
  case "$1" in
    --create-issue) CREATE_ISSUE=true ;;
    --dry-run) DRY_RUN=true ;;
    --window) TIME_WINDOW="$2"; shift ;;
    -h|--help)
      echo "Usage: $0 [--create-issue] [--dry-run] [--window 1h|6h|24h|7d]"
      echo "  --create-issue  Create GitHub issue (default: false)"
      echo "  --dry-run     Don't create issue, just print errors found"
      echo "  --window      Time window (default: 1h)"
      exit 0 ;;
    --token-path) GRAFANA_TOKEN_PATH="$2"; shift ;;
  esac
  shift
done

echo "[$(date '+%Y-%m-%d %H:%M:%S')] SRE Phoenix check started"

# Preflight checks
check_binaries() {
  for bin in curl jq gh python3; do
    command -v "$bin" >/dev/null || { echo "Missing: $bin"; exit 1; }
  done
  echo "Binaries OK"
}

check_credentials() {
  # Check Claude API key
  [[ -n "${ANTHROPIC_API_KEY:-}" ]] || {
    echo "ERROR: ANTHROPIC_API_KEY not set"
    echo "Set with: export ANTHROPIC_API_KEY=sk-ant-..."
    exit 1
  }
  echo "Claude API key: set"

  # Check Grafana token
  [[ -f "$GRAFANA_TOKEN_PATH" ]] || {
    echo "ERROR: Grafana token not found: $GRAFANA_TOKEN_PATH"
    exit 1
  }
  echo "Grafana token: found"
}

# Map time window to Grafana duration string
get_time_range() {
  local window="$1"
  case "$window" in
    1h) echo "now-1h" "now" ;;
    6h) echo "now-6h" "now" ;;
    24h) echo "now-24h" "now" ;;
    7d) echo "now-7d" "now" ;;
    *) echo "now-1h" "now" ;;
  esac
}

# Query Grafana/Loki for errors
query_errors() {
  local window="$1"
  local token
  token=$(cat "$GRAFANA_TOKEN_PATH")

  # Get time range strings
  read -r from_ts to_ts <<< "$(get_time_range "$window")"

  echo "Querying Loki for errors in last $window..."

  # Build the query JSON - using json filter for structured logs
  # This works: {host=~"app1|app2"} | json | level=~"(?i)error|fatal"
  local query_json='{"queries":[{"expr":"{host=~\"app1|app2\"} | json | level=~\"(?i)error|fatal\"","refId":"A","datasource":{"type":"loki","uid":"loki"},"maxLines":100}],"from":"'$from_ts'","to":"'$to_ts'"}'

  # Query Grafana Loki API
  local response
  response=$(curl -sS -H "Authorization: Bearer ${token}" \
    -H "Content-Type: application/json" \
    "https://observability.deemwar.com/api/ds/query?ds_type=loki" \
    --data "$query_json" 2>&1) || true

  echo "$response"
}

# Parse error counts from Loki response
parse_errors() {
  local response="$1"

  # Check for errors in response
  local status
  status=$(echo "$response" | jq -r '.results.A.status // "error"' 2>/dev/null)

  if [[ "$status" != "200" ]]; then
    local error_msg
    error_msg=$(echo "$response" | jq -r '.results.A.error // "Unknown error"' 2>/dev/null)
    echo "Loki query error: $error_msg"
    return 1
  fi

  # Extract log lines from response
  local log_lines
  log_lines=$(echo "$response" | jq -r '.results.A.frames[0].data.values[2][]? // empty' 2>/dev/null)

  if [[ -z "$log_lines" ]]; then
    echo "No error logs found"
    return 1
  fi

  echo "$log_lines"
}

# Analyze with Claude
analyze_errors() {
  local errors="$1"
  [[ -z "$errors" ]] && return 1

  echo "Analyzing with Claude AI..."

  local prompt="You are an SRE analyzing production errors. Summarize in 3 bullet points:
1. Root cause
2. Affected functionality
3. Suggested fix

Errors from Reqsume API production logs:

$errors"

  curl -sS -X POST "https://api.anthropic.com/v1/messages" \
    -H "x-api-key: ${ANTHROPIC_API_KEY}" \
    -H "anthropic-version: 2023-06-01" \
    -H "content-type: application/json" \
    --data-binary "$(jq -n \
      --arg prompt "$prompt" \
      '{
        model: "claude-opus-4-7",
        max_tokens: 512,
        messages: [{"role": "user", "content": $prompt}]
      }')" 2>/dev/null | \
    jq -r '.content[0].text // empty'
}

# Create GitHub issue
create_issue() {
  local summary="$1"
  local errors="$2"
  local analysis="$3"

  echo "Creating GitHub issue..."

  # Check for duplicates (skip if found in last 48h)
  local existing
  existing=$(gh issue list --label sre-alert --state all --limit 10 --json number,title 2>/dev/null | \
    jq -r '.[] | "\(.number) \(.title)"' | \
    grep -i "$(echo "$summary" | cut -d: -f1 | cut -d' ' -f1)" | head -1)

  if [[ -n "$existing" ]]; then
    local existing_num
    existing_num=$(echo "$existing" | awk '{print $1}')
    echo "Duplicate issue found: #$existing_num"
    echo "Adding comment to existing issue..."
    gh issue comment "$existing_num" --body "$(cat <<EOF
## SRE Phoenix Update $(date '+%Y-%m-%d %H:%M')

### New Occurrences

$errors

### Analysis

$analysis

---
*Auto-generated by SRE Phoenix*
EOF
)"
    return 0
  fi

  # Create new issue
  local issue_url
  issue_url=$(gh issue create \
    --title "[sre-alert] $summary" \
    --body "$(cat <<EOF
## SRE Alert

**Detected:** $(date '+%Y-%m-%d %H:%M')
**Environment:** production
**Severity:** warning
**Window:** $TIME_WINDOW

### Errors Found

$errors

### Analysis

$analysis

### Links

- [Grafana Dashboard](https://observability.deemwar.com/d/reqsume-logs)

---
*Generated by SRE Phoenix agent*
EOF
)" \
    --label "sre-alert" \
    --label "bug" 2>/dev/null)

  echo "Issue created: $issue_url"
}

# Main
main() {
  mkdir -p "$LOG_DIR"

  check_binaries
  check_credentials

  echo ""
  echo "=== SRE Phoenix Log Check ==="
  echo ""

  local errors
  errors=$(parse_errors "$(query_errors "$TIME_WINDOW")")

  if [[ -z "$errors" || "$errors" == "No error logs found" ]]; then
    echo ""
    echo "✓ No errors found in last $TIME_WINDOW"
    exit 0
  fi

  echo ""
  echo "Errors detected (counting occurrences)..."

  # Count errors by type
  local error_count
  error_count=$(echo "$errors" | wc -l)
  echo "Total error log lines: $error_count"

  # Show sample
  echo ""
  echo "Sample errors:"
  echo "$errors" | head -5 | while read line; do
    echo "  - ${line:0:120}..."
  done

  # Truncate for analysis (first 5000 chars)
  errors=$(echo "$errors" | head -c 5000)

  echo ""

  # Analyze with Claude
  local analysis
  analysis=$(analyze_errors "$errors")

  echo ""
  echo "Claude Analysis:"
  echo "$analysis"
  echo ""

  if [[ "$DRY_RUN" == "true" ]]; then
    echo "Dry run complete. Use --create-issue to create GitHub issue."
    exit 0
  fi

  if [[ "$CREATE_ISSUE" == "true" ]]; then
    local summary
    summary=$(echo "$errors" | head -3 | grep -o '"msg": "[^"]*"' | head -1 | sed 's/"msg": "//' | sed 's/"//')
    [[ -z "$summary" ]] && summary="Production errors detected"
    create_issue "$summary" "$errors" "$analysis"
  fi

  echo ""
  echo "SRE Phoenix check complete"
}

main "$@"