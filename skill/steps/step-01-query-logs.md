# Step 01: Query Loki for Reqsume Production Errors

Queries Loki for errors from the **Reqsume Production Logs Dashboard**.

**Dashboard:** https://observability.deemwar.com/d/reqsume-logs

## Target Containers

| Container | Description |
|-----------|-------------|
| `reqsume-app-web-*` | Reqsume UI/API web containers |
| `reqsume-api-*` | API containers (if present) |

## Error Patterns to Detect

| Pattern | Severity | Description |
|---------|----------|-------------|
| `ERROR\|FATAL\|PANIC` | Critical | Application crashes |
| `500\|502\|503\|504` | Warning | HTTP errors |
| `validation_failed` | Warning | User-facing validation errors |
| `extraction_failed` | Warning | AI/ML processing failures |
| `timeout\|TIMEOUT` | Warning | Operations exceeding time limits |
| `connection refused` | Warning | Network/service issues |

## Query Methods

### Method 1: Grafana API (Recommended)

Uses the Grafana API token to query Loki through the proxy.

```bash
TOKEN=$(cat infra/observability/Archive/keys/grafana-api-token)
NOW=$(date +%s)000
SINCE=$(( NOW - 3600000 ))  # 1 hour ago

curl -sS -H "Authorization: Bearer ${TOKEN}" \
  "https://observability.deemwar.com/api/ds/query?ds_type=loki" \
  -H "Content-Type: application/json" \
  --data-binary '{
    "queries": [{
      "expr": "{container=~\"reqsume-app-web.*\"} |= \"(?i)error|fatal|panic|failed\"",
      "refId": "A",
      "maxLines": 100
    }],
    "from": "'${SINCE}'",
    "to": "'${NOW}'"
  }' | jq '.results.A.frames[0].data.values[2] | map(select(. != null)) | .[0:10]'
```

### Method 2: Direct SSH + Loki Query

Queries Loki directly on the observability server.

```bash
# SSH to observability server and query Loki directly
ssh -i infra/vault/production/keys/reqsume_prod \
  -o StrictHostKeyChecking=no \
  root@173.249.45.124 \
  'curl -s "http://localhost:3100/loki/api/v1/query_range?query=%7Bcontainer%3D~%22reqsume-app-web.%2A%22%7D&start='$(date -d "1 hour ago" +%s)'&end='$(date +%s)'&limit=100"'
```

## Time Windows

| Window | Flag | Use Case |
|--------|------|----------|
| 1 hour | `--since 1h` | Standard check (default) |
| 6 hours | `--since 6h` | Investigation |
| 24 hours | `--since 24h` | Overnight errors |
| 7 days | `--since 7d` | Trend analysis |

Accept from user or use default of "1 hour".

## Parse and Deduplicate Errors

```bash
# Parse JSON response, extract error lines
jq -r '.results.A.frames[0].data.values[2][]' | \
  grep -v '^$' | \
  sort | \
  uniq -c | \
  sort -rn | \
  head -20
```

## Output Format

```
Found N errors in last 1 hour:
• validation_failed (3 occurrences)
• Gemini API timeout (1 occurrence)
• Database connection refused (1 occurrence)

Full JSON saved to /tmp/sre-phoenix/errors-20260524-100000.json
```

## Save Results for Analysis

```bash
mkdir -p /tmp/sre-phoenix
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
FILENAME="/tmp/sre-phoenix/errors-${TIMESTAMP}.json"
echo "Saved to ${FILENAME}"
```

## Next Step

If errors found → proceed to `step-02-analyze.md` for AI root cause analysis.

If no errors → print "No errors found in last hour" and stop.
