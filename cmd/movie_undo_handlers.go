// movie_undo_handlers.go — undo subcommand handlers (list, by-id, batch, last).
package cmd

import (
	"bufio"
	"fmt"

	"github.com/alimtvnetwork/movie-cli-v5/db"
	"github.com/alimtvnetwork/movie-cli-v5/errlog"
)

func showUndoableList(database *db.DB, f ScopeFilter) {
	fmt.Println("⏪ Recent undoable operations")
	printScopeBanner(f)
	fmt.Println()

	undoableMoves := printUndoableMoves(database, f)
	undoableActions := printUndoableActions(database, f)

	if undoableMoves == 0 && undoableActions == 0 {
		fmt.Println("  📭 Nothing to undo in this scope.")
	}
}

func printUndoableMoves(database *db.DB, f ScopeFilter) int {
	rawMoves, _ := database.ListMoveHistory(50)
	moves := FilterMovesWith(rawMoves, f)
	count := 0
	for _, m := range moves {
		if !m.IsReverted {
			count++
		}
	}
	if count > 0 {
		fmt.Println("  📁 Moves / Renames:")
		for _, m := range moves {
			if m.IsReverted {
				continue
			}
			fmt.Printf("    [move-%d]  %s → %s  (%s)\n", m.ID, m.FromPath, m.ToPath, m.MovedAt)
		}
		fmt.Println()
	}
	return count
}

func printUndoableActions(database *db.DB, f ScopeFilter) int {
	rawActions, _ := database.ListActions(100)
	actions := FilterActionsWith(rawActions, f)
	count := countNonReverted(actions)
	if count == 0 {
		return 0
	}
	fmt.Println("  📋 Actions:")
	for _, a := range actions {
		if a.IsReverted {
			continue
		}
		fmt.Printf("    [action-%d]  %s  %s  (%s%s)\n",
			a.ActionHistoryId, a.FileActionId, actionDetail(a), a.CreatedAt, batchSuffix(a.BatchId))
	}
	fmt.Println()
	return count
}

func countNonReverted(actions []db.ActionRecord) int {
	count := 0
	for _, a := range actions {
		if !a.IsReverted {
			count++
		}
	}
	return count
}

func actionDetail(a db.ActionRecord) string {
	if a.Detail != "" {
		return a.Detail
	}
	return a.FileActionId.String()
}

func batchSuffix(batchID string) string {
	if batchID == "" {
		return ""
	}
	short := batchID
	if len(short) > 8 {
		short = short[:8]
	}
	return fmt.Sprintf("  batch:%s", short)
}

func undoActionByID(database *db.DB, scanner *bufio.Scanner, id int64) {
	action, err := database.GetActionByID(id)
	if err != nil {
		errlog.Error("Cannot find action %d: %v", id, err)
		return
	}
	if action.IsReverted {
		fmt.Printf("⚠️  Action %d has already been reverted.\n", id)
		return
	}

	fmt.Printf("⏪ Undo action %d (%s):\n", action.ActionHistoryId, action.FileActionId)
	if action.Detail != "" {
		fmt.Printf("   %s\n", action.Detail)
	}
	if !confirmUndo(scanner) {
		return
	}

	if err := executeActionUndo(database, action); err != nil {
		errlog.Error("Undo action %d failed: %v", id, err)
		return
	}
	fmt.Printf("✅ Action %d reverted successfully.\n", action.ActionHistoryId)
}

func undoMoveByID(database *db.DB, scanner *bufio.Scanner, id int64) {
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
	if target.IsReverted {
		fmt.Printf("⚠️  Move %d has already been reverted.\n", id)
		return
	}

	fmt.Println("⏪ Undo move:")
	fmt.Printf("   %s → %s\n", target.ToPath, target.FromPath)
	if !confirmUndo(scanner) {
		return
	}

	if err := executeMoveUndo(database, target); err != nil {
		errlog.Error("Undo move %d failed: %v", id, err)
		return
	}
	fmt.Printf("✅ Move %d reverted successfully.\n", target.ID)
}

func undoLastBatch(database *db.DB, scanner *bufio.Scanner, f ScopeFilter) {
	batchID := findLastUndoableBatch(database, f)
	if batchID == "" {
		fmt.Println("📭 No batch operations to undo in this scope.")
		return
	}

	batchActions, err := database.ListActionsByBatch(batchID)
	if err != nil {
		errlog.Error("Cannot read batch %s: %v", batchID, err)
		return
	}

	undoable := countUndoable(batchActions)
	if undoable == 0 {
		fmt.Println("📭 Batch already reverted.")
		return
	}

	fmt.Printf("⏪ Undo batch %s (%d actions):\n", batchID[:8], undoable)
	printUndoableActionsList(batchActions)
	if !confirmUndo(scanner) {
		return
	}

	failed := executeUndoBatch(database, batchActions)
	printUndoBatchResult(batchID[:8], undoable, failed)
}

func undoLastOperation(database *db.DB, scanner *bufio.Scanner, f ScopeFilter) {
	lastMove := pickLastUndoableMove(database, f)
	lastAction := pickLastUndoableAction(database, f)

	haveMove := lastMove != nil
	haveAction := lastAction != nil

	if !haveMove && !haveAction {
		fmt.Println("📭 No operations to undo in this scope.")
		return
	}

	if haveMove && !haveAction {
		undoSingleMove(database, scanner, lastMove)
		return
	}

	if haveAction && !haveMove {
		undoSingleAction(database, scanner, lastAction)
		return
	}

	if lastAction.CreatedAt >= lastMove.MovedAt {
		undoSingleAction(database, scanner, lastAction)
		return
	}
	undoSingleMove(database, scanner, lastMove)
}

// pickLastUndoableMove returns the newest non-reverted move under scope.
func pickLastUndoableMove(database *db.DB, f ScopeFilter) *db.MoveRecord {
	moves, err := database.ListMoveHistory(200)
	if err != nil {
		return nil
	}
	for i := range moves {
		m := moves[i]
		if m.IsReverted {
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

// pickLastUndoableAction returns the newest non-reverted action under scope.
func pickLastUndoableAction(database *db.DB, f ScopeFilter) *db.ActionRecord {
	actions, err := database.ListActions(200)
	if err != nil {
		return nil
	}
	for i := range actions {
		a := actions[i]
		if a.IsReverted {
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

func undoSingleMove(database *db.DB, scanner *bufio.Scanner, m *db.MoveRecord) {
	fmt.Println("⏪ Last move operation:")
	fmt.Printf("   %s → %s\n", m.ToPath, m.FromPath)
	if !confirmUndo(scanner) {
		return
	}
	if err := executeMoveUndo(database, m); err != nil {
		errlog.Error("Undo failed: %v", err)
		return
	}
	fmt.Println("✅ Undo successful!")
}

func undoSingleAction(database *db.DB, scanner *bufio.Scanner, a *db.ActionRecord) {
	printActionUndo(a)
	if !confirmUndo(scanner) {
		return
	}
	if err := executeActionUndo(database, a); err != nil {
		errlog.Error("Undo failed: %v", err)
		return
	}
	fmt.Println("✅ Undo successful!")
}

func findLastUndoableBatch(database *db.DB, f ScopeFilter) string {
	actions, err := database.ListActions(200)
	if err != nil {
		errlog.Error("Cannot read action history: %v", err)
		return ""
	}
	for _, a := range actions {
		if a.IsReverted || a.BatchId == "" {
			continue
		}
		if !batchTouchesScope(database, a.BatchId, f) {
			continue
		}
		return a.BatchId
	}
	return ""
}

// batchTouchesScope returns true if any action in the batch passes the
// dir scope AND glob filter (when set). f.Dir == "" with no globs → true.
func batchTouchesScope(database *db.DB, batchID string, f ScopeFilter) bool {
	if f.Dir == "" && !f.HasGlobs() {
		return true
	}
	rows, err := database.ListActionsByBatch(batchID)
	if err != nil {
		return false
	}
	for _, a := range rows {
		if !ActionInScope(a, f.Dir) {
			continue
		}
		if !f.HasGlobs() || ActionMatchesGlobs(a, f) {
			return true
		}
	}
	return false
}

func countUndoable(actions []db.ActionRecord) int {
	count := 0
	for _, a := range actions {
		if !a.IsReverted {
			count++
		}
	}
	return count
}

func printUndoableActionsList(actions []db.ActionRecord) {
	for _, a := range actions {
		if a.IsReverted {
			continue
		}
		fmt.Printf("   • %s: %s\n", a.FileActionId, actionDetail(a))
	}
}

func executeUndoBatch(database *db.DB, actions []db.ActionRecord) int {
	failed := 0
	for i := len(actions) - 1; i >= 0; i-- {
		if actions[i].IsReverted {
			continue
		}
		if err := executeActionUndo(database, &actions[i]); err != nil {
			errlog.Warn("Failed to undo action %d: %v", actions[i].ActionHistoryId, err)
			failed++
		}
	}
	return failed
}

func printUndoBatchResult(shortBatch string, undoable, failed int) {
	if failed == 0 {
		fmt.Printf("✅ Batch %s reverted (%d actions).\n", shortBatch, undoable)
		return
	}
	fmt.Printf("⚠️  Batch %s: %d reverted, %d failed.\n", shortBatch, undoable-failed, failed)
}
