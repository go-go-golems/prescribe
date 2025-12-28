---
Title: 'File management screens: spec vs implementation'
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
    - Path: prescribe/internal/controller/controller.go
      Note: ReplaceWithFullFile()
    - Path: prescribe/internal/domain/domain.go
      Note: FileChange (Type
    - Path: prescribe/internal/tui/app/view.go
      Note: renderMain() - Main screen rendering
    - Path: prescribe/internal/tui/components/filelist/model.go
      Note: File list component (single-line
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-27T20:00:00-05:00
WhatFor: ""
WhenToUse: ""
---


# File Management Screens: Spec vs Implementation

This document analyzes the file management screens from the original TUI spec (`claude-session-TUI-simulation.md`) against the current implementation.

## Screens Analyzed

1. **Main Screen** (Spec §1)
2. **Edit Context Window** (Spec §2)
3. **Replace with Full File Dialog** (Spec §3)
4. **Filtered Out Files View** (Spec §4)

---

## 1. Main Screen

### Spec Requirements

**Layout:**
- Title: "PR DESCRIPTION GENERATOR"
- Branch info: `source → target`
- Stats: Files changed, token count
- Changed files list with checkboxes `[✓]` / `[ ]`
- Active filters summary
- Additional context summary
- Prompt template preview
- Action buttons: `[C] Edit Context [F] Edit Filters [H] View Hidden [E] Edit Prompt [A] Add Files [T] Add Notes [G] Generate [S] Save [L] Load [Q] Quit`

**Data:**
- File list shows: path, included status, additions/deletions, tokens, type (diff/full_file), version
- Filter summary shows active filter names
- Context summary shows count/type
- Prompt shows current template text

### Current Implementation

**Location:** `internal/tui/app/view.go::renderMain()`

**What Exists:**
- ✅ Title: "PRESCRIBE" (centered, styled)
- ✅ Branch info: `source → target` format
- ✅ Stats line: `Files: X visible, Y filtered | Tokens: Z | Filters: N`
- ✅ File list: Uses `filelist.Model` (bubbles/list) with single-line delegate
- ✅ File display: `[✓/ ] path +N -M (Xt)` format
- ✅ Selection indicator: `▶` marker
- ✅ Filtered view toggle: `V` key (`ToggleFilteredView`)
- ✅ Generate: `G` key
- ✅ Quit: `Q` / `Ctrl+C`

**What's Missing:**
- ❌ **Additional context section**: No UI display of `AdditionalContext` items
- ❌ **Prompt template preview**: No display of current prompt text
- ❌ **Action buttons**: No `[C] Edit Context`, `[A] Add Files`, `[T] Add Notes`, `[E] Edit Prompt`, `[S] Save`, `[L] Load` shortcuts
- ❌ **Filter summary line**: No "FILTERS: X (Y hidden) [H to view]" line
- ⚠️ **File type/version display**: File list doesn't show `[DIFF]` / `[FULL:AFTER]` indicators

**CLI Equivalents:**
- `prescribe session init --save` (create session)
- `prescribe session load` (load session)
- `prescribe session save` (save session)
- `prescribe context add --note "..."` (add note)
- `prescribe context add <file-path>` (add file)
- `prescribe file toggle <path>` (toggle inclusion)

**Code References:**
- Model: `internal/tui/app/state.go::Model` (ModeMain)
- View: `internal/tui/app/view.go::renderMain()`
- File list component: `internal/tui/components/filelist/model.go`
- Domain data: `internal/domain/domain.go::PRData` (ChangedFiles, AdditionalContext, CurrentPrompt)

---

## 2. Edit Context Window

### Spec Requirements

**Layout:**
- Split pane: file list (left) + diff preview (right)
- File list shows: checkbox, path, stats, type indicator `[DIFF]` / `[FULL:AFTER]`
- Right pane shows: first 10 lines of diff/full file content
- Actions: `↑↓ Navigate Space Toggle Enter Full View R Replace Options D Restore Diff F Filter A Add Other Files Esc Back`

**Data:**
- Total token count displayed
- Selected file's diff/full content preview
- File type and version indicators

### Current Implementation

**What Exists:**
- ✅ File list navigation: `↑↓` / `j/k` keys
- ✅ Toggle inclusion: `Space` key
- ✅ Filtered view toggle: `V` key (shows filtered files)
- ✅ Back: `Esc` key
- ✅ Filter screen: `F` key opens filter management

**What's Missing:**
- ❌ **Split-pane layout**: No right-side diff preview pane
- ❌ **Diff preview**: No "first 10 lines" preview of selected file
- ❌ **Full file view**: No `Enter` to view complete diff/file content
- ❌ **Replace with full file**: No `R` key to open Replace Dialog
- ❌ **Restore to diff**: No `D` key to convert full file back to diff
- ❌ **Add files from repo**: No `A` key to add other files
- ❌ **File type indicators**: No `[DIFF]` / `[FULL:AFTER]` display in list

**CLI Equivalents:**
- `prescribe context add <file-path>` (add file from repo)
- Domain supports `ReplaceWithFullFile()` / `RestoreToDiff()` but no TUI

**Code References:**
- Controller: `internal/controller/controller.go::ReplaceWithFullFile()`, `RestoreToDiff()`, `AddContextFile()`
- Domain: `internal/domain/domain.go::FileChange` (Type, Version fields exist)
- File list: `internal/tui/components/filelist/model.go` (single-line only, no preview)

---

## 3. Replace with Full File Dialog

### Spec Requirements

**Layout:**
- Modal dialog showing file path
- Radio options: Before / After / Both (with token counts)
- Current token count shown
- Actions: `↑↓ Select Space Toggle Enter Confirm Esc Cancel`

**Data:**
- File path
- Current tokens (diff mode)
- Token counts for each version option
- Selected option

### Current Implementation

**What Exists:**
- ✅ Domain model supports: `FileTypeFull` with `FileVersionBefore/After/Both`
- ✅ Controller method: `ReplaceWithFullFile(index, version)`
- ✅ Session persistence: `session.go` saves `mode: "full_before"/"full_after"/"full_both"`

**What's Missing:**
- ❌ **Dialog screen**: No TUI dialog/modal for selecting version
- ❌ **Token preview**: No display of token counts per version option
- ❌ **UI flow**: No way to trigger this from TUI

**CLI Equivalents:**
- None (would require CLI command to set file mode)

**Code References:**
- Domain: `internal/domain/domain.go::FileVersion` enum, `ReplaceWithFullFile()` method
- Controller: `internal/controller/controller.go::ReplaceWithFullFile()`
- Session: `internal/session/session.go::FileConfig.Mode` (supports all versions)

---

## 4. Filtered Out Files View

### Spec Requirements

**Layout:**
- Title: "FILTERED OUT FILES"
- Filter name displayed
- Count: "X file hidden from context"
- File list showing filtered files (same format as main list)
- Actions: `Space Toggle to include F Edit Filters Esc Back`

**Data:**
- Active filter name
- List of filtered files with tokens, type, version

### Current Implementation

**Location:** `internal/tui/app/view.go::renderMain()` (when `showFiltered == true`)

**What Exists:**
- ✅ Toggle: `V` key switches between visible/filtered view
- ✅ Header changes: Shows "FILTERED FILES" when `showFiltered == true`
- ✅ File list: Uses same `filelist.Model`, shows `GetFilteredFiles()`
- ✅ Toggle inclusion: `Space` works (but shows toast "Filtered view is read-only")
- ✅ Filter screen: `F` key opens filter management
- ✅ Back: `Esc` returns to main

**What's Missing:**
- ⚠️ **Read-only behavior**: Spec allows toggling filtered files to include them; current implementation blocks this with toast
- ❌ **Filter name display**: No "Filter: X" line showing which filter caused exclusion
- ❌ **Count line**: No "X file hidden from context" summary

**CLI Equivalents:**
- `prescribe file toggle <path>` (can toggle any file, including filtered ones)

**Code References:**
- Model: `internal/tui/app/model.go::showFiltered` flag
- View: `internal/tui/app/view.go::renderMain()` (conditional header + file source)
- Domain: `internal/domain/domain.go::PRData::GetFilteredFiles()`
- Controller: `internal/controller/controller.go::GetFilteredFiles()`

---

## Summary

### Implemented Features
- Main screen file list with navigation and toggle
- Filtered files view (toggle with `V`)
- Basic stats (files visible/filtered, tokens, filter count)
- Branch display
- Generate action

### Missing Features
- **Context management UI**: No display or editing of `AdditionalContext` items
- **Prompt management UI**: No display or editing of prompt template
- **File mode switching**: No UI for ReplaceWithFullFile / RestoreToDiff
- **Diff preview**: No split-pane or preview pane
- **Session save/load dialogs**: Auto-save exists, but no explicit save/load UI
- **File type indicators**: No `[DIFF]` / `[FULL:AFTER]` display

### Architectural Notes
- Domain/controller fully support all spec features (file modes, context, prompts)
- TUI is intentionally minimal (Phase 2 goal: "behavior-compatible root model")
- Many features exist as CLI commands but not TUI screens
- File list component is single-line only (no preview pane support yet)
