---
name: Undo / Redo scope rule
description: movie undo/redo default to cwd. [path] arg overrides. --global disables dir filter. --include / --exclude apply repeatable globs. Post-op summary reports executed/failed/matched/skipped counts.
type: feature
---

# Scoped undo / redo (v2.141.0)

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
  on the explicit record.

## Match rule (dir scope)
An action / move is **in scope** when ANY of its stored paths is rooted
under the scope dir. We check `MoveRecord.FromPath`, `MoveRecord.ToPath`,
every string value found inside `ActionRecord.MediaSnapshot` (decoded as
`map[string]interface{}` and walked recursively), and `Detail` as a
legacy fallback.

## Glob rule
Glob patterns use `filepath.Match` (POSIX shell: `*`, `?`, `[class]`).
Each pattern is tried against:
1. the full path (e.g. `/movies/2024/*.mkv`)
2. the basename (e.g. `*.mkv`)
3. every ancestor basename (e.g. `Trash`, `.temp`)

Logic: excludes evaluated first (any-match → drop), then includes
(any-match → keep; no includes → keep). Excludes always win.

## Post-op summary (v2.141.0)
Every undo/redo run prints a 4-line summary block at the end:

```
📊 Undo summary
   ✅ executed:               3
   ⚠️  failed:                0
   🔍 matched filter:         3
   🚫 skipped (out of scope): 7
```

- `matched`  → rows that passed dir scope + globs
- `executed` → rows actually applied (matched − failed)
- `failed`   → rows where the FS operation errored
- `skipped`  → rows the filter dropped (helps the user notice when
  their `[path]` / `--include` / `--exclude` removed work they expected
  to undo)

List flows (`--list`) print the same block with a `(preview)` suffix —
nothing is executed, only matched and skipped counts are populated.

## Files
- `cmd/path_scope.go` — `ScopeFilter`, `scopeFromArgs`, `buildScopeFilter`,
  `pathInScope`, `MoveInScope`, `ActionInScope`, `FilterMoves`,
  `FilterActions`, `FilterMovesWith`, `FilterActionsWith`,
  `MoveMatchesGlobs`, `ActionMatchesGlobs`, `printScopeBanner`
- `cmd/path_scope_test.go` — prefix-collision + glob include/exclude +
  snapshot basename match coverage
- `cmd/history_summary.go` — `HistorySummary`, `printHistorySummary`,
  `countScopeSkipped`
- `cmd/history_summary_test.go` — clamp coverage
- `cmd/movie_undo.go` / `cmd/movie_redo.go` — accept `[path]` positional
  + `--global`, `--include`, `--exclude` flags
- `cmd/movie_undo_handlers.go` — `runSingleUndoMove/Action`,
  `count*Skipped`, summary at every exit
- `cmd/movie_redo_handlers.go` — `runSingleRedoMove/Action`,
  `count*Skipped`, summary at every exit