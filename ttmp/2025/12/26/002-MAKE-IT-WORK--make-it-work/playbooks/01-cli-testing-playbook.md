---
Title: CLI Testing Playbook for Prescribe
Ticket: 002-MAKE-IT-WORK
Status: active
Topics:
    - testing
    - cli
    - playbook
DocType: playbook
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Comprehensive testing guide for prescribe CLI commands using hierarchical verb structure
LastUpdated: 2025-12-26T19:30:00.000000000-05:00
WhatFor: Testing prescribe functionality for first-time users
WhenToUse: When validating CLI commands work correctly after changes
---

# CLI Testing Playbook for Prescribe

## Overview

This playbook provides step-by-step instructions for testing all `prescribe` CLI functionality. It's designed for users who have never used the program before, walking through each command group and operation systematically.

## Prerequisites

### Build the Binary

```bash
cd /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe
go build -o ./dist/prescribe ./cmd/prescribe
```

### Set Up Test Repository

The test scripts include a setup script that creates a realistic test repository:

```bash
cd /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe
./test-scripts/setup-test-repo.sh
```

This creates `/tmp/pr-builder-test-repo` with:
- A git repository initialized
- A `main` branch with initial files
- A `feature/user-auth` branch with changes
- Realistic TypeScript files and tests

### Command Reference

The binary can be run as:
- `./dist/prescribe` (if built)
- `go run cmd/prescribe/main.go` (direct execution)

For convenience in this playbook, we'll use `PRESCRIBE` variable:

```bash
export PRESCRIBE="./dist/prescribe"  # or "go run cmd/prescribe/main.go"
export REPO="/tmp/pr-builder-test-repo"
```

## Command Structure

Prescribe uses **hierarchical commands** organized into groups:

- **Root commands**: `generate`, `tui`
- **Command groups**: `session`, `filter`, `file`, `context`

### Global Flags

All commands support:
- `-r, --repo string`: Path to git repository (default: ".")
- `-t, --target string`: Target branch (default: main or master)

## Phase 1: Session Management

Session management is the foundation - you must initialize a session before using other features.

### 1.1: Initialize a Session

```bash
cd "$REPO"
rm -rf .pr-builder  # Clean start

# Initialize session (creates in-memory session)
$PRESCRIBE -r "$REPO" -t master session init

# Initialize and auto-save to .pr-builder/session.yaml
$PRESCRIBE -r "$REPO" -t master session init --save

# Initialize with custom output path
$PRESCRIBE -r "$REPO" -t master session init --save --output /tmp/my-session.yaml
```

**Expected behavior:**
- Creates a session from current git state
- Detects changed files between `master` and current branch
- `--save` creates `.pr-builder/session.yaml` (or custom path)

**Validation:**
```bash
# Check session file exists
ls -la .pr-builder/session.yaml

# View session contents
cat .pr-builder/session.yaml
```

### 1.2: Show Session State

```bash
# Human-readable format (default)
$PRESCRIBE -r "$REPO" -t master session show

# YAML format
$PRESCRIBE -r "$REPO" -t master session show --yaml
```

**Expected output includes:**
- Source and target branches
- List of changed files with inclusion status
- Active filters (if any)
- Additional context (if any)
- Prompt configuration

### 1.3: Save Session

```bash
# Save to default location (.pr-builder/session.yaml)
$PRESCRIBE -r "$REPO" -t master session save

# Save to custom location
$PRESCRIBE -r "$REPO" -t master session save /tmp/backup-session.yaml
```

### 1.4: Load Session

```bash
# Load from default location
$PRESCRIBE -r "$REPO" -t master session load

# Load from custom location
$PRESCRIBE -r "$REPO" -t master session load /tmp/backup-session.yaml
```

**Test round-trip:**
```bash
# 1. Initialize and configure
$PRESCRIBE -r "$REPO" -t master session init --save

# 2. Make changes (add filter, toggle files, etc.)
# ... (see phases below)

# 3. Save to backup
$PRESCRIBE -r "$REPO" -t master session save /tmp/test-backup.yaml

# 4. Reset session
rm -rf .pr-builder
$PRESCRIBE -r "$REPO" -t master session init --save

# 5. Load backup
$PRESCRIBE -r "$REPO" -t master session load /tmp/test-backup.yaml

# 6. Verify state matches
$PRESCRIBE -r "$REPO" -t master session show
```

## Phase 2: File Management

File operations let you control which changed files are included in PR generation.

### 2.1: Toggle File Inclusion

```bash
# Initialize session first
$PRESCRIBE -r "$REPO" -t master session init --save

# Toggle a file (include/exclude)
$PRESCRIBE -r "$REPO" -t master file toggle "src/auth/login.ts"

# Verify change
$PRESCRIBE -r "$REPO" -t master session show | grep "login.ts"
```

**Expected behavior:**
- First toggle: excludes file (if included) or includes file (if excluded)
- Subsequent toggles: reverses inclusion state
- Changes are auto-saved if session was initialized with `--save`

**Test multiple files:**
```bash
$PRESCRIBE -r "$REPO" -t master file toggle "src/auth/login.ts"
$PRESCRIBE -r "$REPO" -t master file toggle "src/auth/middleware.ts"
$PRESCRIBE -r "$REPO" -t master file toggle "tests/auth.test.ts"

# View final state
$PRESCRIBE -r "$REPO" -t master session show
```

## Phase 3: Filter Management

Filters use glob patterns to include/exclude files automatically.

### 3.1: List Filters

```bash
# List all active filters
$PRESCRIBE -r "$REPO" -t master filter list
```

**Expected:** Empty list initially, or list of active filters with names and patterns.

### 3.2: Test Filter Pattern (Without Applying)

```bash
# Test what files would match a pattern
# Note: *test* matches filename only, not full path
$PRESCRIBE -r "$REPO" -t master filter test --exclude "*test*"

# Test directory pattern (matches full path)
$PRESCRIBE -r "$REPO" -t master filter test --exclude "tests/*"

# Test recursive pattern (matches anywhere in path)
$PRESCRIBE -r "$REPO" -t master filter test --exclude "**/*test*"

# Test include pattern
$PRESCRIBE -r "$REPO" -t master filter test --include "src/**"

# Test complex pattern
$PRESCRIBE -r "$REPO" -t master filter test --exclude "*test*" --include "src/**"
```

**Expected:** Shows which files would be affected without modifying session.

**Important:** Glob patterns match against:
- Filename only: `*test*` matches `auth.test.ts` but NOT `tests/auth.ts`
- Full path: `tests/*` matches `tests/auth.test.ts`
- Recursive: `**/*test*` matches any file with "test" anywhere in path

### 3.3: Add Filter

```bash
# Simple exclude filter (matches filename only)
$PRESCRIBE -r "$REPO" -t master filter add \
    --name "Exclude test files" \
    --exclude "*test*"

# Exclude directory (matches full path)
$PRESCRIBE -r "$REPO" -t master filter add \
    --name "Exclude tests directory" \
    --exclude "tests/*"

# Recursive pattern (matches anywhere in path)
$PRESCRIBE -r "$REPO" -t master filter add \
    --name "Exclude any test files" \
    --exclude "**/*test*"

# Filter with description
$PRESCRIBE -r "$REPO" -t master filter add \
    --name "Exclude middleware" \
    --description "Hide middleware files from context" \
    --exclude "*middleware*"

# Include-only filter
$PRESCRIBE -r "$REPO" -t master filter add \
    --name "Only TypeScript" \
    --include "*.ts"

# Complex filter (multiple patterns)
$PRESCRIBE -r "$REPO" -t master filter add \
    --name "Backend only" \
    --description "Only backend code, exclude tests and docs" \
    --include "src/**" \
    --exclude "*test*" \
    --exclude "*.md"
```

**Verify filter applied:**
```bash
$PRESCRIBE -r "$REPO" -t master filter list
$PRESCRIBE -r "$REPO" -t master session show
```

### 3.4: Show Filtered Files

```bash
# Show which files are currently filtered out
$PRESCRIBE -r "$REPO" -t master filter show
```

**Expected:** Lists files that match exclude patterns or don't match include patterns.

### 3.5: Remove Filter

```bash
# Remove by index (from filter list)
$PRESCRIBE -r "$REPO" -t master filter remove 0

# Remove by name
$PRESCRIBE -r "$REPO" -t master filter remove "Exclude tests"
```

**Verify removal:**
```bash
$PRESCRIBE -r "$REPO" -t master filter list
```

### 3.6: Clear All Filters

```bash
# Remove all filters at once
$PRESCRIBE -r "$REPO" -t master filter clear

# Verify
$PRESCRIBE -r "$REPO" -t master filter list
```

## Phase 4: Additional Context

Add files or notes that aren't part of the diff but provide context for PR generation.

### 4.1: Add Context File

```bash
# Add a file as additional context
$PRESCRIBE -r "$REPO" -t master context add "README.md"

# Add multiple files (run command once per file)
$PRESCRIBE -r "$REPO" -t master context add "README.md"
$PRESCRIBE -r "$REPO" -t master context add "CONTRIBUTING.md"
```

**Verify:**
```bash
$PRESCRIBE -r "$REPO" -t master session show
```

### 4.2: Add Context Note

```bash
# Add a text note as context
$PRESCRIBE -r "$REPO" -t master context add \
    --note "This PR is part of the Q1 security improvements epic"

# Add multiple notes (run command once per note)
$PRESCRIBE -r "$REPO" -t master context add --note "Related to issue #123"
$PRESCRIBE -r "$REPO" -t master context add --note "Requires database migration"
```

**Note:** `context add` accepts a single note per invocation (run it multiple times to add multiple notes).

**Verify:**
```bash
$PRESCRIBE -r "$REPO" -t master session show --yaml | grep -A 10 "context:"
```

## Phase 5: Generate PR Description

Generate AI-powered PR descriptions from the configured session.

### 5.0: Export generation context payload (no inference)

This prints the **exact text blob** that would be sent to the model (prompt + included files + additional context),
without running inference.

```bash
# Export to stdout (default separator is xml)
$PRESCRIBE -r "$REPO" -t master generate --export-context

# Export with explicit separator selection
$PRESCRIBE -r "$REPO" -t master generate --export-context --separator xml
$PRESCRIBE -r "$REPO" -t master generate --export-context --separator markdown
$PRESCRIBE -r "$REPO" -t master generate --export-context --separator simple
$PRESCRIBE -r "$REPO" -t master generate --export-context --separator begin-end
$PRESCRIBE -r "$REPO" -t master generate --export-context --separator default

# Export to file
$PRESCRIBE -r "$REPO" -t master generate --export-context --separator xml --output-file /tmp/prescribe-context.xml
```

### 5.0b: Export rendered LLM payload (no inference)

This prints the **rendered** `(system,user)` payload that would be sent to the model (no inference).

```bash
# Export to stdout
$PRESCRIBE -r "$REPO" -t master generate --export-rendered

# Export with explicit separator selection
$PRESCRIBE -r "$REPO" -t master generate --export-rendered --separator xml
$PRESCRIBE -r "$REPO" -t master generate --export-rendered --separator markdown

# Export to file
$PRESCRIBE -r "$REPO" -t master generate --export-rendered --separator xml --output-file /tmp/prescribe-rendered.xml
```

### 5.1: Generate with Default Settings

```bash
# Generate and print to stdout
$PRESCRIBE -r "$REPO" -t master generate

# Generate and save to file
$PRESCRIBE -r "$REPO" -t master generate --output-file /tmp/pr-description.md

# View generated file
cat /tmp/pr-description.md
```

### 5.2: Generate with Custom Prompt

```bash
$PRESCRIBE -r "$REPO" -t master generate \
    --prompt "Write a concise 3-sentence PR description focusing on security improvements"
```

### 5.3: Generate with Preset

```bash
# Use a preset prompt template
$PRESCRIBE -r "$REPO" -t master generate --preset concise

# Other presets (check available presets)
$PRESCRIBE -r "$REPO" -t master generate --preset detailed
$PRESCRIBE -r "$REPO" -t master generate --preset technical
```

### 5.4: Generate with Session File

```bash
# Load session from file and generate
$PRESCRIBE -r "$REPO" -t master generate --load-session /tmp/my-session.yaml
```

**Note:** This should work even without an active session in `.pr-builder/`.

## Phase 6: Interactive TUI

Launch the Terminal User Interface for interactive PR building.

### 6.1: Launch TUI

```bash
# Basic TUI launch
$PRESCRIBE -r "$REPO" -t master tui
```

**Expected behavior:**
- Opens full-screen TUI
- Shows file list with keyboard navigation
- Displays current session state

**Keyboard shortcuts (from README):**
- `↑/↓` or `j/k`: Navigate file list
- `Space`: Toggle file inclusion
- `g`: Generate PR description
- `Esc`: Go back (from result screen)
- `q`: Quit

### 6.2: TUI Workflow Test

1. Launch TUI: `$PRESCRIBE -r "$REPO" -t master tui`
2. Navigate files with arrow keys or `j/k`
3. Toggle file inclusion with `Space`
4. Press `g` to generate
5. View generated description
6. Press `Esc` to return to file list
7. Press `q` to quit

**Verify changes persisted:**
```bash
# After quitting TUI, check session
$PRESCRIBE -r "$REPO" -t master session show
```

## Phase 7: Integration Tests

Test complete workflows combining multiple commands.

### 7.1: Complete PR Workflow

```bash
# 1. Initialize session
cd "$REPO"
rm -rf .pr-builder
$PRESCRIBE -r "$REPO" -t master session init --save

# 2. Add filter to exclude tests
$PRESCRIBE -r "$REPO" -t master filter add \
    --name "Exclude tests" \
    --exclude "*test*"

# 3. Toggle specific file
$PRESCRIBE -r "$REPO" -t master file toggle "src/auth/middleware.ts"

# 4. Add context
$PRESCRIBE -r "$REPO" -t master context add \
    --note "This PR enhances authentication security"

# 5. Review session state
$PRESCRIBE -r "$REPO" -t master session show

# 6. Generate PR description
$PRESCRIBE -r "$REPO" -t master generate \
    --output /tmp/final-pr.md \
    --preset concise

# 7. View result
cat /tmp/final-pr.md
```

### 7.2: Session Persistence Workflow

```bash
# 1. Create and configure session
$PRESCRIBE -r "$REPO" -t master session init --save
$PRESCRIBE -r "$REPO" -t master filter add --name "Test filter" --exclude "*test*"
$PRESCRIBE -r "$REPO" -t master context add --note "Test note"

# 2. Save to backup
$PRESCRIBE -r "$REPO" -t master session save /tmp/backup.yaml

# 3. Clear session directory
rm -rf .pr-builder

# 4. Load backup
$PRESCRIBE -r "$REPO" -t master session load /tmp/backup.yaml

# 5. Verify state restored
$PRESCRIBE -r "$REPO" -t master session show
$PRESCRIBE -r "$REPO" -t master filter list
```

### 7.3: Filter Testing Workflow

```bash
# 1. Initialize
$PRESCRIBE -r "$REPO" -t master session init --save

# 2. Test pattern before applying
$PRESCRIBE -r "$REPO" -t master filter test --exclude "*test*"

# 3. Apply filter
$PRESCRIBE -r "$REPO" -t master filter add \
    --name "Exclude tests" \
    --exclude "*test*"

# 4. Verify filtered files
$PRESCRIBE -r "$REPO" -t master filter show

# 5. Check session state
$PRESCRIBE -r "$REPO" -t master session show

# 6. Remove filter
$PRESCRIBE -r "$REPO" -t master filter remove "Exclude tests"

# 7. Verify removal
$PRESCRIBE -r "$REPO" -t master session show
```

## Phase 8: Error Handling Tests

Test error conditions and edge cases.

### 8.1: Invalid Repository Path

```bash
# Should show helpful error
$PRESCRIBE -r "/nonexistent/path" -t master session init
```

### 8.2: Invalid Target Branch

```bash
# Should handle gracefully
$PRESCRIBE -r "$REPO" -t nonexistent-branch session init
```

### 8.3: Operations Without Session

```bash
# Clean start
rm -rf .pr-builder

# Try operations without init
$PRESCRIBE -r "$REPO" -t master session show
$PRESCRIBE -r "$REPO" -t master filter list
$PRESCRIBE -r "$REPO" -t master file toggle "some-file.ts"
```

**Expected:** Should either auto-initialize or show helpful error message.

### 8.4: Invalid File Paths

```bash
# Toggle non-existent file
$PRESCRIBE -r "$REPO" -t master file toggle "nonexistent/file.ts"

# Add context with invalid file
$PRESCRIBE -r "$REPO" -t master context add "nonexistent.md"
```

### 8.5: Invalid Filter Patterns

```bash
# Test invalid glob pattern (if validation exists)
$PRESCRIBE -r "$REPO" -t master filter add \
    --name "Invalid" \
    --exclude "[invalid-glob"
```

## Validation Checklist

After completing all phases, verify:

- [ ] Session init creates valid session file
- [ ] Session show displays correct information
- [ ] Session save/load round-trip works
- [ ] File toggle changes inclusion state
- [ ] Filter add/list/remove/clear all work
- [ ] Filter test shows correct matches
- [ ] Context add works for files and notes
- [ ] Generate produces output (even if mock)
- [ ] TUI launches and responds to keyboard
- [ ] All commands respect `--repo` and `--target` flags
- [ ] Error messages are helpful

## Common Issues and Troubleshooting

### Issue: "Session not found" errors

**Solution:** Ensure you've run `session init --save` first, or use `--session` flag with `generate`.

### Issue: Filters not applying

**Solution:** Check filter patterns match your file paths. Use `filter test` to preview matches.

### Issue: TUI not responding

**Solution:** Ensure terminal supports full-screen mode. Try resizing terminal window.

### Issue: Generated output is empty or mock

**Solution:** This is expected if API is mocked. Check `internal/api/api.go` for mock implementation.

## Next Steps

After completing this playbook:

1. Document any bugs or unexpected behavior
2. Note missing features or unclear error messages
3. Test with real git repositories (not just test repo)
4. Verify session file format matches expected schema
5. Test with different branch names and git states

## Related Documentation

- `analysis/01-prescribe-tui-structure-models-messages-and-control-flow.md` - TUI architecture
- `analysis/02-core-architecture-controller-domain-model-git-session-api-subsystems.md` - Core systems
- `reference/01-go-go-golems-bubbletea-application-guide.md` - Bubbletea patterns
- `test-scripts/` - Automated test scripts (may need updating for hierarchical commands)

