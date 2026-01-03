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
echo "=========================================="
echo "✓ Smoke tests passed"
echo "=========================================="
