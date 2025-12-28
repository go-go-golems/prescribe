#!/usr/bin/env bash
set -euo pipefail

# Compare Gemini output completeness between `prescribe generate --stream` and non-streaming `prescribe generate`.
#
# Goal:
# - Determine whether partial/invalid YAML (e.g. bare `body` line) correlates with streaming.
# - Capture stderr (debug logs) including:
#   - assistant preview/hash
#   - stop_reason + token usage (if provider reports it)
#
# Requirements:
# - You have a Pinocchio profile named "gemini-2.5-pro" (or override PINOCCHIO_PROFILE) in your profiles.yaml
# - That profile supplies gemini API keys and engine/model.
#
# Environment overrides:
# - PRESCRIBE_ROOT: prescribe module directory (default: current workspace)
# - TEST_REPO_DIR: where to create the tiny git repo (default: /tmp/prescribe-gemini-stream-vs-nonstream-repo)
# - TARGET_BRANCH: base branch for prescribe session (default: master)
# - PINOCCHIO_PROFILE: profile to use (default: gemini-2.5-pro)
# - PINOCCHIO_PROFILE_FILE: optional override for profiles.yaml
# - BASE: artifact prefix (default: /tmp/prescribe-gemini-stream-vs-nonstream-<timestamp>)
# - LOG_LEVEL: default: debug
# - EXTRA_ARGS: extra args appended to `prescribe generate` (space-separated; optional)

PRESCRIBE_ROOT="${PRESCRIBE_ROOT:-/home/manuel/workspaces/2025-12-26/prescribe-import/prescribe}"
TEST_REPO_DIR="${TEST_REPO_DIR:-/tmp/prescribe-gemini-stream-vs-nonstream-repo}"
TARGET_BRANCH="${TARGET_BRANCH:-master}"
PINOCCHIO_PROFILE="${PINOCCHIO_PROFILE:-gemini-2.5-pro}"
BASE="${BASE:-/tmp/prescribe-gemini-stream-vs-nonstream-$(date +%Y%m%d-%H%M%S)}"
LOG_LEVEL="${LOG_LEVEL:-debug}"
EXTRA_ARGS="${EXTRA_ARGS:-}"

LOG="${BASE}.log"

OUT_STREAM="${BASE}.stream.out.txt"
ERR_STREAM="${BASE}.stream.err.txt"
OUT_NONSTREAM="${BASE}.nonstream.out.txt"
ERR_NONSTREAM="${BASE}.nonstream.err.txt"

run_quiet() {
  local label="$1"
  shift
  {
    echo
    echo "==> ${label}"
    "$@"
  } >>"$LOG" 2>&1
}

echo "prescribe gemini stream vs nonstream compare" >"$LOG"
echo "PRESCRIBE_ROOT=${PRESCRIBE_ROOT}" >>"$LOG"
echo "TEST_REPO_DIR=${TEST_REPO_DIR}" >>"$LOG"
echo "TARGET_BRANCH=${TARGET_BRANCH}" >>"$LOG"
echo "PINOCCHIO_PROFILE=${PINOCCHIO_PROFILE}" >>"$LOG"
echo "PINOCCHIO_PROFILE_FILE=${PINOCCHIO_PROFILE_FILE:-}" >>"$LOG"
echo "BASE=${BASE}" >>"$LOG"
echo "LOG_LEVEL=${LOG_LEVEL}" >>"$LOG"

# 1) Create tiny repo
run_quiet "setup small test repo" env TEST_REPO_DIR="$TEST_REPO_DIR" bash "${PRESCRIBE_ROOT}/test-scripts/setup-test-repo.sh"

# 2) Initialize + save a session in the small repo (shared across both runs)
run_quiet "session init/save" bash -c \
  "cd \"$PRESCRIBE_ROOT\" && go run ./cmd/prescribe --repo \"$TEST_REPO_DIR\" --target \"$TARGET_BRANCH\" session init --save \
    --title \"gemini stream vs nonstream\" \
    --description \"compare YAML completeness between generate modes\""

# 3) Streaming run
run_quiet "generate (stream)" bash -c \
  "cd \"$PRESCRIBE_ROOT\" && PINOCCHIO_PROFILE=\"$PINOCCHIO_PROFILE\" ${PINOCCHIO_PROFILE_FILE:+PINOCCHIO_PROFILE_FILE=\"$PINOCCHIO_PROFILE_FILE\"} \
    go run ./cmd/prescribe --log-level \"$LOG_LEVEL\" --repo \"$TEST_REPO_DIR\" --target \"$TARGET_BRANCH\" generate --stream ${EXTRA_ARGS} >\"$OUT_STREAM\" 2>\"$ERR_STREAM\""

# 4) Non-streaming run (same repo/session)
run_quiet "generate (non-stream)" bash -c \
  "cd \"$PRESCRIBE_ROOT\" && PINOCCHIO_PROFILE=\"$PINOCCHIO_PROFILE\" ${PINOCCHIO_PROFILE_FILE:+PINOCCHIO_PROFILE_FILE=\"$PINOCCHIO_PROFILE_FILE\"} \
    go run ./cmd/prescribe --log-level \"$LOG_LEVEL\" --repo \"$TEST_REPO_DIR\" --target \"$TARGET_BRANCH\" generate ${EXTRA_ARGS} >\"$OUT_NONSTREAM\" 2>\"$ERR_NONSTREAM\""

echo "=== prescribe gemini stream vs nonstream compare ==="
echo "profile=${PINOCCHIO_PROFILE}"
echo "base=${BASE}"
echo "log=${LOG}"
echo
echo "Artifacts:"
echo "- ${OUT_STREAM}"
echo "- ${ERR_STREAM}"
echo "- ${OUT_NONSTREAM}"
echo "- ${ERR_NONSTREAM}"
echo
echo "Tip: inspect stop_reason + token usage (if present):"
echo "  grep -n \"stop_reason\\|input_tokens\\|output_tokens\" -n \"$ERR_STREAM\" \"$ERR_NONSTREAM\""
echo
echo "Tip: inspect parse summary:"
echo "  grep -n \"Parsed PR data\" -n \"$ERR_STREAM\" \"$ERR_NONSTREAM\""
echo
echo "done"


