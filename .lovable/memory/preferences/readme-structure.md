---
name: readme-structure
description: Required order and rules for the root README.md — write it first, then update; never reorder sections
type: preference
---

# Root README structure (write first, then update)

## Required top-of-README order
1. **Title** (H1 with project name + emoji)
2. **Badges** (CI, release, version, downloads, language, platform, etc.)
3. **Image / logo / hero screenshot** (icon and/or demo gif)
4. **Installation** (one-liner per OS, in this order):
   - 🪟 Windows · PowerShell (FIRST)
   - 🐧 macOS · Linux · Bash (SECOND)
5. **Why this exists** — short, human story explaining the frustration
   that motivated the project (e.g. old DVDs, messy filenames, no ratings)
   and what the tool gives back to the user.
6. Everything else (Highlights, Features, Usage, Installation details, etc.)

## How to apply
- **Write the root README FIRST** when shaping the project narrative,
  THEN update sub-docs (QUICKSTART, spec/03-general/01-install-guide.md,
  CONTRIBUTING) to match. Never let sub-docs drift ahead of README.
- Install headers must read exactly:
  - `🪟 Windows · PowerShell`
  - `🐧 macOS · Linux · Bash`
- Keep the install block immediately after the hero image, before any
  highlights/features list.
- The "Why this exists" section is mandatory and must stay in plain,
  layman language — no jargon, no feature list dump.

**Why:** Users land on README, scan top-down, and decide in seconds
whether to install. Title → proof (badges) → visual → install → story
is the conversion path. Reordering breaks first-impression flow.
