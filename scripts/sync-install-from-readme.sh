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
case "${1:-}" in
  --check)         CHECK_ONLY=1 ;;
  --init-markers)  INIT_MARKERS=1 ;;
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
# Bounded by the "🚀 Install in 10 seconds" line and the closing </table>
# plus the "Auto-detects" caption that follows it.

extract_block() {
  awk '
    /\*\*🚀 Install in 10 seconds/ { capture=1 }
    capture { print }
    capture && /<\/sub>$/ && seen_table { exit }
    /<\/table>/ { seen_table=1 }
  ' "$README"
}

BLOCK="$(extract_block)"
if [[ -z "$BLOCK" ]]; then
  echo "ERROR: install block not found in README.md" >&2
  exit 2
fi

GENERATED=$'<!-- Generated from README.md by scripts/sync-install-from-readme.sh — do not edit by hand -->\n\n'"$BLOCK"

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
