# Ops Phoenix Framework

A general-purpose SRE automation framework that monitors logs, detects errors, analyzes with AI, and automatically creates PRs to fix issues.

## Features

- **Multi-Provider Support**: Works with Grafana Cloud, Loki, Elasticsearch, and more
- **AI-Powered Analysis**: Uses Claude to analyze errors and suggest fixes
- **Auto-Fix**: Generates and applies code fixes automatically
- **Auto-Deploy**: Triggers CI/CD pipelines and monitors deployments
- **Self-Healing**: Retries on failure, creates issues for manual review

## Quick Start

```bash
# Navigate to framework directory
cd infra/skills/reqsume-ops-phoenix/framework

# Run the agent (will ask setup questions on first run)
python3 ops_phoenix.py

# Or use the entry point
python3 ops_phoenix_main.py --dry-run
```

## Configuration

### First Run

On first run, the agent will ask questions to configure:

```
=== OPS PHOENIX CONFIGURATION WIZARD ===

1. What logging/monitoring system?
   - Grafana Cloud (SaaS)
   - Loki (self-hosted)
   - Grafana + Loki (self-hosted)

2. GitHub repository (owner/repo)

3. CI/CD system?
   - GitHub Actions
   - GitLab CI
   - Jenkins
   - Manual

4. Deployment settings
```

### Manual Configuration

Create `~/.ops-phoenix/config.json`:

```json
{
  "observability": {
    "type": "grafana_cloud",
    "grafana_url": "https://your-company.grafana.net",
    "grafana_token_path": "/path/to/token",
    "container_patterns": ["app-*"],
    "host_patterns": ["prod-*"],
    "error_patterns": ["level=~\"(?i)error|fatal\""]
  },
  "github": {
    "repo": "your-org/your-repo",
    "base_branch": "main",
    "default_labels": ["sre-alert"]
  },
  "cicd": {
    "type": "github_actions",
    "workflow_name": "Deploy to Production"
  }
}
```

## Usage

```bash
# First time: Run setup wizard
python3 ops_phoenix.py --setup

# Detect errors only
python3 ops_phoenix.py --dry-run

# Full cycle (detect → analyze → fix → PR → deploy)
python3 ops_phoenix.py --env prod --full-cycle

# Custom time window
python3 ops_phoenix.py --window 24h --dry-run

# Show current config
python3 ops_phoenix.py --config
```

## Supported Providers

### Observability

| Provider | Type | Access Needed |
|----------|------|--------------|
| Grafana Cloud | SaaS | API token |
| Loki | Self-hosted | URL + credentials |
| Grafana + Loki | Self-hosted | URL + API token |
| (More coming) | | |

### CI/CD

| Provider | Status |
|----------|--------|
| GitHub Actions | ✅ Supported |
| GitLab CI | Planned |
| Jenkins | Planned |
| ArgoCD | Planned |

## Architecture

```
ops_phoenix.py (main agent)
    │
    ├── adapters/
    │   └── observability.py    # Log providers
    │
    ├── config_wizard.py        # Setup questions
    │
    └── templates/
        └── sample_config.json  # Config template
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `ANTHROPIC_API_KEY` | Claude API key for AI analysis |
| `ANTHROPIC_BASE_URL` | Custom API proxy (optional) |
| `GITHUB_TOKEN` | GitHub personal access token |

## Security

- API tokens stored in files, not in code
- Config stored in `~/.ops-phoenix/`
- GitHub token via environment variable
- No secrets in config file

## Future Plans

- [ ] GitLab CI adapter
- [ ] Jenkins adapter
- [ ] Elasticsearch adapter
- [ ] AWS CloudWatch adapter
- [ ] Slack notifications
- [ ] PagerDuty integration
- [ ] Web UI for configuration
- [ ] Multi-project support
