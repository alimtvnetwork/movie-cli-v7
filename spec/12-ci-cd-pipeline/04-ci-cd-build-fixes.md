# CI/CD Build Fixes — Recurring Lint & Build Failures

> **Purpose**: This document is a *self-contained playbook* of every CI/CD lint and build failure we have hit, with **root cause analysis**, **exact fix patterns**, and **prevention rules**. Any AI model or engineer reading this file MUST be able to avoid these mistakes on the first try.
>
> **Audience**: AI code generators (Lovable, Copilot, Cursor, etc.) and human contributors.
>
> **Hard rule**: Before committing any Go code in this repo, mentally walk this checklist. If a rule below is violated, the CI pipeline (`golangci-lint v1.64.8`) WILL fail.

---

## How to Use This Document

1. When CI fails, find the matching error category below.
2. Apply the fix pattern exactly — do not improvise.
3. If a *new* class of error appears, add a new section here with: error sample → root cause → fix pattern → prevention rule.
4. Bump version (`version/info.go`) after every fix (per `mem://preferences/version-bump`).

---

## Global Prevention Checklist (run before every commit)

- [ ] All `.go` files pass `gofmt -s -w` (no manual import grouping)
- [ ] Imports: stdlib block first, blank line, then third-party + internal block (goimports order)
- [ ] String concatenation has spaces around `+` when joining identifiers (`` `foo ` + bar + ` baz` ``)
- [ ] No assignment mismatch — check signature of every function called with `:=` or `=`
- [ ] No unused imports (especially after refactors that remove the only usage)
- [ ] Range loops over slices of structs ≥128 bytes use index or pointer (`for i := range s`)
- [ ] Inner `err` variables in nested `if` blocks are renamed (e.g. `addErr`, `markErr`) to avoid shadow
- [ ] Struct fields are ordered largest → smallest pointer/word size to satisfy `fieldalignment`
- [ ] Boolean field/column names use positive form with `Is`/`Has` prefix (no `un/not/no`)
- [ ] Functions ≤15 lines, files ≤300 lines, max 3 params, zero nested if, max 2 conditions per if
- [ ] No `fmt.Errorf` — use `apperror.Wrap()` / `apperror.New()`
- [ ] No magic strings — use constants

---

## Error Catalogue

### 1. `gofmt` — File is not properly formatted

#### 1a. Import order violation

**Sample error:**
```
db/action_history.go:5:1: File is not properly formatted (gofmt)
    "github.com/alimtvnetwork/movie-cli-v5/apperror"
```

**Root cause**: Internal package import (`apperror`) placed inside the stdlib block, or stdlib and third-party imports mixed without the required blank-line separator. `gofmt` (and `goimports`) require:

```
import (
    "fmt"               // stdlib block
    "os"
    "strings"
                        // ← blank line REQUIRED
    "github.com/alimtvnetwork/movie-cli-v5/apperror"   // external/internal block
    "github.com/alimtvnetwork/movie-cli-v5/db"
)
```

**Why it happened**: AI-generated edits inserted the new import alphabetically into the first block instead of the second.

**Fix pattern**:
1. Identify the two groups: stdlib (no `/` in path) vs external (has `/`).
2. Put exactly **one blank line** between them.
3. Sort each group alphabetically.
4. Run `gofmt -s -w <file>` if available.

**Prevention rule**: When adding ANY import that contains `/` (i.e., not stdlib), it MUST go below a blank line, never inside the stdlib block.

---

#### 1b. Struct field alignment whitespace

**Sample error:**
```
db/media.go:25:1: File is not properly formatted (gofmt)
    ID         int64
```

**Root cause**: `gofmt` aligns struct fields by inserting spaces so that all type columns line up *within the same contiguous group*. A blank line breaks the group. When field types differ in length (e.g. `int64` vs `int`), the alignment must be uniform across all consecutive lines.

**Fix pattern**: Never hand-align struct fields. Run `gofmt`. If gofmt is unavailable, use **single space** between name and type and let CI tell you the canonical spacing, then match it.

**Prevention rule**: Do not insert tabs or extra spaces inside struct definitions manually. Let `gofmt` own that whitespace.

---

#### 1c. Struct tag column alignment

**Sample error:**
```
db/media.go:28:1: File is not properly formatted (gofmt)
    FileSize      int64   `json:"file_size,omitempty"`
```

**Root cause**: Extra spaces between the type and the backtick struct tag. Within a contiguous block of tagged fields, gofmt aligns the tag column to the **single space** after the longest type — extra padding is rejected.

**Fix pattern**: Exactly one space between type and `` ` `` unless gofmt aligns the column. Re-run gofmt.

---

#### 1d. String concatenation spacing — TRUST gofmt, DO NOT GUESS

**Sample error:**
```
db/media_query.go:11:1: File is not properly formatted (gofmt)
    rows, err := d.Query(`SELECT ` + mediaColumns + `
```

**Root cause**: `gofmt` has a specific rule for `+` between a raw string literal and an identifier on the same line: it wants **NO spaces** around `+`:

```go
// ✅ CORRECT (what gofmt produces)
d.Query(`SELECT `+mediaColumns+` FROM media`)

// ❌ WRONG (gofmt will rewrite this)
d.Query(`SELECT ` + mediaColumns + ` FROM media`)
```

This is the OPPOSITE of normal arithmetic `+` (where spaces are required). The rule applies specifically when one operand is a string literal directly adjacent to the `+`.

**Prevention rule**: NEVER manually adjust spacing around `+` in string concatenation. Run `gofmt -w` and accept whatever it produces. Bulk `sed` rewrites of these patterns are forbidden — they will fight gofmt.

**The only safe fix command:**
```bash
gofmt -s -w .
```


---

### 2. `assignment mismatch: 1 variable but X returns 2 values`

**Sample error:**
```
cmd/movie_scan_loop.go:93:7: assignment mismatch: 1 variable but database.InsertActionSimple returns 2 values
```

**Root cause**: A function signature was changed (e.g. `InsertActionSimple` was upgraded from `error` return to `(int64, error)`), but call sites were not updated.

**Fix pattern**:
```go
// ❌ WRONG (after signature change)
err := database.InsertActionSimple(...)

// ✅ CORRECT
_, err := database.InsertActionSimple(...)   // if id unused
id, err := database.InsertActionSimple(...)  // if id used
```

**Prevention rule**: When changing a function's return arity, **grep all call sites in the same commit** before pushing:
```bash
grep -rn "InsertActionSimple(" --include="*.go"
```

---

### 3. `imported and not used`

**Sample error:**
```
cmd/movie_history_table.go:8:2: "github.com/alimtvnetwork/movie-cli-v5/db" imported and not used
```

**Root cause**: A refactor removed the last reference to a package, but the import line was left behind.

**Fix pattern**: Delete the import line. If `goimports` is available, it does this automatically.

**Prevention rule**: After deleting any function call or type reference, scan the file's imports and remove orphans.

---

### 4. `rangeValCopy` (gocritic) — copying large structs in range

**Sample error:**
```
cmd/movie_rest_report.go:111:2: rangeValCopy: each iteration copies 368 bytes (consider pointers or indexing) (gocritic)
    for _, m := range items {
```

**Root cause**: Ranging over `[]Media` (or any struct ≥128 bytes) by value copies the entire struct each iteration. `gocritic` flags this above ~128 bytes.

**Fix pattern**:
```go
// ❌ WRONG
for _, m := range items {
    use(m.Title)
}

// ✅ OPTION A — index
for i := range items {
    use(items[i].Title)
}

// ✅ OPTION B — slice of pointers (only if upstream allows)
for _, m := range itemPtrs {
    use(m.Title)
}
```

**Prevention rule**: When the loop variable is a struct (not a primitive, string, or pointer), default to index-based iteration.

---

### 5. `shadow` (govet) — inner `err` shadows outer `err`

**Sample error:**
```
db/media_test.go:208:5: shadow: declaration of "err" shadows declaration at line 200 (govet)
    if err := d.AddTag(int(id), "favorite"); err == nil {
```

**Root cause**: An outer `err` is still in scope and the inner `if err := …` redeclaration shadows it, hiding the outer error from later checks.

**Fix pattern**: Rename the inner variable to describe the operation:
```go
// ❌ WRONG
err := d.InsertMedia(...)
if err := d.AddTag(...); err != nil { ... }

// ✅ CORRECT
err := d.InsertMedia(...)
if tagErr := d.AddTag(...); tagErr != nil { ... }
```

**Prevention rule**: Inside any block where `err` is already declared, never reuse `err` in `if x := …`. Use a descriptive name (`addErr`, `markErr`, `fetchErr`).

---

### 6. `fieldalignment` (govet) — struct memory layout

**Sample error:**
```
cmd/types.go:280:24: fieldalignment: struct with 48 pointer bytes could be 40 (govet)
type AppendUniqueInput struct {
```

**Root cause**: Struct fields are ordered such that padding is inserted between fields of differing alignment. Reordering fields largest → smallest (by `unsafe.Sizeof`) packs them tightly.

**Fix pattern**: Reorder fields by descending size:
```
pointers/slices/maps/strings (16-24 bytes) → int64/float64 (8) → int32/float32 (4) → int16 (2) → bool/byte (1)
```

```go
// ❌ WRONG
type Episode struct {
    ID      int64
    Title   string   // pointer-bearing
    Watched bool
    Runtime int64
}

// ✅ CORRECT
type Episode struct {
    Title   string   // 16 bytes (ptr+len)
    ID      int64    // 8
    Runtime int64    // 8
    Watched bool     // 1
}
```

**Prevention rule**: When defining a new struct, list `string`/slice/map/pointer fields first, then `int64/float64`, then smaller ints, then `bool`/`byte` last.

---

## Recurrence Log

| Date | Version | Errors fixed | Files touched |
|------|---------|--------------|---------------|
| 2026-04-15 | v2.83.0 | assignment mismatch ×6, unused import ×1 | cmd/movie_*.go |
| 2026-04-15 | v2.83.1 | rangeValCopy ×3, shadow ×3 | cmd/, db/media_test.go |
| 2026-04-16 | v2.83.1 | gofmt ×5, fieldalignment ×3 | cmd/, db/ |
| 2026-04-16 | v2.83.2 | gofmt import order ×7 | db/, cmd/movie_resolve.go |
| 2026-04-16 | v2.83.3 | gofmt struct tag + concat spacing | db/media.go, db/*query/cleanup |

When a *new* error class appears that is not catalogued above, append it to the catalogue (don't just log it here) so it cannot recur silently.

---

## Authoritative Tooling Versions

These MUST match `.github/workflows/ci.yml`:

| Tool | Version |
|------|---------|
| `golangci-lint` | `v1.64.8` |
| `govulncheck` | `v1.1.4` |
| `go-winres` | `v0.3.3` |
| Go toolchain | from `go.mod` |

---

## Local Pre-Push Command (recommended)

```bash
gofmt -s -w .
go vet ./...
golangci-lint run --timeout=5m
go test ./...
```

If all four pass locally, CI will pass.

---

*CI/CD build fixes spec — updated: 2026-04-16 — version: v2.83.3*
