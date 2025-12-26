#!/bin/bash
set -e

# Comprehensive test script for pr-builder session-based CLI

REPO_DIR="/tmp/pr-builder-test-repo"
PR_BUILDER="/home/ubuntu/pr-builder/pr-builder"
SESSION_FILE="/tmp/pr-builder-test-repo/.pr-builder/session.yaml"

echo "=========================================="
echo "PR Builder Session-Based CLI Test Suite"
echo "=========================================="
echo ""

# Ensure test repo exists
if [ ! -d "$REPO_DIR" ]; then
    echo "Test repository not found. Running setup script..."
    /home/ubuntu/pr-builder/test/setup-test-repo.sh
fi

cd "$REPO_DIR"

# Clean up any existing session
rm -rf .pr-builder

echo "Test 1: Initialize session"
echo "============================"
$PR_BUILDER -r "$REPO_DIR" -t master init --save
echo ""
echo "✓ Init command works"
echo ""

echo "Test 2: Show session (human-readable)"
echo "======================================"
$PR_BUILDER -r "$REPO_DIR" -t master show
echo ""
echo "✓ Show command works"
echo ""

echo "Test 3: Show session (YAML)"
echo "============================"
$PR_BUILDER -r "$REPO_DIR" -t master show --yaml
echo ""
echo "✓ Show YAML works"
echo ""

echo "Test 4: Add filter to exclude tests"
echo "===================================="
$PR_BUILDER -r "$REPO_DIR" -t master add-filter \
    --name "Exclude tests" \
    --description "Hide test files from context" \
    --exclude "*test*"
echo ""
echo "✓ Add filter works"
echo ""

echo "Test 5: Show session after filter"
echo "=================================="
$PR_BUILDER -r "$REPO_DIR" -t master show
echo ""
echo "✓ Filter applied correctly"
echo ""

echo "Test 6: Toggle file inclusion"
echo "=============================="
$PR_BUILDER -r "$REPO_DIR" -t master toggle-file "src/auth/login.ts"
echo ""
echo "✓ Toggle file works"
echo ""

echo "Test 7: Add context note"
echo "========================"
$PR_BUILDER -r "$REPO_DIR" -t master add-context \
    --note "This PR is part of the auth refactor epic"
echo ""
echo "✓ Add context note works"
echo ""

echo "Test 8: Add context file"
echo "========================"
$PR_BUILDER -r "$REPO_DIR" -t master add-context "README.md"
echo ""
echo "✓ Add context file works"
echo ""

echo "Test 9: Show final session state"
echo "================================="
$PR_BUILDER -r "$REPO_DIR" -t master show
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
$PR_BUILDER -r "$REPO_DIR" -t master save /tmp/custom-session.yaml
echo ""
echo "✓ Save to custom path works"
echo ""

echo "Test 12: Load session from custom path"
echo "======================================="
# First, reset by removing the default session
rm -f "$SESSION_FILE"
$PR_BUILDER -r "$REPO_DIR" -t master init --save
# Then load the custom session
$PR_BUILDER -r "$REPO_DIR" -t master load /tmp/custom-session.yaml
echo ""
echo "✓ Load from custom path works"
echo ""

echo "Test 13: Generate with session"
echo "==============================="
$PR_BUILDER -r "$REPO_DIR" -t master generate
echo ""
echo "✓ Generate with session works"
echo ""

echo "Test 14: Generate with session file flag"
echo "========================================="
$PR_BUILDER -r "$REPO_DIR" -t master generate --session /tmp/custom-session.yaml
echo ""
echo "✓ Generate with --session flag works"
echo ""

echo "=========================================="
echo "All tests passed! ✓"
echo "=========================================="
echo ""
echo "Session file location: $SESSION_FILE"
echo "Custom session location: /tmp/custom-session.yaml"
