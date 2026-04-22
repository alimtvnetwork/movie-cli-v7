// movie_redo.go — movie redo: re-applies the last reverted operation.
//
// Supports redoing:
//   - File moves/renames  (from move_history, IsReverted=1)
//   - Action history ops  (from action_history, IsReverted=1)
//
// Flags:
//
//	--list           Show recent redoable actions
//	--id <id>        Redo a specific action_history record
//	--move-id <id>   Redo a specific move_history record
//	--batch          Redo the entire last reverted batch
package cmd

import (
	"bufio"
	"os"

	"github.com/spf13/cobra"

	"github.com/alimtvnetwork/movie-cli-v5/db"
	"github.com/alimtvnetwork/movie-cli-v5/errlog"
)

var (
	redoListFlag  bool
	redoBatchFlag bool
	redoGlobal    bool
	redoActionID  int64
	redoMoveID    int64
)

var movieRedoCmd = &cobra.Command{
	Use:   "redo [path]",
	Short: "Redo the last reverted operation",
	Long: `Re-applies the most recent reverted operation.

Scope:
  By default, only history rooted under the current working directory
  (or the optional [path] argument) is considered. Pass --global to
  redo across the entire database like in older versions.

Flags:
  --list           Show recent redoable actions
  --id <id>        Redo a specific action_history record by ID
  --move-id <id>   Redo a specific move_history record by ID
  --batch          Redo the entire last reverted batch
  --global         Ignore the cwd / [path] scope`,
	Run: runMovieRedo,
}

func init() {
	movieRedoCmd.Flags().BoolVar(&redoListFlag, "list", false, "Show recent redoable actions")
	movieRedoCmd.Flags().BoolVar(&redoBatchFlag, "batch", false, "Redo entire last reverted batch")
	movieRedoCmd.Flags().BoolVar(&redoGlobal, "global", false, "Ignore cwd / path scope")
	movieRedoCmd.Flags().Int64Var(&redoActionID, "id", 0, "Redo specific action_history record")
	movieRedoCmd.Flags().Int64Var(&redoMoveID, "move-id", 0, "Redo specific move_history record")
}

func runMovieRedo(cmd *cobra.Command, args []string) {
	database, err := db.Open()
	if err != nil {
		errlog.Error(msgDatabaseError, err)
		return
	}
	defer database.Close()

	scanner := bufio.NewScanner(os.Stdin)
	home, _ := os.UserHomeDir()
	scope := scopeFromArgs(args, home, redoGlobal)

	if redoListFlag {
		showRedoableList(database, scope)
		return
	}
	if redoActionID > 0 {
		redoActionByID(database, scanner, redoActionID)
		return
	}
	if redoMoveID > 0 {
		redoMoveByID(database, scanner, redoMoveID)
		return
	}
	if redoBatchFlag {
		redoLastBatch(database, scanner, scope)
		return
	}

	redoLastOperation(database, scanner, scope)
}
