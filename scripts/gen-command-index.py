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
# Each entry: (command, section_label, section_anchor, example_keyword)
#
# - command         : exact CLI invocation, with <placeholders>
# - section_label   : human-readable section name shown in the index
# - section_anchor  : GitHub-style heading anchor (lowercase, & → "", spaces → -)
# - example_keyword : minimal Ctrl/⌘+F-friendly snippet (trailing space if it's
#                     a prefix, no space if it's already complete)
#
# Adding/renaming a command? Edit ONLY this list, then run the script.
COMMANDS: list[tuple[str, str, str, str]] = [
    ("movie cd <id>",                              "File Management",         "#file-management",          "movie cd "),
    ("movie changelog",                            "Configuration & System",  "#configuration--system",    "movie changelog"),
    ("movie cleanup",                              "Maintenance & Debugging", "#maintenance--debugging",   "movie cleanup"),
    ("movie config",                               "Configuration & System",  "#configuration--system",    "movie config"),
    ("movie config get <key>",                     "Configuration & System",  "#configuration--system",    "movie config get "),
    ("movie config set <key> <value>",             "Configuration & System",  "#configuration--system",    "movie config set "),
    ("movie config set source_folder <path>",      "Configuration & System",  "#configuration--system",    "movie config set source_folder "),
    ("movie config set tmdb_api_key <key>",        "Configuration & System",  "#configuration--system",    "movie config set tmdb_api_key "),
    ("movie db",                                   "Maintenance & Debugging", "#maintenance--debugging",   "movie db"),
    ("movie discover",                             "Discovery & Organization","#discovery--organization",  "movie discover"),
    ("movie duplicates",                           "File Management",         "#file-management",          "movie duplicates"),
    ("movie export",                               "Maintenance & Debugging", "#maintenance--debugging",   "movie export"),
    ("movie export --format csv --out <file>",     "Maintenance & Debugging", "#maintenance--debugging",   "movie export --format csv "),
    ("movie export --format json --out <file>",    "Maintenance & Debugging", "#maintenance--debugging",   "movie export --format json "),
    ("movie hello",                                "Configuration & System",  "#configuration--system",    "movie hello"),
    ("movie info <id>",                            "Scanning & Library",      "#scanning--library",        "movie info "),
    ("movie info <id> --json",                     "Scanning & Library",      "#scanning--library",        "movie info --json "),
    ("movie logs",                                 "Maintenance & Debugging", "#maintenance--debugging",   "movie logs"),
    ("movie ls",                                   "Scanning & Library",      "#scanning--library",        "movie ls"),
    ("movie ls --genre <name>",                    "Scanning & Library",      "#scanning--library",        "movie ls --genre "),
    ("movie ls --limit <n>",                       "Scanning & Library",      "#scanning--library",        "movie ls --limit "),
    ("movie ls --year <yyyy> --sort <field>",      "Scanning & Library",      "#scanning--library",        "movie ls --year "),
    ("movie move",                                 "File Management",         "#file-management",          "movie move"),
    ("movie move --all",                           "File Management",         "#file-management",          "movie move --all"),
    ("movie move <id> --to <path>",                "File Management",         "#file-management",          "movie move --to "),
    ("movie play <id>",                            "File Management",         "#file-management",          "movie play "),
    ("movie play <id> --player <bin>",             "File Management",         "#file-management",          "movie play --player "),
    ("movie popout",                               "File Management",         "#file-management",          "movie popout"),
    ("movie redo",                                 "History & Undo",          "#history--undo",            "movie redo"),
    ("movie rename",                               "File Management",         "#file-management",          "movie rename"),
    ("movie rename <id>",                          "File Management",         "#file-management",          "movie rename "),
    ("movie rename --all --pattern <fmt>",         "File Management",         "#file-management",          "movie rename --all --pattern "),
    ("movie rescan",                               "Scanning & Library",      "#scanning--library",        "movie rescan"),
    ("movie rest",                                 "Maintenance & Debugging", "#maintenance--debugging",   "movie rest"),
    ("movie rest --open",                          "Maintenance & Debugging", "#maintenance--debugging",   "movie rest --open"),
    ("movie rest --port <n>",                      "Maintenance & Debugging", "#maintenance--debugging",   "movie rest --port "),
    ("movie scan",                                 "Scanning & Library",      "#scanning--library",        "movie scan"),
    ("movie scan <path>",                          "Scanning & Library",      "#scanning--library",        "movie scan "),
    ("movie scan <path> --dry-run",                "Scanning & Library",      "#scanning--library",        "movie scan --dry-run "),
    ("movie scan <path> --refresh",                "Scanning & Library",      "#scanning--library",        "movie scan --refresh "),
    ("movie search <query>",                       "Scanning & Library",      "#scanning--library",        "movie search "),
    ("movie search <query> --year <yyyy>",         "Scanning & Library",      "#scanning--library",        "movie search --year "),
    ("movie stats",                                "Discovery & Organization","#discovery--organization",  "movie stats"),
    ("movie stats --by <dimension>",               "Discovery & Organization","#discovery--organization",  "movie stats --by "),
    ("movie suggest",                              "Discovery & Organization","#discovery--organization",  "movie suggest"),
    ("movie suggest --genre <name> --limit <n>",   "Discovery & Organization","#discovery--organization",  "movie suggest --genre "),
    ("movie tag add <id> <tag>",                   "Discovery & Organization","#discovery--organization",  "movie tag add "),
    ("movie tag list <id>",                        "Discovery & Organization","#discovery--organization",  "movie tag list "),
    ("movie tag list --all",                       "Discovery & Organization","#discovery--organization",  "movie tag list --all"),
    ("movie tag remove <id> <tag>",                "Discovery & Organization","#discovery--organization",  "movie tag remove "),
    ("movie tag remove <id> --all",                "Discovery & Organization","#discovery--organization",  "movie tag remove --all"),
    ("movie undo",                                 "History & Undo",          "#history--undo",            "movie undo"),
    ("movie undo --id <history-id>",               "History & Undo",          "#history--undo",            "movie undo --id "),
    ("movie undo --list",                          "History & Undo",          "#history--undo",            "movie undo --list"),
    ("movie update",                               "Configuration & System",  "#configuration--system",    "movie update"),
    ("movie version",                              "Configuration & System",  "#configuration--system",    "movie version"),
    ("movie watch add <id>",                       "Discovery & Organization","#discovery--organization",  "movie watch add "),
    ("movie watch add <id> --priority <level>",    "Discovery & Organization","#discovery--organization",  "movie watch add --priority "),
    ("movie watch list",                           "Discovery & Organization","#discovery--organization",  "movie watch list"),
    ("movie watch list --sort <field>",            "Discovery & Organization","#discovery--organization",  "movie watch list --sort "),
]

HTML_BEGIN = "<!-- COMMAND-INDEX:HTML:BEGIN -->"
HTML_END   = "<!-- COMMAND-INDEX:HTML:END -->"
TEXT_BEGIN = "<!-- COMMAND-INDEX:TEXT:BEGIN -->"
TEXT_END   = "<!-- COMMAND-INDEX:TEXT:END -->"


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
    for command, section, anchor, keyword in COMMANDS:
        rid = _row_id(command)
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
    cmd_w = max(len(c) for c, _, _, _ in COMMANDS)
    sec_w = max(len(s) for _, s, _, _ in COMMANDS)
    rows = ["```text"]
    rows.append(f"{'Command'.ljust(cmd_w)}     {'Section'.ljust(sec_w)}     Anchor")
    anchor_w = max(len(f"#{_row_id(c)}") for c, _, _, _ in COMMANDS)
    rows.append(f"{'-'*cmd_w}   {'-'*1}   {'-'*sec_w}   {'-'*anchor_w}")
    for command, section, _, _ in COMMANDS:
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


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--check", action="store_true", help="exit 1 if README would change")
    args = parser.parse_args()

    original = README.read_text(encoding="utf-8")
    updated = _replace_region(original, HTML_BEGIN, HTML_END, render_html())
    updated = _replace_region(updated, TEXT_BEGIN, TEXT_END, render_text())

    if args.check:
        if updated != original:
            sys.stderr.write(
                "README.md command index is stale.\n"
                "Run: python3 scripts/gen-command-index.py\n"
            )
            return 1
        print("README.md command index is up to date.")
        return 0

    if updated == original:
        print("README.md command index already up to date.")
        return 0
    README.write_text(updated, encoding="utf-8")
    print(f"README.md command index regenerated ({len(COMMANDS)} commands).")
    return 0


if __name__ == "__main__":
    sys.exit(main())