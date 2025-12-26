# PR Builder TUI Demo

## Overview

The TUI (Terminal User Interface) provides an interactive way to build PR descriptions using Bubbletea.

## Main Screen

The main screen shows:
- **Header**: Current branch and target branch
- **Stats**: Number of files and total token count
- **File List**: All changed files with:
  - Checkbox indicator (✓ = included,   = excluded)
  - File path
  - Line additions/deletions
  - Token count
- **Navigation**: Arrow keys or j/k to move
- **Actions**:
  - `Space`: Toggle file inclusion
  - `G`: Generate PR description
  - `Q`: Quit

```
╭──────────────────────────────────────────────────────────────────────────────╮
│                                                                              │
│                       PR DESCRIPTION GENERATOR                               │
│                                                                              │
│  feature/user-auth → master                                                 │
│                                                                              │
│  Files: 3 | Tokens: 793                                                     │
│                                                                              │
│  CHANGED FILES                                                              │
│  ──────────────────────────────────────────────────────────────────────────│
│  ▶ [✓] src/auth/login.ts                      +12  -1   (204t)            │
│      [✓] src/auth/middleware.ts               +29  -1   (255t)            │
│      [✓] tests/auth.test.ts                   +28  -3   (334t)            │
│                                                                              │
│  ──────────────────────────────────────────────────────────────────────────│
│  [↑↓/jk] Navigate  [Space] Toggle  [G] Generate  [Q] Quit                  │
│                                                                              │
╰──────────────────────────────────────────────────────────────────────────────╯
```

## Generating Screen

When you press `G`, the TUI shows a loading screen:

```
╭──────────────────────────────────────────────────────────────────────────────╮
│                                                                              │
│                      GENERATING PR DESCRIPTION                               │
│                                                                              │
│  Analyzing changes and generating description...                            │
│                                                                              │
│                                  ⠋                                          │
│                                                                              │
│  This may take a few seconds...                                             │
│                                                                              │
╰──────────────────────────────────────────────────────────────────────────────╯
```

## Result Screen

After generation completes, the result is displayed:

```
╭──────────────────────────────────────────────────────────────────────────────╮
│                                                                              │
│                     GENERATED PR DESCRIPTION                                 │
│                                                                              │
│  ✓ Description generated successfully!                                      │
│                                                                              │
│  ┌────────────────────────────────────────────────────────────────────────┐ │
│  │ # Pull Request: feature/user-auth → master                             │ │
│  │                                                                          │ │
│  │ ## Summary                                                               │ │
│  │ This PR includes changes across 3 files, implementing new features      │ │
│  │ and improvements.                                                        │ │
│  │                                                                          │ │
│  │ ## Changes                                                               │ │
│  │ - **src/auth/login.ts**: +12 lines, -1 lines                           │ │
│  │ - **src/auth/middleware.ts**: +29 lines, -1 lines                      │ │
│  │ - **tests/auth.test.ts**: +28 lines, -3 lines                          │ │
│  │                                                                          │ │
│  │ ## Key Changes                                                           │ │
│  │ - Updated login.ts with new functionality                                │ │
│  │ - Improved middleware with better error handling                         │ │
│  │ - Enhanced authentication system with improved security measures         │ │
│  │                                                                          │ │
│  │ ## Testing                                                               │ │
│  │ - All existing tests pass                                                │ │
│  │ - New tests added for changed functionality                              │ │
│  │ - Manual testing completed for critical paths                            │ │
│  │                                                                          │ │
│  │ ## Breaking Changes                                                      │ │
│  │ None                                                                     │ │
│  └────────────────────────────────────────────────────────────────────────┘ │
│                                                                              │
│  [Esc] Back  [Q] Quit                                                        │
│                                                                              │
╰──────────────────────────────────────────────────────────────────────────────╯
```

## Features

### Session Integration
- TUI automatically loads existing session from `.pr-builder/session.yaml`
- File toggles are saved to session automatically
- Seamless integration with CLI commands

### Keyboard Navigation
- **↑/↓** or **j/k**: Navigate file list
- **Space**: Toggle file inclusion
- **G**: Generate PR description
- **Esc**: Go back (from result screen)
- **Q**: Quit application

### Visual Feedback
- Color-coded file stats (green additions, red deletions, yellow tokens)
- Clear indication of selected file
- Checkbox indicators for file inclusion
- Loading spinner during generation
- Success message on completion

## Running the TUI

```bash
# From any git repository
pr-builder tui

# Or specify repository and target branch
pr-builder -r /path/to/repo -t main tui
```

## Architecture

The TUI follows Bubbletea's Elm architecture:
- **Model**: Contains all state (controller, selected index, flags)
- **Init**: Initializes the model
- **Update**: Handles messages (key presses, generation complete)
- **View**: Renders the current state to a string

The TUI is a thin layer over the core controller - it just provides interactive UI for the same operations available via CLI.
