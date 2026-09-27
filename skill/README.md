# SRE Agent — Agent Skills

Agent skills for SRE Agent: production monitoring, AI-powered root-cause analysis, and autonomous fix generation.

## Contents

| Path | Purpose |
|------|---------|
| `SKILL.md` | Skill definition (triggers, variables, config, commands) |
| `references/steps/` | Step-by-step instructions for Claude |
| `workflow.md` | Scheduled/cron setup guide |
| `references/error-patterns.md` | Common error patterns and severity mappings |

## Usage

Install the skill in Claude Code:

```bash
npx skills add sre-agent
```

Or place the `skill/` directory in your agent's skills folder.

## Trigger Phrases

Claude activates this skill when you say things like:
- "detect incidents" / "check production for errors"
- "analyze the error logs"
- "fix the outage"
- "open an issue" / "create a PR"
- "run the SRE agent"
- "autonomous incident response"

## Prerequisites

1. SRE Agent installed (`sre-agent` binary or `ops_phoenix_main.py`)
2. Config initialized (`sre-agent init`)
3. Grafana token, GitHub auth, and AI API key configured

## How It Works

```
User: "Check production for errors and fix anything critical"
  → Claude activates sre-agent skill
  → step-00: preflight (validate config + connections)
  → step-01: query logs (detect errors)
  → step-02: analyze (AI root cause)
  → step-03: report (issue + PR + deploy)
  → Claude reports results to user
```

## Modes

- **guided** (default): detects, analyzes, opens PR — waits for human approval before deploy
- **autonomous**: detects, analyzes, fixes, deploys, monitors — no human intervention

Set in `{config-dir}/config.json` under `deployment.mode` or pass `--mode` flag.

## State

All state lives in `{config-dir}/`:
- `config.json` — configuration (non-secret)
- `history.json` — run records (issues, PRs, deploys)
- `memory.md` — recent run summary
- `errors-*.json` — raw error data per run
- `work/repo/` — cloned repo for fix generation
