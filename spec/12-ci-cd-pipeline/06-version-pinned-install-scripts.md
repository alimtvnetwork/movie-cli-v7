# 06 — Version-Pinned Install Scripts on Release Pages

## Purpose

When a GitHub Release is published, the `install.ps1` and `install.sh`
attached as release assets MUST install **that exact release version** —
never "latest", never auto-upgraded to a newer sibling repo, never
delegated to `bootstrap.{sh,ps1}`.

This is a hard contract: a user who copies the install one-liner from
the **release page of `vX.Y.Z`** must end up with `movie vX.Y.Z`
installed, even if `vX.Y.Z+1` exists.

---

## Why

1. **Reproducibility** — CI pipelines, Dockerfiles, and onboarding
   docs that pin to a release URL must keep working forever.
2. **Rollback safety** — Users intentionally installing an older
   release (e.g. to bisect a bug) must get exactly what they asked for.
3. **Trust** — A release page that silently installs a different
   version is a footgun. The version printed on the page must match
   the version installed.

---

## How it works

### Generation (`.github/workflows/release.yml`)

The release pipeline generates `install.ps1` and `install.sh` with the
release version baked in as a literal string:

```bash
sed -i "s|VERSION_PLACEHOLDER|$VERSION|g" dist/install.ps1
sed -i "s|VERSION_PLACEHOLDER|$VERSION|g" dist/install.sh
```

Inside each script:

```powershell
$PinnedVersion = "VERSION_PLACEHOLDER"   # → becomes "v2.128.7" at release time
```

```bash
PINNED_VERSION="VERSION_PLACEHOLDER"     # → becomes "v2.128.7" at release time
```

### Release-page one-liner

The release body (rendered on the GitHub Release page) shows:

```powershell
irm https://github.com/<owner>/<repo>/releases/download/v2.128.7/install.ps1 | iex
```

```bash
curl -fsSL https://github.com/<owner>/<repo>/releases/download/v2.128.7/install.sh | bash
```

Both URLs reference the **specific tag** (`v2.128.7`), not `latest`.
The script downloaded from that URL has `PINNED_VERSION="v2.128.7"`
hard-coded, so the installer downloads the matching archive.

### Install behaviour (both scripts)

```
1. PINNED_VERSION is a baked-in literal — never derived at runtime.
2. Script downloads movie-${PINNED_VERSION}-${os}-${arch}.{zip|tar.gz}
   from /releases/download/${PINNED_VERSION}/.
3. Script downloads checksums.txt from the SAME release.
4. Script verifies SHA256 hash.
5. Script installs the binary.
6. NO sibling-repo probing. NO bootstrap delegation. NO --version flag
   override (install.sh exposes one for power users, but the default
   is always the pinned literal).
```

---

## What this is NOT

| Concern | Where it lives |
|---------|----------------|
| "Always install the newest release" | `README.md` quick-install uses `/releases/latest/download/...` |
| "Auto-discover the latest sibling repo (`-vN+1`, `-vN+2`)" | `bootstrap.sh` / `bootstrap.ps1` (separate file, separate URL) |
| "Self-update an installed binary" | `movie update` command (separate code path) |

These three flows MUST stay independent. The release-page installer
is the **anchor** — it is the only one that guarantees an exact
version match.

---

## Acceptance Criteria

- GIVEN a freshly-published release `vX.Y.Z`
  WHEN the user copies the install one-liner from the release page
  THEN running it installs exactly `movie vX.Y.Z`.

- GIVEN release `vX.Y.Z` is published AND `vX.Y.Z+1` is later published
  WHEN a user runs the install one-liner from the `vX.Y.Z` release page
  THEN they still get `vX.Y.Z` (NOT `vX.Y.Z+1`).

- GIVEN the generated `install.ps1` or `install.sh`
  WHEN inspected
  THEN `PINNED_VERSION` / `$PinnedVersion` is a literal string equal
  to the release tag — NOT a runtime lookup.

- GIVEN the generated install scripts
  WHEN inspected
  THEN they contain NO call to `bootstrap.sh`, `bootstrap.ps1`, or any
  sibling-repo probing logic.

---

## Prevention rule for future AI edits

**NEVER** add to a release-page install script:
- `releases/latest/download/...`
- A call to `bootstrap.{sh,ps1}`
- A "check for newer version" prompt
- A `--version` argument default that resolves at runtime

The pinned literal is the contract. If "always-latest" install is
needed, that lives in `README.md` (which uses the GitHub `/latest/`
redirect) — never in the per-release script.

---

*Version-pinned install scripts spec — created: 2026-04-20*
