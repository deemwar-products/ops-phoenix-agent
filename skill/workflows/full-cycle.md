# SRE Phoenix Full Cycle Workflow

Complete automated flow from error detection to production deployment.

## Flow Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           SRE PHOENIX FULL CYCLE                         │
└─────────────────────────────────────────────────────────────────────────┘

  ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
  │   1. DETECT  │ ──▶ │  2. ANALYZE  │ ──▶ │ 3. CREATE   │
  │  Query Loki  │     │  Claude AI   │     │   ISSUE     │
  └──────────────┘     └──────────────┘     └──────────────┘
                                                     │
         ┌───────────────────────────────────────────┘
         │
         ▼
  ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
  │  4. FIX      │ ──▶ │  5. CREATE   │ ──▶ │  6. DEPLOY  │
  │  Claude AI   │     │      PR      │     │   Kamal     │
  │  Generate    │     │  GitHub CLI  │     │  Production │
  │  Code Fix    │     └──────────────┘     └──────────────┘
  └──────────────┘
```

## Step-by-Step

### Step 1: DETECT - Query Loki for Errors

```bash
# Query Grafana/Loki for errors in reqsume-app containers
TOKEN=$(cat ~/Downloads/Archive/keys/grafana-api-token)
curl -sS -H "Authorization: Bearer ${TOKEN}" \
  "https://observability.deemwar.com/api/ds/query?ds_type=loki" \
  --data '{
    "queries": [{
      "expr": "{container=~\"reqsume-app-web.*|reqsume-api.*\", host=~\"app1|app2\"} | json | level=~\"(?i)error|fatal\"",
      "maxLines": 100
    }],
    "from": "now-1h",
    "to": "now"
  }'
```

**Output:** JSON with error logs from reqsume-app containers.

---

### Step 2: ANALYZE - Claude AI Root Cause

```python
# Send errors to Claude for analysis
prompt = f"""Analyze these production errors:

{error_logs}

For each error:
1. Root cause
2. Affected functionality
3. Fix suggestion (code or config)
"""

# Call Claude API
response = claude.messages.create(
    model="claude-opus-4-7",
    messages=[{"role": "user", "content": prompt}]
)
```

**Output:**
```
Root cause: Null pointer in profile validation
Affected: /api/user/profile endpoint
Fix: Add null check for user_id before validation
```

---

### Step 3: CREATE ISSUE - GitHub Issue

```bash
# Create GitHub issue with analysis
gh issue create \
  --title "[sre-alert] {error_type}: {summary}" \
  --body "## Error Analysis\n\n{claude_analysis}\n\n## Links\n\n[Grafana](https://observability.deemwar.com/d/reqsume-logs)" \
  --label "sre-alert" \
  --label "bug"
```

**Output:** GitHub issue URL

---

### Step 4: FIX - Generate Code Fix

```python
# Ask Claude to generate the fix
fix_prompt = f"""Based on this error analysis:

{analysis}

Generate the code fix for the Reqsume Go API.
Return only the code changes needed.
"""

fix_response = claude.messages.create(
    model="claude-opus-4-7",
    messages=[{"role": "user", "content": fix_prompt}]
)
```

**Output:** Code changes to apply

---

### Step 5: CREATE PR - GitHub PR

```bash
# Create branch from main
git checkout main
git pull origin main
git checkout -b fix/sre-{error-type}-{timestamp}

# Apply fix
git apply << 'EOF'
// code changes here
EOF

# Commit and push
git add .
git commit -m "fix(sre): {error_type}: {summary}"
git push origin HEAD

# Create PR
gh pr create \
  --title "fix(sre): {error_type}: {summary}" \
  --body "Fixes #$(gh issue list --label sre-alert --state open --json number | jq '.[0].number')" \
  --label "sre-alert"
```

**Output:** PR URL

---

### Step 6: DEPLOY - Production

```bash
# Merge PR
gh pr merge --squash

# Wait for CI/CD
sleep 30

# Deploy to production
task production:deploy:api
```

**Output:** Deployment complete

---

## Command Reference

### Full Cycle (Automated)

```bash
# Run everything automatically
export ANTHROPIC_API_KEY="sk-ant-..."
python3 scripts/check-errors.py --full-cycle
```

### Individual Steps

```bash
# 1. Detect errors only
python3 scripts/check-errors.py --dry-run

# 2. Create issue only
python3 scripts/check-errors.py --create-issue

# 3. Create fix and PR only (needs issue number)
python3 scripts/create-fix.py --issue 410

# 4. Deploy only (needs PR number)
python3 scripts/deploy-fix.py --pr 411
```

### Time Windows

```bash
# Check last hour
python3 scripts/check-errors.py --window 1h

# Check last 24 hours
python3 scripts/check-errors.py --window 24h

# Check last 7 days
python3 scripts/check-errors.py --window 7d
```

---

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `ANTHROPIC_API_KEY` | Yes | - | Claude API key |
| `ANTHROPIC_BASE_URL` | No | `https://api.opusmax.pro` | Custom API proxy |
| `GRAFANA_TOKEN_PATH` | No | `~/Downloads/Archive/keys/grafana-api-token` | Grafana token path |

---

## Verification

After full cycle completes:

```bash
# Verify no errors in logs
python3 scripts/check-errors.py --dry-run
# Should show: No errors found

# Verify PR merged
gh pr list --label sre-alert --state merged

# Verify deployment
curl https://api.reqsume.com/api/health
```

---

## Rollback

If something goes wrong:

```bash
# Revert to previous commit
git revert HEAD
git push origin main

# Or rollback via Kamal
task production:deploy:api ROLLBACK=true
```

---

## Monitoring

Check SRE Phoenix activity:

```bash
# Recent issues
gh issue list --label sre-alert --state all --limit 10

# Recent PRs
gh pr list --label sre-alert --state all --limit 10

# Recent deployments
task production:logs:api | tail -20
```
