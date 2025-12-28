---
Title: 'Filter management screens: spec vs implementation'
Ticket: 011-TUI-UPDATE
Status: active
Topics:
    - tui
    - bubbletea
    - prescribe
    - ux
    - cli
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
    - Path: prescribe/internal/controller/filter_presets.go
      Note: SaveFilterPreset()
    - Path: prescribe/internal/domain/domain.go
      Note: Filter
    - Path: prescribe/internal/tui/app/view.go
      Note: renderFilters() - Filter management screen
    - Path: prescribe/internal/tui/components/filterpane/model.go
      Note: Filter pane component with rule preview
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-27T20:00:00-05:00
WhatFor: ""
WhenToUse: ""
---


# Filter Management Screens: Spec vs Implementation

This document analyzes the filter management screens from the original TUI spec (`claude-session-TUI-simulation.md`) against the current implementation.

## Screens Analyzed

5. **Edit File Filters Screen** (Spec §5)
6. **Create/Edit Custom Filter Screen** (Spec §6)
7. **Add Filter Rule Dialog** (Spec §7)
8. **Save Filter Preset Dialog** (Spec §8)

---

## 5. Edit File Filters Screen

### Spec Requirements

**Layout:**
- Title: "EDIT FILE FILTERS"
- Stats: Active filters count, files matched, filtered out count
- **QUICK FILTERS** section: Checkboxes for built-in presets (Exclude tests, Exclude docs, Only Go files, etc.)
- **SAVED PRESETS** section: List of project/global presets with location badges `[PROJECT]` / `[GLOBAL]`
- **SESSION FILTERS** section: Unsaved filters with `[TEMP] 💾` indicator
- Actions: `↑↓ Navigate Space Toggle N New E Edit D Delete W Save Preset Enter Apply Esc Cancel`

**Data:**
- Quick filters with pattern descriptions
- Saved presets with name, description, rule count, matched files, location
- Session filters (temporary, unsaved)
- Selected filter index

### Current Implementation

**Location:** `internal/tui/app/view.go::renderFilters()`

**What Exists:**
- ✅ Title: "FILTER MANAGEMENT"
- ✅ Stats: `Active Filters: X | Filtered Files: Y | Files: Z visible`
- ✅ Filter list: Uses `filterpane.Model` (bubbles/list) showing active filters
- ✅ Rule preview: Selected filter shows rules below list (bounded by height)
- ✅ Quick presets: Shows first 3 presets as `[1] Name [2] Name [3] Name`
- ✅ Preset application: `1`/`2`/`3` keys apply presets via `applyQuickPresetByIndex()`
- ✅ Navigation: `↑↓` / `j/k` keys
- ✅ Delete: `D` / `X` keys remove selected filter
- ✅ Clear all: `C` key clears all filters
- ✅ Back: `Esc` returns to main

**What's Missing:**
- ❌ **Quick filters section**: No checkboxes for built-in presets (Exclude tests, Exclude docs, etc.)
- ❌ **Preset location badges**: No `[PROJECT]` / `[GLOBAL]` / `[TEMP]` indicators
- ❌ **Session vs saved distinction**: No separate sections for saved presets vs session filters
- ❌ **Toggle filter on/off**: No `Space` to enable/disable filters (filters are always active when added)
- ❌ **Edit filter**: No `E` key to edit selected filter
- ❌ **Save as preset**: No `W` key to save current filter as preset
- ❌ **New custom filter**: No `N` key to create new filter
- ❌ **Apply action**: No `Enter` to "apply" (filters apply immediately on add)

**CLI Equivalents:**
- `prescribe filter add --name X --exclude "pattern"` (add filter)
- `prescribe filter remove <index-or-name>` (remove filter)
- `prescribe filter clear` (clear all)
- `prescribe filter preset apply <preset-id>` (apply preset)
- `prescribe filter preset list` (list presets)
- `prescribe filter preset save --name X --project/--global` (save preset)

**Code References:**
- Model: `internal/tui/app/state.go::Model` (ModeFilters)
- View: `internal/tui/app/view.go::renderFilters()`
- Filter pane: `internal/tui/components/filterpane/model.go`
- Preset loading: `internal/tui/app/filter_presets.go::loadFilterPresetsCmd()`
- Controller: `internal/controller/controller.go::AddFilter()`, `RemoveFilter()`, `ClearFilters()`

---

## 6. Create/Edit Custom Filter Screen

### Spec Requirements

**Layout:**
- Title: "CREATE CUSTOM FILTER" / "EDIT CUSTOM FILTER"
- Filter name input field
- Description input field
- **RULES** section: Ordered list with `[INCLUDE]` / `[EXCLUDE]` labels and patterns
- **Preview** section: Shows matched/unmatched files with reasons
- Actions: `↑↓ Navigate Enter Edit +Add -Delete Shift+↑↓ Reorder S Save & Use W Save as Preset Esc Cancel`

**Data:**
- Filter name, description
- Ordered rules (type, pattern, order)
- Preview: matched count, example files with inclusion status and reasons

### Current Implementation

**What Exists:**
- ✅ Domain model: `domain.Filter` with `Name`, `Description`, `Rules[]`
- ✅ Filter rules: `domain.FilterRule` with `Type` (include/exclude), `Pattern`, `Order`
- ✅ Controller: `AddFilter(filter)` accepts full filter structure
- ✅ Test filter: `Controller.TestFilter()` returns matched/unmatched files

**What's Missing:**
- ❌ **Screen**: No dedicated Create/Edit Custom Filter screen
- ❌ **Form inputs**: No name/description input fields
- ❌ **Rule editor**: No UI to add/edit/reorder rules
- ❌ **Preview**: No real-time preview of matched files
- ❌ **Rule reordering**: No `Shift+↑↓` to reorder rules

**CLI Equivalents:**
- `prescribe filter add --name X --description Y --exclude "pattern1" --include "pattern2"` (creates filter with multiple rules)

**Code References:**
- Domain: `internal/domain/domain.go::Filter`, `FilterRule`
- Controller: `internal/controller/controller.go::AddFilter()`, `TestFilter()`
- Filter preset save: `internal/controller/filter_presets.go::SaveFilterPreset()`

---

## 7. Add Filter Rule Dialog

### Spec Requirements

**Layout:**
- Title: "ADD FILTER RULE"
- Radio buttons: Include / Exclude
- Glob pattern input field
- Examples section showing pattern syntax
- Actions: `Tab Toggle Type Enter Confirm Esc Cancel`

**Data:**
- Rule type (include/exclude)
- Pattern string
- Editing flag (if editing existing rule)

### Current Implementation

**What Exists:**
- ✅ Domain: `FilterRule` with `Type` (FilterTypeInclude/Exclude) and `Pattern`
- ✅ Pattern matching: `domain.matchesPattern()` uses doublestar glob matching
- ✅ CLI: `--exclude` / `--include` flags accept patterns

**What's Missing:**
- ❌ **Dialog**: No TUI dialog for adding/editing rules
- ❌ **Pattern input**: No text input field for glob pattern
- ❌ **Type toggle**: No UI to switch between include/exclude
- ❌ **Examples**: No inline examples of pattern syntax

**CLI Equivalents:**
- `prescribe filter add --exclude "pattern"` (adds exclude rule)
- `prescribe filter add --include "pattern"` (adds include rule)

**Code References:**
- Domain: `internal/domain/domain.go::FilterRule`, `matchesPattern()`
- Filter syntax docs: `pkg/doc/topics/01-filters-and-glob-syntax.md`

---

## 8. Save Filter Preset Dialog

### Spec Requirements

**Layout:**
- Title: "SAVE FILTER AS PRESET"
- Filter name display
- Radio buttons: Project / Global location
- File path preview for each location
- Actions: `↑↓ Select Location Enter Confirm Esc Cancel`

**Data:**
- Filter name
- Filter rules
- Location options (project/global)
- Selected location index

### Current Implementation

**What Exists:**
- ✅ Controller: `SaveFilterPreset(name, description, rules, location)`
- ✅ Location enum: `PresetLocationProject` / `PresetLocationGlobal`
- ✅ File paths: `.pr-builder/filters/` (project) or `~/.pr-builder/filters/` (global)
- ✅ CLI: `prescribe filter preset save --name X --project/--global`

**What's Missing:**
- ❌ **Dialog**: No TUI dialog for saving presets
- ❌ **Location selection**: No UI to choose project vs global
- ❌ **Path preview**: No display of where preset will be saved

**CLI Equivalents:**
- `prescribe filter preset save --name X --description Y --project` (save to project)
- `prescribe filter preset save --name X --description Y --global` (save to global)
- `prescribe filter preset save --name X --from-filter-index N --project` (save existing filter)

**Code References:**
- Controller: `internal/controller/filter_presets.go::SaveFilterPreset()`
- Domain: `internal/domain/domain.go::PresetLocation`
- CLI: `cmd/prescribe/cmds/filter/preset_save.go`

---

## Summary

### Implemented Features
- Filter list display with rule preview
- Quick preset application (1/2/3 keys)
- Filter deletion and clearing
- Preset loading from project/global directories
- Stats display (active filters, filtered files)

### Missing Features
- **Quick filters section**: No built-in preset checkboxes
- **Filter editor**: No Create/Edit Custom Filter screen
- **Rule editor**: No Add Filter Rule dialog
- **Preset save dialog**: No UI to save filters as presets
- **Filter toggle**: No way to enable/disable filters (they're always active)
- **Rule reordering**: No UI to reorder rules within a filter
- **Preview**: No real-time preview of matched files when editing

### Architectural Notes
- Domain/controller fully support all operations (add, remove, test, save presets)
- TUI shows active filters but doesn't support creating/editing them interactively
- All filter management currently done via CLI commands
- Preset system works (project/global locations, YAML persistence)
- Quick presets (1/2/3 keys) are the only interactive filter creation in TUI
