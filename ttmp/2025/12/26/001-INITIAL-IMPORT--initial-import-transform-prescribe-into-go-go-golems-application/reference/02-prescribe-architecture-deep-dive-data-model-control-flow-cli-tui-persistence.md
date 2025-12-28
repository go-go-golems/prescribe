---
Title: Prescribe Architecture Deep Dive (Data Model, Control Flow, CLI/TUI, Persistence)
Ticket: 001-INITIAL-IMPORT
Status: active
Topics:
    - migration
    - refactoring
    - ci-cd
    - go-module
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-26T17:35:25.051540698-05:00
WhatFor: ""
WhenToUse: ""
---

# Prescribe Architecture Deep Dive (Data Model, Control Flow, CLI/TUI, Persistence)

## Goal

This document explains **how prescribe works end-to-end**: its core data structures, where state lives, how the CLI commands (“verbs”) map to controller operations, how the Bubbletea TUI is structured, and how git/session/prompt subsystems interact. It is intended to be readable by a developer new to this repo, and detailed enough to support refactors.

## Context

Prescribe is a **CLI + TUI** tool that inspects a git repository’s changes between a source branch (current `HEAD`) and a target branch (default branch, or `--target`), allows you to filter/toggle/include context, and then generates a PR description using an (currently mocked) “LLM API”.

At runtime, the “application state” is primarily a single in-memory object: **`*domain.PRData`** (`internal/domain/domain.go`). The CLI and TUI are different front-ends that drive mutations on this state via `controller.Controller`.

### Repo layout (relevant parts)

- **Entrypoint**: `cmd/prescribe/main.go` (`package main`)
- **Cobra root + command registration**: `cmd/prescribe/cmds/root.go` (`package cmds`)
- **Command implementations**:
  - `cmd/prescribe/cmds/generate.go`, `cmd/prescribe/cmds/tui.go`
  - `cmd/prescribe/cmds/session/*` (init/load/save/show)
  - `cmd/prescribe/cmds/filter/*` (add/list/remove/clear/test/show-filtered)
  - `cmd/prescribe/cmds/file/*` (toggle-file, add-context)
- **Core business logic + services**:
  - `internal/domain/` (data structures + core operations)
  - `internal/controller/` (orchestration + “application service” layer)
  - `internal/session/` (YAML persistence)
  - `internal/git/` (git CLI calls)
  - `internal/api/` (mock generator)
  - `internal/tui/` (Bubbletea models + Lipgloss styles)

## Quick Reference

This section is optimized for fast navigation during debugging and refactoring.

### “One diagram” mental model

```text
            ┌─────────────────────────────────────────────────────────┐
            │                       CLI / TUI                         │
            │   Cobra cmds/* (one process per invocation)             │
            │   Bubbletea model (single process with event loop)      │
            └───────────────┬───────────────────────────┬─────────────┘
                            │                           │
                            ▼                           ▼
                  ┌─────────────────┐        ┌─────────────────────────┐
                  │ controller.Controller     │ internal/tui.EnhancedModel
                  │ internal/controller        │ internal/tui/model_enhanced.go
                  └───────────────┬──────────┘
                                  │ owns
                                  ▼
                           ┌───────────────┐
                           │  domain.PRData│
                           │ internal/domain/domain.go
                           └───────┬───────┘
                                   │ uses
            ┌──────────────────────┼───────────────────────────────────┐
            ▼                      ▼                                   ▼
   internal/git.Service     internal/session.Session            internal/api.Service
   (shells out to git)      (YAML save/load/apply)              (mock “LLM” generator)
```

### Core types (data model)

- **`domain.PRData`** (`internal/domain/domain.go`)
  - **branches**: `SourceBranch`, `TargetBranch`
  - **files**: `ChangedFiles []domain.FileChange`
  - **filters**: `ActiveFilters []domain.Filter`
  - **context**: `AdditionalContext []domain.ContextItem`
  - **prompt**: `CurrentPrompt string`, `CurrentPreset *domain.PromptPreset`
  - **generation output**: `GeneratedDescription string`

- **`domain.FileChange`** (`internal/domain/domain.go`)
  - `Path string`
  - `Included bool`
  - `Additions`, `Deletions` (from `git diff --numstat`)
  - `Diff string`, `FullBefore string`, `FullAfter string`
  - `Type domain.FileType` (`diff` vs `full_file`)
  - `Version domain.FileVersion` (`before`, `after`, `both`)
  - `Tokens int` (**rough estimate**, currently `len(text)/4`)

- **`domain.Filter`** and **`domain.FilterRule`**
  - `FilterRule.Type` is `include` or `exclude`
  - `FilterRule.Pattern` is a glob (doublestar)

- **`session.Session`** (`internal/session/session.go`)
  - YAML representation used for persistence; converted to/from `PRData`.

### Where state is stored

- **In-process (ephemeral)**:
  - `*domain.PRData` in `controller.Controller.data`
  - TUI selection state (cursor positions, screen mode) in `internal/tui.EnhancedModel`

- **On disk (persistent)**:
  - **Session YAML**: `internal/session.GetDefaultSessionPath(repoPath)`
    - Currently: `repoPath/.pr-builder/session.yaml` (legacy naming)
  - **Prompt presets**:
    - Project: `repoPath/.pr-builder/prompts/*.y{a,}ml`
    - Global: `~/.pr-builder/prompts/*.y{a,}ml`

### Primary entrypoints and control flow

#### Entrypoint

```text
cmd/prescribe/main.go
  main():
    cmds.Execute()

cmd/prescribe/cmds/root.go
  init():
    define persistent flags: --repo, --target
    registerCommands()
  Execute():
    rootCmd.Execute()
```

#### Controller lifecycle (CLI commands)

```text
RunE():
  repoPath := flag --repo (default ".")
  target   := flag --target (default "" -> controller picks default branch)
  ctrl := controller.NewController(repoPath)
  ctrl.Initialize(target)         // populates PRData from git
  ctrl.LoadSession(defaultPath)   // some commands do this
  mutate ctrl.data and/or run ctrl.GenerateDescription
  ctrl.SaveSession(defaultPath)   // for mutating commands
```

### CLI command surfaces (current)

Command groups are folders, but Cobra commands are currently registered as **top-level** commands:

- **session**: `init`, `save`, `load`, `show`
- **filter**: `add-filter`, `list-filters`, `remove-filter`, `clear-filters`, `test-filter`, `show-filtered`
- **file/context**: `toggle-file`, `add-context`
- **generation/UI**: `generate`, `tui`

### TUI model (current)

The `tui` command uses **`internal/tui.NewEnhancedModel`** (not `internal/tui.NewModel`).

## Usage Examples

All examples use `go run` (preferred) to avoid relying on installed binaries.

### Get oriented (help)

```bash
cd /path/to/your/repo
go run /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/cmd/prescribe --help
```

### Create a session and save it

```bash
go run /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/cmd/prescribe -r . init
go run /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/cmd/prescribe -r . save
```

### Add a filter and inspect what got filtered

```bash
go run /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/cmd/prescribe -r . add-filter --name "No tests" --exclude "**/*test*" --exclude "**/*spec*"
go run /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/cmd/prescribe -r . show-filtered
```

### Generate a PR description

```bash
go run /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/cmd/prescribe -r . generate
```

### Run the TUI

```bash
go run /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/cmd/prescribe -r . tui
```

## Related

This reference complements:

- Ticket analysis doc: `ttmp/.../analysis/01-current-structure-analysis-and-transformation-plan.md`
- Ticket doc audit: `ttmp/.../analysis/02-documentation-analysis-and-transformation-plan.md`

---

# Deep Dive

## 1) Core data structures (`internal/domain`)

The `internal/domain` package is the heart of the app. It defines the “PR context” model and implements methods that compute derived views (visible vs filtered files, total tokens) and apply mutations (toggle inclusion, add/remove filters, add/remove context, set prompt).

### `domain.PRData` (single source of truth)

`domain.PRData` (`internal/domain/domain.go`) contains:

- **Branch info**
  - `SourceBranch`: current branch name
  - `TargetBranch`: base branch used for diff
- **File inputs**
  - `ChangedFiles []FileChange`: list of changed files with diff/full content and metadata
  - `AdditionalContext []ContextItem`: extra file contents or notes
- **Filtering**
  - `ActiveFilters []Filter`: named filter objects made of ordered rules
- **Prompt**
  - `CurrentPrompt string`: prompt template used for generation
  - `CurrentPreset *PromptPreset`: pointer to a preset if selected (optional)
- **Output**
  - `GeneratedDescription string`: last generated description

### `domain.FileChange` (one changed file)

Each `FileChange` holds:

- **Identity**: `Path`
- **Selection**: `Included` (user intent)
- **Stats**: `Additions`, `Deletions`
- **Content**:
  - `Diff` (always fetched)
  - `FullBefore`, `FullAfter` (best-effort, may be empty)
- **View mode**:
  - `Type` is `diff` or `full_file`
  - `Version` is `before` / `after` / `both` when `Type == full_file`
- **Token estimate**: `Tokens` (approx `len(text)/4`)

### Filters: how visibility is computed

Visibility is computed by `PRData.GetVisibleFiles()` and `PRData.GetFilteredFiles()`, both of which call `passesFilters(path)`.

Current behavior of `passesFilters` (as implemented):

```text
passesFilters(path):
  if no filters: return true
  for each filter:
    for each rule in filter:
      matches := doublestar.Match(rule.Pattern, path)
      if rule is exclude and matches: return false
      if rule is include and NOT matches: return false
  return true
```

Important nuance: this is an “AND of all rules”, including include rules, which is stricter than “match at least one include”.

### Tokens are estimates (not a tokenizer)

Token counts are derived from string length (`len(text)/4`). They’re useful for rough budgeting, but not accurate vs real LLM tokenization.

---

## 2) Core orchestration (`internal/controller`)

The controller is the application-service layer. It owns the mutable `*domain.PRData` and wires it to git/session/api operations.

### `controller.Controller` fields

From `internal/controller/controller.go`:

- `data *domain.PRData`
- `gitService *git.Service`
- `apiService *api.Service`
- `repoPath string`

### Lifecycle

```text
ctrl := controller.NewController(repoPath)
ctrl.Initialize(targetBranch)            // populate SourceBranch/TargetBranch/ChangedFiles
ctrl.LoadSession(defaultSessionPath)     // optional
... mutate ctrl.data ...
ctrl.SaveSession(defaultSessionPath)     // optional
```

### Initialize (hydrate from git)

`Controller.Initialize(targetBranch)`:

1. `gitService.GetCurrentBranch()` → `PRData.SourceBranch`
2. If target branch not provided: `gitService.GetDefaultBranch()` → `PRData.TargetBranch`
3. `gitService.GetChangedFiles(source, target)` → `PRData.ChangedFiles`

### GenerateDescription (build request from PRData)

`Controller.GenerateDescription()`:

```text
visible := PRData.GetVisibleFiles()
included := [f in visible where f.Included]
if included empty: error

req := {
  SourceBranch: PRData.SourceBranch,
  TargetBranch: PRData.TargetBranch,
  Files: included,
  AdditionalContext: PRData.AdditionalContext,
  Prompt: PRData.CurrentPrompt,
}
api.ValidateRequest(req)
resp := api.GenerateDescription(req)
PRData.GeneratedDescription = resp.Description
return resp.Description
```

### Filters and filter testing

- `GetFilters()` returns `PRData.ActiveFilters` directly.
- `ClearFilters()` resets `ActiveFilters` to empty.
- `TestFilter(filter)` does a “what would happen” simulation against current `ChangedFiles` without mutating state.

---

## 3) Git interactions (`internal/git`)

Prescribe interacts with git via subprocess calls (not a Go git library). This makes behavior easy to inspect and consistent with local git semantics.

### Key git commands used

- Current branch: `git rev-parse --abbrev-ref HEAD`
- Default branch:
  - preferred: `git symbolic-ref refs/remotes/origin/HEAD`
  - fallback: `git rev-parse --verify main` else `master`
- Changed files: `git diff --numstat target...source`
- Per-file diff: `git diff target...source -- path`
- File contents: `git show ref:path`

### How `[]domain.FileChange` is built

The algorithm is:

1. Parse `--numstat` lines into additions/deletions/path.
2. Fetch per-file diff and full before/after contents.
3. Fill the `domain.FileChange` struct with defaults (`Included=true`, `Type=diff`).

---

## 4) Persistence and where state lives (`internal/session`)

Session persistence is YAML round-tripping between `domain.PRData` and `session.Session`.

### Default session file location (current)

`internal/session.GetDefaultSessionPath(repoPath)` currently returns:

```text
repoPath/.pr-builder/session.yaml
```

This is still “pr-builder” naming and is a future cleanup opportunity.

### Session compatibility checks

When a command loads a session via `Controller.LoadSession`, it requires:

- `sess.SourceBranch == PRData.SourceBranch`

This prevents applying a session captured on another branch to the current branch.

---

## 5) Mock generation (`internal/api`)

The `internal/api` package simulates an LLM:

- sleeps for 2 seconds
- formats a Markdown PR description
- “extracts” key changes heuristically from filenames
- appends additional note context

It is intentionally a placeholder for later integration with a real LLM provider.

---

## 6) CLI verbs (`cmd/prescribe/cmds/**`)

Each Cobra command is “one file, one verb”. Most commands follow the same structure:

1. Read global flags (`--repo`, `--target`)
2. Create controller
3. Initialize from git
4. Optionally load session
5. Perform operation
6. Optionally save session

### Session verbs

- `init` (`cmd/prescribe/cmds/session/init.go`): initialize PRData from git, optionally save.
- `load` (`cmd/prescribe/cmds/session/load.go`): initialize then apply a session file.
- `save` (`cmd/prescribe/cmds/session/save.go`): initialize and write session to disk.
- `show` (`cmd/prescribe/cmds/session/show.go`): initialize, best-effort load session, then print.

### Filter verbs

- `add-filter`: creates a `domain.Filter` and saves session
- `list-filters`: loads session (best-effort) and prints filters + impact
- `remove-filter`: loads session (required), resolves index/name, removes, saves
- `clear-filters`: loads session (required), clears, saves
- `test-filter`: simulates filter impact without applying it
- `show-filtered`: loads session (best-effort) and prints filtered files + active rules

### File/context verbs

- `toggle-file <path>`: toggles the `Included` boolean for a path and saves session
- `add-context [file] --note ...`: adds a repository file’s content or a freeform note; saves session

---

## 7) TUI (`internal/tui`)

The CLI `tui` command runs `internal/tui.NewEnhancedModel(ctrl)`, which implements a simple screen-state machine:

- Main (file list)
- Filters (filter management)
- Generating (spinner)
- Result (generated PR text)

### Event loop and persistence behavior

- Toggling file inclusion or editing filters triggers an immediate `SaveSession(defaultPath)` call.
- Generating runs asynchronously via `tea.Cmd` that calls `controller.GenerateDescription()`.

### Styles

All styling is centralized in `internal/tui/styles.go` via Lipgloss styles.

---

## Known gaps / future cleanup targets

- **Rename `.pr-builder/` → `.prescribe/`** for session + prompt storage (behavior change; update docs/tests).
- **Introduce Cobra groups** (e.g. `prescribe filter add ...`) if desired; currently groups are organizational folders only.
- **Replace token estimation** with a real tokenizer if accuracy matters.
- **Replace mock API** with a real LLM provider integration.
