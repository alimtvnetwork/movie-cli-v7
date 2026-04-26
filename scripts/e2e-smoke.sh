#!/usr/bin/env bash
# e2e-smoke.sh — CLI-level smoke test for the `movie` binary.
#
# Builds the binary, scans a temp folder of fake video files, then runs the
# read-only commands users hit most often. Network calls are disabled by
# clearing TMDB_API_KEY / TMDB_TOKEN / OMDB_API_KEY, so this test is safe
# to run offline and in CI without any secrets.
#
# Exit codes:
#   0 — all smoke checks passed
#   non-zero — first failing command's exit code
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WORK_DIR="$(mktemp -d -t movie-e2e.XXXXXX)"
BINARY="${WORK_DIR}/movie"
SCAN_DIR="${WORK_DIR}/movies"
export HOME="${WORK_DIR}/home"
export USERPROFILE="${HOME}"
export TMDB_API_KEY=""
export TMDB_TOKEN=""
export OMDB_API_KEY=""

cleanup() { rm -rf "${WORK_DIR}"; }
trap cleanup EXIT

echo "▶ Building movie CLI → ${BINARY}"
( cd "${REPO_ROOT}" && go build -o "${BINARY}" . )

echo "▶ Preparing fake media in ${SCAN_DIR}"
mkdir -p "${SCAN_DIR}" "${HOME}"
: > "${SCAN_DIR}/The.Matrix.1999.1080p.BluRay.x264.mkv"
: > "${SCAN_DIR}/Inception (2010) [1080p].mp4"
: > "${SCAN_DIR}/Interstellar.2014.2160p.HDR.mkv"

run() {
  echo "▶ movie $*"
  "${BINARY}" "$@"
}

run version
run scan "${SCAN_DIR}"

echo "▶ movie ls --format json (default = scanned only)"
JSON_OUT="$("${BINARY}" ls --format json)"
echo "${JSON_OUT}" | head -n 5
COUNT="$(printf '%s' "${JSON_OUT}" | grep -oE '"id"' | wc -l | tr -d ' ')"
if [ "${COUNT}" -lt 3 ]; then
  echo "❌ expected ≥3 scanned items, got ${COUNT}"
  exit 1
fi

run ls --missing --format json
run ls --all     --format json
run stats

echo "✅ E2E smoke passed"