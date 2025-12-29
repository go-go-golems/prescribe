#!/usr/bin/env bash
set -euo pipefail

# Integration test for end-to-end PR creation flows.
#
# What it covers (safe/offline GitHub side):
# - `prescribe create --yaml-file` (non-dry-run) using a fake `gh` binary
# - `prescribe generate` persists `.pr-builder/last-generated-pr.yaml`
# - `prescribe create --use-last` (non-dry-run) using the persisted YAML
# - `prescribe generate --create` calls git push + gh pr create (fake)
#
# Notes:
# - We configure a local bare git remote so `git push` succeeds without network.
# - We inject a fake `gh` via PATH so no real PR is created on GitHub.
# - `prescribe generate` still requires real AI step settings (profile / API keys).
#
# Environment overrides:
# - PRESCRIBE_ROOT: path to prescribe module (default: repo root of this script)
# - TEST_REPO_DIR: path for the temporary test repo (default: /tmp/prescribe-test-repo)
# - TARGET_BRANCH: base branch for repo diff (default: master)
# - BASE: output prefix for logs/artifacts (default: /tmp/prescribe-pr-creation-e2e-<timestamp>)
# - SKIP_GENERATE: if set to 1, skip generate-related steps (create-only coverage)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PRESCRIBE_ROOT="${PRESCRIBE_ROOT:-$(cd "$SCRIPT_DIR/.." && pwd)}"

TEST_REPO_DIR="${TEST_REPO_DIR:-/tmp/prescribe-test-repo}"
TARGET_BRANCH="${TARGET_BRANCH:-master}"
BASE="${BASE:-/tmp/prescribe-pr-creation-e2e-$(date +%Y%m%d-%H%M%S)}"
SKIP_GENERATE="${SKIP_GENERATE:-0}"

LOG="${BASE}.log"
GH_LOG="${BASE}.gh.log"
YAML_FILE="${BASE}.pr.yaml"
REMOTE_DIR="${BASE}.remote.git"

run_quiet() {
  local label="$1"
  shift
  {
    echo
    echo "==> ${label}"
    "$@"
  } >>"$LOG" 2>&1
}

prescribe() {
  ( cd "$PRESCRIBE_ROOT" && go run ./cmd/prescribe "$@" )
}

echo "prescribe PR creation integration test" >"$LOG"
echo "PRESCRIBE_ROOT=${PRESCRIBE_ROOT}" >>"$LOG"
echo "TEST_REPO_DIR=${TEST_REPO_DIR}" >>"$LOG"
echo "TARGET_BRANCH=${TARGET_BRANCH}" >>"$LOG"
echo "BASE=${BASE}" >>"$LOG"
echo "REMOTE_DIR=${REMOTE_DIR}" >>"$LOG"
echo "GH_LOG=${GH_LOG}" >>"$LOG"
echo "SKIP_GENERATE=${SKIP_GENERATE}" >>"$LOG"

run_quiet "setup small test repo" env TEST_REPO_DIR="$TEST_REPO_DIR" bash "${PRESCRIBE_ROOT}/test-scripts/setup-test-repo.sh"

rm -rf "$TEST_REPO_DIR/.pr-builder"

# Configure a local bare remote so pushes succeed without network access.
run_quiet "init local bare remote" bash -c "rm -rf \"$REMOTE_DIR\" && git init --bare \"$REMOTE_DIR\" >/dev/null"
run_quiet "set origin remote to local bare remote" bash -c "cd \"$TEST_REPO_DIR\" && (git remote remove origin >/dev/null 2>&1 || true) && git remote add origin \"$REMOTE_DIR\""
run_quiet "push master to origin (set upstream)" bash -c "cd \"$TEST_REPO_DIR\" && git push -u origin master:master >/dev/null"
run_quiet "push current branch to origin (set upstream)" bash -c "cd \"$TEST_REPO_DIR\" && git push -u origin HEAD >/dev/null"

# Fake gh binary (captures invocation; returns success with a fake PR URL).
FAKE_BIN="$(mktemp -d)"
cat >"${FAKE_BIN}/gh" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

GH_LOG="${GH_LOG:-/tmp/prescribe-fake-gh.log}"
{
  echo "gh $*"
} >>"$GH_LOG"

if [[ "${1:-}" == "pr" && "${2:-}" == "create" ]]; then
  echo "https://example.invalid/fake/pr/1"
  exit 0
fi

echo "fake-gh: unsupported command: $*" >&2
exit 2
EOF
chmod +x "${FAKE_BIN}/gh"

export GH_LOG
export PATH="${FAKE_BIN}:${PATH}"

# 1) create --yaml-file (non-dry-run)
cat >"$YAML_FILE" <<YAML
title: "Hello"
body: "World"
YAML

run_quiet "session init (required before generate; harmless otherwise)" \
  prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" session init --save

run_quiet "create --yaml-file (non-dry-run; fake gh)" \
  prescribe --repo "$TEST_REPO_DIR" create --yaml-file "$YAML_FILE" --base "$TARGET_BRANCH" --draft

grep -Fq "https://example.invalid/fake/pr/1" "$LOG"
grep -Fq "gh pr create" "$GH_LOG"
grep -Fq -- "--title Hello" "$GH_LOG"
grep -Fq -- "--base ${TARGET_BRANCH}" "$GH_LOG"
grep -Fq -- "--base ${TARGET_BRANCH}" "$GH_LOG"
grep -Fq -- "--draft" "$GH_LOG"

# 2) generate -> persisted last-generated-pr.yaml
if [[ "$SKIP_GENERATE" == "1" ]]; then
  run_quiet "SKIP_GENERATE=1: skipping generate steps" true
else
  GEN_OUT="${BASE}.generated.md"
  run_quiet "generate (writes last-generated-pr.yaml)" \
    prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" generate --output-file "$GEN_OUT"

  LAST_YAML="${TEST_REPO_DIR}/.pr-builder/last-generated-pr.yaml"
  test -f "$LAST_YAML"
  grep -Eq "^title:" "$LAST_YAML"
  grep -Eq "^body:" "$LAST_YAML"

  # 3) create --use-last (non-dry-run)
  run_quiet "create --use-last (non-dry-run; fake gh)" \
    prescribe --repo "$TEST_REPO_DIR" create --use-last --base "$TARGET_BRANCH" --draft

  grep -Fq "gh pr create" "$GH_LOG"

  # 4) generate --create (non-dry-run; fake gh)
  GEN_CREATE_OUT="${BASE}.generated-create.md"
  run_quiet "generate --create (fake gh)" \
    prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" generate --output-file "$GEN_CREATE_OUT" --create --create-base "$TARGET_BRANCH" --create-draft

  grep -Fq "gh pr create" "$GH_LOG"
  grep -Fq -- "--base ${TARGET_BRANCH}" "$GH_LOG"
  grep -Fq -- "--draft" "$GH_LOG"
fi

echo "=== prescribe PR creation integration test ==="
echo "log=${LOG}"
echo "gh_log=${GH_LOG}"
echo "yaml_file=${YAML_FILE}"
echo "remote_dir=${REMOTE_DIR}"
echo
echo "Verified:"
echo "- create --yaml-file (non-dry-run; fake gh; local git remote)"
if [[ "$SKIP_GENERATE" == "1" ]]; then
  echo "- generate/create steps skipped (SKIP_GENERATE=1)"
else
  echo "- generate persists .pr-builder/last-generated-pr.yaml"
  echo "- create --use-last (non-dry-run; fake gh)"
  echo "- generate --create (non-dry-run; fake gh)"
fi
echo
echo "done"


