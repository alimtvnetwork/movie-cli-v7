#!/usr/bin/env bash
# check-stale-repo.sh
#
# Safely detects whether the local clone is behind / diverged from origin/main
# and guides the user through the recommended remediation steps.
#
# This script is READ-ONLY by default. It will NEVER mutate your working tree
# unless you re-invoke it with --apply AND explicitly confirm each step.
#
# Exit codes:
#   0  repo is in sync (or user safely resolved it)
#   1  repo is stale / diverged (action required)
#   2  not a git repo / git unavailable / fetch failed
#   3  user aborted remediation

set -u

REMOTE="${REMOTE:-origin}"
BRANCH="${BRANCH:-main}"
APPLY=0
ASSUME_YES=0
VERBOSE=0

for arg in "$@"; do
    case "$arg" in
        --apply) APPLY=1 ;;
        --yes|-y) ASSUME_YES=1 ;;
        --verbose|-v) VERBOSE=1 ;;
        -h|--help)
            sed -n '2,16p' "$0" | sed 's/^# \{0,1\}//'
            cat <<'USAGE'

Flags:
  --apply       Interactively run remediation (fetch/reset/clean).
  --yes, -y     Skip confirmation prompts (use with --apply).
  --verbose, -v Print computed SHAs, commit messages, and status checks.
  -h, --help    Show this help.
USAGE
            exit 0
            ;;
        *)
            echo "Unknown argument: $arg" >&2
            exit 2
            ;;
    esac
done

say()  { printf '\033[1;36m[stale-repo]\033[0m %s\n' "$*"; }
warn() { printf '\033[1;33m[stale-repo]\033[0m %s\n' "$*" >&2; }
err()  { printf '\033[1;31m[stale-repo]\033[0m %s\n' "$*" >&2; }
ok()   { printf '\033[1;32m[stale-repo]\033[0m %s\n' "$*"; }
vsay() { [ "$VERBOSE" -eq 1 ] && printf '\033[2;37m[stale-repo:verbose]\033[0m %s\n' "$*" || true; }

confirm() {
    local prompt="$1"
    if [ "$ASSUME_YES" -eq 1 ]; then
        say "$prompt [auto-yes]"
        return 0
    fi
    printf '\033[1;35m[stale-repo]\033[0m %s [y/N]: ' "$prompt"
    read -r reply </dev/tty 2>/dev/null || reply=""
    case "$reply" in
        y|Y|yes|YES) return 0 ;;
        *) return 1 ;;
    esac
}

command -v git >/dev/null 2>&1 || { err "git not found in PATH"; exit 2; }

git rev-parse --is-inside-work-tree >/dev/null 2>&1 \
    || { err "Not inside a git working tree"; exit 2; }

REPO_ROOT="$(git rev-parse --show-toplevel)"
say "Repository: $REPO_ROOT"
say "Tracking:   $REMOTE/$BRANCH"

# --- 1. Fetch latest refs (read-only) ------------------------------------
say "Fetching latest refs from $REMOTE (read-only)..."
if ! git fetch --quiet "$REMOTE" "$BRANCH" 2>/dev/null; then
    err "git fetch $REMOTE $BRANCH failed. Check network / remote URL."
    exit 2
fi

# --- 2. Compare local HEAD to origin/BRANCH ------------------------------
LOCAL_SHA="$(git rev-parse HEAD)"
REMOTE_SHA="$(git rev-parse "$REMOTE/$BRANCH")"
CURRENT_BRANCH="$(git rev-parse --abbrev-ref HEAD)"

say "Current branch: $CURRENT_BRANCH"
say "Local  HEAD:    ${LOCAL_SHA:0:10}"
say "Remote HEAD:    ${REMOTE_SHA:0:10}"

if [ "$LOCAL_SHA" = "$REMOTE_SHA" ]; then
    ok "Repo is in sync with $REMOTE/$BRANCH. Nothing to do."
    exit 0
fi

BEHIND="$(git rev-list --count "$LOCAL_SHA".."$REMOTE_SHA" 2>/dev/null || echo "?")"
AHEAD="$(git rev-list --count "$REMOTE_SHA".."$LOCAL_SHA" 2>/dev/null || echo "?")"
DIRTY=0
git diff --quiet && git diff --cached --quiet || DIRTY=1
UNTRACKED="$(git ls-files --others --exclude-standard | wc -l | tr -d ' ')"

warn "Repo is OUT OF SYNC."
warn "  behind origin:    $BEHIND commit(s)"
warn "  ahead of origin:  $AHEAD commit(s)"
warn "  dirty worktree:   $([ $DIRTY -eq 1 ] && echo yes || echo no)"
warn "  untracked files:  $UNTRACKED"

cat <<EOF

Recommended remediation (DESTRUCTIVE — discards local changes):
    git fetch $REMOTE
    git reset --hard $REMOTE/$BRANCH
    git clean -fd

EOF

if [ "$APPLY" -eq 0 ]; then
    say "Run again with --apply to execute these steps interactively."
    say "Re-run with --apply --yes to skip confirmation prompts."
    exit 1
fi

# --- 3. Apply mode: plan backup refs (do not create yet) ----------------
BACKUP_BRANCH=""
BACKUP_BASE_SHA=""
STASH_MSG=""
WILL_BACKUP=0
WILL_STASH=0

if [ "$DIRTY" -eq 1 ] || [ "$UNTRACKED" -gt 0 ] || [ "$AHEAD" != "0" ]; then
    warn "You have local commits, modifications, and/or untracked files."
    warn "These will be PERMANENTLY LOST by reset --hard + clean -fd."
    if confirm "Create a safety backup branch + stash before proceeding?"; then
        WILL_BACKUP=1
        BACKUP_BRANCH="backup/stale-repo-$(date +%Y%m%d-%H%M%S)"
        BACKUP_BASE_SHA="$LOCAL_SHA"
        if [ "$DIRTY" -eq 1 ] || [ "$UNTRACKED" -gt 0 ]; then
            WILL_STASH=1
            STASH_MSG="stale-repo backup $(date -u +%FT%TZ)"
        fi
    else
        say "Skipping backup."
    fi
fi

# --- 3a. Final pre-flight summary ---------------------------------------
LATEST_LOCAL_MSG="$(git log -1 --pretty=format:'%h %s' HEAD 2>/dev/null)"
LATEST_REMOTE_MSG="$(git log -1 --pretty=format:'%h %s' "$REMOTE/$BRANCH" 2>/dev/null)"

cat <<EOF

────────────────────────────────────────────────────────────────────────
              FINAL CONFIRMATION — review before applying
────────────────────────────────────────────────────────────────────────
  Repository       : $REPO_ROOT
  Branch           : $CURRENT_BRANCH  →  $REMOTE/$BRANCH
  Local  HEAD      : ${LOCAL_SHA:0:10}   $LATEST_LOCAL_MSG
  Remote HEAD      : ${REMOTE_SHA:0:10}   $LATEST_REMOTE_MSG
  Behind remote    : $BEHIND commit(s)    [will be pulled in]
  Ahead of remote  : $AHEAD commit(s)     [will be DISCARDED]
  Dirty worktree   : $([ $DIRTY -eq 1 ] && echo "YES — uncommitted changes will be LOST" || echo "no")
  Untracked files  : $UNTRACKED  $([ $UNTRACKED -gt 0 ] && echo "[will be DELETED by clean -fd]")

  Backup plan:
      Backup branch  : $([ $WILL_BACKUP -eq 1 ] && echo "$BACKUP_BRANCH  →  ${BACKUP_BASE_SHA:0:10}" || echo "(none — no backup will be created)")
      Stash entry    : $([ $WILL_STASH  -eq 1 ] && echo "\"$STASH_MSG\"" || echo "(none)")
      Recover with   : $([ $WILL_BACKUP -eq 1 ] && echo "git checkout $BACKUP_BRANCH$([ $WILL_STASH -eq 1 ] && echo " && git stash pop")" || echo "(no recovery ref — changes will be unrecoverable)")

  Commands that will run (in order):
      $([ $WILL_BACKUP -eq 1 ] && echo "git branch $BACKUP_BRANCH $BACKUP_BASE_SHA")
      $([ $WILL_STASH  -eq 1 ] && echo "git stash push --include-untracked -m \"$STASH_MSG\"")
      git fetch $REMOTE
      git reset --hard $REMOTE/$BRANCH
      git clean -fd
────────────────────────────────────────────────────────────────────────

EOF

confirm "Proceed with the above remediation?" \
    || { err "Aborted by user. No changes made."; exit 3; }

# --- 3b. Create backup refs now (after confirmation) --------------------
if [ "$WILL_BACKUP" -eq 1 ]; then
    if git branch "$BACKUP_BRANCH" "$BACKUP_BASE_SHA" >/dev/null 2>&1; then
        ok "Created backup branch: $BACKUP_BRANCH @ ${BACKUP_BASE_SHA:0:10}"
    else
        err "Failed to create backup branch $BACKUP_BRANCH"; exit 2
    fi
fi
if [ "$WILL_STASH" -eq 1 ]; then
    if git stash push --include-untracked -m "$STASH_MSG" >/dev/null 2>&1; then
        STASH_REF="$(git stash list | grep -F "$STASH_MSG" | head -1 | cut -d: -f1)"
        ok "Stashed working changes as: ${STASH_REF:-stash@{0}}  (\"$STASH_MSG\")"
    else
        warn "Nothing to stash (worktree became clean)."
    fi
fi

confirm "Step 1/3 — Run: git fetch $REMOTE" || { err "Aborted."; exit 3; }
git fetch "$REMOTE" || { err "fetch failed"; exit 2; }

DESTRUCTIVE_PROMPT="Step 2/3 — DESTRUCTIVE: git reset --hard $REMOTE/$BRANCH (discards $AHEAD local commit(s) on $CURRENT_BRANCH"
if [ "$WILL_BACKUP" -eq 1 ]; then
    DESTRUCTIVE_PROMPT="$DESTRUCTIVE_PROMPT; recoverable via $BACKUP_BRANCH)"
else
    DESTRUCTIVE_PROMPT="$DESTRUCTIVE_PROMPT; NO BACKUP — UNRECOVERABLE)"
fi
confirm "$DESTRUCTIVE_PROMPT" || { err "Aborted."; exit 3; }
git reset --hard "$REMOTE/$BRANCH" || { err "reset failed"; exit 2; }

confirm "Step 3/3 — Run: git clean -fd (removes untracked files & dirs)" \
    || { warn "Skipped clean. Repo is reset but may still contain untracked files."; exit 0; }
git clean -fd || { err "clean failed"; exit 2; }

ok "Repo is now in sync with $REMOTE/$BRANCH at ${REMOTE_SHA:0:10}."
