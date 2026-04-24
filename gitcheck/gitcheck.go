// Package gitcheck inspects a local git repository and reports whether it
// is in sync with its expected upstream branch (default: origin/main).
//
// It is consumed by `movie preflight` and `movie doctor` to surface the
// "stale local clone" failure mode documented in
// .lovable/pending-issues/01-local-repo-stale.md before any build or
// updater action is attempted.
//
// All git invocations are read-only (rev-parse, status, fetch). No checkout,
// reset, pull, or clean is ever performed — repair instructions are printed
// for the user to run manually. This keeps gitcheck inside the boundary set
// by mem://constraints/updater-scope.
package gitcheck

import (
	"os/exec"
	"strings"

	"github.com/alimtvnetwork/movie-cli-v6/apperror"
)

// DefaultRemote and DefaultBranch describe the canonical upstream the
// preflight check compares against when the caller does not override them.
const (
	DefaultRemote = "origin"
	DefaultBranch = "main"
)

// Status is the structured result of a single gitcheck pass.
// Field order chosen for govet fieldalignment (strings first, bools last).
type Status struct {
	RepoPath       string
	Remote         string
	Branch         string
	CurrentBranch  string
	LocalCommit    string
	RemoteCommit   string
	Ahead          string
	Behind         string
	IsRepo         bool
	IsOnBranch     bool
	IsClean        bool
	IsUpToDate     bool
	HasFetched     bool
}

// Options controls a Check run.
type Options struct {
	RepoPath string // git working tree root
	Remote   string // default "origin"
	Branch   string // default "main"
	DoFetch  bool   // run `git fetch <remote> <branch>` first (read-only)
}

// Check inspects the repo and returns a Status describing its sync state.
// Returns an error only when git itself is unusable; mismatches are
// reported as fields on Status, not as Go errors.
func Check(opts Options) (*Status, error) {
	status := newStatus(opts)
	if !looksLikeRepo(status.RepoPath) {
		return status, nil
	}
	status.IsRepo = true
	if opts.DoFetch {
		status.HasFetched = runFetch(status) == nil
	}
	populateBranch(status)
	populateCommits(status)
	populateCounts(status)
	populateClean(status)
	status.IsUpToDate = computeUpToDate(status)
	return status, nil
}

// IsStale returns true when the local clone is behind the expected upstream,
// is on a different branch, or has uncommitted local changes that would
// block a fast-forward. It is the single signal callers act on.
func (s *Status) IsStale() bool {
	if !s.IsRepo {
		return true
	}
	if !s.IsOnBranch || !s.IsClean {
		return true
	}
	return !s.IsUpToDate
}

// RecoveryCommands returns the exact shell commands a user must run to
// bring the local clone back in sync with the expected upstream.
func (s *Status) RecoveryCommands() []string {
	remote := s.Remote
	branch := s.Branch
	return []string{
		"git fetch " + remote,
		"git reset --hard " + remote + "/" + branch,
		"git clean -fd",
	}
}

func newStatus(opts Options) *Status {
	remote := strings.TrimSpace(opts.Remote)
	if remote == "" {
		remote = DefaultRemote
	}
	branch := strings.TrimSpace(opts.Branch)
	if branch == "" {
		branch = DefaultBranch
	}
	return &Status{RepoPath: opts.RepoPath, Remote: remote, Branch: branch}
}

func looksLikeRepo(dir string) bool {
	if strings.TrimSpace(dir) == "" {
		return false
	}
	_, err := gitOut(dir, "rev-parse", "--is-inside-work-tree")
	return err == nil
}

func runFetch(s *Status) error {
	_, err := gitOut(s.RepoPath, "fetch", s.Remote, s.Branch)
	if err != nil {
		return apperror.Wrap("git fetch", err)
	}
	return nil
}

func populateBranch(s *Status) {
	cur, err := gitOut(s.RepoPath, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return
	}
	s.CurrentBranch = cur
	s.IsOnBranch = cur == s.Branch
}

func populateCommits(s *Status) {
	if local, err := gitOut(s.RepoPath, "rev-parse", "HEAD"); err == nil {
		s.LocalCommit = shortSHA(local)
	}
	ref := s.Remote + "/" + s.Branch
	if remote, err := gitOut(s.RepoPath, "rev-parse", ref); err == nil {
		s.RemoteCommit = shortSHA(remote)
	}
}

func populateCounts(s *Status) {
	ref := s.Remote + "/" + s.Branch
	out, err := gitOut(s.RepoPath, "rev-list", "--left-right", "--count", "HEAD..."+ref)
	if err != nil {
		return
	}
	parts := strings.Fields(out)
	if len(parts) != 2 {
		return
	}
	s.Ahead = parts[0]
	s.Behind = parts[1]
}

func populateClean(s *Status) {
	out, err := gitOut(s.RepoPath, "status", "--porcelain")
	if err != nil {
		return
	}
	s.IsClean = strings.TrimSpace(out) == ""
}

func computeUpToDate(s *Status) bool {
	if s.LocalCommit == "" || s.RemoteCommit == "" {
		return false
	}
	if s.Behind != "" && s.Behind != "0" {
		return false
	}
	return s.LocalCommit == s.RemoteCommit
}

func shortSHA(full string) string {
	if len(full) >= 7 {
		return full[:7]
	}
	return full
}

func gitOut(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	text := strings.TrimSpace(string(out))
	if err != nil {
		return "", apperror.New("%s", text)
	}
	return text, nil
}