#!/bin/bash
set -e

# Test script for pr-builder CLI

REPO_DIR="/tmp/pr-builder-test-repo"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PRESCRIBE_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Build a local binary for speed/reproducibility (override with PRESCRIBE_BIN if desired).
PRESCRIBE_BIN="${PRESCRIBE_BIN:-/tmp/prescribe}"
if [ ! -x "$PRESCRIBE_BIN" ]; then
  (cd "$PRESCRIBE_ROOT" && go build -o "$PRESCRIBE_BIN" ./cmd/prescribe)
fi

echo "=========================================="
echo "PR Builder CLI Test Suite"
echo "=========================================="
echo ""

# Ensure test repo exists
if [ ! -d "$REPO_DIR" ]; then
    echo "Test repository not found. Running setup script..."
    "$SCRIPT_DIR/setup-test-repo.sh"
fi

cd "$REPO_DIR"

echo "Test 1: Show help"
echo "===================="
$PRESCRIBE_BIN --help
echo ""
echo "✓ Help command works"
echo ""

echo "Test 2: Show version"
echo "===================="
$PRESCRIBE_BIN --version
echo ""
echo "✓ Version command works"
echo ""

echo "Test 3: Session init + show"
echo "============================"
$PRESCRIBE_BIN -r "$REPO_DIR" -t master session init --save
$PRESCRIBE_BIN -r "$REPO_DIR" -t master session show --output json
echo ""
echo "✓ Session init/show works"
echo ""

echo "Test 4: Filter list"
echo "===================="
$PRESCRIBE_BIN -r "$REPO_DIR" -t master filter list --output json
echo ""
echo "✓ Filter list works"
echo ""

echo "Test 5: Generate with output file (optional)"
echo "===================="
if [ "${PRESCRIBE_RUN_GENERATE:-}" = "1" ]; then
  $PRESCRIBE_BIN -r "$REPO_DIR" -t master generate -o /tmp/pr-description.md
  echo "Generated description saved to /tmp/pr-description.md"
  cat /tmp/pr-description.md
  echo ""
  echo "✓ Generate with output file works"
  echo ""
else
  echo "Skipping generate test (set PRESCRIBE_RUN_GENERATE=1 to enable)"
  echo ""
fi

echo "=========================================="
echo "All tests passed! ✓"
echo "=========================================="
