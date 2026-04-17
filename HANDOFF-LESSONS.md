# Self-Update Handoff — Lessons Learned

> A short, AI-shareable document that captures **why the obvious fix is
> wrong** for self-updating CLIs on Windows.
> Read this before changing `updater/` in this repo or implementing a
> similar updater anywhere else.

---

## The problem in one sentence

A running `.exe` on Windows holds a file lock on itself, so it cannot
overwrite its own binary, and it cannot be deleted while running.

## The pattern that works (do this)

1. **Process A** (the running `movie.exe`) copies itself to
   `movie-update-<pid>.exe` next to the original.
2. Process A starts **Process B** from that copy **detached with its own
   console** (`CREATE_NEW_CONSOLE` on Windows, `setsid` on Unix).
3. **Process A exits immediately with code 0.** The OS releases the lock
   on `movie.exe`.
4. Process B runs the build/deploy script (`run.ps1`) which freely
   overwrites `movie.exe`.
5. The script ends by spawning a tiny **detached self-deleter**
   (`cmd /c ping 127.0.0.1 -n 3 & del movie-update-<pid>.exe`).
   Process B exits; ~2 s later the deleter removes the now-idle copy.

A separate `update-cleanup` command exists as a belt-and-braces sweeper
for copies left over from crashes / Ctrl-C; it must `--skip-path` the
worker that is currently running.

## The pattern that looks right but is wrong (do not do this)

> "I'll make the parent block on the worker with `cmd.Run()` so the user
> keeps seeing output in the same terminal."

This re-introduces the exact bug the handoff exists to avoid:

- Process A is still alive → still holds the lock on `movie.exe`.
- The deploy step in `run.ps1` cannot overwrite the active-PATH binary
  → "Active PATH binary is in use; retrying (1/5..5/5)" → "Could not
  sync active PATH binary after retries."
- The cleanup step at the end tries to delete the worker file, but
  the worker file IS Process B currently executing it →
  "Could not remove movie-update-<pid>.exe: Access is denied."
- Net result: deploy *appears* to succeed (because of the rename-first
  trick on the *other* deploy dir) but the binary on `PATH` is never
  updated and the temp worker accumulates forever.

If your motivation is "the user loses console output when the parent
detaches", solve that **inside the new worker console** (use
`CREATE_NEW_CONSOLE`, format `Write-Host` nicely, write a log file).
Do **not** solve it by keeping the parent alive.

## Hard rules

| Rule | Why |
|------|-----|
| Parent exits 0 immediately after spawning the worker | Releases the file lock. |
| Worker runs in its own console window (Windows) | Keeps progress visible. |
| Worker self-deletes via a detached `cmd /c del` after exit | Avoids "Access is denied" on cleanup. |
| `update-cleanup` always honours `--skip-path` for the live worker | Idempotent re-runs are safe. |
| Never use `cmd.Run()` (blocking) for the handoff launch | This is the bug, not the fix. |

## Apology / regression history

In `iteration 3` (16-Apr-2026) we "fixed" a perceived console-detachment
issue by switching the handoff to **synchronous `cmd.Run()`**. That was
wrong and broke real users with the symptoms above. Iteration 4
(17-Apr-2026) reverts to the detached pattern documented here and adds
the worker self-deleter so cleanup never races against the running
worker.

If you are an AI/contributor reading this: **do not "improve" the
handoff back into a blocking call**. The trade-off has been measured on
real Windows machines and the detached pattern is the only one that
works.

---

*Maintained alongside `spec/13-self-update-app-update/03-copy-and-handoff.md`.*
