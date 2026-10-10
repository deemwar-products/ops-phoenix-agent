[CmdletBinding()]
param()

$ErrorActionPreference = "Stop"
$SkillName = "sre-agent-skill"
foreach ($target in @(
    (Join-Path $HOME ".claude\skills\$SkillName"),
    (Join-Path $HOME ".agents\skills\$SkillName")
)) {
    if (Test-Path $target) { Remove-Item -Recurse -Force $target; Write-Host "Removed $target" }
}
Write-Host "Done."
