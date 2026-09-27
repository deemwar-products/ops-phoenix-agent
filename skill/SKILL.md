---
name: sre-agent
description: >
  Autonomous SRE agent that monitors production logs, detects errors, analyzes root causes with AI,
  and creates fixes as GitHub PRs. Trigger this skill  "detect incidents", "fix the outage",
  "analyze the error logs", "open an issue for this", "deploy the fix", "run the SRE agent",
  "autonomous incident response", "self-healing", "ops agent", "SRE automation",
  "check logs", "what failed", "monitor production", "create a fix", "auto-remediate".
  Also trigger when the user reports a production issue, an alert, or asks for help with an incident.
---

# SRE Agent

Autonomous SRE agent: detects errors in production logs, uses AI to find root causes, and creates GitHub PRs with fixes.

## Interfaces

This skill has two surfaces:

1. **CLI** — run directly in terminal (human-driven)
2. **Agent skill** — invoked by Claude Code, Codex, Cursor (AI-driven)

Both use the same engine. Same results.

## Variables

- `{skill-root}` = the `skill/` directory where this file lives
- `{config-dir}` = `~/.config/sre-agent/`
- `{GRAFANA_TOKEN_PATH}` = path to Grafana API token file
- `{GITHUB_REPO}` = target GitHub repo (owner/name)
- `{GITHUB_TOKEN}` = GitHub token env var name
- `{ANTHROPIC_API_KEY}` = Claude API key env var
- `{ANTHROPIC_BASE_URL}` = Claude API base URL (default: https://api.anthropic.com)
- `{CONTAINER_PATTERNS}` = comma-separated container regex patterns (e.g., `app-*,api-*`)
- `{HOST_PATTERNS}` = comma-separated host patterns (e.g., `*`)

## Configuration

Config lives at `{config-dir}/config.json`. Run `sre-agent init` to create it interactively, or write it directly:

```json
{
  "observability": {
    "type": "grafana_self_hosted",
    "grafana_url": "https://observability.example.com",
    "grafana_token_path": "/etc/sre-agent/grafana-token",
    "container_patterns": ["app-*", "api-*"],
    "host_patterns": ["*"],
    "error_patterns": ["level=~\"(?i)error|fatal|panic\""]
  },
  "github": {
    "repo": "owner/repo",
    "token_env_var": "GITHUB_TOKEN",
    "base_branch": "main",
    "labels": ["sre-alert"]
  },
  "ai": {
    "provider": "anthropic",
    "model": "claude-sonnet-4-20250514",
    "api_key_env": "ANTHROPIC_API_KEY",
    "api_url": "https://api.anthropic.com"
  },
  "cicd": {
    "type": "github_actions",
    "workflow_name": "Deploy to Production"
  },
  "deployment": {
    "auto_merge": false,
    "monitor_timeout": 300
  }
}
```

All values are overridable by environment variables. Secrets (API keys, tokens) should NOT be in config — use env vars or a secrets file with mode 0600.

## CLI Commands

Run from the installed binary or `python3 {skill-root}/framework/ops_phoenix_main.py`:

```bash
sre-agent init                    # Interactive setup wizard
sre-agent detect --duration 5m    # Query logs for errors
sre-agent analyze --finding 0     # AI root-cause analysis on finding #0
sre-agent fix --finding 0         # Generate fix diff for finding #0
sre-agent pr --finding 0          # Create PR with the fix
sre-agent full-cycle --duration 1h  # Run complete detect → analyze → issue → fix → deploy → monitor
sre-agent status                  # Show current config and connections
sre-agent test-connections        # Verify Grafana, GitHub, AI connectivity
```

Flags:
- `--duration <window>` — time window for log queries: `5m`, `1h`, `6h`, `24h`, `7d` (default: `1h`)
- `--finding <N>` — operate on the Nth detected error (0-indexed)
- `--mode <guided|autonomous>` — guided waits for human approval; autonomous deploys automatically
- `--dry-run` — detect + analyze only, no PRs or deploys

## Agent-Driven Workflow

When invoked as a skill (not directly from CLI), Claude follows this workflow:

1. Read `references/steps/step-00-preflight.md` — verify config and connections
2. Read `references/steps/step-01-query-logs.md` — detect errors
3. Read `references/steps/step-02-analyze.md` — AI root-cause analysis
4. Read `references/steps/step-03-report.md` — create issue, PR, deploy

After each step, report findings to the user before proceeding.

## Step Files

Each step lives in `{skill-root}/references/steps/`:

| Step | File | Purpose |
|------|------|---------|
| 0 | `step-00-preflight.md` | Validate config, test connections |
| 1 | `step-01-query-logs.md` | Query logs, deduplicate errors |
| 2 | `step-02-analyze.md` | Send errors to AI, get root cause |
| 3 | `step-03-report.md` | Create GitHub issue, PR, optionally deploy |

Read the relevant step file before executing it. Each step file contains the exact commands and logic.

## Modes

| Mode | Behavior |
|------|----------|
| `guided` | Detects → analyzes → opens PR → waits for human approval before deploy |
| `autonomous` | Detects → analyzes → opens PR → auto-merges if CI passes → deploys → monitors |

Default: `guided`. Autonomous mode requires explicit configuration (`deployment.auto_merge: true`).

## Output

Each run produces:
- **Console output** — structured log lines with timestamps
- **GitHub issue** — if errors exceed threshold
- **GitHub PR** — if AI produces a fix diff and threshold is met
- **Run record** — persisted to `{config-dir}/history.json`

Run history is also exported to `{config-dir}/memory.md` for quick review.
