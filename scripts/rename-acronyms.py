#!/usr/bin/env python3
"""rename-acronyms.py — Codemod that renames forbidden acronym MixedCaps in Go identifiers.

Spec: spec/01-coding-guidelines/03-coding-guidelines-spec/03-golang/09-acronym-naming.md

Renames (only when followed by another uppercase letter, i.e. inside a
MixedCaps identifier — never in trailing position like `imdbID`):

    IMDb -> Imdb     TMDb -> Tmdb     API  -> Api
    HTTP -> Http     URL  -> Url      JSON -> Json
    SQL  -> Sql      HTML -> Html     XML  -> Xml

Comments (// and /* */) and string/rune literals are LEFT UNCHANGED so prose
like "the OMDb HTTPS URL" in doc comments is preserved.

Allowlist: trailing-initialism short locals (imdbID, tmdbID, *URL) — these
already pass the lint because the acronym is not followed by another
uppercase letter.

Usage:
    python3 scripts/rename-acronyms.py             # dry-run, prints diff summary
    python3 scripts/rename-acronyms.py --write     # apply changes in place
    python3 scripts/rename-acronyms.py --write path/to/file.go ...

Exit codes:
    0  no changes needed (or all changes applied)
    1  changes needed and dry-run (re-run with --write)
    2  bad arguments
"""
from __future__ import annotations
import os
import re
import sys

# ORDER MATTERS: longer acronyms (IMDb/TMDb) must come before any 3-letter
# acronyms they could overlap with. Each rule rewrites OLD -> NEW only when
# OLD is followed by another uppercase letter (i.e. mid-identifier).
RULES: list[tuple[str, str]] = [
    ("IMDb", "Imdb"),
    ("TMDb", "Tmdb"),
    ("HTTP", "Http"),
    ("JSON", "Json"),
    ("HTML", "Html"),
    ("API",  "Api"),
    ("URL",  "Url"),
    ("SQL",  "Sql"),
    ("XML",  "Xml"),
]

SKIP_DIRS = {".git", ".release", "node_modules", ".gitmap"}


def split_segments(src: str) -> list[tuple[str, str]]:
    """Split Go source into (kind, text) segments preserving everything.

    kind ∈ {"code", "skip"} — only "code" segments are eligible for rewrite.
    "skip" covers line comments, block comments, string and rune literals.
    """
    segs: list[tuple[str, str]] = []
    buf: list[str] = []
    i, n = 0, len(src)

    def flush(kind: str) -> None:
        if buf:
            segs.append((kind, "".join(buf)))
            buf.clear()

    while i < n:
        c = src[i]
        c2 = src[i:i + 2]
        if c2 == "//":
            flush("code")
            j = src.find("\n", i)
            j = n if j == -1 else j
            segs.append(("skip", src[i:j]))
            i = j
            continue
        if c2 == "/*":
            flush("code")
            j = src.find("*/", i + 2)
            j = n if j == -1 else j + 2
            segs.append(("skip", src[i:j]))
            i = j
            continue
        if c == '"' or c == "'":
            flush("code")
            quote = c
            j = i + 1
            while j < n and src[j] != quote:
                if src[j] == "\\" and j + 1 < n:
                    j += 2; continue
                if src[j] == "\n":
                    break
                j += 1
            j = min(j + 1, n)
            segs.append(("skip", src[i:j]))
            i = j
            continue
        if c == "`":
            flush("code")
            j = src.find("`", i + 1)
            j = n if j == -1 else j + 1
            segs.append(("skip", src[i:j]))
            i = j
            continue
        buf.append(c)
        i += 1
    flush("code")
    return segs


def rewrite_code(text: str) -> tuple[str, int]:
    """Apply every RULE to a code segment. Returns (new_text, n_replacements)."""
    total = 0
    for old, new in RULES:
        # OLD followed by an uppercase ASCII letter (mid-identifier only).
        pat = re.compile(rf"{old}(?=[A-Z])")
        text, n = pat.subn(new, text)
        total += n
    return text, total


def process_file(path: str, write: bool) -> int:
    try:
        with open(path, encoding="utf-8") as fh:
            src = fh.read()
    except OSError as exc:
        print(f"skip {path}: {exc}", file=sys.stderr)
        return 0
    segs = split_segments(src)
    out: list[str] = []
    changes = 0
    for kind, text in segs:
        if kind == "code":
            text, n = rewrite_code(text)
            changes += n
        out.append(text)
    if changes == 0:
        return 0
    new_src = "".join(out)
    if write:
        with open(path, "w", encoding="utf-8") as fh:
            fh.write(new_src)
        print(f"rewrote {path}  ({changes} replacement(s))")
    else:
        print(f"would rewrite {path}  ({changes} replacement(s))")
    return changes


def iter_targets(args: list[str]):
    if args:
        for a in args:
            if os.path.isdir(a):
                yield from walk_dir(a)
            elif a.endswith(".go") and os.path.isfile(a):
                yield a
        return
    yield from walk_dir(".")


def walk_dir(root: str):
    for dirpath, dirs, files in os.walk(root):
        dirs[:] = [d for d in dirs if d not in SKIP_DIRS]
        for f in files:
            if f.endswith(".go"):
                yield os.path.join(dirpath, f)


def main(argv: list[str]) -> int:
    write = False
    paths: list[str] = []
    for a in argv:
        if a in ("-w", "--write"):
            write = True
        elif a in ("-h", "--help"):
            print(__doc__); return 0
        elif a.startswith("-"):
            print(f"unknown arg: {a}", file=sys.stderr); return 2
        else:
            paths.append(a)

    total = 0
    files_changed = 0
    for path in iter_targets(paths):
        n = process_file(path, write)
        if n:
            files_changed += 1
            total += n

    if total == 0:
        print("✅ Nothing to rewrite — all identifiers already comply.")
        return 0
    verb = "rewrote" if write else "would rewrite"
    print(f"\n{verb} {total} occurrence(s) across {files_changed} file(s).")
    if not write:
        print("Re-run with --write to apply changes, then run:")
        print("  python3 scripts/check-acronym-naming.py")
        print("  go build ./...   &&   go test ./...")
        return 1
    print("Next: python3 scripts/check-acronym-naming.py && go build ./... && go test ./...")
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
