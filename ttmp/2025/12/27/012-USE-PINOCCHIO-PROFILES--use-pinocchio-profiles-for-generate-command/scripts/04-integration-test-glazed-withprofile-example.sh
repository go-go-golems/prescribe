#!/usr/bin/env bash
set -euo pipefail

# Integration-ish test for: glazed appconfig.WithProfile end-to-end behavior via the example program.
#
# What it checks:
# - default profile applies (redis.host=from-profile-default)
# - env-selected profile applies when env parsing is enabled (MYAPP_PROFILE=prod)
# - config overrides profiles when enabled (--use-config => redis.host=from-config)
#
# Environment overrides:
# - WORKSPACE_ROOT: repo root (default: /home/manuel/workspaces/2025-12-26/prescribe-import)
# - GLAZED_DIR: glazed module directory (default: $WORKSPACE_ROOT/glazed)
# - BASE: output file prefix (default: /tmp/glazed-withprofile-example-<timestamp>)

WORKSPACE_ROOT="${WORKSPACE_ROOT:-/home/manuel/workspaces/2025-12-26/prescribe-import}"
GLAZED_DIR="${GLAZED_DIR:-${WORKSPACE_ROOT}/glazed}"
BASE="${BASE:-/tmp/glazed-withprofile-example-$(date +%Y%m%d-%H%M%S)}"

LOG="${BASE}.log"
OUT_DEFAULT="${BASE}.default.txt"
OUT_ENV="${BASE}.env.txt"
OUT_CONFIG="${BASE}.config.txt"

run() {
  local label="$1"
  shift
  {
    echo
    echo "==> ${label}"
    "$@"
  } >>"$LOG" 2>&1
}

echo "glazed WithProfile example integration test" >"$LOG"
echo "GLAZED_DIR=${GLAZED_DIR}" >>"$LOG"
echo "BASE=${BASE}" >>"$LOG"

run "default (no env, no flags)" bash -c "cd \"$GLAZED_DIR\" && go run ./cmd/examples/appconfig-profiles > \"$OUT_DEFAULT\""
grep -Fq "redis.host=from-profile-default" "$OUT_DEFAULT"

run "env selects prod" bash -c "cd \"$GLAZED_DIR\" && MYAPP_PROFILE=prod go run ./cmd/examples/appconfig-profiles > \"$OUT_ENV\""
grep -Fq "redis.host=from-profile-prod" "$OUT_ENV"

run "config overrides profiles (--use-config)" bash -c "cd \"$GLAZED_DIR\" && go run ./cmd/examples/appconfig-profiles --use-config > \"$OUT_CONFIG\""
grep -Fq "redis.host=from-config" "$OUT_CONFIG"

echo "=== glazed WithProfile example integration test ==="
echo "default_output=${OUT_DEFAULT}"
echo "env_output=${OUT_ENV}"
echo "config_output=${OUT_CONFIG}"
echo "log=${LOG}"
echo
echo "done"


