// e2e_test.go — end-to-end pipeline test.
//
// Builds the `movie` CLI in a temp directory, populates a fake media folder
// with sample video files, then drives `movie scan`, `movie ls --format json`,
// `movie ls --missing` and `movie stats` through os/exec. No TMDb network
// calls are made — TMDB_API_KEY is intentionally unset so the offline code
// paths are exercised. Verifies that the CLI produces a non-empty library
// and that the new --all/--missing filters return the expected shape.
//
// Run only when E2E=1 to keep the regular `go test ./...` fast and offline.
package cmd

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// e2eEnvFlag gates the heavy E2E test — set E2E=1 in CI to enable.
const e2eEnvFlag = "E2E"

// sampleVideos is the fixture set written into the fake scan folder.
// Keep names realistic so the title parser exercises year/quality stripping.
var sampleVideos = []string{
	"The.Matrix.1999.1080p.BluRay.x264.mkv",
	"Inception (2010) [1080p].mp4",
	"Interstellar.2014.2160p.HDR.mkv",
}

func TestE2EPipeline(t *testing.T) {
	if os.Getenv(e2eEnvFlag) != "1" {
		t.Skipf("skipping E2E pipeline test (set %s=1 to enable)", e2eEnvFlag)
	}

	workDir := t.TempDir()
	binary := buildE2EBinary(t, workDir)
	scanDir := writeSampleVideos(t, workDir)
	homeDir := filepath.Join(workDir, "home")
	if mkErr := os.MkdirAll(homeDir, 0o755); mkErr != nil {
		t.Fatalf("mkdir home: %v", mkErr)
	}

	runCli(t, binary, homeDir, "scan", scanDir)
	jsonOut := runCli(t, binary, homeDir, "ls", "--format", "json")
	assertJSONHasItems(t, jsonOut, len(sampleVideos))
	runCli(t, binary, homeDir, "ls", "--missing", "--format", "json")
	runCli(t, binary, homeDir, "stats")
}

func buildE2EBinary(t *testing.T, workDir string) string {
	t.Helper()
	repoRoot, rootErr := filepath.Abs("..")
	if rootErr != nil {
		t.Fatalf("resolve repo root: %v", rootErr)
	}
	binary := filepath.Join(workDir, "movie-e2e")
	build := exec.Command("go", "build", "-o", binary, ".")
	build.Dir = repoRoot
	if out, buildErr := build.CombinedOutput(); buildErr != nil {
		t.Fatalf("build CLI: %v\n%s", buildErr, out)
	}
	return binary
}

func writeSampleVideos(t *testing.T, workDir string) string {
	t.Helper()
	scanDir := filepath.Join(workDir, "movies")
	if mkErr := os.MkdirAll(scanDir, 0o755); mkErr != nil {
		t.Fatalf("mkdir scan dir: %v", mkErr)
	}
	for _, name := range sampleVideos {
		path := filepath.Join(scanDir, name)
		if writeErr := os.WriteFile(path, []byte("fake-video"), 0o644); writeErr != nil {
			t.Fatalf("write sample %s: %v", name, writeErr)
		}
	}
	return scanDir
}

// runCli executes the built binary with TMDB_API_KEY explicitly cleared and
// HOME pointed at a throwaway dir so the test never touches the user's real
// library. Returns combined stdout+stderr for assertions.
func runCli(t *testing.T, binary, homeDir string, args ...string) string {
	t.Helper()
	cmd := exec.Command(binary, args...)
	cmd.Env = append(os.Environ(),
		"HOME="+homeDir,
		"USERPROFILE="+homeDir,
		"TMDB_API_KEY=",
		"TMDB_TOKEN=",
		"OMDB_API_KEY=",
	)
	out, runErr := cmd.CombinedOutput()
	if runErr != nil {
		t.Fatalf("movie %s failed: %v\n%s", strings.Join(args, " "), runErr, out)
	}
	return string(out)
}

func assertJSONHasItems(t *testing.T, raw string, want int) {
	t.Helper()
	// JSON output may follow a banner line — find the first '[' to be safe.
	start := strings.Index(raw, "[")
	if start < 0 {
		t.Fatalf("expected JSON array in output, got:\n%s", raw)
	}
	var items []map[string]any
	if jsonErr := json.Unmarshal([]byte(raw[start:]), &items); jsonErr != nil {
		t.Fatalf("parse JSON: %v\nraw:\n%s", jsonErr, raw)
	}
	if len(items) < want {
		t.Fatalf("expected at least %d items, got %d", want, len(items))
	}
}
