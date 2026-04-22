// history_summary.go — shared post-operation summary for movie undo/redo.
//
// Whenever an undo/redo flow finishes (single op, batch, or just `--list`),
// we want the user to see four numbers:
//
//	matched   — rows that passed the dir scope + glob filter
//	executed  — rows we actually undid/redid (matched − failed for batch;
//	            1 for single op when it succeeded)
//	failed    — rows where the filesystem operation errored out
//	skipped   — rows that exist in the DB but were dropped by the
//	            dir-scope or glob filter (i.e. "out of scope")
//
// Single-op flows just set Matched=1 + Executed=1 (or Failed=1) and
// compute Skipped from raw vs filtered counts.
package cmd

import "fmt"

// HistorySummary captures the counters reported at the end of an
// undo/redo run. Verb is "Undo" or "Redo" — used in the banner.
type HistorySummary struct {
	Verb     string // "Undo" or "Redo"
	Matched  int    // rows passing the filter
	Executed int    // rows actually applied
	Failed   int    // rows that failed during execution
	Skipped  int    // rows dropped by dir/glob filter
}

// printHistorySummary writes a 4-line block. Always called even when all
// counters are zero — gives the user explicit "nothing happened" feedback
// instead of silent exit.
func printHistorySummary(s HistorySummary) {
	fmt.Println()
	fmt.Printf("📊 %s summary\n", s.Verb)
	fmt.Printf("   ✅ executed:          %d\n", s.Executed)
	fmt.Printf("   ⚠️  failed:           %d\n", s.Failed)
	fmt.Printf("   🔍 matched filter:    %d\n", s.Matched)
	fmt.Printf("   🚫 skipped (out of scope): %d\n", s.Skipped)
}

// countScopeSkipped returns the number of rows the filter dropped.
// Negative results clamp to zero (defensive — should never happen).
func countScopeSkipped(raw, kept int) int {
	diff := raw - kept
	if diff < 0 {
		return 0
	}
	return diff
}
