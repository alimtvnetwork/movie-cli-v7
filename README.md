<div align="center">

<img src="assets/icon.png" alt="Movie CLI icon" width="80" height="80">

# 🎬 Movie CLI

**Personal movie & TV show library manager — from the terminal**

[![CI](https://github.com/alimtvnetwork/movie-cli-v6/actions/workflows/ci.yml/badge.svg)](https://github.com/alimtvnetwork/movie-cli-v6/actions/workflows/ci.yml)
[![GitHub Release](https://img.shields.io/github/v/release/alimtvnetwork/movie-cli-v6?style=flat-square&label=version)](https://github.com/alimtvnetwork/movie-cli-v6/releases)
[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev/)
[![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20Linux%20%7C%20macOS-lightgrey?style=flat-square)](https://github.com/alimtvnetwork/movie-cli-v6)
[![SQLite](https://img.shields.io/badge/SQLite-WAL-003B57?style=flat-square&logo=sqlite&logoColor=white)](https://www.sqlite.org/)
[![Go Report Card](https://goreportcard.com/badge/github.com/alimtvnetwork/movie-cli-v6?style=flat-square)](https://goreportcard.com/report/github.com/alimtvnetwork/movie-cli-v6)
[![License](https://img.shields.io/badge/license-MIT-blue?style=flat-square)](https://github.com/alimtvnetwork/movie-cli-v6/blob/main/LICENSE)

_Scan folders, clean filenames, fetch TMDb metadata, organize files, and track your collection._

</div>

---

<div align="center">

## ✨ Highlights

</div>

- 🔍 **Smart scan** — recursively walks folders, cleans messy release names, and matches them against TMDb
- 🖼️ **Posters & metadata** — automatic thumbnail downloads, ratings, genres, cast, runtime
- 📦 **Single binary** — one statically-linked Go executable, no runtime, no dependencies
- 🗂️ **SQLite (WAL)** — fast, durable, zero-config local database in `./data/movie.db`
- ↩️ **Undo / redo** — every move, rename, scan, and delete is reversible
- 🌐 **REST API + web UI** — `movie rest --open` launches a local dashboard
- 🛠️ **Self-updating** — `movie update` pulls, rebuilds, and hands off in-place
- 🔒 **Cross-platform** — Windows, Linux, macOS on `amd64` and `arm64`

---

<div align="center">

## 📑 Table of Contents

</div>

- [Quick Start](#quick-start)
- [Jump to a command](#jump-to-a-command)
- [Demo](#-demo)
- [Installation](#installation)
- [What It Does](#what-it-does)
- [Command Reference](#command-reference)
- [Command Tree](#command-tree)
- [Build & Deploy](#build--deploy)
- [Release Workflow](#release-workflow)
- [Project Structure](#project-structure)
- [Data Storage](#data-storage)
- [Milestones](#milestones)
- [Dependencies](#dependencies)
- [Contributing](#-contributing)
- [Author](#author)
- [License](#license)

---

<div align="center">

## Quick Start

</div>

### Install latest release

Picks up whatever is currently tagged `latest` on GitHub — and if no release has been published yet, automatically falls back to a source-build from `main` so you still end up with a working binary.

**Windows (PowerShell)**

```powershell
irm https://raw.githubusercontent.com/alimtvnetwork/movie-cli-v6/main/get.ps1 | iex
```

**Linux / macOS**

```bash
curl -fsSL https://raw.githubusercontent.com/alimtvnetwork/movie-cli-v6/main/get.sh | bash
```

> The bootstrap probes `releases/latest/download/install.{ps1,sh}` first. If a release exists, it installs the pre-built binary. If not, it transparently falls back to cloning and building from `main` — and prints exactly which path it took. See [Installation](#installation) for flags and details.

### Install a specific version (pinned)

Installs exactly the version in the URL — never auto-upgrades. Use this for CI pipelines, Dockerfiles, reproducible setups, or when you need to roll back. Replace `v2.130.0` with the [release tag](https://github.com/alimtvnetwork/movie-cli-v6/releases) you want.

**Windows (PowerShell)**

```powershell
irm https://github.com/alimtvnetwork/movie-cli-v6/releases/download/v2.130.0/install.ps1 | iex
```

**Linux / macOS**

```bash
curl -fsSL https://github.com/alimtvnetwork/movie-cli-v6/releases/download/v2.130.0/install.sh | bash
```

> **Which one should I use?** Use **latest** for personal machines so you stay current. Use **pinned** anywhere reproducibility matters — the pinned script is hard-locked to the version in the URL and will install that exact tag forever, even after newer releases ship. ([contract spec](spec/12-ci-cd-pipeline/06-version-pinned-install-scripts.md))

### Set up & scan

```bash
movie config set tmdb_api_key YOUR_KEY
movie scan ~/Downloads
movie ls
```

### Search & discover

```bash
movie search "Inception"
movie suggest 5
```

Every command supports `--help` or `-h` for detailed usage.

---

<div align="center">

## Jump to a command

</div>

Skip the demo and jump straight to the command you need. Each link drops you into the matching **Command Reference** subsection — with the animated walkthrough, copy-paste Bash + PowerShell examples, expected output, and the full subcommand table.

| | Section | What's there | Top commands |
|---|---|---|---|
| 📂 | [Scanning & Library](#scanning--library) | Match files against TMDb, browse the library | `movie scan`, `movie rescan`, `movie ls`, `movie search`, `movie info` |
| 📦 | [File Management](#file-management) | Move, rename, flatten, play files | `movie move`, `movie rename`, `movie popout`, `movie play`, `movie cd` |
| ↩️ | [History & Undo](#history--undo) | Reverse any move/rename/scan/delete | `movie undo`, `movie undo --list`, `movie undo --id`, `movie redo` |
| 🎯 | [Discovery & Organization](#discovery--organization) | Recommendations, genres, tags, watchlist | `movie suggest`, `movie discover`, `movie tag`, `movie watch`, `movie stats` |
| 🛠 | [Maintenance & Debugging](#maintenance--debugging) | Stale-entry cleanup, logs, REST server | `movie cleanup`, `movie db`, `movie logs`, `movie rest`, `movie export` |
| ⚙️ | [Configuration & System](#configuration--system) | Settings, TMDb key, version, self-update | `movie config`, `movie config set`, `movie version`, `movie update` |
| 🚑 | [Troubleshooting](#troubleshooting) | Common errors and how to fix them | `tmdb_api_key not set`, `429`, `database is locked`, stale entries |

> First time here? Run the **[env-var check](#command-reference)** at the top of the Command Reference to confirm `TMDB_KEY` is set before you scan.

---

<div align="center">

## 🎥 Demo

</div>

### 📂 Scanning a Folder

<!-- Replace with actual GIF: docs/screenshots/demo-scan.gif -->
<!-- Record with: vhs docs/screenshots/scan.tape  or  asciinema rec -->

```
$ movie scan ~/Downloads

🔍 Scanning: /home/user/Downloads
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  Found 12 video files

  [1/12] Scream.2022.1080p.WEBRip.x264-RARBG.mkv
         → Title: Scream (2022)
         → TMDb: ★ 6.8 | Horror, Mystery, Thriller
         → Poster saved: thumbnails/scream-2022/scream-2022.jpg
         ✅ Saved to database

  [2/12] The.Batman.2022.2160p.BluRay.x265.mkv
         → Title: The Batman (2022)
         → TMDb: ★ 7.7 | Crime, Mystery, Thriller
         → Poster saved: thumbnails/the-batman-2022/the-batman-2022.jpg
         ✅ Saved to database

  ...

  ✅ Done — 12 items scanned, 11 new, 1 updated
```

<p align="center">
  <img src="docs/screenshots/demo-scan.gif" alt="movie scan demo" width="700">
  <br><em>↑ Replace with actual recording</em>
</p>

---

### 📋 Browsing Your Library

<!-- Replace with actual GIF: docs/screenshots/demo-ls.gif -->

```
$ movie ls

🎬 Library — Page 1 of 3 (20 per page)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  #   Title                          Year   Type    Rating
  ─── ────────────────────────────── ────── ─────── ──────
  1   Scream                         2022   🎬      ★ 6.8
  2   The Batman                     2022   🎬      ★ 7.7
  3   Everything Everywhere All...   2022   🎬      ★ 7.8
  4   Breaking Bad                   2008   📺      ★ 8.9
  5   Severance                      2022   📺      ★ 8.4
  ...

  [N]ext  [P]rev  [1-9] Detail  [Q]uit
```

<p align="center">
  <img src="docs/screenshots/demo-ls.gif" alt="movie ls demo" width="700">
  <br><em>↑ Replace with actual recording</em>
</p>

---

### 🎯 Getting Suggestions

<!-- Replace with actual GIF: docs/screenshots/demo-suggest.gif -->

```
$ movie suggest 5

🎯 Movie Suggest
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  Select category:
  1. 🎬 Movie
  2. 📺 TV
  3. 🎲 Random

  Choose: 1

  📽️  Recommendations based on your library:

  #   Title                          Year   Rating   Genre
  ─── ────────────────────────────── ────── ──────── ────────────────
  1   Nope                           2022   ★ 6.8    Horror, Sci-Fi
  2   X                              2022   ★ 6.6    Horror, Mystery
  3   Pearl                          2022   ★ 7.0    Drama, Horror
  4   Bodies Bodies Bodies            2022   ★ 6.5    Comedy, Horror
  5   Barbarian                      2022   ★ 7.0    Horror, Thriller

  🔥 Trending This Week:
  1   Oppenheimer                    2023   ★ 8.1    Drama, History
  2   Killers of the Flower Moon     2023   ★ 7.5    Crime, Drama
  3   Poor Things                    2023   ★ 7.9    Comedy, Drama
```

<p align="center">
  <img src="docs/screenshots/demo-suggest.gif" alt="movie suggest demo" width="700">
  <br><em>↑ Replace with actual recording</em>
</p>

> **📹 Recording your own demos:**
> Use [VHS](https://github.com/charmbracelet/vhs) or [asciinema](https://asciinema.org/) to record terminal sessions as GIFs.
> ```bash
> # VHS (recommended — deterministic, scriptable)
> vhs docs/screenshots/scan.tape
>
> # asciinema + agg (manual recording)
> asciinema rec demo.cast
> agg demo.cast docs/screenshots/demo-scan.gif
> ```

---

<div align="center">

## Installation

</div>

### One-Liner Install (recommended)

Two flavours — pick based on whether you want auto-updates or a frozen version.

#### Latest release (auto-tracks newest)

**Windows (PowerShell)**

```powershell
irm https://raw.githubusercontent.com/alimtvnetwork/movie-cli-v6/main/get.ps1 | iex
```

**Linux / macOS (Bash)**

```bash
curl -fsSL https://raw.githubusercontent.com/alimtvnetwork/movie-cli-v6/main/get.sh | bash
```

`get.{ps1,sh}` first checks `releases/latest/download/install.{ps1,sh}`. If a release is published it installs the pre-built binary; otherwise it falls back to a source-build from `main`, prints exactly which path it took, and tells the maintainer how to publish a release so future installs skip the build step.

#### Pinned to a specific release

**Windows (PowerShell)**

```powershell
irm https://github.com/alimtvnetwork/movie-cli-v6/releases/download/v2.130.0/install.ps1 | iex
```

**Linux / macOS (Bash)**

```bash
curl -fsSL https://github.com/alimtvnetwork/movie-cli-v6/releases/download/v2.130.0/install.sh | bash
```

The script attached to each release has the version baked in (`PINNED_VERSION="v2.130.0"`) and will install **exactly** that tag — it never falls back to "latest" and never delegates to the bootstrap scripts. Replace `v2.130.0` with any [published release](https://github.com/alimtvnetwork/movie-cli-v6/releases).

> **When to use which**
> - **Latest** — personal machines, demos, "just give me the newest one"
> - **Pinned** — CI pipelines, Dockerfiles, onboarding docs, reproducing a bug on a specific version, controlled rollbacks
>
> Both URLs point at installer assets attached to the GitHub Release. The repo-root `install.ps1` and `install.sh` are unrelated source bootstrap scripts for building locally.

### Installer Options

**Windows (PowerShell):**

| Flag | Description | Example |
|---|---|---|
| `-InstallDir` | Custom install directory | `-InstallDir C:\tools\movie` |
| `-Arch` | Force architecture (`amd64`, `arm64`) | `-Arch arm64` |
| `-NoPath` | Skip adding to user PATH | `-NoPath` |

**Linux / macOS (Bash):**

| Flag | Description | Example |
|---|---|---|
| `--dir` | Custom install directory | `--dir ~/bin` |
| `--arch` | Force architecture (`amd64`, `arm64`) | `--arch arm64` |
| `--no-path` | Skip adding to PATH | `--no-path` |

### Clone & Build (Development)

**Prerequisites:**

| Requirement | Minimum | Check |
|---|---|---|
| **Go** | 1.22+ | `go version` |
| **Git** | 2.x | `git --version` |
| **PowerShell** | 5.1+ (Win) / 7+ (Unix) | `$PSVersionTable.PSVersion` |

```bash
git clone https://github.com/alimtvnetwork/movie-cli-v6.git
cd movie-cli-v6
pwsh run.ps1
```

**Using the bootstrap installer:**

```powershell
pwsh install.ps1                      # Fresh install (clone + build + deploy)
pwsh install.ps1 -DeployPath ~/bin    # Custom deploy path
```

### Verify

```bash
movie version
# v1.0.0 (commit: abc1234, built: 2026-04-09)
#   Go:   go1.22.0
#   OS:   linux/amd64
```

> **Tip:** If `movie` is not found, add the deploy path to your `PATH`.
> Default: `E:\bin-run` (Windows) or `/usr/local/bin` (Unix) for source builds.

---

<div align="center">

## What It Does

</div>

A portable CLI that manages your personal movie and TV show library entirely from the terminal. Every scan produces:

- **Database** — structured metadata in SQLite (WAL mode)
- **Thumbnails** — poster images downloaded from TMDb
- **JSON** — per-file metadata written to `./data/json/`
- **Clean filenames** — `Scream.2022.1080p.WEBRip.x264.mkv` → `Scream (2022).mkv`

All data lives in `./data/` at the project root.

---

<div align="center">

## Command Reference

</div>

Each section below shows a real-world example of what the command does.
Each thumbnail is a short looping walkthrough — hover or click to view the full-size still.

<details>
<summary>💡 <strong>PowerShell vs Bash quick reference</strong> — escaping paths & passing env vars in the examples below</summary>

The example commands are written in **Bash** (macOS / Linux / WSL / Git Bash). On **Windows PowerShell** a few things differ — use this table to translate any example before running it:

| Concept | Bash (macOS / Linux / WSL) | PowerShell (Windows) |
|---|---|---|
| Home folder | `~/Downloads` | `$HOME\Downloads` or `$env:USERPROFILE\Downloads` |
| Path with spaces | `"My Movies/Action Films"` (double quotes) | `'My Movies\Action Films'` (single quotes — no variable expansion) |
| Path separator | `/` | `\` (PowerShell also accepts `/`) |
| Escape a literal quote | `\"` inside `"..."` | `` ` " `` (backtick + quote) or use `'...'` |
| Read an env var | `$TMDB_KEY` | `$env:TMDB_KEY` |
| Set env var (one command) | `TMDB_KEY=abc movie scan ~/Downloads` | `$env:TMDB_KEY="abc"; movie scan $HOME\Downloads` |
| Set env var (whole session) | `export TMDB_KEY=abc` | `$env:TMDB_KEY = "abc"` |
| Set env var (persistent) | add `export ...` to `~/.bashrc` / `~/.zshrc` | `[Environment]::SetEnvironmentVariable("TMDB_KEY","abc","User")` |
| Command substitution | `cd $(movie cd Movies)` | `Set-Location (movie cd Movies)` |
| Line continuation | trailing `\` | trailing `` ` `` (backtick) |
| Comments | `# comment` | `# comment` (same) |

**Rule of thumb:** if an example uses `~`, `$VAR`, `\"`, or `$(...)`, swap it for the PowerShell equivalent above. Everything else (flags, subcommands, IDs) is identical across shells.

</details>

<details>
<summary>🔎 <strong>Check your env vars</strong> — confirm <code>TMDB_KEY</code> is set before running the examples</summary>

Run this once at the start of a session. It prints `set` / `MISSING` for each variable the CLI looks at, so you catch a missing TMDb token before a `movie scan` fails halfway through.

**Bash (macOS / Linux / WSL / Git Bash)**

```bash
for v in TMDB_KEY TMDB_API_KEY MOVIE_CONFIG MOVIE_DB_PATH; do
  if [ -n "${!v}" ]; then
    echo "✔ $v is set (${#v} chars: ${!v:0:4}…)"
  else
    echo "✘ $v is MISSING"
  fi
done
```

**PowerShell (Windows)**

```powershell
foreach ($v in 'TMDB_KEY','TMDB_API_KEY','MOVIE_CONFIG','MOVIE_DB_PATH') {
  $val = [Environment]::GetEnvironmentVariable($v)
  if ($val) {
    Write-Host "✔ $v is set ($($val.Length) chars: $($val.Substring(0,[Math]::Min(4,$val.Length)))…)"
  } else {
    Write-Host "✘ $v is MISSING"
  }
}
```

Expected output when everything is configured:

```text
✔ TMDB_KEY is set (32 chars: a1b2…)
✘ TMDB_API_KEY is MISSING        ← optional alias, safe to ignore if TMDB_KEY is set
✔ MOVIE_CONFIG is set (28 chars: /Use…)
✘ MOVIE_DB_PATH is MISSING       ← optional, falls back to the default DB location
```

Only `TMDB_KEY` is required for TMDb-backed commands (`scan`, `search`, `discover`, `suggest`). If it shows `MISSING`, set it with `export TMDB_KEY=...` (Bash) or `$env:TMDB_KEY = "..."` (PowerShell), or persist it via `movie config set tmdb_api_key YOUR_KEY`.

</details>

### Scanning & Library

<p align="center">
  <a href="assets/screenshots/cmd-scan-library.svg">
    <img src="assets/screenshots/cmd-scan-library.gif" alt="Animated walkthrough of movie scan: matching files against TMDb and reporting matches" width="780">
  </a>
  <br>
  <em>📸 <code>movie scan</code> walks a folder, cleans messy release names, and matches each file against TMDb.</em>
</p>

**▶ Try the example from the screenshot** — replace `~/Downloads` with any folder containing video files:

```bash
# 1. Reproduce the walkthrough above
movie scan ~/Downloads               # ← swap for your own scan folder

# 2. Re-run for any unmatched titles after the first pass
movie rescan

# 3. Confirm what landed in the library
movie ls
```

> **Path placeholders:** `~/Downloads` = macOS/Linux home folder. On Windows use `C:\Users\<you>\Downloads` or `$env:USERPROFILE\Downloads` in PowerShell.

<details>
<summary>🪟 <strong>PowerShell version</strong> (copy-paste on Windows)</summary>

```powershell
# 1. Reproduce the walkthrough above
movie scan "$env:USERPROFILE\Downloads"   # ← swap for your own scan folder

# 2. Re-run for any unmatched titles after the first pass
movie rescan

# 3. Confirm what landed in the library
movie ls
```
</details>

<details>
<summary>✅ <strong>Expected output</strong> (sample — yours will list your own files)</summary>

```text
Scanning ~/Downloads ... found 12 video files
  ✔ Inception.2010.1080p.mkv          → Inception (2010)            ★ 8.4
  ✔ The.Batman.2022.WEB.mp4           → The Batman (2022)           ★ 7.8
  ✔ Dune.Part.Two.2024.2160p.mkv      → Dune: Part Two (2024)       ★ 8.3
  ⚠ random_clip.mp4                   → no TMDb match (run `movie rescan` later)
Saved 11 entries to library. Run `movie ls` to browse.
```
</details>

| Command | Description |
|---|---|
| `movie scan [folder]` | Scan folder → DB + TMDb metadata |
| `movie rescan` | Re-fetch TMDb metadata for entries with missing data |
| `movie ls` | Paginated interactive library browser |
| `movie search <name>` | Live TMDb search → save to DB |
| `movie info <id\|title>` | Detail view (local DB → TMDb fallback) |

```bash
movie scan ~/Downloads            # scan folder, fetch metadata + posters
movie rescan                      # re-fetch missing genres/ratings from TMDb
movie ls                          # browse library with pagination
movie search "Inception"          # search TMDb and save result
movie info 1                      # show details for media ID 1
movie info "The Batman"           # search by title
```

---

### File Management

<p align="center">
  <a href="assets/screenshots/cmd-file-management.svg">
    <img src="assets/screenshots/cmd-file-management.gif" alt="Animated walkthrough of movie move showing planned destinations and a batch confirmation" width="780">
  </a>
  <br>
  <em>📸 <code>movie move</code> previews the destination for every file before touching the filesystem — fully reversible with <code>movie undo</code>.</em>
</p>

**▶ Try the example from the screenshot** — preview destinations, accept with `a`, then undo if needed:

```bash
# 1. Interactive preview (the walkthrough's "Select [a]ll, [n]one, or numbers" prompt)
movie move ~/Downloads               # ← swap for your own source folder

# 2. Or batch-route everything by type (Movies/ vs TV/)
movie move --all ~/Downloads

# 3. Changed your mind? Reverse the entire batch
movie undo
```

> **Path placeholders:** `~/Downloads` = macOS/Linux. Windows: `C:\Users\<you>\Downloads` or `$env:USERPROFILE\Downloads`.

<details>
<summary>🪟 <strong>PowerShell version</strong> (copy-paste on Windows)</summary>

```powershell
# 1. Interactive preview (the walkthrough's "Select [a]ll, [n]one, or numbers" prompt)
movie move "$env:USERPROFILE\Downloads"   # ← swap for your own source folder

# 2. Or batch-route everything by type (Movies\ vs TV\)
movie move --all "$env:USERPROFILE\Downloads"

# 3. Changed your mind? Reverse the entire batch
movie undo
```
</details>

<details>
<summary>✅ <strong>Expected output</strong> (sample preview before confirmation)</summary>

```text
Planned moves (3):
  [1] Inception.2010.1080p.mkv      → Movies/Inception (2010)/Inception.2010.1080p.mkv
  [2] The.Batman.2022.WEB.mp4       → Movies/The Batman (2022)/The Batman.2022.mp4
  [3] Breaking.Bad.S01E01.mkv       → TV/Breaking Bad/Season 01/S01E01.mkv
Select [a]ll, [n]one, or numbers (e.g. 1,3): a
✔ Moved 3 files. Undo with `movie undo` (batch id 87).
```
</details>

| Command | Description |
|---|---|
| `movie move [directory]` | Browse, select, move with clean name |
| `movie move --all` | Batch move all files (auto-route by type) |
| `movie rename` | Batch rename to clean format |
| `movie popout [directory]` | Extract video files from subfolders to root |
| `movie play <id>` | Open with default video player |
| `movie cd [folder-name]` | Print path of a scanned folder for quick nav |

```bash
movie move ~/Downloads            # interactive single-file move
movie move --all ~/Downloads      # batch move all files
movie rename                      # clean all filenames
movie popout ~/Downloads          # flatten nested subfolders
movie play 1                      # play with system player
cd $(movie cd Movies)             # navigate to scanned folder
```

---

### History & Undo

<p align="center">
  <a href="assets/screenshots/cmd-history-undo.svg">
    <img src="assets/screenshots/cmd-history-undo.gif" alt="Animated walkthrough of movie undo --list followed by movie undo --id 42 reverting a batch of moves" width="780">
  </a>
  <br>
  <em>📸 Every move, rename, scan, and delete is tracked. <code>movie undo --list</code> shows what can be reversed; <code>movie redo</code> re-applies it.</em>
</p>

**▶ Try the example from the screenshot** — list operations, undo a specific batch by ID, then redo it:

```bash
# 1. List recent operations (the walkthrough's "ID  When  Action  Target" table)
movie undo --list

# 2. Revert the batch you saw — replace 42 with the ID from your own list
movie undo --id 42                   # ← swap 42 for the ID you want to revert

# 3. Re-apply if you undid by mistake
movie redo
```

> **ID placeholder:** `42` is a sample undo ID. Run `movie undo --list` to see your own IDs.

<details>
<summary>🪟 <strong>PowerShell version</strong> (copy-paste on Windows)</summary>

```powershell
# 1. List recent operations (the walkthrough's "ID  When  Action  Target" table)
movie undo --list

# 2. Revert the batch you saw — replace 42 with the ID from your own list
movie undo --id 42                        # ← swap 42 for the ID you want to revert

# 3. Re-apply if you undid by mistake
movie redo
```
</details>

<details>
<summary>✅ <strong>Expected output</strong> (sample — IDs and timestamps will differ)</summary>

```text
ID   When              Action   Target
──   ────────────────  ───────  ─────────────────────────────────────────────
42   2025-04-20 14:02  move     3 files → Movies/
41   2025-04-20 13:55  rename   7 files cleaned
40   2025-04-20 12:10  scan     12 entries added

$ movie undo --id 42
✔ Reverted batch 42 — 3 files restored to original locations.
```
</details>

| Command | Description |
|---|---|
| `movie undo` | Revert last move/rename/delete/scan operation |
| `movie undo --list` | Show recent undoable actions |
| `movie undo --batch` | Undo the entire last batch (e.g. a full scan) |
| `movie undo --id <id>` | Undo a specific action by ID |
| `movie redo` | Re-apply the last undone operation |
| `movie history` | Show history of all tracked operations |

```bash
movie undo                        # revert most recent operation
movie undo --list                 # see what can be undone
movie undo --batch                # undo entire last scan batch
movie undo --id 42                # undo specific action
movie redo                        # re-apply last undone operation
movie history                     # view full operation history
```

---

### Discovery & Organization

<p align="center">
  <a href="assets/screenshots/cmd-discovery.svg">
    <img src="assets/screenshots/cmd-discovery.gif" alt="Animated walkthrough of movie suggest showing personalized recommendations and trending titles" width="780">
  </a>
  <br>
  <em>📸 <code>movie suggest</code> reads your library tastes and surfaces both personalized picks and trending titles from TMDb.</em>
</p>

**▶ Try the example from the screenshot** — get 5 picks, browse a genre, then add one to your watchlist:

```bash
# 1. Reproduce the walkthrough's 5-item recommendation block
movie suggest 5                      # ← change the number for more/fewer picks

# 2. Drill into a specific genre
movie discover Sci-Fi                # ← swap for Action, Comedy, Horror, etc.

# 3. Bookmark something to watch later (use any ID from `movie ls`)
movie watch add 3                    # ← swap 3 for your chosen media ID
```

> **Number / genre / ID placeholders:** `5` = pick count; `Sci-Fi` = any genre; `3` = media ID from your `movie ls`.

<details>
<summary>🪟 <strong>PowerShell version</strong> (copy-paste on Windows)</summary>

```powershell
# 1. Reproduce the walkthrough's 5-item recommendation block
movie suggest 5                           # ← change the number for more/fewer picks

# 2. Drill into a specific genre (quote names containing a hyphen to be safe)
movie discover "Sci-Fi"                   # ← swap for Action, Comedy, Horror, etc.

# 3. Bookmark something to watch later (use any ID from `movie ls`)
movie watch add 3                         # ← swap 3 for your chosen media ID
```
</details>

<details>
<summary>✅ <strong>Expected output</strong> (sample — picks vary based on your library)</summary>

```text
Top 5 picks for you (based on your top genres: Sci-Fi, Thriller):
  1. Arrival (2016)              ★ 7.9   Sci-Fi · Drama
  2. Edge of Tomorrow (2014)     ★ 7.9   Sci-Fi · Action
  3. Ex Machina (2014)           ★ 7.7   Sci-Fi · Thriller
  4. Annihilation (2018)         ★ 6.8   Sci-Fi · Horror
  5. Coherence (2013)            ★ 7.2   Sci-Fi · Mystery

✔ Added "Arrival (2016)" to watchlist (id 3).
```
</details>

| Command | Description |
|---|---|
| `movie suggest [N]` | Genre-based recommendations + trending |
| `movie discover [genre]` | Browse TMDb by genre (interactive picker or direct) |
| `movie tag add <id> <tag>` | Add a tag to a media item |
| `movie tag remove <id> <tag>` | Remove a tag |
| `movie tag list [id]` | List tags (per item or all) |
| `movie watch add <id>` | Add a library item to your watchlist |
| `movie watch done <id>` | Mark a title as watched |
| `movie watch undo <id>` | Revert a title back to to-watch |
| `movie watch rm <id>` | Remove a title from your watchlist |
| `movie watch ls` | List your watchlist |
| `movie watch export` | Export watchlist as JSON for backup |
| `movie watch import <file>` | Import watchlist from JSON |
| `movie stats` | Counts, storage, genre chart, avg ratings |
| `movie duplicates` | Detect duplicate media entries |

```bash
movie suggest 5                   # get 5 recommendations
movie discover                    # interactive genre picker
movie discover Action             # discover Action movies from TMDb
movie discover Comedy --type tv   # discover Comedy TV shows
movie discover Horror --page 2    # page 2 of Horror movies
movie tag add 1 favorite          # tag media #1 as favorite
movie tag list                    # list all tags
movie watch add 3                 # add media #3 to watchlist
movie watch done 3                # mark as watched
movie watch ls                    # view watchlist
movie stats                       # library statistics
movie duplicates                  # find duplicate entries
```

---

### Maintenance & Debugging

<p align="center">
  <a href="assets/screenshots/cmd-maintenance.svg">
    <img src="assets/screenshots/cmd-maintenance.gif" alt="Animated walkthrough of movie stats showing library counts, total size, and a top-genres bar chart" width="780">
  </a>
  <br>
  <em>📸 <code>movie stats</code> renders an instant overview — counts, storage used, top genres, and average rating.</em>
</p>

**▶ Try the example from the screenshot** — view stats, then prune any stale entries it surfaces:

```bash
# 1. Reproduce the walkthrough's library overview + top-genres chart
movie stats

# 2. Dry-run a cleanup to see entries whose files no longer exist
movie cleanup

# 3. Actually remove them once you're happy with the dry-run output
movie cleanup --remove
```

> **No placeholders here** — `movie stats` and `movie cleanup` run as-is.

<details>
<summary>🪟 <strong>PowerShell version</strong> (copy-paste on Windows)</summary>

```powershell
# 1. Reproduce the walkthrough's library overview + top-genres chart
movie stats

# 2. Dry-run a cleanup to see entries whose files no longer exist
movie cleanup

# 3. Actually remove them once you're happy with the dry-run output
movie cleanup --remove
```
</details>

<details>
<summary>✅ <strong>Expected output</strong> (sample — numbers reflect your library)</summary>

```text
Library: 142 titles · 118 movies · 24 TV shows · 1.7 TB
Top genres:  Drama ████████████ 38   Sci-Fi ████████ 26   Action ██████ 19
Average rating: ★ 7.4

$ movie cleanup
Stale entries (files missing on disk): 4
  - Old.Movie.2009.avi          (id 17)
  - Removed.Show.S02E03.mkv     (id 88)
Run `movie cleanup --remove` to delete these from the database.
```
</details>

| Command | Description |
|---|---|
| `movie cleanup` | Find stale entries where files no longer exist |
| `movie cleanup --remove` | Delete stale entries (not just preview) |
| `movie db` | Show resolved database path and status |
| `movie logs` | Display recent error logs from the database |
| `movie rest` | Start a local REST API server for the library |
| `movie rest --open` | Start server and open in browser |
| `movie export [-o path]` | Dump media table as JSON |

```bash
movie cleanup                     # dry run — show stale entries
movie cleanup --remove            # actually remove stale entries
movie db                          # check database location
movie logs                        # view recent error/warning logs
movie rest                        # start REST API on localhost
movie rest --open                 # start and open browser
movie export -o ~/library.json    # export full library as JSON
```

---

### Configuration & System

<p align="center">
  <a href="assets/screenshots/cmd-config-system.svg">
    <img src="assets/screenshots/cmd-config-system.gif" alt="Animated walkthrough of movie config showing config keys, setting tmdb_api_key, and movie version output" width="780">
  </a>
  <br>
  <em>📸 <code>movie config</code> shows every setting; <code>movie version</code> prints the exact build for bug reports.</em>
</p>

**▶ Try the example from the screenshot** — inspect config, set the TMDb key, then verify the build:

```bash
# 1. Reproduce the walkthrough's "Current configuration" block
movie config

# 2. Set your own TMDb API key (replace YOUR_KEY with the real value)
movie config set tmdb_api_key YOUR_KEY        # ← swap YOUR_KEY for your TMDb token

# 3. Confirm exactly which build is running (use this in bug reports)
movie version
```

> **Key placeholder:** `YOUR_KEY` = your TMDb API token from https://www.themoviedb.org/settings/api.

<details>
<summary>🪟 <strong>PowerShell version</strong> (copy-paste on Windows)</summary>

```powershell
# 1. Reproduce the walkthrough's "Current configuration" block
movie config

# 2. Set your own TMDb API key (replace YOUR_KEY with the real value)
#    Tip: store it in an env var first so it doesn't end up in shell history:
#       $env:TMDB_KEY = "your-real-token"
movie config set tmdb_api_key $env:TMDB_KEY   # ← or pass the literal token in quotes

# 3. Confirm exactly which build is running (use this in bug reports)
movie version
```
</details>

<details>
<summary>✅ <strong>Expected output</strong> (sample — your build info will differ)</summary>

```text
Current configuration:
  tmdb_api_key   ********************abcd   (set)
  library_root   ~/Media
  player         vlc
  log_level      info

$ movie config set tmdb_api_key YOUR_KEY
✔ Saved tmdb_api_key.

$ movie version
mahin v2.178.0   commit a1b2c3d   built 2025-04-26   go1.22.2 darwin/arm64
```
</details>

| Command | Description |
|---|---|
| `movie config` | Show all configuration |
| `movie config set <key> <val>` | Set a config value |
| `movie version` | Version, commit, build date, Go, OS/arch |
| `movie update` | Pull latest, rebuild, and deploy (copy-and-handoff) |
| `movie update-cleanup` | Remove leftover temp binaries and `.bak` backups |
| `movie changelog [--latest]` | Show release notes |

```bash
movie config set movies_dir ~/Movies
movie config set tmdb_api_key YOUR_KEY
movie update                          # full self-update: pull → build → deploy
movie update-cleanup                  # remove temp update artifacts
movie changelog --latest
```

**Config keys:**

| Key | Default | Purpose |
|---|---|---|
| `movies_dir` | `~/Movies` | Movie file destination |
| `tv_dir` | `~/TVShows` | TV show destination |
| `archive_dir` | `~/Archive` | Archive destination |
| `scan_dir` | `~/Downloads` | Default scan source |
| `tmdb_api_key` | *(none)* | TMDb API key |
| `page_size` | `20` | Items per page in `ls` |

---

<div align="center">

## Troubleshooting

</div>

### Quick Diagnosis Flowchart

Not sure which error you're seeing? Follow this decision tree to find the right fix in seconds.

```
┌─────────────────────────────────────┐
│  What happened when you ran the     │
│  command?                           │
└──────────────┬──────────────────────┘
               │
    ┌──────────┴──────────┐
    │                     │
┌───▼────┐           ┌────▼────┐
│ Every  │           │ Some or │
│ file   │           │ all got │
│ shows  │           │ skipped │
│ "no    │           │ with an │
│ TMDb   │           │ error   │
│ match" │           │ code    │
└───┬────┘           └────┬────┘
    │                     │
    │         ┌───────────┼───────────┐
    │         │           │           │
    │    ┌────▼────┐ ┌────▼────┐ ┌────▼────┐
    │    │ 429 /   │ │ 401 /   │ │ timeout │
    │    │ "too     │ │ "unauth-│ │ / DNS   │
    │    │ many    │ │ orized" │ │ failure │
    │    │ requests"│ │         │ │         │
    │    └────┬────┘ └────┬────┘ └────┬────┘
    │         │           │           │
    │    ┌────▼────┐ ┌────▼────┐ ┌────▼────┐
    │    │ Wait &  │ │ Check   │ │ Check   │
    │    │ re-run  │ │ your    │ │ network │
    │    │ rescan  │ │ API key │ │ / proxy │
    │    │         │ │         │ │ settings│
    │    └────┬────┘ └────┬────┘ └────┬────┘
    │         │           │           │
    └─────────┴───────────┴───────────┘
              │
    ┌─────────▼─────────┐
    │  "database is      │
    │  locked" or        │
    │  SQLITE_BUSY       │
    └─────────┬─────────┘
              │
    ┌─────────▼─────────┐
    │  Kill any other    │
    │  movie process,    │
    │  then retry        │
    └────────────────────┘
```

**Map the symptom to the fix:**

| Symptom | Likely cause | Jump to fix |
|---|---|---|
| Every file shows `no TMDb match` | API key missing or wrong | [1. `tmdb_api_key not set`](#1-tmdb_api_key-not-set--tmdb-requests-are-skipped) |
| `429 too many requests` | Rate limit hit during large scan | [5. `TMDb 429 Too Many Requests`](#5-tmdb-429-too-many-requests--rate-limited) |
| `database is locked` / `SQLITE_BUSY` | Second `movie` process running | [8. `database is locked`](#8-database-is-locked--second-movie-process-running) |

The most common errors users hit, what each one means, and the exact command to fix it. Each entry links back to the matching walkthrough in the [Command Reference](#command-reference).

**Symptom:** `movie scan` runs but every file is reported as `! no TMDb match — saved as Unknown` (see the warning row in the [scan walkthrough](assets/screenshots/cmd-scan-library.gif)).

**Cause:** No TMDb API key configured. The scanner falls back to filename-only parsing.

**Fix:**

```bash
movie config set tmdb_api_key YOUR_KEY      # see assets/screenshots/cmd-config-system.gif
movie config                                 # confirm: tmdb_api_key = ********  (set)
movie rescan                                 # backfill metadata for previously-unmatched entries
```

If the key is set but matches still fail, see error #5 (rate limits).

---

### 2. `no TMDb match` for a known title

**Symptom:** A file you recognize ends up unmatched in the [scan walkthrough](assets/screenshots/cmd-scan-library.gif), tagged `⚠ no TMDb match`.

**Cause:** The release filename is too noisy for the cleaner (extra release-group tags, unusual separators, foreign titles).

**Fix:** Search and link manually.

```bash
movie search "The Matrix"           # live TMDb search
movie info "The Matrix"             # confirm the right title
movie rescan                        # re-resolve everything still missing metadata
```

If the title genuinely isn't in TMDb, the OMDb fallback kicks in automatically when `OMDB_API_KEY` is set (see error #6).

---

### 3. `move` refuses to run — destination directory missing

**Symptom:** `movie move` aborts before showing the planned destinations from the [file-management walkthrough](assets/screenshots/cmd-file-management.gif), printing `movies_dir does not exist` or `tv_dir does not exist`.

**Cause:** `movies_dir` / `tv_dir` point to a folder that hasn't been created yet.

**Fix:**

```bash
movie config                                 # check current paths
mkdir -p ~/Movies ~/TVShows                  # create the destinations
movie config set movies_dir ~/Movies         # or repoint to an existing folder
movie config set tv_dir ~/TVShows
```

---

### 4. Wrong files moved — need to roll back

**Symptom:** A `movie move --all` or `movie rename` batch put files in unexpected places.

**Fix:** Every operation is tracked. Use the flow shown in the [history & undo walkthrough](assets/screenshots/cmd-history-undo.gif):

```bash
movie undo --list                # find the batch ID (e.g. 42)
movie undo --id 42               # revert exactly that batch
# changed your mind?
movie redo                       # re-apply the last undone operation
```

`movie undo` always works in reverse chronological order — there is no "permanent" move.

---

### 5. `TMDb 429 Too Many Requests` — rate limited

**Symptom:** `movie scan` or `movie suggest` (see the [discovery walkthrough](assets/screenshots/cmd-discovery.gif)) prints `tmdb: 429 too many requests` and skips entries.

**Cause:** TMDb caps free keys at ~50 requests / second. Large scans can briefly exceed it.

**Fix:** The scanner backs off automatically; just re-run the resolver after a short pause:

```bash
sleep 5 && movie rescan          # backfill anything skipped
movie logs                       # inspect any retained warnings
```

---

### 6. `OMDB_API_KEY not set` — fallback tier silently disabled

**Symptom:** Some titles still show as `Unknown` even after `movie rescan`, and `movie logs` shows `omdb: tier skipped (no key)`.

**Cause:** OMDb is the secondary provider used when TMDb has no result. It's opt-in and reads only from the environment — never the config file or repo.

**Fix:**

```bash
export OMDB_API_KEY=your_omdb_key            # add to your shell profile to persist
movie rescan
movie logs                                   # confirm the omdb-skip warnings are gone
```

If you also see `omdb: 401 unauthorized`, the key is wrong — generate a new one at omdbapi.com.

---

### 7. Stale entries — files were moved/deleted outside the CLI

**Symptom:** `movie ls` shows entries whose files no longer exist on disk. `movie stats` (see the [maintenance walkthrough](assets/screenshots/cmd-maintenance.gif)) over-reports `Total size`.

**Fix:**

```bash
movie cleanup                    # dry-run: list stale entries
movie cleanup --remove           # actually delete them from the DB
movie duplicates                 # also surface accidental dupes after a cleanup
```

---

### 8. `database is locked` — second `movie` process running

**Symptom:** Any command exits with `database is locked` or `SQLITE_BUSY`.

**Cause:** SQLite WAL allows many readers but only one writer at a time. A long-running `movie rest` server or a hung `movie scan` can hold the write lock.

**Fix:**

```bash
movie db                         # confirms the path of the locked DB
# stop any running 'movie rest' / 'movie scan'
ps -ef | grep -i movie           # find lingering processes
kill <pid>
```

If the lock persists after killing all processes, delete `data/movie.db-wal` and `data/movie.db-shm` (the live DB file is safe to keep).

---

### 9. `command not found: movie` after `movie update`

**Symptom:** Self-update appears to succeed but the next invocation prints `command not found`.

**Cause:** The new binary was deployed to a directory not on `$PATH`, or shell hash cache is stale.

**Fix:**

```bash
movie update-cleanup             # remove any half-installed temp binaries
hash -r                          # bash/zsh: clear the command cache
which movie                      # verify the resolved path
movie version                    # confirm the new build (see assets/screenshots/cmd-config-system.gif)
```

On Windows, restart the terminal so the updated `PATH` is picked up.

---

### Still stuck?

1. Run `movie version` and include the output in any bug report — it pins down the exact commit and build date.
2. Run `movie logs` — the most recent error rows usually point straight at the failing layer (TMDb / DB / filesystem).
3. Open an issue with the `version` line, the failing command, and the relevant `logs` excerpt.

---

<div align="center">

## Command Tree

</div>

```
movie
├── hello                         # Greeting with version
├── version                       # Version, commit, build date, Go, OS/arch
├── changelog [--latest]          # Show changelog (full or latest version)
├── update                        # Pull → rebuild → deploy (copy-and-handoff)
├── update-cleanup                # Remove temp update artifacts
├── config [get|set] [key]        # View/set configuration
├── scan [folder]                 # Scan folder → DB + TMDb metadata
├── rescan                        # Re-fetch missing TMDb metadata
├── ls                            # Paginated library list (file-backed only)
├── search <name>                 # Live TMDb search → save to DB
├── info <id|title>               # Detail view (local DB → TMDb fallback)
├── suggest [N]                   # Recommendations + trending
├── discover [genre]              # Browse TMDb by genre
├── move [directory]              # Browse, select, move with clean name
├── rename                        # Batch rename to clean format
├── popout [directory]            # Extract videos from subfolders
├── undo [--list|--batch|--id]    # Revert operations (move/delete/scan)
├── redo                          # Re-apply last undone operation
├── history                       # Show all tracked operations
├── play <id>                     # Open with default video player
├── stats                         # Counts, storage, genre chart, avg ratings
├── duplicates                    # Detect duplicate media entries
├── cleanup [--remove]            # Find/remove stale entries
├── tag [add|remove|list]         # Manage user-defined tags
├── watch [add|done|undo|rm|ls|export|import]  # Manage watchlist + sync
├── cd [folder-name]              # Print scanned folder path
├── export [-o path]              # Dump media table as JSON
├── db                            # Show database path and status
├── logs                          # View error/warning logs
└── rest [--open]                 # Start local REST API server
```

---

<div align="center">

## Build & Deploy

</div>

### Makefile Targets

| Target | Description |
|---|---|
| `make build` | Compile for current platform |
| `make build-all` | Cross-compile all 6 targets into `dist/` |
| `make build-windows` | Windows amd64 (with embedded icon) |
| `make build-windows-arm` | Windows arm64 (with embedded icon) |
| `make build-mac-arm` | macOS ARM64 |
| `make build-mac-intel` | macOS amd64 |
| `make build-linux` | Linux amd64 |
| `make build-linux-arm` | Linux arm64 |
| `make install` | Build + copy to `/usr/local/bin` |

### Build via run.ps1

```powershell
.\run.ps1                           # Full pipeline: pull, build, deploy
.\run.ps1 -NoPull                   # Skip git pull
.\run.ps1 -NoPull -NoDeploy        # Build only
.\run.ps1 -R movie scan D:\movies  # Build + run scan
.\run.ps1 -t                       # Run all unit tests
.\run.ps1 -ForcePull               # CI mode: discard changes + pull
```

| Flag | Description |
|---|---|
| `-NoPull` | Skip `git pull` |
| `-NoDeploy` | Skip deploy step |
| `-R` | Run movie after build (trailing args forwarded) |
| `-t` | Run all unit tests |
| `-ForcePull` | CI mode: discard changes + pull |

See [spec/03-general/04-run-guide.md](spec/03-general/04-run-guide.md) for the full usage guide.

---

<div align="center">

## Release Workflow

</div>

Releases are fully automated via GitHub Actions. Pushing to a `release/**` branch or a `v*` tag triggers:

1. **Cross-compilation** — 6 binaries (Windows/Linux/macOS × amd64/arm64)
2. **Packaging** — `.zip` (Windows) and `.tar.gz` (Unix)
3. **SHA256 checksums** — `checksums.txt` with all artifact hashes
4. **Install scripts** — version-pinned `install.ps1` and `install.sh`
5. **GitHub Release** — formatted page with changelog, checksums, and install instructions

### Creating a Release

```bash
# Option A: Push a release branch
git checkout -b release/v1.3.0
git push origin release/v1.3.0

# Option B: Tag directly
git tag v1.3.0
git push origin v1.3.0
```

Both trigger the same pipeline. Version is resolved from the ref name.

> **CI Pipeline:** Pushing a `release/*` branch or `v*` tag triggers GitHub Actions to cross-compile 6 targets, generate checksums, and create a GitHub release with changelog and install instructions.

See [spec/12-ci-cd-pipeline/02-release-pipeline.md](spec/12-ci-cd-pipeline/02-release-pipeline.md) for the full pipeline spec.

---

<div align="center">

## Project Structure

</div>

```
movie-cli-v6/
├── main.go                        # Entry point
├── cmd/                           # Cobra commands (one file per command)
│   ├── root.go                    # Root command, registers subcommands
│   ├── movie_config.go            # config get/set
│   ├── movie_scan.go              # scan folder
│   ├── movie_rescan.go            # re-fetch missing metadata
│   ├── movie_ls.go                # paginated list
│   ├── movie_search.go            # TMDb search
│   ├── movie_info.go              # detail view + shared fetch helpers
│   ├── movie_suggest.go           # recommendations
│   ├── movie_move.go              # interactive move
│   ├── movie_rename.go            # batch rename
│   ├── movie_popout.go            # extract from subfolders
│   ├── movie_undo.go              # undo operations
│   ├── movie_redo.go              # redo undone operations
│   ├── movie_history.go           # operation history
│   ├── movie_play.go              # play with system player
│   ├── movie_stats.go             # library statistics
│   ├── movie_duplicates.go        # duplicate detection
│   ├── movie_cleanup.go           # stale entry cleanup
│   ├── movie_tag.go               # tag management
│   ├── movie_watch.go             # watchlist management
│   ├── movie_cd.go                # folder navigation helper
│   ├── movie_export.go            # JSON export
│   ├── movie_db.go                # database path/status
│   ├── movie_logs.go              # error log viewer
│   ├── movie_rest.go              # REST API server
│   └── movie_resolve.go           # shared ID/title resolver
├── cleaner/cleaner.go             # Filename cleaning + slug generation
├── tmdb/client.go                 # TMDb API client
├── db/                            # SQLite database layer
│   ├── db.go                      # Connection + migrations
│   ├── media.go                   # Media CRUD operations
│   ├── config.go                  # Config get/set
│   └── history.go                 # Move + scan history
├── errlog/                        # Centralized error/warning logging
│   └── errlog.go                  # File + DB logging with stack traces
├── updater/                       # Copy-and-handoff self-update
│   ├── run.go                     # Entry points: Run() + RunWorker()
│   ├── repo.go                    # Repo path resolution
│   ├── handoff.go                 # Binary copy + foreground launch
│   ├── script.go                  # PowerShell script generation
│   └── cleanup.go                 # Temp artifact removal
├── version/version.go             # Build-time version variables
├── .github/
│   └── workflows/
│       ├── ci.yml                 # Lint + test + vulncheck + cross-build
│       ├── release.yml            # Cross-compile + GitHub Release
│       └── vulncheck.yml          # Weekly vulnerability scan
├── run.ps1                        # PowerShell build + deploy pipeline
├── install.ps1                    # Bootstrap installer
├── CHANGELOG.md                   # Release notes
└── spec/                          # Detailed specifications
```

---

<div align="center">

## Data Storage

</div>

All data lives in `./data/`:

```
./data/
├── movie.db                  # SQLite database (WAL mode)
├── thumbnails/               # Downloaded poster images
└── json/
    ├── movie/                # Per-movie JSON metadata
    ├── tv/                   # Per-show JSON metadata
    └── history/              # Move operation logs (RFC3339)
```

---

<div align="center">

## Milestones

</div>

Project milestones are tracked in [`MILESTONES.md`](MILESTONES.md) at the repository root.

- **Location** — `MILESTONES.md` (repo root, version-controlled)
- **Timezone** — Malaysia time (UTC+8, `Asia/Kuala_Lumpur`)
- **Timestamp format** — `dd-MMM-YYYY hh:mm AM/PM` (e.g. `24-Apr-2026 03:33 PM`)
- **Entry format** — one bullet per line under the `## Log` heading:

  ```
  - <event> <dd-MMM-YYYY hh:mm AM/PM> — <short note>
  ```

Example entries:

```
- let's start now 24-Apr-2026 03:33 PM — milestone tracker initialized
- run 24-Apr-2026 07:21 PM — app run logged
```

New entries are appended to the end of the `## Log` section. Generate the timestamp with:

```bash
TZ='Asia/Kuala_Lumpur' date '+%d-%b-%Y %I:%M %p'
```

### Listing & filtering milestones

The `movie milestones` command reads `MILESTONES.md` and prints entries with
optional date / keyword filters:

```bash
movie milestones                              # show all entries
movie milestones --keyword scan               # case-insensitive substring
movie milestones --date 2026-04-24            # only entries on this day
movie milestones --since 2026-04-01           # entries on/after this day
movie milestones --since 2026-04-01 -k run -n 20
```

Flags: `--date YYYY-MM-DD`, `--since YYYY-MM-DD`, `-k/--keyword TEXT`,
`-n/--limit N` (0 = no cap).

### One-shot helper (append + version bump + commit)

The repo ships with a script that appends a new milestone, bumps the **patch**
version in `version/info.go`, and creates a single git commit covering both
changes:

**Linux / macOS**

```bash
scripts/log-milestone.sh                       # default: "- run <ts> — app run logged"
scripts/log-milestone.sh "kickoff complete"    # custom note
scripts/log-milestone.sh --event start "kickoff"
```

**Windows (PowerShell)**

```powershell
pwsh scripts/log-milestone.ps1
pwsh scripts/log-milestone.ps1 -Note "kickoff complete"
pwsh scripts/log-milestone.ps1 -Event start -Note "kickoff"
```

Wire it into the app (e.g. at the end of `movie` startup, or as a `make run`
target) to get a milestone + commit on every run. Commit message format:
`chore(milestone): <event> <timestamp> — <note> (<new-version>)`.

---

<div align="center">

## Dependencies

</div>

| Package | Purpose |
|---|---|
| [`github.com/spf13/cobra`](https://github.com/spf13/cobra) | CLI framework |
| [`modernc.org/sqlite`](https://pkg.go.dev/modernc.org/sqlite) | Pure-Go SQLite driver (no CGo) |

---

## 🤝 Contributing

Contributions are welcome! Here's how to get started:

1. **Fork** the repository
2. **Create a branch** for your feature or fix:
   ```bash
   git checkout -b feature/my-feature
   ```
3. **Follow the coding guidelines** in [`spec/01-coding-guidelines/`](spec/01-coding-guidelines/)
4. **Keep files small** — one file per command, max ~200 lines
5. **Run tests** before submitting:
   ```bash
   make tidy
   go test ./... -v
   ```
6. **Open a Pull Request** against `main` with a clear description

### Development Setup

```bash
git clone https://github.com/alimtvnetwork/movie-cli-v6.git
cd movie-cli-v6
make tidy
make build
```

See the [Install Guide](spec/03-general/01-install-guide.md) for full setup instructions.

### Code Style

- Go standard formatting (`gofmt`)
- Descriptive variable names, no abbreviations
- Error messages start lowercase, no trailing punctuation
- All new code must pass `go vet` and `golangci-lint`

---

## 📜 Code of Conduct

We are committed to providing a welcoming and inclusive experience for everyone.

**Our Standards:**

- Be respectful, constructive, and empathetic
- Welcome newcomers and help them get started
- Accept constructive criticism gracefully
- Focus on what's best for the project and community

**Unacceptable Behavior:**

- Harassment, trolling, or personal attacks
- Publishing others' private information
- Any conduct that would be considered inappropriate in a professional setting

**Enforcement:** Project maintainers may remove, edit, or reject contributions that violate this code. Repeated violations may result in a temporary or permanent ban.

This Code of Conduct is adapted from the [Contributor Covenant v2.1](https://www.contributor-covenant.org/version/2/1/code_of_conduct/).

---

<div align="center">

## Author

### [Md. Alim Ul Karim](https://www.google.com/search?q=alim+ul+karim)

**[Creator & Lead Architect](https://alimkarim.com)** | [Chief Software Engineer](https://www.google.com/search?q=alim+ul+karim), [Riseup Asia LLC](https://riseup-asia.com)

</div>

A system architect with **20+ years** of professional software engineering experience across enterprise, fintech, and distributed systems. His technology stack spans **.NET/C# (18+ years)**, **JavaScript (10+ years)**, **TypeScript (6+ years)**, and **Golang (4+ years)**.

Recognized as a **top 1% talent at Crossover** and one of the top software architects globally. He is also the **Chief Software Engineer of [Riseup Asia LLC](https://riseup-asia.com/)** and maintains an active presence on **[Stack Overflow](https://stackoverflow.com/users/361646/alim-ul-karim)** (2,452+ reputation, member since 2010) and **LinkedIn** (12,500+ followers).

|  |  |
|---|---|
| **Website** | [alimkarim.com](https://alimkarim.com/) · [my.alimkarim.com](https://my.alimkarim.com/) |
| **LinkedIn** | [linkedin.com/in/alimkarim](https://linkedin.com/in/alimkarim) |
| **Stack Overflow** | [stackoverflow.com/users/361646/alim-ul-karim](https://stackoverflow.com/users/361646/alim-ul-karim) |
| **Google** | [Alim Ul Karim](https://www.google.com/search?q=Alim+Ul+Karim) |
| **Role** | Chief Software Engineer, [Riseup Asia LLC](https://riseup-asia.com) |

### Riseup Asia LLC

[Top Leading Software Company in WY (2026)](https://riseup-asia.com)

| | |
|---|---|
| **Website** | [riseup-asia.com](https://riseup-asia.com) |
| **Facebook** | [riseupasia.talent](https://www.facebook.com/riseupasia.talent/) |
| **LinkedIn** | [Riseup Asia](https://www.linkedin.com/company/105304484/) |
| **YouTube** | [@riseup-asia](https://www.youtube.com/@riseup-asia) |

---

<div align="center">

## License

</div>

Released under the [MIT License](LICENSE) — free for personal and commercial use, with no warranty.

<div align="center">

_Built with ❤️ by [Md. Alim Ul Karim](https://alimkarim.com) · [Riseup Asia LLC](https://riseup-asia.com)_

</div>
