---
name: sre-agent-skill
description: "Infrastructure observability and auto-remediation via sre-agent CLI. Trigger: what's broken, check errors, analyze errors, auto-fix, morning SRE check, sre status, infrastructure health, production issues, run diagnostics, error dashboard, fix the deploy, remediation."
---

<!-- version: 0.1.0 -->

# sre-agent-skill

A natural-language front door for infrastructure observability and
auto-remediation. Routes SRE intents to [`sre-agent`](https://github.com/deemwar-products/ops-phoenix-agent)
CLI commands (detect, analyze, fix, status) which query Loki/Grafana logs,
run AI analysis via Anthropic, and optionally generate fix PRs.

The skill is the **routing + recipe layer**. It knows which intent maps to
which sre-agent command, which flags to pass, and how to interpret the
results. It does **not** handle credential storage — that belongs to
sre-agent's own keychain integration and environment files.

## Core Rules

- **sre-agent is the engine. This skill never bypasses it.** All log queries,
  AI analysis, and fix generation go through `sre-agent detect/analyze/fix`.
  Never craft manual Loki queries or call Grafana/Anthropic APIs directly.
- **Run commands in sequence.** `detect` → `analyze` → `fix` is the standard
  pipeline. Only skip steps when the prior step found nothing. Never run
  `analyze` or `fix` without first running `detect` in the same session.
- **Respect the config.** The skill reads `~/.config/sre-agent/config.yaml`
  for container patterns, time window, GitHub repo, and notification
  settings. Never hardcode these — pass `--config` and let sre-agent resolve.
- **Honor dry-run / dev mode.** When `mode: dev` is set in config, sre-agent
  operates without side effects (no PR creation). Report what *would* happen
  rather than what happened.
- **Surface walls honestly.** If sre-agent reports "no errors found", "tests
  failed", or "clone failed", report the wall verbatim. Do not retry,
  do not substitute a manual workaround — the user needs to know the actual
  blocker.
- **Default to signal, not noise.** A generic "check for errors" means the
  user wants the actionable findings, not a raw log dump. Summarize by
  severity (P0/P1/P2/P3), affected service, and root cause. Show the top 5
  by default; offer to expand on request.
- **Never print secrets.** sre-agent stores credentials in the OS keychain
  (macOS) or `.env` file (Linux). Never echo credential values, never show
  `.env` contents, never log token strings.
- **One config, multiple environments.** The config supports `dev` and `prod`
  via separate config files or flags.  Never assume — confirm if
  ambiguous.

## Session Context

Held in conversation memory only — no file writes.

```
config_path:    ~/.config/sre-agent          # default, overridable
environment:    null                          # null = default config, "dev" or "prod" when specified
last_detect:    null                          # result summary from last detect run
last_analyze:   null                          # result summary from last analyze run
```

## Process

1. Read `references/workflow.md` for preflight, intent routing, and
   environment handling.
2. For specific command recipes, load the matching family file:
   - detect → `references/detect.md`
   - analyze → `references/analyze.md`
   - fix → `references/fix.md`
   - status → `references/status.md`
   - auth → `references/auth.md`
3. For setup and configuration, see `references/setup.md`.
4. For error classification and severity interpretation, see
   `references/error-classification.md`.

## Families at a glance

| Intent | Command | Reference |
|---|---|---|
| "what's broken", "check for errors", "are there issues" | `sre-agent detect` | `references/detect.md` |
| "analyze the errors", "root cause", "why is it failing" | `sre-agent analyze` | `references/analyze.md` |
| "fix it", "auto-fix", "generate a patch" | `sre-agent fix` | `references/fix.md` |
| "show status", "recent runs", "sre health" | `sre-agent status` | `references/status.md` |
| "set up credentials", "configure grafana token" | `sre-agent auth` | `references/auth.md` |
| "install sre-agent", "configure sre-agent" | setup guide | `references/setup.md` |

## Self-check

Before trusting the skill's routing, validate its format integrity:

    sh tests/all.sh

Runs:
  - tests/validate-recipes.sh     recipe format + cross-references
  - tests/install-test.sh          install/uninstall idempotency
