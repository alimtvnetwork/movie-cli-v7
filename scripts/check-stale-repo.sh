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

for arg in "$@"; do
    case "$arg" in
        --apply) APPLY=1 ;;
        --yes|-y) ASSUME_YES=1 ;;
        -h|--help)
            sed -n '2,16p' "$0" | sed 's/^# \{0,1\}//'
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

# --- 3. Apply mode -------------------------------------------------------
if [ "$DIRTY" -eq 1 ] || [ "$UNTRACKED" -gt 0 ]; then
    warn "You have local modifications and/or untracked files."
    warn "These will be PERMANENTLY LOST by reset --hard + clean -fd."
    if ! confirm "Create a safety backup branch before proceeding?"; then
        say "Skipping backup."
    else
        BACKUP="backup/stale-repo-$(date +%Y%m%d-%H%M%S)"
        git branch "$BACKUP" >/dev/null 2>&1 \
            && ok "Created backup branch: $BACKUP"
        if [ "$DIRTY" -eq 1 ]; then
            git stash push --include-untracked -m "stale-repo backup $(date -u +%FT%TZ)" \
                >/dev/null 2>&1 && ok "Stashed working changes (see: git stash list)"
        fi
    fi
fi

confirm "Run: git fetch $REMOTE" || { err "Aborted."; exit 3; }
git fetch "$REMOTE" || { err "fetch failed"; exit 2; }

confirm "Run: git reset --hard $REMOTE/$BRANCH (destroys local commits on $CURRENT_BRANCH)" \
    || { err "Aborted."; exit 3; }
git reset --hard "$REMOTE/$BRANCH" || { err "reset failed"; exit 2; }

confirm "Run: git clean -fd (removes untracked files & dirs)" \
    || { warn "Skipped clean. Repo is reset but may still contain untracked files."; exit 0; }
git clean -fd || { err "clean failed"; exit 2; }

ok "Repo is now in sync with $REMOTE/$BRANCH at ${REMOTE_SHA:0:10}."
