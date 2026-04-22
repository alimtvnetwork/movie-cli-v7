---
name: Undo / Redo scope rule
description: movie undo/redo default to cwd scope. Optional [path] arg overrides cwd. --global removes the filter. Any path on the action that lives under scope counts as a match.
type: feature
---

# Scoped undo / redo (v2.139.0)

## Behavior
- `movie undo`              → scope = cwd
- `movie undo /some/dir`    → scope = /some/dir
- `movie undo --global`     → no scope (legacy behavior)
- Same three forms apply to `movie redo`.
- The scope filter applies to `--list`, `--batch`, and the default
  "undo/redo last operation" flow. `--id` and `--move-id` always operate
  on the explicit record (no scope check, the user already named it).

## Match rule
An action / move is **in scope** when ANY of its stored paths is rooted
under the scope dir. We check:
- `MoveRecord.FromPath`, `MoveRecord.ToPath`
- Every string value found inside `ActionRecord.MediaSnapshot` (decoded
  as a generic `map[string]interface{}`, walked recursively)
- `ActionRecord.Detail` as a fallback (some legacy rows store the path
  in the human-readable detail string)

This handles every snapshot shape we currently emit:
- `popout compact` → `{original_path, compact_path}`
- `scan add / delete / restore` → full Media JSON including `file_path`
- future shapes — no code changes needed as long as the path lives in a
  string field somewhere in the JSON

## Files
- `cmd/path_scope.go` — `scopeFromArgs`, `pathInScope`, `MoveInScope`,
  `ActionInScope`, `FilterMoves`, `FilterActions`
- `cmd/path_scope_test.go` — unit coverage incl. prefix-collision case
- `cmd/movie_undo.go` / `cmd/movie_redo.go` — accept `[path]` positional
  + `--global` flag, pass scope into handlers
- `cmd/movie_undo_handlers.go` — `pickLastUndoableMove/Action`,
  `findLastUndoableBatch(scope)`, `batchTouchesScope`
- `cmd/movie_redo_handlers.go` — `pickLastRedoableMove/Action`,
  `findLastRevertedBatchInScope`

## Why
Running `movie undo` from inside a project folder used to surface the
single newest action across the whole DB — often something the user did
in a completely different directory hours earlier. Defaulting to cwd
matches the popout / move / scan defaults set by the CWD-default rule
(see `mem://constraints/cwd-default-rule`).