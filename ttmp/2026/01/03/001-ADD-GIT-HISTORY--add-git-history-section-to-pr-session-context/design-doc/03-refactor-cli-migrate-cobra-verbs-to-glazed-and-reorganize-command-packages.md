---
Title: 'Refactor CLI: migrate Cobra verbs to Glazed and reorganize command packages'
Ticket: 001-ADD-GIT-HISTORY
Status: active
Topics:
    - git
    - pr
    - context
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/prescribe/cmds/context/git/root.go
      Note: Example of new directory-per-subgroup registration root (context git)
    - Path: cmd/prescribe/cmds/generate.go
      Note: Example of Glazed command pattern
    - Path: cmd/prescribe/cmds/root.go
      Note: Root command tree wiring (explicit constructors; no group Init ordering reliance)
ExternalSources: []
Summary: Standardize on Glazed command pattern for all verbs and reorganize command code into a directory-per-subgroup, one-file-per-verb layout.
LastUpdated: 2026-01-04T15:56:23.765933657-05:00
WhatFor: Make the CLI codebase consistent, discoverable, and easier to extend by aligning all verbs on the Glazed command pattern and standard file layout.
WhenToUse: When refactoring the CLI command tree, adding new verbs/subgroups, or cleaning up inconsistent Cobra-only commands.
---


# Refactor CLI: migrate Cobra verbs to Glazed and reorganize command packages

## Executive Summary

`prescribe` already uses Glazed for many commands, but the command implementation patterns and file layout are inconsistent: some verbs are plain Cobra handlers, others are Glazed `cmds.*Command` wrappers, and nested subcommands (e.g., `context git history ...`) live in a single file.

This doc proposes:
1) Convert *all verbs* to Glazed-style commands (using `cmds.CommandDescription` + `cli.BuildCobraCommand`), even when the command is “just print text”.
2) Reorganize command source files into a consistent hierarchy:
   - `cmd/prescribe/cmds/<group>/<subgroup...>/<verb>.go`
   - one file per verb; one directory per subgroup.

This yields a CLI that is easier to navigate, test, and extend, and sets the stage for larger refactors (like plugin-style context providers) without the CLI becoming brittle.

## Problem Statement

We want to evolve `prescribe` without the CLI becoming the bottleneck. Today:
- Command implementations follow multiple patterns (Glazed vs plain Cobra), which makes it harder to add new commands consistently.
- Historically, nested command trees lived in a single large file (e.g., `cmd/prescribe/cmds/context/git.go`), which:
  - discourages fine-grained tests and review,
  - makes “one change per PR” harder,
  - makes it harder to see “what commands exist” by browsing the filesystem.
- Different commands have different layer/flag patterns, even when they share repo/session plumbing.

We want:
- a single, repeatable command template (Glazed-first),
- a directory layout that mirrors the CLI hierarchy,
- commands that all support the same “layered config” idioms (repo settings, optional output formats, debug flag patterns),
- minimal behavior change (mechanical refactor where possible).

## Proposed Solution

### 1) Standardize all verbs on Glazed commands

Use the “Build Your First Glazed Command” patterns (`glaze help build-first-command`) as the baseline:
- every verb file defines a `New...Command()` constructor returning a Glazed command struct embedding `*cmds.CommandDescription`
- every verb file exports a `New<Verb>CobraCommand()` constructor that integrates with Cobra via `cli.BuildCobraCommand(...)`

Command types:
- Default to **BareCommand** (`Run`) for action/printf-style output.
- Use **GlazeCommand** (`RunIntoGlazeProcessor`) only when the command’s primary output is structured rows intended for `--output ...` formats (list/show/token-count/etc).

Even for “plain text” verbs, keeping the Glazed wrapper gives a consistent place to:
- define flags using `parameters.NewParameterDefinition` or `schema.NewSection`,
- reuse repo/session layers,
- share middleware setup (`cli.CobraCommandDefaultMiddlewares`).

### 2) Command source layout mirrors CLI structure

Adopt a strict filesystem-to-CLI mapping:

```
cmd/prescribe/cmds/
  root.go                             # root command wiring (groups + root-level verbs)
  <root-verb>.go                      # root-level verbs (generate/create/tui)
  <group>/
    root.go                           # defines the group command and registers subcommands
    <subgroup>/
      root.go                         # defines the subgroup command and registers subcommands
      <verb>.go
      <subgroup2>/
        root.go
        <verb>.go
```

Interpretation of the user’s desired shape `cmd/prescribe/cmds/$GROUP/$GROUP2/$VERB.go`:
- `$GROUP` is the top-level Cobra group (e.g., `context`, `filter`, `session`, `file`, `tokens`).
- `$GROUP2` is the subgroup path (possibly multiple segments: `git/history`).
- `$VERB.go` is the leaf command file (one verb per file).

For verbs directly under a group, place them next to the group’s `root.go`:
- `cmd/prescribe/cmds/<group>/<verb>.go`

### 3) Package structure and registration

Each directory becomes its own Go package to avoid giant files while keeping import structure simple.

**Important update:** group files are *only* `root.go` and they do all command registration. We do **not** use `Init()` methods (or any `Init...()` registration helpers). The registration is done by constructing the cobra.Command tree directly in `root.go`.

Example for `prescribe context git history show`:
- `cmd/prescribe/cmds/context/root.go` (package `context`): defines `NewContextCmd()` and registers `add` + `git` subcommands.
- `cmd/prescribe/cmds/context/git/root.go` (package `git`): defines `NewGitCmd()` and registers `list/remove/clear/add` + `history` subgroup.
- `cmd/prescribe/cmds/context/git/history/root.go` (package `history`): defines `NewHistoryCmd()` and registers `show/enable/disable/set`.
- `cmd/prescribe/cmds/context/git/history/show.go` (package `history`): defines the Glazed `show` verb and exports `NewShowCobraCommand()` (or similar) returning `*cobra.Command`.

Registration flow (top-down, no Init methods):
- `cmd/prescribe/cmds/root.go` calls constructors for top-level groups and attaches them:
  - `rootCmd.AddCommand(context.NewContextCmd(...))`, `rootCmd.AddCommand(session.NewSessionCmd(...))`, etc.
- each group’s `root.go` constructor registers its verbs/subgroups using `AddCommand(...)`.

This makes it mechanically obvious where a verb lives and how it is wired.

### 4) Constructor naming conventions (final)

Group and subgroup constructors:
- `New<Group>Cmd() (*cobra.Command, error)` in `cmd/prescribe/cmds/<group>/root.go`
- `New<Subgroup>Cmd() (*cobra.Command, error)` in `cmd/prescribe/cmds/<group>/<subgroup>/root.go`

Leaf verb constructors:
- `New<Verb>CobraCommand() (*cobra.Command, error)` in `cmd/prescribe/cmds/<group>/.../<verb>.go`
- If the file defines a Glazed command struct, name it `New<Something>Command()` (e.g. `NewFilterListCommand()`), returning the Glazed command implementation that embeds `*cmds.CommandDescription`.

### 4) Standard layers (repo/session) and command settings

We should standardize “common layers” usage:
- repo/target settings: `prescribe_layers.NewRepositoryLayer()` + `WrapAsExistingCobraFlagsLayer(...)`
- optional command settings/debug: adopt Glazed’s command settings layer where appropriate (see tutorial)

Note: most verbs should include the repository layer, but because `repo`/`target` are persistent flags on the root command you typically want `WrapAsExistingCobraFlagsLayer(...)` to avoid "flag redefined" errors while still parsing inherited values.
As of the current refactor state, all commands in the migrated groups follow this pattern; use the mapping table below as the source-of-truth template for new verbs/subgroups.

### 5) Testing and ergonomics

Refactor should be behavior-preserving:
- keep command names/flags stable unless we explicitly choose to break them
- existing smoke tests (`test-scripts/*`, `test/*`) should keep passing, with updates only if file paths change in internal code (not expected)

We should add a minimal “command tree smoke” test:
- run `prescribe --help` and `prescribe <group> --help` and ensure the new subcommands show up

## Design Decisions

1) **Use Glazed wrappers even for non-table output**
   - Rationale: consistent construction, flags, layers, and middleware; predictable wiring.

2) **Directory-per-subgroup, file-per-verb**
   - Rationale: mirrors the CLI structure; reduces large files; improves discoverability and review diffs.

3) **`root.go` owns registration (no `Init()` methods)**
   - Rationale: keeps registration local to the package; avoids scattered init helpers and ordering reliance.

## Alternatives Considered

1) Keep the current layout and only convert the few remaining Cobra-only verbs
   - Rejected: layout remains inconsistent; future growth will repeat the problem.

2) One package per top-level group only (no subpackages)
   - Rejected: forces large files for nested subcommand trees; loses the “directory mirrors CLI” property.

3) Runtime-generated command tree from a data file
   - Rejected: harder to test and reason about; loses static typing and Go navigation benefits.

## Implementation Plan

1) **Document the mapping**
   - Mapping table (current refactor target state):

| CLI command | Type | File | Constructor |
| --- | --- | --- | --- |
| `prescribe generate` | Bare | `cmd/prescribe/cmds/generate.go` | `NewGenerateCobraCommand()` |
| `prescribe create` | Bare | `cmd/prescribe/cmds/create.go` | `NewCreateCobraCommand()` |
| `prescribe tui` | Bare | `cmd/prescribe/cmds/tui.go` | `NewTuiCobraCommand()` |
| `prescribe context` | Group | `cmd/prescribe/cmds/context/root.go` | `NewContextCmd()` |
| `prescribe context add` | Bare | `cmd/prescribe/cmds/context/add.go` | `NewAddCobraCommand()` |
| `prescribe context git` | Group | `cmd/prescribe/cmds/context/git/root.go` | `NewGitCmd()` |
| `prescribe context git list` | Bare | `cmd/prescribe/cmds/context/git/list.go` | `NewListCobraCommand()` |
| `prescribe context git remove` | Bare | `cmd/prescribe/cmds/context/git/remove.go` | `NewRemoveCobraCommand()` |
| `prescribe context git clear` | Bare | `cmd/prescribe/cmds/context/git/clear.go` | `NewClearCobraCommand()` |
| `prescribe context git add` | Group | `cmd/prescribe/cmds/context/git/add/root.go` | `NewAddCmd()` |
| `prescribe context git add commit` | Bare | `cmd/prescribe/cmds/context/git/add/commit.go` | `NewCommitCobraCommand()` |
| `prescribe context git add commit-patch` | Bare | `cmd/prescribe/cmds/context/git/add/commit_patch.go` | `NewCommitPatchCobraCommand()` |
| `prescribe context git add file-at` | Bare | `cmd/prescribe/cmds/context/git/add/file_at.go` | `NewFileAtCobraCommand()` |
| `prescribe context git add file-diff` | Bare | `cmd/prescribe/cmds/context/git/add/file_diff.go` | `NewFileDiffCobraCommand()` |
| `prescribe context git history` | Group | `cmd/prescribe/cmds/context/git/history/root.go` | `NewHistoryCmd()` |
| `prescribe context git history show` | Bare | `cmd/prescribe/cmds/context/git/history/show.go` | `NewShowCobraCommand()` |
| `prescribe context git history enable` | Bare | `cmd/prescribe/cmds/context/git/history/enable.go` | `NewEnableCobraCommand()` |
| `prescribe context git history disable` | Bare | `cmd/prescribe/cmds/context/git/history/disable.go` | `NewDisableCobraCommand()` |
| `prescribe context git history set` | Bare | `cmd/prescribe/cmds/context/git/history/set.go` | `NewSetCobraCommand()` |
| `prescribe filter` | Group | `cmd/prescribe/cmds/filter/root.go` | `NewFilterCmd()` |
| `prescribe filter add` | Bare | `cmd/prescribe/cmds/filter/add.go` | `NewAddCobraCommand()` |
| `prescribe filter remove` | Bare | `cmd/prescribe/cmds/filter/remove.go` | `NewRemoveCobraCommand()` |
| `prescribe filter clear` | Bare | `cmd/prescribe/cmds/filter/clear.go` | `NewClearCobraCommand()` |
| `prescribe filter list` | Glaze | `cmd/prescribe/cmds/filter/list.go` | `NewListCobraCommand()` |
| `prescribe filter show` | Glaze | `cmd/prescribe/cmds/filter/show.go` | `NewShowCobraCommand()` |
| `prescribe filter test` | Glaze | `cmd/prescribe/cmds/filter/test.go` | `NewTestCobraCommand()` |
| `prescribe filter preset` | Group | `cmd/prescribe/cmds/filter/preset/root.go` | `NewPresetCmd()` |
| `prescribe filter preset list` | Glaze | `cmd/prescribe/cmds/filter/preset/list.go` | `NewListCobraCommand()` |
| `prescribe filter preset save` | Bare | `cmd/prescribe/cmds/filter/preset/save.go` | `NewSaveCobraCommand()` |
| `prescribe filter preset apply` | Bare | `cmd/prescribe/cmds/filter/preset/apply.go` | `NewApplyCobraCommand()` |
| `prescribe session` | Group | `cmd/prescribe/cmds/session/root.go` | `NewSessionCmd()` |
| `prescribe session init` | Bare | `cmd/prescribe/cmds/session/init.go` | `NewInitCobraCommand()` |
| `prescribe session load` | Bare | `cmd/prescribe/cmds/session/load.go` | `NewLoadCobraCommand()` |
| `prescribe session save` | Bare | `cmd/prescribe/cmds/session/save.go` | `NewSaveCobraCommand()` |
| `prescribe session show` | Glaze | `cmd/prescribe/cmds/session/show.go` | `NewShowCobraCommand()` |
| `prescribe session token-count` | Glaze | `cmd/prescribe/cmds/session/token_count.go` | `NewTokenCountCobraCommand()` |
| `prescribe file` | Group | `cmd/prescribe/cmds/file/root.go` | `NewFileCmd()` |
| `prescribe file toggle` | Bare | `cmd/prescribe/cmds/file/toggle.go` | `NewToggleCobraCommand()` |
| `prescribe tokens` | Group | `cmd/prescribe/cmds/tokens/root.go` | `NewTokensCmd()` |
| `prescribe tokens count-xml` | Glaze | `cmd/prescribe/cmds/tokens/count_xml.go` | `NewCountXMLCobraCommand()` |

2) **Move nested `context git ...` from single Cobra file into subpackages**
   - Convert verbs to Glazed commands:
     - repo layer + args/flags as parameter layers
     - `Run()` uses `helpers.NewInitializedControllerFromParsedLayers`
     - load default session, mutate, save
   - Replace `InitGitCmd` style code with `NewGitCmd()` in `root.go` files.

3) **Restructure other groups**
   - Mechanical moves into the new directory layout.
   - Replace group `Init()` functions with `New<Group>Cmd()` constructors in `root.go`.

4) **Update root registration imports**
   - `cmd/prescribe/cmds/root.go` should import only top-level group packages (`context`, `filter`, `session`, `file`, `tokens`) as it does today.

5) **Run tests**
   - `GOWORK=off go test ./...`
   - `bash test-scripts/test-cli.sh`
   - `bash test-scripts/test-all.sh`

6) **Follow repo workflow**
   - commit code changes (no docs) with a focused message
   - update ticket tasks + diary + changelog
   - commit docs

## Open Questions

1) Do we require that all “action commands” also support `--output ...` formats, or only structured/list commands?
2) Should we standardize on `schema.NewSection` for args/flags everywhere, or allow `parameters.NewParameterDefinition` for simple flags?
3) Should the “root subgroup” folder be named `root/` or `default/` (to mirror `schema.DefaultSlug`)?

## References

- `glaze help build-first-command` (Glazed command tutorial)
- Current root initialization: `cmd/prescribe/cmds/root.go`
- Example Glazed commands already in-tree:
  - `cmd/prescribe/cmds/generate.go`
  - `cmd/prescribe/cmds/session/show.go`
