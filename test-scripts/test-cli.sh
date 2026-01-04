#!/usr/bin/env bash
set -euo pipefail

# Lightweight CLI smoke test: help/version + minimal session init + generate.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

TEST_REPO_DIR="${TEST_REPO_DIR:-/tmp/prescribe-test-repo}"
TARGET_BRANCH="${TARGET_BRANCH:-master}"

prescribe() {
	(
		cd "$REPO_ROOT" && GOWORK=off go run ./cmd/prescribe "$@"
	)
}

echo "=========================================="
echo "prescribe - CLI Smoke Test"
echo "=========================================="
echo ""

echo "Setting up test repository..."
bash "$SCRIPT_DIR/setup-test-repo.sh"
echo ""

echo "Test 1: Help"
prescribe --help >/dev/null
echo "✓ Help works"
echo ""

echo "Test 1b: Help tree (subcommands visible)"
OUT_HELP_ROOT="$(prescribe --help)"
echo "$OUT_HELP_ROOT" | grep -Eq "\\bcontext\\b"
echo "$OUT_HELP_ROOT" | grep -Eq "\\bfilter\\b"
echo "$OUT_HELP_ROOT" | grep -Eq "\\bsession\\b"
echo "$OUT_HELP_ROOT" | grep -Eq "\\bfile\\b"
echo "$OUT_HELP_ROOT" | grep -Eq "\\btokens\\b"

OUT_HELP_CONTEXT_GIT="$(prescribe context git --help)"
echo "$OUT_HELP_CONTEXT_GIT" | grep -Eq "\\bhistory\\b"
echo "$OUT_HELP_CONTEXT_GIT" | grep -Eq "\\badd\\b"
echo "$OUT_HELP_CONTEXT_GIT" | grep -Eq "\\blist\\b"

OUT_HELP_CONTEXT_GIT_HISTORY="$(prescribe context git history --help)"
echo "$OUT_HELP_CONTEXT_GIT_HISTORY" | grep -Eq "\\bshow\\b"
echo "$OUT_HELP_CONTEXT_GIT_HISTORY" | grep -Eq "\\benable\\b"
echo "$OUT_HELP_CONTEXT_GIT_HISTORY" | grep -Eq "\\bdisable\\b"
echo "$OUT_HELP_CONTEXT_GIT_HISTORY" | grep -Eq "\\bset\\b"

OUT_HELP_FILTER_PRESET="$(prescribe filter preset --help)"
echo "$OUT_HELP_FILTER_PRESET" | grep -Eq "\\blist\\b"
echo "$OUT_HELP_FILTER_PRESET" | grep -Eq "\\bsave\\b"
echo "$OUT_HELP_FILTER_PRESET" | grep -Eq "\\bapply\\b"

OUT_HELP_SESSION="$(prescribe session --help)"
echo "$OUT_HELP_SESSION" | grep -Eq "\\binit\\b"
echo "$OUT_HELP_SESSION" | grep -Eq "\\bshow\\b"
echo "$OUT_HELP_SESSION" | grep -Eq "\\btoken-count\\b"

OUT_HELP_FILE="$(prescribe file --help)"
echo "$OUT_HELP_FILE" | grep -Eq "\\btoggle\\b"

OUT_HELP_TOKENS="$(prescribe tokens --help)"
echo "$OUT_HELP_TOKENS" | grep -Eq "\\bcount-xml\\b"

echo "✓ Help tree works"
echo ""

echo "Test 2: Version"
prescribe --version >/dev/null
echo "✓ Version works"
echo ""

echo "Test 3: Session init + show"
rm -rf "$TEST_REPO_DIR/.pr-builder"
prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" session init --save
prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" session show >/dev/null
echo "✓ Session init/show works"
echo ""

echo "Test 4: Generate (export rendered payload, includes git history)"
OUT="$(prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" generate --export-rendered --separator markdown)"
echo "$OUT" | grep -Fq "BEGIN COMMITS"
echo "$OUT" | grep -Fq "feat: enhance authentication"
echo "$OUT" | grep -Fq "author=\"Other User\""
echo "✓ Generate export-rendered includes commit history"
echo ""

echo "Test 5: Disable git history removes commits block"
prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" context git history disable >/dev/null
OUT_NO_COMMITS="$(prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" generate --export-rendered --separator markdown)"
if echo "$OUT_NO_COMMITS" | grep -Fq "BEGIN COMMITS"; then
	echo "Expected commit history to be disabled, but BEGIN COMMITS was present"
	exit 1
fi
echo "✓ Disabling git history removes commit history"
echo ""

echo "Test 6: Add explicit git_context item appears in exports"
prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" context git add commit HEAD >/dev/null
OUT_GIT_CTX="$(prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" generate --export-rendered --separator markdown)"
echo "$OUT_GIT_CTX" | grep -Fq "<git_commit"
echo "✓ git_context item appears in export-rendered"
echo ""

echo "=========================================="
echo "All tests passed ✓"
echo "=========================================="
