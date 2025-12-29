#!/usr/bin/env bash
set -euo pipefail

# Smoke test for: prescribe create (dry-run) supports --yaml-file and --use-last.
#
# IMPORTANT: We intentionally run in --dry-run mode because the test repo won't have a GH remote.
#
# What it checks:
# - `create --dry-run --yaml-file` reads YAML and prints redacted gh command
# - `create --dry-run --use-last` reads .pr-builder/last-generated-pr.yaml and prints redacted gh command
#
# Environment overrides:
# - PRESCRIBE_ROOT: path to prescribe module (default: current workspace)
# - BASE: output file prefix (default: /tmp/prescribe-create-dry-run-<timestamp>)
#

PRESCRIBE_ROOT="${PRESCRIBE_ROOT:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}"
BASE="${BASE:-/tmp/prescribe-create-dry-run-$(date +%Y%m%d-%H%M%S)}"

LOG="${BASE}.log"
YAML_FILE="${BASE}.pr.yaml"

run_quiet() {
  local label="$1"
  shift
  {
    echo
    echo "==> ${label}"
    "$@"
  } >>"$LOG" 2>&1
}

echo "prescribe create dry-run smoke test" >"$LOG"
echo "PRESCRIBE_ROOT=${PRESCRIBE_ROOT}" >>"$LOG"
echo "BASE=${BASE}" >>"$LOG"

prescribe() {
  ( cd "$PRESCRIBE_ROOT" && go run ./cmd/prescribe "$@" )
}

# 1) yaml-file mode
cat >"$YAML_FILE" <<YAML
title: "Hello"
body: "World"
YAML

run_quiet "create --dry-run --yaml-file" prescribe create --dry-run --yaml-file "$YAML_FILE"
grep -Fq "source: yaml-file:${YAML_FILE}" "$LOG"
grep -Fq "command: gh pr create --title Hello --body <omitted> --base main" "$LOG"

# 2) use-last mode (write .pr-builder/last-generated-pr.yaml via ticket helper)
run_quiet "write last-generated-pr.yaml (ticket helper)" bash -c \
  "cd \"$PRESCRIBE_ROOT\" && go run ./ttmp/2025/12/29/PR-CREATION--implement-pr-creation-end-to-end/scripts/01-write-last-generated-prdata.go --repo . --title LastTitle --body LastBody"

run_quiet "create --dry-run --use-last" prescribe create --dry-run --use-last
grep -Fq "source: use-last:" "$LOG"
grep -Fq "command: gh pr create --title LastTitle --body <omitted> --base main" "$LOG"

echo "=== prescribe create dry-run smoke test ==="
echo "log=${LOG}"
echo "yaml_file=${YAML_FILE}"
echo
echo "Verified:"
echo "- create --dry-run --yaml-file"
echo "- create --dry-run --use-last"
echo
echo "done"


