# Test Scenario: SRE Phoenix Full Cycle

This document describes how to test the SRE Phoenix agent by introducing a deliberate error in the Reqsume app.

## Test Overview

| Step | Action | Verification |
|------|--------|--------------|
| 1 | Add test error endpoint | Error appears in logs |
| 2 | Trigger the error | Error captured by agent |
| 3 | Run agent | Issue + PR created |
| 4 | Merge PR | Error endpoint removed |

## Step 1: Add Test Error Endpoint

Create a test endpoint that logs errors:

```go
// apps/api/src/handlers/test_sre.go
package handlers

import (
    "net/http"
    "fmt"
    "github.com/gin-gonic/gin"
)

// SRETestErrorEndpoint - Test endpoint for SRE Phoenix agent
// This endpoint deliberately logs errors for testing purposes
// REMOVE AFTER TESTING
func SRETestErrorEndpoint(c *gin.Context) {
    // Log an error with structured format (for Loki parsing)
    fmt.Printf(`{"ts": "%s", "level": "ERROR", "logger": "api", "msg": "sre_test_error", "endpoint": "/api/test/sre", "user_id": "%s"}`,
        time.Now().UTC().Format(time.RFC3339),
        c.Query("user_id"))

    // Also log a validation error
    fmt.Printf(`{"ts": "%s", "level": "ERROR", "logger": "api", "msg": "validation_failed", "field": "email", "error": "invalid_format"}`,
        time.Now().UTC().Format(time.RFC3339))

    c.JSON(http.StatusInternalServerError, gin.H{
        "error": "SRE test error - please ignore",
        "message": "This is a test error for SRE Phoenix agent",
    })
}
```

Add route in `handlers/routes.go`:

```go
// SRE TEST ROUTES - REMOVE AFTER TESTING
testGroup := v1.Group("/test")
{
    testGroup.GET("/sre", SRETestErrorEndpoint)
}
```

## Step 2: Deploy to Production

```bash
# Commit the test changes
git checkout -b test/sre-phoenix-validation
git add apps/api/src/handlers/test_sre.go
git commit -m "test(sre): add test endpoint for SRE Phoenix agent"
git push origin test/sre-phoenix-validation

# Create PR
gh pr create --title "test(sre): SRE Phoenix agent test" --body "Testing the SRE Phoenix agent flow"
gh pr merge --squash

# Deploy
task production:deploy:api
```

## Step 3: Trigger the Error

```bash
# Call the test endpoint multiple times to generate errors
curl https://api.reqsume.com/api/test/sre?user_id=test123
curl https://api.reqsume.com/api/test/sre?user_id=test456
curl https://api.reqsume.com/api/test/sre?user_id=test789
```

Wait 2 minutes for logs to appear in Loki.

## Step 4: Verify Error in Logs

```bash
# Check if errors appear in Loki
export ANTHROPIC_API_KEY="sk-ant-..."
python3 infra/skills/reqsume-sre-phoenix/scripts/check-errors.py --window 1h --dry-run
```

You should see:
```
Errors detected: 3
Error breakdown:
  - api: sre_test_error: 3 occurrences
```

## Step 5: Run Full Cycle

```bash
# Run the full SRE Phoenix cycle
python3 infra/skills/reqsume-sre-phoenix/scripts/check-errors.py --full-cycle
```

Expected output:
```
=== DETECT ===
Errors found: 3

=== ANALYZE ===
Claude analysis complete.

=== ISSUE ===
Issue created: #XXX

=== FIX ===
Branch created: fix/sre-test-error

=== DEPLOY ===
PR created: #YYY
Deploy triggered: production
```

## Step 6: Verify GitHub Issue

Check the created issue at: https://github.com/muthuishere/reqsume/issues

Expected:
- Title: `[sre-alert] api: sre_test_error (3 occurrences)`
- Labels: `sre-alert`, `bug`
- Body contains error details and Claude analysis

## Step 7: Verify PR

Check the created PR at: https://github.com/muthuishere/reqsume/pulls

Expected:
- Title: `fix(sre): Remove test error endpoint`
- Removes the test endpoint code
- Links to the issue

## Step 8: Verify Deployment

```bash
# Check if fix is deployed
curl https://api.reqsume.com/api/test/sre
# Should return 404 after fix deployed
```

## Cleanup

After successful test:

1. **Merge the PR** if not auto-merged
2. **Verify logs show no errors** after cleanup:
   ```bash
   python3 infra/skills/reqsume-sre-phoenix/scripts/check-errors.py --window 1h --dry-run
   # Should show: No errors found
   ```

## Expected Results

| Checkpoint | Expected Result |
|------------|----------------|
| Error detection | 3 errors found |
| Issue created | GitHub issue with `sre-alert` label |
| PR created | PR removing test endpoint |
| Fix deployed | Error endpoint removed |
| Post-fix check | No errors in logs |

## Troubleshooting

### Error not appearing in logs

- Wait 2-5 minutes for Loki to receive logs
- Check Grafana dashboard directly: https://observability.deemwar.com/d/reqsume-logs
- Verify the endpoint is returning 500 status

### Agent not detecting errors

- Check the Loki query in the script matches the container names
- Verify Grafana token is valid
- Check if errors use JSON format with `level: ERROR`

### PR not created

- Ensure `gh auth` is logged in
- Check if branch already exists from previous run
- Verify GitHub has write access

## Simpler Test (No Code Changes)

Instead of adding code, you can trigger existing errors:

1. **Trigger validation errors** by submitting invalid forms
2. **Wait for timeout errors** by making slow requests
3. **Trigger 500 errors** by using known broken endpoints

This tests the detection without needing to deploy code.
