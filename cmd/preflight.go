// preflight.go — `movie preflight` command.
//
// Verifies the local repo is current with its configured upstream tracking
// branch BEFORE running any build/scan operation. The expected remote and
// branch are auto-detected via `git rev-parse @{u}` — never hardcoded — so
// the command works equally for forks, feature branches, and self-hosted
// remotes.
//
// Exit codes:
//
//	0 = clean and up-to-date with upstream
//	1 = error (git unavailable, etc.)
//	4 = stale, diverged, dirty, or no upstream configured
package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/alimtvnetwork/movie-cli-v6/gitcheck"
)

// ExitRepoStale is the documented exit code for a stale/dirty/divergent
// local clone. CI scripts can branch on this without parsing output.
const ExitRepoStale = 4

var preflightJson bool

var preflightCmd = &cobra.Command{
	Use:   "preflight",
	Short: "Verify local repo is current with its upstream tracking branch",
	Long: `Auto-detects the expected remote and branch from the repo's configured
upstream (git rev-parse @{u}) — no hardcoded "origin/main". Reports
whether the local clone is up-to-date, behind, ahead, dirty, or missing
an upstream configuration. When stale, prints the exact recovery
commands tailored to the detected remote/branch.

Exit codes: 0 = clean & current, 1 = error, 4 = stale/dirty/no-upstream.`,
	Run: runPreflight,
}

func runPreflight(cmd *cobra.Command, args []string) {
	status, err := gitcheck.Inspect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "preflight: %v\n", err)
		os.Exit(1)
	}
	if preflightJson {
		emitPreflightJson(status)
		return
	}
	printPreflight(status)
	exitForPreflight(status)
}

func printPreflight(s gitcheck.Status) {
	fmt.Println("==> movie preflight")
	fmt.Println("  --------------------------------------------------")
	fmt.Printf("    branch        : %s\n", orPreflightDash(s.Branch))
	fmt.Printf("    upstream      : %s\n", orPreflightDash(s.RemoteRef))
	fmt.Printf("    remote        : %s\n", orPreflightDash(s.Remote))
	fmt.Printf("    ahead/behind  : %d / %d\n", s.Ahead, s.Behind)
	fmt.Printf("    working tree  : %s\n", cleanLabel(s.IsClean))
	fmt.Println("  --------------------------------------------------")
	fmt.Printf("  %s %s\n", preflightTag(s), s.Summary)
	if s.Recovery != "" {
		fmt.Printf("  hint: %s\n", s.Recovery)
	}
}

func emitPreflightJson(s gitcheck.Status) {
	out, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "preflight --json: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(out))
	exitForPreflight(s)
}

func exitForPreflight(s gitcheck.Status) {
	if !s.IsGitRepo {
		os.Exit(0)
	}
	if !s.HasUpstream || !s.IsCurrent || !s.IsClean {
		os.Exit(ExitRepoStale)
	}
	os.Exit(0)
}

func preflightTag(s gitcheck.Status) string {
	if !s.IsGitRepo {
		return "[ OK ]"
	}
	if s.HasUpstream && s.IsCurrent && s.IsClean {
		return "[ OK ]"
	}
	return "[WARN]"
}

func cleanLabel(isClean bool) string {
	if isClean {
		return "clean"
	}
	return "dirty"
}

func orPreflightDash(v string) string {
	if v == "" {
		return "-"
	}
	return v
}

func init() {
	preflightCmd.Flags().BoolVar(&preflightJson, "json", false,
		"Emit status as JSON for scripting/CI")
}