# setup

Guide for installing and configuring sre-agent on a new machine.

## Prerequisites

| Tool | Required | Install |
|---|---|---|
| **Go 1.24+** | Yes | `brew install go` |
| **Git** | Yes | `brew install git` |
| **Claude Code** or **Codex** | For skill loading | https://claude.com/claude-code |
| **Grafana instance** | Yes | Must have Loki datasource accessible |
| **Anthropic API key** | For analyze/fix | https://console.anthropic.com/ |
| **GitHub CLI (`gh`)** | For fix (PR creation) | `brew install gh && gh auth login` |
| **Teams webhook** | For notifications | Teams Incoming Webhook or Power Automate |

## Install sre-agent

```bash
go install github.com/deemwar-products/ops-phoenix-agent/sre-agent/cmd/sre-agent@latest
```

Verify:
```bash
sre-agent --help
```

## Install this skill

```bash
git clone https://github.com/muthuishere-agent-skills/sre-agent-skill
cd sre-agent-skill
sh install.sh
```

Or install via npm (when published):
```bash
npx @muthuishere/sre-agent-skill install
```

## Configure sre-agent

Run the setup wizard:
```bash
sre-agent setup
```

This creates `~/.config/sre-agent/config.yaml`. Edit manually for advanced
settings.

### Minimal config

```yaml
observability:
  type: grafana_self_hosted
  url: https://your-grafana-instance.com
  container_patterns:
    - your-service-*
  error_pattern: "(?i)(error|panic|fatal|exception|critical)"

github:
  repo: your-org/your-repo
  base_branch: main

ai:
  provider: anthropic
  model: claude-sonnet-4-20250514
  max_tokens: 1024

agent:
  work_dir: ~/.config/sre-agent
  mode: dev

notify:
  teams:
    enabled: false
    on_success: false
```

### Configure credentials

```bash
sre-agent auth status                # check current state
sre-agent auth set grafana_token <token>
sre-agent auth set ANTHROPIC_API_KEY <key>
sre-agent auth set teams_webhook <url>
```

## Verify

```bash
sre-agent status
sre-agent detect --time-window 1h
```

## Environment-specific configs

For dev vs prod, use separate config directories:
```bash
sre-agent --config ~/.config/sre-agent-dev detect
sre-agent --config ~/.config/sre-agent-prod detect
```

## Troubleshooting

| Issue | Fix |
|---|---|
| `grafana self-hosted query: 401` | Token expired — regenerate in Grafana → API Keys |
| `Anthropic API 401` | Key invalid — check console.anthropic.com |
| `clone failed: exit 128` | Run `gh auth login` or set `GITHUB_TOKEN` |
| `no ANTHROPIC_API_KEY` | Run `sre-agent auth set ANTHROPIC_API_KEY <key>` |
| `time_window` ignored | Pass `--time-window` explicitly on the command line |
