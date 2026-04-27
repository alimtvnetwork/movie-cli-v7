#!/usr/bin/env bash
# check-binary-name.sh — Verify the user-facing binary name is consistently
# "movie" across all repo files, and that no banned legacy aliases appear.
#
# This complements the broader "Forbidden term guard" in CI by focusing
# specifically on user-facing binary-name occurrences (CLI examples, version
# banners, env vars, paths, user-agents, install scripts).
#
# Exit codes:
#   0  all clear
#   1  one or more violations found
#   2  bad usage
#
# Usage:
#   bash scripts/check-binary-name.sh
#   bash scripts/check-binary-name.sh --verbose

set -uo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
EXPECTED="movie"
VERBOSE=0

for arg in "$@"; do
    case "$arg" in
        --verbose|-v) VERBOSE=1 ;;
        -h|--help)    sed -n '2,18p' "$0"; exit 0 ;;
        *) echo "Unknown arg: $arg" >&2; exit 2 ;;
    esac
done

cd "$ROOT"

# Patterns we treat as user-facing binary-name occurrences. If any of these
# appear with a non-"movie" identifier, it's a violation. Split string
# literals so this script doesn't trip the broader CI grep.
LEGACY='m''ahin'

EXCLUDES=(
    --exclude-dir=.git
    --exclude-dir=.release
    --exclude-dir=.gitmap
    --exclude-dir=node_modules
    --exclude-dir=dist
    --exclude-dir=build
    --exclude=CHANGELOG.md
    --exclude=check-binary-name.sh
)

# Patterns: anything that looks like a binary invocation, env-var prefix,
# user-agent, /tmp install path, dotfolder convention, or version banner.
PATTERNS=(
    "/tmp/${LEGACY}\b"
    "\b${LEGACY}-cli\b"
    "\b${LEGACY^^}_[A-Z]"
    "\.${LEGACY}/"
    "\b${LEGACY} v[0-9]"
    "Binary: \`${LEGACY}\`"
    "go build -o /tmp/${LEGACY}\b"
)

violations=""
for pat in "${PATTERNS[@]}"; do
    hits=$(grep -rn -E -i "${EXCLUDES[@]}" -e "$pat" . 2>/dev/null || true)
    [ -n "$hits" ] && violations="${violations}${hits}"$'\n'
done

# Also catch any bare legacy-name token in user-facing files.
bare=$(grep -rn -E -i "${EXCLUDES[@]}" \
    --include='*.md' --include='*.sh' --include='*.ps1' \
    --include='*.go' --include='Makefile' --include='*.yml' \
    -e "\b${LEGACY}\b" . 2>/dev/null || true)
[ -n "$bare" ] && violations="${violations}${bare}"$'\n'

violations=$(printf '%s' "$violations" | grep -v '^$' || true)

if [ -z "$violations" ]; then
    echo "✅ Binary name check passed: all user-facing references use '${EXPECTED}'."
    [ "$VERBOSE" -eq 1 ] && {
        echo ""
        echo "Sample of valid '${EXPECTED}' references (first 5):"
        grep -rn -E "\b${EXPECTED} (v[0-9]|version|scan|ls|stats)" \
            "${EXCLUDES[@]}" --include='*.md' . 2>/dev/null | head -5 || true
    }
    exit 0
fi

count=$(printf '%s\n' "$violations" | wc -l | tr -d ' ')
echo "❌ Binary name check failed: ${count} occurrence(s) of banned legacy name."
echo "   Expected user-facing binary name: '${EXPECTED}'"
echo ""
while IFS= read -r line; do
    [ -z "$line" ] && continue
    file=$(printf '%s' "$line" | cut -d: -f1)
    lineno=$(printf '%s' "$line" | cut -d: -f2)
    content=$(printf '%s' "$line" | cut -d: -f3-)
    echo "::error file=${file},line=${lineno}::Banned binary name (use '${EXPECTED}'): ${content}"
done <<< "$violations"
echo ""
echo "Fix: replace every occurrence with '${EXPECTED}' (or 'MOVIE_' for env vars,"
echo "'/tmp/${EXPECTED}' for paths, '${EXPECTED}-cli' for user-agents, '.${EXPECTED}/' for dotfolders)."
exit 1
