#!/usr/bin/env bash
set -euo pipefail

# Comprehensive smoke test for the prescribe CLI commands.
#
# This is intentionally "integration-y": it creates a small throwaway git repo under /tmp,
# runs a sequence of CLI commands, and expects them to succeed.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

TEST_REPO_DIR="${TEST_REPO_DIR:-/tmp/prescribe-test-repo}"
TARGET_BRANCH="${TARGET_BRANCH:-master}"

prescribe() {
	(
		cd "$REPO_ROOT" && GOWORK=off go run ./cmd/prescribe "$@"
	)
}

echo "=========================================="
echo "prescribe - Complete Smoke Test Suite"
echo "=========================================="
echo ""

echo "Setting up test repository..."
bash "$SCRIPT_DIR/setup-test-repo.sh"
echo ""

rm -rf "$TEST_REPO_DIR/.pr-builder"

echo "=== PHASE 1: Session Initialization ==="
echo ""
prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" session init --save
prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" session show
echo ""

echo "=== PHASE 2: Filters + Presets ==="
echo ""
prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" filter add --name "Exclude tests" --exclude "**/*test*"
prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" filter preset save --project --name "Exclude docs" --exclude "**/*.md" --exclude "**/docs/**"
prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" filter preset list --project
prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" filter preset apply exclude_docs.yaml
prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" filter list
prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" filter show
echo ""

echo "=== PHASE 3: File Inclusion ==="
echo ""
prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" file toggle tests/auth.test.ts
prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" session show
echo ""

echo "=== PHASE 4: Additional Context ==="
echo ""
prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" context add --note "This PR is part of the auth refactor epic"
prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" context add README.md
prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" session show
echo ""

echo "=== PHASE 5: Generation ==="
echo ""
OUT_FILE="/tmp/prescribe-pr.md"
prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" generate --export-rendered --separator markdown --output-file "$OUT_FILE"
grep -Fq "BEGIN COMMITS" "$OUT_FILE"
grep -Fq "feat: enhance authentication" "$OUT_FILE"
grep -Fq "author=\"Other User\"" "$OUT_FILE"
echo "Generated rendered payload (with git history) at $OUT_FILE"

echo ""
echo "=== PHASE 6: Git-derived context toggles ==="
echo ""

echo "6.1: Disable git history removes commits block"
prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" context git history disable >/dev/null
OUT_NO_COMMITS="$(prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" generate --export-rendered --separator markdown)"
if echo "$OUT_NO_COMMITS" | grep -Fq "BEGIN COMMITS"; then
	echo "Expected commit history to be disabled, but BEGIN COMMITS was present"
	exit 1
fi
echo "✓ Commit history disabled"

echo ""
echo "6.2: Add explicit git_context item appears in exports"
prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" context git add commit HEAD >/dev/null
OUT_GIT_CTX="$(prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" generate --export-rendered --separator markdown)"
echo "$OUT_GIT_CTX" | grep -Fq "<git_commit"
echo "✓ git_context item appears in export-rendered"

echo ""
prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" context git history enable >/dev/null

echo ""
echo "=========================================="
echo "✓ Smoke tests passed"
echo "=========================================="
