#!/usr/bin/env bash
set -euo pipefail

# Smoke test for: title/description persisted in session.yaml and present in rendered prompt payload.
#
# What it checks:
# - `session init --save --title/--description` persists values into session.yaml
# - `generate --export-rendered` renders those values into the prompt template variables (.title/.description)
#
# Environment overrides:
# - PRESCRIBE_ROOT: path to prescribe module (default: current workspace)
# - TEST_REPO_DIR: where to create the tiny git repo (default: /tmp/prescribe-generate-title-desc-test-repo)
# - TARGET_BRANCH: base branch for prescribe session (default: master)
# - PR_TITLE: title to set via CLI (default: "demo title")
# - PR_DESCRIPTION: description to set via CLI (default: "demo description")
# - BASE: output file prefix (default: /tmp/prescribe-generate-title-desc-<timestamp>)

PRESCRIBE_ROOT="${PRESCRIBE_ROOT:-/home/manuel/workspaces/2025-12-26/prescribe-import/prescribe}"
TEST_REPO_DIR="${TEST_REPO_DIR:-/tmp/prescribe-generate-title-desc-test-repo}"
TARGET_BRANCH="${TARGET_BRANCH:-master}"
PR_TITLE="${PR_TITLE:-demo title}"
PR_DESCRIPTION="${PR_DESCRIPTION:-demo description}"
BASE="${BASE:-/tmp/prescribe-generate-title-desc-$(date +%Y%m%d-%H%M%S)}"

LOG="${BASE}.log"
RENDERED_TXT="${BASE}.rendered.txt"

run_quiet() {
  local label="$1"
  shift
  {
    echo
    echo "==> ${label}"
    "$@"
  } >>"$LOG" 2>&1
}

echo "prescribe generate title/description smoke test" >"$LOG"
echo "PRESCRIBE_ROOT=${PRESCRIBE_ROOT}" >>"$LOG"
echo "TEST_REPO_DIR=${TEST_REPO_DIR}" >>"$LOG"
echo "TARGET_BRANCH=${TARGET_BRANCH}" >>"$LOG"
echo "PR_TITLE=${PR_TITLE}" >>"$LOG"
echo "PR_DESCRIPTION=${PR_DESCRIPTION}" >>"$LOG"
echo "BASE=${BASE}" >>"$LOG"

# 1) Create tiny repo (ensure TEST_REPO_DIR is propagated to the helper script)
run_quiet "setup small test repo" env TEST_REPO_DIR="$TEST_REPO_DIR" bash "${PRESCRIBE_ROOT}/test-scripts/setup-test-repo.sh"

# 2) Use go run so we always test the current source tree.
prescribe() {
  ( cd "$PRESCRIBE_ROOT" && go run ./cmd/prescribe "$@" )
}

# 3) Initialize + save session with title/description
run_quiet "session init/save (with title/description)" prescribe \
  --repo "$TEST_REPO_DIR" \
  --target "$TARGET_BRANCH" \
  session init --save \
  --title "$PR_TITLE" \
  --description "$PR_DESCRIPTION"

# 4) Export rendered payload (system + user)
run_quiet "generate --export-rendered" bash -c \
  "cd \"$PRESCRIBE_ROOT\" && go run ./cmd/prescribe --repo \"$TEST_REPO_DIR\" --target \"$TARGET_BRANCH\" generate --export-rendered > \"$RENDERED_TXT\""

# 5) Assert rendered output contains title/description
grep -Fq "$PR_TITLE" "$RENDERED_TXT"
grep -Fq "$PR_DESCRIPTION" "$RENDERED_TXT"

# 6) Minimal stdout summary
echo "=== prescribe generate title/description smoke test ==="
echo "rendered_payload=${RENDERED_TXT}"
echo "log=${LOG}"
echo
echo "Verified rendered prompt contains:"
echo "- title: $PR_TITLE"
echo "- description: $PR_DESCRIPTION"
echo
echo "done"


