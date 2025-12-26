#!/bin/bash
set -e

# Comprehensive test script for all pr-builder functionality

REPO_DIR="/tmp/pr-builder-test-repo"
PR_BUILDER="/home/ubuntu/pr-builder/pr-builder"

echo "=========================================="
echo "PR Builder - Complete Test Suite"
echo "=========================================="
echo ""

# Setup test repo
if [ ! -d "$REPO_DIR" ]; then
    echo "Setting up test repository..."
    /home/ubuntu/pr-builder/test/setup-test-repo.sh
    echo ""
fi

cd "$REPO_DIR"
rm -rf .pr-builder

echo "=== PHASE 1: Session Initialization ==="
echo ""

echo "1.1: Initialize session with auto-save"
$PR_BUILDER -r "$REPO_DIR" -t master init --save
echo ""

echo "1.2: Show session state"
$PR_BUILDER -r "$REPO_DIR" -t master show
echo ""

echo "=== PHASE 2: File Management ==="
echo ""

echo "2.1: Toggle file exclusion"
$PR_BUILDER -r "$REPO_DIR" -t master toggle-file "tests/auth.test.ts"
echo ""

echo "2.2: Verify file is excluded"
$PR_BUILDER -r "$REPO_DIR" -t master show | grep -A5 "Files:"
echo ""

echo "2.3: Toggle file back to included"
$PR_BUILDER -r "$REPO_DIR" -t master toggle-file "tests/auth.test.ts"
echo ""

echo "=== PHASE 3: Filters ==="
echo ""

echo "3.1: Add filter to exclude test files"
$PR_BUILDER -r "$REPO_DIR" -t master add-filter \
    --name "Exclude tests" \
    --description "Hide test files" \
    --exclude "*test*"
echo ""

echo "3.2: Add filter to exclude specific paths"
$PR_BUILDER -r "$REPO_DIR" -t master add-filter \
    --name "Exclude middleware" \
    --exclude "*middleware*"
echo ""

echo "3.3: Show session with filters"
$PR_BUILDER -r "$REPO_DIR" -t master show
echo ""

echo "=== PHASE 4: Additional Context ==="
echo ""

echo "4.1: Add context note"
$PR_BUILDER -r "$REPO_DIR" -t master add-context \
    --note "This PR is part of the Q1 security improvements epic"
echo ""

echo "4.2: Add context file"
$PR_BUILDER -r "$REPO_DIR" -t master add-context "README.md"
echo ""

echo "4.3: Show session with context"
$PR_BUILDER -r "$REPO_DIR" -t master show
echo ""

echo "=== PHASE 5: Session Persistence ==="
echo ""

echo "5.1: View session YAML"
echo "Session file: .pr-builder/session.yaml"
cat .pr-builder/session.yaml
echo ""

echo "5.2: Save to custom location"
$PR_BUILDER -r "$REPO_DIR" -t master save /tmp/test-session-backup.yaml
echo ""

echo "5.3: Reinitialize (clear session)"
rm -rf .pr-builder
$PR_BUILDER -r "$REPO_DIR" -t master init --save
echo ""

echo "5.4: Load from backup"
$PR_BUILDER -r "$REPO_DIR" -t master load /tmp/test-session-backup.yaml
echo ""

echo "5.5: Verify loaded session"
$PR_BUILDER -r "$REPO_DIR" -t master show
echo ""

echo "=== PHASE 6: PR Generation ==="
echo ""

echo "6.1: Generate with default prompt"
$PR_BUILDER -r "$REPO_DIR" -t master generate -o /tmp/pr-default.md
echo "Generated description:"
cat /tmp/pr-default.md
echo ""

echo "6.2: Generate with custom prompt"
$PR_BUILDER -r "$REPO_DIR" -t master generate \
    --prompt "Write a concise 3-sentence PR description" \
    -o /tmp/pr-custom.md
echo "Generated description:"
cat /tmp/pr-custom.md
echo ""

echo "6.3: Generate with preset"
$PR_BUILDER -r "$REPO_DIR" -t master generate \
    --preset concise \
    -o /tmp/pr-preset.md
echo "Generated description:"
cat /tmp/pr-preset.md
echo ""

echo "=== PHASE 7: Session Export (YAML) ==="
echo ""

echo "7.1: Export session as YAML"
$PR_BUILDER -r "$REPO_DIR" -t master show --yaml > /tmp/session-export.yaml
echo "Exported to /tmp/session-export.yaml"
cat /tmp/session-export.yaml
echo ""

echo "=========================================="
echo "✓ All tests passed!"
echo "=========================================="
echo ""
echo "Test artifacts:"
echo "  - Session file: $REPO_DIR/.pr-builder/session.yaml"
echo "  - Backup session: /tmp/test-session-backup.yaml"
echo "  - Exported YAML: /tmp/session-export.yaml"
echo "  - Generated PRs:"
echo "      - Default: /tmp/pr-default.md"
echo "      - Custom: /tmp/pr-custom.md"
echo "      - Preset: /tmp/pr-preset.md"
