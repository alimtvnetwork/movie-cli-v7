package updater

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/alimtvnetwork/movie-cli-v5/apperror"
)

// executeUpdateWindows writes a temp PowerShell script and runs it.
func executeUpdateWindows(repoPath, targetBinary string) error {
	scriptPath, err := writeUpdateScript(repoPath, targetBinary)
	if err != nil {
		return apperror.Wrap("cannot write update script", err)
	}
	defer os.Remove(scriptPath)

	return runPowerShellScript(scriptPath)
}

// executeUpdateUnix runs the update via pwsh.
func executeUpdateUnix(repoPath, targetBinary string) error {
	if !hasPwshWithRunPS1(repoPath) {
		runPS1 := filepath.Join(repoPath, "run.ps1")
		return apperror.New("pwsh is required to run %s", runPS1)
	}
	scriptPath, err := writeUpdateScript(repoPath, targetBinary)
	if err != nil {
		return apperror.Wrap("cannot write update script", err)
	}
	defer os.Remove(scriptPath)
	return runPowerShellScript(scriptPath)
}

func hasPwshWithRunPS1(repoPath string) bool {
	if _, err := exec.LookPath("pwsh"); err != nil {
		return false
	}
	runPS1 := filepath.Join(repoPath, "run.ps1")
	_, statErr := os.Stat(runPS1)
	return statErr == nil
}

// writeUpdateScript generates a temp PowerShell script for the update.
func writeUpdateScript(repoPath, targetBinary string) (string, error) {
	script := buildUpdateScriptContent(repoPath, targetBinary, currentBinaryPath())

	tmpFile, err := os.CreateTemp(os.TempDir(), "movie-update-script-*.ps1")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	// UTF-8 BOM for PowerShell compatibility
	bom := []byte{0xEF, 0xBB, 0xBF}
	if _, err := tmpFile.Write(bom); err != nil {
		return "", err
	}
	if _, err := tmpFile.WriteString(script); err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}

func currentBinaryPath() string {
	binaryPath, err := os.Executable()
	if err != nil {
		return ""
	}
	if resolved, evalErr := filepath.EvalSymlinks(binaryPath); evalErr == nil {
		return resolved
	}
	return binaryPath
}

// buildUpdateScriptContent generates the PowerShell script content.
func buildUpdateScriptContent(repoPath, targetBinary, workerBinary string) string {
	repoPath = powerShellString(repoPath)
	targetBinary = powerShellString(targetBinary)
	workerBinary = powerShellString(workerBinary)

	return fmt.Sprintf(`$ErrorActionPreference = "Stop"
$repoPath = "%s"
$targetBinary = "%s"
$workerBinary = "%s"

function Resolve-VersionBinary {
    if ($targetBinary) {
        return $targetBinary
    }

    $movieBin = Get-Command movie -ErrorAction SilentlyContinue
    if ($movieBin -and $movieBin.Source -and (Test-Path $movieBin.Source)) {
        return $movieBin.Source
    }

    return $null
}

# Capture current version
$versionBinary = Resolve-VersionBinary
$oldVersion = "unknown"
if ($versionBinary -and (Test-Path $versionBinary)) {
    $oldVersion = (& $versionBinary version 2>&1) -join " "
}
Write-Host "  Version before: $oldVersion" -ForegroundColor Gray

# Wait for parent to release file handles
Start-Sleep -Seconds 1.2

# Build and deploy from repo root via run.ps1
$runScript = Join-Path $repoPath "run.ps1"
if (-not (Test-Path $runScript)) {
    Write-Host "  run.ps1 not found at $runScript" -ForegroundColor Red
    exit 1
}

Write-Host "  Running update via $runScript" -ForegroundColor Cyan
$runArgs = @{ Update = $true }
if ($targetBinary) {
    $deployDir = Split-Path -Parent $targetBinary
    $targetName = Split-Path -Leaf $targetBinary
    if ($deployDir) {
        $runArgs["DeployPath"] = $deployDir
    }
    if ($targetName) {
        $runArgs["BinaryNameOverride"] = $targetName
    }
    Write-Host "  Deploy target: $targetBinary" -ForegroundColor Gray
}
& $runScript @runArgs
$runExit = $LASTEXITCODE
if ($runExit -ne 0) {
    exit $runExit
}

# Compare versions from the original target binary
$versionBinary = Resolve-VersionBinary
$newVersion = "unknown"
if ($versionBinary -and (Test-Path $versionBinary)) {
    $newVersion = (& $versionBinary version 2>&1) -join " "
}

Write-Host ""
if ($oldVersion -eq $newVersion) {
    Write-Host "  WARNING: Version unchanged after update" -ForegroundColor Yellow
    Write-Host "  Was version/info.go bumped?" -ForegroundColor Yellow
} else {
    Write-Host "  Updated: $oldVersion -> $newVersion" -ForegroundColor Green
}

# Show changelog from the updated target binary
if ($versionBinary -and (Test-Path $versionBinary)) {
    Write-Host ""
    $clOutput = & $versionBinary changelog --latest 2>&1
    foreach ($cl in $clOutput) { Write-Host "  $cl" }
}

# Auto-cleanup
if ($versionBinary -and (Test-Path $versionBinary)) {
    $cleanupArgs = @("update-cleanup")
    if ($workerBinary) {
        $cleanupArgs += @("--skip-path", $workerBinary)
    }
    $prevPref = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    $null = & $versionBinary @cleanupArgs 2>&1
    $ErrorActionPreference = $prevPref
}
`, repoPath, targetBinary, workerBinary)
}

func powerShellString(value string) string {
	value = strings.ReplaceAll(value, "`", "``")
	value = strings.ReplaceAll(value, "\"", "`\"")
	return value
}

// runPowerShellScript executes a PowerShell script with output piped to terminal.
func runPowerShellScript(scriptPath string) error {
	psBin := "powershell"
	if runtime.GOOS != "windows" {
		psBin = "pwsh"
	}

	cmd := exec.Command(psBin, "-ExecutionPolicy", "Bypass", "-NoProfile", "-NoLogo", "-File", scriptPath)
	return runAttached(cmd)
}
