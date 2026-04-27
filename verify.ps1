# verify.ps1 — Post-install verification for the movie CLI (Windows / pwsh).
#
# Checks prerequisites (Go, Git) and confirms the installed `movie` binary
# is reachable, executable, and reports a sane version string.
#
# Usage:
#   pwsh verify.ps1
#   pwsh verify.ps1 -Binary movie
#   pwsh verify.ps1 -Dir "$HOME\bin"
#
# Exit codes:
#   0  all checks passed
#   1  one or more checks failed

[CmdletBinding()]
param(
    [string]$Binary = "movie",
    [string]$Dir = ""
)

$ErrorActionPreference = "Continue"
$script:Pass = 0
$script:Fail = 0
$script:Warn = 0

function Write-Pass($m) { Write-Host "  [PASS] $m" -ForegroundColor Green; $script:Pass++ }
function Write-Fail($m) { Write-Host "  [FAIL] $m" -ForegroundColor Red;   $script:Fail++ }
function Write-Warn($m) { Write-Host "  [WARN] $m" -ForegroundColor Yellow; $script:Warn++ }
function Write-Info($m) { Write-Host "  $m" -ForegroundColor DarkGray }
function Write-Hdr($m)  { Write-Host ""; Write-Host "== $m ==" -ForegroundColor Cyan }

Write-Host ""
Write-Host "movie CLI - install verification" -ForegroundColor Cyan

# ── 1. Prerequisites ────────────────────────────────────────
Write-Hdr "Prerequisites"

$git = Get-Command git -ErrorAction SilentlyContinue
if ($git) {
    Write-Pass "git found ($((git --version) -split "`n" | Select-Object -First 1))"
} else {
    Write-Fail "git not found in PATH"
}

$go = Get-Command go -ErrorAction SilentlyContinue
if ($go) {
    $goVerRaw = (& go version) -replace '.*go(\d+\.\d+(\.\d+)?).*', '$1'
    $parts = $goVerRaw.Split('.')
    $major = [int]$parts[0]; $minor = [int]$parts[1]
    if (($major -gt 1) -or ($major -eq 1 -and $minor -ge 22)) {
        Write-Pass "go $goVerRaw (>= 1.22)"
    } else {
        Write-Warn "go $goVerRaw is older than required 1.22"
    }
} else {
    Write-Warn "go not found (only required for building from source)"
}

# ── 2. Locate binary ────────────────────────────────────────
Write-Hdr "Binary"

$binPath = $null
if ($Dir) {
    $candidate = Join-Path $Dir "$Binary.exe"
    if (-not (Test-Path $candidate)) { $candidate = Join-Path $Dir $Binary }
    if (Test-Path $candidate) {
        $binPath = $candidate
        Write-Pass "found $binPath"
    } else {
        Write-Fail "no $Binary executable in $Dir"
    }
} else {
    $cmd = Get-Command $Binary -ErrorAction SilentlyContinue
    if ($cmd) {
        $binPath = $cmd.Source
        Write-Pass "found on PATH at $binPath"
    } else {
        Write-Fail "$Binary not found on PATH"
        Write-Info "try: `$env:PATH += `";$HOME\bin`""
    }
}

# ── 3. Execution check ──────────────────────────────────────
Write-Hdr "Execution"

if ($binPath) {
    try {
        $verOut = & $binPath version 2>&1 | Out-String
        if ($LASTEXITCODE -eq 0 -and $verOut -match 'v\d+\.\d+\.\d+') {
            Write-Pass "version command works"
            Write-Info ($verOut -split "`n" | Select-Object -First 1)
        } elseif ($LASTEXITCODE -eq 0) {
            Write-Warn "version command ran but output looks unexpected"
            Write-Info ($verOut -split "`n" | Select-Object -First 1)
        } else {
            Write-Fail "'$Binary version' exited with code $LASTEXITCODE"
            Write-Info ($verOut -split "`n" | Select-Object -First 3 | Out-String)
        }
    } catch {
        Write-Fail "running '$Binary version' threw: $_"
    }

    try {
        & $binPath --help *> $null
        if ($LASTEXITCODE -eq 0) {
            Write-Pass "help command responds"
        } else {
            & $binPath help *> $null
            if ($LASTEXITCODE -eq 0) { Write-Pass "help command responds" }
            else { Write-Warn "help command did not respond cleanly" }
        }
    } catch {
        Write-Warn "help command threw: $_"
    }
} else {
    Write-Info "skipping execution checks - binary not located"
}

# ── Summary ─────────────────────────────────────────────────
Write-Hdr "Summary"
Write-Host ("  {0} passed  {1} warnings  {2} failed" -f $script:Pass, $script:Warn, $script:Fail)

if ($script:Fail -gt 0) {
    Write-Host ""
    Write-Host "Verification failed. See messages above." -ForegroundColor Red
    Write-Host ""
    exit 1
}
Write-Host ""
Write-Host "All required checks passed - movie CLI is ready." -ForegroundColor Green
Write-Host ""
exit 0
