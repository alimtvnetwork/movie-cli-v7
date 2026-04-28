# 🚀 Quickstart

Get the **movie** CLI running in under a minute. For full docs, flags, and
troubleshooting, see [README.md](README.md) and
[spec/03-general/01-install-guide.md](spec/03-general/01-install-guide.md).

---

## 🚀 One-liner install

<!-- INSTALL:BEGIN -->
<!-- INSTALL:END -->

---

## 🐧 Linux

```bash
curl -fsSL https://raw.githubusercontent.com/alimtvnetwork/movie-cli-v7/main/get.sh | bash
movie version
```

If `movie` is not found:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

---

## 🍎 macOS

```bash
curl -fsSL https://raw.githubusercontent.com/alimtvnetwork/movie-cli-v7/main/get.sh | bash
movie version
```

If `movie` is not found:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

---

## 🪟 Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/alimtvnetwork/movie-cli-v7/main/get.ps1 | iex
movie version
```

If `movie` is not found in the current shell:

```powershell
$env:PATH += ";$HOME\bin"
```

---

## ✅ Verify the install

```bash
# Linux / macOS
curl -fsSL https://raw.githubusercontent.com/alimtvnetwork/movie-cli-v7/main/verify.sh | bash
```

```powershell
# Windows
irm https://raw.githubusercontent.com/alimtvnetwork/movie-cli-v7/main/verify.ps1 | iex
```

Expected: `All required checks passed — movie CLI is ready.`

---

## 🎬 First commands

```bash
movie scan ~/Movies        # scan a folder, match against TMDb
movie ls                   # list your library
movie info "The Matrix"    # show details for one title
movie --help               # full command list
```
