---
name: Undo / Redo scope rule
description: movie undo/redo default to cwd. Optional [path] arg overrides cwd. --global removes the dir filter. --include / --exclude apply repeatable globs on top.
type: feature
---

# Scoped undo / redo (v2.140.0)

## Behavior
- `movie undo`                          → scope = cwd
- `movie undo /some/dir`                → scope = /some/dir
- `movie undo --global`                 → no dir scope (legacy behavior)
- `movie undo --include '*.mkv'`        → keep only actions whose paths match
- `movie undo --exclude 'Trash'`        → drop actions whose paths match
- `--include` / `--exclude` are repeatable and combine with the dir scope.
- Same forms apply to `movie redo`.
- The scope filter applies to `--list`, `--batch`, and the default
  "undo/redo last operation" flow. `--id` and `--move-id` always operate
  on the explicit record (no scope check, the user already named it).

## Match rule (dir scope)
An action / move is **in scope** when ANY of its stored paths is rooted
under the scope dir. We check:
- `MoveRecord.FromPath`, `MoveRecord.ToPath`
- Every string value found inside `ActionRecord.MediaSnapshot` (decoded
  as a generic `map[string]interface{}`, walked recursively)
- `ActionRecord.Detail` as a fallback

## Glob rule
Glob patterns use `filepath.Match` (POSIX shell: `*`, `?`, `[class]`).
Each pattern is tried against:
1. the full path (e.g. `/movies/2024/*.mkv`)
2. the basename (e.g. `*.mkv`)
3. every ancestor basename (e.g. `Trash`, `.temp`)

So both `*.srt` and `Inception` work without the user needing to know
where the path ended up in the snapshot.

Logic: excludes evaluated first (any-match → drop), then includes
(any-match → keep; no includes → keep). Excludes always win.

## Files
- `cmd/path_scope.go` — `ScopeFilter`, `scopeFromArgs`, `buildScopeFilter`,
  `pathInScope`, `MoveInScope`, `ActionInScope`, `FilterMoves`,
  `FilterActions`, `FilterMovesWith`, `FilterActionsWith`,
  `MoveMatchesGlobs`, `ActionMatchesGlobs`, `printScopeBanner`
- `cmd/path_scope_test.go` — prefix-collision + glob include/exclude +
  snapshot basename match
- `cmd/movie_undo.go` / `cmd/movie_redo.go` — accept `[path]` positional
  + `--global`, `--include`, `--exclude` flags, build ScopeFilter,
  pass it into handlers
- `cmd/movie_undo_handlers.go` — `pickLastUndoableMove/Action(filter)`,
  `findLastUndoableBatch(filter)`, `batchTouchesScope(filter)`
- `cmd/movie_redo_handlers.go` — `pickLastRedoableMove/Action(filter)`,
  `findLastRevertedBatchInScope(filter)`

## Why
Dir scope alone is too coarse when a single project folder mixes
different formats (mkv vs srt vs subs vs .temp). Globs let the user
narrow further without having to remember individual action IDs.