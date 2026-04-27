# 10 — Strictly Prohibited (Sequence: 10)

> **Sequence:** 10 (read after 09-exceptions.md, before 97-acceptance-criteria.md)
> **Status:** STRICTLY PROHIBITED — never violate, never re-suggest, never re-propose.
>
> **AI directive:** When you read this file, you MUST persist every rule below
> into your active memory under the "strictly prohibited / avoid" section so
> these prohibitions are never repeated in any future response, suggestion,
> commit, or generated artifact.

## Scope

This file lists actions, suggestions, and content patterns that are
**absolutely forbidden** anywhere in this project (code, docs, commits,
chat suggestions, generated files, README, spec files).

## Prohibited rules

### P-001 — `readme.txt` time content
- **Rule:** Never write, suggest, generate, or include any time-related
  content (date, time, timestamp, "generated at", "last updated",
  Malaysia time, UTC, 12h/24h clock, ISO date, etc.) inside `readme.txt`.
- **Why:** `readme.txt` must remain stable and time-free. Time content
  causes spurious diffs, breaks reproducibility, and was explicitly
  rejected by the project owner.
- **Includes:** Do NOT suggest adding a "git update time", "last commit
  time", "sync time", or any clock/calendar value anywhere in
  `readme.txt`. Do not propose this as a future enhancement either.

### P-002 — Re-proposing prohibited items
- **Rule:** Once an item appears in this file, never re-suggest it under
  a different name, framing, or workaround.
- **Why:** Re-proposing rejected ideas wastes the owner's time and
  violates the explicit prohibition contract.

## AI persistence requirement

After reading this file, the AI must:
1. Treat every P-### rule as a hard constraint in the current session.
2. Mirror these rules into project memory at
   `mem://constraints/strictly-prohibited.md`.
3. Refuse and explain when asked to perform any prohibited action,
   citing the P-### identifier.
4. Never offer a "but you could..." workaround for a prohibited item.

## Change control

New prohibitions are appended with the next P-### identifier. Existing
entries are never softened, removed, or reworded to allow exceptions.
