// workers.go — stale handoff worker discovery + version reading.
package doctor

import (
	"os/exec"
	"path/filepath"
	"strings"
)

func findStaleWorkers(deployDir string) []string {
	if deployDir == "" {
		return nil
	}
	pattern := filepath.Join(deployDir, "*-update-*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil
	}
	return matches
}

func readBinaryVersion(binaryPath string) string {
	if binaryPath == "" {
		return ""
	}
	out, err := exec.Command(binaryPath, "version").Output()
	if err != nil {
		return ""
	}
	return extractSemver(string(out))
}

func extractSemver(text string) string {
	for _, line := range strings.Split(text, "\n") {
		idx := strings.Index(line, "v")
		if idx < 0 {
			continue
		}
		token := strings.Fields(line[idx:])
		if len(token) == 0 {
			continue
		}
		candidate := strings.TrimSpace(token[0])
		if looksLikeSemver(candidate) {
			return candidate
		}
	}
	return strings.TrimSpace(text)
}

func looksLikeSemver(token string) bool {
	if !strings.HasPrefix(token, "v") {
		return false
	}
	return strings.Count(token, ".") >= 2
}
