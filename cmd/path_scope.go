// path_scope.go — shared path-scoping helpers for history commands
// (movie undo / movie redo / future history filters).
//
// Rule (mem://constraints/cwd-default-rule + this file):
//
//	movie undo  [path]   → scope = path or cwd. --global to override.
//	movie redo  [path]   → scope = path or cwd. --global to override.
//
// "In scope" means: ANY path stored on the action (FromPath, ToPath,
// MediaSnapshot.original_path / compact_path / file_path) is rooted under
// the resolved scope directory.
package cmd

import (
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/alimtvnetwork/movie-cli-v5/db"
)

// scopeFromArgs returns the resolved scope directory and a "global" flag.
// When isGlobal is true, the returned dir is "" and callers should not filter.
func scopeFromArgs(args []string, home string, isGlobal bool) string {
	if isGlobal {
		return ""
	}
	dir, err := ResolveTargetDir(args, home)
	if err != nil {
		return ""
	}
	return normalizeScope(dir)
}

// normalizeScope returns a clean absolute-style suffix-friendly form.
func normalizeScope(dir string) string {
	clean := filepath.Clean(dir)
	if !strings.HasSuffix(clean, string(filepath.Separator)) {
		clean += string(filepath.Separator)
	}
	return clean
}

// pathInScope reports whether p is the scope dir itself or sits under it.
func pathInScope(p, scope string) bool {
	if scope == "" || p == "" {
		return scope == ""
	}
	clean := filepath.Clean(p) + string(filepath.Separator)
	return strings.HasPrefix(clean, scope)
}

// MoveInScope returns true when either side of the move touches scope.
func MoveInScope(m db.MoveRecord, scope string) bool {
	if scope == "" {
		return true
	}
	return pathInScope(m.FromPath, scope) || pathInScope(m.ToPath, scope)
}

// ActionInScope inspects the snapshot JSON for any path field that lives
// under scope. We do NOT rely on a typed struct because different actions
// emit different snapshot shapes (compact uses original_path/compact_path,
// scan uses file_path inside Media, etc.).
func ActionInScope(a db.ActionRecord, scope string) bool {
	if scope == "" {
		return true
	}
	if a.Detail != "" && pathInScope(a.Detail, scope) {
		return true
	}
	return snapshotTouchesScope(a.MediaSnapshot, scope)
}

// snapshotTouchesScope decodes the snapshot as a generic map and walks every
// string value, returning true on the first one rooted under scope.
func snapshotTouchesScope(snapshot, scope string) bool {
	if snapshot == "" {
		return false
	}
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(snapshot), &raw); err != nil {
		return false
	}
	return mapHasScopedString(raw, scope)
}

func mapHasScopedString(m map[string]interface{}, scope string) bool {
	for _, v := range m {
		if valueHasScope(v, scope) {
			return true
		}
	}
	return false
}

func valueHasScope(v interface{}, scope string) bool {
	switch t := v.(type) {
	case string:
		return pathInScope(t, scope)
	case map[string]interface{}:
		return mapHasScopedString(t, scope)
	}
	return false
}

// FilterMoves returns only the moves rooted under scope.
func FilterMoves(moves []db.MoveRecord, scope string) []db.MoveRecord {
	if scope == "" {
		return moves
	}
	out := make([]db.MoveRecord, 0, len(moves))
	for _, m := range moves {
		if MoveInScope(m, scope) {
			out = append(out, m)
		}
	}
	return out
}

// FilterActions returns only the actions touching scope (any-path rule).
func FilterActions(actions []db.ActionRecord, scope string) []db.ActionRecord {
	if scope == "" {
		return actions
	}
	out := make([]db.ActionRecord, 0, len(actions))
	for _, a := range actions {
		if ActionInScope(a, scope) {
			out = append(out, a)
		}
	}
	return out
}