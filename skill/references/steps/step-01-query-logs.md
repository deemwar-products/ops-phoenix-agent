# Step 01: Query Logs

Query the observability backend for errors within the configured time window.

## Build Query

Read container and host patterns from `config.json`:
- `observability.container_patterns` — comma-separated regex for container names
- `observability.host_patterns` — comma-separated regex for hosts
- `observability.error_patterns` — Loki logQL or equivalent error filter

Build the query expression from these patterns. Do not hardcode any container names, hostnames, or URLs.

## Execute Query

### Grafana Self-Hosted

```bash
TOKEN=$(cat {GRAFANA_TOKEN_PATH})
curl -sS -H "Authorization: Bearer ${TOKEN}" \
  "{GRAFANA_URL}/api/ds/query?ds_type=loki" \
  -H "Content-Type: application/json" \
  --data-binary '{
    "queries": [{
      "expr": "<built-from-config>",
      "refId": "A",
      "datasource": {"type": "loki", "uid": "loki"},
      "maxLines": 100
    }],
    "from": "<from-ts>",
    "to": "<to-ts>"
  }'
```

### Loki Direct

```bash
curl -sS -G "{LOKI_URL}/loki/api/v1/query_range" \
  --data-urlencode "query=<built-from-config>" \
  --data "limit=100&start=<from-ts>&end=<to-ts>"
```

Add `-u user:pass` if `observability.loki_user` / `loki_password` are set.

## Time Windows

| Flag | Value | Use |
|------|-------|-----|
| `5m` | 300s | Recent check |
| `1h` | 3600s | Standard (default) |
| `6h` | 21600s | Investigation |
| `24h` | 86400s | Overnight |
| `7d` | 604800s | Weekly trend |

## Parse Response

Extract error lines from the response. Deduplicate by normalizing each line to its first 80 characters. Sort by frequency descending. Return the top 20 unique error patterns with occurrence counts.

## Output Format

```
Detected N unique errors in last {duration}:
• <error_pattern_1> (X occurrences)
• <error_pattern_2> (Y occurrences)
...

Top finding: <most_frequent_error> (Z occurrences)
```

Save raw results to `{config-dir}/errors-{timestamp}.json` for the analysis step.

## Next Step

If errors found → proceed to `step-02-analyze.md`.
If no errors → print "No errors found — all systems healthy" and stop.
