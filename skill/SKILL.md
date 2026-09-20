---
name: reqsume-ops-phoenix
description: Self-sufficient Ops agent for Reqsume. Monitors production, detects errors, analyzes with AI, creates issues/PRs, and deploys. Run "ops-phoenix" to start. Refuses to run outside a /reqsume-suffixed git checkout.
---

# Ops Phoenix - Self-Sufficient Ops Agent

## What It Does

Ops Phoenix is a comprehensive ops agent that:
1. **Monitors** - Checks production logs for errors
2. **Analyzes** - Uses AI to find root causes
3. **Fixes** - Generates and applies code fixes
4. **Deploys** - Triggers production deployments
5. **Monitors** - Waits for deployment to complete
6. **Auto-Fixes** - If deploy fails, attempts to fix and retry
7. **Reports** - Creates GitHub issues and PRs

## Full Workflow

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

## Quick Start

```bash
# Invoke the agent - it will ask for required info
ops-phoenix

# Or with specific action
ops-phoenix --action detect
ops-phoenix --action full-cycle
```

## First Time Setup

The agent will ask for:
1. **Claude API Key** - For AI analysis
2. **Grafana Token Path** - Location of Grafana API token
3. **GitHub Auth** - Verify gh CLI is authenticated

## Usage Modes

| Mode | Description | Command Flag |
|------|-------------|--------------|
| `detect` | Check logs for errors only | `--action detect` |
| `analyze` | Detect + AI analysis | `--action analyze` |
| `issue` | Detect + analyze + create issue | `--action issue` |
| `fix` | Detect + analyze + create PR | `--action fix` |
| `deploy` | Detect + analyze + fix + deploy | `--action deploy` |
| `full` | Complete cycle (all steps) | `--action full` |
| `setup` | First-time configuration | `--action setup` |
| `status` | Show current configuration | `--action status` |

## Interactive Mode

When invoked without flags, the agent will ask:

```
=== Ops Phoenix Setup ===

Welcome to Ops Phoenix! This agent will help you monitor
and fix issues in Reqsume production.

First, I need some information:

1. What action do you want to perform?
   - detect (check logs only)
   - full (complete cycle: detect → analyze → fix → deploy)

2. Time window to check?
   - 1h (last hour)
   - 6h (last 6 hours)
   - 24h (last day)
   - 7d (last week)

3. Environment?
   - production (default)
   - dev

4. Claude API Key:
   - Leave empty to use ANTHROPIC_API_KEY env var
   - Or enter your key directly

5. Grafana Token Path:
   - Leave empty for default: ~/Downloads/Archive/keys/grafana-api-token
   - Or enter custom path
```

## Configuration Files

The agent creates config files for persistent settings:

```
~/.ops-phoenix/
├── config.json          # Main configuration
├── secrets.env          # Encrypted secrets (optional)
└── history.log          # Run history
```

## Dashboard Monitored

**URL:** https://observability.deemwar.com/d/reqsume-logs

**Target Containers:**
- `reqsume-app-web-*` - Reqsume UI/API containers
- `reqsume-api-*` - API containers

**NOT monitored:** External services (video-ai-worker, etc.)

## Environment Variables Used

| Variable | Required | Description |
|----------|----------|-------------|
| `ANTHROPIC_API_KEY` | Yes* | Claude API key (*or enter when prompted) |
| `ANTHROPIC_BASE_URL` | No | Custom API proxy (default: https://api.opusmax.pro) |
| `GRAFANA_TOKEN_PATH` | No | Grafana token location (default: ~/Downloads/Archive/keys/grafana-api-token) |
| `GITHUB_REPO` | No | GitHub repo (default: muthuishere/reqsume) |

## Output Examples

### Detect Mode

```
=== OPS PHOENIX ===
Action: detect
Time: 1h
Environment: production

Querying Loki for errors...
Found: 5 errors

Errors:
  - api: validation_failed (3x)
  - api: timeout (2x)

Use --action issue to create GitHub issue
```

### Full Cycle

```
=== OPS PHOENIX ===
Action: full
Time: 1h
Environment: production

[1/6] DETECT: Found 5 errors
[2/6] ANALYZE: Root cause identified
[3/6] ISSUE: #412 created
[4/6] FIX: Code generated
[5/6] PR: #413 created
[6/6] DEPLOY: Triggered

Done! Check your GitHub for issue #412
```

## Error Handling

The agent will:
1. **Ask for missing info** - If config incomplete, prompt for values
2. **Validate inputs** - Check API keys, token paths, etc.
3. **Report failures** - Clear error messages with suggestions
4. **Continue on partial failure** - Don't stop if one step fails

## VM Setup

To run on a VM:

```bash
# 1. Clone the repo
git clone git@github.com:muthuishere/reqsume.git
cd reqsume

# 2. Run setup
python3 infra/skills/reqsume-ops-phoenix/scripts/ops-phoenix.py --action setup

# 3. Add to crontab
crontab -e
# Add: */10 * * * * /path/to/ops-phoenix.py --action detect
```

## Security

- **Never hardcode secrets** - Always use env vars or secure input
- **Validate before use** - Check API keys work before running
- **Audit trail** - All actions logged to history.log
- **Rate limiting** - Max 1 fix per hour per error type

## Files

```
reqsume-ops-phoenix/
├── SKILL.md                        # This file
├── README.md                       # Quick start
├── scripts/
│   ├── ops-phoenix.py              # Main agent (Python)
│   ├── ops-phoenix-interactive.sh  # Interactive wrapper
│   └── ops-phoenix-setup.sh        # Setup script
├── workflows/
│   ├── full-cycle.md               # Complete flow
│   └── test-scenario.md            # Testing guide
└── references/
    └── error-patterns.md           # Known patterns
```

## Context

See `framework/docs/MEMORY.md` for memory/state design and `~/.ops-phoenix/memory.md` for the latest runtime state snapshot.