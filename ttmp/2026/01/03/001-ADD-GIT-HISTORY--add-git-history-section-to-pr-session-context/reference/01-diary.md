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
    - Path: README.md
      Note: Document git_history and git_context usage
    - Path: cmd/prescribe/cmds/context/add.go
      Note: Constructor-based Glazed BareCommand wiring
    - Path: cmd/prescribe/cmds/context/git/root.go
      Note: `context git` registration (subgroup root.go) during CLI refactor
    - Path: cmd/prescribe/cmds/context/git/legacy.go
      Note: Temporary monolithic `context git` verbs kept for incremental split into per-verb packages
    - Path: cmd/prescribe/cmds/context/git/list.go
      Note: `context git list` split into one-verb file during CLI refactor
    - Path: cmd/prescribe/cmds/context/git/remove.go
      Note: `context git remove` split into one-verb file during CLI refactor
    - Path: cmd/prescribe/cmds/context/git/clear.go
      Note: `context git clear` split into one-verb file during CLI refactor
    - Path: cmd/prescribe/cmds/context/root.go
      Note: First group migrated to root.go registration
    - Path: internal/api/prompt.go
      Note: Confirms .commits currently empty
    - Path: internal/controller/controller.go
      Note: |-
        Where generation request is assembled
        Conditionally inject and materialize git context
    - Path: internal/domain/domain.go
      Note: Add git_history/git_context domain types
    - Path: internal/git/context_items.go
      Note: Git materialization + truncation
    - Path: internal/git/git.go
      Note: Commit history extraction and formatting
    - Path: internal/prompts/assets/create-pull-request.yaml
      Note: Prompt contract for commit history
    - Path: internal/session/session.go
      Note: Persist git_history/git_context schema
    - Path: test-scripts/test-all.sh
      Note: Updated smoke suite to assert git history presence
    - Path: test/test-cli.sh
      Note: Export tests assert git_history and BEGIN COMMITS
    - Path: ttmp/2026/01/03/001-ADD-GIT-HISTORY--add-git-history-section-to-pr-session-context/analysis/02-architecture-current-structure-and-modularization-opportunities.md
      Note: Architecture snapshot for plugin refactor
    - Path: ttmp/2026/01/03/001-ADD-GIT-HISTORY--add-git-history-section-to-pr-session-context/design-doc/02-plugin-based-context-providers-proposed-architecture-and-migration-plan.md
      Note: Provider/registry design proposal
    - Path: ttmp/2026/01/03/001-ADD-GIT-HISTORY--add-git-history-section-to-pr-session-context/design-doc/03-refactor-cli-migrate-cobra-verbs-to-glazed-and-reorganize-command-packages.md
      Note: CLI Glazed-first refactor proposal
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

**Commit (code):** 362c0f6 — "Context: add git history to prompt and exports"

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

## Step 5: Design explicit session.yaml controls + `context git` verbs

This step focuses on making Git history inclusion *session-configurable* rather than implicitly always-on. The current implementation derives and injects a default history snippet at request-build time, which is convenient but not controllable or shareable as a session template. The outcome is a design for a `git_history:` block in `session.yaml` and a proposed `prescribe context git history ...` command group to mutate that config.

**Commit (code):** N/A

### What I did
- Audited how `session.yaml` is serialized (`internal/session/session.go`) and how CLI `context` subcommands mutate and save sessions (`cmd/prescribe/cmds/context/add.go` + controller save/load).
- Drafted a `git_history:` schema and a `context git history` verb set (show/enable/disable/set) in a design doc.

### Why
- Git history can be token-expensive and noisy; teams want deterministic control (commit count, merge handling) and the ability to share those settings via committed sessions.

### What worked
- There is a clean extension point: representing git history as a derived source controlled by session config and injected by `BuildGenerateDescriptionRequest()` keeps export/token-count/prompt mapping consistent.

### What didn't work
- N/A (design-only step)

### What I learned
- The current session schema has no “derived context sources” concept; adding one likely requires a top-level config block rather than overloading the `context:` list with large stored content.

### What was tricky to build
- Preserving backwards-compatible behavior for existing sessions: deciding whether “missing git_history” implies enabled defaults or disabled.

### What warrants a second pair of eyes
- CLI UX naming and scope: should this live under `prescribe context git ...` (discoverable) or under `prescribe session ...` (more “config-y”)?

### What should be done in the future
- Implement the proposed schema + commands, then add smoke coverage asserting that disabling history removes the `BEGIN COMMITS` block in `--export-rendered`.

### Code review instructions
- Read `ttmp/2026/01/03/001-ADD-GIT-HISTORY--add-git-history-section-to-pr-session-context/design-doc/01-session-git-history-config-and-context-git-verbs.md`.

## Step 6: Extend the design to support adding specific commits and diffs

This step expands the “git context” concept beyond history summaries. The key insight is that users often need to attach *specific* git artifacts (a particular commit, commit patch, file-at-ref snapshot, or a file diff) as explicit context. These should be persisted as a reference-based `git_context:` list in `session.yaml`, and managed via `prescribe context git ...` verbs, rather than being stored as large literal blobs.

**Commit (code):** N/A

### What I did
- Designed a `git_context:` session list schema to represent explicit git-derived items without storing the derived content.
- Proposed CLI verbs for adding/removing/listing items, including file-scoped and commit-scoped diff options.
- Added token-budget guidance and truncation expectations for patch/diff items.

### Why
- History summaries are not enough when the review narrative depends on a particular intermediate commit or when only one file’s evolution matters.
- Persisting “refs + paths” keeps sessions small, stable, and reviewable.

### What worked
- The existing architecture already supports derived injection at request-build time; `git_context` can follow the same pattern as `git_history`.

### What didn't work
- N/A (design-only step)

### What I learned
- Trying to represent commit/file diffs as regular `context:` items would quickly bloat `session.yaml` and become stale; a dedicated derived schema is the right abstraction.

### What was tricky to build
- Avoiding prompt contract churn: the design keeps `.commits` for history, and treats explicit items as strongly delimited “Additional Context” (or optionally appended to `.diff` when appropriate).

### What warrants a second pair of eyes
- UX shape of the verbs (naming + argument order) and whether we should support path filtering and patch truncation in the first iteration.

### What should be done in the future
- Implement `git_context` end-to-end (schema, controller injection, CLI verbs) and add smoke tests verifying a commit/file-diff item appears in exports.

### Code review instructions
- Start in `ttmp/2026/01/03/001-ADD-GIT-HISTORY--add-git-history-section-to-pr-session-context/design-doc/01-session-git-history-config-and-context-git-verbs.md` (section “Explicit git-derived context items”).

## Step 7: Implement session-controlled git history + explicit git context items

This step turns the design into working code: `git_history:` and `git_context:` are now first-class session schema elements, with CLI verbs to mutate them and controller plumbing to materialize them into the generation request. The result is that commit history is no longer “always-on implicit”; it’s explicitly controllable per session, and you can also attach specific git artifacts (commit metadata, commit patches, file-at-ref, file diffs) as additional context without storing huge blobs in YAML.

The implementation keeps the existing prompt contract stable: `.commits` still represents derived history, while explicit git artifacts appear as strongly-delimited additional context items with distinct context types so exports and token counting can reason about them cleanly.

**Commit (code):** 53272bb — "Context: session git history + git context items"

### What I did
- Extended domain/session schema:
  - added `git_history:` config (enabled/max_commits/include_merges/first_parent/include_numstat)
  - added `git_context:` list (kind + refs/paths only; no embedded diff blobs)
- Added CLI verbs:
  - `prescribe context git history show|enable|disable|set`
  - `prescribe context git add ...`, `list`, `remove`, `clear`
- Updated request building:
  - conditional git history injection based on session config (compat: missing block => enabled defaults)
  - materialized `git_context` items at generation time with strong delimiters and truncation markers
- Updated prompt/export/token-count:
  - mapped explicit git context items into `.context` so they show up under “Additional Context”
  - exported git items with stable CDATA payloads in XML export
  - token-count now reports derived git items in the same pass as derived history
- Updated smoke scripts to assert:
  - disabling history removes `BEGIN COMMITS`
  - a configured `git_context` commit item appears in export output

### Why
- Git history is useful but can be noisy/token-expensive; session-level control makes the behavior deterministic and shareable.
- “Refs + paths” keeps `session.yaml` small and reviewable while still allowing deterministic regeneration of diffs/patches at export/generate time.

### What worked
- `GOWORK=off go test ./...` passes.
- `test-scripts/test-cli.sh` and `test-scripts/test-all.sh` pass with the new disable + git_context assertions.

### What didn't work
- Pre-commit lint initially failed under `go.work` due to a go version mismatch; commit/test workflows need `GOWORK=off` for module-only builds.
- The `exhaustive` linter required adding explicit switch cases for the new `ContextType` values in fallback/export formatters, even when the value is logically excluded by an upstream filter.

### What I learned
- `pflag`’s `Changed` bit is critical for config setters: it allows “set true/false” updates without accidentally overwriting fields that the user didn’t touch.

### What was tricky to build
- Truncation semantics: caps need to be applied to the large diff bodies while keeping the outer delimiters intact so exports remain parseable and reviewers can see what was truncated.

### What warrants a second pair of eyes
- The default caps for `git_context` diff/patch items (bytes/tokens) and whether the truncation marker format should be standardized across other exporters.
- Whether `commit_patch` should default to include commit metadata (currently patch is diff-only; metadata is a separate item).

### What should be done in the future
- Consider adding per-item config knobs (e.g., include_numstat on commit items, patch caps, and diff formatting options) if users need finer-grained control. N/A if current UX is sufficient.

### Code review instructions
- Start in `internal/session/session.go` (schema) and `cmd/prescribe/cmds/context/git.go` (CLI verbs), then follow:
  - `internal/controller/controller.go` (request materialization)
  - `internal/git/context_items.go` (git plumbing + truncation)
  - `internal/api/prompt.go` + `internal/export/context.go` (render/export behavior)
  - `test-scripts/test-cli.sh` (smoke coverage)

## Step 8: Document current architecture + propose plugin-based context providers

This step captures where `prescribe` is architecturally today (layers, critical seams, and current couplings) and turns that into a concrete proposal for evolving toward a plugin-style “context provider” system. The key goal is to make it easy to add new derived context sources (beyond git history/artifacts) without growing controller/export/prompt code into an unmaintainable set of hard-coded feature branches.

The outcome is two ticket-scoped docs: an architecture snapshot and a design doc describing a provider interface + context-type registry, with an incremental migration plan that preserves existing behavior.

**Commit (code):** N/A (docs only)

### What I did
- Wrote an architecture analysis explaining the current data flows and boundaries:
  - domain/session/controller/api/export/cli responsibilities
  - how “literal” vs “derived” context already behaves like a plugin concept
- Wrote a design doc proposing:
  - an internal (compile-time) provider interface run by the controller
  - a registry that routes context types into prompt lanes + export renderers
  - a migration plan that refactors existing git_history/git_context into providers first

### Why
- Derived context sources are multiplying; we need a stable extension surface that doesn’t require editing a half-dozen switches per new context type.
- A provider pipeline keeps `BuildGenerateDescriptionRequest()` as the canonical “what we send” builder while modularizing “how additional context is derived”.

### What worked
- docmgr workflows for adding/relating docs are straightforward and keep the ticket index navigable.

### What didn't work
- N/A (documentation-only step)

### What I learned
- The system already has the critical seam needed for a plugin system: the canonical request builder (`BuildGenerateDescriptionRequest()`). The main missing piece is a registry to centralize routing decisions currently spread across prompt/export/token-count.

### What was tricky to build
- Keeping the “plugin” definition concrete: distinguishing an internal provider registry (practical) from runtime-dynamic plugins (operationally complex and likely unnecessary).

### What warrants a second pair of eyes
- Confirm the desired “plugin level” (in-tree providers only vs out-of-tree compile-time provider packs vs runtime loading).
- Confirm whether provider output should ever be allowed to influence prompt variables beyond existing lanes (`.commits`, `.context`, `.diff`).

### What should be done in the future
- If we adopt this, open a dedicated refactor ticket:
  - add provider interface + provider runner
  - add context type registry
  - port git_history/git_context to providers

### Code review instructions
- Read these docs:
  - `ttmp/2026/01/03/001-ADD-GIT-HISTORY--add-git-history-section-to-pr-session-context/analysis/02-architecture-current-structure-and-modularization-opportunities.md`
  - `ttmp/2026/01/03/001-ADD-GIT-HISTORY--add-git-history-section-to-pr-session-context/design-doc/02-plugin-based-context-providers-proposed-architecture-and-migration-plan.md`

## Step 9: Design CLI refactor to Glazed-first + directory-per-subgroup layout

This step proposes a structural refactor of the CLI command tree to make it consistent and scalable: all verbs should be implemented using the Glazed command pattern (even “print-only” verbs), and command source files should live in a directory structure that mirrors the CLI hierarchy (one file per verb, directory per subgroup).

The immediate motivation is that nested subcommand trees (like `context git history ...`) currently bundle multiple verbs in a single Cobra-only file, which makes it harder to extend and harder to test. Standardizing on Glazed construction and a filesystem layout that mirrors the CLI makes future work (new commands, new context providers) more mechanical and less error-prone.

**Commit (code):** N/A (docs only)

### What I did
- Read the Glazed tutorial guidance (`glaze help build-first-command`) and extracted the key patterns used by Glazed commands (command descriptions, layers, and Cobra integration).
- Drafted a design doc that:
  - defines the target directory layout `cmd/prescribe/cmds/<group>/<subgroup...>/<verb>.go`,
  - standardizes on Glazed wrappers for all verbs (BareCommand or GlazeCommand),
  - proposes top-down `Init()` registration per group/subgroup,
  - includes an incremental migration plan starting with `context git ...`.

### Why
- Consistency: a single pattern for flags/layers/middlewares makes commands easier to add and review.
- Maintainability: directory-per-subgroup avoids “mega command files” and makes the CLI tree discoverable via filesystem browsing.

### What worked
- The repo already uses Glazed for most commands; the remaining Cobra-only areas are localized and good first targets.

### What didn't work
- The `glazed` binary name was not available in PATH; the correct command is `glaze` for reading help/tutorial docs.

### What I learned
- Glazed commands can stay “Cobra-native” at the edges (still a Cobra tree) while centralizing parameter handling and output formats in a single command implementation.

### What was tricky to build
- Designing the filesystem layout in a way that matches Go packaging constraints (directories imply packages) while keeping the root `cmds/root.go` imports stable.

### What warrants a second pair of eyes
- Confirm the final naming convention for the “root subgroup” folder (`root/` vs `default/`) and whether we require structured output flags on action commands.

### What should be done in the future
- Create a dedicated refactor ticket to implement the migration mechanically and keep it reviewable.

### Code review instructions
- Read: `ttmp/2026/01/03/001-ADD-GIT-HISTORY--add-git-history-section-to-pr-session-context/design-doc/03-refactor-cli-migrate-cobra-verbs-to-glazed-and-reorganize-command-packages.md`

## Step 10: Update CLI refactor design (root.go registration, no Init methods)

This step updates the CLI refactor design based on an explicit constraint: group files are just `root.go`, and those files perform all subcommand registration. We should not use `Init()` methods (or `Init*Cmd` helper patterns) in the new structure.

The practical impact is that command trees become self-contained constructors per package: `root.go` builds and wires the cobra.Command subtree by calling verb constructors and `AddCommand(...)` directly. This keeps registration code local and eliminates init ordering concerns.

**Commit (code):** N/A (docs only)

### What I did
- Updated the CLI refactor design doc to:
  - include `root.go` files at each group/subgroup level,
  - remove the `Init()`-based registration plan,
  - adjust example path mappings to match `root.go` ownership.

### Why
- The `Init()` pattern spreads registration logic across many files and depends on call ordering from parents; `root.go` ownership makes the CLI tree easier to reason about and less error-prone during refactors.

### What warrants a second pair of eyes
- Confirm whether leaf verb files should expose `New<Verb>CobraCommand()` (return `*cobra.Command`) or expose the Glazed command constructor and let `root.go` do `cli.BuildCobraCommand(...)` consistently.

## Step 11: Start CLI refactor — remove Init() from `context` group

This step begins the CLI refactor implementation by migrating the `context` group to constructor-based registration (no `Init()` method, no package-level `*Cmd` globals). The intent is to make the command tree wiring local to group `root.go` files, which is the prerequisite for later splitting `context git ...` into one-file-per-verb subpackages and converting all remaining Cobra-only handlers into Glazed commands.

Behavior should remain identical: flags and outputs are unchanged, and smoke tests still pass.

**Commit (code):** eeab311 — "CLI: context command without Init()"

### What I did
- Removed `cmd/prescribe/cmds/context/context.go` and replaced it with `cmd/prescribe/cmds/context/root.go` (`NewContextCmd()` registers subcommands).
- Refactored `cmd/prescribe/cmds/context/add.go` to export a constructor (`NewAddCobraCommand`) instead of `InitAddCmd` + `AddCmd` global.
- Refactored `cmd/prescribe/cmds/context/git.go` to export a constructor (`NewGitCmd`) instead of `InitGitCmd` + `GitCmd` global.
- Updated `cmd/prescribe/cmds/root.go` to add the context group via `context.NewContextCmd()` (stopping the `context.Init()` call path).

### Why
- Establish the “root.go owns registration” invariant in real code, starting with a small/contained group.
- Reduce implicit init ordering and global mutable command vars, making the CLI tree easier to restructure safely.

### What worked
- `GOWORK=off go test ./...` passes.
- `bash test-scripts/test-cli.sh` passes after the change.

### What warrants a second pair of eyes
- Confirm that the constructor naming and error handling patterns (`NewXCmd` returning `(*cobra.Command, error)`) are acceptable as the standard for the broader refactor.

## Step 12: Move `context git` into a dedicated subpackage

This step begins the actual `context git ...` split by moving the git subtree into `cmd/prescribe/cmds/context/git/` and introducing a `root.go` group file. The immediate goal is mechanical: keep behavior identical while establishing the directory-per-subgroup layout that future verb splits can build on.

This is intentionally incremental. The verbs are still implemented in a temporary “legacy” file so we can land the package move as a small diff before splitting one verb per file and converting the remaining Cobra-only handlers to Glazed.

**Commit (code):** f92d96f — "CLI: move context git into subpackage"

### What I did
- Moved the `context git` cobra subtree into `cmd/prescribe/cmds/context/git/` and created `root.go` to own registration.
- Updated `cmd/prescribe/cmds/context/root.go` to attach the git subtree via `git.NewGitCmd()` (no cross-package Init patterns).
- Ran:
  - `GOWORK=off go test ./...`
  - `bash test-scripts/test-cli.sh`

### Why
- Establish the target filesystem layout (`.../context/git/root.go`) before splitting verbs and subgroups, keeping each commit reviewable.

### What was tricky to build
- Avoiding import cycles while changing package boundaries and keeping the `context` group constructor simple.

### What warrants a second pair of eyes
- Confirm the new package boundary is the right one (`context/git` as its own package) and that there are no remaining references to the removed `cmd/prescribe/cmds/context/git.go`.

### What should be done in the future
- Split the leaf verbs into `list.go`, `remove.go`, `clear.go`.
- Split `add` and `history` into their own subpackages (`add/root.go`, `history/root.go`), one file per verb, and convert the verbs to Glazed commands.

## Step 13: Split `context git` leaf verbs into one-file-per-verb

This step pulls the `list`, `remove`, and `clear` verbs out of the temporary monolithic file into `cmd/prescribe/cmds/context/git/{list,remove,clear}.go`. The goal is still mechanical: move code into the target layout without changing behavior, so subsequent commits can focus on converting the remaining verbs and splitting the `add` and `history` subtrees.

**Commit (code):** c77b536 — "CLI: split context git list/remove/clear"

### What I did
- Moved `context git list/remove/clear` cobra command constructors into their own files.
- Ran:
  - `GOWORK=off go test ./...`
  - `bash test-scripts/test-cli.sh`

### What was tricky to build
- Keeping the registration wiring in `git/root.go` unchanged while moving the implementations across files.

### What warrants a second pair of eyes
- Verify the new files don’t accidentally diverge in help text/flags, and that `git/legacy.go` no longer contains duplicate definitions.

### What should be done in the future
- Split `add` and `history` into subpackages and remove the remaining “legacy” file.
