#!/usr/bin/env bash
set -euo pipefail

# Smoke test for: running `prescribe generate` using a Pinocchio profile (Gemini).
#
# Intended usage:
# - You have a profile named "gemini-2.5-pro" in ~/.config/pinocchio/profiles.yaml
# - That profile sets at least:
#   - ai-api-type=gemini
#   - ai-engine=<your gemini model id>
#   - gemini-api-key=<secret> (or otherwise supplies gemini-api-key)
#
# This script:
# - creates a tiny throwaway git repo
# - initializes a prescribe session
# - runs `generate --stream` so stderr includes the parsed PR data summary
# - captures stdout + stderr to /tmp artifacts for inspection
#
# Environment overrides:
# - PRESCRIBE_ROOT: path to prescribe module (default: current workspace)
# - TEST_REPO_DIR: where to create the tiny git repo (default: /tmp/prescribe-gemini-profile-test-repo)
# - TARGET_BRANCH: base branch for prescribe session (default: master)
# - PINOCCHIO_PROFILE: which profile to use (default: gemini-2.5-pro)
# - PINOCCHIO_PROFILE_FILE: optional override for profiles.yaml
# - BASE: output file prefix (default: /tmp/prescribe-gemini-profile-<timestamp>)
# - EXTRA_ARGS: extra args appended to `prescribe generate` (space-separated; optional)

PRESCRIBE_ROOT="${PRESCRIBE_ROOT:-/home/manuel/workspaces/2025-12-26/prescribe-import/prescribe}"
TEST_REPO_DIR="${TEST_REPO_DIR:-/tmp/prescribe-gemini-profile-test-repo}"
TARGET_BRANCH="${TARGET_BRANCH:-master}"
PINOCCHIO_PROFILE="${PINOCCHIO_PROFILE:-gemini-2.5-pro}"
BASE="${BASE:-/tmp/prescribe-gemini-profile-$(date +%Y%m%d-%H%M%S)}"
EXTRA_ARGS="${EXTRA_ARGS:-}"

LOG="${BASE}.log"
OUT_TXT="${BASE}.out.txt"
ERR_TXT="${BASE}.err.txt"

run_quiet() {
  local label="$1"
  shift
  {
    echo
    echo "==> ${label}"
    "$@"
  } >>"$LOG" 2>&1
}

echo "prescribe gemini profile smoke test" >"$LOG"
echo "PRESCRIBE_ROOT=${PRESCRIBE_ROOT}" >>"$LOG"
echo "TEST_REPO_DIR=${TEST_REPO_DIR}" >>"$LOG"
echo "TARGET_BRANCH=${TARGET_BRANCH}" >>"$LOG"
echo "PINOCCHIO_PROFILE=${PINOCCHIO_PROFILE}" >>"$LOG"
echo "PINOCCHIO_PROFILE_FILE=${PINOCCHIO_PROFILE_FILE:-}" >>"$LOG"
echo "BASE=${BASE}" >>"$LOG"

# 1) Create tiny repo (ensure TEST_REPO_DIR is propagated to the helper script)
run_quiet "setup small test repo" env TEST_REPO_DIR="$TEST_REPO_DIR" bash "${PRESCRIBE_ROOT}/test-scripts/setup-test-repo.sh"

# 2) Use go run so we always test the current source tree.
prescribe() {
  ( cd "$PRESCRIBE_ROOT" && go run ./cmd/prescribe "$@" )
}

# 3) Initialize + save a session in the small repo
run_quiet "session init/save" prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" session init --save \
  --title "gemini profile smoke" \
  --description "verify gemini profile selection + YAML parsing"

# 4) Run generation with profile selection via env (streaming so stderr includes parsed PR data summary)
#
# Note: We intentionally *do not* pass --ai-api-type/--ai-engine flags here; the profile should supply them.
run_quiet "generate (gemini profile)" bash -c \
  "cd \"$PRESCRIBE_ROOT\" && PINOCCHIO_PROFILE=\"$PINOCCHIO_PROFILE\" ${PINOCCHIO_PROFILE_FILE:+PINOCCHIO_PROFILE_FILE=\"$PINOCCHIO_PROFILE_FILE\"} go run ./cmd/prescribe --repo \"$TEST_REPO_DIR\" --target \"$TARGET_BRANCH\" generate --stream ${EXTRA_ARGS} >\"$OUT_TXT\" 2>\"$ERR_TXT\""

echo "=== prescribe gemini profile smoke test ==="
echo "profile=${PINOCCHIO_PROFILE}"
echo "stdout=${OUT_TXT}"
echo "stderr=${ERR_TXT}"
echo "log=${LOG}"
echo
echo "tail(stderr):"
tail -n 80 "$ERR_TXT" || true
echo
echo "done"


