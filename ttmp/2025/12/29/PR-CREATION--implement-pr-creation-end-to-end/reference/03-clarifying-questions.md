---
Title: Clarifying questions for PR creation implementation
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
LastUpdated: 2025-12-29T11:56:00.000000000-05:00
WhatFor: ""
WhenToUse: ""
---

# Clarifying questions for PR creation implementation

This document contains the minimal set of clarifying questions needed to lock down the intended PR-creation functionality before task planning. Each question includes:
- A short title
- Why it matters
- Recommended plz-confirm widget type

## Question 1: Integration approach

**Title**: Use GitHub CLI (`gh`) or GitHub API?

**Why it matters**: The analysis doc identifies two options: shell out to `gh pr create` (simpler, matches README suggestion) or use GitHub API directly (more control, requires auth handling). This decision affects dependencies, error handling, and authentication flow.

**Recommended widget**: `select`
- Option 1: Shell out to GitHub CLI (`gh pr create`)
- Option 2: Use GitHub API directly (new dependency)
- Option 3: Support both (CLI as default, API as fallback/option)

## Question 2: CLI surface design

**Title**: New `create` command or flag on `generate`?

**Why it matters**: Determines the user-facing workflow. Should `prescribe create` be a separate command that optionally calls `generate` first, or should `prescribe generate --create` combine both steps?

**Recommended widget**: `select`
- Option 1: New `prescribe create` command (separate from `generate`)
- Option 2: Add `--create` flag to existing `prescribe generate` command
- Option 3: Both (allow `generate --create` and standalone `create`)

## Question 3: PR data source

**Title**: Always generate first, or allow reuse of last generated PR data?

**Why it matters**: Affects workflow efficiency. If a user just ran `prescribe generate`, should `prescribe create` reuse that output, or always regenerate? This impacts session management and user experience.

**Recommended widget**: `confirm`
- Yes: Always generate PR description first (even if session exists)
- No: Allow reuse of last generated PR data from session (with `--use-last` flag or similar)

## Question 4: Branch management

**Title**: Should `prescribe create` handle branch pushing/creation?

**Why it matters**: Currently `prescribe` reads git state but doesn't modify it. Should PR creation include pushing the current branch and/or creating it if it doesn't exist remotely? This affects the end-to-end workflow completeness.

**Recommended widget**: `form` (multiple related choices)
- Push current branch before creating PR? (yes/no)
- Create branch remotely if it doesn't exist? (yes/no)
- Default base branch inference? (select: current default branch, prompt user, require `--base` flag)

## Question 5: Draft PRs

**Title**: Support creating draft PRs?

**Why it matters**: GitHub supports draft PRs (`gh pr create --draft`). Should `prescribe create` support this, and if so, should it be the default or opt-in?

**Recommended widget**: `form`
- Support draft PRs? (yes/no)
- Default to draft? (yes/no, only if support is yes)
- Flag name: `--draft` or `--ready`? (select)

## Question 6: Dry-run / preview mode

**Title**: Should `prescribe create` support a dry-run/preview mode?

**Why it matters**: Users may want to see what would be created without actually creating the PR. This affects error handling, output format, and user confidence.

**Recommended widget**: `confirm`
- Yes: Support `--dry-run` flag that shows what would be created
- No: Always create PR (no preview mode)

## Question 7: Error handling and user feedback

**Title**: What should happen on failure?

**Why it matters**: If PR creation fails (auth error, network issue, branch conflict, etc.), should the tool:
- Exit with error code and message?
- Retry automatically?
- Save generated PR data for manual retry?
- Show partial success (e.g., branch pushed but PR creation failed)?

**Recommended widget**: `select`
- Option 1: Exit with error, show clear message
- Option 2: Exit with error + save generated PR data to file for manual retry
- Option 3: Interactive retry prompt
- Option 4: Combination (save data + clear error message)

## Question 8: Title and body source

**Title**: Allow user override of generated title/body?

**Why it matters**: Users may want to edit the generated PR title or body before creating. Should `prescribe create` support `--title` and `--body` flags to override generated content?

**Recommended widget**: `confirm`
- Yes: Support `--title` and `--body` flags to override generated content
- No: Always use generated title/body (no overrides)

