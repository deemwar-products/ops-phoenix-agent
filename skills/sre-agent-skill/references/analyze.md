# analyze

Feed detected errors to the AI model for root-cause analysis and fix proposals.

## Prerequisites

Must have run `sre-agent detect` first in the session (or confirmed errors
exist via `--time-window`). Never run `analyze` with zero errors.

## Command

```bash
sre-agent analyze --config <config-dir> [--context <level>]
```

## Flags

| Flag | Default | Description |
|---|---|---|
| `--config` | `~/.config/sre-agent` | Config directory |
| `--context` | `medium` | Code context depth: `minimal`, `medium`, `full` |
| `--format` | `text` | Output format: `text` or `json` |
| `--verbose` | `false` | Enable debug logging |

## Output (text format)

```
analyze: found N unique error(s) (window=<duration>)
analyze: querying anthropic (<model>, context=<level>)...

Summary: <one-paragraph summary>

[1] <severity> (severity=<P0-P3> confidence=<0.00-1.00>)
    Error:   <error log line, truncated to 100 chars>
    Cause:   <root cause, truncated to 200 chars>
    Service: <affected service + handler>
    Fix:     <suggested fix, truncated to 300 chars>
```

## Severity Guide

| Level | Meaning |
|---|---|
| P0 | Service down or data loss — immediate action |
| P1 | Feature broken for users — urgent fix |
| P2 | Degraded experience — fix soon |
| P3 | Low impact (noise, probes, minor) — backlog |

Confidence is the AI's self-assessed certainty (0.00–1.00). Below 0.70,
surface the finding but flag it as "low confidence — verify manually."

## Recipe

1. Run `sre-agent analyze --config <config> --context medium`.
2. Parse findings. Group duplicates (same root cause across multiple error
   lines) — the AI often surfaces the same issue multiple times.
3. Summarize: top finding first, then group by severity.
4. Ask whether to proceed to `fix`.
