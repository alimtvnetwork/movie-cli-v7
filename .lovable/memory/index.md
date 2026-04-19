# Memory: index.md
Updated: now

# Project Memory

## Core
Go 1.22 CLI project (NOT web). Binary: `mahin`. Ignore Lovable build errors.
One file per command, max ~200 lines. Shared helpers in movie_info.go and movie_resolve.go.
File naming: `01-name-of-file.md`. Keep folder file counts small.
Plans & suggestions tracked in single files, not per-item files.
Never modify `.release` folder. Any code change bumps at least minor version.
ALWAYS bump version/version.go after every code change. Never forget.
Malaysia timezone (UTC+8) for timestamps. Milestones in `readm.txt`.
Root spec files: lowercase (spec.md, ai-handoff.md, development-log.md). Keep README.md uppercase.
Spec resequenced: foundation 01-06, app at 08, app-issues at 09. Issues in spec/09-app-issues/.
Error spec flattened: spec/02-error-manage-spec/ (no nested subfolder).
HTML JS: single API_BASE variable for all REST calls. Never repeat URL.
Boolean names: never use negative words (un/not/no). Use positive semantic synonyms with Is/Has prefix.
Zero nested if. Max 2 conditions per if. No else after return. Functions ≤15 lines. Files ≤300 lines. Max 3 params.
No magic strings — use constants/enums. No fmt.Errorf — use apperror.Wrap().
Updater rule: in update mode, deploy target = active PATH binary, NEVER powershell.json deployPath.
Updater rule: never write expected os.Remove failures on *-update-* artifacts to stderr (PowerShell turns it into NativeCommandError).
American English ONLY in code/comments/CHANGELOG/spec — misspell uses US locale. behaviour→behavior, optimised→optimized, catalogued→cataloged. Full table in ci-cd playbook.
Acronym MixedCaps: Json/Imdb/Tmdb/Api/Http/Url/Sql/Html/Xml in Go identifiers — NEVER JSON/IMDb/TMDb/API/HTTP/URL/SQL/HTML/XML. Project rule overrides Effective Go. Bare 2-letter `ID` and trailing locals (`imdbID`, `tmdbID`, `imgURL`) exempted.
CI lint failures: every recurring lint error is logged in spec/12-ci-cd-pipeline/05-ci-cd-issues/ — read before fixing similar errors.
Release workflow MUST verify all 6 archives present before upload (see spec/12-ci-cd-pipeline/05-ci-cd-issues/06). Never publish partial releases.


## Memories
- [Project overview](mem://01-project-overview) — Go CLI, command tree, architecture, file structure
- [Conventions](mem://02-conventions) — Code style, naming, build, deploy, config keys
- [Plan](mem://workflow/01-plan) — Done/pending task tracker, prioritized backlog
- [AI success plan](mem://workflow/01-ai-success-plan) — 7 rules for 98% AI success rate
- [Suggestions](mem://suggestions/01-suggestions) — Active suggestion tracker with priority levels
- [Reliability report](mem://reports/01-reliability-risk-report) — Failure map, corrective actions, readiness decision
- [Guideline violations audit](mem://audit/01-guideline-violations) — Full audit: nested ifs, magic strings, oversized funcs/files, 7-phase fix plan
- [Version bump rule](mem://preferences/version-bump) — Always bump version after every code change
- [API base variable](mem://preferences/api-base-variable) — JS must use single API_BASE variable, never repeat URL
- [Boolean naming](mem://constraints/boolean-no-negative-words) — IsUndone→IsReverted; never use un/not/no in boolean names
- [Updater scope](mem://constraints/updater-scope) — Go updater never runs git/build; all git+build belongs in run.ps1
- [Acronym MixedCaps](mem://constraints/acronym-mixedcaps) — Json/Imdb/Tmdb/Api/Http/Url, never JSON/IMDb/etc. Spec issue 05
- [CI/CD build fixes playbook](mem://ci-cd/01-build-fixes-playbook) — All recurring gofmt/govet/misspell/acronym errors with prevention rules
- [Timestamp bug](mem://issues/01-timestamp-bug) — Fixed: hardcoded "now" → RFC3339
- [Duplicate TMDb fetch](mem://issues/02-duplicate-tmdb-fetch) — Fixed: shared helpers
- [Large files](mem://issues/03-large-files) — Fixed: split to <200 lines
- [CI log commit loop](mem://issues/04-ci-log-commit-loop) — Constraint: CI log commits must never trigger new runs; kill feature if loops occur
- [Updater async console](mem://issues/05-updater-async-console) — Updater output and worker handoff sequencing
- [Updater deploy-path mismatch](mem://issues/06-updater-deploypath-mismatch) — Fixed v2.121.0: update target = PATH binary, not deployPath
- [Updater stale-handoff full RCA](mem://issues/07-updater-stale-handoff-loop-full-rca) — Full chain: 5 bugs, why v2.118-v2.120 invisible, v2.121 root-cause + v2.122 self-replace bootstrap + v2.123 polish
