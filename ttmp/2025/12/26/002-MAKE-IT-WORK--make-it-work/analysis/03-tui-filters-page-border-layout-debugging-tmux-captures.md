---
Title: TUI Filters page border/layout debugging (tmux captures)
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
RelatedFiles:
    - Path: internal/tui/app/model.go
      Note: Root layout computation and size propagation (header/footer/body sizing)
    - Path: internal/tui/app/view.go
      Note: Frame rendering; width/height clamping; newline trimming
    - Path: internal/tui/components/filterpane/model.go
      Note: Filter pane height management; rules preview vs list sizing
    - Path: internal/tui/components/status/model.go
      Note: Footer height changes with toast/help; impacts layout
    - Path: internal/tui/styles/styles.go
      Note: BorderBox style and frame sizes; can cause width overflow
    - Path: ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/playbooks/03-smoke-testing-with-tui-tmux-harness.md
      Note: How to run tmux harness and interpret captures
    - Path: ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/scripts/tui-tmux.sh
      Note: Reproduction + capture harness for tmux
ExternalSources: []
Summary: Analysis of a tmux-reproduced border/layout regression on the Filters screen; includes root cause and validation workflow.
LastUpdated: 2025-12-27T00:00:00-05:00
WhatFor: So we can reproduce, reason about, and prevent future border/layout regressions while modularizing the TUI.
WhenToUse: When tmux captures look 'cropped' or borders appear missing; when changing frame sizing, footer/help/toast rendering, or component heights.
---


# TUI Filters page border/layout debugging (tmux captures)

## Context

During ticket `002-MAKE-IT-WORK`, we modularized the TUI (root `internal/tui/app` + child components like `filterpane`). After Phase 5 landed, tmux captures began showing a “missing top border / missing right border” symptom, especially on the Filters screen.

This doc records what we observed, how we reproduced it, and what the likely root causes were.

## Symptoms (as observed in tmux captures)

- **Top border/title appears missing**: captures start with left border (`│`) lines instead of the top-left corner (`╭`) and title.
- **Right border appears missing**: top/bottom border lines end with `─` (no `╮` / `╯`) in captures.
- More noticeable in Filters mode because the screen includes:
  - filter list
  - rules preview block
  - “Quick Add Presets” block
  - toast + help footer

Important: this is often not “border rendering broken”; it can be **terminal overflow** (writing too many rows/cols), which makes tmux captures look cropped.

## Reproduction (tmux harness)

We use the ticket harness to make this repeatable:

```bash
cd /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe && \
./ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/scripts/tui-tmux.sh build && \
./ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/scripts/tui-tmux.sh scenario smoke
```

Captures are written to:

```text
ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/scripts/tui-*.txt
```

To avoid committing captures:

```bash
rm -f ./ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/scripts/tui-*.txt
```

### Capture method note (alt-screen)

Bubbletea draws to the alternate screen. tmux `capture-pane -a -p` can be empty (depending on timing / alt-screen buffer availability). The harness currently tries alt capture, then falls back to normal capture if alt looks blank.

## Root cause hypothesis

### 1) Height overflow (scrolling hides the top border)

If the app writes more than the tmux pane height, tmux shows the “bottom” of the scrollback. In that case, the **top border and title are “missing”** in the capture, but actually scrolled out of view.

Typical causes:
- The body component (filter list / rules preview) is not sized to leave room for:
  - the fixed “Quick Add Presets” block, and/or
  - the status footer (toast + help), whose height changes when help toggles or a toast appears.
- The borderbox render adds padding/border height, but internal content is sized as if it has the whole terminal height.

### 2) Width overflow (right border clipped)

If the app writes more columns than the pane width, the rightmost glyphs (including `╮` / `╯`) can get clipped.

Typical causes:
- Using `lipgloss.PlaceHorizontal(m.width, ...)` inside a container that also has padding/border, making content exceed the available inner width.
- Separator lines like `strings.Repeat("─", m.width-6)` being “close” but not truly derived from the frame’s content width.
- Footer/help line uses width assumptions that don’t match the body width.

### 3) lipgloss width/height semantics

In lipgloss:
- `Style.Width()` / `Height()` apply to the block **before** margins.
- Border + padding are then added.

So setting `BorderBox.Width(m.width)` can overshoot the true terminal width once border+padding are included unless we measure and subtract the frame size or clamp output.

## Fix strategy (what we changed)

We iterated in small compiling commits, repeatedly re-running the tmux harness after each change.

High-level tactics used:
- **Make layout compute frame-aware**: compute body sizes relative to the *inner frame* size (terminal minus border/padding frame).
- **Account for non-body blocks**: Filters mode includes the presets block + footer; reserve space for them so the body doesn’t overflow.
- **Bound component height**: ensure `filterpane` reduces list height when rendering a rules preview, so it doesn’t exceed its assigned height.
- **Reduce “surprise width”**: avoid width-expanding styling on frequently-rendered content (e.g. padding in `styles.Base`), and avoid “almost width” separator calculations when possible.
- **Recompute layout when footer height changes**: toast appearing/disappearing changes footer height; body must shrink/grow accordingly.

### Key commits referenced

See git history around:
- `fd69714...` — frame-aware sizing attempt
- `d86ea73...` — bound filterpane height
- `ac4ad8e...` — account filter presets/footer in layout
- `cd4eb22...` — toast-aware layout recompute + status width fixes
- `0eb5dcf...` — ensure full borders are enabled in styles
- `a4bc17f...` — 1-column slack experiments
- `9619ea3...` — Base style width neutrality

The authoritative source is the diary entry Step 15 (and subsequent updates).

## Validation checklist

After any layout/border change:

- Run:

```bash
cd /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe && \
rm -f ./ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/scripts/tui-*.txt && \
./ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/scripts/tui-tmux.sh build && \
./ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/scripts/tui-tmux.sh scenario smoke
```

- Inspect `tui-00-start-*` and `tui-01-filters-*` captures:
  - Top-left corner `╭` should appear as the first line.
  - Top-right `╮` should appear at the end of the first line.
  - Bottom-right `╯` should appear at the end of the last line.
  - Title lines should be visible (no scrollback cropping).

## Follow-ups / hardening ideas

- Add a small helper in the harness: **assert-corners** that checks capture contains `╭`, `╮`, `╯` (quick regression signal).
- Replace hard-coded header heights in app root with measured header heights (like we do with footer via `lipgloss.Height(status.View())`), so we stop “guessing”.

---
Title: TUI Filters page border/layout debugging (tmux captures)
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
LastUpdated: 2025-12-27T00:22:23.292988926-05:00
WhatFor: ""
WhenToUse: ""
---

