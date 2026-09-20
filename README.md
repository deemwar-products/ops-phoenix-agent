# SRE Agent

Autonomous SRE agent: detect → analyze → fix → deploy.

This repo holds two artifacts:

| Path | What |
|---|---|
| `framework/` | Python agent (`ops_phoenix` package + CLI). Detects production errors, asks Claude for root cause + fix, opens a PR, optionally deploys. |
| `skill/` | Claude skill (`SKILL.md` + steps + scripts + framework). Same capability, but driven by Claude through the chat. |
| `infra/` | VM installer + systemd units + sample config for running the agent unattended on a Linux host. |
| `docs/` | Specs, runbooks, design notes. |

The marketing landing page is at `src/` (Next.js). Run `npm run dev` from the repo root for the site.

## Quick start — VM

```bash
# 1. Generate a fine-grained GitHub PAT with repo + workflow scope.
# 2. Save it to a file, then:
GH_TOKEN_FILE=/path/to/gh-token.txt \
  OPS_PHOENIX_PINNED_REF=origin/main \
  bash infra/install.sh

# Dry-run preview (no changes made):
bash infra/install.sh --dry-run
```

This installs:
- the agent at `/opt/ops-phoenix`
- a systemd timer `ops-phoenix-hourly.timer` that fires at HH:07 every hour
- the `opsphoenix` system user (uid 999, nologin)
- log retention via `/etc/logrotate.d/ops-phoenix`

## Quick start — local

```bash
cd framework
python3 -m venv .venv && source .venv/bin/activate
pip install -e .

# Configure (writes ~/.ops-phoenix/config.json)
ops-phoenix --action setup

# Run one detect → analyze → fix → deploy cycle
ops-phoenix --action full
```

## Manual trigger

```bash
sudo systemctl start ops-phoenix-hourly.service
journalctl -u ops-phoenix-hourly.service -f
```

## Triage

```bash
# Last 5 runs:
ls -lt /var/log/ops-phoenix/runs/ | head -5

# Tail the latest log:
tail -100 /var/log/ops-phoenix/runs/$(ls -t /var/log/ops-phoenix/runs/ | head -1)

# Service journal:
journalctl -u ops-phoenix-hourly.service -n 50

# Heartbeat:
cat /var/lib/ops-phoenix/heartbeat.prom
```

## Kill switch

```bash
# A) Env var (persistent across reboots)
echo 'OPS_PHOENIX_DISABLED=1' | sudo tee -a /etc/ops-phoenix/ops-phoenix.env
sudo systemctl daemon-reload

# B) Sentinel file (immediate, no daemon-reload needed)
sudo touch /etc/ops-phoenix/disabled

# C) Nuclear
sudo systemctl disable --now ops-phoenix-hourly.timer
```

To re-enable: undo whichever you did, then `sudo systemctl enable --now ops-phoenix-hourly.timer`.

## Rotate credentials

- Anthropic key: edit `/etc/ops-phoenix/secret` (preserve 0600 perms)
- GitHub PAT: re-issue at https://github.com/settings/personal-access-tokens then overwrite `/etc/ops-phoenix/gh-token`
- Grafana token: re-issue in Grafana → API keys, then overwrite `/etc/ops-phoenix/grafana-token`

After any rotation, trigger one manual run to verify.

See `docs/055-ops-phoenix-autonomous-vm-runner.md` for the full source-of-truth spec.
