---
Title: 'Architecture: current structure and modularization opportunities'
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
    - Path: cmd/prescribe/cmds/root.go
      Note: CLI command tree initialization and grouping
    - Path: internal/api/prompt.go
      Note: Prompt compilation and template variable routing
    - Path: internal/controller/controller.go
      Note: Canonical request builder and derived context injection seam
    - Path: internal/export/context.go
      Note: Export formatting for context/payloads
    - Path: internal/session/session.go
      Note: Session persistence boundary and schema
ExternalSources: []
Summary: System architecture overview and a modularization map towards plugin-like context sources.
LastUpdated: 2026-01-04T14:28:11.581540961-05:00
WhatFor: Orient contributors and provide a concrete decomposition plan for extracting context generation into modular components.
WhenToUse: When designing new context sources, refactoring CLI/TUI command trees, or introducing plugin-style context providers.
---


# Architecture: current structure and modularization opportunities

## Goal

Capture the **current architecture** of `prescribe` (as it exists today) and identify seams that make it feasible to evolve toward a **plugin-based architecture** for “context providers” (sources that compute + inject derived context into PR generation).

This document focuses on:
- current layers and the critical data flows,
- where responsibilities are currently coupled,
- what “plugin” should mean in this codebase (and what it should not),
- a map of extraction seams that preserve existing behavior while enabling extension.

## Architectural Snapshot (Today)

### Top-level data model (domain)

The core state is `internal/domain.PRData`:
- git identity: `SourceBranch`, `TargetBranch`
- files: `ChangedFiles` (diffs + optional full-file snapshots + token estimates)
- filters: `ActiveFilters` (applied to visibility)
- prompt: `CurrentPrompt` (or a preset)
- context:
  - user-specified: `AdditionalContext` (`file`/`note`)
  - derived config: `GitHistory` config + `GitContext` items (reference-based; materialized later)

Key point: `PRData` is effectively “repo-derived PR state + user session overlays”.

### Session persistence (session.yaml)

`internal/session.Session` persists a *subset* of `PRData`:
- explicit configuration: file inclusion/modes, filters, prompt preset/template, literal `context:` items
- derived configuration:
  - `git_history:` (knobs; history blob is not stored)
  - `git_context:` (refs/paths only; blobs are not stored)

Loading a session applies it onto a freshly initialized controller state (git-derived data first, then session overlays).

### Orchestration: controller as the canonical request builder

`internal/controller.Controller.BuildGenerateDescriptionRequest()` is the primary seam:
- selects visible+included files
- resolves commit refs
- derives + injects derived context:
  - git history (if enabled)
  - configured `git_context` items
- emits `api.GenerateDescriptionRequest` as the canonical “what we send” structure

Export/debug and token-count rely on this. In practice, `BuildGenerateDescriptionRequest()` is the source of truth for the system’s external behavior.

### Git plumbing (internal/git)

`internal/git.Service` provides:
- repo inspection (branch, commit resolution, changed files, diffs, file content at ref)
- derived context generation:
  - commit history snippet
  - explicit git artifacts (commit metadata, commit patch, file-at-ref, file diff), each capped + annotated

### Prompt compilation (internal/api/prompt.go)

Prompt rendering is a “pinocchio-style template” pipeline:
- compile vars from request: `.diff`, `.code`, `.context`, `.commits`, etc.
- default prompt pack renders `.commits` inside a `BEGIN COMMITS` block
- explicit git artifacts are mapped into `.context` so they are visible as additional context blocks

### Export/debug formatting (internal/export + api fallback)

Export is a formatting layer over the canonical request:
- `internal/export/context.go` builds `--export-context` and `--export-rendered` payloads
- `internal/api/api.go` contains a fallback user-context builder for non-pinocchio prompts

### CLI/TUI command trees

The CLI uses Cobra (plus Glazed for tabular commands):
- `cmd/prescribe/cmds/*` registers command groups
- commands typically:
  1) initialize controller from repo,
  2) load default session if it exists,
  3) mutate `PRData`,
  4) save session

The TUI uses the controller as its backing state machine and auto-saves session changes.

## Current “Context” Architecture

It is useful to classify context sources:

### 1) Literal context (persisted blobs)

- `context:` list in session.yaml
- types: `file`, `note`
- behavior: deterministic from session.yaml; no git recomputation required at generate-time

### 2) Derived context (persisted references/config)

- `git_history:` and `git_context:`
- behavior: deterministic from (session config + repo state at generation time)
- explicit non-goal: storing large git-derived blobs in YAML

This is already “plugin-like” in spirit: persisted config drives generation-time derivation, but the implementations are currently hard-coded.

## Couplings / Pain Points (What Blocks Modularity Today)

1) **Controller hard-codes derived context sources**
   - `BuildGenerateDescriptionRequest()` calls specific git builders directly.

2) **Session schema is hand-mapped**
   - adding a new derived context source requires edits across:
     - `internal/domain`
     - `internal/session`
     - controller injection
     - export/prompt mapping (sometimes)

3) **Prompt variables are a global contract**
   - `.commits` is a “special lane”; other context mostly flows through `.context` or `.diff`.
   - a plugin system must decide whether plugins can introduce new variables (likely “no”, initially).

4) **Exports and token-count are coupled to ContextType values**
   - adding context types often requires updating `switch` statements and formatting.

5) **The git service is both repository inspector and context generator**
   - feasible for now, but a plugin system likely wants a higher-level “context provider” interface that can use git utilities without being “the git module”.

## Modularization Map (Incremental)

### Step 1: Establish a first-class “Context Provider” concept in controller (internal plugin system)

Add an internal package (e.g. `internal/contextproviders`) that defines:
- a small provider interface (input: request-building context; output: `[]domain.ContextItem`)
- strong contracts: ordering, error semantics, truncation responsibility

Then controller becomes:
- base request assembly (files/branches/commits/prompt)
- run configured providers to append derived `AdditionalContext`

This is a refactor with no behavior change.

### Step 2: Separate config persistence from provider implementation

Treat session.yaml fields as:
- global config blocks (like `git_history`)
- provider-specific item lists (like `git_context`)

Provider implementations should be responsible for:
- interpreting their config structs,
- materializing derived context items.

### Step 3: Normalize “context item formatting” + truncation contract

Introduce shared helpers:
- a stable envelope format (XML-ish or begin/end) for derived items
- a standard truncation marker contract (consistent across providers)
- standard attributes (`source`, `kind`, `ref`, `path`, `range`) so exporters can be consistent

### Step 4: Replace exporter switch statements with a registry (reduce coupling)

Introduce a registry mapping:
- `ContextType` -> export rendering strategy (xml/markdown/simple/begin-end)
- `ContextType` -> prompt inclusion strategy (commits vs context vs diff vs ignore)

This turns “add a new context type” from “touch 6 switch statements” into “register metadata”.

### Step 5: Define “plugin” levels explicitly (avoid premature complexity)

There are multiple “plugin” meanings in Go; we should choose intentionally:

1) **Internal plugin system (compile-time, in-tree)**
   - providers are Go packages in the repo
   - registered via init/registry calls
   - ideal near-term: simplest, testable, cross-platform

2) **External plugin modules (go modules + build tags)**
   - providers live in separate repos but are compiled into the binary
   - selected by build tags or a “plugin pack” module that imports them
   - good mid-term if teams want private providers without upstreaming

3) **Runtime dynamic loading (`plugin` package)**
   - only works on some OS/arch combos; operationally complex
   - complicates distribution and reproducibility
   - likely a non-goal for `prescribe` unless there is a strong need

Recommendation: start with (1), evolve to (2) if needed; avoid (3) unless a compelling user story appears.

## Minimal “Plugin Surface” Proposal (What a provider should be able to do)

For derived PR context, a provider should:
- determine whether it is enabled (via session config)
- generate zero or more `domain.ContextItem`s with stable types and paths/labels
- enforce caps/truncation (or return raw data + let a shared helper cap it)
- optionally contribute to token-count metadata (but the current system already counts items by content)

What a provider should *not* do (initially):
- mutate core session state beyond its own config area
- change prompt templates or introduce new template variables
- perform network calls during generation (keep generation deterministic)

## Concrete Seams / Refactor Targets

1) `internal/controller.Controller.BuildGenerateDescriptionRequest()`
   - extract derived context injection into a provider pipeline.

2) `internal/session.Session.ApplyToData()` / `NewSession()`
   - isolate provider config serialization/deserialization.

3) `internal/api/prompt.go`
   - replace hard-coded `ContextType` routing with a registry (commits/context/diff).

4) `internal/export/context.go` and `internal/api/api.go` fallback context
   - move display routing behind the same registry.

## Risks and Constraints

- **Reproducibility**: derived providers should be deterministic given repo state + session config.
- **Token safety**: providers must cap their output and surface truncation explicitly.
- **Ordering**: session ordering matters for `git_context` lists; provider pipeline must preserve it.
- **Compatibility**: missing provider config should have an explicit policy (e.g., git_history missing => enabled defaults).

## Next Actions (Doc-driven)

If we pursue this direction, the next design doc should define:
- `ContextProvider` interface
- provider registry + routing metadata
- session schema pattern for provider config blocks
- CLI UX patterns for “context <provider> ...” verbs

