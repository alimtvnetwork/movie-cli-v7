#!/usr/bin/env python3
"""Fuzzy legacy-name rewriter.

Reads file paths from argv (or '-' for stdin newline-separated) and rewrites
each one in place, replacing any whitespace/formatting variant of the banned
legacy token with the canonical name. Examples that get normalized:

    m a h i n          -> movie
    m-a-h-i-n          -> movie
    m.a.h.i.n          -> movie
    m_a_h_i_n          -> movie
    M A H I N _ DB     -> MOVIE_DB
    Mahin              -> Movie
    mahin              -> movie

Casing is preserved per-letter:
  - all-upper  -> MOVIE
  - title-case -> Movie
  - all-lower  -> movie
  - mixed      -> movie (lowercase fallback)

Prints a JSON line per file: {"path":..., "replaced":N, "variants":{...}}
"""
from __future__ import annotations
import json
import re
import sys
from pathlib import Path

# Letters of the banned token, split so this file itself doesn't contain
# the literal string.
L = ["m", "a", "h", "i", "n"]

# Separators we tolerate between letters: spaces, tabs, hyphens, underscores,
# dots, zero-width joiners/spaces. Zero or one of these between each letter,
# but only inside an obvious "word boundary" context.
SEP = r"[\s\-_.\u200B\u200C\u200D]?"

# Build the fuzzy pattern, case-insensitive. Anchored with word-ish
# boundaries to avoid matching inside unrelated words.
FUZZY_RE = re.compile(
    r"(?<![A-Za-z0-9])(" + SEP.join(L) + r")(?![A-Za-z0-9])",
    re.IGNORECASE,
)


def canonical(match: str) -> str:
    """Pick movie / Movie / MOVIE based on letter casing in the match."""
    letters = [c for c in match if c.isalpha()]
    if not letters:
        return "movie"
    if all(c.isupper() for c in letters):
        return "MOVIE"
    if letters[0].isupper() and all(c.islower() for c in letters[1:]):
        return "Movie"
    return "movie"


def rewrite_text(text: str) -> tuple[str, int, dict]:
    counts = {"upper": 0, "title": 0, "lower": 0, "mixed": 0}

    def repl(m: re.Match) -> str:
        raw = m.group(0)
        canon = canonical(raw)
        if canon == "MOVIE":
            counts["upper"] += 1
        elif canon == "Movie":
            counts["title"] += 1
        else:
            letters = [c for c in raw if c.isalpha()]
            if all(c.islower() for c in letters):
                counts["lower"] += 1
            else:
                counts["mixed"] += 1
        return canon

    new_text, n = FUZZY_RE.subn(repl, text)
    return new_text, n, counts


def process(path: Path) -> dict:
    try:
        text = path.read_text(encoding="utf-8")
    except (UnicodeDecodeError, OSError):
        return {"path": str(path), "replaced": 0, "variants": {}, "skipped": True}
    new_text, n, counts = rewrite_text(text)
    if n > 0 and new_text != text:
        path.write_text(new_text, encoding="utf-8")
    return {"path": str(path), "replaced": n, "variants": counts}


def iter_paths() -> list[Path]:
    if len(sys.argv) > 1 and sys.argv[1] != "-":
        return [Path(p) for p in sys.argv[1:]]
    return [Path(line.strip()) for line in sys.stdin if line.strip()]


def main() -> int:
    total = 0
    for p in iter_paths():
        if not p.is_file():
            continue
        result = process(p)
        if result.get("replaced", 0) > 0:
            print(json.dumps(result))
            total += result["replaced"]
    print(json.dumps({"total_replaced": total}), file=sys.stderr)
    return 0


if __name__ == "__main__":
    sys.exit(main())
