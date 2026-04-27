#!/usr/bin/env bash
# pre-release.sh — Run the full pre-release checklist in order.
#
# Steps (halt on first failure):
#   1. go mod tidy           — dependency hygiene
#   2. audit-legacy-paths.sh --strict — block ACTIVE legacy module refs
#   3. go build ./...        — compiles every package
#   4. go test ./...         — full test suite
#
# Usage:
#   bash scripts/pre-release.sh
#   bash scripts/pre-release.sh --skip-tests   # everything except go test
#
# Exit codes:
#   0  all steps passed — safe to tag a release
#   1  a step failed     — see the printed banner for which one
#   2  bad arguments

set -uo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

SKIP_TESTS=0
for arg in "$@"; do
    case "$arg" in
        --skip-tests) SKIP_TESTS=1 ;;
        -h|--help) sed -n '2,16p' "$0"; exit 0 ;;
        *) echo "Unknown arg: $arg" >&2; exit 2 ;;
    esac
done

command -v go >/dev/null 2>&1 || { echo "go is required on PATH" >&2; exit 2; }

step() {
    local n="$1" name="$2"
    echo ""
    echo "════════════════════════════════════════════════════════════════"
    echo "  Step ${n}/4 — ${name}"
    echo "════════════════════════════════════════════════════════════════"
}

fail() {
    local n="$1" name="$2"
    echo ""
    echo "❌ Pre-release FAILED at step ${n}/4 — ${name}"
    echo "   Fix the issue above and re-run: bash scripts/pre-release.sh"
    exit 1
}

step 1 "go mod tidy"
go mod tidy || fail 1 "go mod tidy"

step 2 "legacy module-path auditor (strict)"
bash "${ROOT}/scripts/audit-legacy-paths.sh" --strict \
    || fail 2 "legacy module-path auditor"

step 3 "go build ./..."
go build ./... || fail 3 "go build"

if [ "$SKIP_TESTS" -eq 1 ]; then
    echo ""
    echo "Step 4/4 — go test ./...   (SKIPPED via --skip-tests)"
else
    step 4 "go test ./..."
    go test ./... || fail 4 "go test"
fi

echo ""
echo "════════════════════════════════════════════════════════════════"
echo "  ✅ Pre-release checklist complete — ready to tag a release."
echo "════════════════════════════════════════════════════════════════"
