# 07 — Release Dual-Trigger Race (Partial Release with Missing Archives)

**Severity**: Critical — produces visibly broken releases that 404 the install script.
**First observed**: v2.97.0 (2026-04-16)
**Status**: Fixed in v2.128.5 (workflow now triggers ONLY on `release/**` branches; tag trigger removed).

---

## Symptom

A published GitHub Release contains only a subset of the 6 expected platform archives. The `install.ps1` / `install.sh` one-liners fail with HTTP 404 for the missing archive.

Example — v2.97.0 was published with only 3 of 6 archives:

```
✅ checksums.txt
✅ install.ps1
✅ install.sh
✅ movie-v2.97.0-darwin-amd64.tar.gz
✅ movie-v2.97.0-darwin-arm64.tar.gz
✅ movie-v2.97.0-linux-arm64.tar.gz
❌ movie-v2.97.0-windows-amd64.zip       ← MISSING
❌ movie-v2.97.0-windows-arm64.zip       ← MISSING
❌ movie-v2.97.0-linux-amd64.tar.gz      ← MISSING
```

The build step succeeded — all 6 binaries were produced and zipped. Assets vanished during the publish step.

---

## Root Cause — Two Workflow Runs, Same Release

The original workflow triggered on **both** `release/**` branch pushes AND `v*` tag pushes:

```yaml
on:
  push:
    branches: ["release/**"]
    tags:     ["v*"]
```

`softprops/action-gh-release@v2` **creates the tag** as part of publishing the release. So a single `git push origin release/v2.97.0` produced this sequence:

1. Branch push fires Run A (`release/v2.97.0`).
2. Run A builds 6 binaries, zips them, calls `softprops/action-gh-release` which **creates tag `v2.97.0`** and uploads all 6 archives. ✅ Run A succeeds.
3. The new tag immediately fires Run B (`v2.97.0`) — only 1–2 seconds after Run A started.
4. Run B builds its own 6 binaries in parallel, then calls `softprops/action-gh-release` for the **same tag** that Run A already published.
5. `softprops/action-gh-release@v2` in upload mode **deletes existing assets and re-uploads** when the release already exists. Run B's upload starts replacing assets, then **fails partway through** (network blip, 422 conflict, or asset upload race), leaving the release with only the assets Run B happened to upload before failing.

### Evidence (v2.97.0)

Two GitHub Actions runs for the same release, fired 2 seconds apart:

| Run ID | Trigger ref | Conclusion | Final step |
|---|---|---|---|
| 24534322512 | `release/v2.97.0` (branch) | ✅ success | Created release with 6 archives |
| 24534323295 | `v2.97.0` (tag, fired by Run A's release publish) | ❌ failure | "Create GitHub Release" step failed |

Both runs successfully built and zipped 6 binaries. The race is entirely in the publish step.

### Why concurrency groups did not save us

The original concurrency key was `release-${{ github.ref }}`:

- Run A's `github.ref` = `refs/heads/release/v2.97.0`
- Run B's `github.ref` = `refs/tags/v2.97.0`

Different refs → different concurrency groups → the runs did NOT serialize.

---

## The Fix — Single Trigger Ref

Remove the tag trigger entirely. The release tag is **produced** by the release run, not a precondition for it.

```yaml
on:
  push:
    branches:
      - "release/**"
  ## NO `tags:` block — softprops/action-gh-release creates the tag at publish.
```

And reject any unexpected ref in the version resolver:

```bash
if [[ "$GITHUB_REF" == refs/heads/release/* ]]; then
  VERSION="${GITHUB_REF_NAME#release/}"
else
  echo "::error::Unexpected ref: $GITHUB_REF (this workflow only runs on release/** branches)"
  exit 1
fi
```

### Why this is safe

- The release branch push is the single source of truth for "publish version X".
- The tag is a *side effect* of publishing — never a trigger.
- One workflow run per release ⇒ no race ⇒ no partial uploads possible.
- The existing per-branch concurrency group (`release-${{ github.ref }}`) still protects against re-pushing the same branch racing itself.

---

## Prevention Rules

1. **Never trigger a release workflow on both branch AND tag for the same release**, when the workflow is the thing that creates the tag.
2. **If both triggers are required for distinct flows**, the concurrency group MUST normalize across them (same group key for branch + tag of the same version) AND `cancel-in-progress: false` so the second run waits.
3. **`cancel-in-progress: false` is correct** for releases — partial uploads must finish, not be killed mid-flight. But it does NOT protect against parallel runs in different concurrency groups.
4. **The "verify all 6 archives present" guard** (added in v2.128.4) only protects against build failures — it CANNOT prevent post-publish clobbering by a second workflow run.

---

## How to Recover a Partially-Uploaded Release

Until the partial release is republished or deleted, install scripts will 404 for the missing archives. Options:

1. **Re-push the release branch** (after the fix lands) — this re-runs the workflow, which will overwrite the partial release with all 6 archives.
2. **Delete the partial release + tag**, then re-push:
   ```bash
   gh release delete v2.97.0 --yes --cleanup-tag
   git push --delete origin release/v2.97.0
   git push origin release/v2.97.0
   ```
3. **Bump to a fresh patch version** (cleanest) — `release/v2.97.1`.

---

## Acceptance Criteria

- GIVEN a `release/vX.Y.Z` branch push WHEN the workflow runs THEN exactly **one** workflow run is triggered (no second run from the resulting tag).
- GIVEN the workflow file WHEN inspected THEN there is no `tags:` entry under `on.push`.
- GIVEN any future release WHEN published THEN it contains exactly 6 archives + `checksums.txt` + `install.ps1` + `install.sh`.

---

*CI/CD issue log — added: 2026-04-19*
