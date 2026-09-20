# Ops Phoenix - Quick Start Guide

## Prerequisites

1. **Python 3.8+**
2. **GitHub CLI** (`gh`) authenticated
3. **Claude API Key** (for AI analysis)

## Installation

```bash
# Clone or navigate to the framework
cd infra/skills/reqsume-ops-phoenix/framework

# Make scripts executable
chmod +x ops_phoenix.py config_wizard.py
```

## Quick Start Options

### Option 1: Non-Interactive (Recommended for VMs)

```bash
# Set environment variables
export GRAFANA_URL="https://your-grafana.grafana.net"
export GRAFANA_TOKEN_PATH="/path/to/grafana-api-token"
export GITHUB_REPO="your-org/your-repo"
export GITHUB_TOKEN="ghp_xxx"
export ANTHROPIC_API_KEY="sk-ant-..."

# Run dry run (detect only)
python3 ops_phoenix.py --dry-run

# Run full cycle
python3 ops_phoenix.py --env prod --full-cycle
```

### Option 2: Interactive Setup (First Time)

```bash
# Run without config - will ask questions
python3 ops_phoenix.py

# Or force setup wizard
python3 ops_phoenix.py --setup
```

---

## Common Commands

| Command | Description |
|---------|-------------|
| `python3 ops_phoenix.py --dry-run` | Detect errors only |
| `python3 ops_phoenix.py --env prod --full-cycle` | Full workflow |
| `python3 ops_phoenix.py --window 24h` | Check last 24 hours |
| `python3 ops_phoenix.py --config` | Show current config |
| `python3 ops_phoenix.py --setup` | Re-run setup wizard |

---

## Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `ANTHROPIC_API_KEY` | Claude API key | Yes |
| `GRAFANA_URL` | Grafana Cloud URL | For Grafana Cloud |
| `GRAFANA_TOKEN_PATH` | Path to Grafana token file | For Grafana Cloud |
| `LOKI_URL` | Direct Loki URL | For Loki Direct |
| `GITHUB_REPO` | Repository (owner/repo) | Yes |
| `GITHUB_TOKEN` | GitHub PAT | Yes |
| `AUTO_MERGE` | Auto-merge PRs (true/false) | No |
| `MONITOR_TIMEOUT` | Workflow timeout in seconds | No |

---

## First-Time Setup Questions

When you run `python3 ops_phoenix.py --setup`, you'll be asked:

### 1. Observability System

```
What logging/monitoring system do you use?
  1. grafana_cloud (Grafana Cloud SaaS)
  2. loki_direct (Self-hosted Loki)
  3. grafana_self_hosted (Self-hosted Grafana + Loki)
```

### 2. GitHub Repository

```
GitHub repository (owner/repo format): myorg/myapp
Default branch: main
Labels for issues/PRs: sre-alert,bug
```

### 3. CI/CD System

```
Which CI/CD system?
  1. github_actions (GitHub Actions)
  2. gitlab_ci (GitLab CI)
  3. jenkins (Jenkins)
  4. manual (No auto-deploy)
```

### 4. Deployment Settings

```
Auto-merge PRs? (yes/no): yes
Monitor timeout (seconds): 300
Max retries on failure: 2
```

---

## Example: Reqsume Production

```bash
# Set for Reqsume
export GRAFANA_URL="https://observability.deemwar.com"
export GRAFANA_TOKEN_PATH="~/Downloads/Archive/keys/grafana-api-token"
export GITHUB_REPO="deemwar-products/reqsume"
export ANTHROPIC_API_KEY="sk-ant-..."

# Detect errors
python3 ops_phoenix.py --dry-run

# Full cycle (creates issue, PR, deploys)
python3 ops_phoenix.py --env prod --full-cycle
```

---

## VM Setup (Permanent)

### Crontab (Every 10 minutes)

```bash
crontab -e

# Add line:
*/10 * * * * export ANTHROPIC_API_KEY="sk-ant-..." && /path/to/ops_phoenix.py --env prod --full-cycle >> /var/log/ops-phoenix.log 2>&1
```

### Systemd Service (More Reliable)

```bash
# Create service file
sudo nano /etc/systemd/system/ops-phoenix.service

[Unit]
Description=Ops Phoenix Agent
After=network.target

[Service]
Type=oneshot
ExecStart=/usr/bin/python3 /opt/ops-phoenix/ops_phoenix.py --env prod --full-cycle
Environment=ANTHROPIC_API_KEY=xxx
Environment=GITHUB_TOKEN=xxx
User=ops-phoenix

[Install]
WantedBy=multi-user.target
```

---

## Troubleshooting

### Check Config

```bash
python3 ops_phoenix.py --config
```

### Test Connections

```bash
python3 ops_phoenix.py --dry-run
# Will show:
# - Observability: Connected/Failed
# - GitHub: Authenticated/Not authenticated
# - Errors found/Not found
```

### Reset Config

```bash
rm ~/.ops-phoenix/config.json
python3 ops_phoenix.py --setup
```

---

## Next Steps

1. Read [ARCHITECTURE.md](./ARCHITECTURE.md) for full documentation
2. Check [TROUBLESHOOTING.md](./TROUBLESHOOTING.md) for common issues
3. Review [API.md](./API.md) for programmatic usage