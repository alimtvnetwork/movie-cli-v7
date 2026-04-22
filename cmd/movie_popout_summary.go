// movie_popout_summary.go — final summary report for the popout command.
//
// Printed at the very end of every popout run (success, partial, or zero)
// so the user gets one consolidated bottom-line: how many files were
// flattened to the root and how many subfolders were compacted into
// <root>/.temp/. The batch ID is included so it can be fed straight into
// `movie undo --batch <id>` if needed.
package cmd

import "fmt"

// printPopoutSummary renders the closing report block. Always called once
// per run, after both the move phase and the compaction phase have settled.
func printPopoutSummary(moved, failed, compacted int, batchID string) {
	fmt.Println()
	fmt.Println("  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  📊 Popout Summary")
	fmt.Println("  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("    🎬 Files moved to root  : %d\n", moved)
	if failed > 0 {
		fmt.Printf("    ⚠️  Files failed to move : %d\n", failed)
	}
	fmt.Printf("    📦 Folders → .temp/      : %d\n", compacted)
	fmt.Printf("    📋 Batch ID              : %s\n", shortBatchID(batchID))
	fmt.Println("  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	if moved == 0 && compacted == 0 {
		fmt.Println("  ℹ️  Nothing changed.")
		return
	}
	fmt.Printf("  ↩️  Undo this run: movie undo --batch %s\n", shortBatchID(batchID))
}

// shortBatchID trims the hex batch ID to its first 8 chars for display.
// Falls back to the full string if it's already short.
func shortBatchID(id string) string {
	if len(id) <= 8 {
		return id
	}
	return id[:8]
}
