# fix

Generate a fix for the top error finding and create a PR against the configured
GitHub repo.

## Prerequisites

1. Must have run `sre-agent analyze` first (errors identified).
2. `github.repo` must be set in config.
3. GitHub auth must be available (`gh` CLI logged in, or `GITHUB_TOKEN` env).

## Command

```bash
sre-agent fix --config <config-dir> [--dry-run]
```

## Flags

| Flag | Default | Description |
|---|---|---|
| `--config` | `~/.config/sre-agent` | Config directory |
| `--dry-run` | `false` | Generate diff but don't push or create PR |
| `--verbose` | `false` | Enable debug logging |

## Output

```
fix: analyzing N error(s)...
fix: top finding — <summary> (P<level>, confidence=<X.XX>)
fix: cloning <repo>...
fix: workdir=/tmp/sre-agent-<id>
fix: applying known fix for <pattern>...
fix: diff generated (<N> bytes)
fix: running tests...
fix: tests passed / failed: PASS / FAIL
fix: creating PR...
fix: PR created: https://github.com/<repo>/pull/<N>
```

## Failure Modes

| Step | Failure | Reason | Action |
|---|---|---|---|
| clone | `exit status 128` | No GitHub SSH key or no access | User must set up GitHub auth |
| tests | `tests failed: FAIL` | Fix breaks existing tests | Do not retry — report to user |
| PR | 403/404 | No write access to repo | User must grant permissions |

## Recipe

1. Run `sre-agent fix --config <config>`.
2. Parse output to determine success/failure at each step.
3. If tests failed → report the failure, show the diff if available, do NOT
   create a PR.
4. If PR created → report the PR URL.
5. If clone failed → report the GitHub auth requirement.

## Dev Mode

When `mode: dev` is set in config, sre-agent runs in dry-run mode. No PR is
created, no code is pushed. Report what *would* happen.
