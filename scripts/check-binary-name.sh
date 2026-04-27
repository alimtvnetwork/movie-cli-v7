#!/usr/bin/env bash
# check-binary-name.sh — Verify (and optionally auto-fix) the user-facing
# binary name across the repo. Expected name: "movie".
#
# Modes:
#   (default) check  — report violations, exit 1 if any
#   --dry-run        — show what --fix would change, exit 0
#   --fix            — rewrite files in place, print summary, exit 0
#   --verbose, -v    — extra detail in check mode
#
# Replacement rules (case-sensitive, applied in this order):
#     MAHIN_   -> MOVIE_         (env var prefixes)
#     Mahin    -> Movie          (TitleCase)
#     mahin    -> movie          (lowercase, including paths/user-agents)
#
# Exit codes: 0 ok | 1 violations (check mode) | 2 bad usage
#
# Usage:
#   bash scripts/check-binary-name.sh
#   bash scripts/check-binary-name.sh --dry-run
#   bash scripts/check-binary-name.sh --fix

set -uo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
EXPECTED="movie"
MODE="check"
VERBOSE=0

for arg in "$@"; do
    case "$arg" in
        --fix)        MODE="fix" ;;
        --dry-run)    MODE="dry-run" ;;
        --verbose|-v) VERBOSE=1 ;;
        -h|--help)    sed -n '2,22p' "$0"; exit 0 ;;
        *) echo "Unknown arg: $arg" >&2; exit 2 ;;
    esac
done

cd "$ROOT"

# Split string literals so this script doesn't match itself.
LEGACY_LC='m''ahin'
LEGACY_TC='M''ahin'
LEGACY_UC='M''AHIN'

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

# Find every file containing any casing of the legacy token.
find_offending_files() {
    grep -rl -E "${EXCLUDES[@]}" -e "${LEGACY_LC}" -e "${LEGACY_TC}" -e "${LEGACY_UC}" . 2>/dev/null || true
}

# Show all violation lines for check mode.
list_violations() {
    grep -rn -E "${EXCLUDES[@]}" -e "${LEGACY_LC}" -e "${LEGACY_TC}" -e "${LEGACY_UC}" . 2>/dev/null || true
}

# Apply substitutions to a single file. Returns count of replaced occurrences.
fix_file() {
    local f="$1"
    local before after
    before=$(grep -c -E "${LEGACY_LC}|${LEGACY_TC}|${LEGACY_UC}" "$f" 2>/dev/null || echo 0)
    [ "$before" -eq 0 ] && { echo 0; return; }
    # Order matters: uppercase first (so MAHIN_ doesn't get partially mangled),
    # then TitleCase, then lowercase.
    sed -i \
        -e "s/${LEGACY_UC}_/MOVIE_/g" \
        -e "s/${LEGACY_UC}/MOVIE/g" \
        -e "s/${LEGACY_TC}/Movie/g" \
        -e "s/${LEGACY_LC}/${EXPECTED}/g" \
        "$f"
    after=$(grep -c -E "${LEGACY_LC}|${LEGACY_TC}|${LEGACY_UC}" "$f" 2>/dev/null || echo 0)
    echo "$((before - after))"
}

# ─── Check mode ───────────────────────────────────────────────
if [ "$MODE" = "check" ]; then
    violations=$(list_violations)
    if [ -z "$violations" ]; then
        echo "✅ Binary name check passed: all references use '${EXPECTED}'."
        exit 0
    fi
    count=$(printf '%s\n' "$violations" | wc -l | tr -d ' ')
    echo "❌ Binary name check failed: ${count} occurrence(s) of banned legacy name."
    echo "   Expected user-facing binary name: '${EXPECTED}'"
    echo "   Run: bash scripts/check-binary-name.sh --fix"
    echo ""
    while IFS= read -r line; do
        [ -z "$line" ] && continue
        file=$(printf '%s' "$line" | cut -d: -f1)
        lineno=$(printf '%s' "$line" | cut -d: -f2)
        content=$(printf '%s' "$line" | cut -d: -f3-)
        echo "::error file=${file},line=${lineno}::Banned binary name (use '${EXPECTED}'): ${content}"
    done <<< "$violations"
    exit 1
fi

# ─── Dry-run mode ─────────────────────────────────────────────
if [ "$MODE" = "dry-run" ]; then
    files=$(find_offending_files)
    if [ -z "$files" ]; then
        echo "✅ Nothing to fix — repo already uses '${EXPECTED}' everywhere."
        exit 0
    fi
    echo "── Dry-run: planned replacements ─────────────────────────────"
    echo "  ${LEGACY_UC}_ → MOVIE_   |   ${LEGACY_UC} → MOVIE"
    echo "  ${LEGACY_TC}  → Movie    |   ${LEGACY_LC} → ${EXPECTED}"
    echo ""
    total=0
    fcount=0
    while IFS= read -r f; do
        [ -z "$f" ] && continue
        n=$(grep -c -E "${LEGACY_LC}|${LEGACY_TC}|${LEGACY_UC}" "$f" 2>/dev/null || echo 0)
        printf "  %4d  %s\n" "$n" "$f"
        total=$((total + n))
        fcount=$((fcount + 1))
    done <<< "$files"
    echo ""
    echo "Would update ${fcount} file(s), ${total} occurrence(s)."
    echo "Run with --fix to apply."
    exit 0
fi

# ─── Fix mode ─────────────────────────────────────────────────
files=$(find_offending_files)
if [ -z "$files" ]; then
    echo "✅ Nothing to fix — repo already uses '${EXPECTED}' everywhere."
    exit 0
fi

echo "── Auto-fix: rewriting banned tokens to '${EXPECTED}' ────────"
total_changes=0
files_changed=0
while IFS= read -r f; do
    [ -z "$f" ] && continue
    changed=$(fix_file "$f")
    if [ "$changed" -gt 0 ]; then
        printf "  ✓ %4d replaced  %s\n" "$changed" "$f"
        total_changes=$((total_changes + changed))
        files_changed=$((files_changed + 1))
    fi
done <<< "$files"

echo ""
echo "── Summary ───────────────────────────────────────────────────"
echo "  Files updated : ${files_changed}"
echo "  Replacements  : ${total_changes}"

# Verify nothing slipped through.
remaining=$(list_violations)
if [ -n "$remaining" ]; then
    rcount=$(printf '%s\n' "$remaining" | wc -l | tr -d ' ')
    echo "  Remaining     : ${rcount} (manual review needed)"
    echo ""
    echo "$remaining" | head -20
    exit 1
fi
echo "  Remaining     : 0"
echo "✅ Repo is clean. Don't forget to bump version/info.go."
exit 0
