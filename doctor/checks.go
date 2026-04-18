// checks.go — individual diagnostic checks. Each returns a Finding (or slice).
// Functions stay ≤15 lines; helpers live in paths.go and workers.go.
package doctor

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	idPathMismatch  = "path-mismatch"
	idDeployInPath  = "deploy-in-path"
	idStaleWorker   = "stale-worker"
	idVersionDrift  = "version-drift"
)

func populatePaths(report *Report) error {
	source, err := resolveDeploySource()
	if err != nil {
		return err
	}
	target, _ := resolveActiveBinary()
	report.Source = source
	report.Target = target
	report.DeployDir = filepath.Dir(source)
	return nil
}

func checkPathMismatch(report *Report) Finding {
	if report.Target == "" {
		return finding(idPathMismatch, "PATH-resolved 'movie' binary",
			SeverityErr, "no 'movie' on PATH", "Add deploy dir to PATH", false)
	}
	if pathsEqual(report.Source, report.Target) {
		return finding(idPathMismatch, "Deploy target matches active PATH binary",
			SeverityOK, fmt.Sprintf("both -> %s", report.Target), "", false)
	}
	detail := fmt.Sprintf("deploy=%s vs active=%s", report.Source, report.Target)
	return finding(idPathMismatch, "Deploy path differs from active PATH binary",
		SeverityErr, detail, "Run `movie doctor --fix` (calls self-replace)", true)
}

func checkDeployInPath(report *Report) Finding {
	if report.DeployDir == "" {
		return finding(idDeployInPath, "Deploy directory in PATH",
			SeverityWarn, "deploy dir unknown", "Configure powershell.json deployPath", false)
	}
	if pathContainsDir(report.DeployDir) {
		return finding(idDeployInPath, "Deploy directory is in PATH",
			SeverityOK, report.DeployDir, "", false)
	}
	return finding(idDeployInPath, "Deploy directory is NOT in PATH",
		SeverityWarn, report.DeployDir, "Add it to PATH (User env)", true)
}

func checkStaleWorkers(report *Report) []Finding {
	workers := findStaleWorkers(report.DeployDir)
	if len(workers) == 0 {
		return []Finding{finding(idStaleWorker, "Stale handoff workers",
			SeverityOK, "none found", "", false)}
	}
	detail := strings.Join(workers, "\n      ")
	return []Finding{finding(idStaleWorker, fmt.Sprintf("%d stale *-update-* worker(s) on disk", len(workers)),
		SeverityWarn, detail, "Run `movie doctor --fix` to sweep", true)}
}

func checkVersionDrift(report *Report) Finding {
	source := readBinaryVersion(report.Source)
	target := readBinaryVersion(report.Target)
	if source == "" || target == "" {
		return finding(idVersionDrift, "Version drift", SeverityWarn,
			"could not read one or both versions", "", false)
	}
	if source == target {
		return finding(idVersionDrift, "Active binary matches deployed version",
			SeverityOK, source, "", false)
	}
	detail := fmt.Sprintf("active=%s deployed=%s", target, source)
	return finding(idVersionDrift, "Active binary version differs from deployed",
		SeverityWarn, detail, "Run `movie doctor --fix` (calls self-replace)", true)
}

func finding(id, title string, sev Severity, detail, hint string, fixable bool) Finding {
	return Finding{ID: id, Title: title, Severity: sev,
		Detail: detail, FixHint: hint, IsFixable: fixable}
}

func pathsEqual(a, b string) bool {
	if runtime.GOOS == "windows" {
		return strings.EqualFold(a, b)
	}
	return a == b
}
