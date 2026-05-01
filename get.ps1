<#
.SYNOPSIS
  Smart installer for movie-cli — tries GitHub Release first, then falls back to
  source-build from main, with a clear message either way.

.DESCRIPTION
  Use this when you don't care HOW movie gets installed — only that it does.

  Resolution order:
    1. GitHub Release (releases/latest/download/install.ps1)
       → installs a pre-built binary, no Go toolchain needed.
    2. Source-build fallback (raw.githubusercontent.com/.../main/install.ps1)
       → clones the repo and builds locally; requires Git + Go 1.22+.

  If neither path works, prints actionable next steps for the developer.

.NOTES
  Invoked via:
    irm https://raw.githubusercontent.com/alimtvnetwork/movie-cli-v7/main/get.ps1 | iex
#>

[CmdletBinding()]
param(
    [string]$DeployPath = ""
)

$ErrorActionPreference = 'Stop'

# ── Config ────────────────────────────────────────────────────
$Owner       = 'alimtvnetwork'
$Repo        = 'movie-cli-v7'
$Branch      = 'main'
$ReleaseUrl  = "https://github.com/$Owner/$Repo/releases/latest/download/install.ps1"
$SourceUrl   = "https://raw.githubusercontent.com/$Owner/$Repo/$Branch/install.ps1"
$ReleasesUI  = "https://github.com/$Owner/$Repo/releases"
$ProbeTimeout = 10

# ── UI helpers ────────────────────────────────────────────────
function Write-Step  { param([string]$M) Write-Host "  -> " -ForegroundColor Cyan -NoNewline; Write-Host $M -ForegroundColor Gray }

# Decode IWR .Content safely. GitHub raw sometimes returns a byte[] (when the
# content-type isn't text/*), and Invoke-Expression can't accept byte[]. Always
# coerce to a UTF-8 string before piping into iex / ScriptBlock::Create.
function ConvertTo-ScriptText {
    param($Response)
    $c = $Response.Content
    if ($null -eq $c) { return "" }
    if ($c -is [string]) { return $c }
    if ($c -is [byte[]]) {
        # Strip UTF-8 BOM if present.
        if ($c.Length -ge 3 -and $c[0] -eq 0xEF -and $c[1] -eq 0xBB -and $c[2] -eq 0xBF) {
            return [System.Text.Encoding]::UTF8.GetString($c, 3, $c.Length - 3)
        }
        return [System.Text.Encoding]::UTF8.GetString($c)
    }
    return [string]$c
}
function Write-Ok    { param([string]$M) Write-Host "  OK " -ForegroundColor Green -NoNewline; Write-Host $M -ForegroundColor Green }
function Write-Warn  { param([string]$M) Write-Host "  !! " -ForegroundColor Yellow -NoNewline; Write-Host $M -ForegroundColor Yellow }
function Write-Note  { param([string]$M) Write-Host "     " -NoNewline; Write-Host $M -ForegroundColor DarkGray }

Write-Host ""
Write-Host " +======================================+" -ForegroundColor DarkCyan
Write-Host " | movie smart installer                |" -ForegroundColor Cyan
Write-Host " +======================================+" -ForegroundColor DarkCyan
Write-Host ""

# ── 1. Probe the GitHub Release asset ─────────────────────────
Write-Step "Checking for a published GitHub Release..."

$releaseAvailable = $false
try {
    # Use HEAD via Invoke-WebRequest -Method Head; -MaximumRedirection follows
    # the /releases/latest/ redirect transparently.
    $resp = Invoke-WebRequest -Uri $ReleaseUrl `
                              -Method Head `
                              -UseBasicParsing `
                              -TimeoutSec $ProbeTimeout `
                              -MaximumRedirection 5 `
                              -ErrorAction Stop
    if ($resp.StatusCode -eq 200) {
        $releaseAvailable = $true
    }
} catch {
    # 404, DNS, timeout — anything counts as "no release available"
    $releaseAvailable = $false
}

if ($releaseAvailable) {
    Write-Ok "Release found — installing pre-built binary"
    Write-Note "Source: $ReleaseUrl"
    Write-Host ""
    $installScript = Invoke-WebRequest -Uri $ReleaseUrl -UseBasicParsing -TimeoutSec 30
    Invoke-Expression (ConvertTo-ScriptText $installScript)
    exit $LASTEXITCODE
}

# ── 2. Fall back to source-build ──────────────────────────────
Write-Warn "No published GitHub Release found for $Owner/$Repo."
Write-Note "Falling back to source-build from branch $Branch."
Write-Note "(This needs Git + Go 1.22+ on PATH. Build takes ~30s.)"
Write-Host ""
Write-Note "Tip for maintainers: publish a release at $ReleasesUI"
Write-Note "                    so future installs grab a pre-built binary"
Write-Note "                    instead of cloning + compiling locally."
Write-Host ""
Write-Step "Downloading source installer: $SourceUrl"

try {
    $sourceScript = Invoke-WebRequest -Uri $SourceUrl `
                                      -UseBasicParsing `
                                      -TimeoutSec 30 `
                                      -ErrorAction Stop
} catch {
    Write-Host ""
    Write-Host "  XX " -ForegroundColor Red -NoNewline
    Write-Host "Source installer is also unreachable." -ForegroundColor Red
    Write-Note "URL: $SourceUrl"
    Write-Note "Error: $($_.Exception.Message)"
    Write-Host ""
    Write-Note "What to try next:"
    Write-Note "  1. Check your internet connection / corporate proxy."
    Write-Note "  2. Open $ReleasesUI in a browser to see if releases exist."
    Write-Note "  3. Clone manually:"
    Write-Note "       git clone https://github.com/$Owner/$Repo.git"
    Write-Note "       cd $Repo && pwsh ./install.ps1"
    exit 1
}

# Forward -DeployPath into the source installer when supplied.
if ($DeployPath) {
    # Inject the parameter binding by wrapping in a scriptblock.
    $sb = [ScriptBlock]::Create($sourceScript.Content)
    & $sb -DeployPath $DeployPath
} else {
    Invoke-Expression $sourceScript.Content
}
exit $LASTEXITCODE
