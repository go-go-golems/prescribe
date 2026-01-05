#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

TEST_REPO_DIR="${TEST_REPO_DIR:-/tmp/prescribe-test-repo}"
TARGET_BRANCH="${TARGET_BRANCH:-master}"

prescribe() {
	(
		cd "$REPO_ROOT" && GOWORK=off go run ./cmd/prescribe "$@"
	)
}

echo "========================================="
echo "prescribe - Filter Functionality Smoke Test"
echo "========================================="
echo

echo "Setting up test repository..."
bash "$SCRIPT_DIR/setup-test-repo.sh"
echo

rm -rf "$TEST_REPO_DIR/.pr-builder"
prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" session init --save >/dev/null

echo "Test 1: Add exclude filter"
prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" filter add --name "Exclude tests" --exclude "**/*test*"
echo "✓ Added"
echo

echo "Test 2: List filters"
prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" filter list >/dev/null
echo "✓ Listed"
echo

echo "Test 3: Show filtered files"
prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" filter show >/dev/null
echo "✓ Show works"
echo

echo "Test 4: Clear filters"
prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" filter clear
prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" filter list >/dev/null
echo "✓ Cleared"
echo

echo "Test 5: Preset save/apply"
prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" filter preset save --project --name "Exclude docs" --exclude "**/*.md"
prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" filter preset apply exclude_docs.yaml
echo "✓ Presets ok"
echo

echo "========================================="
echo "All filter smoke tests passed ✓"
echo "========================================="
