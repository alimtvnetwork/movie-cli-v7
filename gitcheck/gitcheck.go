// Package gitcheck inspects the local git clone read-only and reports
// whether it is current with its configured upstream tracking branch.
//
// `movie preflight` uses this to detect stale clones BEFORE running any
// build/scan command, and to print the exact recovery commands.
//
// IMPORTANT: this package MUST stay read-only — no fetch beyond a single
// `git fetch <remote> --quiet`, no checkout, no reset. All mutating git
// work belongs in run.ps1 (see mem://constraints/updater-scope.md).
package gitcheck

import (
	"os/exec"
	"strconv"
	"strings"

	"github.com/alimtvnetwork/movie-cli-v6/apperror"
)

// Status is the result of an upstream-aware preflight inspection.
// Field order optimized for govet fieldalignment (strings, ints, bools last).
type Status struct {
	Branch     string // local branch, e.g. "main"
	Remote     string // detected remote name, e.g. "origin"
	RemoteRef  string // detected upstream short ref, e.g. "origin/main"
	Summary    string // one-line human summary
	Recovery   string // recovery command when not clean+current
	Ahead      int
	Behind     int
	IsGitRepo  bool
	IsClean    bool
	IsCurrent  bool
	HasUpstream bool
}

// Inspect runs the read-only probes and returns a Status. Never returns an
// error for normal "not a repo" / "no upstream" cases — those are encoded
// in the Status fields so callers can render them uniformly.
func Inspect() (Status, error) {
	if !insideGitRepo() {
		return Status{Summary: "not a git repo (skipped)"}, nil
	}
	status := Status{IsGitRepo: true, Branch: currentBranch()}
	if err := resolveUpstream(&status); err != nil {
		return status, nil
	}
	_ = runGit("fetch", status.Remote, "--quiet")
	status.Ahead, status.Behind = aheadBehind(status.RemoteRef)
	status.IsClean = workingTreeClean()
	status.IsCurrent = status.Ahead == 0 && status.Behind == 0
	fillSummary(&status)
	return status, nil
}

// resolveUpstream auto-detects the configured upstream tracking branch via
// `git rev-parse --abbrev-ref --symbolic-full-name @{u}` and splits it into
// remote + short ref. No remote/branch is ever hardcoded.
func resolveUpstream(s *Status) error {
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref",
		"--symbolic-full-name", "@{u}").Output()
	if err != nil {
		s.Summary = "no upstream tracking branch configured for " + s.Branch
		s.Recovery = "git branch --set-upstream-to=origin/" + s.Branch + " " + s.Branch
		return apperror.Wrap("gitcheck: no upstream", err)
	}
	ref := strings.TrimSpace(string(out))
	s.RemoteRef = ref
	s.HasUpstream = true
	s.Remote = splitRemote(ref)
	return nil
}

func splitRemote(ref string) string {
	if i := strings.Index(ref, "/"); i > 0 {
		return ref[:i]
	}
	return "origin"
}

func fillSummary(s *Status) {
	if s.IsCurrent && s.IsClean {
		s.Summary = "up-to-date with " + s.RemoteRef + " on " + s.Branch
		return
	}
	s.Recovery = recoveryFor(s)
	s.Summary = "STALE — " + describeDrift(s) + " — run: " + s.Recovery
}

func recoveryFor(s *Status) string {
	return "git fetch " + s.Remote +
		" && git reset --hard " + s.RemoteRef +
		" && git clean -fd"
}

func describeDrift(s *Status) string {
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

func aheadBehind(remoteRef string) (int, int) {
	out, err := exec.Command("git", "rev-list", "--left-right", "--count",
		"HEAD..."+remoteRef).Output()
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