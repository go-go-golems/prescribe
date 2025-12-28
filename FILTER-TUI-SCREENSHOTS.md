# Filter System - TUI Visual Guide

## Overview

This document provides a visual walkthrough of the filter system in the PR Builder TUI, showing each screen and interaction.

## Screen 1: Main Screen with Filters Active

### Layout

```
╭──────────────────────────────────────────────────────────────────────────────╮
│                          PR DESCRIPTION GENERATOR                             │
│                                                                               │
│ feature/user-auth → master                                                   │
│                                                                               │
│ Files: 2 visible, 1 filtered | Tokens: 459 | Filters: 1                     │
│                                                                               │
│ CHANGED FILES                                                                 │
│ ──────────────────────────────────────────────────────────────────────────── │
│ ▶ [✓] src/auth/login.ts                       +45  -3   (250t)              │
│   [✓] src/auth/middleware.ts                  +30  -5   (209t)              │
│                                                                               │
│                                                                               │
│                                                                               │
│ ──────────────────────────────────────────────────────────────────────────── │
│ [↑↓/jk] Navigate  [Space] Toggle  [F] Filters  [V] View Filtered            │
│ [G] Generate  [Q] Quit                                                       │
╰──────────────────────────────────────────────────────────────────────────────╯
```

### Description

**What's Shown:**
- **Title Bar**: "PR DESCRIPTION GENERATOR"
- **Branch Info**: Shows source branch (feature/user-auth) → target branch (master)
- **Statistics Line**: 
  - `2 visible` - Files passing filters
  - `1 filtered` - Files blocked by filters
  - `459` - Total tokens for visible files
  - `1` - Number of active filters
- **File List**: Shows only visible files (tests/auth.test.ts is filtered out)
- **Help Bar**: Shows available keyboard shortcuts

**Key Features:**
- Filter count prominently displayed in stats
- Only visible files shown in list
- Clear indication of filtering happening (1 filtered)

**Keyboard Shortcuts:**
- `↑↓` or `j/k` - Navigate through files
- `Space` - Toggle file inclusion
- `F` - Open filter management screen
- `V` - Toggle to view filtered files
- `G` - Generate PR description
- `Q` - Quit application

### State

- **Current Selection**: src/auth/login.ts (indicated by ▶)
- **File Status**: Both visible files are included (✓)
- **Active Filters**: 1 filter is active (Exclude tests)
- **Filtered Files**: 1 file is hidden (tests/auth.test.ts)

---

## Screen 2: Main Screen Showing Filtered Files

### Layout

```
╭──────────────────────────────────────────────────────────────────────────────╮
│                          PR DESCRIPTION GENERATOR                             │
│                                                                               │
│ feature/user-auth → master                                                   │
│                                                                               │
│ Files: 2 visible, 1 filtered | Tokens: 459 | Filters: 1                     │
│                                                                               │
│ FILTERED FILES                                                                │
│ ──────────────────────────────────────────────────────────────────────────── │
│ ▶ [✓] tests/auth.test.ts                      +28  -3   (334t)              │
│                                                                               │
│                                                                               │
│                                                                               │
│                                                                               │
│ ──────────────────────────────────────────────────────────────────────────── │
│ [↑↓/jk] Navigate  [Space] Toggle  [F] Filters  [V] View Filtered            │
│ [G] Generate  [Q] Quit                                                       │
╰──────────────────────────────────────────────────────────────────────────────╯
```

### Description

**What's Shown:**
- **Header Changed**: Now shows "FILTERED FILES" instead of "CHANGED FILES"
- **File List**: Shows only the filtered-out files
- **Same Stats**: Statistics remain the same (2 visible, 1 filtered)

**How to Access:**
- Press `V` from main screen to toggle view

**Use Case:**
- Review which files are being excluded
- Verify filters are working as expected
- Decide if you need to adjust filters

**Key Insight:**
- The test file (tests/auth.test.ts) is being filtered out
- It has 334 tokens that won't be included in generation
- File is still marked as "included" but won't be used due to filter

---

## Screen 3: Filter Management Screen (Empty)

### Layout

```
╭──────────────────────────────────────────────────────────────────────────────╮
│                             FILTER MANAGEMENT                                 │
│                                                                               │
│ Active Filters: 0 | Filtered Files: 0                                       │
│                                                                               │
│ ACTIVE FILTERS                                                                │
│ ──────────────────────────────────────────────────────────────────────────── │
│ No active filters                                                             │
│                                                                               │
│                                                                               │
│                                                                               │
│ QUICK ADD PRESETS                                                             │
│ ──────────────────────────────────────────────────────────────────────────── │
│ [1] Exclude Tests  [2] Exclude Docs  [3] Only Source                        │
│                                                                               │
│ ──────────────────────────────────────────────────────────────────────────── │
│ [↑↓/jk] Navigate  [D/X] Delete  [C] Clear All  [1-3] Add Preset             │
│ [Esc] Back                                                                   │
╰──────────────────────────────────────────────────────────────────────────────╯
```

### Description

**What's Shown:**
- **Title**: "FILTER MANAGEMENT"
- **Stats**: 0 active filters, 0 filtered files
- **Filter List**: Empty (shows "No active filters")
- **Presets Section**: Three quick-add presets available
- **Help Bar**: Filter-specific keyboard shortcuts

**How to Access:**
- Press `F` from main screen

**Available Presets:**
1. **Exclude Tests** - Filters out test files (`**/*test*`, `**/*spec*`)
2. **Exclude Docs** - Filters out documentation (`**/*.md`, `**/docs/**`)
3. **Only Source** - Shows only source code files (`.go`, `.ts`, `.js`, `.py`)

**Keyboard Shortcuts:**
- `↑↓` or `j/k` - Navigate through filters (when filters exist)
- `D` or `X` - Delete selected filter
- `C` - Clear all filters
- `1` - Add "Exclude Tests" preset
- `2` - Add "Exclude Docs" preset
- `3` - Add "Only Source" preset
- `Esc` - Return to main screen

---

## Screen 4: Filter Management Screen (With Filters)

### Layout

```
╭──────────────────────────────────────────────────────────────────────────────╮
│                             FILTER MANAGEMENT                                 │
│                                                                               │
│ Active Filters: 2 | Filtered Files: 1                                       │
│                                                                               │
│ ACTIVE FILTERS                                                                │
│ ──────────────────────────────────────────────────────────────────────────── │
│ ▶ [0] Exclude tests - Exclude test files                                    │
│     exclude: tests/*                                                         │
│   [1] Exclude docs - Exclude documentation files                            │
│                                                                               │
│                                                                               │
│ QUICK ADD PRESETS                                                             │
│ ──────────────────────────────────────────────────────────────────────────── │
│ [1] Exclude Tests  [2] Exclude Docs  [3] Only Source                        │
│                                                                               │
│ ──────────────────────────────────────────────────────────────────────────── │
│ [↑↓/jk] Navigate  [D/X] Delete  [C] Clear All  [1-3] Add Preset             │
│ [Esc] Back                                                                   │
╰──────────────────────────────────────────────────────────────────────────────╯
```

### Description

**What's Shown:**
- **Stats Updated**: 2 active filters, 1 file filtered
- **Filter List**: Two filters displayed
  - Filter [0]: "Exclude tests" - Selected (indicated by ▶)
  - Filter [1]: "Exclude docs"
- **Rule Display**: Selected filter shows its rules indented below
  - `exclude: tests/*` pattern shown

**Key Features:**
- **Selection Indicator**: ▶ shows which filter is selected
- **Rule Expansion**: Selected filter's rules are displayed
- **Index Numbers**: [0], [1] for easy reference
- **Descriptions**: Optional descriptions shown after dash

**Interactions:**
- Navigate with `↑↓` to select different filters
- Press `D` or `X` to delete the selected filter
- Press `C` to clear all filters at once
- Press `1-3` to add more preset filters
- Press `Esc` to return to main screen

---

## Screen 5: Filter Management with Multiple Rules

### Layout

```
╭──────────────────────────────────────────────────────────────────────────────╮
│                             FILTER MANAGEMENT                                 │
│                                                                               │
│ Active Filters: 1 | Filtered Files: 3                                       │
│                                                                               │
│ ACTIVE FILTERS                                                                │
│ ──────────────────────────────────────────────────────────────────────────── │
│ ▶ [0] Exclude Tests - Exclude test files                                    │
│     exclude: **/*test*                                                       │
│     exclude: **/*spec*                                                       │
│                                                                               │
│                                                                               │
│ QUICK ADD PRESETS                                                             │
│ ──────────────────────────────────────────────────────────────────────────── │
│ [1] Exclude Tests  [2] Exclude Docs  [3] Only Source                        │
│                                                                               │
│ ──────────────────────────────────────────────────────────────────────────── │
│ [↑↓/jk] Navigate  [D/X] Delete  [C] Clear All  [1-3] Add Preset             │
│ [Esc] Back                                                                   │
╰──────────────────────────────────────────────────────────────────────────────╯
```

### Description

**What's Shown:**
- **Single Filter**: "Exclude Tests" preset added
- **Multiple Rules**: Two rules displayed:
  1. `exclude: **/*test*` - Matches files with "test" anywhere
  2. `exclude: **/*spec*` - Matches files with "spec" anywhere
- **Impact**: 3 files filtered out

**Pattern Explanation:**
- `**/*test*` - Recursive pattern matching any file with "test" in name
- `**/*spec*` - Recursive pattern matching any file with "spec" in name

**This Filter Would Match:**
- `tests/auth.test.ts` ✓
- `src/utils.test.ts` ✓
- `tests/api.spec.ts` ✓
- `__tests__/setup.ts` ✓
- `test-utils.ts` ✓

**This Filter Would NOT Match:**
- `src/auth/login.ts` ✗
- `src/api/users.ts` ✗
- `docs/README.md` ✗

---

## Screen 6: Only Source Preset Applied

### Layout

```
╭──────────────────────────────────────────────────────────────────────────────╮
│                             FILTER MANAGEMENT                                 │
│                                                                               │
│ Active Filters: 1 | Filtered Files: 0                                       │
│                                                                               │
│ ACTIVE FILTERS                                                                │
│ ──────────────────────────────────────────────────────────────────────────── │
│ ▶ [0] Only Source - Include only source code files                          │
│     include: **/*.go                                                         │
│     include: **/*.ts                                                         │
│     include: **/*.js                                                         │
│     include: **/*.py                                                         │
│                                                                               │
│ QUICK ADD PRESETS                                                             │
│ ──────────────────────────────────────────────────────────────────────────── │
│ [1] Exclude Tests  [2] Exclude Docs  [3] Only Source                        │
│                                                                               │
│ ──────────────────────────────────────────────────────────────────────────── │
│ [↑↓/jk] Navigate  [D/X] Delete  [C] Clear All  [1-3] Add Preset             │
│ [Esc] Back                                                                   │
╰──────────────────────────────────────────────────────────────────────────────╯
```

### Description

**What's Shown:**
- **Include Filter**: "Only Source" preset with 4 include rules
- **Multiple Languages**: Supports Go, TypeScript, JavaScript, Python
- **Impact**: 0 files filtered (all files in test repo are .ts)

**Pattern Type: Include**
- Unlike exclude filters, include filters specify what TO show
- Files must match at least one include pattern to be visible
- Useful for focusing on specific file types

**This Filter Would Match:**
- `src/main.go` ✓
- `src/auth/login.ts` ✓
- `src/utils.js` ✓
- `src/models.py` ✓

**This Filter Would NOT Match:**
- `README.md` ✗
- `config.yaml` ✗
- `Dockerfile` ✗
- `package.json` ✗

---

## Interaction Flow

### Adding a Filter

**Steps:**
1. From main screen, press `F` to open filter management
2. Press `1` to add "Exclude Tests" preset
3. Filter is immediately added and saved
4. Rules are displayed when filter is selected
5. Press `Esc` to return to main screen
6. Main screen now shows updated stats (filtered files count)

**Visual Flow:**
```
Main Screen → [F] → Filter Screen → [1] → Filter Added → [Esc] → Main Screen
(0 filters)        (empty list)         (1 filter)              (1 filter active)
```

### Removing a Filter

**Steps:**
1. From main screen, press `F` to open filter management
2. Use `↑↓` or `j/k` to select filter to remove
3. Press `D` or `X` to delete selected filter
4. Filter is immediately removed and saved
5. Stats update to show new filtered file count
6. Press `Esc` to return to main screen

**Visual Flow:**
```
Filter Screen → [↑↓] → Select Filter → [D] → Filter Deleted → [Esc] → Main Screen
(2 filters)            (filter selected)     (1 filter)               (updated stats)
```

### Viewing Filtered Files

**Steps:**
1. From main screen with active filters
2. Press `V` to toggle view
3. Header changes to "FILTERED FILES"
4. List shows only filtered-out files
5. Press `V` again to return to visible files view

**Visual Flow:**
```
Main Screen → [V] → Filtered View → [V] → Main Screen
(visible)           (filtered)            (visible)
```

---

## Use Case Scenarios

### Scenario 1: Exclude Test Files from PR

**Goal**: Generate PR description without test files

**Steps:**
1. Launch TUI: `pr-builder tui`
2. Press `F` to open filters
3. Press `1` to add "Exclude Tests" preset
4. Press `Esc` to return to main
5. Verify stats show filtered files
6. Press `G` to generate

**Result**: PR description generated without test files

---

### Scenario 2: Focus on Specific Module

**Goal**: Only include files from auth module

**Steps:**
1. Launch TUI: `pr-builder tui`
2. Press `F` to open filters
3. Note: Would need CLI to add custom filter
4. From CLI: `pr-builder add-filter --name "Auth Only" --include "src/auth/**"`
5. Launch TUI again: `pr-builder tui`
6. See only auth files visible
7. Press `G` to generate

**Result**: PR description focused on auth module only

---

### Scenario 3: Review What's Being Filtered

**Goal**: Check which files are excluded

**Steps:**
1. Launch TUI with filters active
2. Note stats: "2 visible, 1 filtered"
3. Press `V` to view filtered files
4. See: tests/auth.test.ts is filtered
5. Press `F` to check filter rules
6. See: "Exclude tests" filter with `tests/*` pattern
7. Press `Esc` to return

**Result**: Confirmed test file is correctly filtered

---

### Scenario 4: Clear All Filters

**Goal**: Remove all filters quickly

**Steps:**
1. Launch TUI with multiple filters
2. Press `F` to open filter management
3. Press `C` to clear all filters
4. Confirm all filters removed
5. Press `Esc` to return
6. Stats now show: "3 visible, 0 filtered"

**Result**: All files now visible

---

## Visual Design Elements

### Color Scheme (Conceptual)

- **Title**: Bold, centered
- **Branch Names**: Green (success color)
- **Stats**: White/default
- **Selected Item**: Highlighted background with ▶ indicator
- **Unselected Items**: Normal text
- **Help Text**: Muted/gray
- **Borders**: Box drawing characters

### Layout Principles

1. **Consistent Header**: Title, branch info, stats always at top
2. **Clear Sections**: Separated by horizontal lines
3. **Help Always Visible**: Bottom bar shows available actions
4. **Selection Indicator**: ▶ symbol for current selection
5. **Checkbox Style**: [✓] for included, [ ] for excluded
6. **Indentation**: Rules indented under their filter

### Responsive Design

- Width: 80 characters (fits standard terminal)
- Height: Adjusts to terminal size
- Scrolling: Supported for long lists (not shown in screenshots)
- Wrapping: Long file names truncated with ellipsis

---

## Keyboard Shortcut Reference

### Main Screen

| Key | Action |
|-----|--------|
| `↑` `↓` | Navigate files |
| `j` `k` | Navigate files (vim-style) |
| `Space` | Toggle file inclusion |
| `F` | Open filter management |
| `V` | Toggle filtered files view |
| `G` | Generate PR description |
| `Q` | Quit application |
| `Ctrl+C` | Quit application |

### Filter Screen

| Key | Action |
|-----|--------|
| `↑` `↓` | Navigate filters |
| `j` `k` | Navigate filters (vim-style) |
| `D` | Delete selected filter |
| `X` | Delete selected filter |
| `C` | Clear all filters |
| `1` | Add "Exclude Tests" preset |
| `2` | Add "Exclude Docs" preset |
| `3` | Add "Only Source" preset |
| `Esc` | Return to main screen |
| `Q` | Quit application |
| `Ctrl+C` | Quit application |

---

## Technical Implementation Notes

### Screen State Machine

```
┌─────────────┐
│ Main Screen │◄─────┐
└──────┬──────┘      │
       │             │
    [F]│             │[Esc]
       │             │
       ▼             │
┌─────────────┐      │
│Filter Screen├──────┘
└─────────────┘
```

### Model Structure

```go
type EnhancedModel struct {
    controller      *controller.Controller
    currentScreen   Screen
    selectedIndex   int
    filterIndex     int
    showFilteredFiles bool
}
```

### Update Logic

```go
func (m *EnhancedModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch m.currentScreen {
    case ScreenMain:
        return m.updateMain(msg)
    case ScreenFilters:
        return m.updateFilters(msg)
    }
}
```

### View Rendering

```go
func (m *EnhancedModel) View() string {
    switch m.currentScreen {
    case ScreenMain:
        return m.renderMain()
    case ScreenFilters:
        return m.renderFilters()
    }
}
```

---

## Summary

The filter TUI provides:

✅ **Visual Filter Management** - Dedicated screen for filters  
✅ **Quick Presets** - One-key addition of common filters  
✅ **Real-time Stats** - See impact immediately  
✅ **Filtered Files View** - Toggle to see what's hidden  
✅ **Intuitive Navigation** - Vim-style and arrow keys  
✅ **Clear Visual Feedback** - Selection indicators, stats, help text  
✅ **Seamless Integration** - Filters work across CLI and TUI  

The TUI makes filter management **fast**, **visual**, and **intuitive**, while maintaining full compatibility with the CLI for scripting and automation.
