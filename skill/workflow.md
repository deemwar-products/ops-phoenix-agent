# Workflow: SRE Phoenix Agent (Cron)

How to set up the SRE Phoenix agent to run automatically on a schedule.

## Prerequisites

1. Ubuntu/Debian VM or macOS machine
2. Access to Grafana API token
3. `gh` CLI authenticated (`gh auth login`)
4. `ANTHROPIC_API_KEY` environment variable set

## Setup Steps

### 1. Install Dependencies

```bash
# macOS
brew install jq gh

# Ubuntu/Debian
apt update && apt install jq curl
```

### 2. Authenticate with GitHub

```bash
gh auth login --hostname github.com
gh auth status
```

### 3. Set Environment Variables

```bash
export ANTHROPIC_API_KEY="sk-ant-..."  # Add to ~/.zshrc or ~/.bashrc
```

### 4. Copy Scripts

```bash
cp infra/skills/reqsume-sre-phoenix/scripts/*.sh /usr/local/bin/
chmod +x /usr/local/bin/check-errors.sh
```

### 5. Test Locally

```bash
./check-errors.sh --dry-run
./check-errors.sh --create-issue
```

## Cron Setup

```bash
# Edit crontab
crontab -e

# Add: run every hour at minute 0
0 * * * * /usr/local/bin/check-errors.sh --create-issue >> ~/logs/sre-phoenix.log 2>&1
```

## Output Files

| File | Purpose |
|------|---------|
| `~/logs/sre-phoenix.log` | Error summary |
| `/tmp/sre-phoenix/errors-*.json` | Raw error data |
| `/tmp/sre-phoenix/analyses/*.md` | Claude analysis |

## Monitoring

```bash
# Watch logs
tail -f ~/logs/sre-phoenix.log

# Manual run
./check-errors.sh --dry-run
```

## Troubleshooting

| Problem | Solution |
|---------|-----------|
| "ANTHROPIC_API_KEY not set" | Add to shell profile |
| "gh auth required" | Run `gh auth login` |
| Cron not running | Check `launchctl list` or `systemctl status cron` |
