# workflow

Preflight, handle selection, intent routing, and failure handling for the
sre-agent skill.

## Preflight

Before any sre-agent command:

1. Confirm `sre-agent` is on PATH. If not:
   ```
   go install github.com/deemwar-products/ops-phoenix-agent/sre-agent/cmd/sre-agent@latest
   ```
2. Confirm `~/.config/sre-agent/config.yaml` exists. If not, offer to run
   `sre-agent setup`.
3. Confirm credentials are configured: `sre-agent auth status`. If any are
   missing, offer to set them up.

## Environment Handling

The user may have multiple environments (dev, prod). The skill must determine
which config to use:

- **Explicit mention**: "check production errors" → use `--config
  ~/.config/sre-agent-prod` or the prod config path.
- **Ambiguous**: "check for errors" → use the default config
  (`~/.config/sre-agent`). Ask if the user meant a different environment only
  if `mode` or environment-specific indicators are present in config.
- **Config path**: always pass `--config <path>` explicitly. Never rely on the
  default.

## Intent Routing

| Intent | Command | Flags | Follow-up |
|---|---|---|---|
| "what's broken", "check errors", "any issues" | `detect` | `--time-window <user-specified>` | If errors found → offer analyze |
| "analyze", "root cause", "why is it failing" | `analyze` | `--context <user-specified>` | If findings → offer fix |
| "fix it", "auto-fix", "generate patch" | `fix` | `--dry-run` (ask first) | Report PR or diff |
| "sre status", "show runs", "health" | `status` | none | Report summary |
| "set up", "configure", "auth" | `auth` | `status` / `setup` / `set` | Guide through setup |
| "morning check", "morning SRE" | detect → analyze → fix | default window | Full pipeline |

## Failure Handling

| Failure | Detection | Action |
|---|---|---|
| Grafana unreachable | `status: backend unreachable` | Report URL/token issue, stop |
| No errors found | `Found 0 unique error(s)` | Stop, report clean |
| Tests failed | `tests failed: FAIL` | Do not retry, show diff, report to user |
| Clone failed | `clone failed: exit status 128` | Report GitHub auth requirement |
| AI key invalid | `Anthropic API 401` | Report credential issue, stop |
| Partial detect | `partial` in status | Last step failed, report what ran |

## Output Format

Always summarize findings before acting:

```
sre-agent detect (24h window):
  Found 2 unique errors in reqsume-app-web
    [1] P0 — Gemini API 402 RESOURCE_EXHAUSTED (confidence 0.97)
    [2] P3 — nginx path traversal probe (confidence 0.90)

Proceed to analysis? (yes/no)
```

Group duplicates. The AI often returns the same root cause for multiple log
lines — collapse to a single finding with a count.
