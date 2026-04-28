# Project Memory

## Core
Go 1.22 CLI project (NOT web). Binary: `movie`. Ignore Lovable build errors.
Legacy codename is PERMANENTLY BANNED everywhere — CI guard fails the build on any match.
One file per command, max ~200 lines. Shared helpers in movie_info.go and movie_resolve.go.
File naming: `01-name-of-file.md`. Keep folder file counts small.
Plans & suggestions tracked in single files, not per-item files.
Never modify `.release` folder. Any code change bumps at least minor version.
ALWAYS bump version/version.go after every code change. Never forget.
Malaysia timezone (UTC+8) for timestamps. No milestone file (readm.txt removed).
Root spec files: lowercase (spec.md, ai-handoff.md, development-log.md). Keep README.md uppercase.
Spec resequenced: foundation 01-06, app at 08, app-issues at 09. Issues in spec/09-app-issues/.
Error spec flattened: spec/02-error-manage-spec/ (no nested subfolder).
HTML JS: single API_BASE variable for all REST calls. Never repeat URL.
Boolean names: never use negative words (un/not/no). Use positive semantic synonyms with Is/Has prefix.
Zero nested if. Max 2 conditions per if. No else after return. Functions ≤15 lines. Files ≤300 lines. Max 3 params.
No magic strings — use constants/enums. No fmt.Errorf — use apperror.Wrap().
Questions: never bare. Always full context, layman language, recommended option + why + pros + cons.
Docs: edit root README.md FIRST, then propagate to QUICKSTART/spec/CONTRIBUTING. Sub-docs never lead.

## Memories
- [Question asking style](mem://preferences/question-asking-style) — Full context, layman phrasing, recommendation with pros/cons on every question
- [README structure](mem://preferences/readme-structure) — Root README order: title → badges → image → install (Windows first, then macOS/Linux) → why this exists; write README first then update sub-docs
- [Banned legacy name](mem://constraints/banned-legacy-name) — Legacy codename forbidden everywhere; CI guard enforces it
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
- [Timestamp bug](mem://issues/01-timestamp-bug) — Fixed: hardcoded "now" → RFC3339
- [Duplicate TMDb fetch](mem://issues/02-duplicate-tmdb-fetch) — Fixed: shared helpers
- [Large files](mem://issues/03-large-files) — Fixed: split to <200 lines
- [CI log commit loop](mem://issues/04-ci-log-commit-loop) — Constraint: CI log commits must never trigger new runs; kill feature if loops occur
