# sre-agent-skill

A Claude Code / Codex agent skill that turns natural-language SRE requests —
*"what's broken in production"*, *"run diagnostics"*, *"analyze the latest
errors"*, *"auto-fix the failing service"*, *"show me the error dashboard"*,
*"morning SRE check"* — into sre-agent CLI invocations (detect, analyze,
fix, status). Uses [sre-agent](https://github.com/deemwar-products/ops-phoenix-agent)
as the log-analysis engine and Grafana/Loki for observability.

---

## Requirements

### Always required

| Tool | For | Install |
|---|---|---|
| **Claude Code** or **Codex** | The agent runtime loading this skill | https://claude.com/claude-code |
| **sre-agent** | Log analysis engine | `go install github.com/deemwar-products/ops-phoenix-agent/sre-agent/cmd/sre-agent@latest` |
| **Go 1.24+** | Building sre-agent from source | `brew install go` |
| **Grafana + Loki** | Observability backend | Must have Loki datasource configured |

### Optional (for full functionality)

| Tool | For | Install |
|---|---|---|
| **Anthropic API key** | AI analysis (analyze + fix commands) | https://console.anthropic.com/ |
| **GitHub CLI (`gh`)** | Fix PR creation | `brew install gh && gh auth login` |
| **GitHub PAT** | Fix PR creation (alternative to gh) | GitHub → Settings → Tokens |

### Never required

- **A cloud account** — sre-agent is local-first. It queries your existing
  Grafana/Loki instance; no cloud signup needed.
- **Docker, Python, npm** beyond what your OS ships.

---

## Install

```bash
# 1. Install sre-agent (if not already installed)
go install github.com/deemwar-products/ops-phoenix-agent/sre-agent/cmd/sre-agent@latest

# 2. Install this skill (symlinks into ~/.claude/skills/)
git clone https://github.com/deemwar-products/ops-phoenix-agent
cd ops-phoenix-agent
sh skills/sre-agent-skill/install.sh

# 3. Configure sre-agent
sre-agent setup

# 4. Set credentials
sre-agent auth status
sre-agent auth set grafana_token <your-grafana-token>
sre-agent auth set ANTHROPIC_API_KEY <your-anthropic-key>

# 5. Restart your agent session to pick up the skill
```

Override the Claude skills dir:

```bash
CLAUDE_SKILLS_DIR=/some/other/path sh install.sh
```

## Uninstall

```bash
sh uninstall.sh
```

## What This Skill Does

| You say | sre-agent runs |
|---|---|
| "what's broken in production" | `detect` → summarize errors |
| "analyze the latest errors" | `analyze` → AI root cause |
| "auto-fix the failing service" | `fix` → patch + PR |
| "show SRE status" | `status` → health dashboard |
| "morning SRE check" | Full pipeline: detect → analyze → fix |
| "configure credentials" | `auth status` / `auth set` |

## Skill Structure

```
sre-agent-skill/
  SKILL.md              Frontmatter + core rules + process
  README.md             This file
  LICENSE               MIT
  install.sh            Symlinks into ~/.claude/skills/ + ~/.agents/skills/
  uninstall.sh          Removes symlinks
  bin/
    install.sh          Copy for npm package
    uninstall.sh        Copy for npm package
  references/
    workflow.md         Intent routing, preflight, failure handling
    detect.md           Detect command recipe
    analyze.md          Analyze command recipe
    fix.md              Fix command recipe
    status.md           Status command recipe
    auth.md             Auth command recipe
    setup.md            Setup and configuration guide
    error-classification.md  Severity levels and common patterns
  tests/
    all.sh              Run all tests
    validate-recipes.sh File existence + format check
    install-test.sh     Install/uninstall idempotency
```

## Development

```bash
sh tests/all.sh          # run full test suite
sh tests/validate-recipes.sh   # quick format check
sh tests/install-test.sh       # install/uninstall cycle
```
