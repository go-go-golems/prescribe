---
Title: Prescribe TUI modularization proposal (bobatea-style)
Ticket: 002-MAKE-IT-WORK
Status: active
Topics:
    - tui
    - bubbletea
    - ux
    - refactoring
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-26T19:16:25.055333213-05:00
WhatFor: ""
WhenToUse: ""
---

# Prescribe TUI modularization proposal (bobatea-style)

## Executive Summary

We will rework `prescribe`’s Bubbletea TUI from a monolithic, string-rendering model into a **modular, bobatea-style application** composed of multiple Bubbletea models (“components”), with a typed message layer and explicit layout propagation.

This design keeps the current behavior (screens, keys, auto-save semantics) as the baseline contract, but restructures the code so the ticket 002 features become straightforward:

- robust resize handling (true dynamic layout, no hard-coded widths),
- short/full help via `bubbles/help` with a centralized keymap,
- transient “help bubble”/toast feedback with timeouts,
- select all / unselect all,
- export “context for generation” to clipboard.

The critical architectural shift: **UI actions become messages**, and side-effects (save session, generate, copy to clipboard) become **commands returning result messages**, rather than happening inline in the key handler.

## Problem Statement

### What’s wrong with the current TUI (concretely)

Current root model is `internal/tui/model_enhanced.go` (`EnhancedModel`), launched by `cmd/prescribe/cmds/tui.go`.

It has three structural issues that make feature work risky:

- **Monolithic Update/View**:
  - screen switching, list navigation, controller mutation, persistence, and help rendering all live in one type.
  - adding features multiplies conditional complexity quickly.
- **No real layout system**:
  - the model stores `width/height` but most rendering uses fixed widths (`PlaceHorizontal(80)`, separators `78`, result box `74`).
  - `tea.WindowSizeMsg` does not recompute layout or resize any components.
- **Unstructured side-effects**:
  - the model mutates controller state and saves sessions inline in key handlers.
  - there’s no consistent “UX feedback channel” for side-effects (toast), and no single place that orchestrates them.

### Why “just refactor later” is not enough

Ticket 002 explicitly requires cross-cutting UX improvements (resize correctness, help bubble, clipboard export). These features touch:

- key handling,
- rendering/layout,
- controller interactions,
- persistence and errors.

Without modularization, each change will entangle with every screen.

## Proposed Solution

### Design goals

- **Modularity**: separate models per screen/component; root routes messages.
- **Typed messages**: UI actions and side-effect results are explicit types.
- **Single side-effect boundary**: root handles OS/IO actions (clipboard, save, generate).
- **Layout correctness**:
  - root computes layout on `tea.WindowSizeMsg` and mode changes,
  - child models receive sizes or `tea.WindowSizeMsg` and adapt,
  - no fixed-width rendering.
- **Preserve current semantics** (at least initially):
  - keys still work,
  - auto-save still happens on toggle/filter changes (unless we decide otherwise),
  - generation errors still show as error UI.

### Ground truth constraints (from domain/controller/session)

From `internal/domain/domain.go` and `internal/session/session.go`:

- Stable file identity is the **file path** (`domain.FileChange.Path`).
- Visible/filtered file lists are derived from `PRData.ChangedFiles` via active filters.
- Session persistence is “branch-bound”: `Controller.LoadSession` rejects sessions whose `SourceBranch` does not match current branch.
- Generation requires **at least one included visible file**; otherwise `GenerateDescription` returns error.

This implies:

- Selection should be keyed by file path (stable) rather than list indices.
- UI should treat save/load/generate as operations that can fail and should produce user-visible feedback.

### Proposed code organization (go-go-golems / bobatea style)

We keep the existing module `prescribe/internal/tui`, but introduce subpackages to avoid one giant file.

```text
prescribe/internal/tui/
  app/                      # root model and global orchestration
    model.go                # type Model struct{...} Init/Update/View
    layout.go               # computeLayout/applySizes helpers
    messages.go             # app-level messages (typed)
  components/
    filelist/               # list of files with multi-select + toggle included
      model.go
      keymap.go
      types.go              # item types, stable IDs
    filterpane/             # filters list + presets + delete/clear
      model.go
      keymap.go
    result/                 # generated description viewport
      model.go
      keymap.go
    status/                 # footer/help + toast (“help bubble”)
      toast.go
      statusbar.go
  keys/
    keymap.go               # app keymap (ShortHelp/FullHelp) + mode toggles
  styles/
    styles.go               # central palette + style struct
```

Notes:

- Components are plain Bubbletea models: each can handle `tea.WindowSizeMsg` and `tea.KeyMsg`.
- App-level keymap is the “global contract” (quit/help, mode switching, copy context).
- Per-component keymaps handle local navigation (up/down, select, etc.).

### Proposed model hierarchy (composition)

Root model (single Bubbletea `tea.Model`) composes children:

- `FileListModel` (left/main pane)
- `FilterPaneModel` (when in filter mode, or as a modal)
- `ResultModel` (when in result mode; uses viewport)
- `HelpModel` (`bubbles/help.Model`) + `ToastModel` (status area)

Conceptual diagram:

```text
AppModel (root)
  - controller: *controller.Controller
  - mode: ModeMain | ModeFilters | ModeGenerating | ModeResult | ModeError
  - layout: Layout{w,h, headerH, bodyH, footerH, ...}
  - fileList: filelist.Model
  - filters:  filterpane.Model
  - result:   result.Model
  - help:     help.Model
  - toast:    status.ToastState
```

### Controller lifetime and session loading (one-time boot)

The modular TUI should follow the same lifetime model as the current `prescribe tui` command: **create one controller, keep it for the whole program**.

That controller is the *single source of truth* for the whole app run:

- `Controller.data` (`*domain.PRData`) is the canonical mutable state during the TUI session.
- UI components never create controllers; they only emit intent messages.

Session loading also becomes a **single, explicit boot step** (not a helper sprinkled throughout the codebase):

1) `ctrl := controller.NewController(repoPath)`
2) `ctrl.Initialize(targetBranch)`  // loads branches + changed files
3) `ctrl.LoadSession(ctrl.GetDefaultSessionPath())`
   - ignore “file not found”
   - surface other errors (especially branch mismatch) via toast/error UI

This removes the need for repeated “load default session” logic in subcommands and components.

### Messages: what we will make explicit

We define typed messages in `internal/tui/app/messages.go` (and some in component packages).

#### “Intent” messages (user actions)

Produced by key handling in components or root:

- `ToggleIncludeRequested{FilePath string}`
- `SelectAllRequested{Scope Scope}` / `DeselectAllRequested{Scope Scope}`
- `ToggleShowFilteredRequested{}`
- `OpenFiltersRequested{}` / `CloseFiltersRequested{}`
- `GenerateRequested{}`
- `CopyContextRequested{}` (export generation context)

#### “Result” messages (side-effect outcomes)

Produced by `tea.Cmd` functions:

- `SessionSavedMsg{Path string}`
- `SessionSaveFailedMsg{Err error}`
- `DescriptionGeneratedMsg{Text string}`
- `DescriptionGenerationFailedMsg{Err error}`
- `ClipboardCopiedMsg{What string, Bytes int}`
- `ClipboardCopyFailedMsg{Err error}`

#### “UX” messages (toast + timers)

- `ShowToastMsg{Text string, Level ToastLevel, Duration time.Duration}`
- `ToastExpiredMsg{ID int64}` (ID prevents clearing newer toasts)

### Where side-effects live (single boundary)

The root model owns side-effects. The pattern:

1) Component emits an “intent” message, e.g. `ToggleIncludeRequested{Path}`.
2) Root `Update` handles it by:
   - mutating controller/domain state as needed,
   - returning a `tea.Cmd` to persist (`SaveSession`) or to call the LLM (`GenerateDescription`) or to write clipboard,
   - emitting `ShowToastMsg` (“Saved”, “Copied”, “N selected”).
3) Root handles result messages to update state / error UI.

This keeps domain mutation and OS IO in one place, and child models stay UI-only.

### Layout strategy (resize correctness)

We copy the proven bobatea pattern:

- Root handles `tea.WindowSizeMsg`:
  - store width/height,
  - compute header/footer heights by measuring rendered header/footer,
  - compute body height,
  - call `SetSize(w,h)` or update fields on child models.
- Child models:
  - respond to size changes by setting list/viewport widths and recomputing truncation widths.

The key: **layout is recalculated on resize and on mode changes that change header/footer content** (help expanded, toast visible, filter pane opened).

### Selection and “select all / unselect all”

We model selection using stable IDs:

- The stable ID for files is `domain.FileChange.Path`.
- The “included in generation” bit is already `FileChange.Included`.

This means “select all/unselect all” can be implemented as:

- “select all visible”: set `Included=true` for all files in `PRData.GetVisibleFiles()`.
- “unselect all visible”: set `Included=false` for all files in that list.

Important nuance: `GetVisibleFiles()` returns copies; root must update the underlying `PRData.ChangedFiles` slice. We will:

- build an index map `path -> changedFilesIndex` once per refresh,
- apply mutations via those indices, not by searching each time.

### Clipboard export (“export context to clipboard”)

We define “context” as the exact content that `GenerateDescription()` would send:

- included visible file diffs/full files,
- additional context items,
- prompt template.

We’ll implement a pure formatter function (no UI) that returns the text:

- `internal/controller` helper or new package `internal/export`:
  - `BuildGenerationContextText(data *domain.PRData) (string, error)`

Then root handles `CopyContextRequested` by:

- calling the builder
- writing to clipboard
- showing a toast with the byte count

### Help bubble / toast with display time

Root owns toast state:

- `toast{Text, Deadline, ID}`
- `ShowToastMsg` sets it and schedules `tea.Tick(duration, ...)` to emit `ToastExpiredMsg{ID}`.
- `ToastExpiredMsg` clears only if ID matches the current toast (prevents a newer toast from being cleared).

Rendering:

- status area draws help (`help.Model.View(keymap)`) and, if toast visible, draws a short toast line above it or inline.

### Migration path: keep behavior while making structure better

This proposal is explicitly staged to avoid a big-bang rewrite.

## Design Decisions

### 1) Root model owns side-effects (save/generate/clipboard)

**Decision:** child models don’t talk to disk/clipboard/network; they emit typed messages.

**Why:** keeps IO centralized and testable, and prevents “random save calls” from deep inside UI code.

### 2) Stable IDs for selection are file paths

**Decision:** selection state and toggling operations use `FileChange.Path` as the ID.

**Why:** session serialization uses paths; filters reorder lists; indices are not stable.

### 3) Layout is a first-class subsystem

**Decision:** add explicit layout recomputation (`computeLayout/applySizes`) triggered by resize and by UI mode toggles that affect header/footer height.

**Why:** fixes resize issues once and prevents future regressions.

### 4) Preserve current auto-save semantics (initially)

**Decision:** keep auto-save on toggle/filter changes, but execute via commands and surface errors via toast/error mode.

**Why:** minimizes user-visible behavior changes while modularizing; later we can revisit UX.

### 5) Use standard Charmbracelet bubbles where possible

**Decision:** use `bubbles/list` or `viewport` for lists and result viewing, and `bubbles/help` for help.

**Why:** reduces custom rendering and improves behavior (scrolling, truncation, etc.).

## Alternatives Considered

### Alternative A: Keep one model, just “clean it up”

Rejected because it does not solve cross-cutting feature complexity; it only postpones the pain.

### Alternative B: Put controller mutations inside child models

Rejected because it scatters side-effects and makes it very hard to reason about persistence and toasts.

### Alternative C: Rewrite to bobatea’s diff model wholesale

We will borrow patterns, but a wholesale replacement is too risky and would likely change UI semantics unexpectedly.

## Implementation Plan

### Phase 0: Documentation + contracts (done / in progress)

- Baseline TUI structure analysis:
  - `analysis/01-prescribe-tui-structure-models-messages-and-control-flow.md`
- Bobatea-style Bubbletea guide:
  - `reference/01-go-go-golems-bubbletea-application-guide.md`

### Phase 1: Introduce new modular root model (no behavior changes)

- Add `internal/tui/app` root model that:
  - wraps the existing behavior (modes/screens)
  - has layout plumbing and a basic help/footer region
- Keep rendering close to current output, but remove fixed widths.

### Phase 2: Extract file list into a component model

- Implement `components/filelist`:
  - cursor navigation
  - visible/filtered toggle as an input/state
  - emits `ToggleIncludeRequested{Path}`
- Root handles domain mutation + save.

### Phase 3: Extract filter management into a component model

- Implement `components/filterpane`:
  - navigation + delete/clear/presets
  - emits `RemoveFilterRequested{Index}` etc.
- Root handles mutations + save.

### Phase 4: Add result viewport + generating spinner model

- `components/result` uses `viewport.Model` for scrolling generated text.
- `components/status` supports toast and long/short help.

### Phase 5: Add ticket 002 feature work (now easy)

- help bubble/toast
- robust resize propagation (already wired)
- select all / unselect all (set Included for all visible paths)
- export context to clipboard (builder + clipboard write)

### Phase 6: Cleanup

- delete legacy `internal/tui/model.go` if no longer used
- unify style vars into a `Styles` struct and pass it down
- decide fate of `internal/model` (it currently duplicates `internal/domain` concepts)

## Open Questions

- **Autosave UX**: keep forever, or add explicit “save” and make autosave optional?
- **“Filtered files” view semantics**: should toggling include on filtered files be allowed, or should filtered view be read-only?
- **Where to put “context management” UI** (additional context items): separate screen, or a side panel?
- **Clipboard implementation**: do we adopt `atotto/clipboard` (bobatea uses it) or prefer OSC52 fallback? (proposal assumes `atotto/clipboard` for speed.)

## References

- `reference/01-go-go-golems-bubbletea-application-guide.md` (ticket 002)
- `analysis/01-prescribe-tui-structure-models-messages-and-control-flow.md` (ticket 002)
- `bobatea/docs/charmbracelet-bubbletea-guidelines.md` (repo reference)
- Current TUI code:
  - `prescribe/internal/tui/model_enhanced.go`
  - `prescribe/internal/tui/styles.go`
- Domain/controller invariants:
  - `prescribe/internal/domain/domain.go`
  - `prescribe/internal/controller/controller.go`
  - `prescribe/internal/session/session.go`
