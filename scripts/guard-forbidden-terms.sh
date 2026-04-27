#!/usr/bin/env bash
# guard-forbidden-terms.sh — Repo-wide grep guard for the banned legacy
# project name. Single source of truth used by both local devs and CI.
#
# Self-exclusions are computed via `basename` so the guard works the same
# whether run from the repo root, a subdir, or invoked with `./` prefixes.
# `grep --exclude=NAME` matches by basename only (GNU grep behavior),
# so a basename-derived list is path-independent.
#
# Exit codes: 0 clean | 1 violations found

set -uo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

# Banned patterns, case-insensitive. Split literals so this file itself
# does not contain the banned token.
patterns=( 'm''ahin' 'M''AHIN' )

# Files that legitimately mention the banned token (the guard itself, the
# auto-fixer, historical logs, the constraint memory). Listed by full path
# so renames are easy to track; only the basename is fed to grep.
self_files=(
    ".github/workflows/ci.yml"
    "scripts/guard-forbidden-terms.sh"
    "scripts/check-binary-name.sh"
    "scripts/audit-legacy-paths.sh"
    "scripts/rename-acronyms.py"
    "scripts/check-acronym-naming.py"
    "CHANGELOG.md"
    ".lovable/memory/constraints/banned-legacy-name.md"
)

exclude_args=()
for p in "${self_files[@]}"; do
    exclude_args+=( "--exclude=$(basename "$p")" )
    [ -f "$p" ] || echo "::warning::self-exclude target missing: $p" >&2
done

violations=$(grep -rn -E -i \
    --exclude-dir=.git --exclude-dir=.release --exclude-dir=node_modules \
    --exclude-dir=dist --exclude-dir=build --exclude-dir=.gitmap \
    "${exclude_args[@]}" \
    -e "$(IFS='|'; echo "${patterns[*]}")" . 2>/dev/null || true)

if [ -z "$violations" ]; then
    echo "✅ No forbidden legacy-name references."
    exit 0
fi

echo "Forbidden legacy-name references found:"
echo "$violations"
echo ""
while IFS= read -r line; do
    [ -z "$line" ] && continue
    file=$(printf '%s' "$line" | cut -d: -f1)
    lineno=$(printf '%s' "$line" | cut -d: -f2)
    content=$(printf '%s' "$line" | cut -d: -f3-)
    echo "::error file=${file},line=${lineno}::Forbidden legacy name: ${content}"
done <<< "$violations"
echo ""
echo "The legacy project name is permanently banned. Use 'movie' instead."
echo "Also banned: /tmp/<legacy>, <LEGACY>_DB, .<legacy>/, <legacy>-cli."
echo "Auto-fix:  bash scripts/check-binary-name.sh --fix"
exit 1
