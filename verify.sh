#!/usr/bin/env bash
# verify.sh — Post-install verification for the movie CLI.
#
# Checks prerequisites (Go, Git) and confirms the installed `movie` binary
# is reachable, executable, and reports a sane version string.
#
# Usage:
#   bash verify.sh
#   bash verify.sh --binary movie
#   bash verify.sh --dir ~/.local/bin
#
# Exit codes:
#   0  all checks passed
#   1  one or more checks failed
#   2  bad usage

set -uo pipefail

BINARY_NAME="movie"
INSTALL_DIR=""
MIN_GO_MAJOR=1
MIN_GO_MINOR=22

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
GRAY='\033[0;90m'
NC='\033[0m'

PASS=0
FAIL=0
WARN=0

pass() { printf "  ${GREEN}[PASS]${NC} %s\n" "$1"; PASS=$((PASS+1)); }
fail() { printf "  ${RED}[FAIL]${NC} %s\n" "$1"; FAIL=$((FAIL+1)); }
warn() { printf "  ${YELLOW}[WARN]${NC} %s\n" "$1"; WARN=$((WARN+1)); }
info() { printf "  ${GRAY}%s${NC}\n" "$1"; }
hdr()  { printf "\n${CYAN}== %s ==${NC}\n" "$1"; }

usage() {
    cat <<EOF
Usage: bash verify.sh [--binary NAME] [--dir PATH]

  --binary NAME   Binary name to look for (default: movie)
  --dir PATH      Specific install directory to probe
  -h, --help      Show this help
EOF
}

# ── Parse args ─────────────────────────────────────────────────
while [ $# -gt 0 ]; do
    case "$1" in
        --binary) BINARY_NAME="${2:-}"; shift 2 ;;
        --dir)    INSTALL_DIR="${2:-}"; shift 2 ;;
        -h|--help) usage; exit 0 ;;
        *) echo "Unknown arg: $1"; usage; exit 2 ;;
    esac
done

printf "\n${CYAN}movie CLI — install verification${NC}\n"

# ── 1. Prerequisites ───────────────────────────────────────────
hdr "Prerequisites"

if command -v git >/dev/null 2>&1; then
    pass "git found ($(git --version | head -n1))"
else
    fail "git not found in PATH"
fi

if command -v go >/dev/null 2>&1; then
    GO_VER_RAW="$(go version | awk '{print $3}' | sed 's/^go//')"
    GO_MAJOR="$(echo "$GO_VER_RAW" | cut -d. -f1)"
    GO_MINOR="$(echo "$GO_VER_RAW" | cut -d. -f2)"
    if [ "$GO_MAJOR" -gt "$MIN_GO_MAJOR" ] 2>/dev/null || \
       { [ "$GO_MAJOR" -eq "$MIN_GO_MAJOR" ] && [ "$GO_MINOR" -ge "$MIN_GO_MINOR" ]; } 2>/dev/null; then
        pass "go $GO_VER_RAW (>= ${MIN_GO_MAJOR}.${MIN_GO_MINOR})"
    else
        warn "go $GO_VER_RAW is older than required ${MIN_GO_MAJOR}.${MIN_GO_MINOR}"
    fi
else
    warn "go not found (only required for building from source)"
fi

# ── 2. Locate binary ───────────────────────────────────────────
hdr "Binary"

BIN_PATH=""
if [ -n "$INSTALL_DIR" ]; then
    if [ -x "$INSTALL_DIR/$BINARY_NAME" ]; then
        BIN_PATH="$INSTALL_DIR/$BINARY_NAME"
        pass "found $BIN_PATH"
    else
        fail "no executable $BINARY_NAME in $INSTALL_DIR"
    fi
else
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        BIN_PATH="$(command -v "$BINARY_NAME")"
        pass "found on PATH at $BIN_PATH"
    else
        fail "$BINARY_NAME not found on PATH"
        info "try: export PATH=\"\$HOME/.local/bin:\$PATH\""
    fi
fi

# ── 3. Execution check ────────────────────────────────────────
hdr "Execution"

if [ -n "$BIN_PATH" ]; then
    if [ -x "$BIN_PATH" ]; then
        pass "binary is executable"
    else
        fail "binary exists but is not executable (chmod +x $BIN_PATH)"
    fi

    if VER_OUT="$("$BIN_PATH" version 2>&1)"; then
        if echo "$VER_OUT" | grep -Eq 'v[0-9]+\.[0-9]+\.[0-9]+'; then
            pass "version command works"
            info "$(echo "$VER_OUT" | head -n1)"
        else
            warn "version command ran but output looks unexpected"
            info "$(echo "$VER_OUT" | head -n1)"
        fi
    else
        fail "running '$BINARY_NAME version' failed"
        info "$(echo "$VER_OUT" | head -n3)"
    fi

    if "$BIN_PATH" --help >/dev/null 2>&1 || "$BIN_PATH" help >/dev/null 2>&1; then
        pass "help command responds"
    else
        warn "help command did not respond cleanly"
    fi
else
    info "skipping execution checks — binary not located"
fi

# ── Summary ───────────────────────────────────────────────────
hdr "Summary"
printf "  ${GREEN}%d passed${NC}  ${YELLOW}%d warnings${NC}  ${RED}%d failed${NC}\n" \
    "$PASS" "$WARN" "$FAIL"

if [ "$FAIL" -gt 0 ]; then
    printf "\n${RED}Verification failed.${NC} See messages above.\n\n"
    exit 1
fi
printf "\n${GREEN}All required checks passed — movie CLI is ready.${NC}\n\n"
exit 0
