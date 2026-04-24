#!/usr/bin/env bash
# log-milestone.sh — append a Malaysia-time entry to MILESTONES.md,
# bump the patch version in version/info.go, and create a single git commit.
#
# Usage:
#   scripts/log-milestone.sh                    # default note: "app run logged"
#   scripts/log-milestone.sh "custom note"      # custom note text
#   scripts/log-milestone.sh --event start "kickoff"   # custom event prefix
set -euo pipefail

EVENT="run"
if [[ "${1:-}" == "--event" ]]; then
  EVENT="${2:?--event requires a value}"
  shift 2
fi
NOTE="${1:-app run logged}"

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

MILESTONES="MILESTONES.md"
VERSION_FILE="version/info.go"

if [[ ! -f "$MILESTONES" ]]; then
  echo "error: $MILESTONES not found at repo root" >&2
  exit 1
fi
if [[ ! -f "$VERSION_FILE" ]]; then
  echo "error: $VERSION_FILE not found" >&2
  exit 1
fi

# Malaysia time (UTC+8), format: dd-MMM-YYYY hh:mm AM/PM
TS="$(TZ='Asia/Kuala_Lumpur' date '+%d-%b-%Y %I:%M %p')"
ENTRY="- ${EVENT} ${TS} — ${NOTE}"

# Ensure file ends with a newline before appending.
[[ -s "$MILESTONES" && "$(tail -c1 "$MILESTONES")" != "" ]] && printf '\n' >> "$MILESTONES"
printf '%s\n' "$ENTRY" >> "$MILESTONES"

# Bump patch version: vMAJOR.MINOR.PATCH -> vMAJOR.MINOR.(PATCH+1)
CURRENT="$(grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+' "$VERSION_FILE" | head -n1)"
if [[ -z "$CURRENT" ]]; then
  echo "error: could not parse current version from $VERSION_FILE" >&2
  exit 1
fi
IFS='.' read -r MAJ MIN PAT <<<"${CURRENT#v}"
NEW="v${MAJ}.${MIN}.$((PAT + 1))"

# Portable in-place edit (works on GNU + BSD sed via a temp file).
tmp="$(mktemp)"
sed "s|${CURRENT}|${NEW}|" "$VERSION_FILE" > "$tmp" && mv "$tmp" "$VERSION_FILE"

# Commit both files together.
git add "$MILESTONES" "$VERSION_FILE"
if git diff --cached --quiet; then
  echo "nothing to commit"
  exit 0
fi
git commit -m "chore(milestone): ${EVENT} ${TS} — ${NOTE} (${NEW})"

echo "✓ logged: ${ENTRY}"
echo "✓ version: ${CURRENT} → ${NEW}"
echo "✓ committed on $(git rev-parse --abbrev-ref HEAD)"