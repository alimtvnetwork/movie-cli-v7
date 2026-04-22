---
name: Undo / Redo scope rule
description: movie undo/redo default to cwd. [path] arg overrides. --global disables dir filter. --include / --exclude apply repeatable globs. --list shows resolved scope + matched/skipped per kind.
type: feature
---

# Scoped undo / redo (v2.142.0)

## Behavior
- `movie undo`                          → scope = cwd
- `movie undo /some/dir`                → scope = /some/dir
- `movie undo --global`                 → no dir scope (legacy)
- `movie undo --include '*.mkv'`        → keep only matching paths
- `movie undo --exclude 'Trash'`        → drop matching paths
- `--include` / `--exclude` repeatable, combine with the dir scope.
- Same forms apply to `movie redo`.
- Filter applies to `--list`, `--batch`, and the default last-op flow.
  `--id` / `--move-id` always operate on the explicit record.

## --list output (v2.142.0)
The list view now always prints a 3-block header:

```
⏪ Recent undoable operations
   scope:    /home/me/movies/2024/    (or "<global> (no directory filter)")
   include:  *.mkv
   exclude:  Trash
   matched:  4  (1 moves, 3 actions)
   skipped:  9  (2 moves, 7 actions)

  📁 Moves / Renames:
    [move-12]  /a → /b  (2024-04-21T18:33:00+08:00)

  📋 Actions:
    [action-44]  scan_add  Inception (2010)  (2024-04-21T18:30:00+08:00  batch:9f3ac01a)

📊 Undo (preview) summary
   ✅ executed:               0
   ⚠️  failed:                0
   🔍 matched filter:         4
   🚫 skipped (out of scope): 9
```

- The `scope:` line is always printed (`<global>` label when --global).
- `matched` / `skipped` lines split moves vs actions so the user can
  tell at a glance which kind got dropped by their filter.
- The trailing `(preview)` summary block stays for consistency with
  real undo/redo runs.

## Match rule (dir scope)
Action / move is in scope when ANY stored path is rooted under the
scope dir. Checks `MoveRecord.FromPath`, `MoveRecord.ToPath`, every
string value inside `ActionRecord.MediaSnapshot` (decoded as
`map[string]interface{}` and walked recursively), plus `Detail` as a
legacy fallback.

## Glob rule
`filepath.Match` syntax (`*`, `?`, `[class]`). Each pattern tried
against full path, basename, and every ancestor basename. Excludes
evaluated first; excludes always win.

## Post-op summary
Every undo/redo exit prints:

```
📊 Undo summary
   ✅ executed:               3
   ⚠️  failed:                0
   🔍 matched filter:         3
   🚫 skipped (out of scope): 7
```

## Files
- `cmd/path_scope.go` — `ScopeFilter`, `buildScopeFilter`,
  `printScopeBanner`, `scopeLabel`, `printScopeMatchedCounts`,
  filter helpers
- `cmd/path_scope_test.go` — prefix-collision + globs + snapshot match
- `cmd/history_summary.go` — `HistorySummary`, `printHistorySummary`,
  `countScopeSkipped`
- `cmd/history_summary_test.go` — clamp coverage
- `cmd/movie_undo.go` / `cmd/movie_redo.go` — `[path]` positional +
  `--global` / `--include` / `--exclude` flags
- `cmd/movie_undo_handlers.go` — `runSingleUndoMove/Action`,
  `count*Skipped`, `countMatchedUndo*`, summary at every exit
- `cmd/movie_redo_handlers.go` — `runSingleRedoMove/Action`,
  `count*Skipped`, `countMatchedRedo*`, summary at every exit