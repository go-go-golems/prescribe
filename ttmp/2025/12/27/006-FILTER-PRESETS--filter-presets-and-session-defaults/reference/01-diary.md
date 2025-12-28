---
Title: Diary
Ticket: 006-FILTER-PRESETS
Status: active
Topics:
    - prescribe
    - filters
    - qol
    - session
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: prescribe/cmd/prescribe/cmds/filter/filter.go
      Note: Register preset command group under filter (code commit 4880311)
    - Path: prescribe/cmd/prescribe/cmds/filter/preset.go
      Note: Defines filter preset command group (code commit 4880311)
    - Path: prescribe/cmd/prescribe/cmds/filter/preset_apply.go
      Note: Implements prescribe filter preset apply (code commit 4880311)
    - Path: prescribe/cmd/prescribe/cmds/filter/preset_list.go
      Note: Implements prescribe filter preset list (code commit 4880311)
    - Path: prescribe/cmd/prescribe/cmds/filter/preset_save.go
      Note: Implements prescribe filter preset save (code commit 4880311)
    - Path: prescribe/internal/controller/repo_defaults.go
      Note: |-
        Load .pr-builder/config.yaml and apply defaults.filter_presets (code commit cc52899)
        Exports LoadFilterPresetByID used by defaults and CLI (code commit 4880311)
    - Path: prescribe/internal/tui/app/boot.go
      Note: Apply repo defaults when session.yaml missing (code commit cc52899)
    - Path: prescribe/internal/tui/app/boot_test.go
      Note: Boot behavior test for defaults (code commit cc52899)
    - Path: prescribe/internal/tui/app/model.go
      Note: Toast + UI resync handling for defaults/session load (code commit cc52899)
    - Path: prescribe/internal/tui/events/events.go
      Note: New events for default filters applied/failed (code commit cc52899)
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-27T16:58:14.979386247-05:00
WhatFor: ""
WhenToUse: ""
---



# Diary

## Goal

Capture the step-by-step implementation work for `006-FILTER-PRESETS`, including what changed, why, what was tricky, and how to validate.

## Step 1: Add controller-level filter preset load/save (project + global)

This step establishes the foundational persistence layer for filter presets, modeled directly after the existing prompt preset implementation. The output is intentionally “boring”: stable filename-based IDs, and a minimal YAML schema that round-trips cleanly and is easy to reason about.

**Commit (code):** bc3149d5c22d3c6bab3cdcbdebdf73efbe11e101 — "Filters: add controller load/save for filter presets"

### What I did
- Added a `domain.FilterPreset` type to represent a saved preset with a stable ID + location.
- Implemented controller APIs:
  - `LoadProjectFilterPresets()` → `<repo>/.pr-builder/filters/*.yaml`
  - `LoadGlobalFilterPresets()` → `~/.pr-builder/filters/*.yaml`
  - `SaveFilterPreset(...)` writing YAML with `name`, `description`, `rules[{type,pattern}]`
- Added unit tests for project + global save/load.

### Why
- We need the same persistence affordances that prompt presets already have so we can later implement:
  - repo defaults on “new session” boot, and
  - CLI/TUI affordances to list/apply/save presets.

### What worked
- `go test ./...` passes with new tests covering round-trip load/save behavior.

### What was tricky to build
- Keeping YAML schema stable while `domain.FilterRule` does not carry YAML tags (solved by using a dedicated YAML struct in controller, similar to `internal/session`).

### What warrants a second pair of eyes
- **ID strategy**: we currently treat `ID == filename` (including `.yaml`), matching prompt presets. Confirm this is the desired long-term UX for CLI/TUI.

### What should be done in the future
- Add an example preset file in docs/fixtures to make the schema discoverable without reading Go code.
- Build the boot-time default filter application off these APIs (only when session is missing).

### Code review instructions
- Start in `internal/controller/filter_presets.go` and compare it against the prompt preset implementation in `internal/controller/controller.go`.
- Validate with:
  - `cd prescribe && go test ./... -count=1`

### Technical details
YAML schema:

```yaml
name: Exclude tests
description: Exclude test files
rules:
  - type: exclude
    pattern: "**/*test*"
  - type: exclude
    pattern: "**/*spec*"
```

## Step 2: Apply repo-default filter presets on TUI boot when session.yaml is missing

This step wires “defaults for new sessions” into the TUI boot sequence: if the repo has no `.pr-builder/session.yaml`, we look for `.pr-builder/config.yaml` and apply any configured `defaults.filter_presets` into `ActiveFilters`. This keeps session state as the higher-precedence, explicit source of truth while enabling “first run” defaults.

**Commit (code):** cc52899e19f42a044ec00000e55462b2c7a10a5c — "TUI: apply repo default filters when session missing"

### What I did
- Added controller support to read `<repo>/.pr-builder/config.yaml` and apply `defaults.filter_presets` via the filter preset loader.
- Updated `bootCmd` to apply defaults on “missing session” instead of silently skipping.
- Added new TUI events for “defaults applied” / “defaults failed” and toast + UI resync handling.
- Added a unit test that exercises boot behavior end-to-end with a temp repo.

### Why
- “New session has no filters” was making repo defaults impossible unless a session file already existed. This makes defaults opt-in via repo config.

### What worked
- `go test ./...` passes and includes a regression test for boot-time default application.

### What was tricky to build
- Ensuring the UI resyncs file visibility when filters are applied at boot (same issue applies to session load).

### What warrants a second pair of eyes
- Precedence semantics: session load wins; defaults only apply on `os.ErrNotExist` for session.
- Error semantics: a bad config/preset shows a warning toast but does not prevent the app from running.

### What should be done in the future
- Decide whether to auto-save a new session after applying defaults (to make the defaults “sticky” even if config changes).
- Decide whether CLI commands should also apply defaults when session is missing (parity with TUI).

### Code review instructions
- Start in `internal/tui/app/boot.go` and `internal/controller/repo_defaults.go`.
- Validate with:
  - `cd prescribe && go test ./... -count=1`

## Step 3: Add CLI commands for filter presets (list/save/apply)

This step exposes the new filter preset persistence via a minimal CLI surface under `prescribe filter preset ...`. The intent is to unlock scripting and pave the way for TUI affordances, while keeping session semantics intact (load session if present, then mutate, then save).

**Commit (code):** 48803118333e6d0145e965d97b09045344e644bc — "CLI: add filter preset list/save/apply"

### What I did
- Added a `filter preset` command group with:
  - `prescribe filter preset list [--project|--global|--all]`
  - `prescribe filter preset save --name ... (--project|--global) [--from-filter-index N | --exclude ... --include ...]`
  - `prescribe filter preset apply PRESET_ID`
- Exported `(*controller.Controller).LoadFilterPresetByID` so both TUI defaults and CLI can resolve preset IDs consistently.

### Why
- We need a first-class way to manage presets without editing YAML by hand, and to apply presets into the persisted session workflow.

### What worked
- `go test ./... -count=1` passes.

### What was tricky to build
- Keeping CLI repo/target parsing consistent with existing Glazed-based filter commands, while still supporting an arg-based `apply PRESET_ID` subcommand.

### What warrants a second pair of eyes
- Command UX: confirm the `save` command’s `--from-filter-index` workflow is the right minimum viable interface (vs selecting by name / saving multiple active filters).

### What should be done in the future
- Add help/docs examples for common recipes (repo defaults authoring, applying multiple presets).

### Code review instructions
- Start in `cmd/prescribe/cmds/filter/preset*.go` and `cmd/prescribe/cmds/filter/filter.go`.
- Validate with:
  - `cd prescribe && go test ./... -count=1`
  - `cd prescribe && go run ./cmd/prescribe --help | grep -n \"filter preset\"`
