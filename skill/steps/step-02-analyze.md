# Step 02: Analyze Errors with Claude AI

Uses Claude to analyze errors, determine root cause, and suggest fixes.

## Prerequisites

- `ANTHROPIC_API_KEY` environment variable set
- Error JSON file from `step-01-query-logs.md`
- `jq` for parsing

## Claude Analysis Prompt

Construct a prompt that includes:

1. Error patterns found
2. Frequency of each error type
3. Request context (timestamps, user IDs)
4. Ask for root cause and fix

## Example Prompt

```
You are an SRE analyzing production errors. Given these log entries, identify:
1. Root cause (one sentence)
2. Affected users/requests
3. Suggested fix (code/config change)
4. Severity (critical/warning/info)

Errors from reqsume API (2026-05-24):

## validation_failed (3 occurrences)
[log line 1]
[log line 2]
[log line 3]

## Gemini 503 timeout (1 occurrence)
[log line]
```

## Claude API Call

```bash
curl -sS -X POST "https://api.anthropic.com/v1/messages" \
  -H "x-api-key: ${ANTHROPIC_API_KEY}" \
  -H "anthropic-version: 2023-06-01" \
  -H "content-type: application/json" \
  -d '{
    "model": "claude-opus-4-7",
    "max_tokens": 1024,
    "messages": [{
      "role": "user",
      "content": "Analyze these production errors..."
    }]
  }' | jq '.content'
```

## Parse Claude Response

Extract:
- Root cause summary
- Fix suggestion (code snippet if applicable)
- Severity assessment

## Save Analysis

```bash
mkdir -p /tmp/sre-phoenix/analyses
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
echo "ANALYSIS_TIMESTAMP=${TIMESTAMP}" > /tmp/sre-phoenix/analyses/latest.env
```

## Output Format

```
Analysis complete (took Xs)

Root Cause: [one line summary]
Severity: [critical/warning/info]
Suggested Fix: [brief description]

Full analysis saved to /tmp/sre-phoenix/analyses/${TIMESTAMP}.md
```

## Guardrails

- Never suggest database migrations without review
- Never suggest destructive operations
- Always recommend testing in dev first
- Flag as "requires manual review" for complex changes

## Next Step

If analysis complete → proceed to `step-03-report.md` for GitHub issue creation.
