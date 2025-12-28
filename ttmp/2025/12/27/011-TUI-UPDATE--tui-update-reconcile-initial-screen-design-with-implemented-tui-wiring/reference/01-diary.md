---
Title: Diary
Ticket: 011-TUI-UPDATE
Status: active
Topics:
    - tui
    - bubbletea
    - prescribe
    - ux
    - cli
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-27T19:40:42.784552066-05:00
WhatFor: ""
WhenToUse: ""
---

# Diary

## Goal

Keep a **research/implementation diary** while reconciling the original multi-screen TUI design spec with the **currently implemented** TUI in `prescribe/internal/tui/`.

## Context

We have an early “full product” TUI spec in:

- `/home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/ttmp/2025/12/26/001-INITIAL-IMPORT--initial-import-transform-prescribe-into-go-go-golems-application/design/claude-session-TUI-simulation.md`

Now that core functionality exists, this ticket builds a **screen-by-screen wiring map**: models, inputs, Bubble Tea messages/commands, and which CLI verbs feed/launch the TUI.

## Quick Reference

### Ticket workspace

- Ticket: `011-TUI-UPDATE`
- Analysis doc: `analysis/01-tui-wiring-analysis-screens-models-messages-cli-entrypoints.md`
- Diary: `reference/01-diary.md` (this file)

## Usage Examples

N/A (diary only).

## Related

See ticket index for linked artifacts (screenshots, playbooks, original spec).

## Step 1: Create ticket + initial TUI inventory (what exists today)

This step established the working area for the ticket and did the first “reality check”: what the repo *actually implements* vs the older multi-screen spec. The key outcome is that we already have a modular Bubble Tea app with a small state machine (Main/Filters/Generating/Result), plus a dedicated `prescribe tui` command that requires an initialized session.

### What I did
- Created ticket `011-TUI-UPDATE` and created two docs:
  - Analysis doc (screens/models/messages/CLI entrypoints)
  - Diary (this file)
- Collected and related existing artifacts that contain “screenshots” / visual references:
  - `prescribe/TUI-SCREENSHOTS.md`
  - `prescribe/TUI-SCREENSHOTS.pdf`
  - `prescribe/FILTER-TUI-SCREENSHOTS.md`
  - `prescribe/TUI-DEMO.md`
  - `prescribe/PLAYBOOK-Bubbletea-TUI-Development.md`
  - The original spec: `.../claude-session-TUI-simulation.md`
- Located current TUI implementation entrypoints and core files:
  - CLI entrypoint: `prescribe/cmd/prescribe/cmds/tui.go`
  - Root model/state machine: `prescribe/internal/tui/app/state.go`, `model.go`, `view.go`
  - Shared message vocabulary: `prescribe/internal/tui/events/events.go`
  - Key bindings: `prescribe/internal/tui/keys/keymap.go`
  - Components: `prescribe/internal/tui/components/{filelist,filterpane,result,status}/model.go`

### What worked
- Found a clear and relatively clean architecture:
  - `cmds/tui.go` creates/initializes controller + loads session, then runs Bubble Tea with `tea.WithAltScreen()`
  - `internal/tui/app` is the orchestrator (routes keypresses + “intent messages” to controller side effects)
  - `internal/tui/components/*` are UI-only submodels that emit `events.*Requested` messages (cycle-free)

### What I learned
- **Spec vs implementation gap is now concrete**:
  - The old spec describes many screens (edit context window, replace-with-full-file dialog, prompt editor/presets, save dialogs, etc.).
  - The current TUI implements **4 modes**: `ModeMain`, `ModeFilters`, `ModeGenerating`, `ModeResult`, plus a “copy context” action and a footer help/toast system.
- The TUI currently **requires an initialized session**:
  - If the default session is missing, the CLI verb returns an error instructing to run `prescribe session init --save` first.

### What was tricky to build
- N/A (this was an inventory step), but one potential sharp edge showed up:
  - `internal/tui/app/boot.go` comment says “missing file: SessionLoadFailedMsg (TUI requires an initialized session)”; `events.SessionLoadSkippedMsg` exists but is currently unused here.

### What warrants a second pair of eyes
- Confirm we want the strict “must init session first” UX for `prescribe tui` long-term (it’s enforced in `cmds/tui.go` and also echoed in `bootCmd`).

### Code review instructions
- Start at `prescribe/cmd/prescribe/cmds/tui.go` (entrypoint), then follow into `prescribe/internal/tui/app/`.

## Step 2: Trace domain/session/controller invariants that define “screen truth”

This step focused on the “data truth” that each screen is ultimately rendering: `domain.PRData` (visible vs filtered files, included bits, token totals), how sessions serialize/apply that state, and which controller methods the TUI actually calls for side effects.

The key outcome is a crisp, code-backed contract: **generation uses visible+included files only**, token totals include additional context, and loading a session recomputes token counts based on the chosen file mode (diff vs full file).

### What I did
- Read the controller methods the TUI relies on:
  - `Controller.SetFileIncludedByPath`
  - `Controller.SetAllVisibleIncluded`
  - `Controller.BuildGenerateDescriptionRequest`
  - `Controller.GenerateDescription`
  - `Controller.LoadProjectFilterPresets` / `LoadGlobalFilterPresets`
  - `Controller.LoadFilterPresetByID`
  - `Controller.SaveSession` / `LoadSession` / `GetDefaultSessionPath`
- Read the underlying domain/session types:
  - `domain.PRData` and derived lists (`GetVisibleFiles`, `GetFilteredFiles`)
  - session schema (`session.Session`) and apply logic (`ApplyToData`)
- Reviewed the existing “previous TUI work” docs:
  - `002-MAKE-IT-WORK/design-doc/01-prescribe-tui-modularization-proposal-bobatea-style.md`
  - tmux recording harness + playbook: `.../scripts/tui-tmux.sh`, `.../playbooks/02-recording-bubbletea-with-tmux.md`

### Key findings (high signal)
- **Visible vs filtered is the primary view split**:
  - `PRData.GetVisibleFiles()` returns files passing active filters; `GetFilteredFiles()` returns the rest.
  - The Main screen’s `showFiltered` flag is *only* a view toggle; it doesn’t change the underlying “included” bits.
- **Generation input selection is “visible + included” (not “all included”)**:
  - `Controller.BuildGenerateDescriptionRequest()` builds the request from `visibleFiles := c.data.GetVisibleFiles()` and then filters for `Included == true`.
  - If there are no included visible files, it returns `error("no files included for generation")`.
- **Session load is branch-bound**:
  - `Controller.LoadSession()` rejects a session if `sess.SourceBranch != c.data.SourceBranch`.
- **Token counts are kept consistent on session load**:
  - `session.Session.ApplyToData()` recomputes `FileChange.Tokens` based on mode:
    - `diff` ⇒ count tokens on `Diff`
    - `full_*` ⇒ count tokens on the chosen full content (or sum for `both`)
- **Glob matching uses doublestar**:
  - `domain.matchesPattern()` is based on `doublestar.Match(pattern, path)`; invalid patterns fall back to substring match.

### What warrants a second pair of eyes
- `PRData.passesFilters` currently applies each filter’s rules in a strict “AND all rules” way, including `include` rules (“if include and !matches → false”).
  - That makes “include-only” filters behave like “must match every include rule”, which may or may not be intended (many UIs expect include rules to be OR’d).
  - This impacts *what is visible* and therefore *what can be generated* in the TUI.

### What should be done in the future
- When reconciling the original spec’s richer filter editor UI, clarify the intended filter semantics (AND vs OR vs ordered rule evaluation), because it’s a correctness/UX contract that affects every screen.
