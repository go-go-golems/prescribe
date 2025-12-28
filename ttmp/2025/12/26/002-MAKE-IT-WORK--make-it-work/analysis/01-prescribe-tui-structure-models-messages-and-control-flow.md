---
Title: 'Prescribe TUI structure: models, messages, and control flow'
Ticket: 002-MAKE-IT-WORK
Status: active
Topics:
    - tui
    - bubbletea
    - ux
    - refactoring
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-26T19:05:06.370508085-05:00
WhatFor: ""
WhenToUse: ""
---

# Prescribe TUI structure: models, messages, and control flow

## Executive summary (what exists today)

`prescribe` currently ships **two Bubbletea models** under `prescribe/internal/tui/`:

- `*tui.EnhancedModel` (`internal/tui/model_enhanced.go`): **the model actually launched** by `prescribe tui`.
- `*tui.Model` (`internal/tui/model.go`): a simpler legacy model; currently not referenced by the CLI entrypoint.

Both models are “monolithic”: they own screen state, selection state, and also directly mutate the domain/controller (including auto-saving sessions). They do not compose “child models” (bubbles list/viewport/help) yet; instead they render strings manually and handle keys via `msg.String()`.

The **message set is extremely small**:

- `tea.WindowSizeMsg` updates width/height (but does not recompute layout beyond storing the ints).
- `tea.KeyMsg` drives almost all control flow.
- `generateCompleteMsg` is an internal message emitted by the `generateCmd()` tea command.

This doc maps:

- how the TUI is wired from Cobra → Bubbletea,
- which models exist and how they’re composed,
- which messages exist and which transitions they trigger,
- what state each model owns,
- where domain mutations happen (controller calls), including persistence (auto-save).

## Entry point and wiring

### Cobra → Bubbletea program

The TUI is started by `cmd/prescribe/cmds/tui.go`:

- Create controller via `helpers.NewInitializedController(...)`
- Best-effort load default session via `helpers.LoadDefaultSessionIfExists(ctrl)`
- Create the Bubbletea program: `tea.NewProgram(tui.NewEnhancedModel(ctrl), tea.WithAltScreen())`

This means:

- `EnhancedModel` is the root model.
- The program runs in alt screen mode.
- The controller provided to the model is already initialized (git diff loaded) and may already have a session loaded.

## Models and composition (what models exist and how they relate)

### Model 1: `tui.EnhancedModel` (active model)

**File:** `prescribe/internal/tui/model_enhanced.go`

**Role:** Root Bubbletea model implementing a **screen state machine** with four screens:

- `ScreenMain` (file list)
- `ScreenFilters` (filter management)
- `ScreenGenerating` (blocking “please wait” view)
- `ScreenResult` (generated PR description view)

**Composition:** No child Bubbletea models. Rendering is string-builder + lipgloss.

### Model 2: `tui.Model` (legacy / unused by CLI)

**File:** `prescribe/internal/tui/model.go`

**Role:** Earlier “single-screen” model with two boolean flags:

- `generating` (blocking)
- `generated` (shows result)

This model does not have filter UI or filtered-file toggling.

**Composition:** No child Bubbletea models.

### Shared: `styles.go`

**File:** `prescribe/internal/tui/styles.go`

**Role:** Central Lipgloss styles used by both models.

This is already aligned with the “centralized palette” goal, but it is currently a set of global variables (not a `Styles` struct passed around).

## Message catalog

### External Bubbletea messages

- `tea.WindowSizeMsg`
  - **Observed handling:** both models store `m.width = msg.Width`, `m.height = msg.Height`.
  - **Missing behavior:** no layout recomputation; views still use fixed widths in many places.
- `tea.KeyMsg`
  - **Observed handling:** keys are interpreted via `msg.String()` and conditional on screen/flags.

### Internal application messages

#### `generateCompleteMsg`

**Defined in:** `internal/tui/model.go` (and referenced by `model_enhanced.go` as well)

Fields:

- `description string`
- `err error`

**Produced by:** `generateCmd()` (`tea.Cmd`) which synchronously calls `m.controller.GenerateDescription()` and returns a `generateCompleteMsg`.

Implication:

- Generation happens in the command function (which Bubbletea runs asynchronously relative to UI), but it is still a synchronous call inside that command. UI “spinner” is not animated (it always renders the first rune).

## EnhancedModel: state and transitions (the real behavior today)

### State owned by the model

`EnhancedModel` fields:

- `controller *controller.Controller`: domain/service boundary
- `width, height int`: stored from WindowSizeMsg, not fully used in layout
- `currentScreen Screen`: the state machine control
- `selectedIndex int`: selected row index for file list
- `filterIndex int`: selected row index for filter list
- `err error`: terminal error state (renders error screen)
- `generatedDesc string`: cached result
- `showFilteredFiles bool`: toggles which list is shown on main screen

### ScreenMain: keys and side-effects

Handled in `updateMain`:

- `ctrl+c` / `q`: quit
- `up`/`k`, `down`/`j`: move `selectedIndex` within **visible files** list
- `space`: toggle file inclusion
  - finds selected file by path in `data.ChangedFiles`
  - calls `m.controller.ToggleFileInclusion(i)`
  - immediately calls `m.controller.SaveSession(m.controller.GetDefaultSessionPath())` (auto-save)
- `f`: switch to `ScreenFilters` and reset `filterIndex`
- `v`: toggle `showFilteredFiles`
- `g`: switch to `ScreenGenerating` and return `m.generateCmd()`

### ScreenFilters: keys and side-effects

Handled in `updateFilters`:

- `ctrl+c` / `q` / `esc`: return to `ScreenMain`
- `up`/`k`, `down`/`j`: move `filterIndex`
- `d` / `x`: delete selected filter
  - `m.controller.RemoveFilter(m.filterIndex)`
  - `m.controller.SaveSession(defaultPath)` (auto-save)
  - adjusts `filterIndex` if it’s now out of range
- `c`: clear all filters
  - `m.controller.ClearFilters()`
  - `m.controller.SaveSession(defaultPath)` (auto-save)
- `1`, `2`, `3`: create and add preset filters (hardcoded), then auto-save

### ScreenGenerating: keys and transitions

- Key input is ignored (`Update` returns `m, nil`).
- Only transition is via `generateCompleteMsg`.

### ScreenResult: keys and transitions

Handled in `updateResult`:

- `ctrl+c` / `q`: quit
- `esc`: return to `ScreenMain`

### generateCompleteMsg handling (global transition)

In `Update`, on `generateCompleteMsg`:

- if error: set `m.err` and go back to `ScreenMain`
- else: set `m.generatedDesc`, go to `ScreenResult`

## View composition (EnhancedModel)

EnhancedModel’s `View()` is essentially:

- if `m.err != nil`: show error view (with “Press q to quit”)
- else switch on `currentScreen` and render:
  - `renderMain()`
  - `renderFilters()`
  - `renderGenerating()`
  - `renderResult()`

Key rendering notes:

- Many elements are hardcoded to width `80` and separators `78`.
- The “spinner” is not animated (no tick messages).
- Help is hand-built per screen (strings), not using `bubbles/help`.

## Controller boundary (where domain state is mutated)

The model directly calls controller methods on user key actions:

- session persistence:
  - `SaveSession(GetDefaultSessionPath())` happens on many user actions (toggle, add/remove filter)
- file inclusion:
  - `ToggleFileInclusion(index)`
- filters:
  - `GetFilters()`, `AddFilter(filter)`, `RemoveFilter(index)`, `ClearFilters()`
- generation:
  - `GenerateDescription()`

This implies the current “UI contract” includes:

- toggles and filter operations are persisted immediately
- generation is based on the controller’s current session state

### Important controller invariants that leak into the UI

These are not “UI code”, but they matter for how the UI behaves and what errors can happen:

- **`Controller.Initialize(target)` must run before most operations**
  - It populates `PRData.SourceBranch`, `PRData.TargetBranch`, and `PRData.ChangedFiles`.
  - The CLI guarantees this by constructing the controller via `helpers.NewInitializedController(...)`.
- **Session load is branch-bound**
  - `Controller.LoadSession(path)` verifies the session’s `SourceBranch` matches the current branch; otherwise it errors.
  - The CLI uses a best-effort load for the default session (`helpers.LoadDefaultSessionIfExists`), and ignores errors.
- **Generation requires “included” files**
  - `Controller.GenerateDescription()` collects *visible* files and then filters to `Included == true`.
  - If none are included, generation errors with `no files included for generation`.
  - `EnhancedModel` currently does not special-case this; it will show an error state.

## Known structure gaps (useful for the upcoming refactor)

This section is intentionally “analysis + implications”, because it’s the most useful part when refactoring.

- **Width/height is stored but not used**: `renderMain()` and friends use fixed widths; resize events do not adapt layout.
- **Selection uses indices**: `selectedIndex` is tied to the currently shown list; toggling between visible/filtered lists risks selection mismatch semantics.
- **No component boundaries**: list rendering, key handling, and persistence are all fused together; adding features like select-all and clipboard export will amplify branching.
- **No transient UX channel**: all side-effects are “silent” except for view changes; this motivates a toast/help bubble pattern.

## Appendix: legacy `tui.Model` (why it exists and how it differs)

`tui.Model` (`internal/tui/model.go`) is a smaller predecessor of `EnhancedModel`.

Notable differences:

- It has no filter screen, no filtered-file viewing toggle, and no filter presets.
- It uses booleans (`generating`, `generated`) instead of an explicit `Screen` enum.
- It defines `generateCompleteMsg` (also used by `EnhancedModel`).

This matters for future refactors because:

- we can delete it once we’re confident `EnhancedModel` (or its successor) is the only root model,
- or we can repurpose it as a small “demo model” if we want to keep a minimal baseline.

