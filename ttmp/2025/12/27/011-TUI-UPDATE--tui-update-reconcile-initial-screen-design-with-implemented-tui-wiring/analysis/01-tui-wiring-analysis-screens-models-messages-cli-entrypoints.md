---
Title: 'TUI wiring analysis: screens, models, messages, CLI entrypoints'
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
    - Path: prescribe/cmd/prescribe/cmds/tui.go
      Note: CLI entrypoint launching Bubble Tea app (session requirement
    - Path: prescribe/internal/controller/controller.go
      Note: Controller operations the TUI calls (toggle/include
    - Path: prescribe/internal/controller/session.go
      Note: Session load/save semantics and branch mismatch behavior
    - Path: prescribe/internal/domain/domain.go
      Note: Domain data model (FileChange
    - Path: prescribe/internal/session/session.go
      Note: Session YAML schema + token recomputation on load
    - Path: prescribe/internal/tui/app/model.go
      Note: Root Update() wiring
    - Path: prescribe/internal/tui/app/state.go
      Note: Root TUI state machine (modes
    - Path: prescribe/internal/tui/app/view.go
      Note: Per-mode View() rendering (main/filters/generating/result)
    - Path: prescribe/internal/tui/events/events.go
      Note: Typed message vocabulary between app and components
    - Path: prescribe/internal/tui/keys/keymap.go
      Note: Central key bindings (inputs)
ExternalSources: []
Summary: Screen-by-screen wiring map from the original TUI spec to the currently implemented Bubble Tea TUI in prescribe (models, inputs, messages/commands, and CLI entrypoints).
LastUpdated: 2025-12-27T19:40:42.707991222-05:00
WhatFor: ""
WhenToUse: ""
---


## Goal

Produce an in-depth, code-linked analysis of the TUI described in the original design spec and reconcile it with the **currently implemented** TUI. For each screen we want:

- a “screenshot” reference (ASCII screenshot from spec and/or current screenshot docs)
- what the screen’s **model** is (state fields + where they live)
- what **inputs** it consumes (key bindings and any CLI args/config)
- what **messages** it handles/emits (Bubble Tea message types + `internal/tui/events`)
- how it is wired up (which functions call which; side effects in controller)
- which **CLI verbs** are relevant for launching/feeding the TUI state

## Sources (primary)

- Spec: `prescribe/ttmp/2025/12/26/001-INITIAL-IMPORT--initial-import-transform-prescribe-into-go-go-golems-application/design/claude-session-TUI-simulation.md`
- Current TUI screenshots: `prescribe/TUI-SCREENSHOTS.md` (+ `.pdf`)
- Filter UX screenshots: `prescribe/FILTER-TUI-SCREENSHOTS.md`
- Bubble Tea playbook: `prescribe/PLAYBOOK-Bubbletea-TUI-Development.md`

## Current implementation “spine” (where to start reading code)

### CLI entrypoint

- `prescribe/cmd/prescribe/cmds/tui.go`
  - Builds controller from parsed glazed layers
  - Requires default session exists (`prescribe session init --save` if missing)
  - Starts Bubble Tea program: `tea.NewProgram(app.New(...), tea.WithAltScreen())`

### Root app model + state machine

- `prescribe/internal/tui/app/state.go` defines `Mode`:
  - `ModeMain`, `ModeFilters`, `ModeGenerating`, `ModeResult`
- `prescribe/internal/tui/app/model.go` handles:
  - global key routing / mode switching
  - controller side effects (save session, generate, copy context)
  - synchronization of subcomponents (`filelist`, `filterpane`, `result`, `status`)
- `prescribe/internal/tui/app/view.go` renders per mode.

### UI-only components (emit intent messages)

- `prescribe/internal/tui/components/filelist/model.go`
- `prescribe/internal/tui/components/filterpane/model.go`
- `prescribe/internal/tui/components/result/model.go`
- `prescribe/internal/tui/components/status/model.go`

### Shared message vocabulary + keybindings

- `prescribe/internal/tui/events/events.go` (cycle-free message types)
- `prescribe/internal/tui/keys/keymap.go`

## Screen inventory: spec vs current

### Screens described in the original spec

The spec describes (at least) the following:

- Main screen
- Edit Context window
- Replace-with-full-file dialog
- Filtered out files view
- Edit file filters screen
- Create/Edit custom filter screen
- Add filter rule dialog
- Save filter preset dialog
- Edit prompt template screen
- Select prompt preset screen
- Save prompt preset dialog
- External editor “waiting” screen

### Screens currently implemented (as of this repo snapshot)

From `internal/tui/app/state.go` and screenshot docs:

- Main file list (`ModeMain`)
- Filter management (`ModeFilters`)
- Generating (`ModeGenerating`)
- Result (`ModeResult`)

## Wiring overview (data + message flow)

### Controller as the “domain boundary”

The TUI itself is presentation-layer only: it reads from `ctrl.GetData()` and requests side effects through controller methods (toggle included, save session, add/remove filters, generate).

### Messages

- **Bubble Tea runtime messages**:
  - `tea.WindowSizeMsg` for resize
  - `tea.KeyMsg` for input
- **App internal messages (`internal/tui/events`)**:
  - Intent messages emitted by components (e.g. `ToggleFileIncludedRequested`)
  - Result messages emitted by commands (e.g. `SessionSavedMsg`, `DescriptionGeneratedMsg`)
  - UX messages (toasts)

### Key bindings (current)

Defined in `internal/tui/keys/keymap.go`:

- Global: `q`/`ctrl+c` quit, `?` help, `esc` back
- Navigation: `up`/`k`, `down`/`j`
- Main: `space` toggle included, `v` view filtered, `f` filters, `g` generate, `y` copy context, `a` select all, `A` unselect all
- Filters: `d`/`x` delete filter, `c` clear filters, `1`..`3` apply quick presets

## Per-screen analysis (current implementation)

> NOTE: “Screenshots” below are intentionally anchored to `TUI-SCREENSHOTS.md` / `FILTER-TUI-SCREENSHOTS.md` (current) and the original spec’s ASCII blocks (spec).

### Screen: Main File List (`ModeMain`)

- Screenshot (current): `prescribe/TUI-SCREENSHOTS.md` “Screen 1: Main File List View”
- Screenshot (spec): “Main Screen” in `claude-session-TUI-simulation.md`
- View renderer: `internal/tui/app/view.go` → `renderMain()`
- Submodel: `internal/tui/components/filelist.Model`
- Model state used:
  - `Model.showFiltered`
  - `Model.filelist` selection and items
  - `ctrl.GetData()` (branches, visible/filtered files, token totals, active filter count)
- Inputs:
  - `↑/↓` or `k/j`: list cursor
  - `space`: emits `events.ToggleFileIncludedRequested{Path}`
  - `v`: toggles `showFiltered` (read-only filtered list)
  - `f`: enters `ModeFilters`
  - `g`: enters `ModeGenerating` and kicks off `generateCmd`
  - `y`: runs `copyContextCmd`
  - `a` / `A`: bulk include toggles
- Messages handled (non-exhaustive):
  - `events.ToggleFileIncludedRequested` → `ctrl.SetFileIncludedByPath` → `saveSessionCmd`
  - `events.SetAllVisibleIncludedRequested` → `ctrl.SetAllVisibleIncluded` → `saveSessionCmd`
  - `events.ClipboardCopiedMsg` / `events.ClipboardCopyFailedMsg` → toast

#### “Generate” input contract (this is the important part)

The TUI’s `g` key ultimately calls `Controller.GenerateDescription(ctx)` (via `generateCmd`). That in turn uses `Controller.BuildGenerateDescriptionRequest()` to decide *exactly* what is sent to inference:

Pseudocode (mirrors `internal/controller/controller.go`):

```go
visible := data.GetVisibleFiles()           // filters applied
included := filter(visible, f.Included)     // only included files
if len(included) == 0 { error("no files included for generation") }

req := {
  source_branch: data.SourceBranch,
  target_branch: data.TargetBranch,
  files: included,
  additional_context: data.AdditionalContext,
  prompt: data.CurrentPrompt,
}
```

This is why “filtered view is read-only” matters: filtered files are not part of `visible` and therefore cannot be generated from unless the filters change.

### Screen: Filter Management (`ModeFilters`)

- Screenshot (current): `prescribe/FILTER-TUI-SCREENSHOTS.md` “Filter Management Screen”
- View renderer: `internal/tui/app/view.go` → `renderFilters()`
- Submodel: `internal/tui/components/filterpane.Model`
- Inputs:
  - `↑/↓` or `k/j`: select filter
  - `d`/`x`: delete selected filter (emits `events.RemoveFilterRequested{Index}`)
  - `c`: clear all (emits `events.ClearFiltersRequested{}`)
  - `1`/`2`/`3`: apply quick presets (loaded from preset dirs at runtime)
  - `esc`: back to main
- Messages handled:
  - `events.RemoveFilterRequested` → `ctrl.RemoveFilter(i)` → `saveSessionCmd`
  - `events.ClearFiltersRequested` → `ctrl.ClearFilters()` → `saveSessionCmd`
  - `events.FilterPresetsLoadedMsg` (async during Init) → `Model.filterPresets = …`

#### Quick presets wiring (1/2/3 keys)

- Discovery happens during `Model.Init()` via `loadFilterPresetsCmd(ctrl)`
  - project: `<repo>/.pr-builder/filters`
  - global: `~/.pr-builder/filters`
- Application happens in `Model.Update()` when matching `keymap.Preset{1,2,3}`
  - it resolves the `presetID` through `ctrl.LoadFilterPresetByID(presetID)`
  - then adds it to active filters via `ctrl.AddFilter(domain.Filter{...})`
  - then triggers `saveSessionCmd(ctrl)`

### Screen: Generating (`ModeGenerating`)

- Screenshot (current): `prescribe/TUI-SCREENSHOTS.md` “Generating Screen”
- View renderer: `internal/tui/app/view.go` → `renderGenerating()`
- Trigger:
  - key `g` in main: sets `ModeGenerating`, calls `generateCmd(ctrl)`
- Command:
  - `generateCmd` calls `ctrl.GenerateDescription(context.Background())`
- Completion messages:
  - `events.DescriptionGeneratedMsg` → set result content + enter `ModeResult`
  - `events.DescriptionGenerationFailedMsg` → set `Model.err` + enter `ModeResult`

### Screen: Result (`ModeResult`)

- Screenshot (current): `prescribe/TUI-SCREENSHOTS.md` “Result Screen”
- View renderer: `internal/tui/app/view.go` → `renderResult()`
- Submodel:
  - `internal/tui/components/result.Model` (bubbles `viewport` for scroll)
- Inputs:
  - `esc`: back to main
  - any viewport scroll keys supported by `bubbles/viewport` (only active while in `ModeResult`)

## Spec-to-current reconciliation notes (high signal)

### What’s already implemented (subset of spec)

- File inclusion toggling
- Filter management (including presets from project/global preset dirs)
- Generation + result display
- Session save/load behavior and session requirement
- Copy generation context to clipboard (`y`)

### What the spec describes but is not yet present in this TUI

- Edit Context window + diff/full-file replacement flows
- Prompt template editing + prompt presets UI
- Save dialogs for presets/sessions
- External editor workflow UI

## Relevant CLI verbs (how TUI state is created/modified outside the TUI)

The TUI is not an island: the same underlying session (`.pr-builder/session.yaml`) is manipulated by dedicated CLI verbs. This matters because:

- `prescribe tui` currently **requires** a saved session to exist (explicit UX contract).
- many “spec screens” (prompt editor, context editor, preset save dialogs) exist today as CLI verbs, even if not yet implemented as TUI screens.

### Session lifecycle

- `prescribe session init [--save] [--path PATH]`
  - Initializes controller from git; can apply repo defaults from `.pr-builder/config.yaml` (default filter presets)
  - If `--save`, writes session YAML (default path if `--path` omitted)
- `prescribe session load [PATH]`
  - Loads YAML session into controller (branch-bound)
- `prescribe session save [PATH]`
  - Saves current controller state to YAML; prefers loading existing session first

### Files (inclusion)

- `prescribe file toggle PATH`
  - Toggles `Included` bit for a changed file path in the current session, then saves.

### Filters

- `prescribe filter add --name NAME [--description ...] (--exclude PAT ... | --include PAT ...)`
  - Adds a filter to active filters and saves.
- `prescribe filter remove INDEX_OR_NAME`
  - Removes a filter by index (0-based) or name and saves.
- `prescribe filter clear`
  - Clears all filters and saves.
- `prescribe filter preset list [--project|--global|--all]`
  - Lists filter presets (project/global).
- `prescribe filter preset apply PRESET_ID`
  - Adds a preset to active filters and saves.
- `prescribe filter preset save --name NAME (--project|--global) [--from-filter-index N | --exclude/--include ...]`
  - Saves a new preset file under preset dirs.

### Additional context

- `prescribe context add [FILE_PATH]` **or** `prescribe context add --note "..."` (mutually exclusive)
  - Adds context items and saves; updates total token count.

### Generation

- `prescribe generate`
  - Loads default session if present (and/or `--load-session` from generation layer), allows prompt override/preset selection, then calls controller generate.
  - Export modes:
    - `--export-context`: print “generation context” (no inference)
    - `--export-rendered`: print rendered LLM payload (no inference)

