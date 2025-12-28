#!/usr/bin/env bash
set -euo pipefail

# Compare providers (via Pinocchio profiles) for `prescribe generate` output completeness.
#
# Runs the same tiny repo session through multiple PINOCCHIO_PROFILE values and captures:
# - stdout (final description output)
# - stderr (debug logs + parsed PR data summary)
#
# Default profiles:
# - gemini-2.5-pro
# - o4-mini
# - sonnet-4.5
#
# Environment overrides:
# - PRESCRIBE_ROOT: prescribe module directory (default: current workspace)
# - TEST_REPO_DIR: repo path to create/use (default: /tmp/prescribe-provider-compare-repo)
# - TARGET_BRANCH: default: master
# - BASE: artifact prefix (default: /tmp/prescribe-provider-compare-<timestamp>)
# - PROFILES: comma-separated list of profile names to run (default: gemini-2.5-pro,o4-mini,sonnet-4.5)
# - LOG_LEVEL: default: debug
# - EXTRA_ARGS: extra args appended to `prescribe generate` (space-separated; optional)
# - PINOCCHIO_PROFILE_FILE: optional override for profiles.yaml

PRESCRIBE_ROOT="${PRESCRIBE_ROOT:-/home/manuel/workspaces/2025-12-26/prescribe-import/prescribe}"
TEST_REPO_DIR="${TEST_REPO_DIR:-/tmp/prescribe-provider-compare-repo}"
TARGET_BRANCH="${TARGET_BRANCH:-master}"
BASE="${BASE:-/tmp/prescribe-provider-compare-$(date +%Y%m%d-%H%M%S)}"
PROFILES="${PROFILES:-gemini-2.5-pro,o4-mini,sonnet-4.5}"
LOG_LEVEL="${LOG_LEVEL:-debug}"
EXTRA_ARGS="${EXTRA_ARGS:-}"

LOG="${BASE}.log"

run_quiet() {
  local label="$1"
  shift
  {
    echo
    echo "==> ${label}"
    "$@"
  } >>"$LOG" 2>&1
}

echo "prescribe provider compare" >"$LOG"
echo "PRESCRIBE_ROOT=${PRESCRIBE_ROOT}" >>"$LOG"
echo "TEST_REPO_DIR=${TEST_REPO_DIR}" >>"$LOG"
echo "TARGET_BRANCH=${TARGET_BRANCH}" >>"$LOG"
echo "BASE=${BASE}" >>"$LOG"
echo "PROFILES=${PROFILES}" >>"$LOG"
echo "LOG_LEVEL=${LOG_LEVEL}" >>"$LOG"
echo "PINOCCHIO_PROFILE_FILE=${PINOCCHIO_PROFILE_FILE:-}" >>"$LOG"

# 1) Create tiny repo
run_quiet "setup small test repo" env TEST_REPO_DIR="$TEST_REPO_DIR" bash "${PRESCRIBE_ROOT}/test-scripts/setup-test-repo.sh"

# 2) Initialize + save session
run_quiet "session init/save" bash -c \
  "cd \"$PRESCRIBE_ROOT\" && go run ./cmd/prescribe --repo \"$TEST_REPO_DIR\" --target \"$TARGET_BRANCH\" session init --save --title \"provider compare\" --description \"compare YAML completeness\""

IFS=',' read -r -a profile_list <<<"$PROFILES"

for p in "${profile_list[@]}"; do
  p="$(echo "$p" | xargs)" # trim
  [ -z "$p" ] && continue

  OUT_TXT="${BASE}.${p}.out.txt"
  ERR_TXT="${BASE}.${p}.err.txt"

  run_quiet "generate (${p})" bash -c \
    "cd \"$PRESCRIBE_ROOT\" && PINOCCHIO_PROFILE=\"$p\" ${PINOCCHIO_PROFILE_FILE:+PINOCCHIO_PROFILE_FILE=\"$PINOCCHIO_PROFILE_FILE\"} go run ./cmd/prescribe --log-level \"$LOG_LEVEL\" --repo \"$TEST_REPO_DIR\" --target \"$TARGET_BRANCH\" generate --stream ${EXTRA_ARGS} >\"$OUT_TXT\" 2>\"$ERR_TXT\""
done

echo "=== prescribe provider compare ==="
echo "base=${BASE}"
echo "log=${LOG}"
echo
echo "Artifacts:"
for p in "${profile_list[@]}"; do
  p="$(echo "$p" | xargs)"
  [ -z "$p" ] && continue
  echo "- ${BASE}.${p}.out.txt"
  echo "- ${BASE}.${p}.err.txt"
done
echo
echo "Tip: grep parsed summary:"
echo "  grep -n \"Parsed PR data\" -n ${BASE}.*.err.txt"
echo
echo "done"


