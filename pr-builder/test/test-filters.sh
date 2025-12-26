#!/bin/bash
set -e

REPO="/tmp/pr-builder-test-repo"
APP="/home/ubuntu/pr-builder/pr-builder"

echo "========================================="
echo "PR Builder - Filter Functionality Tests"
echo "========================================="
echo

# Ensure test repo exists
if [ ! -d "$REPO" ]; then
    echo "Setting up test repository..."
    /home/ubuntu/pr-builder/test/setup-test-repo.sh
fi

cd "$REPO"

# Clean start
rm -rf .pr-builder
echo "✓ Clean start"
echo

# Test 1: Initialize session
echo "Test 1: Initialize session"
echo "-------------------------------------------"
$APP -r "$REPO" -t master init --save
$APP -r "$REPO" -t master show
echo

# Test 2: List filters (should be empty)
echo "Test 2: List filters (empty)"
echo "-------------------------------------------"
$APP -r "$REPO" -t master list-filters
echo

# Test 3: Test a filter without applying
echo "Test 3: Test filter pattern"
echo "-------------------------------------------"
$APP -r "$REPO" -t master test-filter --name "Test Exclude Tests" --exclude "*test*"
echo

# Test 4: Add exclude filter
echo "Test 4: Add exclude filter"
echo "-------------------------------------------"
$APP -r "$REPO" -t master add-filter --name "Exclude tests" --exclude "*test*"
$APP -r "$REPO" -t master show
echo

# Test 5: List filters
echo "Test 5: List filters"
echo "-------------------------------------------"
$APP -r "$REPO" -t master list-filters
echo

# Test 6: Show filtered files
echo "Test 6: Show filtered files"
echo "-------------------------------------------"
$APP -r "$REPO" -t master show-filtered
echo

# Test 7: Add another filter
echo "Test 7: Add multiple filters"
echo "-------------------------------------------"
$APP -r "$REPO" -t master add-filter --name "Only TypeScript" --include "*.ts"
$APP -r "$REPO" -t master list-filters
echo

# Test 8: Test glob patterns
echo "Test 8: Test various glob patterns"
echo "-------------------------------------------"
echo "Pattern: **/*.test.ts"
$APP -r "$REPO" -t master test-filter --exclude "**/*.test.ts"
echo
echo "Pattern: src/**"
$APP -r "$REPO" -t master test-filter --include "src/**"
echo
echo "Pattern: tests/*"
$APP -r "$REPO" -t master test-filter --exclude "tests/*"
echo

# Test 9: Remove filter by index
echo "Test 9: Remove filter by index"
echo "-------------------------------------------"
$APP -r "$REPO" -t master remove-filter 0
$APP -r "$REPO" -t master list-filters
echo

# Test 10: Remove filter by name
echo "Test 10: Remove filter by name"
echo "-------------------------------------------"
$APP -r "$REPO" -t master add-filter --name "Exclude tests" --exclude "tests/*"
$APP -r "$REPO" -t master remove-filter "Exclude tests"
$APP -r "$REPO" -t master list-filters
echo

# Test 11: Add multiple filters and clear all
echo "Test 11: Clear all filters"
echo "-------------------------------------------"
$APP -r "$REPO" -t master add-filter --name "Filter 1" --exclude "*test*"
$APP -r "$REPO" -t master add-filter --name "Filter 2" --exclude "*.md"
$APP -r "$REPO" -t master add-filter --name "Filter 3" --include "src/*"
$APP -r "$REPO" -t master list-filters
echo "Clearing all filters..."
$APP -r "$REPO" -t master clear-filters
$APP -r "$REPO" -t master list-filters
echo

# Test 12: Session persistence
echo "Test 12: Session persistence"
echo "-------------------------------------------"
$APP -r "$REPO" -t master add-filter --name "Persistent filter" --exclude "*test*" --exclude "*.md"
echo "Session YAML:"
cat .pr-builder/session.yaml
echo

# Test 13: Complex filter with multiple rules
echo "Test 13: Complex filter with multiple rules"
echo "-------------------------------------------"
$APP -r "$REPO" -t master clear-filters
$APP -r "$REPO" -t master add-filter \
    --name "Complex Filter" \
    --description "Exclude tests and docs, include only src" \
    --exclude "*test*" \
    --exclude "*.md" \
    --include "src/**"
$APP -r "$REPO" -t master list-filters
$APP -r "$REPO" -t master show-filtered
echo

# Test 14: Generate with filters active
echo "Test 14: Generate with filters"
echo "-------------------------------------------"
$APP -r "$REPO" -t master clear-filters
$APP -r "$REPO" -t master add-filter --name "Exclude tests" --exclude "*test*"
$APP -r "$REPO" -t master show
echo "Generating..."
$APP -r "$REPO" -t master generate
echo

echo "========================================="
echo "All filter tests completed successfully!"
echo "========================================="
