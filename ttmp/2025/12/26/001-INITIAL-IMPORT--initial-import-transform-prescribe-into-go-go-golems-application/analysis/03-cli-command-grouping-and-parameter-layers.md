---
Title: CLI command grouping and parameter layers
Ticket: 001-INITIAL-IMPORT
Status: active
Topics:
    - migration
    - refactoring
    - ci-cd
    - go-module
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-26T18:19:48.501995678-05:00
WhatFor: ""
WhenToUse: ""
---

# CLI command grouping and parameter layers

## Context / goal

We want the `prescribe` CLI to follow the same **command organization pattern** as pinocchio:

- **Command groups are folders** (eg `cmd/pinocchio/cmds/tokens/…`)
- Each group is exposed on the CLI as a **parent command** (`pinocchio tokens …`)
- Leaf subcommands live under the group (`pinocchio tokens encode|decode|…`)

Before this work, prescribe already had *folders* like `cmd/prescribe/cmds/filter/…`, but the commands were still registered *flat on the root* (`prescribe add-filter`, `prescribe clear-filters`, …). This doc captures:

- the **CLI verbs** (what commands exist, and how they should be grouped),
- the **settings/flags inventory**, and
- the emerging **parameter layers** that can be reused across commands.

## Pinocchio reference (what we’re mirroring)

Pinocchio’s root (`pinocchio/cmd/pinocchio/main.go`) registers groups via exported helpers like:

- `tokens.RegisterCommands(rootCmd)`
- `helpers.RegisterHelperCommands(rootCmd)`
- `kagi.RegisterKagiCommands()` returning a group command

The key point is: **groups are packages** and provide a single entrypoint that attaches leaf commands.

## Prescribe command tree (target shape)

### Global flags (root persistent flags)

- `--repo, -r`: path to git repository (default `"."`)
- `--target, -t`: target branch (default auto / empty means “main or master”)

### Root-level commands (singletons)

- `prescribe tui`: interactive TUI
- `prescribe generate`: generate PR description (non-interactive)

### Command groups (parents)

- `prescribe filter …`
  - `filter add`
  - `filter list`
  - `filter remove <index|name>`
  - `filter clear`
  - `filter test`
  - `filter show` (shows filtered-out files)
- `prescribe session …`
  - `session init`
  - `session load [path]`
  - `session save [path]`
  - `session show`
- `prescribe file …`
  - `file toggle <path>`
- `prescribe context …`
  - `context add [file-path] --note "…"`

## Command/flag inventory (what exists today)

### `tui`

- **Consumes**: repo layer (`--repo`, `--target`)
- **Implicit behavior**: best-effort loads default session, then starts Bubbletea.

### `generate`

- **Consumes**: repo layer (`--repo`, `--target`)
- **Flags**:
  - `--output, -o`: output file (default stdout)
  - `--prompt, -p`: prompt override
  - `--preset`: preset id
  - `--session, -s`: load session file before generating

### `filter add`

- **Flags**:
  - `--name, -n` (required)
  - `--description, -d`
  - `--exclude, -e` (repeatable)
  - `--include, -i` (repeatable)
- **Important semantics**:
  - should load an existing session before mutating/saving, otherwise it clobbers prior state.

### `filter test`

- **Flags**:
  - `--name, -n` (for display; default `"test"`)
  - `--exclude, -e`
  - `--include, -i`
- **Semantics**:
  - does not persist (no save); it’s a pure “what would happen” operation.

### `session show`

- **Flags**:
  - `--yaml, -y`: dump session YAML instead of human-readable output.

### `session init`

- **Flags**:
  - `--save`: save after init
  - `--path, -p`: where to save (default: app default session path)

## Parameter layers (reusable settings “bundles”)

This repo is not yet using geppetto/glazed parameter layers, but we can still structure Cobra code in the same spirit: **a small number of reusable flag bundles + extractors**.

### Layer 1: Repo / target layer (global)

**Used by**: almost all commands.

- Flags live on root persistent flags.
- Extraction should be reusable.

Implemented now as:
- `cmd/prescribe/cmds/helpers.GetRepoParams(cmd)`

### Layer 2: Controller initialization layer (repo → controller + Initialize)

**Used by**: almost all commands.

Implemented now as:
- `cmd/prescribe/cmds/helpers.NewInitializedController(cmd)`

### Layer 3: Session loading semantics (best-effort vs required)

There are two patterns:

- **Best-effort load**: commands should work even when no session exists
  - e.g. `tui`, `filter list`, `filter show`, `file toggle` (can still operate on current git state)
- **Required load**: command needs existing persisted state
  - e.g. `filter remove`, `filter clear`

Implemented now as:
- `helpers.LoadDefaultSessionIfExists(ctrl)`
- `helpers.LoadDefaultSession(ctrl) error`

### Layer 4: Filter specification layer (name/description/include/exclude)

Currently spread across `filter/add.go` and `filter/test.go`. Potential refactor:

- Shared `FilterSpec` struct
- Shared `BindFilterSpecFlags(cmd)` + `ValidateFilterSpec(spec)`

This would remove duplication and enforce consistent validation.

### Layer 5: Prompt selection layer (prompt vs preset)

Currently local to `generate`. Potential refactor:

- `PromptParams{ PromptText, PresetID }`
- `ApplyPromptSelection(ctrl, params)` to unify behavior

### Layer 6: Output layer (stdout vs file)

Currently local to `generate`. Potential refactor:

- `OutputParams{ OutputFile }`
- `WriteOutput(outputFile, content)` helper

## Compatibility notes (CLI surface change)

This regrouping changes command invocation:

- old: `prescribe add-filter`
- new: `prescribe filter add`

If we want backwards compatibility, we could add hidden root-level shims that forward to the new group subcommands. For now, the intent is to converge on the pinocchio-style grouped surface and then update scripts/docs accordingly.

