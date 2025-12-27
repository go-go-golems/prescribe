---
Title: Diary
Ticket: 002-MAKE-IT-WORK
Status: active
Topics:
    - tui
    - bubbletea
    - ux
    - refactoring
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-26T19:05:06.273846214-05:00
WhatFor: ""
WhenToUse: ""
---

# Diary

## Goal

Track the step-by-step work for ticket `002-MAKE-IT-WORK`, focusing on:

- adapting `prescribe`’s TUI to bobatea-style Bubbletea architecture,
- capturing the decision process and UX tradeoffs,
- recording failures and “gotchas” as we iterate.

## Context

This ticket builds on the initial import and CLI regrouping. We’re now focusing on making the TUI robust and consistent:

- correct resize behavior
- better keymap/help patterns
- bulk selection helpers (select all / unselect all)
- clipboard export of generated PR context
- transient “help bubble”/toast UX feedback

---

## Step 1: Start a thorough prescribe TUI structure analysis (models, messages, state, control flow)

This step creates a concrete map of the current `prescribe/internal/tui` implementation: what models exist, how they are wired from the CLI, what messages drive control flow, and what state each screen owns. The intent is to make later refactors (bobatea-style keymaps, resize propagation, toasts, bulk selection, clipboard export) safe and reviewable because we’ll know exactly what the current semantics are.

**Commit (code):** N/A — analysis/documentation phase

### What I did
- Created ticket documents:
  - `reference/02-diary.md` (this diary)
  - `analysis/01-prescribe-tui-structure-models-messages-and-control-flow.md` (the deep dive)
- Read the current wiring: `cmd/prescribe/cmds/tui.go` constructs the Bubbletea program with `tui.NewEnhancedModel(ctrl)`.
- Read current models:
  - `internal/tui/model_enhanced.go` (the model used by the CLI)
  - `internal/tui/model.go` (legacy/simple model, not currently used by `prescribe tui`)
  - `internal/tui/styles.go` (central style variables used by both models)
- Read controller boundaries used by the UI:
  - `internal/controller/controller.go`
  - `internal/controller/session.go`

### Why
- We need a faithful snapshot of the current TUI behavior before changing it.
- Many requested features (resize handling, help bubble, clipboard export, select all/unselect all) are cross-cutting; a baseline map prevents “fix one thing, break another”.

### What worked
- Found a very small message set (mostly `tea.KeyMsg`, `tea.WindowSizeMsg`, and one internal completion message), which makes the current control flow easy to reason about.

### What was tricky to build
- The current rendering uses fixed widths (e.g. `PlaceHorizontal(80, ...)`) and fixed separators; this will interact with resize changes and needs careful refactoring planning.

### What warrants a second pair of eyes
- Whether we should preserve the current “auto-save on toggle/filter change” semantics exactly, or introduce an explicit save action with a toast.

### What should be done in the future
- Extend the analysis doc with a “behavioral contract” section: what exactly happens on each key in each screen, and what controller mutations occur.
- Use the analysis doc to plan the bobatea-style component decomposition (list component, footer/help component, toast component).

---

## Step 2: Write a modularization proposal (bobatea-style models + typed messages + layout system)

This step turns the raw “what exists today” analysis into an actionable architecture proposal. The goal is to make the upcoming feature work (resize correctness, help bubble/toast, select-all/unselect-all, clipboard export) low-risk by introducing clear boundaries: a root orchestrator model, child component models, typed messages, and a single side-effect boundary.

**Commit (code):** N/A — design phase

### What I did
- Read the domain and session structures to ground the proposal in real invariants:
  - `internal/domain/domain.go` (file identity, filters, context items)
  - `internal/session/session.go` (session YAML schema and apply semantics)
- Wrote the design doc:
  - `design-doc/01-prescribe-tui-modularization-proposal-bobatea-style.md`

### Why
- The current monolithic model makes cross-cutting features brittle.
- A typed-message + component-model architecture aligns with bobatea patterns and enables better testing and UI responsiveness.

### What was tricky to build
- Preserving behavior while changing structure: the proposal explicitly stages the migration so we can keep existing semantics (especially auto-save) until we decide otherwise.

### What warrants a second pair of eyes
- Where we draw the side-effect boundary: the proposal puts persistence/clipboard/generation in the root model; confirm this matches long-term go-go-golems style for small apps like prescribe.

### What should be done in the future
- Convert the proposal phases into implementation tasks and start Phase 1 (new root model + layout plumbing).

---

## Step 3: Start a deep-dive on the core architecture (Controller, domain model, git/session/api)

This step shifts focus from UI to the “engine room”: the Controller and the subsystems it orchestrates. The intent is to produce a didactic core-architecture document that explains the nouns (PRData, filters, context, presets, session YAML), the verbs (initialize, mutate, persist, generate), and the failure modes. This becomes the foundation for refactoring the TUI without breaking correctness or persistence semantics.

**Commit (code):** N/A — analysis/documentation phase

### What I did
- Created a new analysis doc in ticket 002:
  - `analysis/02-core-architecture-controller-domain-model-git-session-api-subsystems.md`
- Read core packages to map real control/data flow:
  - `internal/controller/controller.go` + `internal/controller/session.go`
  - `internal/domain/domain.go`
  - `internal/session/session.go`
  - `internal/git/git.go`
  - `internal/api/api.go` (currently a mock/simulated API)

### Why
- The TUI modularization proposal assumes a clear side-effect boundary and stable IDs; both are determined by core code (paths, session schema, branch checks, etc.).
- The next feature work (export context to clipboard) needs an explicit definition of “context for generation”, which currently spans files, filters, additional context items, and prompt state.

### What worked
- The system is intentionally simple and synchronous: controller orchestrates plain functions; PRData is a single struct; the session schema is straightforward.

### What was tricky to build
- Some boundaries are still “in progress” from the import: there is duplicate `internal/model` that mirrors parts of `internal/domain`, and the prompt preset storage still uses the legacy `.pr-builder` directory name. These will need explicit decisions during modernization.

### What warrants a second pair of eyes
- Whether we should rename `.pr-builder/` to `.prescribe/` as part of ticket 002 (breaking change) or defer it (and document it as legacy).


## Usage Examples

<!-- Show how to use this reference in practice -->

## Related

<!-- Link to related documents or resources -->
