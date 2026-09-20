# Ops Phoenix Framework Documentation

> **Self-Healing SRE Agent for Automated Error Detection and Fix**

---

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Quick Start](#quick-start)
4. [Configuration](#configuration)
5. [Observability Providers](#observability-providers)
6. [CI/CD Integration](#cicd-integration)
7. [Workflow](#workflow)
8. [API Reference](#api-reference)
9. [Troubleshooting](#troubleshooting)
10. [Future Plans](#future-plans)

---

## Overview

### What is Ops Phoenix?

Ops Phoenix is a **general-purpose SRE automation framework** that:

1. **Monitors** production logs for errors
2. **Analyzes** errors using AI to find root causes
3. **Fixes** code automatically when errors are detected
4. **Deploys** fixes via CI/CD pipelines
5. **Monitors** deployments and auto-retries on failure

### Key Features

| Feature | Description |
|---------|-------------|
| 🎯 **Multi-Provider** | Works with Grafana Cloud, Loki, Elasticsearch, and more |
| 🤖 **AI-Powered** | Uses Claude to analyze errors and suggest fixes |
| 🔧 **Auto-Fix** | Generates and applies code fixes automatically |
| 🚀 **Auto-Deploy** | Triggers CI/CD pipelines and monitors deployments |
| 🔄 **Self-Healing** | Retries on failure, creates issues for manual review |
| ⚙️ **Configurable** | Asks questions to configure for any project |

---

## Architecture

### System Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           OPS PHOENIX FRAMEWORK                               │
└─────────────────────────────────────────────────────────────────────────────┘

                              ┌─────────────────┐
                              │   User Input    │
                              │  (env vars or   │
                              │   wizard)       │
                              └────────┬────────┘
                                       │
                                       ▼
┌──────────────────────────────────────────────────────────────────────────────┐
│                           CONFIGURATION LAYER                                 │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                        ~/.ops-phoenix/config.json                    │    │
│  │                                                                       │    │
│  │   {                                                                   │    │
│  │     "observability": { "type": "...", ... },                        │    │
│  │     "github": { "repo": "...", ... },                               │    │
│  │     "cicd": { "type": "...", ... },                                 │    │
│  │     "deployment": { ... }                                           │    │
│  │   }                                                                   │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
└──────────────────────────────────────────────────────────────────────────────┘
                                       │
         ┌──────────────────────────────┼──────────────────────────────┐
         │                              │                              │
         ▼                              ▼                              ▼
┌─────────────────┐         ┌─────────────────┐         ┌─────────────────┐
│   OBSERVABILITY │         │     GITHUB      │         │      CI/CD      │
│     ADAPTERS    │         │    ADAPTERS     │         │    ADAPTERS     │
├─────────────────┤         ├─────────────────┤         ├─────────────────┤
│                 │         │                 │         │                 │
│  ┌───────────┐  │         │  ┌───────────┐  │         │  ┌───────────┐  │
│  │  Grafana  │  │         │  │   GitHub  │  │         │  │   GitHub  │  │
│  │   Cloud   │  │         │  │     API   │  │         │  │  Actions  │  │
│  └───────────┘  │         │  └───────────┘  │         │  └───────────┘  │
│                 │         │                 │         │                 │
│  ┌───────────┐  │         │  ┌───────────┐  │         │  ┌───────────┐  │
│  │    Loki   │  │         │  │  GitLab   │  │         │  │  GitLab   │  │
│  │  (direct) │  │         │  │     API   │  │         │  │     CI    │  │
│  └───────────┘  │         │  └───────────┘  │         │  └───────────┘  │
│                 │         │                 │         │                 │
│  ┌───────────┐  │         │  ┌───────────┐  │         │  ┌───────────┐  │
│  │  Grafana  │  │         │  │  Jenkins  │  │         │  │  Jenkins  │  │
│  │ Self-Host │  │         │  │     API   │  │         │  │    API    │  │
│  └───────────┘  │         │  └───────────┘  │         │  └───────────┘  │
│                 │         │                 │         │                 │
└────────┬────────┘         └────────┬────────┘         └────────┬────────┘
         │                             │                             │
         ▼                             │                             ▼
┌─────────────────────────────────────┴────────────────────────────────────────┐
│                              OPS PHOENIX AGENT                                │
│                                                                               │
│    ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐     │
│    │  DETECT  │→ │ ANALYZE  │→ │  ISSUE   │→ │   FIX   │→ │   PR     │     │
│    │          │  │ (Claude) │  │          │  │(Claude) │  │          │     │
│    └──────────┘  └──────────┘  └──────────┘  └──────────┘  └──────────┘     │
│                                                      │             │         │
│                                                      │             ▼         │
│                                              ┌──────────┐  ┌──────────┐      │
│                                              │  DEPLOY  │→ │ MONITOR  │      │
│                                              │          │  │          │      │
│                                              └──────────┘  └────┬─────┘      │
│                                                                      │         │
│                                                      ┌───────────────┴──────┐ │
│                                                      │                       │ │
│                                                      ▼                       ▼ │
│                                              ┌──────────┐           ┌──────────┐
│                                              │ SUCCESS  │           │  AUTO    │
│                                              │   ✓      │           │   FIX    │
│                                              └──────────┘           └──────────┘
└──────────────────────────────────────────────────────────────────────────────┘
                                       │
                                       ▼
┌──────────────────────────────────────────────────────────────────────────────┐
│                           EXTERNAL SERVICES                                    │
│                                                                               │
│    ┌────────────────┐   ┌────────────────┐   ┌────────────────┐                │
│    │  Grafana Cloud │   │  GitHub       │   │  CI/CD System  │                │
│    │     Loki       │   │  Repository   │   │  (Actions)    │                │
│    └────────────────┘   └────────────────┘   └────────────────┘                │
│                                                                               │
└──────────────────────────────────────────────────────────────────────────────┘
```

### Data Flow

```
┌────────────────────────────────────────────────────────────────────────┐
│                        DATA FLOW DIAGRAM                                │
└────────────────────────────────────────────────────────────────────────┘

   User/VM                      Ops Phoenix                        External
     │                               │                               │
     │  python3 ops_phoenix.py        │                               │
     │───────────────────────────────▶                               │
     │                               │                               │
     │                    ┌──────────┴──────────┐                    │
     │                    │    Load Config       │                    │
     │                    │  ~/.ops-phoenix/config │                    │
     │                    └──────────┬──────────┘                    │
     │                               │                               │
     │                               │  1. DETECT                     │
     │                               │  ──────────                    │
     │                               │                               │
     │                    ┌──────────┴──────────┐                    │
     │                    │ Query Observability  │                    │
     │                    │   Provider (Loki)    │◀──────────────────│
     │                    └──────────┬──────────┘   logs/errors     │
     │                               │                               │
     │                    ┌──────────┴──────────┐                    │
     │                    │  Parse & Count       │                    │
     │                    │    Error Types        │                    │
     │                    └──────────┬──────────┘                    │
     │                               │                               │
     │            errors found?      │                               │
     │                    ┌─────────┴─────────┐                      │
     │                    │                   │                      │
     │                   YES                  NO                      │
     │                    │                   │                      │
     │                    ▼                   │                      │
     │         ┌──────────────────┐             │                      │
     │         │ 2. ANALYZE       │             │                      │
     │         │ (Claude AI)     │             │                      │
     │         └────────┬─────────┘             │                      │
     │                  │                       │                      │
     │                  ▼                       │                      │
     │         ┌──────────────────┐             │                      │
     │         │ 3. CREATE ISSUE  │             │                      │
     │         │ (GitHub)         │             │                      │
     │         └────────┬─────────┘             │                      │
     │                  │                       │                      │
     │                  ▼                       │                      │
     │         ┌──────────────────┐             │                      │
     │         │ 4. GENERATE FIX  │             │                      │
     │         │ (Claude AI)     │             │                      │
     │         └────────┬─────────┘             │                      │
     │                  │                       │                      │
     │                  ▼                       │                      │
     │         ┌──────────────────┐             │                      │
     │         │ 5. CREATE PR      │             │                      │
     │         │ (GitHub)         │──────────────▶  PR created         │
     │         └────────┬─────────┘             │                      │
     │                  │                       │                      │
     │                  ▼                       │                      │
     │         ┌──────────────────┐             │                      │
     │         │ 6. MERGE PR      │──────────────▶  Branch merged      │
     │         └────────┬─────────┘             │                      │
     │                  │                       │                      │
     │                  ▼                       │                      │
     │         ┌──────────────────┐             │                      │
     │         │ 7. TRIGGER DEPLOY│──────────────▶  Workflow triggered  │
     │         └────────┬─────────┘             │                      │
     │                  │                       │                      │
     │                  ▼                       │                      │
     │         ┌──────────────────┐             │                      │
     │         │ 8. MONITOR       │──────────────▶  Poll workflow      │
     │         │                  │◀───────────────── status updates   │
     │         └────────┬─────────┘             │                      │
     │                  │                       │                      │
     │         ┌────────┴────────┐              │                      │
     │         │                 │              │                      │
     │        YES               NO              │                      │
     │         │                 │              │                      │
     │         ▼                 │              │                      │
     │  ┌──────────┐              │              │                      │
     │  │ SUCCESS  │              │              │                      │
     │  └──────────┘              │              │                      │
     │                            ▼              │                      │
     │                   ┌──────────────────┐    │                      │
     │                   │ 8a. AUTO-FIX     │    │                      │
     │                   │ (analyze failure)│    │                      │
     │                   └────────┬─────────┘    │                      │
     │                            │              │                      │
     │                            ▼              │                      │
     │                   ┌──────────────────┐    │                      │
     │                   │ 8b. RETRY DEPLOY  │──────▶  Retry workflow   │
     │                   └──────────────────┘    │                      │
     │                            │              │                      │
     │                            │              │                      │
     │                   (max retries exceeded)  │                      │
     │                            │              │                      │
     │                            ▼              │                      │
     │                   ┌──────────────────┐    │                      │
     │                   │ 8c. CREATE ISSUE  │──────▶  Manual review    │
     │                   │ (for human)      │    │                      │
     │                   └──────────────────┘    │                      │
     │                                              │                      │
     │◀───────────────────────────────────────────│                      │
     │  ✓ All systems healthy!                    │                      │
```

---

## Quick Start

### Option 1: With Environment Variables (Non-Interactive)

```bash
# Set required environment variables
export GRAFANA_URL="https://your-company.grafana.net"
export GITHUB_REPO="your-org/your-repo"
export ANTHROPIC_API_KEY="sk-ant-..."

# Run dry run (detect only)
python3 ops_phoenix.py --dry-run

# Run full cycle (detect → analyze → fix → deploy)
python3 ops_phoenix.py --env prod --full-cycle
```

### Option 2: Interactive Setup (First Time)

```bash
# Run without config - will ask questions
python3 ops_phoenix.py

# Or force setup wizard
python3 ops_phoenix.py --setup
```

The wizard will ask:

```
=== OPS PHOENIX CONFIGURATION WIZARD ===

--- OBSERVABILITY SETUP ---
What logging/monitoring system do you use?
  1. grafana_cloud (SaaS)
  2. loki_direct (self-hosted)
  3. grafana_self_hosted

Enter Grafana Cloud URL: https://your-company.grafana.net
Path to Grafana API token: /path/to/token

Container patterns (regex): app-*,api-*
Host patterns (regex): prod-*
Error patterns: level=~"(?i)error|fatal"

--- GITHUB SETUP ---
GitHub repository (owner/repo): your-org/your-repo
Default branch: main

--- CI/CD SETUP ---
Which CI/CD system? 1. GitHub Actions
Workflow name: Deploy to Production
```

---

## Configuration

### Config File Location

```
~/.ops-phoenix/config.json
```

### Full Configuration Example

```json
{
  "observability": {
    "type": "grafana_cloud",
    "grafana_url": "https://your-company.grafana.net",
    "grafana_token_path": "/path/to/grafana-api-token",
    "loki_url": "",
    "loki_user": "",
    "loki_password": "",
    "container_patterns": ["app-*", "api-*"],
    "host_patterns": ["prod-*", "app1|app2"],
    "error_patterns": ["level=~\"(?i)error|fatal\""]
  },
  "github": {
    "repo": "your-org/your-repo",
    "token_env_var": "GITHUB_TOKEN",
    "base_branch": "main",
    "default_labels": ["sre-alert", "bug"]
  },
  "cicd": {
    "type": "github_actions",
    "workflow_name": "Deploy to Production"
  },
  "deployment": {
    "auto_merge": true,
    "monitor_timeout_seconds": 300,
    "max_retries": 2
  }
}
```

### Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `ANTHROPIC_API_KEY` | Claude API key | Yes |
| `GRAFANA_URL` | Grafana Cloud URL | For grafana_cloud |
| `GRAFANA_TOKEN_PATH` | Path to Grafana token | For grafana_cloud |
| `LOKI_URL` | Direct Loki URL | For loki_direct |
| `GITHUB_REPO` | Repository (owner/repo) | Yes |
| `GITHUB_TOKEN` | GitHub personal access token | Yes |

---

## Observability Providers

### 1. Grafana Cloud

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│ Your App    │────▶│ Grafana     │────▶│ Loki        │
│ (containers)│     │ Cloud       │     │ (storage)   │
└─────────────┘     └──────┬──────┘     └─────────────┘
                           │
                    ┌──────┴──────┐
                    │ Grafana API │  ← Ops Phoenix queries this
                    └─────────────┘
```

**Config:**
```json
{
  "observability": {
    "type": "grafana_cloud",
    "grafana_url": "https://your-company.grafana.net",
    "grafana_token_path": "/path/to/token"
  }
}
```

### 2. Loki Direct (Self-Hosted)

```
┌─────────────┐     ┌─────────────┐
│ Your App    │────▶│ Loki        │
│             │     │ (on VM)     │
└─────────────┘     └──────┬──────┘
                    ┌──────┴──────┐
                    │ Loki API    │  ← Ops Phoenix queries this
                    │ :3100       │
                    └─────────────┘
```

**Config:**
```json
{
  "observability": {
    "type": "loki_direct",
    "loki_url": "http://loki:3100",
    "loki_user": "admin",
    "loki_password": "secret"
  }
}
```

### 3. Grafana Self-Hosted

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│ Your App    │────▶│ Loki        │────▶│ Grafana     │
└─────────────┘     │             │     │ (frontend)   │
                    └─────────────┘     └──────┬──────┘
                                            ┌──────┴──────┐
                                            │ Grafana API │  ← Ops Phoenix
                                            └─────────────┘
```

**Config:**
```json
{
  "observability": {
    "type": "grafana_self_hosted",
    "grafana_url": "http://grafana:3000",
    "grafana_token_path": "/path/to/token"
  }
}
```

---

## CI/CD Integration

### GitHub Actions

```yaml
# .github/workflows/ops-phoenix.yml
name: Ops Phoenix

on:
  schedule:
    - cron: '*/10 * * * *'  # Every 10 minutes

jobs:
  ops-phoenix:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Python
        uses: actions/setup-python@v5
        with:
          python-version: '3.10'

      - name: Run Ops Phoenix
        env:
          ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
          python3 ops_phoenix.py --env prod --full-cycle
```

### Environment Modes

| Mode | Flag | Behavior |
|------|------|----------|
| `dev` | `--env dev` | Creates PR but does NOT auto-merge |
| `prod` | `--env prod` | Auto-merges and deploys |

---

## Workflow

### Full Cycle

```
┌─────────────────────────────────────────────────────────────────┐
│                    OPS PHOENIX FULL CYCLE                       │
└─────────────────────────────────────────────────────────────────┘

START
  │
  ▼
┌─────────────────┐
│ 1. DETECT       │
│ Query logs for  │
│ errors          │
└────────┬────────┘
         │
         ▼
┌─────────────────┐     NO ERRORS
│ Any errors?     │──────────────▶ END: All healthy ✓
└────────┬────────┘
         │ YES
         ▼
┌─────────────────┐
│ 2. ANALYZE      │
│ Claude AI       │
│ Root cause      │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ 3. CREATE ISSUE │
│ GitHub          │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ 4. GENERATE FIX │
│ Claude AI       │
│ Code changes    │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ 5. CREATE PR    │
│ GitHub          │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ 6. DEPLOY       │
│ Merge & trigger │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ 7. MONITOR      │
│ Wait for        │
│ completion      │
└────────┬────────┘
         │
    ┌────┴────┐
    │         │
  SUCCESS   FAILURE
    │         │
    │         ▼
    │ ┌─────────────────┐
    │ │ 8. AUTO-FIX    │
    │ │ Analyze failure│
    │ │ Generate fix   │
    │ └────────┬──────┘
    │          │
    │    ┌─────┴─────┐
    │    │           │
    │  SUCCESS     FAIL
    │    │           │
    │    │     ┌─────┴─────┐
    │    │     │ 9. CREATE │
    │    │     │   ISSUE   │
    │    │     │  (manual) │
    │    │     └───────────┘
    │    │
    └────┘
     │
     ▼
   END ✓
```

---

## API Reference

### Command Line Arguments

| Argument | Description | Example |
|----------|-------------|---------|
| `--dry-run` | Detect errors only, no changes | `ops_phoenix.py --dry-run` |
| `--full-cycle` | Complete workflow | `ops_phoenix.py --full-cycle` |
| `--env <mode>` | Environment (dev/prod) | `ops_phoenix.py --env prod --full-cycle` |
| `--window <time>` | Time window (1h/6h/24h/7d) | `ops_phoenix.py --window 24h` |
| `--setup` | Run configuration wizard | `ops_phoenix.py --setup` |
| `--config` | Show current config | `ops_phoenix.py --config` |

### Programmatic Usage

```python
from ops_phoenix import OpsPhoenix

# Create agent
agent = OpsPhoenix()

# Run detection
errors = agent.detect_errors("1h")

# Run full cycle
result = agent.run_full_cycle()

print(f"Status: {result['status']}")
print(f"Issue: #{result['issue']}")
print(f"PR: #{result['pr']}")
```

---

## Troubleshooting

### Common Issues

| Issue | Solution |
|-------|----------|
| `Config not found` | Run `python3 ops_phoenix.py --setup` |
| `Token not found` | Check `GRAFANA_TOKEN_PATH` env var |
| `GitHub not authenticated` | Run `gh auth login` |
| `No errors found` | System is healthy, increase time window |
| `Workflow timeout` | Increase `monitor_timeout_seconds` in config |

### Debug Mode

```bash
# Enable verbose output
export DEBUG=1
python3 ops_phoenix.py --dry-run
```

---

## Future Plans

### Planned Adapters

| Provider | Status | Priority |
|----------|--------|----------|
| Elasticsearch | 🔜 Planned | High |
| AWS CloudWatch | 🔜 Planned | Medium |
| Azure Monitor | 🔜 Planned | Low |
| Datadog | 🔜 Planned | Low |

### CI/CD Adapters

| Provider | Status | Priority |
|----------|--------|----------|
| GitLab CI | 🔜 Planned | High |
| Jenkins | 🔜 Planned | Medium |
| ArgoCD | 🔜 Planned | Low |
| CircleCI | 🔜 Planned | Low |

### Feature Roadmap

- [ ] Web UI for configuration
- [ ] Multi-project support
- [ ] Slack notifications
- [ ] PagerDuty integration
- [ ] Dashboard for monitoring
- [ ] Scheduled reports
- [ ] Custom fix templates
- [ ] A/B testing for fixes

---

## Architecture Diagram (Colorful)

```
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃                                                                              ┃
┃                         🦅  OPS PHOENIX  🦅                                ┃
┃                       Self-Healing SRE Agent                                ┃
┃                                                                              ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛

                              ┌──────────────────┐
                              │   💻  USER / VM   │
                              │  Environment Vars │
                              └────────┬─────────┘
                                       │
                                       ▼
    ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
    ┃                    🎨  CONFIGURATION LAYER                            ┃
    ┃                                                                   ┃
    ┃   ┌────────────────────────────────────────────────────────────┐  ┃
    ┃   │          🗂️  ~/.ops-phoenix/config.json                    │  ┃
    ┃   │                                                             │  ┃
    ┃   │   ┌──────────────┐  ┌──────────────┐  ┌──────────────┐    │  ┃
    ┃   │   │ OBSERVABILITY │  │   GITHUB     │  │    CI/CD     │    │  ┃
    ┃   │   │  🟢 Config    │  │  🔵 Config   │  │  🟡 Config   │    │  ┃
    ┃   │   └──────────────┘  └──────────────┘  └──────────────┘    │  ┃
    ┃   └────────────────────────────────────────────────────────────┘  ┃
    ┃                                                                   ┃
    ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
                                       │
          ┌────────────────────────────┼────────────────────────────┐
          │                            │                            │
          ▼                            ▼                            ▼
    ┏━━━━━━━━━━━━━━━┓         ┏━━━━━━━━━━━━━━━┓         ┏━━━━━━━━━━━━━━━┓
    ┃  OBSERVABILITY ┃         ┃    GITHUB     ┃         ┃     CI/CD      ┃
    ┃    ADAPTERS    ┃         ┃   ADAPTERS    ┃         ┃   ADAPTERS     ┃
    ┃               ┃         ┃               ┃         ┃               ┃
    ┃ ┌───────────┐ ┃         ┃ ┌───────────┐ ┃         ┃ ┌───────────┐ ┃
    ┃ │ 🟢 Grafana│ ┃         ┃ │   🟢 GitHub│ ┃         ┃ │ 🟢 GitHub │ ┃
    ┃ │   Cloud   │ ┃         ┃ │    API    │ ┃         ┃ │  Actions  │ ┃
    ┃ └───────────┘ ┃         ┃ └───────────┘ ┃         ┃ └───────────┘ ┃
    ┃               ┃         ┃               ┃         ┃               ┃
    ┃ ┌───────────┐ ┃         ┃ ┌───────────┐ ┃         ┃ ┌───────────┐ ┃
    ┃ │ 🟡 Loki   │ ┃         ┃ │ 🟡GitLab  │ ┃         ┃ │ 🟡GitLab  │ ┃
    ┃ │  Direct   │ ┃         ┃ │    API    │ ┃         ┃ │     CI    │ ┃
    ┃ └───────────┘ ┃         ┃ └───────────┘ ┃         ┃ └───────────┘ ┃
    ┃               ┃         ┃               ┃         ┃               ┃
    ┃ ┌───────────┐ ┃         ┃ ┌───────────┐ ┃         ┃ ┌───────────┐ ┃
    ┃ │ 🔴 Grafana│ ┃         ┃ │ 🔴Jenkins │ ┃         ┃ │ 🔴Jenkins │ ┃
    ┃ │Self-Hosted│ ┃         ┃ │    API    │ ┃         ┃ │    API    │ ┃
    ┃ └───────────┘ ┃         ┃ └───────────┘ ┃         ┃ └───────────┘ ┃
    ┃               ┃         ┃               ┃         ┃               ┃
    ┗━━━━━━━┳━━━━━━━┛         ┗━━━━━━━┳━━━━━━━┛         ┗━━━━━━━┳━━━━━━━┛
            │                        │                        │
            └────────────────────────┼────────────────────────┘
                                     │
                                     ▼
    ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
    ┃                      🔷  OPS PHOENIX AGENT 🔷                       ┃
    ┃                                                                   ┃
    ┃    ╔════════╗   ╔════════╗   ╔════════╗   ╔════════╗   ╔════════╗ ┃
    ┃    ║   1.   ║   ║   2.   ║   ║   3.   ║   ║   4.   ║   ║   5.   ║ ┃
    ┃    ║DETECT  ║──▶║ANALYZE ║──▶║ ISSUE  ║──▶║  FIX   ║──▶║   PR   ║ ┃
    ┃    ║ 🔍    ║   ║  🤖    ║   ║ 📝    ║   ║  🔧    ║   ║ 📦    ║ ┃
    ┃    ╚════════╝   ╚════════╝   ╚════════╝   ╚════════╝   ╚════════╝ ┃
    ┃                                                             │      ┃
    ┃                                                             ▼      ┃
    ┃                                                ╔════════╗   ╔════════╗
    ┃                                                ║   6.   ║──▶║   7.   ║
    ┃                                                ║ DEPLOY ║   ║MONITOR ║
    ┃                                                ║  🚀   ║   ║  📊   ║
    ┃                                                ╚════════╝   ╚═══╤════╝
    ┃                                                               │
    ┃                                             ┌──────────────────┴──────┐
    ┃                                             │                         │
    ┃                                             ▼                         ▼
    ┃                                     ┌──────────┐              ┌──────────┐
    ┃                                     │  ✅ SUCCESS │            │  ❌ FAILURE  │
    ┃                                     │   Done!    │            │  8. AUTO-FIX │
    ┃                                     └──────────┘              └──────┬─────┘
    ┃                                                                     │
    ┃                                                     ┌───────────────┴───────────────┐
    ┃                                                     │                               │
    ┃                                                     ▼                               ▼
    ┃                                            ┌──────────────┐                 ┌──────────────┐
    ┃                                            │   Retry OK   │                 │  Max Retries  │
    ┃                                            │  ✅ Success  │                 │  ❌ Exceeded   │
    ┃                                            └──────────────┘                 └──────┬───────┘
    ┃                                                                                       │
    ┃                                                                                       ▼
    ┃                                                                              ┌──────────────┐
    ┃                                                                              │ 9. Create    │
    ┃                                                                              │    Manual    │
    ┃                                                                              │    Issue     │
    ┃                                                                              └──────────────┘
    ┃                                                                   ┃
    ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
                                     │
                                     ▼
    ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
    ┃                      🔷  EXTERNAL SERVICES 🔷                         ┃
    ┃                                                                   ┃
    ┃   ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐   ┃
    ┃   │   🟢 Grafana    │  │  🟢  GitHub     │  │ 🟢  CI/CD      │   ┃
    ┃   │  Cloud / Loki   │  │   Repository   │  │   (Actions)    │   ┃
    ┃   └─────────────────┘  └─────────────────┘  └─────────────────┘   ┃
    ┃                                                                   ┃
    ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
                                     │
                                     ▼
                              ┌──────────────┐
                              │   ✅ All     │
                              │  Systems     │
                              │  Healthy!    │
                              └──────────────┘

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

                    🟢 = Production Ready
                    🟡 = Coming Soon
                    🔴 = Planned