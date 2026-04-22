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

## Cwd scope confirmation prompt (v2.143.0)
When the scope was inferred from cwd (no `[path]` arg, no `--global`)
and a destructive flow runs (`undo`, `undo --batch`, `redo`,
`redo --batch`), we now prompt:

```
🎯 Undo scope detected from current directory:
   /home/me/movies/2024/
   Use this scope?  [Y]es / [g]lobal / [n]o :
```

- `Enter` / `y` → proceed with cwd scope
- `g`           → switch to `--global` for this run
- anything else → cancel

`--list`, `--id`, `--move-id` skip the prompt — they're either read-only
or already explicit.

Implementation: `ScopeFilter.UserProvidedPath` is set true when the user
passed `[path]` or `--global`. `ConfirmCwdScope(scanner, f, verb)` in
`cmd/path_scope.go` short-circuits when that flag is true.

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

## Cwd scope confirmation prompt (v2.143.0)
When scope is inferred from cwd (no `[path]`, no `--global`), destructive
flows prompt before acting:

```
🎯 Undo scope detected from current directory:
   /home/me/movies/2024/
   Use this scope?  [Y]es / [g]lobal / [n]o :
```

`Enter`/`y` proceed, `g` switches to `--global`, anything else cancels.
`--list`, `--id`, `--move-id` skip the prompt. Implemented via
`ScopeFilter.UserProvidedPath` + `ConfirmCwdScope` in `cmd/path_scope.go`.

## Preview-summary block for `--list` (v2.144.0)
`movie undo --list` / `movie redo --list` no longer reuse
`printHistorySummary` (which printed misleading `executed: 0` /
`failed: 0` lines in preview mode). They now use a dedicated
`PreviewSummary` block with per-kind breakdowns:

```
📋 Undo preview (no changes made)
   🎯 ready to undo:    4   (1 moves, 3 actions)
   🚫 out of scope:     9   (2 moves, 7 actions)
   ℹ️  run `movie undo` (without --list) to actually undo.
```

Per-section headers also carry the matched count, e.g.
`📁 Moves / Renames  — 1 ready to undo:` so totals in the footer
can be cross-checked against each section.

Destructive flows (`undo`, `undo --batch`, `redo`, `redo --batch`)
keep using `printHistorySummary` so they still report executed/failed.

Implementation: `cmd/history_summary.go` — `PreviewSummary` struct,
`printPreviewSummary`. Wired into `showUndoableList` and
`showRedoableList`.