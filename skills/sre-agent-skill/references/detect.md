# detect

Query Loki/Grafana for errors across configured container patterns.

## Command

```bash
sre-agent detect --config <config-dir> --time-window <duration>
```

## Flags

| Flag | Default | Description |
|---|---|---|
| `--config` | `~/.config/sre-agent` | Config directory |
| `--time-window` | from config (`24h`) | How far back to query (`1h`, `6h`, `24h`, `7d`) |
| `--verbose` | `false` | Enable debug logging |

## Output

```
Found N unique error(s) in the last <window>

  [<container>] (x<count>) <timestamp> <log-line-preview>
    first: <time>  last: <time>

notify: teams ✓ posted
```

## Recipe

1. Run `sre-agent detect --config <config>` with the user's configured time
   window. If the user specifies a shorter window (e.g. "last hour"), pass
   `--time-window 1h`.
2. Parse the output. Each `[container]` block is one deduplicated error group.
3. If `Found 0 unique error(s)` → report "no errors found in the last <window>"
   and stop. Do not run analyze or fix.
4. If errors found → summarize by severity and affected service, then ask
   whether to proceed to analysis.
5. Teams notification is automatic when `notify.teams.enabled: true` in config.
   Report the `notify: teams ✓ posted` line if present.

## Error Interpretation

- `first: 00:00:00  last: 00:00:00` means the error appeared once (no
  repeated occurrences in the window).
- Multiple `(xN)` on the same container = same error repeated N times.
- Container names come from `container_patterns` in config — they may be
  long Docker container IDs if labels aren't set.
