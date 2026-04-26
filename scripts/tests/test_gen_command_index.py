"""
Unit tests for scripts/gen-command-index.py.

Why these tests exist
---------------------
The slug algorithm and the anchor rewriter have subtle edge cases that have
already caused real bugs (`#scan--library` instead of `#scanning--library`,
a `&`-collapsing slugger that produced `#scanning-library`, etc.). These
tests pin the GitHub-compatible behaviour so future refactors can't quietly
regress it.

Run from repo root:
    python3 -m unittest scripts.tests.test_gen_command_index -v
    # or:
    python3 -m unittest discover -s scripts/tests -v

The script under test has a hyphen in its filename, so we load it via
importlib instead of `import scripts.gen_command_index`.
"""
from __future__ import annotations

import importlib.util
import pathlib
import sys
import unittest


def _load_module():
    """Load scripts/gen-command-index.py as a module, importable despite the hyphen."""
    here = pathlib.Path(__file__).resolve().parent
    src = here.parent / "gen-command-index.py"
    spec = importlib.util.spec_from_file_location("gen_command_index", src)
    if spec is None or spec.loader is None:
        raise RuntimeError(f"could not load spec for {src}")
    module = importlib.util.module_from_spec(spec)
    sys.modules["gen_command_index"] = module
    spec.loader.exec_module(module)
    return module


gci = _load_module()


class SlugAlgorithmTests(unittest.TestCase):
    """Lock down _section_slug() against GitHub's heading slugger."""

    def test_ampersand_with_spaces_becomes_double_hyphen(self):
        # The historic bug: a naive slugger collapses "Scanning & Library"
        # to "scanning-library" instead of GitHub's "scanning--library".
        self.assertEqual(gci._section_slug("Scanning & Library"), "scanning--library")
        self.assertEqual(gci._section_slug("History & Undo"), "history--undo")
        self.assertEqual(gci._section_slug("Configuration & System"), "configuration--system")

    def test_plain_space_becomes_single_hyphen(self):
        self.assertEqual(gci._section_slug("File Management"), "file-management")
        self.assertEqual(gci._section_slug("Quick Start"), "quick-start")

    def test_single_word_label(self):
        self.assertEqual(gci._section_slug("Installation"), "installation")
        self.assertEqual(gci._section_slug("Troubleshooting"), "troubleshooting")

    def test_lowercases_and_strips_other_punctuation(self):
        # GitHub drops punctuation (besides hyphens) entirely. Spaces around
        # the dropped char survive — that's what produces the doubled hyphen.
        self.assertEqual(gci._section_slug("Foo (Bar)"), "foo-bar")
        self.assertEqual(gci._section_slug("Foo: Bar"), "foo-bar")
        self.assertEqual(gci._section_slug("Foo / Bar"), "foo--bar")

    def test_existing_hyphens_preserved(self):
        self.assertEqual(gci._section_slug("Quick-Start"), "quick-start")
        self.assertEqual(gci._section_slug("multi-word-label"), "multi-word-label")

    def test_leading_trailing_whitespace_stripped(self):
        # The strip(" ") inside _section_slug guards against accidental
        # leading/trailing hyphens that would mismatch GitHub's anchor.
        self.assertEqual(gci._section_slug("  File Management  "), "file-management")


class SectionAnchorTests(unittest.TestCase):
    """The thin wrapper that prefixes '#' and validates the label."""

    def test_known_label_returns_hash_prefixed_slug(self):
        self.assertEqual(gci._section_anchor("File Management"), "#file-management")
        self.assertEqual(gci._section_anchor("Scanning & Library"), "#scanning--library")

    def test_unknown_label_exits(self):
        # _section_anchor() rejects labels that aren't in SECTION_LABELS so
        # typos can't silently produce dead anchors.
        with self.assertRaises(SystemExit):
            gci._section_anchor("Definitely Not A Section")

    def test_section_anchors_map_matches_helper(self):
        # The pre-computed lookup table must agree with the helper for every
        # managed label — they're used interchangeably in the rewriter.
        for label in gci.SECTION_LABELS:
            self.assertEqual(gci.SECTION_ANCHORS[label], gci._section_anchor(label))


class FingerprintRewriteTests(unittest.TestCase):
    """End-to-end behaviour of _rewrite_section_anchors()."""

    def _rewrite(self, body: str):
        # Wrap snippets in just enough boilerplate that the function's
        # skip-span lookups (HTML/TEXT/SECTION-CMDS markers) find nothing
        # and therefore process every anchor in the body.
        return gci._rewrite_section_anchors(body)

    def test_double_hyphen_section_canonicalises(self):
        # `#scanning_library` and `#history-undo` both fingerprint-match
        # their managed labels (alphanumerics only), so the rewriter must
        # rewrite them to the canonical doubled-hyphen forms produced by
        # `_section_slug` for "Scanning & Library" and "History & Undo".
        body = "See [scan](#scanning_library) and [hist](#history-undo)."
        new, changes = self._rewrite(body)
        self.assertIn("#scanning--library", new)
        self.assertIn("#history--undo", new)
        self.assertEqual(len(changes), 2)

    def test_underscore_and_uppercase_variants_collapse(self):
        # Same fingerprint → same canonical slug. Covers the most common
        # hand-typed mistakes for anchors with `&` in them.
        body = (
            '[a](#FILE_MANAGEMENT) [b](#filemanagement) '
            '<a href="#File-Management">x</a>'
        )
        new, _ = self._rewrite(body)
        self.assertEqual(new.count("#file-management"), 3)
        self.assertNotIn("#FILE_MANAGEMENT", new)
        self.assertNotIn("#File-Management", new)

    def test_extra_doc_anchors_are_managed(self):
        body = "[q](#quick_start) [t](#TROUBLESHOOTING) [i](#Installation)"
        new, changes = self._rewrite(body)
        self.assertIn("#quick-start", new)
        self.assertIn("#troubleshooting", new)
        # `#Installation` differs only in case from `#installation` — the
        # rewriter must catch that too, not just punctuation differences.
        self.assertIn("#installation", new)
        self.assertNotIn("#Installation", new)
        self.assertEqual(len(changes), 3)

    def test_canonical_anchor_unchanged_and_unreported(self):
        body = "[ok](#scanning--library) [ok2](#file-management)"
        new, changes = self._rewrite(body)
        self.assertEqual(new, body)
        self.assertEqual(changes, [])

    def test_unmanaged_anchor_left_alone(self):
        # Fingerprint `installationguide` matches no managed label, so the
        # rewriter must NOT collapse it to `#installation`.
        body = "[doc](#installation-guide) [other](#some-random-anchor)"
        new, changes = self._rewrite(body)
        self.assertEqual(new, body)
        self.assertEqual(changes, [])

    def test_html_href_form_rewritten(self):
        body = '<a href="#FILE_MANAGEMENT">file mgmt</a>'
        new, changes = self._rewrite(body)
        self.assertIn('href="#file-management"', new)
        self.assertEqual(len(changes), 1)

    def test_anchor_inside_command_index_region_skipped(self):
        # Anchors inside the auto-generated COMMAND-INDEX regions are owned
        # by render_html()/render_text() and must not be touched by the
        # anchor pass — otherwise we'd double-process them.
        body = (
            f"{gci.HTML_BEGIN}\n"
            '<a href="#FILE_MANAGEMENT">stale-but-skipped</a>\n'
            f"{gci.HTML_END}\n"
            'outside: <a href="#FILE_MANAGEMENT">should rewrite</a>\n'
        )
        new, changes = self._rewrite(body)
        # The inside-region anchor is preserved verbatim…
        self.assertIn(
            f'{gci.HTML_BEGIN}\n<a href="#FILE_MANAGEMENT">stale-but-skipped</a>\n{gci.HTML_END}',
            new,
        )
        # …while the outside one is canonicalised.
        self.assertIn('outside: <a href="#file-management">', new)
        self.assertEqual(len(changes), 1)

    def test_change_log_lines_carry_label_and_line_number(self):
        body = "first line\n[bad](#FILE_MANAGEMENT)\n"
        _, changes = self._rewrite(body)
        self.assertEqual(len(changes), 1)
        # Format is "  line N: #raw → #canonical  (Label)" — assert each
        # piece so a future format tweak fails loudly here instead of in CI.
        msg = changes[0]
        self.assertIn("line 2:", msg)
        self.assertIn("#FILE_MANAGEMENT", msg)
        self.assertIn("#file-management", msg)
        self.assertIn("File Management", msg)


class WhitelistTests(unittest.TestCase):
    """The escape hatch must short-circuit the rewriter, not be silently overridden."""

    def setUp(self):
        # Snapshot the module-level whitelist so each test can mutate it
        # without affecting the others or any later import.
        self._saved_list = gci.ANCHOR_WHITELIST
        self._saved_fp = gci._WHITELIST_FINGERPRINTS
        self._saved_by_fp = gci._WHITELIST_BY_FINGERPRINT

    def tearDown(self):
        gci.ANCHOR_WHITELIST = self._saved_list
        gci._WHITELIST_FINGERPRINTS = self._saved_fp
        gci._WHITELIST_BY_FINGERPRINT = self._saved_by_fp

    def _set_whitelist(self, entries):
        """
        Install a whitelist for one test. Accepts the same shapes the real
        constant accepts — bare strings (global) or (anchor, scope) tuples.
        Mirrors the module-load logic in gen-command-index.py so behaviour
        under test exactly matches production.
        """
        import re
        gci.ANCHOR_WHITELIST = entries
        by_fp: dict[str, set] = {}
        for entry in entries:
            anchor, scope = gci._normalize_whitelist_entry(entry)
            fp = re.sub(r"[^a-z0-9]", "", anchor.lower())
            by_fp.setdefault(fp, set()).add(scope)
        gci._WHITELIST_BY_FINGERPRINT = by_fp
        gci._WHITELIST_FINGERPRINTS = frozenset(
            fp for fp, scopes in by_fp.items() if None in scopes
        )

    def test_whitelisted_fingerprint_is_left_verbatim(self):
        # Without a whitelist this anchor wouldn't be rewritten anyway —
        # but with a whitelist we must guarantee it's *never* touched, even
        # in cases where the fingerprint also matches a managed label.
        self._set_whitelist(("legacy-quick-link",))
        body = "[legacy](#LEGACY_QUICK_LINK) [legacy2](#legacyquicklink)"
        new, changes = gci._rewrite_section_anchors(body)
        self.assertEqual(new, body)
        self.assertEqual(changes, [])

    def test_whitelist_uses_fingerprint_match(self):
        # The user explicitly chose fingerprint-based matching — a whitelist
        # entry of "Foo Bar" must cover #foo-bar, #FOO_BAR, #foobar.
        self._set_whitelist(("Foo Bar",))
        for raw in ("#foo-bar", "#FOO_BAR", "#foobar", "#Foo_Bar"):
            body = f"[x]({raw})"
            new, changes = gci._rewrite_section_anchors(body)
            self.assertEqual(new, body, f"whitelist failed to cover variant {raw}")
            self.assertEqual(changes, [])

    # ─── Per-section scoped entries ─────────────────────────────────────────

    def test_scoped_entry_only_suppresses_managed_anchor_in_its_section(self):
        # `quick_start` fingerprint-matches the managed Quick Start label, so
        # the rewriter would normally rewrite it to `#quick-start`. A scoped
        # whitelist entry must allow that variant inside the Quick Start
        # section but still rewrite it elsewhere.
        self._set_whitelist((("quick_start", "Quick Start"),))
        body = (
            "## Quick Start\n"
            "[inside](#quick_start)\n"
            "## Other\n"
            "[outside](#quick_start)\n"
        )
        new, changes = gci._rewrite_section_anchors(body)
        # Inside Quick Start: untouched
        self.assertIn("[inside](#quick_start)", new)
        # Outside: canonicalised
        self.assertIn("[outside](#quick-start)", new)
        # Exactly one rewrite reported, and it's the outside one.
        self.assertEqual(len(changes), 1)
        self.assertIn("#quick_start", changes[0])
        self.assertIn("#quick-start", changes[0])

    def test_global_and_scoped_entries_for_same_fingerprint_combine(self):
        # If both a global and a scoped entry exist for the same fingerprint,
        # the global one wins (every occurrence is suppressed). Documents the
        # union semantics rather than letting it be discovered by accident.
        self._set_whitelist(("quick_start", ("quick_start", "Quick Start")))
        body = "## Other\n[outside](#quick_start)\n"
        new, changes = gci._rewrite_section_anchors(body)
        self.assertEqual(new, body)
        self.assertEqual(changes, [])

    def test_invalid_scope_label_rejected_at_normalize(self):
        # "Sample setup" isn't in _SCOPABLE_LABELS — must abort, not silently
        # downgrade to a global entry that would surprise the user.
        with self.assertRaises(SystemExit):
            gci._normalize_whitelist_entry(("legacy-link", "Sample setup"))

    def test_malformed_entry_rejected(self):
        with self.assertRaises(SystemExit):
            gci._normalize_whitelist_entry(("only-one-element",))  # type: ignore[arg-type]
        with self.assertRaises(SystemExit):
            gci._normalize_whitelist_entry(123)  # type: ignore[arg-type]

    def test_scoped_entry_no_op_when_section_absent(self):
        # If the README doesn't contain the scoped section heading, a scoped
        # entry must NOT accidentally suppress matches elsewhere. It should
        # behave as if the entry didn't exist for that document.
        self._set_whitelist((("quick_start", "Quick Start"),))
        body = "## Other\n[x](#quick_start)\n"
        new, changes = gci._rewrite_section_anchors(body)
        self.assertIn("#quick-start", new)  # rewritten
        self.assertEqual(len(changes), 1)


class CheckDiagnosticsTests(unittest.TestCase):
    """The helpers that decorate `--check` failure output."""

    def test_format_anchor_change_adds_region_tag_and_source_line(self):
        # Simulate a 5-line README where line 3 is the offending anchor.
        content = "alpha\nbeta\n[bad](#FILE_MANAGEMENT)\ndelta\nepsilon\n"
        change = "  line 3: #FILE_MANAGEMENT → #file-management  (File Management)"
        decorated = gci._format_anchor_change(content, change)
        self.assertIn("[outside index]", decorated)
        # The source line itself must appear under the change, prefixed with
        # '> ' so it visually nests in CI logs.
        self.assertIn("> [bad](#FILE_MANAGEMENT)", decorated)

    def test_format_anchor_change_handles_missing_line(self):
        # If the change record references a line past EOF (shouldn't happen
        # in practice, but defensive), the helper must not crash and must
        # still emit the region tag.
        decorated = gci._format_anchor_change("only one line\n", "  line 99: #x → #y  (X)")
        self.assertIn("[outside index]", decorated)
        # No '> ' source preview when the line lookup returns empty.
        self.assertNotIn("\n      > ", decorated)

    def test_stale_regions_returns_empty_when_identical(self):
        # Wrap a body in a real BEGIN/END marker pair and feed the same
        # content as both `original` and `updated` — the diff should be empty.
        body = "<table><tr><td>same</td></tr></table>"
        wrapped = f"prefix\n{gci.HTML_BEGIN}\n{body}\n{gci.HTML_END}\nsuffix\n"
        self.assertEqual(gci._stale_regions(wrapped, wrapped), [])

    def test_stale_regions_emits_diff_for_changed_region(self):
        # Same wrapper, different body inside the markers — the helper must
        # report exactly one drifted region, named for the marker pair.
        before = f"{gci.HTML_BEGIN}\nrow A\nrow B\n{gci.HTML_END}\n"
        after = f"{gci.HTML_BEGIN}\nrow A\nrow C\n{gci.HTML_END}\n"
        drifted = gci._stale_regions(before, after)
        self.assertEqual(len(drifted), 1)
        name, diff = drifted[0]
        self.assertEqual(name, "COMMAND-INDEX:HTML")
        self.assertIn("-row B", diff)
        self.assertIn("+row C", diff)

    def test_stale_regions_skips_absent_markers(self):
        # Content that never had the markers in the first place must not
        # be reported as drifted — regions are opt-in.
        self.assertEqual(gci._stale_regions("plain readme\n", "plain readme\n"), [])


class NearestHeadingTests(unittest.TestCase):
    """The breadcrumb helper used by _format_anchor_change."""

    def test_returns_nearest_heading_at_any_level(self):
        content = (
            "# Top\n"          # line 1
            "intro\n"           # line 2
            "## Section A\n"    # line 3
            "body\n"            # line 4
            "### Sub A.1\n"     # line 5
            "[link]\n"          # line 6  ← target
        )
        result = gci._nearest_heading_above(content, 6)
        self.assertEqual(result, (5, "### Sub A.1"))

    def test_falls_back_to_higher_level_when_no_intervening_subheading(self):
        content = (
            "## Section\n"   # line 1
            "para\n"          # line 2
            "[link]\n"        # line 3  ← target, no sub-heading between
        )
        result = gci._nearest_heading_above(content, 3)
        self.assertEqual(result, (1, "## Section"))

    def test_returns_none_when_no_heading_precedes(self):
        content = "para 1\npara 2\n[link]\n"
        self.assertIsNone(gci._nearest_heading_above(content, 3))

    def test_ignores_atx_lookalike_inside_fenced_code_block(self):
        # A `# comment` inside a bash block is not a real heading. Without
        # fence-tracking the helper would incorrectly report it as the
        # nearest section header.
        content = (
            "## Real Section\n"   # line 1
            "before\n"             # line 2
            "```bash\n"            # line 3  (open fence)
            "# not a heading\n"    # line 4
            "echo hi\n"            # line 5
            "```\n"                # line 6  (close fence)
            "[link]\n"             # line 7  ← target
        )
        result = gci._nearest_heading_above(content, 7)
        self.assertEqual(result, (1, "## Real Section"))

    def test_target_line_itself_is_inclusive(self):
        # If the offending anchor is on the heading line itself (rare but
        # possible for `## [text](#anchor)`), the heading IS the breadcrumb.
        content = "## [Inline](#bad)\n"
        result = gci._nearest_heading_above(content, 1)
        self.assertEqual(result, (1, "## [Inline](#bad)"))

    def test_format_anchor_change_includes_under_breadcrumb(self):
        # End-to-end: _format_anchor_change must surface the heading line.
        content = "## Quick Start\n[bad](#FILE_MANAGEMENT)\n"
        change = "  line 2: #FILE_MANAGEMENT → #file-management  (File Management)"
        decorated = gci._format_anchor_change(content, change)
        self.assertIn("under: ## Quick Start (line 1)", decorated)
        self.assertIn("> [bad](#FILE_MANAGEMENT)", decorated)

    def test_format_anchor_change_omits_breadcrumb_when_no_heading(self):
        # No heading precedes the offending line — the breadcrumb line must
        # not appear at all (don't print "under: None").
        content = "[bad](#FILE_MANAGEMENT)\n"
        change = "  line 1: #FILE_MANAGEMENT → #file-management  (File Management)"
        decorated = gci._format_anchor_change(content, change)
        self.assertNotIn("under:", decorated)
        self.assertIn("> [bad](#FILE_MANAGEMENT)", decorated)

    def test_ignores_atx_inside_indented_code_block(self):
        # CommonMark: a line indented with 4+ spaces is an indented code
        # block, so a `#` there is code, not a heading.
        content = (
            "## Real\n"               # line 1
            "\n"                       # line 2
            "    # not a heading\n"   # line 3 (4-space indent → code)
            "[link]\n"                # line 4
        )
        result = gci._nearest_heading_above(content, 4)
        self.assertEqual(result, (1, "## Real"))

    def test_ignores_atx_inside_tab_indented_code_block(self):
        # A leading tab also marks an indented code block.
        content = (
            "## Real\n"          # line 1
            "\n"                  # line 2
            "\t# not heading\n"  # line 3 (tab → code)
            "[link]\n"           # line 4
        )
        result = gci._nearest_heading_above(content, 4)
        self.assertEqual(result, (1, "## Real"))

    def test_tracks_tilde_fences(self):
        # ~~~ fences must be recognised exactly like ``` fences.
        content = (
            "## Real\n"           # line 1
            "~~~\n"               # line 2 (open)
            "# not heading\n"    # line 3
            "~~~\n"               # line 4 (close)
            "[link]\n"            # line 5
        )
        result = gci._nearest_heading_above(content, 5)
        self.assertEqual(result, (1, "## Real"))

    def test_inner_shorter_fence_does_not_close_outer(self):
        # CommonMark: a closing fence must use the same char and length
        # ≥ the opener. An inner ``` inside a ```` block is content.
        content = (
            "## Real\n"               # line 1
            "````\n"                  # line 2 (open, len 4)
            "```\n"                   # line 3 (NOT a close — too short)
            "# still inside fence\n" # line 4
            "```\n"                   # line 5 (still not a close)
            "````\n"                  # line 6 (close)
            "[link]\n"                # line 7
        )
        result = gci._nearest_heading_above(content, 7)
        self.assertEqual(result, (1, "## Real"))

    def test_inner_different_char_fence_does_not_toggle(self):
        # An inner ~~~ inside a ``` block must not close it.
        content = (
            "## Real\n"            # line 1
            "```\n"                # line 2 (open ` fence)
            "~~~\n"                # line 3 (different char — content)
            "# not heading\n"     # line 4
            "~~~\n"                # line 5 (still content)
            "```\n"                # line 6 (close)
            "[link]\n"             # line 7
        )
        result = gci._nearest_heading_above(content, 7)
        self.assertEqual(result, (1, "## Real"))

    def test_indented_atx_up_to_three_spaces_still_counts(self):
        # CommonMark: ATX headings may have up to 3 leading spaces.
        content = (
            "intro\n"             # line 1
            "   ## Indented\n"   # line 2 (3 spaces — still a heading)
            "[link]\n"            # line 3
        )
        result = gci._nearest_heading_above(content, 3)
        # Stored verbatim including the 3-space indent.
        self.assertEqual(result, (2, "   ## Indented"))

    def test_fence_with_backticks_in_info_string_is_not_a_fence(self):
        # ``` openers forbid backticks in the info string. A line like
        # "```foo`bar" is not a real fence opener, so a later '#' line
        # outside any real fence remains a heading candidate.
        content = (
            "## Real\n"             # line 1
            "```foo`bar\n"          # line 2 (NOT a fence opener)
            "## Inner Heading\n"    # line 3 (a real heading)
            "[link]\n"              # line 4
        )
        result = gci._nearest_heading_above(content, 4)
        self.assertEqual(result, (3, "## Inner Heading"))


class BreadcrumbLevelsTests(unittest.TestCase):
    """The --breadcrumb-levels CLI option and its level-filtered helper."""

    def test_h2_only_skips_h3_subheading(self):
        # With H2 as the only permitted level, an H3 between the H2 and
        # the target line must be ignored — the H2 is reported instead.
        content = (
            "## Section A\n"   # line 1
            "para\n"            # line 2
            "### Sub A.1\n"     # line 3
            "[link]\n"          # line 4
        )
        result = gci._nearest_heading_above_levels(content, 4, frozenset({2}))
        self.assertEqual(result, (1, "## Section A"))

    def test_returns_none_when_only_disallowed_levels_precede(self):
        # Only H1/H3 precede; with H2 as the only permitted level the
        # helper must yield None rather than the closest disallowed one.
        content = (
            "# Top\n"           # line 1
            "### Deep\n"        # line 2
            "[link]\n"          # line 3
        )
        result = gci._nearest_heading_above_levels(content, 3, frozenset({2}))
        self.assertIsNone(result)

    def test_multi_level_set_picks_nearest_match(self):
        # Permit {2,3}: an H4 between an H3 and the target is ignored,
        # but the H3 is reported (closer than the H2 above).
        content = (
            "## Top\n"          # line 1
            "### Mid\n"         # line 2
            "#### Deep\n"       # line 3 (disallowed)
            "[link]\n"          # line 4
        )
        result = gci._nearest_heading_above_levels(content, 4, frozenset({2, 3}))
        self.assertEqual(result, (2, "### Mid"))

    def test_format_anchor_change_default_is_h2_only(self):
        # Default `breadcrumb_levels` arg is H2-only — an H3 closer to the
        # offending line must not appear in the breadcrumb.
        content = (
            "## Quick Start\n"           # line 1
            "### Install\n"               # line 2
            "[bad](#FILE_MANAGEMENT)\n"  # line 3
        )
        change = "  line 3: #FILE_MANAGEMENT → #file-management  (File Management)"
        decorated = gci._format_anchor_change(content, change)
        self.assertIn("under: ## Quick Start (line 1)", decorated)
        self.assertNotIn("### Install", decorated)

    def test_format_anchor_change_honours_caller_levels(self):
        # Passing {3} flips selection to the H3 instead of the H2.
        content = (
            "## Quick Start\n"           # line 1
            "### Install\n"               # line 2
            "[bad](#FILE_MANAGEMENT)\n"  # line 3
        )
        change = "  line 3: #FILE_MANAGEMENT → #file-management  (File Management)"
        decorated = gci._format_anchor_change(content, change, frozenset({3}))
        self.assertIn("under: ### Install (line 2)", decorated)

    def test_parse_breadcrumb_levels_basic_csv(self):
        self.assertEqual(gci._parse_breadcrumb_levels("2,3"), frozenset({2, 3}))

    def test_parse_breadcrumb_levels_default_token_is_h2(self):
        # The argparse default is the literal string "2".
        self.assertEqual(gci._parse_breadcrumb_levels("2"), frozenset({2}))

    def test_parse_breadcrumb_levels_clamps_out_of_range(self):
        # 0 → 1, 7 → 6. Both clamped values appear in the result.
        self.assertEqual(
            gci._parse_breadcrumb_levels("0,7"), frozenset({1, 6})
        )

    def test_parse_breadcrumb_levels_ignores_non_integer_tokens(self):
        # 'foo' is dropped; the rest survives.
        self.assertEqual(
            gci._parse_breadcrumb_levels("foo,3"), frozenset({3})
        )

    def test_parse_breadcrumb_levels_ignores_empty_tokens(self):
        # Trailing/leading commas and whitespace must not crash or pollute.
        self.assertEqual(
            gci._parse_breadcrumb_levels(" 2 , , 3 ,"), frozenset({2, 3})
        )

    def test_parse_breadcrumb_levels_falls_back_to_h2_when_empty(self):
        # Nothing valid → fall back to the documented default.
        self.assertEqual(gci._parse_breadcrumb_levels("foo,bar"), frozenset({2}))
        self.assertEqual(gci._parse_breadcrumb_levels(",,"), frozenset({2}))

    def test_parse_breadcrumb_levels_dedupes(self):
        self.assertEqual(
            gci._parse_breadcrumb_levels("2,2,3,3"), frozenset({2, 3})
        )


class IgnoredRegionsTests(unittest.TestCase):
    """IGNORED_REGIONS must suppress both generation and drift reporting."""

    def setUp(self):
        # Snapshot module-level ignore state so tests are isolated.
        self._saved_tuple = gci.IGNORED_REGIONS
        self._saved_set = gci._IGNORED_REGION_SET

    def tearDown(self):
        gci.IGNORED_REGIONS = self._saved_tuple
        gci._IGNORED_REGION_SET = self._saved_set

    def _set_ignored(self, names: tuple[str, ...]):
        gci.IGNORED_REGIONS = names
        gci._IGNORED_REGION_SET = frozenset(names)

    def test_canonical_region_names_covers_all_managed_regions(self):
        # Sanity check: the typo guard's allowlist must include every region
        # name `_all_regions()` produces, otherwise valid entries would be
        # rejected at module load.
        produced = {name for name, _, _ in gci._all_regions()}
        self.assertEqual(produced, gci._canonical_region_names())

    def test_is_region_ignored_respects_runtime_state(self):
        self._set_ignored(("COMMAND-INDEX:HTML",))
        self.assertTrue(gci._is_region_ignored("COMMAND-INDEX:HTML"))
        self.assertFalse(gci._is_region_ignored("COMMAND-INDEX:TEXT"))

    def test_stale_regions_skips_ignored_region(self):
        # A region whose body differs would normally be reported as drifted;
        # adding it to IGNORED_REGIONS must make the report drop it entirely.
        before = f"{gci.HTML_BEGIN}\nrow A\n{gci.HTML_END}\n"
        after = f"{gci.HTML_BEGIN}\nrow B\n{gci.HTML_END}\n"
        # Without ignoring: drift is reported.
        self.assertEqual(len(gci._stale_regions(before, after)), 1)
        # With ignoring: drift is suppressed.
        self._set_ignored(("COMMAND-INDEX:HTML",))
        self.assertEqual(gci._stale_regions(before, after), [])

    def test_replace_section_blocks_skips_ignored_section(self):
        # Build a doc with a SECTION-CMDS region whose body differs from
        # what render_section_block() would generate. With the section
        # ignored, _replace_section_blocks must NOT touch it.
        label = gci.SECTION_LABELS[1]  # "File Management"
        begin = gci.SECTION_CMDS_BEGIN.format(label=label)
        end = gci.SECTION_CMDS_END.format(label=label)
        stale_body = "```bash\nFROZEN — do not regenerate\n```"
        doc = f"prefix\n{begin}\n{stale_body}\n{end}\nsuffix\n"
        self._set_ignored((f"SECTION-CMDS:{label}",))
        new_doc, processed = gci._replace_section_blocks(doc)
        self.assertEqual(new_doc, doc)
        self.assertNotIn(label, processed)


if __name__ == "__main__":
    unittest.main()