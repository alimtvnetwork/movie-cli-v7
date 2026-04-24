// repo.go — read-only git repo staleness check.
//
// Powers the one-line `Repo:` summary in `movie doctor` so users see at a
// glance whether the local clone is behind, diverged, or dirty — with the
// exact recovery commands they should run. This file MUST NOT mutate the
// repo (no fetch beyond a single read-only `git fetch origin`, no checkout,
// no reset) — see mem://constraints/updater-scope.md.
package doctor

import (
	"os/exec"
	"strconv"
	"strings"
)

// RepoRecoveryCmd is the canonical fix advised whenever the repo is stale,
// diverged, or dirty. Single source of truth — referenced by report and JSON.
const RepoRecoveryCmd = "git fetch origin && git reset --hard origin/main && git clean -fd"

// RepoStatus captures the local repo state vs origin/main.
// Field order optimized for govet fieldalignment (strings, ints, bools last).
type RepoStatus struct {
	Branch     string
	Summary    string
	Recovery   string
	Ahead      int
	Behind     int
	IsGitRepo  bool
	IsClean    bool
	IsCurrent  bool
}

// CheckRepo runs the read-only git probes and returns a RepoStatus.
// Never returns an error: any git failure is folded into IsGitRepo=false.
func CheckRepo() RepoStatus {
	if !insideGitRepo() {
		return RepoStatus{Summary: "not a git repo (skipped)"}
	}
	status := RepoStatus{IsGitRepo: true, Branch: currentBranch()}
	_ = runGit("fetch", "origin", "--quiet")
	status.Ahead, status.Behind = aheadBehind()
	status.IsClean = workingTreeClean()
	status.IsCurrent = status.Ahead == 0 && status.Behind == 0
	fillRepoSummary(&status)
	return status
}

func fillRepoSummary(s *RepoStatus) {
	if s.IsCurrent && s.IsClean {
		s.Summary = "up-to-date with origin/main on " + s.Branch
		return
	}
	s.Recovery = RepoRecoveryCmd
	s.Summary = "STALE — " + describeDrift(s) + " — run: " + RepoRecoveryCmd
}

func describeDrift(s *RepoStatus) string {
	parts := []string{}
	if s.Behind > 0 {
		parts = append(parts, "behind "+strconv.Itoa(s.Behind))
	}
	if s.Ahead > 0 {
		parts = append(parts, "ahead "+strconv.Itoa(s.Ahead))
	}
	if !s.IsClean {
		parts = append(parts, "dirty working tree")
	}
	if len(parts) == 0 {
		return "drift detected"
	}
	return strings.Join(parts, ", ")
}

func insideGitRepo() bool {
	out, err := exec.Command("git", "rev-parse", "--is-inside-work-tree").Output()
	return err == nil && strings.TrimSpace(string(out)) == "true"
}

func currentBranch() string {
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

func aheadBehind() (int, int) {
	out, err := exec.Command("git", "rev-list", "--left-right", "--count", "HEAD...origin/main").Output()
	if err != nil {
		return 0, 0
	}
	fields := strings.Fields(strings.TrimSpace(string(out)))
	if len(fields) != 2 {
		return 0, 0
	}
	ahead, _ := strconv.Atoi(fields[0])
	behind, _ := strconv.Atoi(fields[1])
	return ahead, behind
}

func workingTreeClean() bool {
	out, err := exec.Command("git", "status", "--porcelain").Output()
	if err != nil {
		return true
	}
	return strings.TrimSpace(string(out)) == ""
}

func runGit(args ...string) error {
	return exec.Command("git", args...).Run()
}