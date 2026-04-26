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
        body = "See [scan](#scan--library) and [hist](#history-undo)."
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


if __name__ == "__main__":
    unittest.main()