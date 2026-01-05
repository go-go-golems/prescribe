---
Title: 'Research: Git history section for PR session context'
Ticket: 001-ADD-GIT-HISTORY
Status: active
Topics:
    - git
    - pr
    - context
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/prescribe/cmds/context/add.go
      Note: Context CLI; currently only adds file/note
    - Path: cmd/prescribe/cmds/session/token_count.go
      Note: Token breakdown command; implications depend on representation choice
    - Path: internal/api/prompt.go
      Note: Template variable mapping; commits currently empty
    - Path: internal/controller/controller.go
      Note: Canonical request construction (Source/Target branches, commit refs, files, AdditionalContext)
    - Path: internal/export/context.go
      Note: Export formats; add Git history section for debug parity
    - Path: internal/git/git.go
      Note: Git operations; needs commit history/stat/patch support
    - Path: internal/prompts/assets/create-pull-request.yaml
      Note: Prompt pack contract for commit history via .commits
    - Path: internal/session/session.go
      Note: Session YAML schema; currently only supports file/note context items
ExternalSources: []
Summary: Identify where PR session context is assembled today and propose designs for adding a Git history section (commits/authors/diffstat and optionally targeted diffs).
LastUpdated: 2026-01-03T16:00:34.522863908-05:00
WhatFor: Guide implementation of a Git history section in the PR generation context and prompt variables.
WhenToUse: Before implementing git-history capture/rendering or adding session/CLI/TUI options for commit history.
---


# Research: Git history section for PR session context

## Executive summary

Prescribe already has:
- a canonical “generation request” (`internal/controller/controller.go:BuildGenerateDescriptionRequest`) containing branches, commit refs, included file content/diffs, and `AdditionalContext` items, and
- an embedded prompt pack (`internal/prompts/assets/create-pull-request.yaml`) that explicitly supports a `.commits` variable meant to hold commit history.

However, commit history is not currently computed or attached anywhere: `internal/api/prompt.go` sets `"commits": ""`, and the exported “Commits” sections only contain **ref SHAs** (`SourceCommit`/`TargetCommit`), not a history range.

This document maps the relevant code paths and proposes design options to add a Git history section (commit list + author + diffstat, and optionally targeted diffs per commit/file) while controlling token cost and preserving deterministic formatting for both “export context” and “rendered payload” debug workflows.

## Problem statement

For PR description generation, diffs and selected full files explain *what changed*, but not always *why* or *how the change evolved*. Commit history provides:
- intent (commit messages),
- ownership (authors),
- sequencing (order of work),
- scope signal (diffstat per commit / per file),
and can allow the model to produce a more accurate narrative and changelog.

The risk is that commit history can be extremely noisy (merge commits, large refactors) and can dominate token budget if full patches are included.

## What “PR session context” means in this repo (today)

There are three distinct “context” representations:

1) **Persisted session state** (`internal/session/session.go`)
- `Session.Context []ContextConfig` supports only `type: file|note` and stores literal `content` for context items.
- No configuration exists for “derive context from git”, so any git history integration must either:
  - be stored as raw text content in `session.yaml`, or
  - introduce a new structured session field (recommended if we want persistent options without storing huge text blobs), or
  - be computed at generation/export time (no persistence).

2) **Canonical generation request** (`internal/controller/controller.go:BuildGenerateDescriptionRequest`)
- Constructs `api.GenerateDescriptionRequest` with:
  - `SourceBranch`, `TargetBranch`
  - `SourceCommit`, `TargetCommit` (resolved via `git rev-parse`)
  - `Files` (visible+included only)
  - `AdditionalContext` (file/note items)
  - `Prompt`, `Title`, `Description`
- This is the single best integration point if commit history should be part of “session context” semantics consistently across CLI, TUI, and exports.

3) **Rendered payload / export formats**
- Prompt compilation + template vars: `internal/api/prompt.go`
  - maps request fields into pinocchio-style variables `.diff`, `.code`, `.context`, `.description`, `.title`
  - `.commits` exists in the prompt pack contract, but is currently hard-coded to empty.
- Export-only context blob: `internal/export/context.go:BuildGenerationContext`
  - serializes the request into XML/markdown/simple/etc.
  - includes a “Commits” section in markdown/XML, but this is just `SourceCommit`/`TargetCommit`.

## Existing prompt contract for commits (high-signal)

The embedded prompt pack (`internal/prompts/assets/create-pull-request.yaml`) contains:
- a `commits` flag (stringFromFile) and
- a template block:

```
{{ if .commits }} ... --- BEGIN COMMITS ... {{ .commits }} ... --- END COMMITS {{ end }}
```

This strongly suggests the intended design: make “commit history text” available as `.commits` for prompt rendering. Today that channel is unused.

## Relevant code touchpoints (implementation map)

**Git data acquisition**
- `internal/git/git.go`
  - already supports: `ResolveCommit`, `GetDiff`, `GetChangedFiles` (via `git diff --numstat`), file content and file diff
  - does not support: commit range listing, per-commit numstat, patch extraction by commit

**Session persistence**
- `internal/session/session.go`
  - `ContextConfig.Type` is effectively an enum of `"file"|"note"` (stringly typed)
  - adding a new persisted “git history” concept would require schema extension

**CLI: adding context**
- `cmd/prescribe/cmds/context/add.go`
  - only `--note` or `<file-path>` argument; no git-derived items

**Prompt compilation**
- `internal/api/prompt.go`
  - `buildTemplateVars()` currently sets `"commits": ""`
  - diffs are emitted as XML-ish per-file blocks inside `.diff` for delimiter safety

**Exports / debug**
- `internal/export/context.go`
  - formats “generation context” and “rendered payload”
  - currently uses “Commits” heading for commit SHAs, not history

**Token-count UX**
- `cmd/prescribe/cmds/session/token_count.go`
  - counts tokens for included files and all `AdditionalContext` items
  - if Git history becomes an `AdditionalContext` item, token-count “just works”; if it becomes a new request field, token-count needs updating (or an explicit row for git history)

## Design options

### Option A (recommended): First-class “commit history” field feeding `.commits`

Add a new field (or structured sub-object) to `api.GenerateDescriptionRequest`, populated by the controller from git, then mapped into template vars:
- `GenerateDescriptionRequest.Commits string` (pre-rendered text) or
- `GenerateDescriptionRequest.GitHistory GitHistory` (structured, then rendered in `buildTemplateVars`).

Pros
- Aligns with existing prompt contract (`.commits`) without overloading `AdditionalContext`.
- Lets us render commits in a format optimized for the model (and independently for export formats).
- Avoids persisting huge blobs in `session.yaml` unless explicitly desired.

Cons
- Requires changes in multiple call sites (controller, api, export, token-count).

### Option B: New `domain.ContextTypeGitHistory` stored as `AdditionalContext`

Introduce a new `ContextType` (e.g. `"git_history"`) and store a single context item containing a formatted commit history section.

Pros
- Minimal surface change: it fits the existing “additional context” concept and token-count path.
- Session persistence can be “free” if we store the full rendered history as content.

Cons
- The prompt pack expects `.commits`, not “some context file content”; it may still work but loses the explicit contract.
- Persisting full history in `session.yaml` can become stale and can bloat session files.
- Harder to support “options” (max commits, stat vs patch) without introducing schema anyway.

### Option C: Export-only git history section (human debug only)

Only add history to `internal/export/context.go` outputs, without sending it to the model.

Pros
- Low risk and avoids token blowups.

Cons
- Does not actually improve generation quality (the model never sees it), so it doesn’t meet the likely intent.

## Recommended content shape (what to include)

### Baseline (default) history payload (high signal, low tokens)

For a PR range `target..source`, include the top N commits:
- short SHA
- author name (and optionally email)
- author date (ISO)
- subject line

Example (single-line per commit):
`<sha> <date> <author>: <subject>`

### Add diffstat (opt-in or default-on with limits)

Two escalating levels:
- **per-commit summary**: `(+X/-Y, K files)` (cheap)
- **per-commit per-file numstat**: `path +a -d` (more expensive; keep top K files per commit)

Avoid including full patches by default; instead provide optional targeted patches (below).

### Targeted diffs from specific commits/files (explicit opt-in)

Only include patches when the user requests them, with hard limits:
- max commits with patches (e.g. 1–3)
- max files per commit (e.g. 5)
- max total bytes/tokens for patches (truncate with explicit marker)

Rationale: patches often duplicate the main `Files` diff already in the context; their value is highest for:
- multi-commit refactors where the final diff is hard to read,
- revert/fixup sequences,
- identifying intent in intermediate commits.

## How to compute commit history (git plumbing suggestions)

### Commit range

For “commits in source not in target”, use:
- `git log <target>..<source>`

This matches PR semantics more closely than symmetric difference (`...`).

### Stable parsing strategy (recommended)

Use a delimiter-based format rather than human-oriented output:
- `git log --date=iso-strict --pretty=format:'%H%x1f%an%x1f%ae%x1f%ad%x1f%s%x1e' <target>..<source>`
- If stats are needed, run `git show --numstat --format=... <sha>` per commit (bounded by max commits), or experiment with `git log --numstat` and parse between `%x1e` records.

Avoid parsing `--stat` output if we need structured data; prefer `--numstat` for machine readability.

### Merge commits

Default behavior should likely:
- exclude merges, or
- use `--first-parent` to reduce noise on feature branches that merge main frequently.

Make this configurable.

## Where to render “Git history” in the exported context

To minimize confusion with existing “Commits” (which are ref SHAs), adjust headings:
- keep current section but rename it to **Commit refs** (source/target SHA)
- add a new section **Git history** for the commit list/diffstat

This applies to:
- fallback markdown context (when prompt is empty): `internal/api/api.go:buildUserContext`
- export context formats: `internal/export/context.go:buildMarkdown` and `buildXML`
- rendered payload exports: `internal/export/context.go:BuildRenderedLLMPayload` (optional, but useful for debugging)

## Session persistence and UX design suggestions

### If we want persistence (recommended for TUI usability)

Extend `session.yaml` schema with a new optional block rather than storing raw text:

```yaml
git_history:
  enabled: true
  max_commits: 30
  include_author_email: false
  include_diffstat: true
  include_numstat: false
  include_merge_commits: false
  first_parent: true
  patches:
    enabled: false
    max_commits: 2
    paths: []         # optional file path filters
    max_tokens: 1500
```

This allows regeneration on demand and avoids stale/bloated sessions.

### CLI flags (minimal, composable)

Expose the above as generate/session flags rather than only “context add”:
- `prescribe generate --include-git-history`
- `--git-history-max-commits N`
- `--git-history-diffstat`
- `--git-history-numstat`
- `--git-history-first-parent`
- `--git-history-include-merges`
- `--git-history-patches` (plus bounds)

Rationale: history is fundamentally derived from source/target branches and belongs to the generation input set, not just ad-hoc context items.

### TUI affordances

Treat git history as a distinct “derived context source”:
- toggle on/off
- max commits slider/input
- show estimated tokens for history independently from file diffs/full files

## Token budgeting guidance (defaults)

Suggested defaults that keep context small but useful:
- `max_commits`: 30
- `first_parent`: true
- include only subject lines (no bodies)
- include diffstat summary per commit; omit per-file stats unless requested
- never include patches by default
- truncate with explicit markers:
  - `... (truncated: exceeded max_commits)` or `... (truncated: exceeded max_tokens)`

## Open questions (need decisions before implementation)

1) Should commit history be included by default for generation, or be opt-in?
2) Should “git history” be persisted as configuration in `session.yaml` or computed ad hoc each run?
3) Should the canonical representation live in `GenerateDescriptionRequest` (first-class field) or in `AdditionalContext` (new context type)?
4) How should we avoid leaking sensitive commit messages (optional redaction, or just explicit opt-in)?

## Proposed next steps (implementation outline)

1) Add git service support for commit history range + (optional) numstat/patch extraction.
2) Decide representation (Option A vs B) and wire into:
   - `internal/controller/controller.go` request construction
   - `internal/api/prompt.go` to populate `.commits`
   - `internal/export/context.go` and fallback `buildUserContext` so exports match runtime inputs
   - `cmd/prescribe/cmds/session/token_count.go` if history is not modeled as `AdditionalContext`
3) Add minimal tests around prompt compilation (ensuring `.commits` renders) and export context formatting.
