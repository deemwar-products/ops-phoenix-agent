# Step 00: Preflight

Validate configuration and test all connections before starting a run.

## Load Config

Read `{config-dir}/config.json`. If it doesn't exist, tell the user to run `sre-agent init` first.

Required config keys:
- `observability.type` — one of `grafana_cloud`, `grafana_self_hosted`, `loki_direct`
- `observability.grafana_url` — base URL of the Grafana instance (for grafana adapters)
- `github.repo` — target repo in `owner/name` format
- `ai.api_key_env` — env var name for the AI API key (e.g., `ANTHROPIC_API_KEY`)

## Test Connections

Run these checks and report results:

1. **AI API key**: Check that `os.environ[ai.api_key_env]` is set. If missing, tell the user to export it.
2. **Grafana token**: Read `observability.grafana_token_path`. Verify the file exists and is readable.
3. **GitHub auth**: Run `gh auth status`. If it fails, tell the user to run `gh auth login`.
4. **GitHub repo access**: Run `gh repo view {github.repo}`. Verify access.

## Report

Output a summary like:
```
Preflight: OK
  AI: connected (model: claude-sonnet-4-20250514)
  Grafana: connected (grafana_self_hosted)
  GitHub: connected (owner/repo)
```

If any check fails, stop and tell the user what's missing. Do not proceed with the run.
