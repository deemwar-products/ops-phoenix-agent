# Step 02: Analyze

Send detected errors to AI for root-cause analysis.

## Prerequisites

- Errors must be loaded from the previous step (step-01-query-logs)
- `ai.api_key_env` must be set in environment
- At least one error pattern must exist

## Build Analysis Prompt

Send the error summary to the AI model configured in `config.json`:

```
You are an SRE analyzing production errors.

## Error Summary
{top_10_errors_with_counts}

Provide:
1. ROOT_CAUSE: One sentence explaining the root cause
2. AFFECTED: What functionality is affected
3. SEVERITY: critical/high/medium/low
4. FIX_SUGGESTION: How to fix this
5. FILES_AFFECTED: Which files in the repo likely need changes
6. CONFIDENCE: 0.0–1.0 score

Format:
ROOT_CAUSE: ...
AFFECTED: ...
SEVERITY: ...
FIX_SUGGESTION: ...
FILES_AFFECTED: ...
CONFIDENCE: ...
```

## Parse Response

Extract the structured fields from the AI response. If the response doesn't match the expected format, retry once with a stricter prompt. If still failing, return `confidence: 0.0` and skip the fix step.

## Threshold Check

Read `{config-dir}/thresholds.json` if it exists. Compare the finding's severity and confidence against configured thresholds:

| Severity | Min confidence for action |
|----------|--------------------------|
| critical | 0.5 |
| high | 0.7 |
| medium | 0.85 |
| low | 0.95 |

If confidence is below threshold, report the finding but do not create a PR. Route to issue-only mode.

## Output Format

```
=== ANALYSIS ===
Root cause: <root_cause>
Severity: <severity>
Affected: <affected>
Fix: <fix_suggestion>
Files: <files_affected>
Confidence: <score> (threshold: <threshold>)

Decision: <proceed_to_fix | issue_only>
```

## Next Step

If `proceed_to_fix` → proceed to `step-03-report.md`.
If `issue_only` → create GitHub issue only, no PR.
