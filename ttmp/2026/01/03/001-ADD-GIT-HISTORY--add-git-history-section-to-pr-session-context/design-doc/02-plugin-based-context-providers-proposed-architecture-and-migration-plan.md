---
Title: 'Plugin-based context providers: proposed architecture and migration plan'
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
    - Path: internal/api/prompt.go
      Note: Target for registry-based routing
    - Path: internal/controller/controller.go
      Note: Target for provider pipeline refactor
    - Path: internal/domain/domain.go
      Note: ContextType and PRData extension surface
    - Path: internal/export/context.go
      Note: Target for registry-based export rendering
ExternalSources: []
Summary: Define a provider/registry architecture to modularize derived context generation and enable a plugin-like system.
LastUpdated: 2026-01-04T14:29:54.157011274-05:00
WhatFor: Provide a concrete design for modular context providers, including registry, config, and migration steps.
WhenToUse: When refactoring controller/export/prompt routing away from hard-coded derived context sources, or adding new derived context providers.
---


# Plugin-based context providers: proposed architecture and migration plan

## Executive Summary

`prescribe` has grown beyond “diff + notes”: it now supports derived git history (`git_history`) and explicit git artifacts (`git_context`). These are functionally “context sources” that produce additional context at generation time, but the implementations are currently hard-coded into controller, exporter, and prompt mapping.

This design proposes an internal plugin-style architecture centered on **Context Providers**:
- providers are responsible for materializing derived context items from repo/session state,
- a **registry** defines how each context type is routed into prompts and exports,
- controller request building runs a provider pipeline to assemble `AdditionalContext`.

The plan is incremental: first refactor to an internal provider interface + registry (no behavior change), then enable optional out-of-tree providers (compile-time module plugins) if needed.

## Problem Statement

We want to keep adding context features (git artifacts, blame, issue references, changelog summaries, build logs, test summaries, etc.) without the codebase devolving into:
- a controller that accumulates one-off injection logic per feature,
- exporter and prompt code that grows large switch statements for every new context type,
- session schema changes that require touching many parts of the system.

We also want:
- deterministic, reproducible context derivation (especially for export-only workflows),
- token safety (caps + explicit truncation),
- a clear extension surface so new context sources can be added “like plugins”.

## Proposed Solution

### Definitions

**Context Provider**
- A component that, given repo/session state, can produce zero or more `domain.ContextItem`s.
- Providers are allowed to read from git and session config; they are not allowed to mutate unrelated session state.

**Context Type Registry**
- A mapping from `domain.ContextType` to:
  - export rendering behavior (per separator format),
  - prompt routing behavior (which “lane” it belongs to: `.commits`, `.context`, `.diff`, or ignored),
  - optional metadata (human label, ordering group, etc).

### Provider interface (internal)

Add a small interface in a new package (illustrative only):

```go
type Provider interface {
	ID() string
	Enabled(pr *domain.PRData) bool
	Build(ctx context.Context, in BuildInput) ([]domain.ContextItem, error)
}

type BuildInput struct {
	RepoPath      string
	SourceBranch  string
	TargetBranch  string
	SourceCommit  string
	TargetCommit  string
	PR            *domain.PRData
	Git           *git.Service
}
```

### Provider pipeline in the controller

Refactor `BuildGenerateDescriptionRequest()` to:
- build the “base request” (files/commits/prompt)
- run a list of providers to build derived items
- append provider items to `AdditionalContext` (respecting ordering rules)

The initial provider set would include:
- `GitHistoryProvider` (materializes `ContextTypeGitHistory` when enabled)
- `GitContextProvider` (materializes `git_context` list into per-kind context types)

### Context type registry

Add a registry (e.g., `internal/contexttypes`) that declares:
- routing target: `commits` vs `context` vs `diff` vs `notes`
- exporter support: how each type renders in XML/markdown/etc

Then:
- `internal/api/prompt.go` uses the registry to route items (instead of hard-coded switch cases)
- `internal/export/context.go` uses the registry to render items (instead of hard-coded switch cases)
- token-count can tag derived items consistently using type metadata

### Session schema and provider config

Provider configuration stays persisted in `session.yaml`, but the pattern becomes explicit:
- top-level config blocks for provider toggles/knobs (`git_history`)
- provider-owned item lists (`git_context`)

The session layer remains responsible for (de)serialization, but the design encourages keeping provider config scoped and strongly typed.

### Token/byte caps contract

Providers must either:
- enforce caps themselves (like current git_context implementations), or
- return raw data that is capped by shared helper functions before producing final `ContextItem.Content`.

In either case, truncation must be explicit and standardized (e.g., `... [TRUNCATED: ...]` markers).

## Design Decisions

1) **Internal plugin system first (compile-time)**
   - Rationale: simplest to test, cross-platform, avoids runtime plugin complexity.

2) **Registry-driven routing**
   - Rationale: centralizes prompt/export behavior so adding new context types doesn’t require editing many switch statements.

3) **No plugin-defined prompt variables (initially)**
   - Rationale: stabilizes the prompt contract; providers can route into existing lanes (`.commits`, `.context`, `.diff`).

4) **Session stores references, not blobs**
   - Rationale: preserves reproducibility and reviewability; avoids stale or enormous session files.

5) **Controller remains the canonical request builder**
   - Rationale: exports/token-count/generate stay aligned and deterministic.

## Alternatives Considered

1) **Keep hard-coded injection logic in controller**
   - Rejected: does not scale; increases coupling and review surface area.

2) **Move derivation into prompt templates**
   - Rejected: templates should be pure rendering; git operations don’t belong there.

3) **Runtime dynamic plugins (`plugin` package)**
   - Rejected (for now): operational complexity and limited platform support; makes reproducibility harder.

4) **Generic “context item schema” with free-form YAML**
   - Rejected: loses type safety and discoverability; harder to validate and document; pushes complexity to runtime.

## Implementation Plan

1) Introduce provider interface + provider runner in controller; port git_history and git_context to providers (no behavior change).
2) Introduce context type registry; refactor prompt + export routing to consult the registry.
3) Add a “provider list” debug export (optional) to show which providers ran and what they produced.
4) Document provider authoring guidelines (how to define new derived context sources safely).
5) Optional: support out-of-tree providers via a “plugin pack” module that imports provider packages to register them.

## Open Questions

1) Should the provider `Enabled` decision be purely config-driven (session.yaml), or can it depend on repo state (e.g., “only if diff > N”)?
2) Should we standardize a single envelope format for derived context items (XML-ish only), or allow provider-specific envelopes?
3) Do we want a stable ordering model (priority numbers / phases) across providers?
4) How should providers surface errors: fail the request, or attach an “error context item” with diagnostics?

## References

- `ttmp/2026/01/03/001-ADD-GIT-HISTORY--add-git-history-section-to-pr-session-context/analysis/02-architecture-current-structure-and-modularization-opportunities.md`
- `internal/controller/controller.go`
- `internal/api/prompt.go`
- `internal/export/context.go`
- `internal/session/session.go`
