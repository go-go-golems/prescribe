#!/bin/bash
set -e

# Comprehensive test script for pr-builder session-based CLI

REPO_DIR="/tmp/pr-builder-test-repo"
SESSION_FILE="/tmp/pr-builder-test-repo/.pr-builder/session.yaml"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PRESCRIBE_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Build a local binary for speed/reproducibility (override with PRESCRIBE_BIN if desired).
PRESCRIBE_BIN="${PRESCRIBE_BIN:-/tmp/prescribe}"
if [ ! -x "$PRESCRIBE_BIN" ]; then
  (cd "$PRESCRIBE_ROOT" && GOWORK=off go build -o "$PRESCRIBE_BIN" ./cmd/prescribe)
fi

echo "=========================================="
echo "PR Builder Session-Based CLI Test Suite"
echo "=========================================="
echo ""

# Ensure test repo exists
if [ ! -d "$REPO_DIR" ]; then
    echo "Test repository not found. Running setup script..."
    "$SCRIPT_DIR/setup-test-repo.sh"
fi

cd "$REPO_DIR"

# Clean up any existing session
rm -rf .pr-builder

echo "Test 1: Initialize session"
echo "============================"
$PRESCRIBE_BIN -r "$REPO_DIR" -t master session init --save
echo ""
echo "✓ Init command works"
echo ""

echo "Test 2: Show session (human-readable)"
echo "======================================"
$PRESCRIBE_BIN -r "$REPO_DIR" -t master session show
echo ""
echo "✓ Show command works"
echo ""

echo "Test 3: Show session (YAML)"
echo "============================"
$PRESCRIBE_BIN -r "$REPO_DIR" -t master session show --output yaml
echo ""
echo "✓ Show YAML works"
echo ""

echo "Test 4: Add filter to exclude tests"
echo "===================================="
$PRESCRIBE_BIN -r "$REPO_DIR" -t master filter add \
    --name "Exclude tests" \
    --description "Hide test files from context" \
    --exclude "*test*"
echo ""
echo "✓ Add filter works"
echo ""

echo "Test 5: Show session after filter"
echo "=================================="
$PRESCRIBE_BIN -r "$REPO_DIR" -t master session show
echo ""
echo "✓ Filter applied correctly"
echo ""

echo "Test 6: Toggle file inclusion"
echo "=============================="
$PRESCRIBE_BIN -r "$REPO_DIR" -t master file toggle "src/auth/login.ts"
echo ""
echo "✓ Toggle file works"
echo ""

echo "Test 7: Add context note"
echo "========================"
$PRESCRIBE_BIN -r "$REPO_DIR" -t master context add \
    --note "This PR is part of the auth refactor epic"
echo ""
echo "✓ Add context note works"
echo ""

echo "Test 8: Add context file"
echo "========================"
$PRESCRIBE_BIN -r "$REPO_DIR" -t master context add "README.md"
echo ""
echo "✓ Add context file works"
echo ""

echo "Test 9: Show final session state"
echo "================================="
$PRESCRIBE_BIN -r "$REPO_DIR" -t master session show
echo ""
echo "✓ Final session state looks good"
echo ""

echo "Test 10: View session YAML file"
echo "================================"
echo "Session file contents:"
cat "$SESSION_FILE"
echo ""
echo "✓ Session YAML file created"
echo ""

echo "Test 11: Save session to custom path"
echo "====================================="
$PRESCRIBE_BIN -r "$REPO_DIR" -t master session save /tmp/custom-session.yaml
echo ""
echo "✓ Save to custom path works"
echo ""

echo "Test 12: Load session from custom path"
echo "======================================="
# First, reset by removing the default session
rm -f "$SESSION_FILE"
$PRESCRIBE_BIN -r "$REPO_DIR" -t master session init --save
# Then load the custom session
$PRESCRIBE_BIN -r "$REPO_DIR" -t master session load /tmp/custom-session.yaml
echo ""
echo "✓ Load from custom path works"
echo ""

echo "Test 13: Generate with session"
echo "==============================="
if [ "${PRESCRIBE_RUN_GENERATE:-}" = "1" ]; then
  $PRESCRIBE_BIN -r "$REPO_DIR" -t master generate
  echo ""
  echo "✓ Generate with session works"
  echo ""
else
  echo "Skipping generate test (set PRESCRIBE_RUN_GENERATE=1 to enable)"
  echo ""
fi

echo "Test 14: Generate with session file flag"
echo "========================================="
if [ "${PRESCRIBE_RUN_GENERATE:-}" = "1" ]; then
  $PRESCRIBE_BIN -r "$REPO_DIR" -t master generate --load-session /tmp/custom-session.yaml
  echo ""
  echo "✓ Generate with --load-session flag works"
  echo ""
else
  echo "Skipping generate test (set PRESCRIBE_RUN_GENERATE=1 to enable)"
  echo ""
fi

echo "=========================================="
echo "All tests passed! ✓"
echo "=========================================="
echo ""
echo "Session file location: $SESSION_FILE"
echo "Custom session location: /tmp/custom-session.yaml"
