# Workflow: SRE Agent (Cron / Scheduled)

How to set up SRE Agent to run automatically on a schedule.

## Prerequisites

1. Linux VM or macOS machine with network access to Grafana and GitHub
2. Grafana API token with Viewer permissions
3. `gh` CLI authenticated (`gh auth login`)
4. AI API key set in environment (e.g., `ANTHROPIC_API_KEY`)

## Setup Steps

### 1. Install Dependencies

```bash
# macOS
brew install jq gh

# Ubuntu/Debian
apt update && apt install jq curl python3
```

### 2. Configure the Agent

```bash
sre-agent init
```

This runs the interactive wizard that asks for:
- Grafana URL
- Grafana token path
- GitHub repo
- AI API key env var
- Mode (guided / autonomous)

Config is saved to `{config-dir}/config.json`.

### 3. Test Connections

```bash
sre-agent test-connections
```

Verify Grafana, GitHub, and AI connectivity before scheduling.

### 4. Test a Dry Run

```bash
sre-agent detect --duration 5m
```

### 5. Schedule

#### systemd timer (Linux)

```bash
# /etc/systemd/system/sre-agent.service
[Unit]
Description=SRE Agent — autonomous incident cycle

[Service]
Type=oneshot
ExecStart=/opt/sre-agent/bin/run.sh
TimeoutStartSec=900

# /etc/systemd/system/sre-agent.timer
[Timer]
OnCalendar=hourly
Persistent=true
RandomizedDelaySec=90

[Install]
WantedBy=timers.target
```

```bash
systemctl enable --now sre-agent.timer
```

#### cron (any Unix)

```bash
# Run every hour at minute 17
17 * * * * /opt/sre-agent/bin/run.sh >> /var/log/sre-agent/runs.log 2>&1
```

## Output Files

| File | Purpose |
|------|---------|
| `{config-dir}/history.json` | Run records (issues, PRs, deploys) |
| `{config-dir}/memory.md` | Recent run summary |
| `{config-dir}/errors-*.json` | Raw error data per run |
| `/var/log/sre-agent/runs/*.log` | Run logs |

## Monitoring

```bash
# Watch logs
tail -f /var/log/sre-agent/runs/*.log

# Manual run
sre-agent full-cycle --duration 1h
```

## Troubleshooting

| Problem | Solution |
|---------|----------|
| "config not found" | Run `sre-agent init` |
| "AI API key not set" | Export the env var configured in `ai.api_key_env` |
| "gh auth required" | Run `gh auth login` |
| "Grafana connection failed" | Check token file path and URL in config |
| Cron not running | Check `systemctl status sre-agent.timer` or `crontab -l` |
