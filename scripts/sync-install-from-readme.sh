#!/usr/bin/env bash
# sync-install-from-readme.sh
#
# Extracts the canonical install block from README.md and rewrites the
# install snippets in QUICKSTART.md and spec/03-general/01-install-guide.md
# so wording and headers stay identical to the root README.
#
# Source of truth: README.md (per mem://preferences/readme-structure).
# Sub-docs MUST contain sentinel markers around the install block:
#
#   <!-- INSTALL:BEGIN -->
#   ...generated content...
#   <!-- INSTALL:END -->
#
# Anything outside the markers is preserved verbatim.
#
# Usage:
#   scripts/sync-install-from-readme.sh                  # rewrite files
#   scripts/sync-install-from-readme.sh --check          # exit 1 if drift
#   scripts/sync-install-from-readme.sh --init-markers   # add sentinels if missing
#
# Exit codes:
#   0  success / no drift
#   1  drift detected (--check) or missing markers
#   2  README install block not found / unknown option

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
README="$ROOT/README.md"
TARGETS=(
  "$ROOT/QUICKSTART.md"
  "$ROOT/spec/03-general/01-install-guide.md"
)

BEGIN_MARK="<!-- INSTALL:BEGIN -->"
END_MARK="<!-- INSTALL:END -->"

CHECK_ONLY=0
INIT_MARKERS=0
PRINT_ONLY=0
case "${1:-}" in
  --check)         CHECK_ONLY=1 ;;
  --init-markers)  INIT_MARKERS=1 ;;
  --print)         PRINT_ONLY=1 ;;
  "")              ;;
  *) echo "Unknown option: $1" >&2; exit 2 ;;
esac

# --- 0. Optionally insert sentinels into targets that lack them -------------
# Appends a fresh install section (with INSTALL:BEGIN / INSTALL:END markers)
# to the end of any target file that is missing the sentinels. Existing
# install content is left untouched — the sync step below will fill the new
# block on the next run.

init_markers() {
  local file="$1"
  if [[ ! -f "$file" ]]; then
    echo "SKIP: $file not found" >&2
    return 0
  fi
  if grep -qF "$BEGIN_MARK" "$file" && grep -qF "$END_MARK" "$file"; then
    echo "OK:    $file already has sentinels"
    return 0
  fi
  {
    printf '\n\n## Install\n\n'
    printf '%s\n\n%s\n' "$BEGIN_MARK" "$END_MARK"
  } >> "$file"
  echo "INIT:  $file sentinels appended"
}

if [[ "$INIT_MARKERS" -eq 1 ]]; then
  rc=0
  for f in "${TARGETS[@]}"; do
    init_markers "$f" || rc=1
  done
  exit $rc
fi

# --- 1. Extract install block from README -----------------------------------
# Strategy (most → least reliable):
#   1. Explicit sentinels in README:
#        <!-- README-INSTALL:BEGIN -->  ...  <!-- README-INSTALL:END -->
#      This is the canonical source — surrounding headings and tables can
#      change freely without breaking the sync.
#   2. Heuristic fallback for older READMEs: capture from the
#      "🚀 Install in 10 seconds" line through the first <sub>…</sub>
#      caption that follows the closing </table>.
# Both strategies trim trailing blank lines so the generated block is stable.

extract_by_sentinels() {
  awk '
    /<!-- README-INSTALL:BEGIN -->/ { capture=1; next }
    /<!-- README-INSTALL:END -->/   { exit }
    capture { print }
  ' "$README"
}

extract_by_heuristic() {
  awk '
    /\*\*🚀 Install in 10 seconds/ { capture=1 }
    capture { print }
    /<\/table>/ { seen_table=1 }
    capture && seen_table && /<\/sub>[[:space:]]*$/ { exit }
  ' "$README"
}

trim_blank_edges() {
  awk '
    { lines[NR]=$0 }
    END {
      first=1; last=NR
      while (first<=last && lines[first] ~ /^[[:space:]]*$/) first++
      while (last>=first && lines[last]  ~ /^[[:space:]]*$/) last--
      for (i=first; i<=last; i++) print lines[i]
    }
  '
}

extract_block() {
  local out
  out="$(extract_by_sentinels | trim_blank_edges)"
  if [[ -n "$out" ]]; then
    echo "$out"
    return 0
  fi
  echo "INFO: README sentinels not found — falling back to heuristic" >&2
  extract_by_heuristic | trim_blank_edges
}

BLOCK="$(extract_block)"
if [[ -z "$BLOCK" ]]; then
  echo "ERROR: install block not found in README.md (no sentinels, no heuristic match)" >&2
  exit 2
fi

GENERATED=$'<!-- Generated from README.md by scripts/sync-install-from-readme.sh — do not edit by hand -->\n\n'"$BLOCK"

# --- 1b. --print: emit the extracted block to stdout and exit ---------------
# Useful for previewing exactly what would be written before touching any
# target file. Header/footer go to stderr so the body on stdout stays
# pipe-friendly (e.g. `... --print | less`, `... --print > preview.md`,
# `diff <(... --print) <(sed -n '/INSTALL:BEGIN/,/INSTALL:END/p' QUICKSTART.md)`).

if [[ "$PRINT_ONLY" -eq 1 ]]; then
  {
    echo "----- README install block (extracted from README.md) -----"
    echo "----- $(printf '%s' "$BLOCK" | wc -l | tr -d ' ') lines, $(printf '%s' "$BLOCK" | wc -c | tr -d ' ') bytes -----"
  } >&2
  printf '%s\n' "$BLOCK"
  echo "----- end of block -----" >&2
  exit 0
fi

# --- 2. Replace block in each target between sentinels ----------------------

sync_file() {
  local file="$1"

  if [[ ! -f "$file" ]]; then
    echo "SKIP: $file not found" >&2
    return 0
  fi

  if ! grep -qF "$BEGIN_MARK" "$file" || ! grep -qF "$END_MARK" "$file"; then
    echo "ERROR: $file missing $BEGIN_MARK / $END_MARK sentinels" >&2
    return 1
  fi

  local tmp
  tmp="$(mktemp)"

  awk -v begin="$BEGIN_MARK" -v end="$END_MARK" -v block="$GENERATED" '
    BEGIN { inside=0 }
    {
      if (index($0, begin)) {
        print begin
        print ""
        print block
        print ""
        print end
        inside=1
        next
      }
      if (inside && index($0, end)) { inside=0; next }
      if (!inside) print
    }
  ' "$file" > "$tmp"

  if [[ "$CHECK_ONLY" -eq 1 ]]; then
    local h1 h2
    h1="$(sha256sum < "$file" | awk '{print $1}')"
    h2="$(sha256sum < "$tmp"  | awk '{print $1}')"
    if [[ "$h1" != "$h2" ]]; then
      echo "DRIFT: $file is out of sync with README.md — run scripts/sync-install-from-readme.sh" >&2
      rm -f "$tmp"
      return 1
    fi
    rm -f "$tmp"
    echo "OK:    $file in sync"
  else
    mv "$tmp" "$file"
    echo "WROTE: $file"
  fi
}

rc=0
for f in "${TARGETS[@]}"; do
  sync_file "$f" || rc=1
done

exit $rc
