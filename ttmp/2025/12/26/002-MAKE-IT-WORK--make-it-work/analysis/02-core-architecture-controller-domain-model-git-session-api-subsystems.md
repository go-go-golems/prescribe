---
Title: 'Core architecture: Controller, domain model, git/session/api subsystems'
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
LastUpdated: 2025-12-26T19:22:37.155000399-05:00
WhatFor: ""
WhenToUse: ""
---

# Core architecture: Controller, domain model, git/session/api subsystems

## Executive summary

Prescribe’s “core” is intentionally small and pragmatic: a single `Controller` owns a single in-memory `PRData` value and orchestrates a few subsystems:

- `internal/git.Service`: shells out to `git` to discover branches and diffs
- `internal/session`: serializes/deserializes the persistent session YAML
- `internal/api.Service`: generates PR descriptions (currently a mock, but shaped like an LLM API)
- `internal/domain.PRData`: the authoritative in-memory model for files, filters, prompt, context, and generated output

If you remember one mental model, make it this:

> **Initialize** (read-only) populates `PRData` from git → **mutations** toggle/filters/context/prompt modify `PRData` → **persist** saves a projection of `PRData` to YAML → **generate** builds a request from `PRData` and calls the API.

This document explains the nouns and verbs of that system, with concrete file/symbol references and a focus on invariants and failure modes (per the didactic writing guidelines).

## The mental model (nouns and verbs)

### Nouns (data you should know)

#### `domain.PRData` (the authoritative state)

**File:** `prescribe/internal/domain/domain.go`

`PRData` is the whole app state:

- git: `SourceBranch`, `TargetBranch`
- files: `ChangedFiles []FileChange`
- filters: `ActiveFilters []Filter`
- prompt: `CurrentPrompt`, `CurrentPreset`
- extra context: `AdditionalContext []ContextItem`
- output: `GeneratedDescription`

Crucial point: **file identity is the file path** (`FileChange.Path`).

#### `session.Session` (the persistent projection)

**File:** `prescribe/internal/session/session.go`

Session YAML stores a projection of `PRData` that’s designed to survive restart:

- branches (`source_branch`, `target_branch`)
- per-file config: `path`, `included`, `mode`
- filters: name/description + rules
- additional context: notes/files
- prompt: preset id or template

Session does **not** store the entire diffs; it stores *configuration* applied back onto the latest git-derived file list.

#### `controller.Controller` (the orchestrator)

**Files:**

- `prescribe/internal/controller/controller.go`
- `prescribe/internal/controller/session.go`

The controller has:

- `data *domain.PRData`
- `gitService *git.Service`
- `apiService *api.Service`
- `repoPath string`

This is the only layer that knows how to:

- populate `PRData` from git (`Initialize`)
- call the LLM API (`GenerateDescription`)
- bridge persistence (`SaveSession`/`LoadSession`)

### Verbs (flows you should know)

The system basically has 4 verbs:

1) **Initialize**: git → PRData
2) **Mutate**: user actions → PRData
3) **Persist**: PRData → session YAML (and back)
4) **Generate**: PRData → API request → generated text

The rest of this doc walks those verbs end-to-end.

## Flow 1: Initialize (git → PRData)

This is the “read-only discovery” step.

### What happens

- `Controller.Initialize(targetBranch)`:
  - discovers current branch via `gitService.GetCurrentBranch()`
  - selects a target branch:
    - uses argument if provided, else `gitService.GetDefaultBranch()`
  - populates `PRData.SourceBranch`, `PRData.TargetBranch`
  - populates `PRData.ChangedFiles` by calling `gitService.GetChangedFiles(sourceBranch, targetBranch)`

### Where it lives

- Controller: `prescribe/internal/controller/controller.go`
- Git: `prescribe/internal/git/git.go`

### Key technical details (git service)

`git.Service.GetChangedFiles`:

- runs `git diff --numstat target...source` to list changed files and line counts
- for each file:
  - reads a per-file diff via `git diff target...source -- <path>`
  - reads “full before” (`git show target:<path>`) and “full after” (`git show source:<path>`)
  - estimates tokens as `len(diff)/4`
  - sets `Included = true` by default
  - sets `Type = diff` by default

### Important invariants and failure modes

- **Repo must be a git repo**: `git.NewService` checks `<repo>/.git` exists.
- **Default branch detection**:
  - tries `refs/remotes/origin/HEAD` first
  - falls back to `main` else `master`
- **File enumeration is “best-effort” for diff/full content**:
  - if per-file diff fails, it sets diff to empty and continues
  - full before/after errors are ignored (empty string)

Implication: UI should tolerate “missing diff content” and treat it as non-fatal.

## Flow 2: Mutations (user actions → PRData)

These are the operations a TUI/CLI triggers.

### File inclusion and file modes

Mutations in `domain.PRData`:

- `ToggleFileInclusion(index)` flips `ChangedFiles[index].Included`
- `ReplaceWithFullFile(index, version)`:
  - sets type + version
  - recalculates tokens based on content
- `RestoreToDiff(index)`:
  - sets type back to diff
  - resets version
  - recalculates tokens based on diff

Today, the TUI calls controller methods that delegate to these domain methods.

### Filters

Filters are a list of named filters, and each filter contains ordered rules (include/exclude) using doublestar globbing.

Core functions:

- `PRData.GetVisibleFiles()` filters `ChangedFiles` through active filters
- `PRData.GetFilteredFiles()` returns the complement

Note: filtering is done per file path; there is no “match set” caching yet.

### Additional context

Additional context items are either:

- file content snapshots (`ContextTypeFile`)
- arbitrary notes (`ContextTypeNote`)

Controller helpers:

- `AddContextFile(path)` reads file content from the current source branch and adds it as a context item with tokens estimate
- `AddContextNote(text)` adds a note context item with tokens estimate

### Prompt selection

Prompt can be either:

- “custom template” (a string), or
- a preset (builtin/project/global).

Controller has:

- `LoadPromptPreset(presetID)` searching:
  - builtins (`domain.GetBuiltinPresets()`)
  - project presets under `<repo>/.pr-builder/prompts`
  - global presets under `$HOME/.pr-builder/prompts`

Note: preset directories still use the legacy name `.pr-builder`.

## Flow 3: Persistence (PRData ↔ session YAML)

Persistence is explicit and separate from initialization.

### Save path and file format

- Default path is `<repo>/.pr-builder/session.yaml` (`session.GetDefaultSessionPath`)
- Format is YAML, modeled by `session.Session` structs.

### Save flow (PRData → YAML)

- `Controller.SaveSession(path)`:
  - creates `session.NewSession(c.data)`
  - calls `sess.Save(path)` which:
    - ensures directory exists (`os.MkdirAll`)
    - `yaml.Marshal`
    - `os.WriteFile`

### Load flow (YAML → PRData)

- `Controller.LoadSession(path)`:
  - reads YAML to `session.Session`
  - verifies `sess.SourceBranch == PRData.SourceBranch` (branch-bound)
  - applies config onto current `PRData` via `sess.ApplyToData(c.data)`:
    - toggles included/mode for files that still exist (by matching path)
    - replaces filters entirely
    - replaces additional context entirely
    - replaces prompt (preset if found in builtins, else template)

### Key invariants and failure modes

- **Branch mismatch kills load**: this is a strong invariant today.
  - UX implication: “load session” needs to tell user to switch branches or regenerate.
- **Session is a projection**: it can’t restore files that are no longer changed.
  - A file config for a path not present in `ChangedFiles` is simply ignored.
- **Prompt preset resolution is partial**:
  - session Apply currently checks only builtin presets.
  - if preset id is unknown, it falls back to template if present.

## Flow 4: Generate (PRData → API request → text)

This is the core “produce PR description” operation.

### What is the “generation context” exactly?

In code, it is `api.GenerateDescriptionRequest`:

- `SourceBranch`, `TargetBranch`
- `Files`: *included files only*, derived from `PRData.GetVisibleFiles()` and then `Included == true`
- `AdditionalContext`: all context items
- `Prompt`: `PRData.CurrentPrompt`

### Validation rules (where errors come from)

Controller’s `GenerateDescription()` enforces:

- at least one included visible file, else error `no files included for generation`

API service validates:

- source/target required
- files non-empty and at least one included

### Current implementation note

`internal/api` is currently a mock:

- sleeps 2 seconds
- emits a markdown-ish description derived from file list and context notes

This is important because UI work around “generating” and “spinner” should be designed as if it were a real network call, even if today it’s simulated.

## The big picture: why this core matters for the upcoming TUI rework

The modular TUI proposal (ticket 002 design doc) depends on several concrete facts from core code:

- **Stable IDs are paths**: selection, session config, and filter behavior all pivot on `FileChange.Path`.
- **Some operations are strict** (branch-bound session load) and must surface good UX messages.
- **Generation context is well-defined** and can be exported to clipboard by reusing the same builder inputs.
- **Most “errors” are user-correctable** (include at least one file, switch branches, provide valid preset), so the UI should prefer toasts and inline errors over fatal exits.
