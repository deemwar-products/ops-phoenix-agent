[CmdletBinding()]
param(
    [switch]$Force,
    [switch]$SkipCodex
)

$ErrorActionPreference = "Stop"
$SkillName = "sre-agent-skill"
$Source = (Resolve-Path (Join-Path $PSScriptRoot ".")).Path

function Install-Skill([string]$TargetRoot, [string]$Label) {
    $target = Join-Path $TargetRoot $SkillName
    New-Item -ItemType Directory -Force -Path $TargetRoot | Out-Null
    if (Test-Path $target) {
        if (-not $Force) { throw "$Label destination already exists: $target. Re-run with -Force to update it." }
        Remove-Item -Recurse -Force $target
    }
    Copy-Item -Path $Source -Destination $target -Recurse -Force
    Write-Host "  $Label -> $target"
}

Write-Host "Installing $SkillName..."
if (-not (Get-Command sre-agent -ErrorAction SilentlyContinue)) {
    Write-Warning "'sre-agent' is not on PATH. Install with: go install github.com/deemwar-products/ops-phoenix-agent/sre-agent/cmd/sre-agent@latest"
}
Install-Skill (Join-Path $HOME ".claude\skills") "Claude Code"
if (-not $SkipCodex -and (Get-Command codex -ErrorAction SilentlyContinue)) {
    Install-Skill (Join-Path $HOME ".agents\skills") "Codex Agent"
}
Write-Host "Done. Restart Claude Code or Codex to load the skill."
