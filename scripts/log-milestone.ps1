# log-milestone.ps1 — append a Malaysia-time entry to MILESTONES.md,
# bump the patch version in version/info.go, and create a single git commit.
#
# Usage:
#   pwsh scripts/log-milestone.ps1
#   pwsh scripts/log-milestone.ps1 -Note "kickoff complete"
#   pwsh scripts/log-milestone.ps1 -Event "start" -Note "kickoff"
param(
  [string]$Event = "run",
  [string]$Note  = "app run logged"
)
$ErrorActionPreference = "Stop"

$repoRoot = (& git rev-parse --show-toplevel).Trim()
Set-Location $repoRoot

$milestones  = "MILESTONES.md"
$versionFile = "version/info.go"

if (-not (Test-Path $milestones))  { throw "MILESTONES.md not found at repo root" }
if (-not (Test-Path $versionFile)) { throw "version/info.go not found" }

# Malaysia time, format: dd-MMM-YYYY hh:mm tt
$tz = [System.TimeZoneInfo]::FindSystemTimeZoneById(
  ($IsWindows ? "Singapore Standard Time" : "Asia/Kuala_Lumpur"))
$now = [System.TimeZoneInfo]::ConvertTimeFromUtc([DateTime]::UtcNow, $tz)
$ts  = $now.ToString("dd-MMM-yyyy hh:mm tt", [Globalization.CultureInfo]::InvariantCulture)
$entry = "- $Event $ts — $Note"

# Ensure trailing newline before append.
$content = Get-Content $milestones -Raw
if ($content.Length -gt 0 -and -not $content.EndsWith("`n")) {
  Add-Content -Path $milestones -Value ""
}
Add-Content -Path $milestones -Value $entry

# Bump patch version.
$verContent = Get-Content $versionFile -Raw
if ($verContent -notmatch 'v(\d+)\.(\d+)\.(\d+)') {
  throw "could not parse current version from $versionFile"
}
$current = $Matches[0]
$new = "v$($Matches[1]).$($Matches[2]).$([int]$Matches[3] + 1)"
(Get-Content $versionFile -Raw).Replace($current, $new) |
  Set-Content -NoNewline $versionFile

# Commit both files together.
& git add $milestones $versionFile | Out-Null
& git diff --cached --quiet
if ($LASTEXITCODE -eq 0) { Write-Host "nothing to commit"; exit 0 }
& git commit -m "chore(milestone): $Event $ts — $Note ($new)"

Write-Host "✓ logged: $entry"
Write-Host "✓ version: $current → $new"
Write-Host "✓ committed on $((& git rev-parse --abbrev-ref HEAD).Trim())"