#!/usr/bin/env bash
# audit-legacy-paths.sh — Scan the entire repo for legacy module paths
# (movie-cli-v1 .. movie-cli-v6) and print a categorized report.
#
# This is a READ-ONLY audit tool. It never modifies files.
#
# Categories:
#   HISTORICAL (allowed)  — references in CHANGELOG.md, spec/, .lovable/
#                           memory, and this script itself. These are
#                           legitimate records of past renames.
#   ACTIVE    (must fix)  — references in user-facing code, scripts, configs,
#                           README, QUICKSTART, install scripts, CI workflows,
#                           Go source, go.mod, and Makefile. Any of these
#                           pointing at v1-v6 is a regression.
#
# Usage:
#   bash scripts/audit-legacy-paths.sh
#   bash scripts/audit-legacy-paths.sh --strict   # exit 1 if any ACTIVE found
#   bash scripts/audit-legacy-paths.sh --json     # machine-readable output

set -uo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PATTERN='movie-cli-v[123456]\b'

STRICT=0
JSON=0
for arg in "$@"; do
    case "$arg" in
        --strict) STRICT=1 ;;
        --json)   JSON=1 ;;
        -h|--help)
            sed -n '2,20p' "$0"; exit 0 ;;
        *) echo "Unknown arg: $arg" >&2; exit 2 ;;
    esac
done

cd "$ROOT"

# Collect all matches, excluding directories that have no business here.
RAW="$(grep -rnE "$PATTERN" \
    --exclude-dir=.git \
    --exclude-dir=node_modules \
    --exclude-dir=.release \
    --exclude-dir=.gitmap \
    . 2>/dev/null || true)"

if [ -z "$RAW" ]; then
    if [ "$JSON" -eq 1 ]; then
        echo '{"historical":[],"active":[],"summary":{"historical":0,"active":0}}'
    else
        echo "✅ No legacy module path references (movie-cli-v1..v6) found anywhere."
    fi
    exit 0
fi

# Classify each line: HISTORICAL vs ACTIVE.
# HISTORICAL paths (allowed — these document past renames):
#   ./CHANGELOG.md
#   ./spec/...
#   ./.lovable/...
#   ./scripts/audit-legacy-paths.sh   (this file mentions the pattern itself)
HISTORICAL_FILTER='^\./(CHANGELOG\.md|spec/|\.lovable/|scripts/audit-legacy-paths\.sh)'

historical=""
active=""
while IFS= read -r line; do
    [ -z "$line" ] && continue
    if echo "$line" | grep -qE "$HISTORICAL_FILTER"; then
        historical="${historical}${line}"$'\n'
    else
        active="${active}${line}"$'\n'
    fi
done <<< "$RAW"

h_count=$(printf '%s' "$historical" | grep -c . || true)
a_count=$(printf '%s' "$active" | grep -c . || true)

if [ "$JSON" -eq 1 ]; then
    # Minimal hand-rolled JSON (no jq dependency).
    json_array() {
        local input="$1"
        local first=1
        printf '['
        while IFS= read -r line; do
            [ -z "$line" ] && continue
            file=$(printf '%s' "$line" | cut -d: -f1)
            lineno=$(printf '%s' "$line" | cut -d: -f2)
            content=$(printf '%s' "$line" | cut -d: -f3- | sed 's/\\/\\\\/g; s/"/\\"/g')
            [ "$first" -eq 0 ] && printf ','
            printf '{"file":"%s","line":%s,"content":"%s"}' "$file" "$lineno" "$content"
            first=0
        done <<< "$input"
        printf ']'
    }
    printf '{"historical":'
    json_array "$historical"
    printf ',"active":'
    json_array "$active"
    printf ',"summary":{"historical":%s,"active":%s}}\n' "$h_count" "$a_count"
    [ "$STRICT" -eq 1 ] && [ "$a_count" -gt 0 ] && exit 1
    exit 0
fi

# Plain-text categorized report.
echo ""
echo "════════════════════════════════════════════════════════════════"
echo "  Legacy module-path audit — pattern: ${PATTERN}"
echo "  Root: ${ROOT}"
echo "════════════════════════════════════════════════════════════════"
echo ""
echo "── HISTORICAL (allowed) ─────────────────  ${h_count} match(es)"
echo "  Files: CHANGELOG.md, spec/**, .lovable/**, this script."
echo ""
if [ "$h_count" -gt 0 ]; then
    printf '%s' "$historical" | sed 's/^/  /'
    echo ""
fi

echo "── ACTIVE (must fix) ─────────────────────  ${a_count} match(es)"
echo "  Anything outside the historical zone — README, scripts, Go code,"
echo "  go.mod, CI workflows, install scripts, etc."
echo ""
if [ "$a_count" -gt 0 ]; then
    printf '%s' "$active" | sed 's/^/  /'
    echo ""
    echo "❌ ${a_count} active legacy reference(s) found."
    echo "   Replace each with: github.com/alimtvnetwork/movie-cli-v7"
else
    echo "  ✅ none"
fi
echo ""
echo "── Summary ──────────────────────────────────────────────────────"
echo "  Historical : ${h_count}  (informational, OK to keep)"
echo "  Active     : ${a_count}  $( [ "$a_count" -gt 0 ] && echo '(action required)' || echo '(clean)')"
echo ""

if [ "$STRICT" -eq 1 ] && [ "$a_count" -gt 0 ]; then
    exit 1
fi
exit 0
