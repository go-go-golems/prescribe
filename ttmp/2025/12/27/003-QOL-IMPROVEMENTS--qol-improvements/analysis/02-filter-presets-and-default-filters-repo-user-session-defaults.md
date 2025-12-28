---
Title: Filter presets and default filters (repo/user/session defaults)
Ticket: 003-QOL-IMPROVEMENTS
Status: active
Topics:
    - prescribe
    - qol
    - docs
    - tokens
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/controller/controller.go
      Note: Prompt preset load/save precedent using .pr-builder/prompts and ~/.pr-builder/prompts
    - Path: internal/domain/domain.go
      Note: Filter/FilterRule domain types and ActiveFilters
    - Path: internal/session/session.go
      Note: Session YAML includes filters and is saved under .pr-builder/session.yaml
    - Path: internal/tui/app/boot.go
      Note: Startup loads session; missing session currently means no defaults
    - Path: internal/tui/app/model.go
      Note: Hardcoded quick filter presets (exclude-tests/docs/only-source)
ExternalSources: []
Summary: Current prescribe supports persisting active filters in per-repo session.yaml and has a few hardcoded TUI presets; it does not support named filter presets (repo/global) nor defaults for new sessions. This doc describes the gaps and what to implement.
LastUpdated: 2025-12-27T16:33:12.260157283-05:00
WhatFor: Answer whether filter presets/defaults are supported today, and outline concrete code changes to add repo/global filter presets and per-repo default filter sets.
WhenToUse: Before implementing filter preset persistence or adding default filters-on-init behavior.
---


## Overview

You asked whether `prescribe` supports:

- **Saving filters to per-repository presets**
- **Saving filters to user config presets**
- **Setting defaults for which filters are applied to new sessions in the current repo**

Short version:

- **Persisting active filters per repo** is supported today via `.pr-builder/session.yaml` (session state).
- **Named filter presets (repo/global)** are **not** supported (only hardcoded “quick presets” exist in the TUI).
- **Default filters applied to new sessions** are **not** supported (if no session exists, no filters are applied).

## Current behavior (what exists today)

### 1) Filters are part of session state (per-repo persistence)

`prescribe` persists filters as part of the session YAML (in the repo):

- Default session path: `session.GetDefaultSessionPath(repoPath)` → `<repo>/.pr-builder/session.yaml`
- TUI startup attempts to load this session; if missing, it does nothing.
- CLI commands often call “load default session if exists” before mutating state.

So, **a repo can have a stable default set of filters** *only* if a session file already exists and is loaded.

### 2) Hardcoded “quick presets” exist (TUI only)

The TUI supports a few built-in preset IDs (e.g. exclude tests/docs, only source). These are **not persisted as named preset files** and are not listable/editable beyond code changes.

### 3) Prompt presets DO support project/global storage (as precedent)

There is a working precedent for project/global presets (but for prompts, not filters):

- project: `<repo>/.pr-builder/prompts/*.yaml`
- global: `~/.pr-builder/prompts/*.yaml`

This is implemented in `controller.LoadProjectPresets`, `controller.LoadGlobalPresets`, and `controller.SavePromptPreset`.

## Answering the 3 bullets

### A) Saving filters to per-repository presets

**Not supported** as “named presets”.

What you *do* have:

- **Per-repo persistence of the currently active filter set** via `.pr-builder/session.yaml`

What’s missing:

- A place like `<repo>/.pr-builder/filters/*.yaml` containing named filter definitions
- Load/list/apply/delete operations
- UI/CLI affordances to “save current filters as preset”

### B) Saving filters to user config presets

**Not supported**.

Missing:

- A global preset directory like `~/.pr-builder/filters/*.yaml`
- Load/list/apply/delete operations

### C) Defaults for which filters are applied to new sessions in the current repo

**Not supported**.

Today, “new session” behavior is:

- TUI boot tries `LoadSession(<repo>/.pr-builder/session.yaml)`:
  - if missing → “skip load”, leaving `ActiveFilters` empty
- `PRData.NewPRData()` initializes `ActiveFilters` to empty

So if no session exists, no default filters are applied.

## What should be done (recommended implementation)

### 1) Add “filter preset” storage mirroring prompt presets

Add directories:

- **Project**: `<repo>/.pr-builder/filters/*.yaml`
- **Global**: `~/.pr-builder/filters/*.yaml`

Define a YAML schema similar to prompt presets (minimal and stable):

```yaml
name: Exclude tests
description: Exclude test files
rules:
  - type: exclude
    pattern: "**/*test*"
  - type: exclude
    pattern: "**/*spec*"
```

Add controller APIs modeled after prompt presets:

- `LoadProjectFilterPresets() ([]domain.Filter, error)` (or `[]domain.FilterPreset`)
- `LoadGlobalFilterPresets() ...`
- `SaveFilterPreset(name, description string, rules []domain.FilterRule, location domain.PresetLocation) error`

Implementation approach:

- Reuse the same “load all YAML from dir” pattern used in `loadPresetsFromDir`
- Derive preset ID from filename (e.g. `backend_only.yaml`) similarly to prompt presets

### 2) Add CLI commands to manage/apply filter presets

Minimal CLI surface (mirroring prompt behavior):

- `prescribe filter preset list [--global|--project|--all]`
- `prescribe filter preset save --name ... [--global|--project] [--from-active|--exclude ... --include ...]`
- `prescribe filter preset apply PRESET_ID`

This can be implemented either in Cobra or the Glazed ported command structure (note: `filter/add.go` in the current tree is already Glazed-based).

### 3) Add per-repo default filter presets for “new session”

Add a small repo config file, for example:

- `<repo>/.pr-builder/config.yaml`

Proposed shape:

```yaml
defaults:
  filter_presets:
    - exclude-tests.yaml
    - exclude-docs.yaml
```

Then apply defaults when:

- TUI starts and session is missing (`bootCmd` sees `os.ErrNotExist`)
- (Optionally) CLI commands that currently do “load session if exists” should apply defaults when session is missing

Important precedence:

- If `session.yaml` exists: it should **win** (explicit state)
- If no session exists: apply repo defaults (and optionally save a new session to make this sticky)

### 4) Decide whether include rules should remain ANDed

Unrelated but relevant: current filter semantics treat multiple include rules as AND (not OR). This strongly impacts “presets” because presets often want OR semantics (e.g. “only go OR ts OR js”).

If you keep the current semantics, document it and encourage `{alt1,...}` patterns.
If you change semantics, do it as a separate, explicit decision (it’s a behavior change).

## Files to review / starting points

- `prescribe/internal/tui/app/boot.go`: “new session” behavior (missing session → no defaults)
- `prescribe/internal/session/session.go`: session YAML includes filters (current persistence mechanism)
- `prescribe/internal/tui/app/model.go`: hardcoded quick presets via `filterPreset(id)`
- `prescribe/internal/domain/domain.go`: `Filter`, `FilterRule`, `ActiveFilters`
- `prescribe/internal/controller/controller.go`: prompt preset load/save precedent (`.pr-builder/prompts`)

