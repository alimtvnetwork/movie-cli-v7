// movie_redo_handlers.go — redo subcommand handlers (list, by-id, batch, last).
package cmd

import (
	"bufio"
	"fmt"

	"github.com/alimtvnetwork/movie-cli-v5/db"
	"github.com/alimtvnetwork/movie-cli-v5/errlog"
)

func showRedoableList(database *db.DB, f ScopeFilter) {
	fmt.Println("⏩ Recent redoable operations")
	printScopeBanner(f)

	moveSkipped := countRedoableMoveSkipped(database, f)
	actionSkipped := countRedoableActionSkipped(database, f)
	matchedMoves := countMatchedRedoMoves(database, f)
	matchedActions := countMatchedRedoActions(database, f)
	printScopeMatchedCounts(matchedMoves, matchedActions, moveSkipped, actionSkipped)
	fmt.Println()

	redoableMoves := printRedoableMoves(database, f)
	redoableActions := printRedoableActions(database, f)

	if redoableMoves == 0 && redoableActions == 0 {
		fmt.Println("  📭 Nothing to redo in this scope.")
	}

	printHistorySummary(HistorySummary{
		Verb:    "Redo (preview)",
		Matched: redoableMoves + redoableActions,
		Skipped: moveSkipped + actionSkipped,
	})
}

func countMatchedRedoMoves(database *db.DB, f ScopeFilter) int {
	raw, _ := database.ListMoveHistory(50)
	return countRevertedMoves(FilterMovesWith(raw, f))
}

func countMatchedRedoActions(database *db.DB, f ScopeFilter) int {
	raw, _ := database.ListActions(200)
	return countReverted(FilterActionsWith(raw, f))
}

func countRedoableMoveSkipped(database *db.DB, f ScopeFilter) int {
	raw, _ := database.ListMoveHistory(50)
	kept := FilterMovesWith(raw, f)
	return countScopeSkipped(countRevertedMoves(raw), countRevertedMoves(kept))
}

func countRedoableActionSkipped(database *db.DB, f ScopeFilter) int {
	raw, _ := database.ListActions(200)
	kept := FilterActionsWith(raw, f)
	return countScopeSkipped(countReverted(raw), countReverted(kept))
}

func countRevertedMoves(moves []db.MoveRecord) int {
	count := 0
	for _, m := range moves {
		if m.IsReverted {
			count++
		}
	}
	return count
}

func printRedoableMoves(database *db.DB, f ScopeFilter) int {
	rawMoves, _ := database.ListMoveHistory(50)
	moves := FilterMovesWith(rawMoves, f)
	count := 0
	for _, m := range moves {
		if m.IsReverted {
			count++
		}
	}
	if count > 0 {
		fmt.Println("  📁 Moves / Renames:")
		for _, m := range moves {
			if !m.IsReverted {
				continue
			}
			fmt.Printf("    [move-%d]  %s → %s  (%s)\n", m.ID, m.FromPath, m.ToPath, m.MovedAt)
		}
		fmt.Println()
	}
	return count
}

func printRedoableActions(database *db.DB, f ScopeFilter) int {
	rawActions, _ := database.ListActions(200)
	actions := FilterActionsWith(rawActions, f)
	count := countReverted(actions)
	if count == 0 {
		return 0
	}
	fmt.Println("  📋 Actions:")
	for _, a := range actions {
		if !a.IsReverted {
			continue
		}
		fmt.Printf("    [action-%d]  %s  %s  (%s%s)\n",
			a.ActionHistoryId, a.FileActionId, actionDetail(a), a.CreatedAt, batchSuffix(a.BatchId))
	}
	fmt.Println()
	return count
}

func redoActionByID(database *db.DB, scanner *bufio.Scanner, id int64) {
	action, err := database.GetActionByID(id)
	if err != nil {
		errlog.Error("Cannot find action %d: %v", id, err)
		return
	}
	if !action.IsReverted {
		fmt.Printf("⚠️  Action %d is not reverted — nothing to redo.\n", id)
		return
	}

	fmt.Printf("⏩ Redo action %d (%s):\n", action.ActionHistoryId, action.FileActionId)
	if action.Detail != "" {
		fmt.Printf("   %s\n", action.Detail)
	}
	if !confirmRedo(scanner) {
		return
	}

	if err := executeActionRedo(database, action); err != nil {
		errlog.Error("Redo action %d failed: %v", id, err)
		return
	}
	fmt.Printf("✅ Action %d redone successfully.\n", action.ActionHistoryId)
}

func redoMoveByID(database *db.DB, scanner *bufio.Scanner, id int64) {
	moves, err := database.ListMoveHistory(1000)
	if err != nil {
		errlog.Error("Cannot read move history: %v", err)
		return
	}
	var target *db.MoveRecord
	for i := range moves {
		if moves[i].ID == id {
			target = &moves[i]
			break
		}
	}
	if target == nil {
		errlog.Error("Move %d not found.", id)
		return
	}
	if !target.IsReverted {
		fmt.Printf("⚠️  Move %d is not reverted — nothing to redo.\n", id)
		return
	}

	fmt.Println("⏩ Redo move:")
	fmt.Printf("   %s → %s\n", target.FromPath, target.ToPath)
	if !confirmRedo(scanner) {
		return
	}

	if err := executeMoveRedo(database, target); err != nil {
		errlog.Error("Redo move %d failed: %v", id, err)
		return
	}
	fmt.Printf("✅ Move %d redone successfully.\n", target.ID)
}

func redoLastBatch(database *db.DB, scanner *bufio.Scanner, f ScopeFilter) {
	batchID := findLastRevertedBatchInScope(database, f)
	if batchID == "" {
		fmt.Println("📭 No reverted batch operations to redo in this scope.")
		return
	}

	batchActions, err := database.ListActionsByBatch(batchID)
	if err != nil {
		errlog.Error("Cannot read batch %s: %v", batchID, err)
		return
	}

	scoped := FilterActionsWith(batchActions, f)
	redoable := countReverted(scoped)
	skipped := countScopeSkipped(countReverted(batchActions), redoable)
	if redoable == 0 {
		fmt.Println("📭 Batch has no reverted actions to redo (or all filtered out).")
		printHistorySummary(HistorySummary{Verb: "Redo", Skipped: skipped})
		return
	}

	shortBatch := batchID
	if len(shortBatch) > 8 {
		shortBatch = shortBatch[:8]
	}

	fmt.Printf("⏩ Redo batch %s (%d actions, %d skipped by filter):\n",
		shortBatch, redoable, skipped)
	printRevertedActions(scoped)
	if !confirmRedo(scanner) {
		return
	}

	failed := executeRedoBatch(database, scoped)
	printRedoBatchResult(shortBatch, redoable, failed)
	printHistorySummary(HistorySummary{
		Verb:     "Redo",
		Matched:  redoable,
		Executed: redoable - failed,
		Failed:   failed,
		Skipped:  skipped,
	})
}

func redoLastOperation(database *db.DB, scanner *bufio.Scanner, f ScopeFilter) {
	lastMove := pickLastRedoableMove(database, f)
	lastAction := pickLastRedoableAction(database, f)
	skipped := countRedoableMoveSkipped(database, f) +
		countRedoableActionSkipped(database, f)

	haveMove := lastMove != nil
	haveAction := lastAction != nil

	if !haveMove && !haveAction {
		fmt.Println("📭 No reverted operations to redo in this scope.")
		printHistorySummary(HistorySummary{Verb: "Redo", Skipped: skipped})
		return
	}

	if haveMove && !haveAction {
		runSingleRedoMove(database, scanner, lastMove, skipped)
		return
	}

	if haveAction && !haveMove {
		runSingleRedoAction(database, scanner, lastAction, skipped)
		return
	}

	// Both available — pick the newest reverted
	if lastAction.CreatedAt >= lastMove.MovedAt {
		runSingleRedoAction(database, scanner, lastAction, skipped)
		return
	}
	runSingleRedoMove(database, scanner, lastMove, skipped)
}

func runSingleRedoMove(database *db.DB, scanner *bufio.Scanner, m *db.MoveRecord, skipped int) {
	executed, failed := 0, 0
	if !redoSingleMoveOK(database, scanner, m) {
		failed = 1
	} else {
		executed = 1
	}
	printHistorySummary(HistorySummary{
		Verb: "Redo", Matched: 1, Executed: executed, Failed: failed, Skipped: skipped,
	})
}

func runSingleRedoAction(database *db.DB, scanner *bufio.Scanner, a *db.ActionRecord, skipped int) {
	executed, failed := 0, 0
	if !redoSingleActionOK(database, scanner, a) {
		failed = 1
	} else {
		executed = 1
	}
	printHistorySummary(HistorySummary{
		Verb: "Redo", Matched: 1, Executed: executed, Failed: failed, Skipped: skipped,
	})
}

func redoSingleMoveOK(database *db.DB, scanner *bufio.Scanner, m *db.MoveRecord) bool {
	fmt.Println("⏩ Redo last move:")
	fmt.Printf("   %s → %s\n", m.FromPath, m.ToPath)
	if !confirmRedo(scanner) {
		return false
	}
	if err := executeMoveRedo(database, m); err != nil {
		errlog.Error("Redo failed: %v", err)
		return false
	}
	fmt.Println("✅ Redo successful!")
	return true
}

func redoSingleActionOK(database *db.DB, scanner *bufio.Scanner, a *db.ActionRecord) bool {
	printActionRedo(a)
	if !confirmRedo(scanner) {
		return false
	}
	if err := executeActionRedo(database, a); err != nil {
		errlog.Error("Redo failed: %v", err)
		return false
	}
	fmt.Println("✅ Redo successful!")
	return true
}

// pickLastRedoableMove returns the newest reverted move under scope.
func pickLastRedoableMove(database *db.DB, f ScopeFilter) *db.MoveRecord {
	moves, err := database.ListMoveHistory(200)
	if err != nil {
		return nil
	}
	for i := range moves {
		m := moves[i]
		if !m.IsReverted {
			continue
		}
		if !MoveInScope(m, f.Dir) {
			continue
		}
		if f.HasGlobs() && !MoveMatchesGlobs(m, f) {
			continue
		}
		return &m
	}
	return nil
}

// pickLastRedoableAction returns the newest reverted action under scope.
func pickLastRedoableAction(database *db.DB, f ScopeFilter) *db.ActionRecord {
	actions, err := database.ListActions(200)
	if err != nil {
		return nil
	}
	for i := range actions {
		a := actions[i]
		if !a.IsReverted {
			continue
		}
		if !ActionInScope(a, f.Dir) {
			continue
		}
		if f.HasGlobs() && !ActionMatchesGlobs(a, f) {
			continue
		}
		return &a
	}
	return nil
}

// findLastRevertedBatchInScope finds the most recent reverted batch where at
// least one action passes the scope+glob filter. Empty filter → unfiltered.
func findLastRevertedBatchInScope(database *db.DB, f ScopeFilter) string {
	actions, err := database.ListActions(200)
	if err != nil {
		return ""
	}
	for _, a := range actions {
		if !a.IsReverted || a.BatchId == "" {
			continue
		}
		if !batchTouchesScope(database, a.BatchId, f) {
			continue
		}
		return a.BatchId
	}
	return ""
}

