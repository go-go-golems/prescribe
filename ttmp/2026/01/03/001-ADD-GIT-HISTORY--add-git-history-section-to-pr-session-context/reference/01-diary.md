---
Title: Diary
Ticket: 001-ADD-GIT-HISTORY
Status: active
Topics:
    - git
    - pr
    - context
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/api/prompt.go
      Note: Confirms .commits currently empty
    - Path: internal/controller/controller.go
      Note: Where generation request is assembled
    - Path: internal/git/git.go
      Note: Commit history extraction and formatting
    - Path: internal/prompts/assets/create-pull-request.yaml
      Note: Prompt contract for commit history
    - Path: test-scripts/test-all.sh
      Note: Updated smoke suite to assert git history presence
    - Path: test/test-cli.sh
      Note: Export tests assert git_history and BEGIN COMMITS
ExternalSources: []
Summary: ""
LastUpdated: 2026-01-03T16:00:33.996872013-05:00
WhatFor: ""
WhenToUse: ""
---



# Diary

## Goal

Research and design notes for adding a **Git history** section to the PR session context used by `prescribe generate` (and export/debug output).

## Step 1: Create the ticket and locate prior work

This step established the documentation workspace for the change and looked for any existing design/implementation threads so we don’t reinvent a parallel mechanism. The key outcome was finding a prior ticket that scoped “git history with stat”, but also noticing that the current codebase still only supports file/note context items.

**Commit (code):** N/A

### What I did
- Read workflow instructions: `~/.cursor/commands/docmgr.md`, `~/.cursor/commands/diary.md`, `~/.cursor/commands/git-commit-instructions.md`
- Created ticket `001-ADD-GIT-HISTORY` and seeded docs (diary + analysis)
- Reviewed prior ticket `ttmp/2025/12/27/010-GIT-HISTORY--add-git-history-with-stat-to-context` to understand earlier intent and suggested API shapes

### Why
- Ensure the new work fits the existing docs + code architecture, and reuses prompt/template expectations already present in the repo.

### What worked
- `docmgr` ticket + doc creation workflow is functional in this repo; topic vocabulary needed seeding because it started empty.

### What didn't work
- N/A

### What I learned
- There is a prior analysis ticket proposing `ContextTypeGitHistory`, but the current Go codebase does not yet contain any `git-history` implementation (no matches in `cmd/` or `internal/`).

### What was tricky to build
- N/A (pure research/setup step)

### What warrants a second pair of eyes
- Confirm whether `010-GIT-HISTORY` was superseded by later changes (or if there’s another branch/PR with partial implementation not present in this workspace).

### What should be done in the future
- If `010-GIT-HISTORY` is still relevant, consolidate: either close it in favor of `001-ADD-GIT-HISTORY` or explicitly link the scopes (avoid duplicated “git history” designs).

### Code review instructions
- Start with `ttmp/2025/12/27/010-GIT-HISTORY--add-git-history-with-stat-to-context/analysis/01-analysis-add-git-history-with-stat-to-context.md`

### Technical details
- Ticket created at `ttmp/2026/01/03/001-ADD-GIT-HISTORY--add-git-history-section-to-pr-session-context/`

## Step 2: Locate the PR session context assembly points

This step identified the concrete code touchpoints that define “PR session context” today: what gets persisted in `session.yaml`, what gets compiled into prompt template variables, and what gets exported as a debug context blob. The key outcome is that the embedded prompt pack already has a `.commits` channel intended for commit history, but it is currently always empty.

**Commit (code):** N/A

### What I did
- Traced “generation context” assembly in `internal/controller/controller.go` and `internal/api/prompt.go`
- Verified exported context formats in `internal/export/context.go`
- Located the embedded prompt pack contract in `internal/prompts/assets/create-pull-request.yaml` (expects `.commits`)
- Confirmed session persistence constraints in `internal/session/session.go` and `cmd/prescribe/cmds/context/add.go` (only file/note context)

### Why
- To decide whether Git history should be a new `ContextType`, a first-class request field, or a prompt-only variable populated at render time.

### What worked
- The code paths are relatively centralized:
  - `Controller.BuildGenerateDescriptionRequest()` is the single place that constructs `api.GenerateDescriptionRequest`.
  - `api.buildTemplateVars()` is the single place mapping request fields into prompt variables (`.diff`, `.code`, `.context`, etc).

### What didn't work
- The “commit history” pathway does not exist yet: `.commits` is hard-coded to `""` in `internal/api/prompt.go`, so even the default prompt can’t use it.

### What I learned
- There are two “commit” concepts in the repo:
  - commit refs (`SourceCommit` / `TargetCommit`) already exist and are exported under a “Commits” section, but these are *not* history
  - commit history is explicitly supported by the prompt pack variable `.commits`, but not wired up by the app

### What was tricky to build
- N/A (research), but the future implementation will need careful token budgeting because commit history can easily dominate context.

### What warrants a second pair of eyes
- Confirm whether commit history should feed `.commits` (prompt contract) or be represented as a new `AdditionalContext` item; this affects session persistence, token-count UX, and export formatting.

### What should be done in the future
- Decide on the canonical representation (structured vs pre-rendered string) before adding flags/TUI affordances, to avoid later migrations in `session.yaml`.

### Code review instructions
- Start in `internal/prompts/assets/create-pull-request.yaml` to see the `.commits` contract.
- Then inspect `internal/api/prompt.go` to see that `.commits` is currently empty.

## Step 3: Draft design options and recommend a direction

This step translated the code-path findings into concrete design choices, focusing on minimizing token blowups while still exploiting the existing prompt pack’s “commits history” channel. The main decision is whether git history becomes a first-class request field (feeding `.commits`) or a new `AdditionalContext` item type; the analysis leans toward the first-class approach because it matches the prompt contract and avoids persisting huge blobs.

**Commit (code):** N/A

### What I did
- Wrote a design/implementation map and compared three representation options (first-class request field vs new context type vs export-only).
- Proposed a conservative default payload shape: one-line-per-commit + optional diffstat; patches only as explicit opt-in with hard limits.
- Drafted session/CLI/TUI configuration suggestions to make behavior controllable without storing large raw text in `session.yaml`.

### Why
- Git history is high-signal but also high-volume; without clear defaults and truncation rules, it can crowd out the actual diffs and file context.

### What worked
- The embedded prompt pack already provides a clear insertion point (`.commits`) so the feature can be prompt-driven rather than “hope the model reads it from somewhere”.

### What didn't work
- N/A

### What I learned
- The current “Commits” sections in exports refer to commit SHAs (source/target) and would likely need a rename (e.g. “Commit refs”) to avoid confusion once history is added.

### What was tricky to build
- The future implementation will need robust parsing if we want per-commit file stats; `--numstat` is parseable, `--stat` is not.

### What warrants a second pair of eyes
- Confirm the chosen git log range semantics (`target..source` vs `target...source`) and merge handling (`--first-parent` vs include merges) to match the team’s PR workflow.

### What should be done in the future
- Decide whether commit history is included by default or opt-in; this is primarily a token-budget/product decision.

### Code review instructions
- Read `ttmp/2026/01/03/001-ADD-GIT-HISTORY--add-git-history-section-to-pr-session-context/analysis/01-research-git-history-section-for-pr-session-context.md` for the full option matrix and recommended defaults.

## Step 4: Wire Git history into exported/rendered context + augment smoke tests

This step implemented a minimal end-to-end “Git history section” that flows into the rendered prompt payload (via the existing `.commits` template variable) and into export/debug outputs, then updated the mock-repo smoke scripts to assert it’s present. The design choice here is pragmatic: represent git history as a special `AdditionalContext` item (`type=git_history`) injected at request-build time, then map that to `.commits` instead of treating it like a note.

**Commit (code):** N/A

### What I did
- Added `git_history` as a `domain.ContextType` and taught prompt templating to map it to `.commits`.
- Implemented a git history extractor using `git log --no-merges --numstat` and a separator-based parser, producing an “XML-ish” text snippet (commit + author + date + subject + diffstat summary).
- Injected the computed history into `Controller.BuildGenerateDescriptionRequest()` so it is included in:
  - `generate --export-context`
  - `generate --export-rendered` (and thus the default prompt’s `--- BEGIN COMMITS` block)
- Updated export formatting and fallback user-context formatting to show:
  - **Commit refs** (source/target SHA), and
  - a separate **Git history** section (without duplicating it under “Additional context”).
- Updated token-count to compute git history via the canonical request builder so it stays aligned with what `generate` will send.
- Augmented mock-repo scripts to create multiple commits with different authors and assert the commit-history block is present.

### Why
- The prompt pack already has a `.commits` contract; wiring that up is the smallest change that improves PR generation quality without requiring AI inference for validation.
- Export-only smoke tests must not depend on API keys; `--export-rendered` and `--export-context` allow deterministic verification.

### What worked
- `go test ./...` passes.
- Smoke scripts now validate git history presence without requiring inference.

### What didn't work
- Initial parsing strategy put the record separator *after* the header; `--numstat` lines then appeared as “headers” when splitting. Fix was to emit the record separator at the *start* of each commit record.

### What I learned
- For `git log` + `--numstat`, delimiter placement is critical: record separators must anchor the beginning of a commit record to keep stats with the right commit.

### What was tricky to build
- Parsing `git log` output reliably without relying on human-oriented formatting; `--numstat` is machine-friendly but still needs careful record grouping.

### What warrants a second pair of eyes
- Confirm we want `--no-merges` by default; some workflows rely on merge commits for context.
- Confirm the range semantics (`target..source`) match how you want PR history presented.

### What should be done in the future
- Add configuration/flags to control commit count and merge handling, and potentially support per-file numstat or targeted patches behind explicit opt-in.

### Code review instructions
- Start in `internal/controller/controller.go` (`BuildGenerateDescriptionRequest`) and follow into:
  - `internal/git/git.go` (`BuildCommitHistoryText`)
  - `internal/api/prompt.go` (`buildTemplateVars` mapping `.commits`)
  - `internal/export/context.go` (Git history formatting)
  - `test-scripts/` and `test/` (smoke coverage)
