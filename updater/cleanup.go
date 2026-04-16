package updater

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/alimtvnetwork/movie-cli-v5/apperror"
)

// powershellConfigName is the build/deploy config file at the repo root.
const powershellConfigName = "powershell.json"

// powershellConfig mirrors the subset of powershell.json we need.
type powershellConfig struct {
	DeployPath  string `json:"deployPath"`
	BuildOutput string `json:"buildOutput"`
}

// Cleanup removes leftover temp binaries and backup files from previous updates.
// It scans every directory where artifacts may land:
//   - active binary's directory
//   - OS temp directory
//   - deployPath from powershell.json (the actual deploy target)
//   - buildOutput from powershell.json (intermediate build dir, e.g. ./bin)
//
// Patterns cleaned per directory:
//   - <name>-update-*[.exe]   handoff copies from any prior PID
//   - *.old                   rename-first deploy backups
//   - *.bak                   legacy backup files
//   - legacy binaries (movie*, mahin*) and their handoff copies
func Cleanup() (int, error) {
	cleaned := 0

	selfPath, err := os.Executable()
	if err != nil {
		return 0, apperror.Wrap("cannot determine executable path", err)
	}
	if resolved, evalErr := filepath.EvalSymlinks(selfPath); evalErr == nil {
		selfPath = resolved
	}

	baseName := binaryBaseName(selfPath)
	dirs := candidateDirs(selfPath)

	fmt.Println("  Scanning directories:")
	for _, dir := range dirs {
		fmt.Printf("    - %s\n", dir)
	}

	for _, dir := range dirs {
		cleaned += cleanDir(dir, baseName, selfPath)
	}

	return cleaned, nil
}

// binaryBaseName returns the running binary name without extension,
// e.g. "gitmap.exe" -> "gitmap", "movie" -> "movie".
func binaryBaseName(selfPath string) string {
	name := filepath.Base(selfPath)
	ext := filepath.Ext(name)
	if ext != "" {
		name = strings.TrimSuffix(name, ext)
	}
	return name
}

// candidateDirs returns the unique directories to scan for leftovers:
// active binary dir, OS temp dir, deployPath, and buildOutput from powershell.json.
func candidateDirs(selfPath string) []string {
	seen := map[string]struct{}{}
	var dirs []string
	add := func(d string) {
		if d == "" {
			return
		}
		abs, err := filepath.Abs(d)
		if err != nil {
			return
		}
		key := strings.ToLower(abs)
		if _, ok := seen[key]; ok {
			return
		}
		if info, statErr := os.Stat(abs); statErr != nil || !info.IsDir() {
			return
		}
		seen[key] = struct{}{}
		dirs = append(dirs, abs)
	}

	add(filepath.Dir(selfPath))
	add(os.TempDir())

	cfg := loadPowershellConfig()
	if cfg != nil {
		add(cfg.DeployPath)
		add(cfg.BuildOutput)
	}

	return dirs
}

// loadPowershellConfig reads powershell.json from the repo root.
// Returns nil if the repo or file cannot be located.
func loadPowershellConfig() *powershellConfig {
	repoPath, _, err := findRepoPath()
	if err != nil || repoPath == "" {
		return nil
	}
	data, readErr := os.ReadFile(filepath.Join(repoPath, powershellConfigName))
	if readErr != nil {
		return nil
	}
	var cfg powershellConfig
	if jsonErr := json.Unmarshal(data, &cfg); jsonErr != nil {
		return nil
	}
	if cfg.BuildOutput != "" && !filepath.IsAbs(cfg.BuildOutput) {
		cfg.BuildOutput = filepath.Join(repoPath, cfg.BuildOutput)
	}
	return &cfg
}

// cleanDir removes all known leftover patterns in a single directory.
func cleanDir(dir, baseName, selfPath string) int {
	patterns := []string{
		filepath.Join(dir, baseName+"-update-*"),
		filepath.Join(dir, "*.old"),
		filepath.Join(dir, "*.bak"),
	}
	// Legacy artifacts from previous project names (movie-cli, mahin).
	// The selfPath guard in cleanGlob still protects the running binary.
	for _, legacy := range legacyBaseNames {
		if legacy == baseName {
			continue
		}
		patterns = append(patterns,
			filepath.Join(dir, legacy+"-update-*"),
			filepath.Join(dir, legacy+".exe"),
			filepath.Join(dir, legacy),
		)
	}
	cleaned := 0
	for _, pattern := range patterns {
		cleaned += cleanGlob(pattern, selfPath)
	}
	return cleaned
}

// legacyBaseNames are previous binary names whose leftovers should be swept.
// Add new names here when the project is renamed again.
var legacyBaseNames = []string{"movie", "mahin"}

// cleanGlob removes files matching a glob pattern, skipping the running binary.
func cleanGlob(pattern, selfPath string) int {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return 0
	}

	cleaned := 0
	for _, match := range matches {
		if shouldSkip(match, selfPath) {
			continue
		}
		if err := os.Remove(match); err != nil {
			fmt.Fprintf(os.Stderr, "  ⚠ Could not remove %s: %v\n", filepath.Base(match), err)
			continue
		}
		fmt.Printf("  Removed: %s\n", filepath.Base(match))
		cleaned++
	}
	return cleaned
}

// shouldSkip returns true when the match is the running binary or a directory.
func shouldSkip(match, selfPath string) bool {
	abs, _ := filepath.Abs(match)
	if pathsEqual(abs, selfPath) {
		return true
	}
	info, err := os.Stat(match)
	if err != nil {
		return true
	}
	return info.IsDir()
}

// pathsEqual compares two paths case-insensitively on Windows.
func pathsEqual(a, b string) bool {
	if runtime.GOOS == "windows" {
		return strings.EqualFold(a, b)
	}
	return a == b
}
