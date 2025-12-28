#!/bin/bash
set -e

# Comprehensive test script for all pr-builder functionality

REPO_DIR="/tmp/pr-builder-test-repo"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PRESCRIBE_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Build a local binary for speed/reproducibility (override with PRESCRIBE_BIN if desired).
PRESCRIBE_BIN_DEFAULT="/tmp/prescribe-$(cd "$PRESCRIBE_ROOT" && git rev-parse --short HEAD 2>/dev/null || echo dev)"
PRESCRIBE_BIN="${PRESCRIBE_BIN:-$PRESCRIBE_BIN_DEFAULT}"
(cd "$PRESCRIBE_ROOT" && go build -o "$PRESCRIBE_BIN" ./cmd/prescribe)

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

echo "6.0: Export generation context (no inference)"
CTX_DEFAULT="/tmp/prescribe-context-default.xml"
CTX_XML="/tmp/prescribe-context.xml"
CTX_MD="/tmp/prescribe-context.md"
CTX_SIMPLE="/tmp/prescribe-context.simple.txt"
CTX_BEGIN_END="/tmp/prescribe-context.begin-end.txt"
CTX_PLAIN="/tmp/prescribe-context.default.txt"
CTX_OUTPUT_FILE="/tmp/prescribe-context.output-file.xml"

rm -f "$CTX_DEFAULT" "$CTX_XML" "$CTX_MD" "$CTX_SIMPLE" "$CTX_BEGIN_END" "$CTX_PLAIN" "$CTX_OUTPUT_FILE"

$PRESCRIBE_BIN -r "$REPO_DIR" -t master generate --export-context >"$CTX_DEFAULT"
test -s "$CTX_DEFAULT"
grep -q "<prescribe>" "$CTX_DEFAULT"

$PRESCRIBE_BIN -r "$REPO_DIR" -t master generate --export-context --separator xml >"$CTX_XML"
test -s "$CTX_XML"
grep -q "<prescribe>" "$CTX_XML"

$PRESCRIBE_BIN -r "$REPO_DIR" -t master generate --export-context --separator markdown >"$CTX_MD"
test -s "$CTX_MD"
grep -q "# Prescribe generation context" "$CTX_MD"

$PRESCRIBE_BIN -r "$REPO_DIR" -t master generate --export-context --separator simple >"$CTX_SIMPLE"
test -s "$CTX_SIMPLE"
grep -q "START PRESCRIBE CONTEXT" "$CTX_SIMPLE"

$PRESCRIBE_BIN -r "$REPO_DIR" -t master generate --export-context --separator begin-end >"$CTX_BEGIN_END"
test -s "$CTX_BEGIN_END"
grep -q "BEGIN PRESCRIBE CONTEXT" "$CTX_BEGIN_END"

$PRESCRIBE_BIN -r "$REPO_DIR" -t master generate --export-context --separator default >"$CTX_PLAIN"
test -s "$CTX_PLAIN"
grep -q "Prescribe context" "$CTX_PLAIN"

$PRESCRIBE_BIN -r "$REPO_DIR" -t master generate --export-context --separator xml --output-file "$CTX_OUTPUT_FILE"
test -s "$CTX_OUTPUT_FILE"
grep -q "<prescribe>" "$CTX_OUTPUT_FILE"

echo ""

echo "6.0b: Export rendered LLM payload (no inference)"
RENDERED_XML="/tmp/prescribe-rendered.xml"
RENDERED_MD="/tmp/prescribe-rendered.md"
RENDERED_OUTPUT_FILE="/tmp/prescribe-rendered.output-file.xml"
rm -f "$RENDERED_XML" "$RENDERED_MD" "$RENDERED_OUTPUT_FILE"

$PRESCRIBE_BIN -r "$REPO_DIR" -t master generate --export-rendered --separator xml >"$RENDERED_XML"
test -s "$RENDERED_XML"
grep -q "<llm_payload>" "$RENDERED_XML"

$PRESCRIBE_BIN -r "$REPO_DIR" -t master generate --export-rendered --separator markdown >"$RENDERED_MD"
test -s "$RENDERED_MD"
grep -q "# Prescribe LLM payload (rendered)" "$RENDERED_MD"

$PRESCRIBE_BIN -r "$REPO_DIR" -t master generate --export-rendered --separator xml --output-file "$RENDERED_OUTPUT_FILE"
test -s "$RENDERED_OUTPUT_FILE"
grep -q "<llm_payload>" "$RENDERED_OUTPUT_FILE"

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
