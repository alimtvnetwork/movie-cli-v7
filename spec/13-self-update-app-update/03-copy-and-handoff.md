# 03 — Copy-and-Handoff Mechanism

> **Status**: ✅ Authoritative — supersedes all prior versions.
> **Last rewritten**: 2026-04-17 after the **synchronous-handoff regression**
> documented at the bottom of this file. Read that section before changing
> anything in `updater/handoff.go` or `updater/script.go`.

## Purpose

Define how the running binary creates a temporary copy of itself and
delegates the update work to that copy, **then exits immediately**, so the
OS releases its file lock before `run.ps1` tries to overwrite it.

---

## Why Handoff is Needed

On Windows, a running `.exe` holds a file lock. If the build pipeline tries
to overwrite the binary while the original process is still alive, the OS
blocks it. The handoff solves this by:

1. The running binary is **Process A** (holds the lock on `movie.exe`).
2. Process A copies itself to `movie-update-<pid>.exe` (a different file).
3. Process A launches **Process B** from the copy **detached** with its
   own console (`CREATE_NEW_CONSOLE` on Windows, `setsid` on Unix).
4. **Process A exits immediately with code 0.** ← critical step
   - The OS releases the lock on the original `movie.exe`.
   - The user's terminal returns to the prompt; Process B keeps running
     in its own window and prints the update progress there.
5. Process B runs `run.ps1` which freely overwrites `movie.exe`.
6. When Process B is done, it spawns a tiny **detached self-deleter**
   (`cmd /c ping ... & del movie-update-<pid>.exe`) and exits. The
   self-deleter waits ~2 s, then removes the now-idle worker binary.

> **Forbidden alternative.** Do **NOT** keep Process A alive with
> `cmd.Run()` to "preserve the console". That keeps the lock on
> `movie.exe`, makes `run.ps1`'s active-PATH-binary sync loop fail with
> *"Active PATH binary is in use; retrying"*, and makes the cleanup step
> die with *"Access is denied"* when it tries to delete the worker that is
> still running. See the regression note at the bottom.

---

## Copy Location

The handoff copy is placed at:

1. **Same directory as the binary** (preferred):
   `<binary-dir>/movie-update-<pid>.exe`
2. **Fallback to temp directory**:
   `%TEMP%/movie-update-<pid>.exe` (Windows)
   `/tmp/movie-update-<pid>` (Unix)

The PID suffix ensures uniqueness if multiple updates run simultaneously.

---

## File Naming

| OS | Format |
|----|--------|
| Windows | `movie-update-<pid>.exe` |
| Linux/macOS | `movie-update-<pid>` |

---

## Launch Arguments

The parent launches the copy with:

```
movie-update-12345.exe update-runner --repo-path <repo> --target-binary <orig>
```

- `update-runner` is a **hidden command** (not shown in help).
- `--repo-path` passes the resolved repo path so the worker doesn't
  re-resolve it.
- `--target-binary` is the **original** `movie.exe` path the user
  launched, so `run.ps1` can deploy back to the exact same location.

---

## Detached Execution (CRITICAL)

The parent MUST launch the worker **detached with its own console**, then
**exit 0 immediately**:

```go
// Windows
cmd := exec.Command(copyPath, args...)
cmd.SysProcAttr = &syscall.SysProcAttr{
    CreationFlags: 0x00000010 | 0x00000200, // CREATE_NEW_CONSOLE | CREATE_NEW_PROCESS_GROUP
}
if err := cmd.Start(); err != nil {
    return err
}
_ = cmd.Process.Release()  // detach handle so the parent can exit cleanly
return nil                  // ← parent returns; main exits 0
```

```go
// Unix
cmd := exec.Command(copyPath, args...)
cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
cmd.Stdout, cmd.Stderr = nil, nil  // fully detached
if err := cmd.Start(); err != nil {
    return err
}
_ = cmd.Process.Release()
return nil
```

Why a **new console** (Windows) and not just `DETACHED_PROCESS`:
without a console the worker has no `Stdout`/`Stderr` to write progress
into and PowerShell's `Write-Host` swallows everything. A new console
window keeps the live output visible to the user.

---

## Exit Code Propagation

Because the parent exits as soon as the worker is launched, **the parent
always exits 0** for a successful handoff. The worker is responsible for
its own exit code in its own console. This is the same model used by
`gitmap update` and is acceptable because:

- The user already sees the worker's output in the new console window.
- The original shell prompt returns instantly — terminals are never left
  hanging on a long-running update.
- Failures inside the worker are visible in the worker's console and
  recorded by `errlog`.

---

## Self-Deletion of the Worker

The worker cannot delete itself while it is running. The temp
PowerShell script that the worker executes ends with a small detached
deleter:

```powershell
# Inside the temp update-script-*.ps1, AFTER everything else:
$deleter = "cmd.exe"
$deleterArgs = @(
    "/c", "ping", "127.0.0.1", "-n", "3", ">", "nul", "&",
    "del", "/f", "/q", "`"$workerBinary`""
)
Start-Process -FilePath $deleter -ArgumentList $deleterArgs -WindowStyle Hidden
```

The `ping ... -n 3` gives the worker ~2 s to exit before `del` runs.
On Unix the equivalent is `(sleep 2 && rm -f "$workerBinary") &`.

`movie update-cleanup` is still the **belt-and-braces sweeper** for any
worker copies that escaped the self-deleter (e.g. crashed runs, Ctrl-C).
It must `--skip-path` the currently-running worker to stay idempotent.

---

## Unix Behaviour

On Linux/macOS in-place binary replacement is allowed by the kernel, so
strictly speaking the handoff is not required. We still use the same
detached pattern on all platforms for code symmetry and so the parent
shell prompt always returns immediately.

---

## Pseudocode (full flow)

```go
func Run(repoPathFlag string) error {
    repoPath, _ := findRepoPath(repoPathFlag)
    selfPath, _  := resolveSelfPath()
    copyPath, _  := createHandoffCopy(selfPath)

    // Detached spawn → parent returns → main() exits 0 → lock released.
    return launchHandoffDetached(copyPath, repoPath, selfPath)
}

func RunWorker(repoPath, targetBinary string) error {
    // We are now Process B, running from <binary-dir>/movie-update-<pid>.exe.
    // The original movie.exe is unlocked.
    return executeUpdate(repoPath, targetBinary)
    // executeUpdate() writes a temp .ps1 that ends with the self-deleter.
}
```

---

## Regression Note — 2026-04-16 → 2026-04-17

**What went wrong.** Issue
[`05-updater-async-console`](../../.lovable/memory/issues/05-updater-async-console.md)
("Updater Async Console Breakage") was fixed in the *opposite direction*
of this spec: the parent was changed to **`cmd.Run()` (blocking)** so
that "the console stays attached". That solved the cosmetic console
detachment but **re-introduced the original Windows file-lock bug** that
the handoff exists to avoid. Symptoms reported by the user:

- `run.ps1`'s active-PATH-binary sync loop printed *"Active PATH binary
  is in use; retrying (1..5/5)"* and then *"Could not sync active PATH
  binary after retries."*
- The `update-cleanup` step at the end of the script printed
  *"Could not remove movie-update-<pid>.exe: Access is denied"*
  because that file was the still-running worker.

**Apology.** The synchronous fix was a mistake. The correct trade-off is
"new console window for the worker, parent exits immediately" — that
keeps progress visible to the user **and** releases the lock. The
synchronous variant cannot satisfy both constraints on Windows.

**Hard rules going forward.**

- `updater/handoff.go` MUST use detached spawn + parent exit. Do not
  reintroduce `cmd.Run()` there. Ever.
- The temp update script MUST end with a detached self-deleter for the
  worker binary. Do not rely solely on `update-cleanup`.
- Any future "console looks weird" complaint must be solved inside the
  new worker console (e.g. `Write-Host` formatting), not by making the
  parent block.

---

*Copy-and-handoff mechanism — rewritten: 2026-04-17*
