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

    def tearDown(self):
        gci.ANCHOR_WHITELIST = self._saved_list
        gci._WHITELIST_FINGERPRINTS = self._saved_fp

    def _set_whitelist(self, entries: tuple[str, ...]):
        import re
        gci.ANCHOR_WHITELIST = entries
        gci._WHITELIST_FINGERPRINTS = frozenset(
            re.sub(r"[^a-z0-9]", "", a.lower()) for a in entries
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