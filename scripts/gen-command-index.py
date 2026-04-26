#!/usr/bin/env python3
"""
Single source of truth for the README's "Flat alphabetical command index".

Why this exists
---------------
The index appears twice in README.md:

  1. An HTML <table> for nicely-rendered GitHub viewing.
  2. A fenced ```text``` block for terminals / monospace editors.

Hand-editing both blocks is how drift starts. This script owns the data and
regenerates both regions in place, between explicit marker comments:

  <!-- COMMAND-INDEX:HTML:BEGIN -->   ... <!-- COMMAND-INDEX:HTML:END -->
  <!-- COMMAND-INDEX:TEXT:BEGIN -->   ... <!-- COMMAND-INDEX:TEXT:END -->

It also keeps the six command-section anchors (Scanning & Library, File
Management, History & Undo, Discovery & Organization, Maintenance &
Debugging, Configuration & System) in sync everywhere they appear in the
README. Each section label has exactly ONE canonical GitHub slug, computed
from the label itself; any `(#stale)` link to one of these sections is
rewritten in place when you run the script, and `--check` fails CI if any
such rewrite is still pending. Anchors unrelated to the six sections are
left untouched.

Usage
-----
  python3 scripts/gen-command-index.py            # rewrite README.md in place
  python3 scripts/gen-command-index.py --check    # exit 1 if README is stale

The --check mode is what CI runs to enforce no-drift.
"""
from __future__ import annotations

import argparse
import html
import pathlib
import re
import sys

REPO_ROOT = pathlib.Path(__file__).resolve().parent.parent
README = REPO_ROOT / "README.md"

# ─── Source of truth ────────────────────────────────────────────────────────
# Each entry: (command, section_label, example_keyword)
#
# - command         : exact CLI invocation, with <placeholders>
# - section_label   : human-readable section name shown in the index. Must be
#                     one of SECTION_LABELS below — its anchor is derived from
#                     the label via _section_slug() so renaming a section in
#                     ONE place updates every link that references it.
# - example_keyword : minimal Ctrl/⌘+F-friendly snippet (trailing space if it's
#                     a prefix, no space if it's already complete)
#
# Adding/renaming a command? Edit ONLY this list, then run the script.
SECTION_LABELS: tuple[str, ...] = (
    "Scanning & Library",
    "File Management",
    "History & Undo",
    "Discovery & Organization",
    "Maintenance & Debugging",
    "Configuration & System",
)

COMMANDS: list[tuple[str, str, str]] = [
    ("movie cd <id>",                              "File Management",         "movie cd "),
    ("movie changelog",                            "Configuration & System",  "movie changelog"),
    ("movie cleanup",                              "Maintenance & Debugging", "movie cleanup"),
    ("movie config",                               "Configuration & System",  "movie config"),
    ("movie config get <key>",                     "Configuration & System",  "movie config get "),
    ("movie config set <key> <value>",             "Configuration & System",  "movie config set "),
    ("movie config set source_folder <path>",      "Configuration & System",  "movie config set source_folder "),
    ("movie config set tmdb_api_key <key>",        "Configuration & System",  "movie config set tmdb_api_key "),
    ("movie db",                                   "Maintenance & Debugging", "movie db"),
    ("movie discover",                             "Discovery & Organization","movie discover"),
    ("movie duplicates",                           "File Management",         "movie duplicates"),
    ("movie export",                               "Maintenance & Debugging", "movie export"),
    ("movie export --format csv --out <file>",     "Maintenance & Debugging", "movie export --format csv "),
    ("movie export --format json --out <file>",    "Maintenance & Debugging", "movie export --format json "),
    ("movie hello",                                "Configuration & System",  "movie hello"),
    ("movie info <id>",                            "Scanning & Library",      "movie info "),
    ("movie info <id> --json",                     "Scanning & Library",      "movie info --json "),
    ("movie logs",                                 "Maintenance & Debugging", "movie logs"),
    ("movie ls",                                   "Scanning & Library",      "movie ls"),
    ("movie ls --genre <name>",                    "Scanning & Library",      "movie ls --genre "),
    ("movie ls --limit <n>",                       "Scanning & Library",      "movie ls --limit "),
    ("movie ls --year <yyyy> --sort <field>",      "Scanning & Library",      "movie ls --year "),
    ("movie move",                                 "File Management",         "movie move"),
    ("movie move --all",                           "File Management",         "movie move --all"),
    ("movie move <id> --to <path>",                "File Management",         "movie move --to "),
    ("movie play <id>",                            "File Management",         "movie play "),
    ("movie play <id> --player <bin>",             "File Management",         "movie play --player "),
    ("movie popout",                               "File Management",         "movie popout"),
    ("movie redo",                                 "History & Undo",          "movie redo"),
    ("movie rename",                               "File Management",         "movie rename"),
    ("movie rename <id>",                          "File Management",         "movie rename "),
    ("movie rename --all --pattern <fmt>",         "File Management",         "movie rename --all --pattern "),
    ("movie rescan",                               "Scanning & Library",      "movie rescan"),
    ("movie rest",                                 "Maintenance & Debugging", "movie rest"),
    ("movie rest --open",                          "Maintenance & Debugging", "movie rest --open"),
    ("movie rest --port <n>",                      "Maintenance & Debugging", "movie rest --port "),
    ("movie scan",                                 "Scanning & Library",      "movie scan"),
    ("movie scan <path>",                          "Scanning & Library",      "movie scan "),
    ("movie scan <path> --dry-run",                "Scanning & Library",      "movie scan --dry-run "),
    ("movie scan <path> --refresh",                "Scanning & Library",      "movie scan --refresh "),
    ("movie search <query>",                       "Scanning & Library",      "movie search "),
    ("movie search <query> --year <yyyy>",         "Scanning & Library",      "movie search --year "),
    ("movie stats",                                "Discovery & Organization","movie stats"),
    ("movie stats --by <dimension>",               "Discovery & Organization","movie stats --by "),
    ("movie suggest",                              "Discovery & Organization","movie suggest"),
    ("movie suggest --genre <name> --limit <n>",   "Discovery & Organization","movie suggest --genre "),
    ("movie tag add <id> <tag>",                   "Discovery & Organization","movie tag add "),
    ("movie tag list <id>",                        "Discovery & Organization","movie tag list "),
    ("movie tag list --all",                       "Discovery & Organization","movie tag list --all"),
    ("movie tag remove <id> <tag>",                "Discovery & Organization","movie tag remove "),
    ("movie tag remove <id> --all",                "Discovery & Organization","movie tag remove --all"),
    ("movie undo",                                 "History & Undo",          "movie undo"),
    ("movie undo --id <history-id>",               "History & Undo",          "movie undo --id "),
    ("movie undo --list",                          "History & Undo",          "movie undo --list"),
    ("movie update",                               "Configuration & System",  "movie update"),
    ("movie version",                              "Configuration & System",  "movie version"),
    ("movie watch add <id>",                       "Discovery & Organization","movie watch add "),
    ("movie watch add <id> --priority <level>",    "Discovery & Organization","movie watch add --priority "),
    ("movie watch list",                           "Discovery & Organization","movie watch list"),
    ("movie watch list --sort <field>",            "Discovery & Organization","movie watch list --sort "),
]

HTML_BEGIN = "<!-- COMMAND-INDEX:HTML:BEGIN -->"
HTML_END   = "<!-- COMMAND-INDEX:HTML:END -->"
TEXT_BEGIN = "<!-- COMMAND-INDEX:TEXT:BEGIN -->"
TEXT_END   = "<!-- COMMAND-INDEX:TEXT:END -->"


def _section_slug(label: str) -> str:
    """
    GitHub-style heading anchor for one of the six command-section labels.

    Mirrors GitHub's slugger:
      - lowercase
      - drop characters that aren't alphanumeric, space, or hyphen
      - spaces → '-'

    So `&` is dropped (becoming a doubled hyphen between its neighbours):
      "Scanning & Library"      → "scanning--library"
      "Configuration & System"  → "configuration--system"
      "File Management"         → "file-management"
    """
    s = label.lower()
    s = re.sub(r"[^\w\s-]", "", s, flags=re.UNICODE)
    s = s.replace("_", "")
    s = re.sub(r"\s+", "-", s.strip())
    return s


def _section_anchor(label: str) -> str:
    """'#'-prefixed section anchor for a known command-section label."""
    if label not in SECTION_LABELS:
        sys.exit(f"error: unknown section label '{label}' (not in SECTION_LABELS)")
    return f"#{_section_slug(label)}"


# Pre-compute label → canonical anchor for fast lookup during the rewrite pass.
SECTION_ANCHORS: dict[str, str] = {label: _section_anchor(label) for label in SECTION_LABELS}


def _row_id(command: str) -> str:
    """Stable per-row anchor: 'movie scan <path> --dry-run' → 'movie-scan-path-dry-run'."""
    slug = re.sub(r"[<>]", "", command)
    slug = re.sub(r"\s+--", "-", slug)
    slug = re.sub(r"[\s_]+", "-", slug)
    slug = re.sub(r"-+", "-", slug)
    return slug.strip("-").lower()


def render_html() -> str:
    head = (
        "<table>\n"
        "<thead><tr>\n"
        '<th align="left" width="38%">Command</th>\n'
        '<th align="center" width="4%">→</th>\n'
        '<th align="left" width="20%">Section</th>\n'
        '<th align="left" width="22%">Example keyword</th>\n'
        '<th align="right" width="16%">Anchor</th>\n'
        "</tr></thead>\n"
        "<tbody>"
    )
    body = []
    for command, section, keyword in COMMANDS:
        rid = _row_id(command)
        anchor = SECTION_ANCHORS[section]
        cmd_html = html.escape(command)
        kw_html = html.escape(keyword)
        body.append(
            f'<tr id="{rid}">'
            f'<td><a href="{anchor}" title="Jump to the {section} section"><code>{cmd_html}</code></a></td>'
            f'<td align="center">→</td>'
            f'<td><a href="{anchor}">{section}</a></td>'
            f'<td><code>{kw_html}</code></td>'
            f'<td align="right"><code>#{rid}</code></td>'
            f'</tr>'
        )
    tail = "</tbody>\n</table>"
    return head + "\n" + "\n".join(body) + "\n" + tail


def render_text() -> str:
    cmd_w = max(len(c) for c, _, _ in COMMANDS)
    sec_w = max(len(s) for _, s, _ in COMMANDS)
    rows = ["```text"]
    rows.append(f"{'Command'.ljust(cmd_w)}     {'Section'.ljust(sec_w)}     Anchor")
    anchor_w = max(len(f"#{_row_id(c)}") for c, _, _ in COMMANDS)
    rows.append(f"{'-'*cmd_w}   {'-'*1}   {'-'*sec_w}   {'-'*anchor_w}")
    for command, section, _ in COMMANDS:
        rid = _row_id(command)
        rows.append(f"{command.ljust(cmd_w)}   →   {section.ljust(sec_w)}   #{rid}")
    rows.append("```")
    return "\n".join(rows)


def _replace_region(content: str, begin: str, end: str, body: str) -> str:
    pattern = re.compile(
        re.escape(begin) + r"\n.*?\n" + re.escape(end),
        flags=re.DOTALL,
    )
    replacement = f"{begin}\n{body}\n{end}"
    new_content, count = pattern.subn(replacement, content)
    if count != 1:
        sys.exit(f"error: expected exactly one '{begin}'..'{end}' region, found {count}")
    return new_content


def _rewrite_section_anchors(content: str) -> tuple[str, list[str]]:
    """
    Find every Markdown link or HTML href whose anchor target should resolve
    to one of the six command sections, and rewrite the anchor to its current
    canonical slug. Returns (new_content, list_of_human_readable_changes).

    Matching rule (intentionally narrow):
      A `(#xxx)` or `href="#xxx"` is considered "owned by" a section if its
      slug, lowercased and stripped of '-', equals the section label's
      alphanumeric-only fingerprint. So `#file-management`, `#filemanagement`,
      `#FILE-MANAGEMENT` all map to "File Management"; but `#installation`,
      `#troubleshooting`, `#quick-start` etc. don't match any section and are
      left alone. Anchors INSIDE the COMMAND-INDEX regions are skipped — those
      blocks are fully regenerated by render_html() / render_text() and would
      never contain a stale value at the point this pass runs.
    """
    # label fingerprint (a–z, 0–9 only) → label
    fingerprint_to_label: dict[str, str] = {}
    for label in SECTION_LABELS:
        fp = re.sub(r"[^a-z0-9]", "", label.lower())
        fingerprint_to_label[fp] = label

    # Carve out the two index regions so we don't double-process them.
    def _spans(begin: str, end: str) -> tuple[int, int] | None:
        i = content.find(begin)
        j = content.find(end, i + len(begin)) if i != -1 else -1
        return (i, j + len(end)) if i != -1 and j != -1 else None

    skip_spans: list[tuple[int, int]] = []
    for begin, end in ((HTML_BEGIN, HTML_END), (TEXT_BEGIN, TEXT_END)):
        sp = _spans(begin, end)
        if sp:
            skip_spans.append(sp)

    def _in_skip(pos: int) -> bool:
        return any(s <= pos < e for s, e in skip_spans)

    changes: list[str] = []

    # Patterns: Markdown link target `(#xxx)` and HTML attribute `href="#xxx"`.
    md_pat = re.compile(r"\(#([A-Za-z0-9_-]+)\)")
    html_pat = re.compile(r'href="#([A-Za-z0-9_-]+)"')

    def _maybe_rewrite(match: re.Match[str], wrap: tuple[str, str]) -> str:
        if _in_skip(match.start()):
            return match.group(0)
        raw = match.group(1)
        fp = re.sub(r"[^a-z0-9]", "", raw.lower())
        label = fingerprint_to_label.get(fp)
        if label is None:
            return match.group(0)  # not a known section — leave alone
        canonical = _section_slug(label)
        if raw == canonical:
            return match.group(0)
        line_no = content.count("\n", 0, match.start()) + 1
        changes.append(f"  line {line_no}: #{raw} → #{canonical}  ({label})")
        return f"{wrap[0]}#{canonical}{wrap[1]}"

    new_content = md_pat.sub(lambda m: _maybe_rewrite(m, ("(", ")")), content)
    new_content = html_pat.sub(lambda m: _maybe_rewrite(m, ('href="', '"')), new_content)
    return new_content, changes


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--check", action="store_true", help="exit 1 if README would change")
    args = parser.parse_args()

    original = README.read_text(encoding="utf-8")

    # 1. Rewrite stale section anchors first so the COMMAND-INDEX regions are
    #    rebuilt from the same canonical slugs everything else now uses.
    updated, anchor_changes = _rewrite_section_anchors(original)

    # 2. Regenerate the two index regions in place.
    updated = _replace_region(updated, HTML_BEGIN, HTML_END, render_html())
    updated = _replace_region(updated, TEXT_BEGIN, TEXT_END, render_text())

    if args.check:
        if updated != original:
            sys.stderr.write("README.md is stale:\n")
            if anchor_changes:
                sys.stderr.write(
                    f"  {len(anchor_changes)} section anchor(s) need rewriting:\n"
                )
                for change in anchor_changes:
                    sys.stderr.write(change + "\n")
            else:
                sys.stderr.write("  command-index region(s) need regenerating\n")
            sys.stderr.write("Run: python3 scripts/gen-command-index.py\n")
            return 1
        print("README.md command index and section anchors are up to date.")
        return 0

    if updated == original:
        print("README.md command index and section anchors already up to date.")
        return 0
    README.write_text(updated, encoding="utf-8")
    msg = f"README.md command index regenerated ({len(COMMANDS)} commands)."
    if anchor_changes:
        msg += f"\nRewrote {len(anchor_changes)} section anchor reference(s):"
        for change in anchor_changes:
            msg += "\n" + change
    print(msg)
    return 0


if __name__ == "__main__":
    sys.exit(main())