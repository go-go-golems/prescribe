#!/usr/bin/env bash
set -euo pipefail

# Lightweight CLI smoke test: help/version + minimal session init + generate.

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
echo "prescribe - CLI Smoke Test"
echo "=========================================="
echo ""

echo "Setting up test repository..."
bash "$SCRIPT_DIR/setup-test-repo.sh"
echo ""

echo "Test 1: Help"
prescribe --help >/dev/null
echo "✓ Help works"
echo ""

echo "Test 2: Version"
prescribe --version >/dev/null
echo "✓ Version works"
echo ""

echo "Test 3: Session init + show"
rm -rf "$TEST_REPO_DIR/.pr-builder"
prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" session init --save
prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" session show >/dev/null
echo "✓ Session init/show works"
echo ""

echo "Test 4: Generate (export rendered payload, includes git history)"
OUT="$(prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" generate --export-rendered --separator markdown)"
echo "$OUT" | grep -Fq "BEGIN COMMITS"
echo "$OUT" | grep -Fq "feat: enhance authentication"
echo "$OUT" | grep -Fq "author=\"Other User\""
echo "✓ Generate export-rendered includes commit history"
echo ""

echo "=========================================="
echo "All tests passed ✓"
echo "=========================================="
