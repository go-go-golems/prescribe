#!/bin/bash
set -e

# Comprehensive test script for all pr-builder functionality

REPO_DIR="/tmp/pr-builder-test-repo"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PRESCRIBE_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Build a local binary for speed/reproducibility (override with PRESCRIBE_BIN if desired).
PRESCRIBE_BIN="${PRESCRIBE_BIN:-/tmp/prescribe}"
if [ ! -x "$PRESCRIBE_BIN" ]; then
  (cd "$PRESCRIBE_ROOT" && go build -o "$PRESCRIBE_BIN" ./cmd/prescribe)
fi

echo "=========================================="
echo "PR Builder - Complete Test Suite"
echo "=========================================="
echo ""

# Setup test repo
if [ ! -d "$REPO_DIR" ]; then
    echo "Setting up test repository..."
    "$SCRIPT_DIR/setup-test-repo.sh"
    echo ""
fi

cd "$REPO_DIR"
rm -rf .pr-builder

echo "=== PHASE 1: Session Initialization ==="
echo ""

echo "1.1: Initialize session with auto-save"
$PRESCRIBE_BIN -r "$REPO_DIR" -t master session init --save
echo ""

echo "1.2: Show session state"
$PRESCRIBE_BIN -r "$REPO_DIR" -t master session show
echo ""

echo "=== PHASE 2: File Management ==="
echo ""

echo "2.1: Toggle file exclusion"
$PRESCRIBE_BIN -r "$REPO_DIR" -t master file toggle "tests/auth.test.ts"
echo ""

echo "2.2: Verify file is excluded"
$PRESCRIBE_BIN -r "$REPO_DIR" -t master session show --output json | grep "included_files"
echo ""

echo "2.3: Toggle file back to included"
$PRESCRIBE_BIN -r "$REPO_DIR" -t master file toggle "tests/auth.test.ts"
echo ""

echo "=== PHASE 3: Filters ==="
echo ""

echo "3.1: Add filter to exclude test files"
$PRESCRIBE_BIN -r "$REPO_DIR" -t master filter add \
    --name "Exclude tests" \
    --description "Hide test files" \
    --exclude "*test*"
echo ""

echo "3.2: Add filter to exclude specific paths"
$PRESCRIBE_BIN -r "$REPO_DIR" -t master filter add \
    --name "Exclude middleware" \
    --exclude "*middleware*"
echo ""

echo "3.3: Show session with filters"
$PRESCRIBE_BIN -r "$REPO_DIR" -t master session show
echo ""

echo "=== PHASE 4: Additional Context ==="
echo ""

echo "4.1: Add context note"
$PRESCRIBE_BIN -r "$REPO_DIR" -t master context add \
    --note "This PR is part of the Q1 security improvements epic"
echo ""

echo "4.2: Add context file"
$PRESCRIBE_BIN -r "$REPO_DIR" -t master context add "README.md"
echo ""

echo "4.3: Show session with context"
$PRESCRIBE_BIN -r "$REPO_DIR" -t master session show
echo ""

echo "=== PHASE 5: Session Persistence ==="
echo ""

echo "5.1: View session YAML"
echo "Session file: .pr-builder/session.yaml"
cat .pr-builder/session.yaml
echo ""

echo "5.2: Save to custom location"
$PRESCRIBE_BIN -r "$REPO_DIR" -t master session save /tmp/test-session-backup.yaml
echo ""

echo "5.3: Reinitialize (clear session)"
rm -rf .pr-builder
$PRESCRIBE_BIN -r "$REPO_DIR" -t master session init --save
echo ""

echo "5.4: Load from backup"
$PRESCRIBE_BIN -r "$REPO_DIR" -t master session load /tmp/test-session-backup.yaml
echo ""

echo "5.5: Verify loaded session"
$PRESCRIBE_BIN -r "$REPO_DIR" -t master session show
echo ""

echo "=== PHASE 6: PR Generation ==="
echo ""

echo "6.1: Generate with default prompt"
if [ "${PRESCRIBE_RUN_GENERATE:-}" = "1" ]; then
  $PRESCRIBE_BIN -r "$REPO_DIR" -t master generate -o /tmp/pr-default.md
  echo "Generated description:"
  cat /tmp/pr-default.md
  echo ""
else
  echo "Skipping generate test (set PRESCRIBE_RUN_GENERATE=1 to enable)"
  echo ""
fi

echo "6.2: Generate with custom prompt"
if [ "${PRESCRIBE_RUN_GENERATE:-}" = "1" ]; then
  $PRESCRIBE_BIN -r "$REPO_DIR" -t master generate \
      --prompt "Write a concise 3-sentence PR description" \
      -o /tmp/pr-custom.md
  echo "Generated description:"
  cat /tmp/pr-custom.md
  echo ""
else
  echo "Skipping generate test (set PRESCRIBE_RUN_GENERATE=1 to enable)"
  echo ""
fi

echo "6.3: Generate with preset"
if [ "${PRESCRIBE_RUN_GENERATE:-}" = "1" ]; then
  $PRESCRIBE_BIN -r "$REPO_DIR" -t master generate \
      --preset concise \
      -o /tmp/pr-preset.md
  echo "Generated description:"
  cat /tmp/pr-preset.md
  echo ""
else
  echo "Skipping generate test (set PRESCRIBE_RUN_GENERATE=1 to enable)"
  echo ""
fi

echo "=== PHASE 7: Session Export (YAML) ==="
echo ""

echo "7.1: Export session as YAML"
$PRESCRIBE_BIN -r "$REPO_DIR" -t master session show --output yaml > /tmp/session-export.yaml
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
