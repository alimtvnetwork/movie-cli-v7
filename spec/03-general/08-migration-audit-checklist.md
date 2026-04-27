# Pre-release migration audit checklist

Run this before tagging any new release that follows a module rename
(or whenever you suspect stale module-path references).

## 1. Run the auditor

```bash
bash scripts/audit-legacy-paths.sh --strict
```

Exit 0 = clean. Exit 1 = ACTIVE references remain — fix before release.
CI runs the same command in the `lint` job; do not skip locally.

## 2. Surfaces the auditor must come back clean on

The auditor scans the entire repo, but these are the surfaces a release
actually breaks if they slip through. Eyeball them after a clean run:

### Module + build
- `go.mod` — `module` line and any `replace` directives
- `go.sum` — should regenerate cleanly with `go mod tidy`
- All Go imports under `cmd/`, `db/`, `tmdb/`, `apperror/`, `version/`, etc.

### User-facing install / docs
- `README.md` — badges, install one-liners, clone URL
- `QUICKSTART.md` — copy-paste blocks for Linux / macOS / Windows
- `install.sh`, `install.ps1`
- `get.sh`, `get.ps1`
- `verify.sh`, `verify.ps1`
- `bootstrap.sh`, `bootstrap.ps1`

### Build + release scripts
- `run.ps1` — `-ldflags` module path
- `scripts/*.sh`, `scripts/*.ps1`, `scripts/*.py`
- Anything that constructs a GitHub release-asset URL

### CI / automation
- `.github/workflows/*.yml` — guard patterns are HISTORICAL and OK,
  but action steps that clone or `go install` must use `-v7`

## 3. Variants the auditor flags

| Variant     | Example                                                | Fix                                       |
|-------------|--------------------------------------------------------|-------------------------------------------|
| VERSIONED   | `github.com/alimtvnetwork/movie-cli-v6`                | rename to `…-v7`                          |
| UNVERSIONED | `github.com/alimtvnetwork/movie-cli`                   | add `-v7` suffix                          |
| TAGGED      | `github.com/alimtvnetwork/movie-cli-v7@v1.2.3`         | drop the `@vX.Y.Z` pin                    |
| REPLACE     | `replace github.com/alimtvnetwork/movie-cli-v6 => …`   | remove or repoint the `replace` directive |

## 4. Allowed (HISTORICAL) zones

These can keep legacy mentions and never need rewriting:

- `CHANGELOG.md`
- `spec/**`
- `.lovable/**`
- `scripts/audit-legacy-paths.sh` (defines the patterns)
- `.github/workflows/ci.yml` (guard patterns name legacy paths to ban them)

## 5. Final pre-release sequence

```bash
go mod tidy
bash scripts/audit-legacy-paths.sh --strict
go build ./...
go test ./...
```

All four must succeed before bumping `version/info.go` and tagging.
