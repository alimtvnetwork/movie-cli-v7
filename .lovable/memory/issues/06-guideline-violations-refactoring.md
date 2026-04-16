# Issue: Guideline Violations Audit & Nested-If Refactoring

> **Status**: ✅ Resolved (Phase 1-3 of 7)  
> **Severity**: Medium  
> **Iteration**: 3 (16-Apr-2026)

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
- Created comprehensive audit report in `.lovable/memory/audit/01-guideline-violations`
- Catalogued all violations by category with file locations

### Phase 2: Nested-If Elimination
- Refactored top 20 worst files using early returns and guard clauses
- Extracted helper functions for complex conditions
- Created `cmd/movie_scan_helpers_print.go` for extracted print logic
- Version bumped through v2.3.0 → v2.4.0+

### Phase 3: Magic String → Constants (v2.24.0)
- Replaced 3 remaining raw `"Database error: %v"` strings with `msgDatabaseError` constant
- Files fixed: `movie_scan.go`, `movie_rescan.go`, `movie_ls_table.go`
- Switched from `fmt.Fprintf(os.Stderr, ...)` to `errlog.Error(...)` for consistency
- All magic string violations now resolved (0 remaining)

## Remaining Phases (⏳ Pending)

- Phase 4: fmt.Errorf → apperror.Wrap()
- Phase 5: Oversized functions split
- Phase 6: Oversized files split
- Phase 7: Final consistency pass

## Learning

- Guideline enforcement should happen continuously, not in batch audits
- Early returns dramatically simplify control flow
- Guard clauses at function top reduce cognitive load
