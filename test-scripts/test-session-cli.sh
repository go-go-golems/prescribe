#!/usr/bin/env bash
set -euo pipefail

# Smoke test for session-centric flows (init/show/save/load) using the current prescribe CLI.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

TEST_REPO_DIR="${TEST_REPO_DIR:-/tmp/prescribe-test-repo}"
TARGET_BRANCH="${TARGET_BRANCH:-master}"

SESSION_FILE="$TEST_REPO_DIR/.pr-builder/session.yaml"

prescribe() {
	(
		cd "$REPO_ROOT" && go run ./cmd/prescribe "$@"
	)
}

echo "=========================================="
echo "prescribe - Session CLI Smoke Test"
echo "=========================================="
echo ""

echo "Setting up test repository..."
bash "$SCRIPT_DIR/setup-test-repo.sh"
echo ""

rm -rf "$TEST_REPO_DIR/.pr-builder"

echo "Test 1: Initialize + save session"
prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" session init --save >/dev/null
test -f "$SESSION_FILE"
echo "✓ Session saved at $SESSION_FILE"
echo ""

echo "Test 2: Show session"
prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" session show >/dev/null
echo "✓ Session show works"
echo ""

echo "Test 3: Save session to custom path"
CUSTOM_SESSION="/tmp/prescribe-custom-session.yaml"
prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" session save "$CUSTOM_SESSION" >/dev/null
test -f "$CUSTOM_SESSION"
echo "✓ Saved $CUSTOM_SESSION"
echo ""

echo "Test 4: Load session from custom path"
rm -f "$SESSION_FILE"
prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" session init --save >/dev/null
prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" session load "$CUSTOM_SESSION" >/dev/null
echo "✓ Loaded $CUSTOM_SESSION"
echo ""

echo "Test 5: Generate using loaded session"
prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" generate >/dev/null
echo "✓ Generate works"
echo ""

echo "=========================================="
echo "All tests passed ✓"
echo "=========================================="
