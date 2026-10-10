# status

Show sre-agent health, credential status, and recent run history.

## Command

```bash
sre-agent status --config <config-dir>
```

## Output

```
status: backend reachable / unreachable

sre-agent status
────────────────
Observability:  grafana_self_hosted @ https://...
GitHub:         <repo> (base: main)
AI:             <provider> / <model>
Mode:           dev (dry_run=false) / prod

Credentials:
  github       ✓ gh_cli / ✗ missing
  anthropic    ✓ keychain / ✗ missing
  grafana      ✓ keychain / ✗ missing
  teams        ✓ keychain / ✗ missing

Recent runs:
  ✓ analyze    success 3m ago  N findings
  ✗ detect     partial 3m ago  N error(s) in <window>
  ✗ fix        failed  3m ago  <last finding summary>
```

## Interpretation

- `status: backend reachable` — Grafana/Loki is reachable and auth works.
- `status: backend unreachable` — Grafana URL is wrong, token is invalid, or
  network issue. Check `observability.url` in config and `GRAFANA_API_TOKEN`.
- Credential `✓` = resolved (from env, keychain, or gh CLI). `✗` = missing
  — the corresponding commands will fail until configured.
- `partial` for detect means errors were found but a subsequent step failed.
- `failed` with a finding summary means the AI analysis completed but the
  fix step failed.

## Recipe

1. Run `sre-agent status --config <config>`.
2. Report the health summary: backend status, credential status, recent runs.
3. If any credential is `✗`, offer to run `sre-agent auth setup`.
