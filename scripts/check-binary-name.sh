#!/usr/bin/env bash
# check-binary-name.sh — Verify (and optionally auto-fix) the user-facing
# binary name across the repo. Expected name: "movie".
#
# Modes:
#   (default) check  — report violations, exit 1 if any
#   --dry-run        — show what --fix would change, exit 0
#   --fix            — rewrite files in place, print summary, exit 0
#   --json PATH      — also write JSON summary to PATH (default with --fix:
#                      /mnt/documents/binary-name-fix-summary.json if writable,
#                      else .lovable/reports/binary-name-fix-summary.json)
#   --verbose, -v    — extra detail in check mode
#
# JSON schema (fix mode):
#   {
#     "expected": "movie",
#     "timestamp": "2026-04-27T12:34:56+08:00",
#     "files_changed": N, "total_replacements": N, "remaining": N,
#     "files": [
#       { "path": "...", "before": N, "after": N, "replaced": N,
#         "by_pattern": { "UPPER_PREFIX": N, "UPPER": N, "TITLE": N, "LOWER": N } }
#     ]
#   }
#
# Replacement rules (case-sensitive, applied in this order). The legacy
# token is referred to here as <LEGACY> so this file does not itself contain
# the banned literal and trip the broader CI guard:
#     <LEGACY>_  (UPPER) -> MOVIE_     (env var prefixes)
#     <Legacy>   (Title) -> Movie      (TitleCase)
#     <legacy>   (lower) -> movie      (paths, user-agents, CLI examples)
#
# Exit codes: 0 ok | 1 violations (check mode) | 2 bad usage
#
# Usage:
#   bash scripts/check-binary-name.sh
#   bash scripts/check-binary-name.sh --dry-run
#   bash scripts/check-binary-name.sh --fix
#   bash scripts/check-binary-name.sh --fix --json /tmp/summary.json

set -uo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
EXPECTED="movie"
MODE="check"
VERBOSE=0
JSON_PATH=""
FUZZY=0

while [ $# -gt 0 ]; do
    case "$1" in
        --fix)        MODE="fix" ;;
        --dry-run)    MODE="dry-run" ;;
        --fuzzy)      FUZZY=1 ;;
        --verbose|-v) VERBOSE=1 ;;
        --json)       JSON_PATH="${2:-}"; shift ;;
        --json=*)     JSON_PATH="${1#--json=}" ;;
        -h|--help)    sed -n '2,36p' "$0"; exit 0 ;;
        *) echo "Unknown arg: $1" >&2; exit 2 ;;
    esac
    shift
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

# Apply substitutions to a single file. Echoes 6 space-separated integers:
#   replaced before after upper_prefix upper title lower
# (per-pattern counts measured *before* the in-place rewrite).
fix_file() {
    local f="$1"
    local before after up_prefix up tl lo
    before=$(grep -cE "${LEGACY_LC}|${LEGACY_TC}|${LEGACY_UC}" "$f" 2>/dev/null | tr -d '[:space:]')
    : "${before:=0}"
    if [ "$before" -eq 0 ]; then
        echo "0 0 0 0 0 0 0"; return
    fi
    # Per-pattern occurrence counts (sum of occurrences, not matched lines).
    up_prefix=$(grep -oE "${LEGACY_UC}_" "$f" 2>/dev/null | wc -l | tr -d '[:space:]')
    up=$(grep -oE "${LEGACY_UC}[^_]|${LEGACY_UC}\$" "$f" 2>/dev/null | wc -l | tr -d '[:space:]')
    tl=$(grep -oE "${LEGACY_TC}" "$f" 2>/dev/null | wc -l | tr -d '[:space:]')
    lo=$(grep -oE "${LEGACY_LC}" "$f" 2>/dev/null | wc -l | tr -d '[:space:]')
    : "${up_prefix:=0}"; : "${up:=0}"; : "${tl:=0}"; : "${lo:=0}"
    sed -i \
        -e "s/${LEGACY_UC}_/MOVIE_/g" \
        -e "s/${LEGACY_UC}/MOVIE/g" \
        -e "s/${LEGACY_TC}/Movie/g" \
        -e "s/${LEGACY_LC}/${EXPECTED}/g" \
        "$f"
    after=$(grep -cE "${LEGACY_LC}|${LEGACY_TC}|${LEGACY_UC}" "$f" 2>/dev/null | tr -d '[:space:]')
    : "${after:=0}"
    echo "$((before - after)) ${before} ${after} ${up_prefix} ${up} ${tl} ${lo}"
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
        n=$(grep -cE "${LEGACY_LC}|${LEGACY_TC}|${LEGACY_UC}" "$f" 2>/dev/null | tr -d '[:space:]')
        : "${n:=0}"
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

# Optional fuzzy pre-pass: handle whitespace/formatting variants
# (m a h i n, m-a-h-i-n, m_a_h_i_n, m.a.h.i.n, zero-width-joined, etc.)
# before the strict sed pass mops up any remaining canonical occurrences.
fuzzy_replaced=0
fuzzy_files=0
if [ "$FUZZY" -eq 1 ]; then
    echo "── Fuzzy pre-pass: normalizing whitespace/formatting variants ──"
    # Build a candidate file list: anything text-like, minus excluded dirs
    # and the self-exclude list. Use git ls-files when available for speed
    # and accuracy; fall back to find.
    if command -v git >/dev/null 2>&1 && [ -d .git ]; then
        candidates=$(git ls-files 2>/dev/null)
    else
        candidates=$(find . -type f \
            -not -path '*/.git/*' -not -path '*/.release/*' \
            -not -path '*/node_modules/*' -not -path '*/dist/*' \
            -not -path '*/build/*' -not -path '*/.gitmap/*' \
            2>/dev/null | sed 's|^\./||')
    fi
    # Exclude self files by basename.
    self_basenames=( "ci.yml" "check-binary-name.sh" "_fuzzy_rewrite.py"
                     "guard-forbidden-terms.sh" "audit-legacy-paths.sh"
                     "rename-acronyms.py" "check-acronym-naming.py"
                     "CHANGELOG.md" "banned-legacy-name.md" )
    filtered=$(printf '%s\n' "$candidates" | while IFS= read -r p; do
        [ -z "$p" ] && continue
        skip=0
        bn=$(basename "$p")
        for sb in "${self_basenames[@]}"; do
            [ "$bn" = "$sb" ] && { skip=1; break; }
        done
        [ "$skip" -eq 0 ] && printf '%s\n' "$p"
    done)
    # Run the Python rewriter; it only touches files that actually match.
    if [ -n "$filtered" ]; then
        fuzzy_out=$(printf '%s\n' "$filtered" | python3 scripts/_fuzzy_rewrite.py - 2>/dev/null || true)
        fuzzy_err=$(printf '%s\n' "$filtered" | python3 scripts/_fuzzy_rewrite.py - 2>&1 >/dev/null || true)
        if [ -n "$fuzzy_out" ]; then
            while IFS= read -r jl; do
                [ -z "$jl" ] && continue
                fp=$(printf '%s' "$jl" | python3 -c "import sys,json;d=json.loads(sys.stdin.read());print(d['path'],d['replaced'])")
                printf "  ~ fuzzy  %s\n" "$fp"
                fuzzy_files=$((fuzzy_files + 1))
            done <<< "$fuzzy_out"
            fuzzy_replaced=$(printf '%s' "$fuzzy_err" | python3 -c "import sys,json;
data=sys.stdin.read().strip().splitlines()
print(json.loads(data[-1]).get('total_replaced',0) if data else 0)" 2>/dev/null || echo 0)
        fi
        : "${fuzzy_replaced:=0}"
        echo "  Fuzzy files   : ${fuzzy_files}"
        echo "  Fuzzy replaced: ${fuzzy_replaced}"
    fi
    echo ""
fi

files=$(find_offending_files)
if [ -z "$files" ]; then
    echo "✅ Nothing to fix — repo already uses '${EXPECTED}' everywhere."
    # Still emit an empty JSON summary if requested.
    if [ -n "$JSON_PATH" ] || [ "$MODE" = "fix" ]; then
        : # fall through to JSON write below with zero entries
    else
        exit 0
    fi
fi

echo "── Auto-fix: rewriting banned tokens to '${EXPECTED}' ────────"
total_changes=0
files_changed=0
# Accumulate per-file JSON entries in a temp file (avoid quoting hell).
entries_tmp=$(mktemp)
trap 'rm -f "$entries_tmp"' EXIT

while IFS= read -r f; do
    [ -z "$f" ] && continue
    read -r changed before after up_prefix up tl lo <<< "$(fix_file "$f")"
    if [ "${changed:-0}" -gt 0 ]; then
        printf "  ✓ %4d replaced  %s  (UPPER_=%d UPPER=%d Title=%d lower=%d)\n" \
            "$changed" "$f" "$up_prefix" "$up" "$tl" "$lo"
        total_changes=$((total_changes + changed))
        files_changed=$((files_changed + 1))
        # Escape backslash and double-quote for JSON.
        esc_path=$(printf '%s' "$f" | sed 's/\\/\\\\/g; s/"/\\"/g')
        printf '    {"path":"%s","before":%d,"after":%d,"replaced":%d,"by_pattern":{"UPPER_PREFIX":%d,"UPPER":%d,"TITLE":%d,"LOWER":%d}}\n' \
            "$esc_path" "$before" "$after" "$changed" "$up_prefix" "$up" "$tl" "$lo" \
            >> "$entries_tmp"
    fi
done <<< "$files"

echo ""
echo "── Summary ───────────────────────────────────────────────────"
echo "  Files updated : ${files_changed}"
echo "  Replacements  : ${total_changes}"

# Verify nothing slipped through.
remaining=$(list_violations)
remaining_count=0
if [ -n "$remaining" ]; then
    remaining_count=$(printf '%s\n' "$remaining" | wc -l | tr -d ' ')
    echo "  Remaining     : ${remaining_count} (manual review needed)"
    echo ""
    echo "$remaining" | head -20
else
    echo "  Remaining     : 0"
fi

# ─── Write JSON summary ───────────────────────────────────────
# Default JSON path when --fix is used without explicit --json.
if [ -z "$JSON_PATH" ]; then
    if [ -d /mnt/documents ] && [ -w /mnt/documents ]; then
        JSON_PATH="/mnt/documents/binary-name-fix-summary.json"
    else
        mkdir -p .lovable/reports 2>/dev/null || true
        JSON_PATH=".lovable/reports/binary-name-fix-summary.json"
    fi
fi

ts=$(date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo "unknown")
{
    printf '{\n'
    printf '  "expected": "%s",\n' "$EXPECTED"
    printf '  "timestamp": "%s",\n' "$ts"
    printf '  "files_changed": %d,\n' "$files_changed"
    printf '  "total_replacements": %d,\n' "$total_changes"
    printf '  "remaining": %d,\n' "$remaining_count"
    printf '  "patterns": {\n'
    printf '    "UPPER_PREFIX": "<LEGACY>_ -> MOVIE_",\n'
    printf '    "UPPER":        "<LEGACY> -> MOVIE",\n'
    printf '    "TITLE":        "<Legacy> -> Movie",\n'
    printf '    "LOWER":        "<legacy> -> %s"\n' "$EXPECTED"
    printf '  },\n'
    printf '  "files": [\n'
    if [ -s "$entries_tmp" ]; then
        # Join entries with commas.
        sed '$!s/$/,/' "$entries_tmp"
    fi
    printf '  ]\n'
    printf '}\n'
} > "$JSON_PATH"

echo "  JSON summary  : ${JSON_PATH}"

[ "$remaining_count" -gt 0 ] && exit 1
echo "✅ Repo is clean. Don't forget to bump version/info.go."
exit 0

