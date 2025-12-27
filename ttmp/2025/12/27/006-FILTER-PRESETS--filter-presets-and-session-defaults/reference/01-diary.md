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
    - Path: prescribe/internal/controller/filter_presets.go
      Note: Filter preset YAML schema + controller load/save (code commit bc3149d)
    - Path: prescribe/internal/controller/filter_presets_test.go
      Note: Round-trip tests for preset save/load (code commit bc3149d)
    - Path: prescribe/internal/domain/domain.go
      Note: domain.FilterPreset type (code commit bc3149d)
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
