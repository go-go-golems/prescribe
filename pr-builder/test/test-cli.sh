#!/bin/bash
set -e

# Test script for pr-builder CLI

REPO_DIR="/tmp/pr-builder-test-repo"
PR_BUILDER="/home/ubuntu/pr-builder/pr-builder"

echo "=========================================="
echo "PR Builder CLI Test Suite"
echo "=========================================="
echo ""

# Ensure test repo exists
if [ ! -d "$REPO_DIR" ]; then
    echo "Test repository not found. Running setup script..."
    /home/ubuntu/pr-builder/test/setup-test-repo.sh
fi

cd "$REPO_DIR"

echo "Test 1: Show help"
echo "===================="
$PR_BUILDER --help
echo ""
echo "✓ Help command works"
echo ""

echo "Test 2: Show version"
echo "===================="
$PR_BUILDER --version
echo ""
echo "✓ Version command works"
echo ""

echo "Test 3: Status command"
echo "===================="
$PR_BUILDER -r "$REPO_DIR" -t master status
echo ""
echo "✓ Status command works"
echo ""

echo "Test 4: List files (visible only)"
echo "===================="
$PR_BUILDER -r "$REPO_DIR" -t master list-files
echo ""
echo "✓ List files command works"
echo ""

echo "Test 5: List all files"
echo "===================="
$PR_BUILDER -r "$REPO_DIR" -t master list-files --all
echo ""
echo "✓ List all files command works"
echo ""

echo "Test 6: Generate PR description (default prompt)"
echo "===================="
$PR_BUILDER -r "$REPO_DIR" -t master generate
echo ""
echo "✓ Generate command works"
echo ""

echo "Test 7: Generate with output file"
echo "===================="
$PR_BUILDER -r "$REPO_DIR" -t master generate -o /tmp/pr-description.md
echo "Generated description saved to /tmp/pr-description.md"
cat /tmp/pr-description.md
echo ""
echo "✓ Generate with output file works"
echo ""

echo "Test 8: Generate with custom prompt"
echo "===================="
$PR_BUILDER -r "$REPO_DIR" -t master generate --prompt "Write a concise PR description in bullet points"
echo ""
echo "✓ Generate with custom prompt works"
echo ""

echo "Test 9: Generate with preset"
echo "===================="
$PR_BUILDER -r "$REPO_DIR" -t master generate --preset concise
echo ""
echo "✓ Generate with preset works"
echo ""

echo "=========================================="
echo "All tests passed! ✓"
echo "=========================================="
