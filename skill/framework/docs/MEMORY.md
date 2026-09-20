# Ops Phoenix - Memory & Issue Tracker

> **This file tracks what Ops Phoenix did, what errors it found, and what issues it fixed.**

---

## How It Works

Every time Ops Phoenix runs, it:

1. **Detects errors** from logs
2. **Analyzes root cause** using Claude AI
3. **Creates GitHub issues** for tracking
4. **Generates fixes** using Claude AI
5. **Creates PRs** to apply fixes
6. **Deploys** to production
7. **Records run history** in `~/.ops-phoenix/history.json`
8. **Updates state memory** in `~/.ops-phoenix/memory.md`

---

## Quick Stats

| Metric | Count |
|--------|-------|
| Total runs | Run `python3 ops_phoenix.py --history` to see |
| Issues created | Check history |
| PRs merged | Check history |
| Errors fixed | Check history |

---

## Run History

### Viewing History

```bash
# Show summary
python3 ops_phoenix.py --history

# Export as markdown
python3 ops_phoenix.py --history-export

# View raw history
cat ~/.ops-phoenix/history.json
```

### History File Location

```
~/.ops-phoenix/
├── config.json        # Your configuration
├── history.json       # All run records (auto-created)
└── memory.md          # Current lifecycle + repeated issue states
```

---

## What Gets Recorded

### Per Run

| Field | Description |
|-------|-------------|
| `timestamp` | When the run started |
| `mode` | `dev` or `prod` |
| `time_window` | `1h`, `6h`, `24h`, `7d` |
| `status` | `healthy`, `complete`, `failed` |
| `errors_found` | Number of error logs detected |
| `error_types` | Dict of error type → count |

### Per Issue

| Field | Description |
|-------|-------------|
| `number` | GitHub issue number |
| `title` | Issue title |
| `url` | Link to issue |
| `created_at` | When created |

### Per PR

| Field | Description |
|-------|-------------|
| `number` | GitHub PR number |
| `title` | PR title |
| `url` | Link to PR |
| `merged` | Boolean |
| `created_at` | When created |

### Per Deployment

| Field | Description |
|-------|-------------|
| `run_id` | GitHub Actions run ID |
| `workflow` | Workflow name |
| `status` | `triggered`, `success`, `failed:reason` |
| `triggered_at` | When triggered |

### Per Auto-Fix

| Field | Description |
|-------|-------------|
| `error_type` | What error was fixed |
| `fix_applied` | What Claude suggested |
| `success` | Whether it worked |

---

## Example Run Record

```json
{
  "timestamp": "2026-05-30T10:30:00",
  "mode": "prod",
  "time_window": "1h",
  "status": "complete",
  "errors_found": 15,
  "error_types": {
    "api: connection timeout": 10,
    "db: query failed": 5
  },
  "issues_created": [
    {
      "number": 123,
      "title": "[ops-alert] api: connection timeout (10 occurrences)",
      "url": "https://github.com/org/repo/issues/123"
    }
  ],
  "prs_created": [
    {
      "number": 124,
      "title": "fix: resolve connection timeout issues",
      "url": "https://github.com/org/repo/pull/124",
      "merged": true
    }
  ],
  "deploys_triggered": [
    {
      "run_id": "123456789",
      "workflow": "Deploy to Production",
      "status": "success"
    }
  ],
  "auto_fixes": [],
  "manual_issues": []
}
```

---

## Memory File vs History File

| Aspect | `~/.ops-phoenix/memory.md` | `~/.ops-phoenix/history.json` |
|--------|-----------------------------|------------------------------|
| **Location** | Home (`~/.ops-phoenix/`) | Home (`~/.ops-phoenix/`) |
| **Purpose** | Current agent state snapshot | Permanent run records |
| **Contents** | Lifecycle state + repeated issue status | Every run, every issue, every PR |
| **Format** | Markdown (machine-generated) | JSON (machine-generated) |
| **Who reads** | Agent + humans | Ops Phoenix CLI |

---

## What to Ask Claude

When you want to know:

- **"What did Ops Phoenix do?"** → Run `python3 ops_phoenix.py --history`
- **"Show me all fixes"** → Run `python3 ops_phoenix.py --history-export`
- **"What errors are recurring?"** → Check `error_types` in history
- **"What is the current state right now?"** → Open `~/.ops-phoenix/memory.md`
- **"Did the last deploy work?"** → Run `gh run view` or check history

---

## Example Workflows

### Check Last Run

```bash
python3 ops_phoenix.py --history | head -50
```

### Find Recurring Errors

```bash
cat ~/.ops-phoenix/history.json | jq '[.runs[].error_types | to_entries[].key] | group_by(.) | map({key: .[0], count: length}) | sort_by(-.count)'
```

### Find All Fixed Issues

```bash
python3 ops_phoenix.py --history --fixes
```

---

## Retention

- **History file**: Keeps all runs forever (grows over time)
- **Export**: Can export to markdown for reports
- **Cleanup**: Not implemented yet (future feature)

---

## State Rules (for repeated issues)

- First time an error signature appears: create issue + PR and run deploy flow.
- When the same signature appears again: update `memory.md` state row (runs seen, last seen, state) instead of duplicating tracking records.
- Deploy failures keep state as `failed` until a later successful run moves it to `deployed`/`healthy`.

## Future Features

- [ ] Auto-cleanup of old runs (keep last 90 days)
- [ ] Export to external system (Slack, PagerDuty)
- [ ] Performance trends (how long to fix, success rate)