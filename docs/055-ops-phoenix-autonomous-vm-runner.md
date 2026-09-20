# Spec 055 — Ops Phoenix Autonomous VM Runner

**Status:** In Progress
**Date:** 2026-06-15
**Owner:** Muthukumaran
**Branch:** TBD
**Component:** `infra/skills/reqsume-ops-phoenix/`
**Related:** spec 040 (`infra-skills`), spec 012 (`production-observability`), spec 047 (`admin-sre-simulation-endpoint`)

---

## Problem

The `reqsume-ops-phoenix` skill (`infra/skills/reqsume-ops-phoenix/`) is a self-healing SRE agent that already does the full cycle end-to-end:

1. Query Loki (via Grafana API) for `error|fatal` level logs from `reqsume-app-web-*` / `reqsume-api-*` on `app1|app2`
2. Analyze with Claude
3. Create a GitHub issue
4. Generate a code fix
5. Open a PR
6. (optionally) merge + trigger the `Deploy to Production New` workflow

It runs locally on a developer's laptop today, driven by `gh auth login` + interactive `input()` prompts in the `setup()` wizard. There's no way to invoke it without a human in front of a terminal, and no way to run it on a cadence.

**Goal:** run this agent fully autonomously, every hour, on the **dev VM** (not production servers — see [Target host](#target-host) below), with **zero interactive prompts**. The agent should:

- Wake up on its own (cron / systemd timer)
- Have credentials pre-loaded (Grafana token, `ANTHROPIC_API_KEY`, `gh` auth)
- Use only non-interactive code paths (no `input()`, no `getpass`, no `os.isatty`-gated branches)
- Auto-decide action level via the existing `threshold_manager.py` (deduplication, severity, rate limiting)
- Be observable (logs on disk, Prometheus-textfile heartbeat for monitoring)
- Be safe to run unattended (idempotent, no destructive operations outside what the agent already does, with kill-switch via env var)

This is a runner / packaging spec, not a rewrite of the agent. The agent's behavior (Loki query, Claude analysis, GitHub issue/PR) is unchanged. We're wiring it to a clock and removing the human from the loop.

---

## Target host

The dev VM (the single Hetzner/Contabo VM currently used for the dev API + dev Postgres on `:5433`, referenced in `infra/postgres-forward` and the existing `task dev:deploy:all`). The agent runs **as a cron job on the dev VM**, not on `app1`/`app2`/`db`. It queries Loki for the **production** Loki stream (`observability.deemwar.com`) regardless of where it runs from — the host only needs outbound HTTPS to Grafana + Anthropic + GitHub.

Why dev VM, not a prod app host:
- It already has `gh` authenticated for the team account used to open PRs.
- It's not in the request path of any user traffic — a hung or buggy agent loop can't degrade prod.
- It has a non-root user (`reqsume`) the team can `ssh` into for triage without a prod escalation.
- It already has `task` + `git` + `python3` installed (see dev setup in `CLAUDE.md`).

---

## What "autonomous" means here

Concretely, every code path that currently calls `input()` or `getpass.getpass()` must be unreachable when the runner launches the agent. The existing scripts already support this for `--action detect` and `--action full-cycle` (they read everything from env vars or `~/.ops-phoenix/config.json`), but two scripts — `scripts/ops-phoenix.py::setup()` and the interactive fallback in `scripts/ops-phoenix.py::run()` — call `input()`. The runner will never invoke these; instead it will:

1. Pre-stage `~/.ops-phoenix/config.json` from a templated block in the runner (one-shot, not via the wizard).
2. Always pass an explicit `--action` flag so the `else: # Interactive mode` branch is unreachable.
3. Refuse to start if any required credential is missing (fail loud at the wrapper level, not via `input()`).

The current `framework/ops_phoenix.py` already has `run_setup_non_interactive()` for exactly this case (it builds config from env vars). The runner will use that path.

---

## Architecture

```
┌──────────────────────────────────────────────────────────────────────┐
│ DEV VM (Linux, systemd) │
│ │
│ ┌─────────────────────────────────────────────────────────────────┐ │
│ │ systemd timer: ops-phoenix-hourly.timer │ │
│ │ OnCalendar=*:0/15 │ │
│ │ Persistent=true │ │
│ └────────────────────────────┬────────────────────────────────────┘ │
│ │ │
│ ▼ │
│ ┌─────────────────────────────────────────────────────────────────┐ │
│ │ systemd service: ops-phoenix-hourly.service (Type=oneshot)  │ │
│ │ ExecStart=/opt/ops-phoenix/bin/ops-phoenix-runner.sh │ │
│ │ EnvironmentFile=/etc/ops-phoenix/ops-phoenix.env │ │
│ │  User=opsphoenix Group=opsphoenix │ │
│ └────────────────────────────┬────────────────────────────────────┘ │
│ │ │
│ ▼ │
│ ┌─────────────────────────────────────────────────────────────────┐ │
│ │ /opt/ops-phoenix/bin/ops-phoenix-runner.sh  │ │
│ │ 1. preflight (jq, curl, gh, python3, disk space)  │ │
│ │ 2. export ANTHROPIC_API_KEY from /etc/ops-phoenix/secret │ │
│ │ 3. lock (flock /var/lock/ops-phoenix.lock) │ │
│ │ 4. python3 framework/ops_phoenix.py --action auto-cycle \ │ │
│ │ --env prod --config /etc/ops-phoenix/config.json │ │
│ │ 5. tail -200 /var/log/ops-phoenix/runs/*.log to journal │ │
│ │ 6. write Prometheus heartbeat to │ │
│ │ /var/lib/node_exporter/textfile_collector/ops_phoenix.prom│ │
│ └─────────────────────────────────────────────────────────────────┘ │
│ │
│ /opt/ops-phoenix/ (skill code, vendored copy from repo) │
│ /etc/ops-phoenix/ (config + env file, mode 0640 root:ops)│
│ /etc/ops-phoenix/secret (ANTHROPIC_API_KEY, mode 0600 ops) │
│ /var/log/ops-phoenix/runs/ (rotated by logrotate, 30d) │
│ /var/lib/ops-phoenix/state/ (mirrored from ~/.ops-phoenix/) │
└──────────────────────────────────────────────────────────────────────┘
 │ │ │
 ▼ ▼  ▼
 observability.deemwar.com api.anthropic.com github.com
 (Loki / Grafana API) (Claude analysis) (issues + PRs)
```

The wrapper script is the only new code. The agent itself is **unmodified** — we're just running it under a non-interactive environment with a clock.

---

## Implementation

### Slice 1 — install + initial dry run (target: 1 PR)

#### 1.1 Vendor the skill into `/opt/ops-phoenix/`

The dev VM is plain Ubuntu/Debian with no symlink to the repo. The skill code must live somewhere stable. Two options:

| Option | Pros | Cons |
|---|---|---|
| **(a) `git clone` into `/opt/ops-phoenix/`** | Easy `git pull` to update; matches how skills already live in the repo | Drift from the in-repo version |
| **(b) Symlink to the repo checkout** | Always in sync with what the team has reviewed | Requires a working tree at a known path; breaks on `git pull --rebase` |

**Decision: (a)**, with a thin `bin/update-skill.sh` wrapper that does `git pull --ff-only` from the dev branch of the repo. The vendored copy is a *runtime snapshot*, not a working tree — agents are not developed on the VM.

```bash
# /opt/ops-phoenix/bin/update-skill.sh
#!/usr/bin/env bash
set -euo pipefail
cd /opt/ops-phoenix
if ! git diff --quiet; then
 echo "Local changes present, refusing to update" >&2
 exit 1
fi
git fetch --depth 1 origin <pinned-ref-or-branch>
git reset --hard origin/<pinned-ref-or-branch>
```

`update-skill.sh` is itself a `Type=oneshot` systemd service (`ops-phoenix-update.service`) triggered by another timer (weekly, with `RandomizedDelaySec=2h`) so we never pull live during a run.

#### 1.2 Create the unprivileged user + dirs

```bash
sudo useradd --system --home /var/lib/ops-phoenix --shell /usr/sbin/nologin opsphoenix
sudo install -d -o opsphoenix -g opsphoenix -m 0750 /opt/ops-phoenix
sudo install -d -o opsphoenix -g opsphoenix -m 0750 /var/lib/ops-phoenix/state
sudo install -d -o opsphoenix -g opsphoenix -m 0750 /var/log/ops-phoenix/runs
sudo install -d -o root  -g opsphoenix -m 0750 /etc/ops-phoenix
```

`/var/lib/ops-phoenix/state` becomes `HOME` for the agent process — `~/.ops-phoenix/` maps to `<state>/.ops-phoenix/` via `HOME=/var/lib/ops-phoenix/state` in the systemd unit. The agent's rate-limit state and history live there.

#### 1.3 Drop credentials into `/etc/ops-phoenix/`

```bash
# /etc/ops-phoenix/secret (mode 0600, owner root:opsphoenix)
ANTHROPIC_API_KEY=sk-ant-...

# /etc/ops-phoenix/gh-token (mode 0600, owner root:opsphoenix, used by gh)
# contents: a GitHub PAT or fine-grained token with `repo` + `workflow` scopes
# OR the output of `gh auth login --with-token` on a teammate's machine
```

The `ops-phoenix-runner.sh` wrapper will:

1. `set -a; . /etc/ops-phoenix/secret; set +a` (loads env vars without echoing them)
2. `export GH_TOKEN=$(cat /etc/ops-phoenix/gh-token)` (so `gh` picks it up without needing an interactive `gh auth login`)

These files are NOT committed. They are created once by hand, by a teammate with `gh auth login` already done locally, and then `scp`'d (or `ansible.builtin.copy` with `mode: 0600`) to the dev VM. The credential file is excluded by `.gitignore` at the repo level (the team already does this for the in-repo `.env` symlink workflow — same model).

The Grafana token the agent reads from disk (`$GRAFANA_TOKEN_PATH`, default `~/Downloads/Archive/keys/grafana-api-token`) is **replaced** by an env var path under the runner. We set `GRAFANA_TOKEN_PATH=/etc/ops-phoenix/grafana-token` and the `scripts/ops-phoenix.py` and `framework/ops_phoenix.py` will read from that path the same way they already read from the default — no code change needed, just an env var.

#### 1.4 The runner wrapper

`/opt/ops-phoenix/bin/ops-phoenix-runner.sh`:

```bash
#!/usr/bin/env bash
# Ops Phoenix — autonomous hourly runner.
# Invoked by systemd; never by hand. Never prompts for input.
set -euo pipefail
IFS=$'\n\t'

umask 0077
export HOME=/var/lib/ops-phoenix/state
export PATH="/opt/ops-phoenix/.venv/bin:/usr/local/bin:/usr/bin:/bin"

# --- 1. Lock so two runs cannot collide (manual run + timer run) ---
exec 9>/var/lock/ops-phoenix.lock
if ! flock -n 9; then
 echo "[$(date -Iseconds)] another run is in progress, exiting" \
 | tee -a /var/log/ops-phoenix/runs/skip.log >&2
 exit 0
fi

# --- 2. Load secrets ---
set -a
# shellcheck disable=SC1091
. /etc/ops-phoenix/secret
set +a
export GH_TOKEN
GH_TOKEN=$(cat /etc/ops-phoenix/gh-token)
export GRAFANA_TOKEN_PATH=/etc/ops-phoenix/grafana-token
export OPS_PHOENIX_CONFIG=/etc/ops-phoenix/config.json

# --- 3. Preflight ---
for bin in jq curl gh python3 git; do
 command -v "$bin" >/dev/null 2>&1 || { echo "missing $bin"; exit 1; }
done
[[ -f "$GRAFANA_TOKEN_PATH" ]] || { echo "grafana token missing"; exit 1; }
[[ -n "${ANTHROPIC_API_KEY:-}" ]] || { echo "ANTHROPIC_API_KEY missing"; exit 1; }
[[ -f "$OPS_PHOENIX_CONFIG" ]] || { echo "config missing"; exit 1; }
# verify gh auth (will be silent because GH_TOKEN is set)
gh auth status >/dev/null 2>&1 || { echo "gh not authenticated"; exit 1; }

# --- 4. Run ---
TS=$(date -u +%Y%m%dT%H%M%SZ)
RUN_LOG=/var/log/ops-phoenix/runs/${TS}.log
mkdir -p "$(dirname "$RUN_LOG")"

cd /opt/ops-phoenix
python3 framework/ops_phoenix.py \
 --action auto-cycle \
 --env prod \
 --config "$OPS_PHOENIX_CONFIG" \
 >"$RUN_LOG" 2>&1
RC=$?

# --- 5. Heartbeat (Prometheus textfile) ---
HB_DIR=/var/lib/node_exporter/textfile_collector
mkdir -p "$HB_DIR"
cat >"$HB_DIR/ops_phoenix.prom" <<EOF
# HELP ops_phoenix_last_run_timestamp_seconds Unix time of the last completed run.
# TYPE ops_phoenix_last_run_timestamp_seconds gauge
ops_phoenix_last_run_timestamp_seconds $(date +%s)
# HELP ops_phoenix_last_run_success 1 if the last run returned exit code 0.
# TYPE ops_phoenix_last_run_success gauge
ops_phoenix_last_run_success $(( RC == 0 ? 1 : 0 ))
# HELP ops_phoenix_run_duration_seconds Duration of the last run.
# TYPE ops_phoenix_run_duration_seconds gauge
ops_phoenix_run_duration_seconds ${SECONDS:-0}
EOF

# --- 6. Surface failure to journald ---
if [[ $RC -ne 0 ]]; then
 echo "ops-phoenix run failed (rc=$RC); tail:" >&2
 tail -50 "$RUN_LOG" >&2 || true
 exit $RC
fi
```

Notes:
- The wrapper, not the agent, owns `flock` and credential loading. The agent code is not changed.
- `set -a; . secret; set +a` keeps the secret out of the process's argv (visible in `/proc/<pid>/environ` but not `/proc/<pid>/cmdline`).
- `exec 9>` keeps the lock FD open for the entire run; `flock` releases it on exit.
- The Prometheus heartbeat reuses whatever path `node_exporter --collector.textfile.directory` is already configured to scrape on this VM (the dev VM already runs node_exporter — see spec 012 for the observability topology). No new monitoring infra.

#### 1.5 `config.json` (non-interactive)

`/etc/ops-phoenix/config.json` — staged by hand on the VM, not via the wizard:

```json
{
 "observability": {
 "type": "grafana_cloud",
 "grafana_url": "https://observability.deemwar.com",
 "grafana_token_path": "/etc/ops-phoenix/grafana-token",
 "container_patterns": ["reqsume-app-web-*", "reqsume-api-*"],
 "host_patterns": ["app1", "app2"],
 "error_patterns": ["level=~\"(?i)error|fatal\""]
 },
 "github": {
 "repo": "deemwar-products/reqsume",
 "base_branch": "main",
  "token_env_var": "GH_TOKEN",
 "default_labels": ["sre-alert", "ops-phoenix"]
 },
 "cicd": {
 "type": "github_actions",
 "workflow_name": "Deploy to Production New"
 },
 "deployment": {
 "auto_merge": true,
 "monitor_timeout_seconds": 600,
 "max_retries": 2
 }
}
```

This is the same shape as `framework/templates/sample_config.json` (see lines 1–29 of that file). The two material changes vs. the sample are:
- `grafana_token_path` points at the runner-managed secret location.
- `github.token_env_var` is `GH_TOKEN` (the runner's name) rather than `GITHUB_TOKEN`, because we deliberately use a fine-grained PAT scoped only to `deemwar-products/reqsume` + `workflow` + `contents`.

#### 1.6 Add `--action auto-cycle` to the agent

The current `framework/ops_phoenix.py::main()` accepts `--action` with values `setup | status | detect | analyze | issue | fix | deploy | full`. None of these match what the runner needs:

| Need | Current closest flag | Why it falls short |
|---|---|---|
| "Decide action level from `threshold_manager.should_create_pr()` and run the appropriate slice" | `--action full` | Always runs the full chain, ignoring rate limits and dedup. Will spam duplicate issues. |
| "Never ask for input" | `--action detect` | Doesn't exist for `framework/ops_phoenix.py` (only `scripts/ops-phoenix.py`) — and the wizard fallback in `framework/ops_phoenix.py::run()` calls `ConfigWizard` on missing config |
| "Pass a config path explicitly" | `--config` is read but not used as a flag — `OpsPhoenix.__init__()` always reads `~/.ops-phoenix/config.json` | Path is hardcoded; no way to point it at `/etc/ops-phoenix/config.json` |

So slice 1 also adds a thin patch to `framework/ops_phoenix.py`:

1. **Add `--config PATH` flag** to `main()` that overrides `CONFIG_FILE` (currently `Path.home() / ".ops-phoenix" / "config.json"`). The wrapper uses this to point at `/etc/ops-phoenix/config.json` without symlink tricks.
2. **Add `--action auto-cycle`** which:
 - Calls `ThresholdManager().should_create_pr(error_types, analysis, recent_issues)` and short-circuits when `should_act` is `False`.
 - On `action_level == "full_cycle"`, runs the full chain (current `--action full` path).
 - On `action_level == "pr_only"`, runs detect → analyze → issue → fix → PR, but stops before the merge.
 - On `action_level == "issue_only"`, runs detect → analyze → issue only.
 - On `action_level == "update_existing"`, runs detect → analyze → comment on the existing issue.
 - On `action_level == "log_only"` or `"ignore"`, exits 0 with a single log line.
3. **Make the wizard unreachable** when invoked from a non-TTY environment. Change `run_setup()` (currently called from `__main__` when no flags are passed) to check `sys.stdin.isatty()` and bail with a clear error if not. The wrapper always passes `--config` and `--action`, so this is a defense-in-depth guard.
4. **Harden `run_setup_non_interactive()`** — currently it only sets keys for the keys it knows about (observability / github / cicd / deployment). It needs to also accept `MONITOR_TIMEOUT`, `MAX_RETRIES`, `AUTO_MERGE` from env (it already does this — good), and to refuse to write a partial config when any required key is missing. Specifically: if `GITHUB_REPO` is unset, log a hard error and exit non-zero; do not write a broken config.

The Python patch is small (≈40 lines added to `framework/ops_phoenix.py`); the rest is unit tests in `framework/tests/test_auto_cycle.py` that drive `--action auto-cycle` against a fixture config and assert:
- The wizard is never instantiated when stdin is closed.
- `auto-cycle` exits 0 with `"IGNORE: count below threshold"` when given 1 error under `min_error_count=3`.
- `auto-cycle` exits 0 with `"LOG ONLY: rate limited"` when fed two runs inside `rate_cooldown_minutes`.
- `auto-cycle` does not invoke `gh issue create` when the dedup match returns a number.

#### 1.7 systemd unit + timer

`/etc/systemd/system/ops-phoenix-hourly.service`:

```ini
[Unit]
Description=Ops Phoenix — autonomous SRE cycle
After=network-online.target
Wants=network-online.target

[Service]
Type=oneshot
User=opsphoenix
Group=opsphoenix
EnvironmentFile=/etc/ops-phoenix/ops-phoenix.env
ExecStart=/opt/ops-phoenix/bin/ops-phoenix-runner.sh
Nice=10
IOSchedulingClass=best-effort
IOSchedulingPriority=7
# Don't let one run eat the whole day
TimeoutStartSec=900
# If something hangs, systemd kills it; the next timer tick still runs
KillMode=mixed
KillSignal=SIGTERM
```

`/etc/systemd/system/ops-phoenix-hourly.timer`:

```ini
[Unit]
Description=Hourly Ops Phoenix cycle
Requires=ops-phoenix-hourly.service

[Timer]
# Every hour on the hour, but offset by 7 minutes to avoid colliding with
# other cron jobs on the VM (kamal access logs, vsync pulls, etc.)
OnCalendar=*:07:00
# If the VM was off at the scheduled time, run on boot
Persistent=true
# Randomize the start by ±90s so multiple dev VMs don't fan out
# (this is a single-VM setup, but the jitter is cheap insurance)
RandomizedDelaySec=90
Unit=ops-phoenix-hourly.service

[Install]
WantedBy=timers.target
```

`/etc/ops-phoenix/ops-phoenix.env`:

```bash
# Sourced by EnvironmentFile. Plain key=value; no secrets here.
OPS_PHOENIX_LOG_RETENTION_DAYS=30
OPS_PHOENIX_DRY_RUN=false
```

`ops-phoenix.env` deliberately holds **only non-secret defaults** — the secrets stay in the `secret` file with `umask 0077` and perms `0600`. systemd `EnvironmentFile` doesn't enforce perms, so a typo in this file that put a key=value would be world-readable.

Enable + start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now ops-phoenix-hourly.timer
systemctl list-timers ops-phoenix-hourly.timer
```

The user is also expected to leave a comment in `/etc/ops-phoenix/README` (a one-time hand-written file) explaining:
- what this service is for,
- where logs go (`/var/log/ops-phoenix/runs/`),
- who to page if it fails (linked from the team's runbook),
- the kill-switch (see below).

#### 1.8 logrotate

`/etc/logrotate.d/ops-phoenix`:

```
/var/log/ops-phoenix/runs/*.log {
 daily
 rotate 30
 compress
 delaycompress
 missingok
 notifempty
 create 0640 opsphoenix opsphoenix
  dateext
 dateformat -%Y%m%d
}
```

30 days of history is enough to debug a "what did the agent do last Friday" question without filling the disk. The Prometheus heartbeat on disk is the durable signal for "did the run happen"; the run logs are the why.

#### 1.9 Manual trigger guard

The wrapper script intentionally lives at `/opt/ops-phoenix/bin/`, which is **not on the PATH** for any interactive user. The only way to start a run is:

- the timer (scheduled), or
- `sudo systemctl start ops-phoenix-hourly.service` (explicit human trigger)

This prevents an accidental `./ops-phoenix-runner.sh` from a teammate's SSH session from racing with the timer — systemd serializes them via the service unit.

---

### Slice 2 — kill-switch + observability dashboard (target: 1 PR)

#### 2.1 Kill-switch env var

In the runner, before invoking python:

```bash
if [[ "${OPS_PHOENIX_DISABLED:-0}" == "1" ]]; then
 echo "OPS_PHOENIX_DISABLED=1 — exiting without running"
 exit 0
fi
```

Plus a sentinel file:

```bash
if [[ -f /etc/ops-phoenix/disabled ]]; then
 echo "disabled sentinel present at /etc/ops-phoenix/disabled"
 exit 0
fi
```

Both let a teammate stop the agent without uninstalling the systemd unit. To re-enable: `rm /etc/ops-phoenix/disabled` or unset the env var and `systemctl daemon-reload`.

#### 2.2 Status command

A read-only command that prints the last 5 runs, last heartbeat, and current rate-limit state:

```bash
sudo /opt/ops-phoenix/bin/ops-phoenix-status.sh
# Last run: 2026-06-15T14:07:23Z status=success duration=87s
# Last run: 2026-06-15T13:07:19Z status=success duration=42s
# Last run: 2026-06-15T12:07:31Z status=failure duration=12s (gh not authenticated)
# Heartbeat: ops_phoenix_last_run_success=1 age=47m
# Rate limit: 0/4 runs in last hour
```

Output goes to stdout; nothing to interpret. Reads from `/var/log/ops-phoenix/runs/` + the prom textfile.

#### 2.3 Dashboard panel

Add a row to the existing **reqsume-prod** dashboard (Grafana, https://observability.deemwar.com/d/reqsume-prod, spec 012) titled "Ops Phoenix (dev VM)":

- Stat: `time() - ops_phoenix_last_run_timestamp_seconds` — turns red if > 2h, yellow if > 1.5h
- Stat: `ops_phoenix_last_run_success`
- Time series: `ops_phoenix_run_duration_seconds` over 7d

Data source: the existing Prometheus instance that already scrapes the dev VM's node_exporter textfile collector.

---

### Slice 3 — guardrails for destructive behavior (target: 1 PR, after we have 2 weeks of slice 1+2 data)

After slice 1+2 has been running for 2 weeks with `OPS_PHOENIX_DRY_RUN=true` (a no-op flag in the wrapper that runs the whole detect → analyze → issue path but skips the PR + merge), review the issue stream:

- Is the agent opening issues for things humans would ignore? → raise the threshold.
- Is the `auto_merge: true` flag appropriate? → switch to `false` and require a teammate to merge the PRs the agent opens, until we trust the fix generator.

This slice is intentionally a stop-gap, not a permanent design. The agent generates a **single-file Markdown stub** today (see `framework/ops_phoenix.py::create_pr()`, lines 399–413 — it writes `ops-phoenix-fix.md` with the analysis summary). That's safe to auto-merge because the fix does not touch any real code path. Once the fix generator learns to write actual Go/TypeScript diffs, this slice gets re-scoped to require human merge approval.

---

## Files this spec creates / modifies

| File | Action | Notes |
|---|---|---|
| `infra/skills/reqsume-ops-phoenix/framework/ops_phoenix.py` | **Modify** | Add `--config`, `--action auto-cycle`, harden `run_setup_non_interactive()`, make wizard TTY-gated. ≈ 40 lines net. |
| `infra/skills/reqsume-ops-phoenix/framework/tests/test_auto_cycle.py` | **New** | Unit tests for the new flag + non-interactive behavior. |
| `infra/skills/reqsume-ops-phoenix/framework/ops_phoenix_main.py` | **Modify** | Pass through `--config` to the underlying agent. |
| `infra/setup/ansible/ops-phoenix.yml` (or a small bash installer in `infra/setup/`) | **New** | Idempotent installer: user, dirs, systemd units, logrotate, sample config (no secrets). Run once per env. |
| `docs/specs/in-progress/055-ops-phoenix-autonomous-vm-runner.md` | **New** | This spec. |
| `/opt/ops-phoenix/` (on dev VM) | **New** | Vendored copy of `infra/skills/reqsume-ops-phoenix/`. |
| `/etc/ops-phoenix/` (on dev VM) | **New** | `config.json`, `ops-phoenix.env`, `README` (no secrets). |
| `/etc/ops-phoenix/secret`, `/etc/ops-phoenix/gh-token`, `/etc/ops-phoenix/grafana-token` (on dev VM) | **New** | Created by hand, never committed. Mode `0600`, owner `root:opsphoenix`. |
| `/etc/systemd/system/ops-phoenix-hourly.{service,timer}` (on dev VM) | **New** | systemd unit + timer. |
| `/etc/logrotate.d/ops-phoenix` (on dev VM) | **New** | 30d rotation. |

`CLAUDE.md` is **not** updated for slice 1+2 because the dev setup section is for *developer* machines, and this is a *server-side* component. Slice 3 will add a one-paragraph "Ops Phoenix (autonomous)" entry to the `Architecture` section pointing at this spec.

---

## VM bootstrap — what's actually been done (2026-06-15)

Status of the dev VM `deemwar-dev` as of this spec update:

**Done (partial bootstrap from a prior session):**
- `opsphoenix` user created (uid 999, `/usr/sbin/nologin`)
- `/opt/ops-phoenix/`, `/var/lib/ops-phoenix/state/`, `/var/log/ops-phoenix/runs/`, `/etc/ops-phoenix/` directories exist
- `/etc/ops-phoenix/anthropic-key` and `/etc/ops-phoenix/grafana-token` are present (mode 0600/0640, owner `root:opsphoenix`)

**Done in `feat/ops-phoenix-vm-installer` (this PR):**
- `infra/setup/ops-phoenix/` — all staged assets: `config.json`, `ops-phoenix.env`, `README`, `ops-phoenix-runner.sh`, `update-skill.sh`, systemd units (4 files), `ops-phoenix-logrotate`
- `infra/setup/install-ops-phoenix.sh` — idempotent installer with `--dry-run` mode; reads `GH_TOKEN_FILE` from env so the GitHub PAT never touches argv or shell history
- The runner wrapper defensively `unset ANTHROPIC_AUTH_TOKEN` to avoid the dual-auth-header problem discovered during local verification (when a calling env has both `ANTHROPIC_API_KEY` and `ANTHROPIC_AUTH_TOKEN` — e.g. when a developer runs the wrapper from a laptop with Claude Code installed — the Anthropic Python SDK sends both `X-Api-Key` and `Authorization: Bearer`, which claudestore.store and other resellers reject)

**Not yet done (waiting on team input):**
- `gh-token` is not in the repo (correctly — it's a secret). The installer pulls it from `GH_TOKEN_FILE` on the host running the installer.
- Admin approval of the fine-grained GitHub PAT (in flight with the org admin)
- node_exporter / Prometheus textfile collector on the dev VM (decision deferred; the wrapper writes to a local-only path `/var/lib/ops-phoenix/heartbeat.prom` which can be scraped later without code changes)
- First live run on the VM (blocked on PAT approval)

**Verification done locally (Mac, before VM deploy):**
- Confirmed `api3.claudestore.store` is reachable and the team's reseller key is valid
- Confirmed `claude-haiku-4-5`, `claude-sonnet-4-6`, `claude-opus-4-8` all work on the reseller
- Confirmed the realistic SRE-triage prompt shape (error logs → JSON with `root_cause` / `severity` / `summary`) round-trips correctly with `max_tokens: 256`
- Confirmed the Anthropic Python SDK reads `ANTHROPIC_BASE_URL` from env and routes correctly when only one of `ANTHROPIC_API_KEY` / `ANTHROPIC_AUTH_TOKEN` is set
- Confirmed the SDK fails with `Invalid key prefix` when both env vars are set — the runner wrapper's `unset ANTHROPIC_AUTH_TOKEN` defends against this

**Decision pending: `auto_merge`.** Per spec open-question #1, ship with `auto_merge: false`. The agent opens issues and PRs; humans merge. The slice 3 review (after 2 weeks of dry-run) is the explicit decision point on whether to flip it.

---

## Acceptance

### Slice 1

1. On the dev VM, `systemctl list-timers ops-phoenix-hourly.timer` shows the next fire time within the next 90 minutes.
2. `sudo systemctl start ops-phoenix-hourly.service` runs end-to-end in under 5 minutes when there are no errors in the production Loki stream. Exit code 0. The Prometheus textfile at `/var/lib/node_exporter/textfile_collector/ops_phoenix.prom` is updated with `ops_phoenix_last_run_success 1` and a fresh timestamp.
3. With `ANTHROPIC_API_KEY` unset in `/etc/ops-phoenix/secret`, the runner exits non-zero with `ANTHROPIC_API_KEY missing` and writes `ops_phoenix_last_run_success 0` to the heartbeat.
4. With `gh` unauthenticated, the runner exits non-zero and logs `gh not authenticated`; the journald entry includes the last 50 lines of the run log.
5. Two simultaneous invocations (`systemctl start` from a terminal + the timer firing) do not collide — the second one exits 0 with `another run is in progress, exiting`, and the first one completes.
6. `framework/tests/test_auto_cycle.py` passes locally: `python3 -m pytest framework/tests/test_auto_cycle.py -v`.
7. `framework/ops_phoenix.py` does not import `getpass` or call `input()` from any code path reachable via `--action auto-cycle --config /path/to/cfg`. Verified by `grep -nE "input\(|getpass" framework/ops_phoenix.py` returning only matches in the `setup()` function and the new TTY guard.

### Slice 2

8. `OPS_PHOENIX_DISABLED=1` in `ops-phoenix.env` causes the next timer fire to exit 0 without invoking python. Verified by `journalctl -u ops-phoenix-hourly.service -n 5`.
9. `ops-phoenix-status.sh` shows the last 5 runs, regardless of log rotation.
10. The new dashboard panel renders in the reqsume-prod dashboard and turns red if no run has happened in 2h.

### Slice 3 (deferred)

11. After 2 weeks of `OPS_PHOENIX_DRY_RUN=true`, the team reviews the issues opened and the false-positive rate, then decides on the auto-merge policy for slice 3. **No code lands in slice 3 until that review is done.**

---

## Security

- **Credential files** (`secret`, `gh-token`, `grafana-token`) are mode `0600`, owner `root:opsphoenix`, and never logged. The wrapper uses `set -a; . secret; set +a` to load them into the env without echoing, and the runner script's own log lines never reference these paths.
- **No `getpass` / `input()`** reachable from the runner's invocation path (slice 1 acceptance #7).
- **No tokens in process argv** — `GH_TOKEN` is read from a file and exported, not passed as a CLI flag.
- **The agent user `opsphoenix` has no password, no shell (`/usr/sbin/nologin`)**, no home directory other than `/var/lib/ops-phoenix/state`, and no `sudo` rights. A compromised agent can only read its own state dir and the credential files (which it can already read by design).
- **GitHub token scope** is the minimum required: `repo` (to open PRs / read code) + `workflow` (to trigger `Deploy to Production New`). No `admin:org`, no `delete_repo`, no package scopes. A teammate can revoke this single token in isolation.
- **The `gh` auth refresh** lives on the VM; rotating the token means a 1-line edit + `systemctl restart` — no rebuild, no deploy.
- **Rate limiting** is already implemented in `threshold_manager.py::check_rate_limit()` (max 4 runs/hour, 15-min cooldown). The wrapper's `flock` provides a second layer — even if rate limiting is bypassed, only one process can run at a time.
- **Deduplication** is already implemented in `threshold_manager.py::check_deduplication()` (48-hour window). The `--action auto-cycle` flag respects it (acceptance test: an issue opened Monday is *not* re-opened by the Wednesday run if the same error recurs).
- **The agent never `rm -rf`s, never `kubectl delete`s, never `kamal rollback`s.** Its destructive surface is limited to `gh issue create`, `gh pr create`, `gh pr merge --squash --delete-branch` (only on `full_cycle` action level, and only with `auto_merge: true` in config), and `gh workflow run`. All four are visible in the team's GitHub audit log.

---

## Failure modes

| Symptom | Likely cause | Recovery |
|---|---|---|
| Timer fires, but service exits 1 with `gh not authenticated` | `GH_TOKEN` expired or `gh-token` file got truncated | Re-run `gh auth login --with-token` on a teammate's laptop, `scp` the new token, `systemctl start ops-phoenix-hourly.service` to retry. |
| Heartbeat shows `ops_phoenix_last_run_success 0` for two consecutive runs | Loki API changed shape, or the agent's parser broke on a new log format | Read the last 50 lines of `/var/log/ops-phoenix/runs/<ts>.log`. If it's a parser issue, patch `framework/adapters/observability.py` and re-run `update-skill.sh`. |
| `another run is in progress, exiting` in journal for > 1h | A previous run wedged — the agent is mid-`monitor_workflow()` (15s sleep loop in `scripts/check-errors.py::monitor_workflow()`, max 5 min) | Wait for the 15-min `TimeoutStartSec` to kill it. If recurring, the agent is stuck in an infinite Claude API call — check the `ANTHROPIC_API_KEY` quota. |
| Disk fills up because `/var/log/ops-phoenix/runs/` is huge | An error spike is generating 1k+ error log lines per run, each run log is 200KB+ | Lower `maxLines: 100` in the config to `maxLines: 50`. logrotate already prunes to 30d, so this self-heals within a day. |
| Agent opens 50 issues in a night | Threshold config is too loose, or a new class of error is bypassing `ignore_error_patterns` | Set `OPS_PHOENIX_DISABLED=1` to stop the bleeding, then adjust `threshold_manager.py::ThresholdConfig` (`min_error_count` / `ignore_error_patterns`) and `systemctl daemon-reload`. |

---

## Open questions

1. **Should the agent ever `auto_merge`?** The current `framework/templates/sample_config.json` has `"auto_merge": true`, but that's a placeholder. Slice 1 ships with `auto_merge: false` — the agent opens PRs and the team merges. Slice 3 is the explicit decision point on flipping it to `true`. Marked in the spec as deferred.
2. **Notification on failure?** A failed run shows up in the Prometheus heartbeat, but nobody is paged. The team's existing alerting path is **Telegram** (see `infra/telegram/`), and the existing `notify.Send` (spec 015) and `notify.SendToAllAdmins` (apps/api) infra can deliver a Slack/Telegram message — but only from the API process, not from a server-side script. Slice 2 ships with the dashboard panel only; a Telegram alert on `ops_phoenix_last_run_success == 0` for >2h is a follow-up. Tracked separately.
3. **Multi-environment support?** The team has a dev, staging, and prod Grafana. The current skill hardcodes the prod dashboard URL (`https://observability.deemwar.com/d/reqsume-logs`). The runner config above targets prod; a future "ops-phoenix-staging" service is a copy-paste away but not in this spec.
4. **What happens when the skill code in the repo moves?** The `update-skill.sh` wrapper does `git pull --ff-only` from a pinned ref. If the team refactors `framework/ops_phoenix.py` to break the `--action auto-cycle` interface, the next weekly update will surface that immediately (the next run will fail with `unrecognized argument: auto-cycle`). A 5-minute response window is acceptable.
5. **Should we add a `--dry-run` flag to the wrapper itself?** Already supported by `OPS_PHOENIX_DRY_RUN=true` in the env file. The wrapper translates that to `--action detect` (no issue, no PR, no merge). Verified in slice 1 acceptance #6.

---

## Out of scope

- A different agent runtime (k8s CronJob, AWS EventBridge, GitHub Actions schedule). The dev VM is fine; if/when the team migrates to k8s, this spec is revisited.
- Replacing the agent's `framework/ops_phoenix.py` with a Claude Code sub-process invocation. The skill is intentionally a stand-alone Python tool so the VM doesn't need Claude Code installed.
- LLM-based deduplication beyond the title-substring match in `threshold_manager.py::check_deduplication()`. Today's approach (substring match on the first error logger name) is good enough for the current issue volume; revisit if the issue count grows.
- Per-incident summaries in the team's existing `notify` system. The agent already creates GitHub issues; that's the audit trail.
