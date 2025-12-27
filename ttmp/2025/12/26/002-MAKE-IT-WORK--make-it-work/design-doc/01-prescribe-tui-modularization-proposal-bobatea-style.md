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
RelatedFiles:
    - Path: prescribe/internal/controller/controller.go
      Note: Controller APIs that new app model will orchestrate
    - Path: prescribe/internal/controller/session.go
      Note: Session load/save semantics and branch mismatch behavior
    - Path: prescribe/internal/domain/domain.go
      Note: Domain data/IDs (paths) and filter/token behavior
    - Path: prescribe/internal/session/session.go
      Note: Session YAML schema + default path (.pr-builder/session.yaml)
    - Path: prescribe/internal/tui/model_enhanced.go
      Note: Current monolithic TUI baseline being refactored
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

We keep the existing module `prescribe/internal/tui`, but introduce subpackages with **explicit boundaries**:

- `app` owns orchestration + side-effects
- `components/*` are UI-only models
- `events` is the shared, cycle-free typed message vocabulary
- `layout/keys/styles` are leaf packages used everywhere

```text
prescribe/internal/tui/
  app/                           # root model and global orchestration (side-effect boundary)
    model.go                     # type Model struct{...} Init/Update/View
    boot.go                      # initial load (load session, initial toast, etc.)
    reducer.go                   # message routing + state transitions (pure-ish helpers)
    commands.go                  # tea.Cmd constructors for save/generate/clipboard
    view.go                      # root view composition
    state.go                     # Mode + app state structs
  components/
    filelist/
      model.go                   # bubbles/list-based file list
      item.go                    # list.Item + render delegate
      messages.go                # filelist -> events (or local msgs) emission types
    filterpane/
      model.go                   # list of filters + preset actions
      item.go
      messages.go
    result/
      model.go                   # viewport for generated description
      messages.go
    status/
      model.go                   # bubbles/help + toast line
      toast.go                   # toast state machine (tick/expire)
  events/                        # shared message types (avoids import cycles)
    events.go                    # intent + result + UX messages
  layout/
    layout.go                    # Layout struct + Compute() + helpers
  keys/
    keymap.go                    # centralized keymap implementing bubbles/help.KeyMap
  styles/
    styles.go                    # Styles struct (wraps lipgloss styles; no globals)
  export/
    context.go                   # Build "context for generation" text (clipboard/export)
```

#### Package/API boundary rules (non-negotiable)

- `components/*` **must not import** `controller` or `session` or do disk/network/clipboard IO.
- `app` is the only package that may call:
  - `ctrl.SaveSession(...)`
  - `ctrl.GenerateDescription()`
  - clipboard writes
- Cross-cutting messages go into `events/` so that `app` can import components and components can import `events` without cycles.

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

Session loading becomes a **single, explicit boot step** owned by the app model (not helpers sprinkled throughout the codebase).

We will keep the CLI behavior (“try load, ignore if missing”), but make it *precise*:

- If the session file is missing: ignore (no toast).
- If the session exists but fails to load/apply (YAML error, branch mismatch, etc.): show a toast and keep running with the freshly initialized session.

Implementation sketch:

```go
// app/boot.go
func bootCmd(ctrl *controller.Controller) tea.Cmd {
  return func() tea.Msg {
    path := ctrl.GetDefaultSessionPath()
    if err := ctrl.LoadSession(path); err != nil {
      // ctrl.LoadSession wraps session.Load errors; errors.Is can still detect os.ErrNotExist.
      if errors.Is(err, os.ErrNotExist) {
        return events.SessionLoadSkippedMsg{Path: path}
      }
      return events.SessionLoadFailedMsg{Path: path, Err: err}
    }
    return events.SessionLoadedMsg{Path: path}
  }
}
```

This removes repeated “load default session” logic in subcommands and components *and* gives us a place to surface important errors.

### AppModel: core state + dependencies (concrete)

The root model is the only stateful orchestrator. Components keep UI state; `ctrl` keeps domain state.

```go
// app/state.go
type Mode int

const (
  ModeMain Mode = iota
  ModeFilters
  ModeGenerating
  ModeResult
)

type Model struct {
  ctrl *controller.Controller

  // deps keep side-effects mockable (clipboard/time) without leaking into components
  deps Deps

  mode Mode

  // view flags
  showFiltered bool
  showFullHelp bool

  // layout
  width, height int
  layout layout.Layout

  // caches for stable-ID mutations
  changedFileIndexByPath map[string]int

  // composed children
  fileList  filelist.Model
  filterPane filterpane.Model
  result    result.Model
  status    status.Model
}
```

`Deps` is intentionally tiny; it exists to keep clipboard/time mockable while keeping `controller` as the real source of truth for domain state.

```go
// app/state.go (or app/deps.go)
type Deps interface {
  Now() time.Time
  ClipboardWriteAll(text string) error
}
```

### Messages: what we will make explicit

We define a small, explicit typed message vocabulary in `internal/tui/events` so that `app` and `components/*` can share message types without import cycles.

Guideline: **components emit intent**, `app` performs side-effects and emits results.

#### Message taxonomy (concrete)

```go
// events/events.go
package events

import "time"

// --- boot/session ---
type SessionLoadedMsg struct{ Path string }
type SessionLoadSkippedMsg struct{ Path string } // e.g. missing file
type SessionLoadFailedMsg struct {
  Path string
  Err  error
}

type SessionSavedMsg struct{ Path string }
type SessionSaveFailedMsg struct {
  Path string
  Err  error
}

// --- intents (user actions) ---
type ToggleFileIncludedRequested struct{ Path string }

// SetAllVisibleIncludedRequested is the canonical "select all/unselect all".
type SetAllVisibleIncludedRequested struct{ Included bool }

type ToggleShowFilteredRequested struct{}
type OpenFiltersRequested struct{}
type CloseFiltersRequested struct{}

type RemoveFilterRequested struct{ Index int }
type ClearFiltersRequested struct{}
type AddFilterPresetRequested struct{ PresetID string }

type GenerateRequested struct{}
type CopyContextRequested struct{}

// --- results (side-effects) ---
type DescriptionGeneratedMsg struct{ Text string }
type DescriptionGenerationFailedMsg struct{ Err error }

type ClipboardCopiedMsg struct {
  What  string
  Bytes int
}
type ClipboardCopyFailedMsg struct{ Err error }

// --- UX (toasts) ---
type ToastLevel int

const (
  ToastInfo ToastLevel = iota
  ToastSuccess
  ToastWarning
  ToastError
)

type ShowToastMsg struct {
  Text     string
  Level    ToastLevel
  Duration time.Duration
}

type ToastExpiredMsg struct{ ID int64 }
```

Notes:

- Components are free to define additional local message types for internal UI mechanics (cursor moves, list selection changes), but **all cross-component intents** should be `events.*Requested`.
- App owns the translation from `events.*Requested` → controller mutation / IO → `events.*Msg` + `events.ShowToastMsg`.

### Where side-effects live (single boundary)

The root model owns side-effects. The pattern:

1) Component emits an “intent” message, e.g. `events.ToggleFileIncludedRequested{Path}`.
2) Root `Update` handles it by:
   - mutating controller/domain state as needed,
   - returning a `tea.Cmd` to persist (`SaveSession`) or to call the LLM (`GenerateDescription`) or to write clipboard,
   - emitting `events.ShowToastMsg` (“Saved”, “Copied”, “N selected”).
3) Root handles result messages to update state / error UI.

This keeps domain mutation and OS IO in one place, and child models stay UI-only.

### Layout strategy (resize correctness)

We copy the proven bobatea pattern, but make it **explicit and testable** via a `layout.Layout` struct.

Root handles `tea.WindowSizeMsg` by recomputing layout and pushing sizes into children.

```go
// layout/layout.go
package layout

type Layout struct {
  Width, Height int

  // "Chrome" heights. These are computed from rendered header/footer, not constants.
  HeaderH int
  FooterH int

  BodyH int

  // Optional: left/right pane splits if we do a two-pane UI later.
  BodyW int
}

// Compute derives a stable Layout from a window size and rendered chrome heights.
func Compute(width, height, headerH, footerH int) Layout {
  bodyH := height - headerH - footerH
  if bodyH < 0 {
    bodyH = 0
  }
  return Layout{
    Width: width, Height: height,
    HeaderH: headerH, FooterH: footerH,
    BodyH: bodyH, BodyW: width,
  }
}
```

Key invariant: **layout is recomputed** on resize *and* on mode/help/toast changes that affect header/footer height (help expanded, toast visible, filter pane opened).

### Selection and “select all / unselect all”

We model selection using stable IDs:

- The stable ID for files is `domain.FileChange.Path`.
- The “included in generation” bit is already `FileChange.Included`.

This means “select all/unselect all” can be implemented as:

- “select all visible”: set `Included=true` for all files in `PRData.GetVisibleFiles()`
- “unselect all visible”: set `Included=false` for all files in that list

Important nuance: `GetVisibleFiles()` returns copies; root must update the underlying `PRData.ChangedFiles` slice. We will:

- build an index map `path -> changedFilesIndex` once after init/load and refresh it after any operation that can change `ChangedFiles`,
- apply mutations via those indices, not by searching each time.

#### API design: avoid “toggle loops” for bulk operations

The current controller API exposes `ToggleFileInclusion(index int) error`, which is awkward for “select all” (we would need to toggle conditionally).

We should add explicit setters (used by both TUI and future scripting):

```go
// internal/controller/controller.go
func (c *Controller) SetFileIncludedByPath(path string, included bool) (changed bool, err error)
func (c *Controller) SetAllVisibleIncluded(included bool) (changedCount int, err error)
```

If we *don’t* add these, the app root can mutate `ctrl.GetData().ChangedFiles[idx].Included` directly (still within the side-effect boundary), but adding the controller helpers makes the intent clearer and reduces UI-specific domain poking.

### Clipboard export (“export context to clipboard”)

We define “context” as the exact content that `Controller.GenerateDescription()` would send to the API:

- included visible file diffs/full files,
- additional context items,
- prompt template.

#### API design: single source of truth for “what generation sees”

Right now, `Controller.GenerateDescription()` builds an `api.GenerateDescriptionRequest` inline. For clipboard export (and later “export context to file”), we should not duplicate that logic in the TUI.

Proposed refactor in `internal/controller`:

```go
func (c *Controller) BuildGenerateDescriptionRequest() (api.GenerateDescriptionRequest, error)
```

- `GenerateDescription()` calls `BuildGenerateDescriptionRequest()` and then executes the API call.
- `CopyContextRequested` calls `BuildGenerateDescriptionRequest()` and then formats it for humans.

#### Formatting (pure function)

We’ll implement a pure formatter (no UI) that returns text:

- `prescribe/internal/tui/export`:
  - `BuildGenerationContextText(req api.GenerateDescriptionRequest) (string, error)`

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

### Phase 0: Documentation + baseline contracts (done / in progress)

- [x] Baseline TUI structure analysis:
  - `analysis/01-prescribe-tui-structure-models-messages-and-control-flow.md`
- [x] Core architecture analysis:
  - `analysis/02-core-architecture-controller-domain-model-git-session-api-subsystems.md`
- [x] Bobatea-style Bubbletea guide:
  - `reference/01-go-go-golems-bubbletea-application-guide.md`

### Phase 1: Add scaffolding packages (compile-only, no UI changes yet)

- [ ] Create `internal/tui/events` with the message taxonomy (intent/result/toast)
- [ ] Create `internal/tui/layout` with `Layout` + `Compute()`
- [ ] Create `internal/tui/keys` with an app-wide keymap implementing `bubbles/help.KeyMap`
- [ ] Create `internal/tui/styles` with a `Styles` struct (stop adding more globals)
- [ ] Create `internal/tui/components/status` with:
  - toast state machine (`ShowToastMsg` + tick + expire)
  - help rendering (`bubbles/help.Model`)

Exit criteria:
- `go test ./...` passes
- no behavior changes yet (still using `internal/tui/model_enhanced.go`)

### Phase 2: Introduce `app.Model` and switch `prescribe tui` to use it (behavior-compatible)

- [ ] Add `internal/tui/app` root model that:
  - owns `Mode` and screen transitions (Main ⇄ Filters ⇄ Result, plus Generating)
  - does boot-time session load via `bootCmd` (ignore missing file, toast on other errors)
  - owns side-effect commands (save/generate/clipboard)
  - recomputes layout on `tea.WindowSizeMsg`
- [ ] Wire `cmd/prescribe/cmds/tui.go` to run the new model
  - add a small real `Deps` implementation (time + clipboard)
- [ ] Preserve current key contract initially:
  - Main: `j/k` or arrows, `space` toggle include, `f` filters, `v` toggle filtered view, `g` generate, `q` quit
  - Filters: `j/k`, `d/x` delete, `c` clear, `1-3` presets, `esc` back
  - Result: `esc` back, `q` quit

Exit criteria:
- No hard-coded widths remain in root rendering (no `PlaceHorizontal(80)` / constant separators)
- Session auto-save behavior remains the same (but executed via `tea.Cmd` with toast on failure)

### Phase 3: Small controller API improvements (reduce UI hacks)

These are tiny, high-leverage helpers used by the new UI flow:

- [ ] Add `Controller.SetFileIncludedByPath(path, included)` and `Controller.SetAllVisibleIncluded(included)`
- [ ] Add `Controller.BuildGenerateDescriptionRequest()` so generation and export share one source of truth

Exit criteria:
- app never needs to “toggle-loop” to implement bulk selection
- clipboard export can be implemented without duplicating generation request construction

### Phase 4: Extract `filelist` component (first real component)

- [ ] Implement `components/filelist` using `bubbles/list` (or a minimal custom list if needed)
  - stable ID = file path
  - renders included/filtered state and token counts
  - emits `events.ToggleFileIncludedRequested{Path}`
  - emits `events.SetAllVisibleIncludedRequested{Included:true|false}` for select-all/unselect-all keys
- [ ] Root becomes the adapter:
  - builds `[]filelist.Item` from `ctrl.GetVisibleFiles()` (or filtered list when `showFiltered`)
  - applies mutations + save

### Phase 5: Extract `filterpane` component

- [ ] Implement `components/filterpane`:
  - list filters (name, description, rules preview)
  - delete/clear actions
  - presets (mapped to concrete `domain.Filter` definitions)
  - emits `events.RemoveFilterRequested{Index}`, `events.ClearFiltersRequested`, `events.AddFilterPresetRequested{PresetID}`
- [ ] Root handles filter mutations + save + toast

### Phase 6: Extract `result` component + generating state

- [ ] Implement `components/result` with `viewport.Model` (scrolling, copy-to-clipboard shortcut)
- [ ] Add explicit generating state (spinner) while `GenerateDescription` runs

### Phase 7: Ticket 002 feature work (now straightforward)

- [ ] Robust resize propagation (already wired by layout + `SetSize`)
- [ ] Toast/help bubble with timeout (`events.ShowToastMsg` + tick)
- [ ] Select all / unselect all (via controller helpers)
- [ ] Export “context for generation” to clipboard:
  - use `Controller.BuildGenerateDescriptionRequest()`
  - format via `internal/tui/export`
  - write via `Deps.ClipboardWriteAll`

### Phase 8: Cleanup + deletion

- [ ] Remove / archive legacy `internal/tui/model.go` and old rendering helpers if unused
- [ ] Consolidate style usage into `styles.Styles` (no new global styles)
- [ ] Decide fate of `internal/model` (duplicate of `internal/domain`) and document/plan its removal

## Open Questions

- **Autosave UX**: keep forever, or add explicit “save” and make autosave optional?
- **“Filtered files” view semantics**: today `showFilteredFiles` is effectively *view-only* (the selection/toggle logic still targets the visible list). In the refactor we should make this explicit: either (A) filtered view is read-only with a clearly separate cursor, or (B) toggling acts on the currently displayed list (filtered vs visible) consistently.
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
