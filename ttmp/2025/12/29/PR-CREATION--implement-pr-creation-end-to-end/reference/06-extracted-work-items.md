---
Title: Extracted work items for PR creation implementation
Ticket: PR-CREATION
Status: active
Topics:
    - cli
    - git
    - prescribe
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-29T12:06:00.000000000-05:00
WhatFor: ""
WhenToUse: ""
---

# Extracted work items for PR creation implementation

This document contains the concrete, checkable work items extracted from the codebase analysis and clarification answers.

## Work items

### 1. Create new `create` command structure
- Create `prescribe/cmd/prescribe/cmds/create.go` with cobra command structure
- Add command to root command initialization (`InitRootCmd`)
- Command should accept flags: `--use-last`, `--yaml-file`, `--title`, `--body`, `--draft`, `--dry-run`, `--base`
- Output: New `prescribe create` command exists (even if not functional yet)

### 2. Add `--create` flag to existing `generate` command
- Modify `prescribe/cmd/prescribe/cmds/generate.go`
- Add `--create` flag that triggers PR creation after generation
- Reuse generation logic, then call PR creation
- Output: `prescribe generate --create` flag exists and calls creation flow

### 3. Implement GitHub CLI integration (`gh pr create`)
- Create `prescribe/internal/github/github.go` (or similar) with `CreatePR` function
- Shell out to `gh pr create` with appropriate flags (`--title`, `--body`, `--draft`, `--base`)
- Handle `gh` command execution and capture output/errors
- Output: Function that can create a PR via `gh pr create`

### 4. Implement branch pushing before PR creation
- Extend `prescribe/internal/git/git.go` with `PushBranch` function
- Call `git push` before creating PR (when `--create` or `create` command is used)
- Handle push errors gracefully
- Output: Branch is pushed before PR creation

### 5. Implement session data reuse (`--use-last`)
- Read last generated PR data from session file (`.pr-builder/session.yaml`)
- Parse `GeneratedPRData` from session
- Use this data when `--use-last` flag is provided
- Output: `prescribe create --use-last` uses last generated PR data

### 6. Implement YAML file input (`--yaml-file`)
- Add `--yaml-file` flag to `create` command
- Read and parse YAML file containing `GeneratedPRData`
- Use parsed data for PR creation
- Output: `prescribe create --yaml-file path/to/pr.yaml` works

### 7. Implement title/body override flags (`--title`, `--body`)
- Add `--title` and `--body` flags to `create` command
- Override generated title/body when flags are provided
- Support both flags together or individually
- Output: `prescribe create --title "..." --body "..."` overrides generated content

### 8. Implement draft PR support (`--draft`)
- Add `--draft` flag to `create` command
- Pass `--draft` flag to `gh pr create` when flag is set
- Default to false (not draft)
- Output: `prescribe create --draft` creates a draft PR

### 9. Implement dry-run mode (`--dry-run`)
- Add `--dry-run` flag to `create` command
- When set, show what would be created without actually calling `gh pr create`
- Display: title, body, base branch, draft status, etc.
- Output: `prescribe create --dry-run` shows preview without creating PR

### 10. Implement error handling with PR data save
- On PR creation failure, save generated PR data to a file (e.g., `.pr-builder/pr-data-<timestamp>.yaml`)
- Include clear error message indicating where data was saved
- Exit with appropriate error code
- Output: On failure, PR data is saved and user is informed

### 11. Implement base branch handling (`--base`)
- Add `--base` flag to `create` command
- Default to `main` (or detected default branch via `git.GetDefaultBranch()`)
- Pass `--base` flag to `gh pr create`
- Output: `prescribe create --base develop` sets base branch correctly

### 12. Wire up `generate --create` flow
- After successful generation in `generate` command, call PR creation logic
- Use generated PR data for creation
- Handle errors appropriately
- Output: `prescribe generate --create` generates and creates PR in one command

### 13. Add tests for PR creation
- Create `prescribe/internal/github/github_test.go` (or similar)
- Test `gh pr create` command construction with various flags
- Mock `gh` command execution for unit tests
- Output: Test coverage for PR creation logic

### 14. Update documentation
- Update `prescribe/README.md` with `create` command usage
- Document `--use-last`, `--yaml-file`, `--title`, `--body`, `--draft`, `--dry-run`, `--base` flags
- Add examples for common workflows
- Output: README documents PR creation functionality

### 15. Integration test: end-to-end PR creation
- Test full flow: `prescribe generate` → `prescribe create --use-last`
- Test: `prescribe generate --create`
- Test: `prescribe create --yaml-file <file>`
- Verify PR is actually created (or mocked appropriately)
- Output: Integration tests verify end-to-end functionality

## Notes

- Tasks are ordered roughly by dependency (infrastructure first, then features, then tests/docs)
- Each task should be completable independently where possible
- Tasks 1-2 can be done in parallel (separate commands)
- Tasks 3-11 build on the command structure
- Tasks 12-15 are integration and validation

