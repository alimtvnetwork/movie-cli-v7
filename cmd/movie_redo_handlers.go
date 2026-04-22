// movie_redo_handlers.go — redo subcommand handlers (list, by-id, batch, last).
package cmd

import (
	"bufio"
	"fmt"

	"github.com/alimtvnetwork/movie-cli-v5/db"
	"github.com/alimtvnetwork/movie-cli-v5/errlog"
)

func showRedoableList(database *db.DB, scope string) {
	fmt.Println("⏩ Recent redoable operations")
	if scope != "" {
		fmt.Printf("   scope: %s\n", scope)
	}
	fmt.Println()

	redoableMoves := printRedoableMoves(database, scope)
	redoableActions := printRedoableActions(database, scope)

	if redoableMoves == 0 && redoableActions == 0 {
		fmt.Println("  📭 Nothing to redo in this scope.")
	}
}

func printRedoableMoves(database *db.DB, scope string) int {
	rawMoves, _ := database.ListMoveHistory(50)
	moves := FilterMoves(rawMoves, scope)
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

func printRedoableActions(database *db.DB, scope string) int {
	rawActions, _ := database.ListActions(200)
	actions := FilterActions(rawActions, scope)
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

func redoLastBatch(database *db.DB, scanner *bufio.Scanner, scope string) {
	batchID := findLastRevertedBatchInScope(database, scope)
	if batchID == "" {
		fmt.Println("📭 No reverted batch operations to redo in this scope.")
		return
	}

	batchActions, err := database.ListActionsByBatch(batchID)
	if err != nil {
		errlog.Error("Cannot read batch %s: %v", batchID, err)
		return
	}

	redoable := countReverted(batchActions)
	if redoable == 0 {
		fmt.Println("📭 Batch has no reverted actions to redo.")
		return
	}

	shortBatch := batchID
	if len(shortBatch) > 8 {
		shortBatch = shortBatch[:8]
	}

	fmt.Printf("⏩ Redo batch %s (%d actions):\n", shortBatch, redoable)
	printRevertedActions(batchActions)
	if !confirmRedo(scanner) {
		return
	}

	failed := executeRedoBatch(database, batchActions)
	printRedoBatchResult(shortBatch, redoable, failed)
}

func redoLastOperation(database *db.DB, scanner *bufio.Scanner, scope string) {
	lastMove := pickLastRedoableMove(database, scope)
	lastAction := pickLastRedoableAction(database, scope)

	haveMove := lastMove != nil
	haveAction := lastAction != nil

	if !haveMove && !haveAction {
		fmt.Println("📭 No reverted operations to redo in this scope.")
		return
	}

	if haveMove && !haveAction {
		redoSingleMove(database, scanner, lastMove)
		return
	}

	if haveAction && !haveMove {
		redoSingleAction(database, scanner, lastAction)
		return
	}

	// Both available — pick the newest reverted
	if lastAction.CreatedAt >= lastMove.MovedAt {
		redoSingleAction(database, scanner, lastAction)
		return
	}
	redoSingleMove(database, scanner, lastMove)
}

// pickLastRedoableMove returns the newest reverted move under scope.
func pickLastRedoableMove(database *db.DB, scope string) *db.MoveRecord {
	moves, err := database.ListMoveHistory(200)
	if err != nil {
		return nil
	}
	for i := range moves {
		m := moves[i]
		if !m.IsReverted {
			continue
		}
		if !MoveInScope(m, scope) {
			continue
		}
		return &m
	}
	return nil
}

// pickLastRedoableAction returns the newest reverted action under scope.
func pickLastRedoableAction(database *db.DB, scope string) *db.ActionRecord {
	actions, err := database.ListActions(200)
	if err != nil {
		return nil
	}
	for i := range actions {
		a := actions[i]
		if !a.IsReverted {
			continue
		}
		if !ActionInScope(a, scope) {
			continue
		}
		return &a
	}
	return nil
}

// findLastRevertedBatchInScope finds the most recent reverted batch where at
// least one action touches the scope dir. scope == "" → unfiltered.
func findLastRevertedBatchInScope(database *db.DB, scope string) string {
	actions, err := database.ListActions(200)
	if err != nil {
		return ""
	}
	for _, a := range actions {
		if !a.IsReverted || a.BatchId == "" {
			continue
		}
		if !batchTouchesScope(database, a.BatchId, scope) {
			continue
		}
		return a.BatchId
	}
	return ""
}

func redoSingleMove(database *db.DB, scanner *bufio.Scanner, m *db.MoveRecord) {
	fmt.Println("⏩ Redo last move:")
	fmt.Printf("   %s → %s\n", m.FromPath, m.ToPath)
	if !confirmRedo(scanner) {
		return
	}
	if err := executeMoveRedo(database, m); err != nil {
		errlog.Error("Redo failed: %v", err)
		return
	}
	fmt.Println("✅ Redo successful!")
}

func redoSingleAction(database *db.DB, scanner *bufio.Scanner, a *db.ActionRecord) {
	printActionRedo(a)
	if !confirmRedo(scanner) {
		return
	}
	if err := executeActionRedo(database, a); err != nil {
		errlog.Error("Redo failed: %v", err)
		return
	}
	fmt.Println("✅ Redo successful!")
}
