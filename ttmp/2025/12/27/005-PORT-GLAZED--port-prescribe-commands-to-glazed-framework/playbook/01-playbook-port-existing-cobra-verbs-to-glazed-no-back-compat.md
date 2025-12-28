---
Title: 'Playbook: Port existing Cobra verbs to Glazed (no back-compat)'
Ticket: 005-PORT-GLAZED
Status: active
Topics:
    - glazed
    - prescribe
    - porting
    - cli
DocType: playbook
Intent: long-term
Owners: []
RelatedFiles:
    - Path: prescribe/cmd/prescribe/cmds/helpers/controller_from_layers.go
      Note: Controller init from parsed layers (used by all ports)
    - Path: prescribe/cmd/prescribe/cmds/root.go
      Note: Explicit root init + command tree registration
    - Path: prescribe/pkg/layers/existing_cobra_flags_layer.go
      Note: Wrapper to avoid re-adding root persistent flags
    - Path: prescribe/pkg/layers/repository.go
      Note: Repository layer (repo/target)
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-27T16:21:17.150661827-05:00
WhatFor: ""
WhenToUse: ""
---


# Playbook: Port existing Cobra verbs to Glazed (no back-compat)

## Purpose

This playbook is a **step-by-step recipe** for migrating existing Cobra-based commands (“verbs”) in `prescribe` to the Glazed framework.

This ticket intentionally does **not** preserve backwards compatibility. The goal is to:
- standardize parsing through Glazed `layers` (and `schema`/`fields` aliases),
- standardize output through Glazed structured output (for query commands),
- keep command initialization explicit (no Go `init()` ordering footguns),
- keep the implementation consistent so a new developer can port additional commands quickly.

## Environment Assumptions

- You’re working in the `prescribe` git worktree:
  - `/home/manuel/workspaces/2025-12-26/prescribe-import/prescribe`
- Go toolchain works (this repo uses a Go worktree setup; run git commands from `prescribe/`).
- You can run:
  - `go test ./... -count=1`
  - `go run ./cmd/prescribe ...`

## Key concepts (Glazed mental model)

### 1) Command types

- **BareCommand**: imperative commands; prints its own output (good for state changes)
- **GlazeCommand**: emits rows via a Glaze processor (good for queries / listings)

### 2) Layers and ParsedLayers

Glazed parses flags into a `*layers.ParsedLayers`. Each command defines a set of schema sections (“layers”), and then initializes a typed settings struct from a given layer slug:

```go
settings := &MySettings{}
if err := parsedLayers.InitializeStruct("my-layer-slug", settings); err != nil { ... }
```

### 3) Explicit initialization (no `init()`)

Do **not** use package-level `func init()` to build/register commands. Instead:
- each command package exposes `Init...Cmd()` functions that build cobra commands,
- each command group has an explicit `Init()` that calls `Init...Cmd()` and adds subcommands,
- root command calls these `Init()` functions in a deterministic order.

### 4) Root persistent flags (repo/target) and “existing flags” wrapper

`prescribe` currently has persistent root flags (eg `--repo`, `--target`). If you add a Glazed layer that *would* add these flags again, Cobra will panic (“flag redefined”).

Solution: wrap the schema section so it **does not re-add flags** but can still parse inherited flag values:

- `prescribe/pkg/layers/existing_cobra_flags_layer.go`

## What to port first (recommended order)

- Query/list commands (GlazeCommand):
  - `filter list`, `filter show`, `filter test`, `session show`
- Then state-changing commands (BareCommand):
  - `filter add/remove/clear`, `session init/load/save`, `file toggle`, `context add`

Skip `generate` for now (ticket decision).

## Commands

```bash
cd /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe

# Run unit tests for the module
go test ./... -count=1

# Smoke test a command you're porting
go run ./cmd/prescribe <group> <subcommand> --help
go run ./cmd/prescribe <group> <subcommand> --output json
```

## Porting recipe (copy/paste workflow)

### Step A: Pick the target and classify it

- If it’s a **query** (lists things / shows data): implement `cmds.GlazeCommand`
- If it **mutates state** (writes session / toggles file): implement `cmds.BareCommand`

### Step B: Build the Glazed command definition (schema)

In the relevant file (eg `cmd/prescribe/cmds/filter/foo.go`):

1) Create a `NewXxxCommand()` constructor.
2) Attach layers via `cmds.WithLayersList(...)`.
3) For repo/target, create the repository layer and wrap it:

```go
repoLayer, _ := prescribe_layers.NewRepositoryLayer()
repoLayerExisting, _ := prescribe_layers.WrapAsExistingCobraFlagsLayer(repoLayer)
```

4) If the command needs flags, prefer a **command-specific** layer (no back-compat constraints):

```go
layer, _ := schema.NewSection(
  "my-slug",
  "My Section",
  schema.WithFields(
    fields.New("flag", fields.TypeString, fields.WithDefault(""), fields.WithHelp("...")),
  ),
)
```

### Step C: Implement the execution method

#### For GlazeCommand

```go
func (c *MyCmd) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
  // Initialize controller
  ctrl, err := helpers.NewInitializedControllerFromParsedLayers(parsedLayers)
  if err != nil { return err }

  // Load session if needed
  helpers.LoadDefaultSessionIfExists(ctrl)

  // Emit rows
  row := types.NewRow(types.MRP("field", "value"))
  return gp.AddRow(ctx, row)
}
```

#### For BareCommand

```go
func (c *MyCmd) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
  ctrl, err := helpers.NewInitializedControllerFromParsedLayers(parsedLayers)
  if err != nil { return err }

  // Load / mutate / save
  helpers.LoadDefaultSessionIfExists(ctrl)
  // ...
  return nil
}
```

### Step D: Expose an explicit Init function for Cobra wiring

Replace any direct `var XxxCmd = &cobra.Command{...}` construction. Use:

```go
var XxxCmd *cobra.Command

func InitXxxCmd() error {
  glazedCmd, err := NewXxxCommand()
  if err != nil { return err }

  cobraCmd, err := cli.BuildCobraCommand(glazedCmd,
    cli.WithParserConfig(cli.CobraParserConfig{
      MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
    }),
  )
  if err != nil { return err }

  XxxCmd = cobraCmd
  return nil
}
```

### Step E: Register it in the group Init()

Example: `cmd/prescribe/cmds/filter/filter.go`:

```go
func Init() error {
  // ...
  if err := InitXxxCmd(); err != nil { return err }
  FilterCmd.AddCommand(XxxCmd)
  return nil
}
```

### Step F: Tests + smoke test

```bash
cd prescribe
go test ./... -count=1
go run ./cmd/prescribe <group> <subcommand> --help
go run ./cmd/prescribe <group> <subcommand> --output json
```

### Step G: Commit + ticket bookkeeping (docmgr)

1) Commit code from the `prescribe/` directory.
2) Update ticket tasks + changelog + relate files.
3) Commit docs separately.

## Pitfalls / gotchas (read this first)

### Go init ordering

Avoid `init()` entirely for command wiring. The crash we hit early on was caused by one file trying to register a subcommand before another file had initialized it.

### Root persistent flags vs Glazed layers

If you accidentally re-add a flag already defined on the root command, Cobra will error/panic.

Use `WrapAsExistingCobraFlagsLayer(...)` whenever you attach a layer whose flags already exist as persistent flags.

### Broken pipe during `head`

When you pipe JSON output to `head`, the consumer can close early and you may see a “broken pipe” exit. That’s expected for this kind of test.

## Exit Criteria

- Target command is implemented as a Glazed command (BareCommand or GlazeCommand).
- No `init()` ordering reliance introduced.
- `go test ./... -count=1` passes.
- `go run ./cmd/prescribe <cmd> --help` shows Glazed output flags (for GlazeCommand) or the intended flags (for BareCommand).
- `go run ./cmd/prescribe <cmd> --output json` works for GlazeCommands.
- A code commit exists, and ticket docs are updated (task state / changelog / diary / related files).

## Notes

- Current “base” building blocks in this repo:
  - `prescribe/pkg/layers/*` (repository/session/filter/generation sections)
  - `prescribe/pkg/layers/existing_cobra_flags_layer.go` (don’t re-add root flags)
  - `cmd/prescribe/cmds/helpers/controller_from_layers.go` (controller init from parsed layers)
  - `cmd/prescribe/cmds/root.go` (`NewRootCmd`, `InitRootCmd` for explicit command wiring)
