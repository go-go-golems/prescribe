#!/bin/bash
set -e

REPO="/tmp/pr-builder-test-repo"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PRESCRIBE_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Build a local binary for speed/reproducibility (override with PRESCRIBE_BIN if desired).
PRESCRIBE_BIN="${PRESCRIBE_BIN:-/tmp/prescribe}"
if [ ! -x "$PRESCRIBE_BIN" ]; then
  (cd "$PRESCRIBE_ROOT" && GOWORK=off go build -o "$PRESCRIBE_BIN" ./cmd/prescribe)
fi

echo "========================================="
echo "PR Builder - Filter Functionality Tests"
echo "========================================="
echo

# Ensure test repo exists
if [ ! -d "$REPO" ]; then
    echo "Setting up test repository..."
    "$SCRIPT_DIR/setup-test-repo.sh"
fi

cd "$REPO"

# Clean start
rm -rf .pr-builder
echo "✓ Clean start"
echo

# Test 1: Initialize session
echo "Test 1: Initialize session"
echo "-------------------------------------------"
$PRESCRIBE_BIN -r "$REPO" -t master session init --save
$PRESCRIBE_BIN -r "$REPO" -t master session show
echo

# Test 2: List filters (should be empty)
echo "Test 2: List filters (empty)"
echo "-------------------------------------------"
$PRESCRIBE_BIN -r "$REPO" -t master filter list
echo

# Test 3: Test a filter without applying
echo "Test 3: Test filter pattern"
echo "-------------------------------------------"
$PRESCRIBE_BIN -r "$REPO" -t master filter test --name "Test Exclude Tests" --exclude "*test*"
echo

# Test 4: Add exclude filter
echo "Test 4: Add exclude filter"
echo "-------------------------------------------"
$PRESCRIBE_BIN -r "$REPO" -t master filter add --name "Exclude tests" --exclude "*test*"
$PRESCRIBE_BIN -r "$REPO" -t master session show
echo

# Test 5: List filters
echo "Test 5: List filters"
echo "-------------------------------------------"
$PRESCRIBE_BIN -r "$REPO" -t master filter list
echo

# Test 6: Show filtered files
echo "Test 6: Show filtered files"
echo "-------------------------------------------"
$PRESCRIBE_BIN -r "$REPO" -t master filter show
echo

# Test 7: Add another filter
echo "Test 7: Add multiple filters"
echo "-------------------------------------------"
$PRESCRIBE_BIN -r "$REPO" -t master filter add --name "Only TypeScript" --include "*.ts"
$PRESCRIBE_BIN -r "$REPO" -t master filter list
echo

# Test 8: Test glob patterns
echo "Test 8: Test various glob patterns"
echo "-------------------------------------------"
echo "Pattern: **/*.test.ts"
$PRESCRIBE_BIN -r "$REPO" -t master filter test --exclude "**/*.test.ts"
echo
echo "Pattern: src/**"
$PRESCRIBE_BIN -r "$REPO" -t master filter test --include "src/**"
echo
echo "Pattern: tests/*"
$PRESCRIBE_BIN -r "$REPO" -t master filter test --exclude "tests/*"
echo

# Test 9: Remove filter by index
echo "Test 9: Remove filter by index"
echo "-------------------------------------------"
$PRESCRIBE_BIN -r "$REPO" -t master filter remove 0
$PRESCRIBE_BIN -r "$REPO" -t master filter list
echo

# Test 10: Remove filter by name
echo "Test 10: Remove filter by name"
echo "-------------------------------------------"
$PRESCRIBE_BIN -r "$REPO" -t master filter add --name "Exclude tests" --exclude "tests/*"
$PRESCRIBE_BIN -r "$REPO" -t master filter remove "Exclude tests"
$PRESCRIBE_BIN -r "$REPO" -t master filter list
echo

# Test 11: Add multiple filters and clear all
echo "Test 11: Clear all filters"
echo "-------------------------------------------"
$PRESCRIBE_BIN -r "$REPO" -t master filter add --name "Filter 1" --exclude "*test*"
$PRESCRIBE_BIN -r "$REPO" -t master filter add --name "Filter 2" --exclude "*.md"
$PRESCRIBE_BIN -r "$REPO" -t master filter add --name "Filter 3" --include "src/*"
$PRESCRIBE_BIN -r "$REPO" -t master filter list
echo "Clearing all filters..."
$PRESCRIBE_BIN -r "$REPO" -t master filter clear
$PRESCRIBE_BIN -r "$REPO" -t master filter list
echo

# Test 12: Session persistence
echo "Test 12: Session persistence"
echo "-------------------------------------------"
$PRESCRIBE_BIN -r "$REPO" -t master filter add --name "Persistent filter" --exclude "*test*" --exclude "*.md"
echo "Session YAML:"
cat .pr-builder/session.yaml
echo

# Test 13: Complex filter with multiple rules
echo "Test 13: Complex filter with multiple rules"
echo "-------------------------------------------"
$PRESCRIBE_BIN -r "$REPO" -t master filter clear
$PRESCRIBE_BIN -r "$REPO" -t master filter add \
    --name "Complex Filter" \
    --description "Exclude tests and docs, include only src" \
    --exclude "*test*" \
    --exclude "*.md" \
    --include "src/**"
$PRESCRIBE_BIN -r "$REPO" -t master filter list
$PRESCRIBE_BIN -r "$REPO" -t master filter show
echo

# Test 14: Generate with filters active
echo "Test 14: Generate with filters"
echo "-------------------------------------------"
$PRESCRIBE_BIN -r "$REPO" -t master filter clear
$PRESCRIBE_BIN -r "$REPO" -t master filter add --name "Exclude tests" --exclude "*test*"
$PRESCRIBE_BIN -r "$REPO" -t master session show
echo "Generating..."
if [ "${PRESCRIBE_RUN_GENERATE:-}" = "1" ]; then
  $PRESCRIBE_BIN -r "$REPO" -t master generate
else
  echo "Skipping generate test (set PRESCRIBE_RUN_GENERATE=1 to enable)"
fi
echo

echo "========================================="
echo "All filter tests completed successfully!"
echo "========================================="
