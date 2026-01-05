---
Title: Session git_history config and context git verbs
Ticket: 001-ADD-GIT-HISTORY
Status: active
Topics:
    - git
    - pr
    - context
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/prescribe/cmds/context
      Note: add new context git subcommands here (new files under this dir)
    - Path: cmd/prescribe/cmds/context/add.go
      Note: Existing context mutation command pattern (load session
    - Path: cmd/prescribe/cmds/context/context.go
      Note: Context command group; place to add context git subcommands
    - Path: internal/api/prompt.go
      Note: Decide how git_context-derived diffs/contents map into prompt variables (.context vs .diff)
    - Path: internal/controller/controller.go
      Note: Where git history injection would become conditional
    - Path: internal/export/context.go
      Note: Ensure git_context-derived items are exported with stable delimiters
    - Path: internal/git/git.go
      Note: History builder; would be parameterized by config knobs
    - Path: internal/session/session.go
      Note: Session YAML schema to extend with git_history config
ExternalSources: []
Summary: Make Git history inclusion controllable by persisting a git_history config block in session.yaml and adding `prescribe context git *` verbs to manage it.
LastUpdated: 2026-01-03T18:26:50.29488178-05:00
WhatFor: Define the session.yaml schema and CLI UX to enable/disable and tune Git history context generation.
WhenToUse: When implementing explicit Git history controls instead of always injecting a default history block.
---



# Session git_history config and context git verbs

## Executive Summary

Git history is currently injected into the generation request automatically (defaulting to 30 non-merge commits) and is not persisted in `session.yaml`. This makes it hard to:
- disable history for a given session,
- tune commit count / merge handling deterministically, and
- share the same history configuration across a team (via committed sessions/templates).

This design introduces an explicit `git_history:` configuration block in `session.yaml` and a set of `prescribe context git ...` verbs to manage that configuration. The intent is to treat Git history as a **derived context source** controlled by session state, not as an opaque text blob stored in `context:`.

## Problem Statement

We want the Git history section to be:
- **controllable** (enabled/disabled; knobs like max commits, merge handling),
- **persistent** (stored in `session.yaml` so it survives save/load and can be shared),
- **reproducible** (same session config yields the same history output in export modes),
- **safe for token budgets** (defaults remain conservative; allow truncation policy),
- and **compatible** with the existing prompt contract (`.commits` in `internal/prompts/assets/create-pull-request.yaml`).

Today:
- history injection happens during `Controller.BuildGenerateDescriptionRequest()` and is unconditional,
- `session.yaml` has no concept of derived git context,
- `prescribe context add` only persists raw file/note items.

## Proposed Solution

### 1) Persist a `git_history` block in `session.yaml`

Add a new optional top-level session field:

```yaml
git_history:
  enabled: true
  max_commits: 30
  include_merges: false
  first_parent: false
  include_diffstat: true
  include_numstat: false
  # future:
  # include_patches: false
  # max_patch_commits: 2
  # max_patch_files: 5
  # max_tokens: 1500
```

Notes:
- This config describes **how to derive** the `.commits` payload; it does not store the derived text.
- `max_commits` controls the size; `include_merges` / `first_parent` reduce noise.
- `include_numstat` is more detailed than the current “summary only”; it should default off.
- Patches should remain opt-in with strict limits if we support them.

### 2) Represent the config in the domain model

Add something like `domain.GitHistoryConfig` on `domain.PRData`, so:
- `internal/session` can serialize/deserialize it
- CLI/TUI can display and mutate it as part of session state
- `Controller.BuildGenerateDescriptionRequest()` can conditionally add a `git_history` context item using this config

This keeps the controller as the single “canonical request builder” and avoids writing ad-hoc YAML patch logic in CLI commands.

### 3) Make injection conditional

`BuildGenerateDescriptionRequest()` should include git history only when:
- `data.GitHistory.Enabled == true` (or when legacy behavior applies; see Compatibility below).

The injected representation can remain what exists today:
- append a `domain.ContextItem{Type: ContextTypeGitHistory, ... Content: <xml-ish commits>}` to `AdditionalContext`
- map that context item to the prompt variable `.commits`

This preserves:
- token-count accounting (via request builder)
- export formatting (Git history section is derived from that context item)
- prompt contract (`.commits`)

### 4) Add `prescribe context git *` verbs to manage the config

Add a new `context git` command group with these verbs:

#### `prescribe context git history show`
Shows the effective config (including defaults) for the current session:
- `enabled`
- `max_commits`
- merge options
- current derived range (`target..source`)

#### `prescribe context git history enable|disable`
Toggles `session.yaml: git_history.enabled` and saves the session.

#### `prescribe context git history set`
Updates one or more fields and saves the session:

```bash
prescribe context git history set --enabled=true --max-commits 50
prescribe context git history set --include-merges=true
prescribe context git history set --first-parent=true
prescribe context git history set --include-numstat=true
```

Design notes:
- Keep flags additive and explicit (`--enabled/--disabled` or `--enabled=true/false`).
- Avoid too many knobs until they’re implemented in `internal/git`.

#### Optional: `prescribe context git history clear`
Resets to defaults (writes an explicit default block) or removes the block (depending on compatibility approach).

### 5) Update `session init --save` to write the block explicitly

If “explicit in session.yaml” is the goal, then sessions created with `--save` should contain a `git_history:` block (even if it matches defaults).

This can be done by:
- having `session.NewSession` always populate the field (no `omitempty`), or
- setting defaults at init time.

## Design Decisions

### Why a top-level `git_history:` block (vs `context:` items)

- `context:` today is for **literal content** (files and notes) stored in the session.
- Git history is **derived** from branches and can be regenerated; storing it literally makes sessions:
  - large,
  - stale,
  - and hard to diff-review.

A dedicated config block makes the derived nature explicit and keeps sessions small.

### Why manage it under `prescribe context git ...`

Users already think of “things sent to the model besides file diffs” as *context*. A `context git` group also leaves room for future git-derived context sources:
- `context git diff` (explicitly include a full diff blob)
- `context git blame` (targeted blame for selected lines/files)
- `context git files` (changed-files summary)
- `context git commit` (include an entire commit, or a commit-scoped diff)
- `context git file` (include a file at a given ref, or a file diff between refs)

## Explicit git-derived context items (commits, commit diffs, file diffs)

Git history answers “what happened over time”, but sometimes you also need to add *specific* git artifacts as context:
- a single commit (message + author) that captures intent,
- the patch for a specific commit (e.g. when the final branch diff is too large/noisy),
- the diff for a single file between two refs,
- the content of a file at a specific ref for comparison.

These should be modeled as **derived context items** and persisted as configuration in the session, not stored as large blobs.

### Session schema: `git_context` list

Add a new optional top-level list in `session.yaml`:

```yaml
git_context:
  - kind: commit
    ref: 362c0f6
    include_message: true
    include_author: true
    include_patch: false
    include_numstat: true

  - kind: commit_patch
    ref: 362c0f6
    paths: ["internal/git/git.go"]   # optional; restrict patch to selected paths

  - kind: file_at_ref
    ref: master
    path: internal/api/prompt.go

  - kind: file_diff
    from: master
    to: feature/user-auth
    path: internal/api/prompt.go
```

Key properties:
- Small and stable: these are references (refs + paths), not embedded content.
- Deterministic: can be regenerated from git at generation/export time.
- Composable: multiple items can be added and ordered.

### CLI verbs: `prescribe context git ...`

Add a `context git` command group with subcommands to manage `git_history` (config) and `git_context` (items).

#### `prescribe context git list`
List configured git context items (with indices/IDs) and show effective token estimate (best effort).

#### `prescribe context git remove <index>`
Remove a configured item by index (stable ordering in `session.yaml`).

#### `prescribe context git clear`
Clear all configured git context items.

#### `prescribe context git add commit <ref>`
Adds a commit metadata item (default: message + author + date + numstat summary; no patch).

Flags:
- `--with-patch` (opt-in)
- `--paths <glob/path>...` (limits patch content)
- `--message-only` / `--no-author` (optional)

#### `prescribe context git add commit-patch <ref>`
Adds an explicit commit patch item (diff), optionally restricted to paths.

#### `prescribe context git add file-at <ref> <path>`
Adds full file content at a specific ref (for comparison or “what did this look like on base?”).

#### `prescribe context git add file-diff --from <ref> --to <ref> --path <path>`
Adds a diff for a single file between two refs.

### Formatting and prompt integration

To avoid expanding the prompt contract, these derived git items should be attached as additional context items with distinct types, e.g.:
- `git_commit` (metadata-only)
- `git_commit_patch` (diff text)
- `git_file_at_ref` (file content)
- `git_file_diff` (diff text)

Mapping:
- `git_history` continues to feed `.commits` (existing prompt contract).
- Explicit git items should render under “Additional Context” (or optionally be appended to `.diff` when they are diffs), but always with strong delimiters.

### Token-budget and truncation

Defaults should be conservative:
- metadata-only commit items are “cheap” and can default-on
- patches/diffs must be explicitly requested and constrained:
  - cap total bytes/tokens per item
  - cap number of hunks/lines emitted
  - write explicit truncation markers (`... (truncated)`), so reviewers can see what happened


### Compatibility / default behavior

We must decide how to interpret sessions that don’t have `git_history:`:

Option A (recommended): **Missing config == enabled default**
- Preserves current behavior for existing users/sessions.
- `session init --save` writes the block explicitly going forward.

Option B: **Missing config == disabled**
- Would be a behavior change: current generation context would lose history unless users enable it.
- Lower token cost by default, but surprising relative to current behavior.

## Alternatives Considered

### A) Store history as a `context:` item (`type: git_history`, `content: ...`)
Pros:
- No new schema beyond `context` list.
Cons:
- Bloats sessions, can go stale, and cannot be safely regenerated.

### B) Only expose CLI flags on `generate` (no session persistence)
Pros:
- Simple; avoids schema changes.
Cons:
- Not shareable/reproducible across runs and teammates; TUI/session UX becomes inconsistent.

### C) New `session git-history ...` command group instead of `context git ...`
Pros:
- More semantically “session config”.
Cons:
- Users already look in `context` for “what extra context is included”; `context git` is more discoverable.

### D) Add commit/file diffs as normal `context:` items (stored content)
Pros:
- Minimal schema changes (reuse existing `context` list).
Cons:
- Stored diffs go stale and explode `session.yaml` size; hard to review and share.
  Derived, reference-based items are a better fit.

## Implementation Plan

1) **Schema**
   - Add `GitHistoryConfig` types to `internal/session/session.go` and `internal/domain/domain.go`.
   - Add `GitContextItem` list types to `internal/session/session.go` and `internal/domain/domain.go`.
   - Load/save round-trip via `session.NewSession` and `Session.ApplyToData`.

2) **Controller**
   - Make git history injection conditional based on config.
   - Apply config knobs to `internal/git` history building.
   - Expand request-building to materialize `git_context` items into additional context payloads.

3) **CLI**
   - Add `cmd/prescribe/cmds/context/git/...` commands to show/enable/disable/set.
   - Add `cmd/prescribe/cmds/context/git/...` commands to add/list/remove git_context items.
   - Ensure commands load default session, mutate config, and save.

4) **Docs + tests**
   - Update `README.md` session format example to include `git_history` block.
   - Update `README.md` session format example to include `git_context` items once implemented.
   - Update smoke tests to:
     - verify enabling/disabling toggles output in `--export-rendered`
     - assert `BEGIN COMMITS` present only when enabled
     - verify a commit/file-diff item shows up in `--export-context`/`--export-rendered`

## Open Questions

1) Should `first_parent` default to true when merges are included (recommended), or be a separate explicit knob?
2) Should “include merges” be supported at all initially, or only after we have token truncation knobs?
3) How do we handle extremely large histories (hard max tokens/bytes truncation vs max commits only)?
4) Do we want a TUI toggle + config panel for this, or keep it CLI-only initially?

## References

- Prompt contract: `internal/prompts/assets/create-pull-request.yaml` (uses `.commits`)
- Current injection path: `internal/controller/controller.go` (`BuildGenerateDescriptionRequest`)
- Current history builder: `internal/git/git.go` (`BuildCommitHistoryText`)

## Design Decisions

<!-- Document key design decisions and rationale -->

## Alternatives Considered

<!-- List alternative approaches that were considered and why they were rejected -->

## Implementation Plan

<!-- Outline the steps to implement this design -->

## Open Questions

<!-- List any unresolved questions or concerns -->

## References

<!-- Link to related documents, RFCs, or external resources -->
