#!/usr/bin/env bash
set -euo pipefail

# Smoke test for: prescribe generate loads pinocchio profiles.yaml (bootstrap selection + apply values)
#
# This script intentionally uses a *small* test repo created via prescribe's smoke-test helper,
# so we avoid running against the real (huge) repo.
#
# What it checks:
# - `--profile/--profile-file` flags are accepted
# - profile selection is applied at the right precedence level by verifying parsed provenance
#
# Expected outcome:
# - `generate --print-parsed-parameters` output contains a `profiles` parse step for `separator`
#   and the resolved value is `markdown`.
#
# Environment overrides:
# - PRESCRIBE_ROOT: path to prescribe module (default: current workspace)
# - TEST_REPO_DIR: where to create the tiny git repo (default: /tmp/prescribe-generate-profiles-test-repo)
# - TARGET_BRANCH: base branch for prescribe session (default: master)
# - PROFILE_NAME: profile name to select (default: demo)
# - BASE: output file prefix (default: /tmp/prescribe-generate-profiles-<timestamp>)

PRESCRIBE_ROOT="${PRESCRIBE_ROOT:-/home/manuel/workspaces/2025-12-26/prescribe-import/prescribe}"
TEST_REPO_DIR="${TEST_REPO_DIR:-/tmp/prescribe-generate-profiles-test-repo}"
TARGET_BRANCH="${TARGET_BRANCH:-master}"
PROFILE_NAME="${PROFILE_NAME:-demo}"
BASE="${BASE:-/tmp/prescribe-generate-profiles-$(date +%Y%m%d-%H%M%S)}"

LOG="${BASE}.log"
PROFILES_YAML="${BASE}.profiles.yaml"
PARSED_TXT="${BASE}.print-parsed-parameters.txt"

run_quiet() {
  local label="$1"
  shift
  {
    echo
    echo "==> ${label}"
    "$@"
  } >>"$LOG" 2>&1
}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "prescribe generate profiles smoke test" >"$LOG"
echo "PRESCRIBE_ROOT=${PRESCRIBE_ROOT}" >>"$LOG"
echo "TEST_REPO_DIR=${TEST_REPO_DIR}" >>"$LOG"
echo "TARGET_BRANCH=${TARGET_BRANCH}" >>"$LOG"
echo "PROFILE_NAME=${PROFILE_NAME}" >>"$LOG"
echo "BASE=${BASE}" >>"$LOG"

# 1) Create tiny repo (ensure TEST_REPO_DIR is propagated to the helper script)
run_quiet "setup small test repo" env TEST_REPO_DIR="$TEST_REPO_DIR" bash "${PRESCRIBE_ROOT}/test-scripts/setup-test-repo.sh"

# 2) Create a tiny profiles.yaml that sets a generate default-layer param:
# GenerateExtraSettings.Separator lives in the default layer as `separator`.
cat >"$PROFILES_YAML" <<YAML
${PROFILE_NAME}:
  default:
    separator: markdown
YAML

# 3) Use go run so we always test the current source tree.
prescribe() {
  ( cd "$PRESCRIBE_ROOT" && go run ./cmd/prescribe "$@" )
}

run_quiet "session init/save" prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" session init --save

# 4) Print parsed parameters so we can see provenance for `separator`
run_quiet "generate --print-parsed-parameters (profile-selected)" bash -c \
  "cd \"$PRESCRIBE_ROOT\" && go run ./cmd/prescribe --repo \"$TEST_REPO_DIR\" --target \"$TARGET_BRANCH\" generate --profile \"$PROFILE_NAME\" --profile-file \"$PROFILES_YAML\" --print-parsed-parameters > \"$PARSED_TXT\""

# 5) Minimal stdout summary
echo "=== prescribe generate profiles smoke test ==="
echo "profiles_yaml=${PROFILES_YAML}"
echo "parsed_parameters=${PARSED_TXT}"
echo "log=${LOG}"
echo
echo "Looking for: separator=markdown coming from source=profiles"
echo

# Display a small excerpt around "separator" for quick human inspection (and verify provenance).
python3 "${SCRIPT_DIR}/extract-parsed-parameter-snippet.py" \
  --file "$PARSED_TXT" \
  --param separator \
  --context 12 \
  --require-source profiles

echo
echo "done"


