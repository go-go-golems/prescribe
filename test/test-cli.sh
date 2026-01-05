#!/bin/bash
set -e

# Test script for pr-builder CLI

REPO_DIR="/tmp/pr-builder-test-repo"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PRESCRIBE_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Build a local binary for speed/reproducibility (override with PRESCRIBE_BIN if desired).
PRESCRIBE_BIN_DEFAULT="/tmp/prescribe-$(cd "$PRESCRIBE_ROOT" && git rev-parse --short HEAD 2>/dev/null || echo dev)"
PRESCRIBE_BIN="${PRESCRIBE_BIN:-$PRESCRIBE_BIN_DEFAULT}"
(cd "$PRESCRIBE_ROOT" && GOWORK=off go build -o "$PRESCRIBE_BIN" ./cmd/prescribe)

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

echo "Test 1b: Help tree (subcommands visible)"
echo "===================="

OUT_HELP_ROOT="$($PRESCRIBE_BIN --help)"
echo "$OUT_HELP_ROOT" | grep -Eq "\\bcontext\\b"
echo "$OUT_HELP_ROOT" | grep -Eq "\\bfilter\\b"
echo "$OUT_HELP_ROOT" | grep -Eq "\\bsession\\b"
echo "$OUT_HELP_ROOT" | grep -Eq "\\bfile\\b"
echo "$OUT_HELP_ROOT" | grep -Eq "\\btokens\\b"

OUT_HELP_CONTEXT_GIT="$($PRESCRIBE_BIN context git --help)"
echo "$OUT_HELP_CONTEXT_GIT" | grep -Eq "\\bhistory\\b"
echo "$OUT_HELP_CONTEXT_GIT" | grep -Eq "\\badd\\b"
echo "$OUT_HELP_CONTEXT_GIT" | grep -Eq "\\blist\\b"

OUT_HELP_CONTEXT_GIT_HISTORY="$($PRESCRIBE_BIN context git history --help)"
echo "$OUT_HELP_CONTEXT_GIT_HISTORY" | grep -Eq "\\bshow\\b"
echo "$OUT_HELP_CONTEXT_GIT_HISTORY" | grep -Eq "\\benable\\b"
echo "$OUT_HELP_CONTEXT_GIT_HISTORY" | grep -Eq "\\bdisable\\b"
echo "$OUT_HELP_CONTEXT_GIT_HISTORY" | grep -Eq "\\bset\\b"

OUT_HELP_FILTER_PRESET="$($PRESCRIBE_BIN filter preset --help)"
echo "$OUT_HELP_FILTER_PRESET" | grep -Eq "\\blist\\b"
echo "$OUT_HELP_FILTER_PRESET" | grep -Eq "\\bsave\\b"
echo "$OUT_HELP_FILTER_PRESET" | grep -Eq "\\bapply\\b"

OUT_HELP_SESSION="$($PRESCRIBE_BIN session --help)"
echo "$OUT_HELP_SESSION" | grep -Eq "\\binit\\b"
echo "$OUT_HELP_SESSION" | grep -Eq "\\bshow\\b"
echo "$OUT_HELP_SESSION" | grep -Eq "\\btoken-count\\b"

OUT_HELP_FILE="$($PRESCRIBE_BIN file --help)"
echo "$OUT_HELP_FILE" | grep -Eq "\\btoggle\\b"

OUT_HELP_TOKENS="$($PRESCRIBE_BIN tokens --help)"
echo "$OUT_HELP_TOKENS" | grep -Eq "\\bcount-xml\\b"

echo ""
echo "✓ Help tree works"
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

echo "Test 5: Export generation context (no inference)"
echo "===================="

CTX_DEFAULT="/tmp/prescribe-context-default.xml"
CTX_XML="/tmp/prescribe-context.xml"
CTX_MD="/tmp/prescribe-context.md"
CTX_SIMPLE="/tmp/prescribe-context.simple.txt"
CTX_BEGIN_END="/tmp/prescribe-context.begin-end.txt"
CTX_PLAIN="/tmp/prescribe-context.default.txt"
CTX_OUTPUT_FILE="/tmp/prescribe-context.output-file.xml"

rm -f "$CTX_DEFAULT" "$CTX_XML" "$CTX_MD" "$CTX_SIMPLE" "$CTX_BEGIN_END" "$CTX_PLAIN" "$CTX_OUTPUT_FILE"

# Default separator is xml
$PRESCRIBE_BIN -r "$REPO_DIR" -t master generate --export-context >"$CTX_DEFAULT"
test -s "$CTX_DEFAULT"
grep -q "<prescribe>" "$CTX_DEFAULT"
grep -Eq "<source_commit>[0-9a-f]{7,40}</source_commit>" "$CTX_DEFAULT"
grep -Eq "<target_commit>[0-9a-f]{7,40}</target_commit>" "$CTX_DEFAULT"
grep -q "<git_history>" "$CTX_DEFAULT"

$PRESCRIBE_BIN -r "$REPO_DIR" -t master generate --export-context --separator xml >"$CTX_XML"
test -s "$CTX_XML"
grep -q "<prescribe>" "$CTX_XML"
grep -Eq "<source_commit>[0-9a-f]{7,40}</source_commit>" "$CTX_XML"
grep -Eq "<target_commit>[0-9a-f]{7,40}</target_commit>" "$CTX_XML"
grep -q "<git_history>" "$CTX_XML"

$PRESCRIBE_BIN -r "$REPO_DIR" -t master generate --export-context --separator markdown >"$CTX_MD"
test -s "$CTX_MD"
grep -q "# Prescribe generation context" "$CTX_MD"
grep -q "## Git history" "$CTX_MD"

$PRESCRIBE_BIN -r "$REPO_DIR" -t master generate --export-context --separator simple >"$CTX_SIMPLE"
test -s "$CTX_SIMPLE"
grep -q "START PRESCRIBE CONTEXT" "$CTX_SIMPLE"

$PRESCRIBE_BIN -r "$REPO_DIR" -t master generate --export-context --separator begin-end >"$CTX_BEGIN_END"
test -s "$CTX_BEGIN_END"
grep -q "BEGIN PRESCRIBE CONTEXT" "$CTX_BEGIN_END"

$PRESCRIBE_BIN -r "$REPO_DIR" -t master generate --export-context --separator default >"$CTX_PLAIN"
test -s "$CTX_PLAIN"
grep -q "Prescribe context" "$CTX_PLAIN"

# Verify --output-file also works in export mode
$PRESCRIBE_BIN -r "$REPO_DIR" -t master generate --export-context --separator xml --output-file "$CTX_OUTPUT_FILE"
test -s "$CTX_OUTPUT_FILE"
grep -q "<prescribe>" "$CTX_OUTPUT_FILE"
grep -Eq "<source_commit>[0-9a-f]{7,40}</source_commit>" "$CTX_OUTPUT_FILE"
grep -Eq "<target_commit>[0-9a-f]{7,40}</target_commit>" "$CTX_OUTPUT_FILE"
grep -q "<git_history>" "$CTX_OUTPUT_FILE"

echo ""
echo "✓ Export context works"
echo ""

echo "Test 6: Export rendered LLM payload (no inference)"
echo "===================="

RENDERED_XML="/tmp/prescribe-rendered.xml"
RENDERED_MD="/tmp/prescribe-rendered.md"
RENDERED_OUTPUT_FILE="/tmp/prescribe-rendered.output-file.xml"
rm -f "$RENDERED_XML" "$RENDERED_MD" "$RENDERED_OUTPUT_FILE"

$PRESCRIBE_BIN -r "$REPO_DIR" -t master generate --export-rendered --separator xml >"$RENDERED_XML"
test -s "$RENDERED_XML"
grep -q "<llm_payload>" "$RENDERED_XML"
grep -Eq "<source_commit>[0-9a-f]{7,40}</source_commit>" "$RENDERED_XML"
grep -Eq "<target_commit>[0-9a-f]{7,40}</target_commit>" "$RENDERED_XML"
grep -q "<git_history>" "$RENDERED_XML"

$PRESCRIBE_BIN -r "$REPO_DIR" -t master generate --export-rendered --separator markdown >"$RENDERED_MD"
test -s "$RENDERED_MD"
grep -q "# Prescribe LLM payload (rendered)" "$RENDERED_MD"
grep -q "BEGIN COMMITS" "$RENDERED_MD"
grep -q "author=\\\"Other User\\\"" "$RENDERED_MD"

$PRESCRIBE_BIN -r "$REPO_DIR" -t master generate --export-rendered --separator xml --output-file "$RENDERED_OUTPUT_FILE"
test -s "$RENDERED_OUTPUT_FILE"
grep -q "<llm_payload>" "$RENDERED_OUTPUT_FILE"
grep -Eq "<source_commit>[0-9a-f]{7,40}</source_commit>" "$RENDERED_OUTPUT_FILE"
grep -Eq "<target_commit>[0-9a-f]{7,40}</target_commit>" "$RENDERED_OUTPUT_FILE"
grep -q "<git_history>" "$RENDERED_OUTPUT_FILE"

echo ""
echo "✓ Export rendered payload works"
echo ""

echo "Test 6b: Disable git history removes commits block"
$PRESCRIBE_BIN -r "$REPO_DIR" -t master context git history disable >/dev/null
OUT_NO_COMMITS="$($PRESCRIBE_BIN -r "$REPO_DIR" -t master generate --export-rendered --separator markdown)"
if echo "$OUT_NO_COMMITS" | grep -Fq "BEGIN COMMITS"; then
  echo "Expected commit history to be disabled, but BEGIN COMMITS was present"
  exit 1
fi
$PRESCRIBE_BIN -r "$REPO_DIR" -t master context git history enable >/dev/null
echo "✓ Disabling git history removes commit history"
echo ""

echo "Test 6c: Add explicit git_context item appears in exports"
$PRESCRIBE_BIN -r "$REPO_DIR" -t master context git add commit HEAD >/dev/null
CTX_GIT="/tmp/prescribe-context.git.xml"
rm -f "$CTX_GIT"
$PRESCRIBE_BIN -r "$REPO_DIR" -t master generate --export-context --separator xml >"$CTX_GIT"
test -s "$CTX_GIT"
grep -q "<git_commit" "$CTX_GIT"
OUT_GIT_CTX="$($PRESCRIBE_BIN -r "$REPO_DIR" -t master generate --export-rendered --separator markdown)"
echo "$OUT_GIT_CTX" | grep -Fq "<git_commit"
echo "✓ git_context item appears in export-context and export-rendered"
echo ""

echo "Test 7: Generate with output file (optional)"
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
