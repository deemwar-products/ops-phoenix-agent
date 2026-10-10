# error-classification

How sre-agent classifies errors and what each severity means.

## Severity Levels

| Level | Label | Criteria | Action |
|---|---|---|---|
| P0 | Critical | Service down, data loss, security breach | Immediate alert, page on-call |
| P1 | Urgent | Feature broken for users, high error rate | Fix within hours |
| P2 | Warning | Degraded experience, workaround available | Fix within days |
| P3 | Low | Noise, probes, minor issues, no user impact | Backlog |

## Common Patterns

### Gemini API 402 RESOURCE_EXHAUSTED
- **Severity**: P0 if extraction is user-facing (resume uploads failing), P1 otherwise
- **Root cause**: Google AI Studio prepayment credits depleted
- **Fix**: Top up credits, add fallback provider, configure billing alerts

### validation_failed
- **Severity**: P2 (client-side issue, not systemic)
- **Root cause**: Client submitting invalid payload, or server-side validation
  rules too strict / poorly logged
- **Fix**: Add detailed validation error logging, return 422 with field-level
  errors, implement input sanitization

### nginx 404 on .php paths
- **Severity**: P3 (automated scanner probing)
- **Root cause**: External bot scanning for known PHP vulnerabilities
- **Fix**: Add nginx deny rule for `.php$`, consider fail2ban for repeat scanners

### Circuit breaker open
- **Severity**: P0 if blocking user requests, P1 otherwise
- **Root cause**: Repeated downstream failures (often 402/429/5xx)
- **Fix**: Fix upstream issue first, then tune breaker thresholds. Add
  fallback provider for non-retryable errors (402 = billing, don't trip breaker)

### DB connection errors
- **Severity**: P0
- **Root cause**: DB down, connection pool exhausted, wrong credentials
- **Fix**: Check DB health, connection limits, credentials

## Confidence Scores

- **0.90–1.00**: High confidence — root cause clear from logs
- **0.70–0.89**: Medium confidence — likely cause, verify manually
- **< 0.70**: Low confidence — AI is guessing, surface as "unconfirmed"

## Duplicate Detection

sre-agent deduplicates errors by container + error pattern within the time
window. When multiple log lines share the same root cause, they collapse to
a single finding. The `(xN)` count in `detect` output shows how many times
the same error occurred.
