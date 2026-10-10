# SRE Agent installation

This guide installs the `sre-agent` CLI and the optional Claude Code/Codex
skill. The CLI can be used by itself; the skill translates natural-language
requests into CLI commands.

## Prerequisites

Required: Go 1.24+, Git, access to Grafana/Loki, and Claude Code or Codex if
you want the agent skill. Anthropic, GitHub CLI, and Teams credentials are
optional for AI analysis, pull requests, and notifications.

## macOS

```bash
brew install go git
go install github.com/deemwar-products/ops-phoenix-agent/sre-agent/cmd/sre-agent@latest
git clone https://github.com/deemwar-products/ops-phoenix-agent.git
cd ops-phoenix-agent
sh skills/sre-agent-skill/install.sh
sre-agent setup
sre-agent auth set grafana_token '<grafana-token>'
sre-agent auth set ANTHROPIC_API_KEY '<anthropic-key>'
sre-agent detect --time-window 1h
```

Credentials are stored in the macOS Keychain. Never commit tokens.

## Windows (PowerShell)

```powershell
winget install GoLang.Go
winget install Git.Git
```

Open a new PowerShell window, then run:

```powershell
go install github.com/deemwar-products/ops-phoenix-agent/sre-agent/cmd/sre-agent@latest
git clone https://github.com/deemwar-products/ops-phoenix-agent.git
Set-Location ops-phoenix-agent
& .\skills\sre-agent-skill\install.ps1
sre-agent setup
```

If PowerShell blocks local scripts:

```powershell
Set-ExecutionPolicy -Scope CurrentUser RemoteSigned
```

Update an existing skill installation with:

```powershell
& .\skills\sre-agent-skill\install.ps1 -Force
```

The current credential store is macOS-Keychain-backed, so Windows users should
use environment variables:

```powershell
$env:GRAFANA_API_TOKEN = '<grafana-token>'
$env:ANTHROPIC_API_KEY = '<anthropic-key>'
```

The default Windows config is `%USERPROFILE%\.config\sre-agent\config.yaml`.
Verify the installation:

```powershell
sre-agent --help
sre-agent detect --time-window 1h
```

## Separate environments

Use separate config directories for development and production:

```bash
sre-agent --config ~/.config/sre-agent-prod detect --time-window 1h
```

```powershell
sre-agent --config "$HOME\.config\sre-agent-prod" detect --time-window 1h
```

The setup wizard provides defaults, but replace placeholder Grafana URLs,
container patterns, repositories, and branches with real values. The default
detection window is `1h` unless overridden with `--time-window`.

## Uninstalling the skill

```bash
sh skills/sre-agent-skill/uninstall.sh
```

```powershell
& .\skills\sre-agent-skill\uninstall.ps1
```

Restart Claude Code or Codex after installation or removal.
