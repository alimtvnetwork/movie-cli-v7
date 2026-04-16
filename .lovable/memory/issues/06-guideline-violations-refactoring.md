# Issue: Guideline Violations Audit & Refactoring

> **Status**: ✅ Resolved (Phase 1-5 of 7)  
> **Severity**: Medium  
> **Iteration**: 4 (16-Apr-2026)

## Root Cause

Codebase accumulated 280+ violations of Go coding guidelines over incremental development:
- 50+ nested if statements
- Magic strings throughout
- Functions >15 lines
- `fmt.Errorf` usage instead of `apperror.Wrap()`
- `else` after `return`
- Files >300 lines

## Solution Applied

### Phase 1: Audit
- Created comprehensive audit report with violations catalogued by category

### Phase 2: Nested-If Elimination
- Refactored top 20 worst files using early returns and guard clauses
- Extracted helper functions, created `cmd/movie_scan_helpers_print.go`

### Phase 3: Magic String → Constants (v2.24.0)
- Replaced 3 raw `"Database error: %v"` with `msgDatabaseError` constant
- Switched to `errlog.Error()` for consistency

### Phase 4: fmt.Errorf → apperror.Wrap (v2.24.0)
- Already resolved — only `fmt.Errorf` remaining is inside `apperror/apperror.go` (correct)

### Phase 5: Oversized Functions Split (v2.27.0)
- Split `movie_discover.go` `runMovieDiscover` into smaller helpers
- `updater/run.go` and `movie_move.go` already compliant

## Remaining Phases (⏳ Pending)

- Phase 6: >3 params → option structs (19 functions)
- Phase 7: Final consistency pass

## Learning

- Guideline enforcement should happen continuously, not in batch audits
- Early returns dramatically simplify control flow
- Guard clauses at function top reduce cognitive load
