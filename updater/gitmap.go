package updater

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// GitMapRelease represents a .gitmap/release/*.json entry.
type GitMapRelease struct {
	Version      string `json:"version"`
	Branch       string `json:"branch"`
	SourceBranch string `json:"sourceBranch"`
	Tag          string `json:"tag"`
	IsLatest     bool   `json:"isLatest"`
}

// GitmapDir is the relative path to the gitmap release directory.
const GitmapDir = ".gitmap/release"

// ReadGitMapLatest reads .gitmap/release/latest.json from the given repo root.
func ReadGitMapLatest(repoRoot string) (*GitMapRelease, error) {
	path := filepath.Join(repoRoot, GitmapDir, "latest.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var rel GitMapRelease
	if err := json.Unmarshal(data, &rel); err != nil {
		return nil, err
	}
	return &rel, nil
}
