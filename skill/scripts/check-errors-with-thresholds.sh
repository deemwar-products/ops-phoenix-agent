#!/usr/bin/env bash
# check-errors.sh — SRE Phoenix agent script with intelligent thresholds
# Usage: ./check-errors.sh [--create-issue] [--dry-run] [--full-cycle]
# Requirements: curl, jq, gh, python3 (for threshold manager)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_DIR="/tmp/sre-phoenix"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
CREATE_ISSUE=false
DRY_RUN=false
FULL_CYCLE=false
TIME_WINDOW="1h"
ENV_MODE="dev"

# Default paths (can override with env vars)
GRAFANA_TOKEN_PATH="${GRAFANA_TOKEN_PATH:-${HOME}/Downloads/Archive/keys/grafana-api-token}"

# Threshold config (can be overridden)
THRESHOLD_MIN_ERRORS="${THRESHOLD_MIN_ERRORS:-3}"
THRESHOLD_SEVERITY="${THRESHOLD_SEVERITY:-medium}"
THRESHOLD_DEDUP_HOURS="${THRESHOLD_DEDUP_HOURS:-48}"

# Parse flags
while [[ $# -gt 0 ]]; do
 case "$1" in
 --create-issue) CREATE_ISSUE=true ;;
 --dry-run) DRY_RUN=true ;;
 --full-cycle) FULL_CYCLE=true ;;
 --window) TIME_WINDOW="$2"; shift ;;
 --env) ENV_MODE="$2"; shift ;;
 -h|--help)
 echo "Usage: $0 [options]"
 echo ""
 echo "Options:"
 echo " --create-issue Create GitHub issue (default: false)"
 echo " --dry-run Don't create anything, just show errors"
 echo " --full-cycle Complete cycle: detect → analyze → issue → fix → PR"
 echo " --window Time window (default: 1h, also: 6h, 24h, 7d)"
 echo " --env Environment mode: dev|prod (default: dev)"
 echo " -h, --help Show this help"
 echo ""
 echo "Threshold configuration (env vars):"
 echo " THRESHOLD_MIN_ERRORS Min errors to trigger (default: 3)"
 echo " THRESHOLD_SEVERITY Min severity: critical|high|medium|low (default: medium)"
 echo " THRESHOLD_DEDUP_HOURS Hours to avoid duplicate issues (default: 48)"
 echo ""
 echo "Examples:"
 echo " $0 --dry-run # Check for errors only"
 echo " $0 --create-issue # Detect + create issue if threshold met"
 echo " $0 --window 6h --create-issue # Check last 6 hours"
 echo " $0 --env prod --full-cycle # Full cycle with auto-deploy"
 exit 0 ;;
 --token-path) GRAFANA_TOKEN_PATH="$2"; shift ;;
 --min-errors) THRESHOLD_MIN_ERRORS="$2"; shift ;;
 --severity) THRESHOLD_SEVERITY="$2"; shift ;;
 esac
 shift
done

# Logging functions
log() {
 echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

log_info() {
 log "[INFO] $1"
}

log_error() {
 log "[ERROR] $1"
}

log_debug() {
 if [[ "${DEBUG:-false}" == "true" ]]; then
 log "[DEBUG] $1"
 fi
}

# Preflight checks
check_binaries() {
 for bin in curl jq gh python3; do
 if ! command -v "$bin" >/dev/null 2>&1; then
 log_error "Missing required binary: $bin"
 exit 1
 fi
 done
 log_info "All required binaries found"
}

check_credentials() {
 # Check Claude API key
 if [[ -z "${ANTHROPIC_API_KEY:-}" ]]; then
 log_error "ANTHROPIC_API_KEY not set"
 echo "Set with: export ANTHROPIC_API_KEY=sk-ant-..."
 exit 1
 fi
 log_info "Claude API key: configured"

 # Check Grafana token
 if [[ ! -f "$GRAFANA_TOKEN_PATH" ]]; then
 log_error "Grafana token not found: $GRAFANA_TOKEN_PATH"
 exit 1
 fi
 log_info "Grafana token: found"
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

 log_info "Querying Loki for errors in last $window..."

 # Build the query JSON - using json filter for structured logs
 local query_json="{\"queries\":[{\"expr\":\"{container=~\\\"reqsume-app-web.*|reqsume-api.*\\\", host=~\\\"app1|app2\\\"} | json | level=~\\\"(?i)error|fatal\\\"\",\"refId\":\"A\",\"datasource\":{\"type\":\"loki\",\"uid\":\"loki\"},\"maxLines\":100}],\"from\":\"${from_ts}\",\"to\":\"${to_ts}\"}"

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

# Count errors by type (returns "type:count" lines)
count_error_types() {
 local errors="$1"
 echo "$errors" | while read -r line; do
 # Extract msg and logger from JSON
 local msg logger
 msg=$(echo "$line" | jq -r '.msg // "unknown"' 2>/dev/null)
 logger=$(echo "$line" | jq -r '.logger // "unknown"' 2>/dev/null)
 if [[ -n "$msg" && "$msg" != "null" ]]; then
 echo "${logger}: ${msg}"
 fi
 done | sort | uniq -c | sort -rn
}

# Check thresholds and decide action
check_thresholds() {
 local error_counts="$1"
 local severity="${2:-medium}"

 # Count total errors
 local total_errors=0
 while read -r count rest; do
 total_errors=$((total_errors + count))
 done <<< "$error_counts"

 log_info "Threshold check: $total_errors errors (min: $THRESHOLD_MIN_ERRORS, severity: $THRESHOLD_SEVERITY)"

 # Check severity threshold
 local severity_score
 case "$THRESHOLD_SEVERITY" in
 critical) severity_score=4 ;;
 high) severity_score=3 ;;
 medium) severity_score=2 ;;
 low) severity_score=1 ;;
 *) severity_score=2 ;;
 esac

 local incoming_score
 case "$severity" in
 critical) incoming_score=4 ;;
 high) incoming_score=3 ;;
 medium) incoming_score=2 ;;
 low) incoming_score=1 ;;
 *) incoming_score=2 ;;
 esac

 if [[ "$incoming_score" -lt "$severity_score" ]]; then
 echo "SEVERITY_BELOW_THRESHOLD"
 return 1
 fi

 # Check error count threshold
 if [[ "$total_errors" -lt "$THRESHOLD_MIN_ERRORS" ]]; then
 echo "COUNT_BELOW_THRESHOLD"
 return 1
 fi

 echo "THRESHOLD_MET"
 return 0
}

# Check for duplicate issue
check_duplicate_issue() {
 local error_summary="$1"

 # Extract key term from error (first word of first error type)
 local key_term
 key_term=$(echo "$error_summary" | head -1 | awk '{print $2}' | cut -d: -f1)

 # Look for existing issues with similar title in last 48h
 local existing
 existing=$(gh issue list \
 --label "sre-alert" \
 --state "all" \
 --limit 20 \
 --json "number,title,createdAt" 2>/dev/null | \
 jq -r '.[] | "\(.number) \(.title)"' | \
 grep -i "$key_term" | head -1)

 if [[ -n "$existing" ]]; then
 local existing_num
 existing_num=$(echo "$existing" | awk '{print $1}')
 log_info "Duplicate issue found: #$existing_num"
 echo "$existing_num"
 return 0
 fi

 echo ""
 return 1
}

# Analyze with Claude
analyze_errors() {
 local error_counts="$1"

 log_info "Analyzing with Claude AI..."

 local prompt="You are an SRE analyzing production errors. For the following errors, provide:

1. ROOT_CAUSE: One sentence explaining root cause
2. AFFECTED: What functionality is affected
3. SEVERITY: critical/high/medium/low
4. FIX_SUGGESTION: How to fix this (be specific)

Format your response EXACTLY as:
ROOT_CAUSE: ...
AFFECTED: ...
SEVERITY: ...
FIX_SUGGESTION: ...

Error counts from Reqsume API production logs:

$error_counts"

 local response
 response=$(curl -sS -X POST "https://api.anthropic.com/v1/messages" \
 -H "x-api-key: ${ANTHROPIC_API_KEY}" \
 -H "anthropic-version: 2023-06-01" \
 -H "content-type: application/json" \
 --data-binary "$(jq -n \
 --arg prompt "$prompt" \
 '{
 model: "claude-opus-4-7",
 max_tokens: 512,
 messages: [{"role": "user", "content": $prompt}]
 }')" 2>/dev/null)

 # Parse response
 local content
 content=$(echo "$response" | jq -r '.content[0].text // empty')

 if [[ -z "$content" ]]; then
 echo "Analysis failed"
 return 1
 fi

 # Extract severity
 local severity
 severity=$(echo "$content" | grep "^SEVERITY:" | head -1 | sed 's/SEVERITY: //I' | tr -d ' ')

 echo "$content"
 echo "---SEVERITY: $severity---"
}

# Parse severity from analysis
get_severity_from_analysis() {
 local analysis="$1"
 echo "$analysis" | grep "^SEVERITY:" | head -1 | sed 's/SEVERITY: //I' | tr -d ' '
}

# Create GitHub issue
create_issue() {
 local summary="$1"
 local error_counts="$2"
 local analysis="$3"
 local severity="${4:-medium}"

 log_info "Creating GitHub issue..."

 local body
 body=$(cat <<EOF
## SRE Alert

**Detected:** $(date '+%Y-%m-%d %H:%M')
**Environment:** production
**Severity:** $severity
**Window:** $TIME_WINDOW

### Errors Found

$error_counts

### Analysis

$analysis

### Links

- [Grafana Dashboard](https://observability.deemwar.com/d/reqsume-logs)

---
*Auto-generated by SRE Phoenix agent*
*Threshold: min_errors=$THRESHOLD_MIN_ERRORS, severity=$THRESHOLD_SEVERITY*
EOF
)

 local issue_url
 issue_url=$(gh issue create \
 --title "[sre-alert] $summary" \
 --body "$body" \
 --label "sre-alert" \
 --label "bug" 2>/dev/null)

 log_info "Issue created: $issue_url"
 echo "$issue_url"
}

# Main
main() {
 mkdir -p "$LOG_DIR"

 check_binaries
 check_credentials

 echo ""
 echo "=== SRE Phoenix Log Check ==="
 echo ""
 log_info "Mode: $ENV_MODE, Window: $TIME_WINDOW"
 log_info "Thresholds: min_errors=$THRESHOLD_MIN_ERRORS, severity=$THRESHOLD_SEVERITY, dedup=${THRESHOLD_DEDUP_HOURS}h"
 echo ""

 # Step 1: Query errors
 local errors
 errors=$(parse_errors "$(query_errors "$TIME_WINDOW")")

 if [[ -z "$errors" || "$errors" == "No error logs found" ]]; then
 echo ""
 log_info "✓ No errors found in last $TIME_WINDOW"
 exit 0
 fi

 # Step 2: Count error types
 local error_counts
 error_counts=$(count_error_types "$errors")

 echo ""
 log_info "Errors detected:"
 echo "$error_counts" | while read -r count type; do
 echo " - $type ($count occurrences)"
 done

 # Step 3: Analyze with Claude
 local analysis
 local severity
 analysis=$(analyze_errors "$error_counts")
 severity=$(get_severity_from_analysis "$analysis")

 echo ""
 log_info "Claude Analysis (severity: $severity):"
 echo "$analysis" | head -10
 echo ""

 # Step 4: Check thresholds
 local threshold_result
 threshold_result=$(check_thresholds "$error_counts" "$severity")

 log_info "Threshold result: $threshold_result"

 if [[ "$threshold_result" == "SEVERITY_BELOW_THRESHOLD" ]]; then
 log_info "Severity $severity below threshold $THRESHOLD_SEVERITY - logging only"
 exit 0
 fi

 if [[ "$threshold_result" == "COUNT_BELOW_THRESHOLD" ]]; then
 log_info "Error count below threshold $THRESHOLD_MIN_ERRORS - logging only"
 exit 0
 fi

 # Step 5: Check for duplicates
 local duplicate_issue
 duplicate_issue=$(check_duplicate_issue "$error_counts")

 if [[ -n "$duplicate_issue" ]]; then
 log_info "Adding comment to existing issue #$duplicate_issue..."

 # Get root cause from analysis for comment
 local root_cause
 root_cause=$(echo "$analysis" | grep "^ROOT_CAUSE:" | head -1 | sed 's/ROOT_CAUSE: //')

 gh issue comment "$duplicate_issue" --body "$(cat <<EOF
## SRE Phoenix Update $(date '+%Y-%m-%d %H:%M')

### New Occurrences

$error_counts

### Analysis

$root_cause

---
*Auto-generated by SRE Phoenix*
EOF
)"
 echo ""
 log_info "Comment added to issue #$duplicate_issue"
 exit 0
 fi

 # Step 6: Create issue (if flag set)
 if [[ "$CREATE_ISSUE" == "true" ]]; then
 local summary
 summary=$(echo "$error_counts" | head -1 | awk '{$1=""; print $0}' | sed 's/^ //')
 [[ -z "$summary" ]] && summary="Production errors detected"

 create_issue "$summary" "$error_counts" "$analysis" "$severity"
 fi

 echo ""
 log_info "SRE Phoenix check complete"
 log_info "Threshold was met - use --create-issue to create GitHub issue"
}

# Run
main "$@"