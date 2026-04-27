#!/usr/bin/env bash
# check-contributing-toc.sh — verifies that every link in the
# CONTRIBUTING.md "Table of Contents" block points at a heading that
# actually exists in the file, and (in the other direction) that every
# top-level "## " section is referenced by the TOC.
#
# Why this exists:
#   The TOC is hand-maintained. When a new section like
#   "Undo / Redo Worked Examples" is added, the matching anchor in the
#   TOC and the heading text MUST agree exactly under GitHub's slugging
#   rules — otherwise the link silently 404s inside the rendered doc.
#
# Slugging rules implemented (matches github.com markdown renderer):
#   1. Lowercase.
#   2. Strip every character that is NOT a letter, digit, space, or '-'.
#   3. Replace runs of whitespace with a single '-'.
#   That collapses "Undo / Redo Worked Examples" → "undo--redo-worked-examples"
#   because the two spaces around '/' both become '-' and the '/' itself is
#   stripped (rule 2). Tested against the live anchor GitHub generates.
#
# Exit codes:
#   0 — TOC and headings agree.
#   1 — at least one mismatch (TOC link with no heading, or vice versa).
#
# Run locally:        bash scripts/check-contributing-toc.sh
# Run against a file: bash scripts/check-contributing-toc.sh path/to/FILE.md

set -uo pipefail

FILE="${1:-CONTRIBUTING.md}"

if [ ! -f "$FILE" ]; then
    echo "::error file=${FILE}::file not found"
    exit 1
fi

slugify() {
    # Read one line on stdin, emit the GitHub-style anchor on stdout.
    # Rules (verified against github.com renderer):
    #   1. Lowercase.
    #   2. Drop every char that is NOT a letter, digit, space, or '-'.
    #   3. Replace EACH remaining space with '-' (do NOT collapse runs —
    #      that is why "Undo / Redo" → "undo--redo": the two spaces around
    #      '/' both survive step 2 and each becomes its own '-').
    awk '{
        s = tolower($0)
        out = ""
        for (i = 1; i <= length(s); i++) {
            c = substr(s, i, 1)
            if (c ~ /[a-z0-9 -]/) out = out c
        }
        gsub(/ /, "-", out)
        print out
    }'
}

# 1. Pull the TOC block — everything between the first "## Table of Contents"
#    line and the next "## " heading. Then extract anchors that look like
#    "- [Title](#anchor)".
toc_block=$(awk '
    /^## Table of Contents/ { capture = 1; next }
    capture && /^## /       { exit }
    capture                 { print }
' "$FILE")

if [ -z "$toc_block" ]; then
    echo "::error file=${FILE}::no '## Table of Contents' block found"
    exit 1
fi

toc_anchors=$(echo "$toc_block" \
    | grep -oE '\]\(#[a-z0-9-]+\)' \
    | sed -E 's/^\]\(#//; s/\)$//' \
    | sort -u)

# 2. Compute the expected anchor for every "## " heading in the file
#    (skip the TOC heading itself — it is not self-referenced).
heading_anchors=$(grep -E '^## ' "$FILE" \
    | sed -E 's/^## //' \
    | grep -v '^Table of Contents$' \
    | while IFS= read -r h; do
        printf '%s\n' "$h" | slugify
      done \
    | sort -u)

fail=0
report() {
    echo "::error file=${FILE}::$1"
    echo "  [FAIL] $1"
    fail=1
}

# 3. Every TOC anchor MUST exist as a heading-derived anchor.
while IFS= read -r anchor; do
    [ -z "$anchor" ] && continue
    if ! echo "$heading_anchors" | grep -qx "$anchor"; then
        report "TOC links to '#${anchor}' but no '## ' heading produces that slug"
    fi
done <<<"$toc_anchors"

# 4. Every "## " heading MUST be referenced by the TOC.
while IFS= read -r anchor; do
    [ -z "$anchor" ] && continue
    if ! echo "$toc_anchors" | grep -qx "$anchor"; then
        report "Heading slug '#${anchor}' is not listed in the Table of Contents"
    fi
done <<<"$heading_anchors"

# 5. Targeted assertion for the section that triggered this check —
#    fail loudly with a hand-written message if "Undo / Redo Worked
#    Examples" or its anchor disappears, so future renames stay obvious.
if ! grep -qE '^## Undo / Redo Worked Examples$' "$FILE"; then
    report "missing required heading: '## Undo / Redo Worked Examples'"
fi
if ! echo "$toc_anchors" | grep -qx 'undo--redo-worked-examples'; then
    report "TOC missing anchor '#undo--redo-worked-examples' for the Worked Examples section"
fi

if [ "$fail" -ne 0 ]; then
    echo ""
    echo "CONTRIBUTING.md TOC check failed."
    exit 1
fi
echo "✅ CONTRIBUTING.md TOC and headings are in sync."
exit 0
