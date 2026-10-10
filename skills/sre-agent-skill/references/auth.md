# auth

Manage sre-agent credentials (Grafana, Anthropic, GitHub, Teams).

## Commands

```bash
# Check credential status
sre-agent auth status

# Set a credential (stores in OS keychain on macOS, .env on Linux)
sre-agent auth set grafana_token <token>
sre-agent auth set ANTHROPIC_API_KEY <key>
sre-agent auth set github_token <token>

# Interactive setup wizard
sre-agent auth setup
```

## Credential Resolution Order

1. Environment variables (`GRAFANA_API_TOKEN`, `ANTHROPIC_API_KEY`, etc.)
2. macOS keychain (via `security` CLI)
3. Linux keyring (falls back to noop without secret service)
4. `.env` file at `<config-dir>/.env` (sourced by `sre-agent-run`)

## Recipe

When the user asks to configure credentials:

1. Run `sre-agent auth status` to show current state.
2. For missing credentials, guide through `sre-agent auth set <name> <value>`.
3. Grafana: needs a Grafana API key with `Admin` or `Editor` role from
   Grafana → Connections → API Keys.
4. Anthropic: needs an API key from console.anthropic.com.
5. GitHub: needs a PAT with `repo` scope, or `gh auth login` must be run.
6. Teams: needs a Teams webhook URL from Power Automate or Teams Incoming
   Webhook connector.

Never echo credential values back to the user.
