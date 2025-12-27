---
Title: Go-go-golems Bubbletea application guide
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
LastUpdated: 2025-12-26T18:33:55.245410601-05:00
WhatFor: ""
WhenToUse: ""
---

# Go-go-golems Bubbletea application guide

## Goal

This guide teaches you how to build **Bubbletea TUIs in the go-go-golems style**, using **bobatea** as the concrete reference implementation, and **prescribe** as the “current app we’re evolving”.

The intended outcome is that a developer can:

- design a Bubbletea app with a clean mental model (state machine + components + messages),
- keep `Update()` and `View()` readable and testable,
- implement “polish features” (help, resize, selection, clipboard, transient notifications) without turning the model into spaghetti,
- align new TUIs with existing go-go-golems patterns so they look/feel consistent across repos.

## Context

In this repo set we have:

- **bobatea**: a growing “toolkit + examples” repository that codifies patterns for Bubbletea apps (components like diff viewer, file picker, timeline shell, and keymap/help organization).
  - See `bobatea/docs/charmbracelet-bubbletea-guidelines.md` for the baseline conventions.
  - See `bobatea/pkg/diff/*`, `bobatea/pkg/filepicker/*`, `bobatea/pkg/chat/*` for concrete, production-ish models.
- **prescribe**: a Bubbletea app that currently lives in `prescribe/internal/tui/*` and is fairly monolithic: it handles keys via `msg.String()` switches and renders help manually. It also currently uses fixed-width rendering in places.

The feature work planned under ticket **002-MAKE-IT-WORK** (resize correctness, selection helpers, clipboard export, “help bubble”) is exactly the sort of polish that tends to either:

- **snap into place** if the app is structured with component boundaries and message patterns, or
- **explode** into ad-hoc conditionals if everything is a single giant update loop.

This guide is written to push us into the first path.

## Quick Reference

### The go-go-golems Bubbletea mental model (the “nouns and verbs”)

You can usually describe a TUI with 4 nouns and 4 verbs:

- **Nouns**
  - **Root model**: owns the app state machine and cross-cutting UI concerns.
  - **Components**: reusable sub-models (`list`, `viewport`, custom widgets, bobatea components).
  - **Keymap**: declarative bindings + help strings.
  - **Styles**: a palette in one place (Lipgloss).
- **Verbs**
  - **Update routing**: decide which component gets the message.
  - **Layout**: compute sizes and propagate `SetSize(...)` (especially on `tea.WindowSizeMsg`).
  - **Side effects**: IO happens via `tea.Cmd` and returns typed messages back into Update.
  - **Rendering**: assemble header/body/footer; compute heights based on rendered content if needed.

### Recommended project layout (pinocchio/bobatea-inspired)

This is a *pattern*, not a hard rule. The point is: keep keymaps/styles separate, keep screen models small.

```text
cmd/<app>/...            # wiring only (cobra -> tea.NewProgram -> root model)
internal/tui/
  root_model.go          # the root state machine and routing
  screens/               # optional: one file per screen/state
    main.go
    filters.go
    result.go
  components/            # custom widgets or thin wrappers around bubbles/bobatea
    file_list.go
    help_toast.go
  keys/                  # keymaps, implements bubbles/help.KeyMap
    keymap.go
  styles/                # centralized lipgloss styles / palette
    styles.go
```

**Concrete reference:** bobatea’s guideline doc explicitly recommends splitting `keys`, `view/styles`, “one concept per file” and using `bubbles/help` (`bobatea/docs/charmbracelet-bubbletea-guidelines.md`).

### Resize handling: the “must-do” checklist

On every `tea.WindowSizeMsg`, do *all* of these:

- Store `m.width`, `m.height` on the root model.
- Recompute layout for header/body/footer.
- Set child component sizes (lists, viewports, textareas).
- If you use a viewport, refresh its content after resizing (bobatea does this in several models).

**Concrete references:**

- `bobatea/pkg/diff/model.go`: `computeLayout()` + `applyContentSizes()` after `tea.WindowSizeMsg`.
- `bobatea/pkg/repl/model.go`: resizes input + timeline shell and triggers refresh.
- `bobatea/pkg/chat/model.go`: sets size then calls `recomputeSize()` which measures header/help heights.

### Selection: model it as a set, not as indices

If your list is filterable/sortable (it almost always becomes that), don’t store selection as “indices”.

- Prefer `map[StableID]bool` (e.g. file path string → selected).
- Derive list row rendering from that map.
- “Select all / deselect all” becomes trivial, and survives filtering.

**Concrete reference:** `bobatea/pkg/filepicker/filepicker.go` uses `multiSelected map[string]bool` and has dedicated bindings for `SelectAll`/`DeselectAll`.

### Clipboard: make it a side-effect requested by the UI, executed at the top

The clean pattern is:

- Leaf component emits “copy requested” message (typed).
- Root model performs the actual clipboard write.
- Root model emits a transient “copied” notification (toast/help bubble).

**Concrete reference:** `bobatea/pkg/chat/model.go` handles `timeline.CopyTextRequestedMsg` / `timeline.CopyCodeRequestedMsg` by calling `clipboard.WriteAll(...)` in the root `Update`.

### Transient status (“help bubble” / toast)

Implement this as a small, local “toast state machine” in the root model:

- Set toast text + expiration timestamp
- Return a `tea.Tick(duration, func(...) tea.Msg { return toastExpiredMsg{} })`
- On toastExpiredMsg, clear toast

This avoids goroutine mutation and makes it testable.

### Bubbletea + bobatea “house style” (what makes apps feel consistent)

These are the conventions that show up across bobatea components and demos:

- **Keymaps are declarative** (via `bubbles/key`) and can drive help output:
  - per-component bindings (`bobatea/pkg/diff/keymap.go`)
  - mode-aware bindings (chat uses mode tags and a mode-enabler, `bobatea/pkg/chat/keymap.go`)
- **Layout is explicit**:
  - store width/height on the model
  - compute `headerHeight`, `footerHeight`, `bodyHeight` by measuring the rendered header/footer when needed (`bobatea/pkg/diff/model.go`, `bobatea/pkg/chat/model.go`)
  - apply sizes to child bubbles using frame sizes (`lipgloss.Style.GetFrameSize()`) rather than “guessing” (diff does this in `applyContentSizes()`)
- **Side-effects are centralized**:
  - leaf models request actions via typed messages (e.g. copy-to-clipboard requests)
  - the “top” model executes OS-level effects (`bobatea/pkg/chat/model.go` writes to clipboard in response to timeline copy requests)
- **Selection uses stable IDs**:
  - store selected items in a set map (`map[string]bool`)
  - implement select-all/deselect-all easily (`bobatea/pkg/filepicker/filepicker.go`)

## Usage Examples

### Example 1: A minimal “help bubble” toast (copy/paste template)

```go
// msgs
type toastMsg struct {
	Text     string
	Duration time.Duration
}

type toastExpiredMsg struct{}

// state
type toastState struct {
	text      string
	deadline  time.Time
	visible   bool
}

func showToast(text string, d time.Duration) tea.Cmd {
	return func() tea.Msg { return toastMsg{Text: text, Duration: d} }
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch v := msg.(type) {
	case toastMsg:
		m.toast.text = v.Text
		m.toast.deadline = time.Now().Add(v.Duration)
		m.toast.visible = true
		return m, tea.Tick(v.Duration, func(time.Time) tea.Msg { return toastExpiredMsg{} })
	case toastExpiredMsg:
		// Only clear if our deadline has truly passed (handles overlapping toasts).
		if m.toast.visible && time.Now().After(m.toast.deadline) {
			m.toast.visible = false
			m.toast.text = ""
		}
		return m, nil
	}
	return m, nil
}
```

**Implication:** this composes cleanly with clipboard actions, saving, “N selected”, etc.

### Example 2: A bobatea-style keymap + help integration

If you want discoverable shortcuts, use `bubbles/key` bindings and satisfy the `help.KeyMap` interface.

Reference implementation:

- `bobatea/pkg/diff/keymap.go` implements `ShortHelp()` and `FullHelp()`
- `bobatea/pkg/filepicker/filepicker.go` stores `help help.Model` and renders different help per mode
- `bobatea/pkg/chat/keymap.go` uses mode tags and a help key for toggling help

Pattern sketch:

```go
type KeyMap struct {
	Quit key.Binding
	Help key.Binding
	SelectAll key.Binding
	DeselectAll key.Binding
	Copy key.Binding
}

func (k KeyMap) ShortHelp() []key.Binding { return []key.Binding{k.Help, k.Copy, k.Quit} }
func (k KeyMap) FullHelp() [][]key.Binding { return [][]key.Binding{{k.SelectAll, k.DeselectAll}, {k.Copy, k.Quit}} }
```

Then:

- Root model owns `help help.Model`
- Root `View()` prints `help.View(keyMap)` at the bottom

### Example 3: Applying these patterns to prescribe (what to change conceptually)

Prescribe today:

- handles `tea.WindowSizeMsg` but does not recompute layout or propagate sizes (`prescribe/internal/tui/model.go`, `model_enhanced.go`)
- renders fixed-width blocks (hardcoded `80` and separators) and manually prints help strings
- selection is an index (`selectedIndex`) rather than a set keyed by file path

To implement ticket 002 features in a go-go-golems way:

- **Resize correctness**
  - replace fixed widths with `m.width` and use `lipgloss.Width/Height`-based layout like bobatea’s diff/chat models
  - factor layout into `computeLayout()` + `applyContentSizes()` methods
- **Select all / unselect all**
  - add `selected map[string]bool` keyed by file path (or reuse `Included` as the “selected” bit)
  - bindings: `a` select all, `A` deselect all (mirroring `bobatea/pkg/filepicker/filepicker.go`)
- **Export context to clipboard**
  - root model emits a `copyContextRequestedMsg` when user presses key
  - side-effect executes clipboard write at the top level (like bobatea chat does)
  - show toast “Copied context (N chars)”
- **Help bubble**
  - implement toast state machine and render it next to help footer (or above it)

## Design deep dive (how to think about a real app)

This section is the “mental model first” part: it’s the reasoning you want before you start coding.

### 1) Treat the TUI as a state machine

If you can’t name your states, you will end up with boolean soup.

Prescribe currently models screens as:

- `ScreenMain`, `ScreenFilters`, `ScreenGenerating`, `ScreenResult` in `prescribe/internal/tui/model_enhanced.go`

That’s a good start, but the next step is to make sure **each state has**:

- a clear **input routing** rule (which keys are handled in this state?),
- a clear **view composition** (which components render?),
- explicit **transitions** (what messages cause state changes?).

Bobatea chat is a strong reference here: it has explicit states (`StateUserInput`, `StateMovingAround`, etc.) and changes the enabled keybindings when state changes (see `bobatea/pkg/chat/model.go` and `bobatea/pkg/chat/keymap.go`).

### 2) Build “UI components” that can be resized and embedded

In go-go-golems TUIs, we treat many UI pieces as reusable components (either directly from `bubbles`, or wrappers):

- lists (bubbles `list.Model`, or custom listbox)
- viewports (bubbles `viewport.Model`)
- text inputs / textareas (bubbles `textinput`, bobatea’s memoized `textarea`)
- “panes” (styled containers with borders, focus indication)

The critical interface for embeddability is:

- each component must be able to accept size changes (`SetSize` or by updating fields when it receives `tea.WindowSizeMsg`)
- each component must not assume fixed widths

Bobatea diff is the canonical pattern: it stores `width/height`, computes `leftWidth/rightWidth/bodyHeight`, then calls `SetSize` on the list/detail widgets (`bobatea/pkg/diff/model.go`).

### 3) Prefer “layout by measurement”, not hard-coded heights

For non-trivial screens, header and footer height often depends on:

- whether a search box is visible,
- whether help is short vs full,
- whether a toast/notification is visible.

Bobatea diff computes header/footer height by rendering them and measuring `lipgloss.Height(...)` (`bobatea/pkg/diff/model.go`).
Bobatea chat’s `recomputeSize()` does similar measurement for header and help height (`bobatea/pkg/chat/model.go`).

This is the most robust way to “properly handle resize events” because it also handles “UI mode changes that change height”.

### 4) Separate “what happens” from “how it’s shown”

When you add clipboard export, save-to-file, or generate-with-LLM:

- the **trigger** is a keybinding → message
- the **effect** happens in a `tea.Cmd` or in the root update as a response to a typed “requested” message
- the **UX** is a toast/help bubble + maybe a status area

Bobatea uses explicit “requested side effects” messages in its timeline (`timeline.CopyTextRequestedMsg`, etc.) and performs the clipboard write at the top (`bobatea/pkg/chat/model.go`).

This keeps it testable: you can unit-test that a copy-request message is produced without needing a real clipboard.

## Implementation walkthrough (prescribe, guided by the 002 feature list)

This is a practical “if you’re implementing this ticket, here’s the map” section.

### Step 1: Adopt bobatea-style keymap + help rendering

**Why:** manual help strings drift quickly, and you can’t easily vary help per mode/screen.

**Plan shape:**

- Create `internal/tui/keys/keymap.go`:
  - define `type KeyMap struct { ... }`
  - provide `ShortHelp()` / `FullHelp()` so we can plug in `bubbles/help`
- Add `help help.Model` to the root model
- Render `help.View(keyMap)` in the footer

**Concrete inspiration:**

- `bobatea/pkg/filepicker/filepicker.go` embeds `help.Model` and shows full help by mode.
- `bobatea/pkg/diff/keymap.go` shows how to implement `ShortHelp/FullHelp` cheaply.

### Step 2: Make resizing first-class (not an afterthought)

**Why:** Resize bugs are usually “layout computation bugs”. Fix the pattern once, then new features won’t re-break it.

**Pattern:**

- On `tea.WindowSizeMsg`:
  - set width/height
  - call `computeLayout()` (measure header/footer)
  - call `applyContentSizes()` (set child sizes)
  - re-render or refresh viewports as needed

**Concrete inspiration:**

- `bobatea/pkg/diff/model.go`: `computeLayout` and `applyContentSizes`
- `bobatea/pkg/listbox/listbox.go`: updates truncation width on resize

### Step 3: Add select-all / unselect-all

**Why:** Once you have selection as a set, everything else (bulk operations, clipboard export, summary counts) becomes easy.

**Pattern:**

- Track selection in a set:
  - `selected map[string]bool` where key is file path (stable)
- Provide:
  - `selectAllVisible()`
  - `deselectAllVisible()`
- Bind keys:
  - `a` select all (mirrors bobatea filepicker)
  - `A` deselect all
- Render each row based on the set membership.

**Concrete inspiration:**

- `bobatea/pkg/filepicker/filepicker.go`: `SelectAll` / `DeselectAll` bindings, and multi-selection map.

### Step 4: Export context to clipboard

**Why:** It’s a “side-effect + UX” feature; it’s easy to keep clean if you follow the bobatea side-effect pattern.

**Pattern:**

- Define a message:
  - `type CopyContextRequestedMsg struct{ Text string }`
- In the component/screen, emit it when the user presses the binding.
- In the root model, handle it:
  - write to clipboard (`atotto/clipboard` is already used in bobatea)
  - show toast “Copied context”

**Concrete inspiration:**

- `bobatea/pkg/chat/model.go`: handles copy requested messages and calls `clipboard.WriteAll`.

### Step 5: “Help bubble” with display time

**Why:** This gives immediate UX feedback (“Copied”, “Saved”, “Filters cleared”) without modal dialogs.

**Pattern:**

- Implement toast state machine (Example 1 above).
- Render toast near footer help:
  - either on its own line above help
  - or appended to the left side of help output
- Trigger from all side-effects:
  - clipboard copy
  - save
  - bulk selection

**Implementation note:** Use `tea.Tick` (not goroutines) to clear it; bobatea already uses tick scheduling for refresh/polling in multiple places (`bobatea/pkg/repl/model.go`).

## Related

- `bobatea/docs/charmbracelet-bubbletea-guidelines.md` (baseline conventions and patterns)
- `bobatea/pkg/diff/model.go` (layout/resize + split panes + search widget layout)
- `bobatea/pkg/filepicker/filepicker.go` (selection sets + select all/deselect all + help per mode)
- `bobatea/pkg/chat/model.go` and `bobatea/pkg/chat/keymap.go` (help model + side-effect messages like clipboard)
- `prescribe/internal/tui/model_enhanced.go` (current prescribe model we’ll be adapting)
- `prescribe/internal/tui/styles.go` (existing centralized lipgloss styles; likely to evolve into a Styles struct)

## Usage Examples

<!-- Show how to use this reference in practice -->

## Related

<!-- Link to related documents or resources -->
