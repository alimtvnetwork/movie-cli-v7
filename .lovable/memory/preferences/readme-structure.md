---
name: readme-structure
description: Root README is the source of truth — write/update it FIRST, then propagate changes to sub-docs. Required section order is fixed.
type: preference
---

# Root README rule (write FIRST, then update sub-docs)

## The rule (non-negotiable)
**The root `README.md` is the canonical narrative for the project.**
Whenever project messaging, install steps, features, or positioning change:

1. **Edit `README.md` FIRST.** Land the change there before touching anything else.
2. **THEN propagate** the same change to sub-docs that mirror it:
   - `QUICKSTART.md`
   - `spec/03-general/01-install-guide.md`
   - `CONTRIBUTING.md`
   - any other doc that repeats README content
3. **Never** let a sub-doc ship an instruction, command, or claim that the
   root README does not already state. Sub-docs may expand on README, never
   contradict or precede it.

If you catch yourself opening a sub-doc first, **stop and open README.md instead.**

## Required top-of-README order (fixed)
1. **Title** (H1 with project name + emoji)
2. **Badges** (CI, release, version, downloads, language, platform, etc.)
3. **Image / logo / hero screenshot**
4. **Installation** — one-liner per OS, in this exact order with these exact headers:
   - `🪟 Windows · PowerShell` (FIRST)
   - `🐧 macOS · Linux · Bash` (SECOND)
5. **Why this exists** — short, plain-language story explaining the
   frustration that motivated the project (old DVDs, messy filenames,
   no ratings) and what the tool gives back to the user.
6. Everything else (Highlights, Features, Usage, detailed Installation, etc.)

## How to apply
- Install block sits **immediately after the hero image**, before highlights.
- Install headers must read **exactly**:
  - `🪟 Windows · PowerShell`
  - `🐧 macOS · Linux · Bash`
- The "Why this exists" section is **mandatory** and must stay in
  layman language — no jargon, no feature-list dump.
- When updating install commands, version pins, or project tagline:
  README.md → QUICKSTART.md → spec install guide → everything else.

**Why:** Users land on README and decide in seconds whether to install.
Title → proof (badges) → visual → install → story is the conversion path.
Sub-docs that drift ahead of README cause contradictions, broken
copy-paste commands, and wasted trust. README-first eliminates drift at
the source.
