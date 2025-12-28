---
Title: 'Bug report: Filters screen border/right-edge regression (tmux capture)'
Ticket: 002A-FIX-FILTER-VIEW
Status: active
Topics:
    - tui
    - bubbletea
    - ux
    - refactoring
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/tui/app/model.go
      Note: Layout + frame sizing + footer accounting
    - Path: internal/tui/app/view.go
      Note: BorderBox sizing and rendering (border regressions surface here)
    - Path: internal/tui/components/filterpane/model.go
      Note: Filters body sizing/rules preview; can trigger overflow
    - Path: ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/scripts/tui-tmux.sh
      Note: Reproduction harness for tmux captures
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-27T12:26:35.22801816-05:00
WhatFor: ""
WhenToUse: ""
---


# Bug report: Filters screen border/right-edge regression (tmux capture)

## Goal

Provide a copy/paste-ready bug report for the Filters screen border regression seen under tmux captures (top/right borders “missing”), including reproduction, expected/actual, suspected root cause, and validation checklist.

## Context

During ticket `002-MAKE-IT-WORK`, we refactored the TUI into a modular Bubbletea app (`internal/tui/app`) and introduced child components (`filelist`, `filterpane`, result viewport, status/help/toasts). After Phase 5 (filter pane component) and several layout changes, tmux captures showed the Filters screen looking “cropped”, especially missing the **top** and **right** border edges.

Key detail: tmux captures can make **overflow** (too wide / too tall output) look like a border rendering bug. We use the ticket tmux harness for reproducible snapshots.

## Quick Reference

### Summary

- **Bug**: In tmux captures, the Filters screen appears to lose borders (top border/title scrolled away; right border clipped).
- **Impact**: UI regression during refactor; undermines using tmux captures as “screenshots”.
- **Likely cause**: Layout sizing mismatch between terminal size and lipgloss frame semantics (border/padding), plus dynamic footer height (toast/help) and fixed blocks (Filters “Quick Add Presets”).

### Reproduction (tmux harness)

```bash
cd /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe && \
rm -f ./ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/scripts/tui-*.txt && \
./ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/scripts/tui-tmux.sh build && \
./ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/scripts/tui-tmux.sh scenario smoke
```

Inspect:

```bash
ls -1t ./ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/scripts/tui-01-filters-*.txt | head -1
```

### Expected

- Captures include full frame corners:
  - top-left `╭`
  - top-right `╮`
  - bottom-right `╯`
- Title/header visible (no scrollback cropping).

### Actual (symptoms)

- Capture starts with `│ ...` lines (top border/title missing).
- Top/bottom border lines end with `─` (right corner missing), suggesting right side clipped.

### Validation checklist (post-fix)

```bash
cd /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe && \
./ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/scripts/tui-tmux.sh build && \
TMUX_COLS=120 TMUX_ROWS=34 ./ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/scripts/tui-tmux.sh stop || true && \
TMUX_COLS=120 TMUX_ROWS=34 ./ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/scripts/tui-tmux.sh start && \
tmux capture-pane -t prescribe-002:tui.0 -p | grep -q '╭' && echo top-left-ok && \
tmux capture-pane -t prescribe-002:tui.0 -p | grep -q '╮' && echo top-right-ok && \
tmux capture-pane -t prescribe-002:tui.0 -p | grep -q '╯' && echo bottom-right-ok && \
TMUX_COLS=120 TMUX_ROWS=34 ./ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/scripts/tui-tmux.sh stop
```

### Suspected root cause (technical)

- **lipgloss `Style.Width/Height` are pre-border sizes**; if you set `BorderBox.Width(termWidth)`, the rendered box becomes `termWidth + borderW + paddingW`, overflowing and clipping the right edge.
- **Footer height is dynamic** (toast appears/disappears; help toggles short/full) and must be accounted for in layout.
- **Filters screen has extra fixed blocks** (quick presets) beyond the body component.

## Usage Examples

### Filing the issue (paste template)

```text
Title: TUI Filters screen borders missing under tmux capture (top scroll + right clip)

Repro:
- Run: ./ttmp/.../scripts/tui-tmux.sh scenario smoke
- Inspect: tui-01-filters-*.txt

Expected:
- Full box corners: ╭ ... ╮ and ╰ ... ╯
- Title/header visible

Actual:
- Missing top/title (capture starts with │ lines)
- Missing right corners (line ends with ─)

Notes:
- Likely overflow: height overflow scrolls top away; width overflow clips right edge.
```

## Related

- Ticket 002 analysis: `ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/analysis/03-tui-filters-page-border-layout-debugging-tmux-captures.md`
- Ticket 002 diary: `ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/reference/02-diary.md` (Step 15)
