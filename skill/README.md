# Ops Phoenix Agent

Automated Ops agent that monitors **Reqsume Production Logs Dashboard** and runs the complete self-healing cycle.

**Dashboard:** https://observability.deemwar.com/d/reqsume-logs

## Quick Start

```bash
# Set Claude API key
export ANTHROPIC_API_KEY="sk-ant-..."

# Dry run (detect only)
python3 scripts/check-errors.py --dry-run

# Detect + create issue
python3 scripts/check-errors.py --create-issue

# Full cycle (detect + issue + fix + PR + deploy)
python3 scripts/check-errors.py --full-cycle
```

## What It Does

```
┌─────────────────────────────────────────────────────────────────────┐
│                      OPS PHOENIX FULL FLOW                          │
└─────────────────────────────────────────────────────────────────────┘

  ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
  │  1. DETECT   │ ──▶ │  2. ANALYZE  │ ──▶ │  3. ISSUE   │
  │  Query Loki  │     │  Claude AI   │     │  GitHub     │
  │  Find errors │     │  Root cause  │     │  Create     │
  └──────────────┘     └──────────────┘     └──────────────┘
                                                    │
         ┌──────────────────────────────────────────┘
         │
         ▼
  ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
  │  4. FIX     │ ──▶ │  5. PR       │ ──▶ │  6. DEPLOY  │
  │  Claude AI   │     │  GitHub CLI   │     │  GitHub      │
  │  Generate    │     │  Branch+PR    │     │  Actions    │
  │  Code        │     └──────────────┘     └──────────────┘
  └──────────────┘                              │
                                                  ▼
                                   ┌─────────────────────────┐
                                   │  7. MONITOR            │
                                   │  Wait for completion   │
                                   └─────────────────────────┘
                                                  │
                              ┌────────────────────┴────────────────────┐
                              │                                         │
                              ▼                                         ▼
                   ┌──────────────────┐                   ┌──────────────────┐
                   │   SUCCESS        │                   │    FAILURE       │
                   │   ✓ Done         │                   │   8. AUTO-FIX    │
                   └──────────────────┘                   │   Analyze logs   │
                                                           │   Apply fix      │
                                                           │   Redeploy       │
                                                           └──────────────────┘
```

## Auto-Fix on Deploy Failure

If the deployment workflow fails, Ops Phoenix will:
1. **Analyze failure** - Detect if it's a test, build, or deploy failure
2. **Attempt fix** - Generate code fix if possible
3. **Retry deploy** - Trigger new deployment
4. **Create issue** - If auto-fix fails, create issue for manual review

## Prerequisites

1. **Claude API Key** — `ANTHROPIC_API_KEY` environment variable
2. **Grafana Token** — at `~/Downloads/Archive/keys/grafana-api-token`
3. **GitHub CLI** — authenticated with `gh auth login`

## Target Containers

| Container | Monitored |
|-----------|-----------|
| `reqsume-app-web-*` | ✅ Yes |
| `reqsume-api-*` | ✅ Yes |
| `video-ai-worker-*` | ❌ No (external) |
| `kamal-proxy` | ❌ No |

## Files

```
reqsume-ops-phoenix/
├── SKILL.md                      # Full skill documentation
├── README.md                     # This file
├── scripts/
│   └── check-errors.py          # Main agent script
├── workflows/
│   ├── full-cycle.md             # Complete flow
│   └── test-scenario.md          # Testing guide
└── references/
    └── error-patterns.md         # Known patterns
```

## Usage

```bash
# Dry run - detect errors only
python3 scripts/check-errors.py --dry-run

# Detect + create GitHub issue
python3 scripts/check-errors.py --create-issue

# Full cycle (detect + issue + fix + PR + deploy)
python3 scripts/check-errors.py --full-cycle

# Custom time window
python3 scripts/check-errors.py --window 7d --create-issue
```

## Cron Setup (for VM)

```bash
# Edit crontab
crontab -e

# Run every hour at minute 0
0 * * * * /opt/ops-phoenix/check-errors.py --create-issue >> /var/log/ops-phoenix.log 2>&1
```

## Secrets for VM

```bash
# Create directory
sudo mkdir -p /opt/ops-phoenix/secrets
sudo chmod 700 /opt/ops-phoenix/secrets

# Store Claude API key
echo "sk-ant-..." | sudo tee /opt/ops-phoenix/secrets/claude-api-key
sudo chmod 600 /opt/ops-phoenix/secrets/claude-api-key
```

## Example Output

```
[2026-05-24 10:00:00] Ops Phoenix started
Mode: dry-run

=== DETECT ===
Querying Loki for Reqsume production errors...
✓ No errors found in reqsume-app containers (last 1h)
```

## Context

See `framework/docs/MEMORY.md` for memory/state design and `~/.ops-phoenix/memory.md` for the latest runtime state snapshot.
