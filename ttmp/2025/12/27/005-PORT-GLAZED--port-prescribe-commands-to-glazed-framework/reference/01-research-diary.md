---
Title: Research Diary
Ticket: 005-PORT-GLAZED
Status: active
Topics:
    - glazed
    - prescribe
    - porting
    - cli
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Step-by-step research diary documenting the analysis of Glazed framework patterns and Prescribe command structure
LastUpdated: 2025-12-27T15:08:53.002761943-05:00
WhatFor: Tracking research process and findings for porting Prescribe commands to Glazed
WhenToUse: Reference during implementation to understand decisions and context
---

# Research Diary

## Goal

Document the research process for understanding how to port Prescribe CLI commands to the Glazed framework. This diary captures what was learned, what files were analyzed, and what patterns were identified.

## Step 1: Understanding Glazed Command Structure

**What I did:**
- Read `glazed/pkg/doc/tutorials/05-build-first-command.md` (785 lines)
- Analyzed the tutorial's command structure patterns
- Identified key API symbols and patterns

**What I learned:**

### Core Glazed Command Pattern

1. **Command Struct**: Embeds `*cmds.CommandDescription`
   ```go
   type ListUsersCommand struct {
       *cmds.CommandDescription
   }
   ```

2. **Settings Struct**: Maps CLI flags to Go fields using `glazed.parameter` tags
   ```go
   type ListUsersSettings struct {
       Limit      int    `glazed.parameter:"limit"`
       NameFilter string `glazed.parameter:"name-filter"`
       Active     bool   `glazed.parameter:"active-only"`
   }
   ```

3. **Interface Implementation**: `RunIntoGlazeProcessor` method
   ```go
   func (c *ListUsersCommand) RunIntoGlazeProcessor(
       ctx context.Context,
       parsedLayers *layers.ParsedLayers,
       gp middlewares.Processor,
   ) error {
       settings := &ListUsersSettings{}
       if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
           return err
       }
       // ... business logic ...
       // Output structured data as rows
       for _, item := range items {
           row := types.NewRow(
               types.MRP("key", value),
           )
           if err := gp.AddRow(ctx, row); err != nil {
               return err
           }
       }
       return nil
   }
   ```

4. **Command Construction**: Uses `cmds.NewCommandDescription` with layers
   ```go
   cmdDesc := cmds.NewCommandDescription(
       "command-name",
       cmds.WithShort("Short description"),
       cmds.WithLong("Long description"),
       cmds.WithFlags(
           parameters.NewParameterDefinition(
               "param-name",
               parameters.ParameterTypeString,
               parameters.WithDefault("default"),
               parameters.WithHelp("Help text"),
               parameters.WithShortFlag("p"),
           ),
       ),
       cmds.WithLayersList(glazedLayer, customLayer),
   )
   ```

5. **Cobra Integration**: Uses `cli.BuildCobraCommand`
   ```go
   cobraCmd, err := cli.BuildCobraCommand(glazedCmd,
       cli.WithParserConfig(cli.CobraParserConfig{
           ShortHelpLayers: []string{layers.DefaultSlug},
           MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
       }),
   )
   ```

**Key API Symbols Identified:**
- `github.com/go-go-golems/glazed/pkg/cmds` - Core command types
- `github.com/go-go-golems/glazed/pkg/cmds/layers` - Parameter layers
- `github.com/go-go-golems/glazed/pkg/cmds/parameters` - Parameter definitions
- `github.com/go-go-golems/glazed/pkg/cli` - Cobra integration
- `github.com/go-go-golems/glazed/pkg/middlewares` - Data processing
- `github.com/go-go-golems/glazed/pkg/types` - Structured data types
- `github.com/go-go-golems/glazed/pkg/settings` - Built-in layers

**Files Referenced:**
- `glazed/pkg/doc/tutorials/05-build-first-command.md`

## Step 2: Understanding Custom Layer Creation

**What I did:**
- Read `glazed/pkg/doc/tutorials/custom-layer.md` (1019 lines)
- Analyzed the logging layer example
- Understood layer composition patterns

**What I learned:**

### Layer Creation Pattern

1. **Settings Struct**: Type-safe configuration with struct tags
   ```go
   type LoggingSettings struct {
       Level      string `glazed.parameter:"log-level"`
       Format     string `glazed.parameter:"log-format"`
       File       string `glazed.parameter:"log-file"`
       WithCaller bool   `glazed.parameter:"with-caller"`
       Verbose    bool   `glazed.parameter:"verbose"`
   }
   ```

2. **Layer Definition**: Uses `layers.NewParameterLayer`
   ```go
   func NewLoggingLayer() (layers.ParameterLayer, error) {
       return layers.NewParameterLayer(
           LoggingSlug,
           "Logging Configuration",
           layers.WithParameterDefinitions(
               parameters.NewParameterDefinition(
                   "log-level",
                   parameters.ParameterTypeChoice,
                   parameters.WithHelp("Set the logging level"),
                   parameters.WithDefault("info"),
                   parameters.WithChoices("debug", "info", "warn", "error", "fatal", "panic"),
               ),
               // ... more parameters ...
           ),
       )
   }
   ```

3. **Helper Functions**: Extract settings from parsed layers
   ```go
   func GetLoggingSettings(parsedLayers *layers.ParsedLayers) (*LoggingSettings, error) {
       settings := &LoggingSettings{}
       if err := parsedLayers.InitializeStruct(LoggingSlug, settings); err != nil {
           return nil, fmt.Errorf("failed to initialize logging settings: %w", err)
       }
       return settings, nil
   }
   ```

4. **Layer Usage in Commands**: Add to command layers list
   ```go
   cmds.WithLayersList(loggingLayer, glazedLayer)
   ```

**Key Patterns:**
- Layers encapsulate related parameters
- Settings structs provide type-safe access
- Helper functions simplify initialization
- Layers can be composed across commands

**Files Referenced:**
- `glazed/pkg/doc/tutorials/custom-layer.md`

## Step 3: Understanding AppConfig API

**What I did:**
- Read `glazed/cmd/examples/appconfig-parser/main.go` (89 lines)
- Analyzed `glazed/pkg/appconfig/parser.go`
- Read `glazed/pkg/cmds/schema/schema.go` and `glazed/pkg/cmds/fields/fields.go`
- Understood the new appconfig pattern

**What I learned:**

### AppConfig Pattern (Newer API)

1. **Grouped Settings Struct**: Top-level struct with nested settings
   ```go
   type AppSettings struct {
       Redis RedisSettings
       DB    DBSettings
   }
   
   type RedisSettings struct {
       Host string `glazed.parameter:"host"`
       Port int    `glazed.parameter:"port"`
   }
   ```

2. **Schema Sections**: Using `schema.NewSection` (wraps `layers.NewParameterLayer`)
   ```go
   redisLayer := schema.NewSection(
       string(RedisSlug),
       "Redis",
       schema.WithPrefix("redis-"),
       schema.WithFields(
           fields.New("host", fields.TypeString, fields.WithDefault("127.0.0.1")),
           fields.New("port", fields.TypeInteger, fields.WithDefault(6379)),
       ),
   )
   ```

3. **Parser Registration**: Register layers with binders
   ```go
   parser, err := appconfig.NewParser[AppSettings](
       appconfig.WithValuesForLayers(map[string]map[string]interface{}{
           string(RedisSlug): {"host": "cache.local", "port": 6380},
       }),
   )
   
   parser.Register(RedisSlug, redisLayer, func(t *AppSettings) any { 
       return &t.Redis 
   })
   ```

4. **Parsing**: Single call returns populated struct
   ```go
   cfg, err := parser.Parse()
   ```

**Key API Symbols:**
- `github.com/go-go-golems/glazed/pkg/appconfig` - New parser API
- `github.com/go-go-golems/glazed/pkg/cmds/schema` - Schema sections (alias for layers)
- `github.com/go-go-golems/glazed/pkg/cmds/fields` - Field definitions (alias for parameters)
- `appconfig.LayerSlug` - Type-safe layer identifiers
- `appconfig.NewParser[T]` - Generic parser constructor
- `parser.Register(slug, layer, binder)` - Register layer with binder function
- `parser.Parse()` - Execute middleware chain and populate struct

**Files Referenced:**
- `glazed/cmd/examples/appconfig-parser/main.go`
- `glazed/pkg/appconfig/parser.go`
- `glazed/pkg/appconfig/options.go`
- `glazed/pkg/cmds/schema/schema.go`
- `glazed/pkg/cmds/fields/fields.go`

**Key Differences:**
- AppConfig uses a grouped settings struct pattern
- Schema/Fields are aliases for Layers/Parameters (newer naming)
- Parser pattern is more declarative but may be overkill for simple commands
- Traditional layer pattern is more flexible for command composition

## Step 4: Analyzing Prescribe Commands

**What I did:**
- Listed all command directories in `prescribe/cmd/prescribe/cmds/`
- Read all command implementations
- Cataloged all flags and their usage patterns
- Identified common patterns across commands

**What I learned:**

### Command Structure

**Root Command Groups:**
1. `filter` - Filter management (6 subcommands)
2. `session` - Session management (4 subcommands)
3. `file` - File operations (1 subcommand)
4. `context` - Context management (1 subcommand)
5. Root-level: `generate`, `tui`

**Persistent Flags (Root):**
- `--repo, -r` (string, default: ".")
- `--target, -t` (string, default: "")

**Command Inventory:**

1. **filter add**
   - `--name, -n` (string, required)
   - `--description, -d` (string)
   - `--exclude, -e` ([]string)
   - `--include, -i` ([]string)

2. **filter list**
   - No flags (reads from session)

3. **filter remove**
   - Takes index or name as argument

4. **filter clear**
   - No flags

5. **filter test**
   - `--name, -n` (string, default: "test")
   - `--exclude, -e` ([]string)
   - `--include, -i` ([]string)

6. **filter show**
   - No flags (shows filtered files)

7. **session init**
   - `--save` (bool)
   - `--path, -p` (string)

8. **session load**
   - Takes optional path as argument

9. **session save**
   - Takes optional path as argument

10. **session show**
    - `--yaml, -y` (bool)

11. **file toggle**
    - Takes file path as argument

12. **context add**
    - `--note` (string)
    - Takes optional file path as argument

13. **generate**
    - `--output, -o` (string)
    - `--prompt, -p` (string)
    - `--preset` (string)
    - `--session, -s` (string)

14. **tui**
    - No flags (uses persistent flags)

**Common Patterns:**
- All commands use `helpers.NewInitializedController(cmd)` to get controller
- Most commands call `helpers.LoadDefaultSessionIfExists(ctrl)` or `helpers.LoadDefaultSession(ctrl)`
- Many commands save session after modification
- Commands output human-readable text (not structured data)
- Controller pattern abstracts git operations and session management

**Files Referenced:**
- `prescribe/cmd/prescribe/cmds/root.go`
- `prescribe/cmd/prescribe/cmds/filter/*.go` (6 files)
- `prescribe/cmd/prescribe/cmds/session/*.go` (4 files)
- `prescribe/cmd/prescribe/cmds/file/toggle.go`
- `prescribe/cmd/prescribe/cmds/context/add.go`
- `prescribe/cmd/prescribe/cmds/generate.go`
- `prescribe/cmd/prescribe/cmds/tui.go`
- `prescribe/cmd/prescribe/cmds/helpers/controller.go`

## Step 5: Identifying Layer Candidates

**What I did:**
- Analyzed flag patterns across all commands
- Grouped related flags by functionality
- Identified reusable configuration patterns

**What I learned:**

### Potential Layers

1. **Repository Layer** (already exists as persistent flags)
   - `repo` (string) - Repository path
   - `target` (string) - Target branch
   - Used by: ALL commands

2. **Session Layer** (could be extracted)
   - `session-path` (string) - Custom session path
   - `auto-save` (bool) - Auto-save after operations
   - Used by: session commands, filter commands, file commands, context commands

3. **Output Layer** (for structured output commands)
   - `output-format` (choice: text, json, yaml, csv)
   - `output-file` (string)
   - Used by: generate, session show, filter list, filter show

4. **Filter Layer** (for filter operations)
   - `filter-name` (string)
   - `filter-description` (string)
   - `exclude-patterns` ([]string)
   - `include-patterns` ([]string)
   - Used by: filter add, filter test

5. **Generation Layer** (for generate command)
   - `prompt` (string)
   - `preset` (string)
   - `load-session` (string)
   - Used by: generate

**Key Insight:**
- Most commands are stateful (operate on session)
- Many commands could benefit from structured output (JSON/YAML)
- Repository configuration is already shared but could be a layer
- Session management is implicit but could be explicit layer

## Step 6: Mapping Commands to Glazed Patterns

**What I did:**
- Categorized each command by output type
- Identified which commands should be GlazeCommand vs BareCommand
- Planned layer composition for each command

**What I learned:**

### Command Categories

**Structured Output Candidates (GlazeCommand):**
- `filter list` - Could output filter list as rows
- `filter show` - Could output filtered files as rows
- `session show` - Already has `--yaml` flag, could be structured
- `generate` - Could output generation metadata as structured data

**Text Output Only (BareCommand or Dual):**
- `filter add` - Success message + impact summary
- `filter remove` - Success message
- `filter clear` - Success message
- `filter test` - Test results (could be structured)
- `session init` - Initialization summary
- `session load` - Load summary
- `session save` - Save confirmation
- `file toggle` - Toggle confirmation
- `context add` - Add confirmation
- `tui` - Interactive (no output)

**Dual Command Candidates:**
- `filter list` - Text table OR structured rows
- `filter show` - Text list OR structured rows
- `session show` - Text summary OR structured data
- `filter test` - Text results OR structured rows

**Key Decision Points:**
- Commands that modify state might stay as BareCommand (simpler)
- Commands that query/display data should be GlazeCommand or Dual
- TUI command stays as BareCommand (interactive)
- Generate command could be Dual (text output + metadata)

## Summary

**Key Findings:**
1. Glazed provides structured output without format-specific code
2. Layers enable reusable configuration across commands
3. Prescribe commands follow consistent controller pattern
4. Many commands could benefit from structured output
5. Repository and session configuration are good layer candidates

**Next Steps:**
- Create detailed command mapping document
- Design layer structure
- Plan migration strategy
- Identify which commands to port first

## Step 7: Initialize Prescribe with Glazed logging + help system

This step wires Prescribe into the “Glazed program initialization” pattern: logging flags are registered on the root command and logging is initialized early (via `PersistentPreRunE`) so every command run has consistent logging. In the same step, we boot a Glazed help system and load Prescribe’s own markdown help topics into it, so `prescribe help ...` can surface richer docs than plain Cobra help.

This is intentionally “groundwork only”: it should not change existing command semantics, but it unlocks follow-up work where we can port individual subcommands to Glazed (dual-mode or full GlazeCommand) while keeping `prescribe` feeling like a single cohesive CLI.

**Commit (code):** 90d79514c295a366d53a3c035d6a3356f5777c23 — "prescribe: init glazed help + logging"

### What I did
- Added `PersistentPreRunE` on `cmd/prescribe/cmds/root.go` to call `logging.InitLoggerFromCobra(cmd)`.
- Updated `cmd/prescribe/main.go` to:
  - call `logging.AddLoggingLayerToRootCommand(rootCmd, "prescribe")`
  - create a `help.NewHelpSystem()` and call `help_cmd.SetupCobraRootCommand(...)`
  - load Prescribe’s embedded help topics via `helpSystem.LoadSectionsFromFS(...)`
- Added a small `pkg/doc` Go package to embed `pkg/doc/topics/*.md` for the help system.

### Why
- We want Prescribe to behave like a Glazed application so future command ports can reuse Glazed layers, structured output, and help pages.
- Initializing logging in `PersistentPreRunE` ensures subcommands and middlewares can emit logs consistently (and early).
- Loading markdown help topics via Glazed’s help system gives us a scalable path for richer CLI docs as we add Glazed-based commands.

### What worked
- `go test ./... -count=1` passed in the `prescribe/` module after the changes.

### What didn't work
- I initially tried to run `git` from the workspace root (`/home/manuel/workspaces/2025-12-26/prescribe-import`) and hit “not a git repository”; `prescribe/` is a git worktree with its gitdir outside the workspace, so commits need to be done from the `prescribe/` directory.

### What I learned
- The Glazed help system can be incrementally adopted: it works fine with a small embedded doc set, and `LoadSectionsFromFS` is resilient (warns but doesn’t hard-fail when dirs don’t exist).

### What was tricky to build
- Keeping the initialization small and non-invasive: Glazed layers/help are added without restructuring existing Cobra command implementations yet.

### What warrants a second pair of eyes
- Confirm the help system integration doesn’t interfere with Cobra help output in surprising ways for existing commands (especially piping/help text destination).

### What should be done in the future
- Add more Prescribe help topics and link them to commands/flags as we port commands to Glazed.
- Decide whether we want to configure the help writer (stdout vs stderr) explicitly for Prescribe (see `glazed/pkg/doc/topics/01-help-system.md`).

### Code review instructions
- Start with `cmd/prescribe/main.go` and verify the initialization order (logging, help system, doc loading).
- Then review `cmd/prescribe/cmds/root.go` for the `PersistentPreRunE` addition.
- Run:
  - `cd prescribe && go test ./... -count=1`
  - `cd prescribe && go run ./cmd/prescribe --help` and `cd prescribe && go run ./cmd/prescribe help prescribe-filters-and-glob-syntax`

### Technical details
- The embedded docs live in `pkg/doc/topics/*.md` and are loaded with:
  - `helpSystem.LoadSectionsFromFS(prescribe_doc.FS, "topics")`

## Step 8: Add Phase 1 Glazed parameter layers (repository/session/filter/generation)

This step extracts Prescribe’s most common “global config” and shared flag groups into Glazed-style parameter sections. The immediate benefit is that future Glazed commands can compose these layers consistently without duplicating flag definitions in every command.

I intentionally kept this to “schema definitions + settings extraction helpers” only—no behavioral changes and no command ports yet. That keeps review small and makes the next step (controller init from parsed layers, then dual-mode command ports) much easier.

**Commit (code):** cb59b50746bb2f02a4203d7baf0bd755110d8f58 — "prescribe: add glazed parameter layers"

### What I did
- Added `prescribe/pkg/layers/` with four sections implemented using `schema.NewSection()` + `fields.New()`:
  - `repository` (`repo`, `target`)
  - `session` (`session-path`, `auto-save`)
  - `filter` (`filter-name`, `filter-description`, `exclude-patterns`, `include-patterns`)
  - `generation` (`prompt`, `preset`, `load-session`, `output-file`)
- Added `Get…Settings(parsedLayers)` helpers for each layer, returning typed settings structs.

### Why
- These flags are reused across many commands; layers make the reuse explicit and keep help/flag definitions consistent.
- Typed settings helpers reduce the amount of “flag plumbing” in command implementations and make it easier to test parsing behavior.

### What worked
- `go test ./... -count=1` passed in the `prescribe/` module after adding the new package.

### What didn't work
- N/A.

### What I learned
- Using the `schema`/`fields` aliases keeps new Glazed code in Prescribe readable and aligned with the newer Glazed docs and examples.

### What was tricky to build
- Keeping flag names aligned with existing Cobra flags (`repo`, `target`, etc.) so we don’t accidentally introduce breaking changes when we start using these layers in commands.

### What warrants a second pair of eyes
- Confirm the chosen layer slugs and field names match what we’ll want long-term (especially `session-path` and `auto-save`) before many commands start depending on them.

### What should be done in the future
- Add small unit tests around the layer definitions + defaults (to lock down flag names, defaults, and short flags).
- Add controller initialization helpers that read these settings from `*layers.ParsedLayers` (instead of Cobra flags).

### Code review instructions
- Start in `prescribe/pkg/layers/` and verify:
  - slugs (`RepositorySlug`, etc.)
  - field names and short flags
  - defaults
  - helper error wrapping and nil handling
- Run:
  - `cd prescribe && go test ./... -count=1`

## Step 9: Add controller initialization helper that consumes Glazed parsed layers

This step creates a small bridge between “Glazed world” and Prescribe’s existing controller architecture: Glazed commands don’t want to read Cobra flags directly, so we need a helper that takes `*layers.ParsedLayers`, extracts the repository settings, and initializes a `*controller.Controller`.

This unlocks the next Phase 2 work (dual-mode commands). We can port a command to Glazed output without rewriting the controller layer or duplicating repo/target parsing logic.

**Commit (code):** 8e294d1b599458921d37cfbf7a8648a09638e6e5 — "prescribe: init controller from parsed layers"

### What I did
- Added `cmd/prescribe/cmds/helpers/controller_from_layers.go` with:
  - `NewInitializedControllerFromParsedLayers(parsedLayers *layers.ParsedLayers)`
  - It calls `prescribe/pkg/layers.GetRepositorySettings` and then `controller.NewController(...)` + `ctrl.Initialize(...)`.

### Why
- Glazed commands are built around `*layers.ParsedLayers`, so controller initialization must be possible without reading Cobra flags.
- Centralizing the initialization keeps future command ports consistent and reduces copy/paste.

### What worked
- `go test ./... -count=1` passed in the `prescribe/` module.

### What didn't work
- N/A.

### What I learned
- The “minimal bridge” approach keeps the migration incremental: we can keep Prescribe’s controller stable while we modernize the CLI surface area.

### What was tricky to build
- Avoiding package name collisions (`glazed/pkg/cmds/layers` vs `prescribe/pkg/layers`) while still keeping the helper readable.

### What warrants a second pair of eyes
- Confirm error wrapping is consistent with the rest of Prescribe/Glazed and that no important context gets lost.

### What should be done in the future
- Add unit tests around `NewInitializedControllerFromParsedLayers` using a `layers.ParsedLayers` with repository settings filled in (to lock down expected parsing).

### Code review instructions
- Start in `cmd/prescribe/cmds/helpers/controller_from_layers.go`.
- Verify imports/aliases and error handling.
- Run:
  - `cd prescribe && go test ./... -count=1`

## Step 10: Remove init()-based Cobra wiring and port `filter list` to dual-mode Glazed output

This step addresses a real sharp edge we hit: relying on `init()` ordering in Cobra command packages caused a runtime panic when `filter/filter.go` tried to register `ListFiltersCmd` before it had been constructed. To make the migration safe and predictable, I refactored Prescribe’s command wiring to use explicit initialization functions instead of package-level `init()` hooks.

With deterministic initialization in place, I also completed the first Phase 2 port: `prescribe filter list` now supports **dual-mode** execution (classic text output by default, and Glazed structured output when `--with-glaze-output` is provided).

**Commit (code):** da425db88a8e1c0d10eaa9edd4fb0965bfc38924 — "prescribe: explicit cobra init + dual-mode filter list"

### What I did
- Removed all `init()` functions under `cmd/prescribe/cmds/**` and replaced them with explicit `Init…()` functions:
  - `cmds.InitRootCmd(rootCmd)` wires global flags and registers command trees
  - `filter.Init()`, `session.Init()`, `context.Init()`, `file.Init()` set up their subcommands and flag bindings deterministically
- Refactored `cmd/prescribe/main.go` to:
  - build `rootCmd := cmds.NewRootCmd()`
  - initialize Glazed logging/help as before
  - call `cmds.InitRootCmd(rootCmd)` before execution
- Ported `filter list` to Glazed dual-mode:
  - added a Glazed command implementation (`FilterListCommand` implementing `cmds.GlazeCommand`)
  - wired classic output through `cli.BuildCobraCommandFromCommandAndFunc(..., cli.WithDualMode(true))`
  - switched both modes to use `helpers.NewInitializedControllerFromParsedLayers`
- Added `prescribe/pkg/layers/existing_cobra_flags_layer.go` to avoid “flag redefined” errors when a layer’s flags already exist as inherited persistent flags on the root command (eg. `--repo`, `--target`).

### Why
- `init()` ordering across files is not deterministic; explicit initialization eliminates a whole class of startup/runtime panics.
- Dual-mode lets us add structured output without breaking existing text-based usage.
- The “existing flags layer” wrapper allows us to reuse Glazed parsing without duplicating root-level persistent flags.

### What worked
- `go test ./... -count=1` still passes.
- `go run ./cmd/prescribe filter list --help` no longer panics.

### What didn't work
- Before the refactor, `go run ./cmd/prescribe filter list --help` panicked due to a nil subcommand being added during init-time registration.

### What I learned
- A safe migration path for Cobra→Glazed in an existing app is: **keep Cobra as the container**, but make all Glazed command construction happen explicitly (no init ordering), and treat root persistent flags as “already defined” when building Glazed layers.

### What was tricky to build
- Avoiding “flag redefined” errors while still allowing Glazed to parse inherited persistent flags (`repo`, `target`) for Glazed commands.

### What warrants a second pair of eyes
- Review the `ExistingCobraFlagsLayer` wrapper carefully: it intentionally suppresses flag registration but still parses from Cobra, so we need to ensure this doesn’t hide flags unexpectedly or lead to confusing help output in other commands.

### What should be done in the future
- Consider adding a small unit test for the “existing flags layer” wrapper and for `filter list` glaze output shape (column names + nil handling).
- Decide whether to move `repo/target` entirely into a Glazed layer on the root command (and remove the current Cobra persistent flags), or keep the current mixed approach.

### Code review instructions
- Start in:
  - `cmd/prescribe/cmds/root.go` (`NewRootCmd`, `InitRootCmd`)
  - `cmd/prescribe/main.go` (initialization ordering)
  - `cmd/prescribe/cmds/filter/list.go` (dual-mode command wiring)
  - `pkg/layers/existing_cobra_flags_layer.go` (wrapper semantics)
- Validate:
  - `cd prescribe && go test ./... -count=1`
  - `cd prescribe && go run ./cmd/prescribe filter list --help`
  - `cd prescribe && go run ./cmd/prescribe filter list --with-glaze-output --output json`

## Step 11: Port `filter show` to dual-mode Glazed output

This step ports `prescribe filter show` to the same dual-mode pattern as `filter list`: classic text output remains the default, but users can request Glazed structured output with `--with-glaze-output`. This keeps backwards compatibility while enabling JSON/YAML/CSV export of the filtered file list.

The command is still implemented on top of the existing controller and session model. The port is mostly “plumbing”: constructing a Glazed command description, parsing repository settings from inherited root flags, and emitting a row per filtered file.

**Commit (code):** 3f05fca7ea9a7019664590be0abf48a952bec7f2 — "prescribe: dual-mode filter show"

### What I did
- Converted `filter/show.go` to build `ShowFilteredCmd` via Glazed’s Cobra builder (dual-mode).
- Added `InitShowFilteredCmd()` and wired it into `filter.Init()` so command registration is explicit and deterministic.
- Implemented Glaze output as one row per filtered file with:
  - `file_path`, `additions`, `deletions`, `tokens`
  - plus summary counts (`total_files`, `visible_files`, `filtered_files`) for convenience.

### Why
- `filter show` is naturally “query output”, which benefits from structured output formats.
- Keeping classic output as default avoids breaking existing usage and keeps the UX for interactive terminal usage unchanged.

### What worked
- `go test ./... -count=1` passed.
- Smoke test:
  - `cd prescribe && go run ./cmd/prescribe filter show --help`
  - `cd prescribe && go run ./cmd/prescribe filter show --with-glaze-output --output json`

### What didn't work
- N/A.

### What I learned
- Once the “existing persistent flags” wrapper is in place, porting additional query commands becomes repetitive in a good way: build command description + output rows + keep a classic run function.

### What was tricky to build
- Ensuring `filter.Init()` initializes `ShowFilteredCmd` before registering it, so we never reintroduce init-order panics.

### What warrants a second pair of eyes
- Confirm the chosen Glaze row schema (especially the inclusion of summary counts on every row) is what we want long-term vs emitting a separate “summary row”.

### What should be done in the future
- Consider a small shared helper for “emit file rows + counts” to keep future ports (`filter show`, `filter test`, maybe `tui export`) consistent.

### Code review instructions
- Start in `cmd/prescribe/cmds/filter/show.go`.
- Validate with:
  - `cd prescribe && go run ./cmd/prescribe filter show --help`
  - `cd prescribe && go run ./cmd/prescribe filter show --with-glaze-output --output json`

## Step 12: Port `filter test` to dual-mode Glazed output (keep legacy flags)

This step ports `prescribe filter test` to dual-mode structured output while keeping the existing CLI surface. The command still accepts `--name/-n`, `--exclude/-e`, and `--include/-i` exactly as before, and classic output remains the default behavior.

In Glaze mode (`--with-glaze-output`), the command emits one row per file with a boolean indicating whether the file would be visible (matched) or filtered out, plus some convenience counts to make it easy to summarize results in downstream tooling.

**Commit (code):** 3a028b58a1f964045128f2b660a7aa0e96c1ba50 — "prescribe: dual-mode filter test"

### What I did
- Reworked `cmd/prescribe/cmds/filter/test.go` to build `TestFilterCmd` via the Glazed dual-mode Cobra builder.
- Added a dedicated layer (`filter-test`) that defines the legacy flags:
  - `--name/-n` (default `"test"`)
  - `--exclude/-e` (string list)
  - `--include/-i` (string list)
- Implemented Glaze mode output as rows with:
  - `filter_name`, `file_path`, `matched` (bool)
  - `total_files`, `matched_files`, `filtered_files`

### Why
- `filter test` is frequently used as “quick feedback”; structured output makes it scriptable (JSON/YAML/CSV) without rewriting formatting logic.
- Keeping legacy flags avoids breaking existing usage while we migrate internals.

### What worked
- `go test ./... -count=1` passed.
- Smoke test:
  - `cd prescribe && go run ./cmd/prescribe filter test --exclude '**/*.md'`
  - `cd prescribe && go run ./cmd/prescribe filter test --exclude '**/*.md' --with-glaze-output --output json`

### What didn't work
- N/A (the `head` pipe during smoke testing can trigger a broken-pipe exit code; the command output itself is correct).

### What I learned
- For dual-mode ports that must keep legacy flags, the simplest approach is often a small command-specific layer (rather than forcing reuse of a shared “FilterLayer” with different names).

### What was tricky to build
- Keeping the classic output identical while switching the parsing source to `*layers.ParsedLayers` (no direct access to `*cobra.Command` in the classic run function).

### What warrants a second pair of eyes
- Validate the row schema choices (`matched` vs `filtered` naming) and whether counts should be repeated per-row or emitted separately.

### What should be done in the future
- Consider extracting the “rules builder” into a shared helper once we port more filter commands to reduce duplication.

### Code review instructions
- Start in `cmd/prescribe/cmds/filter/test.go` and review:
  - legacy flag mapping in the `filter-test` layer
  - row schema in `RunIntoGlazeProcessor`
- Validate with:
  - `cd prescribe && go run ./cmd/prescribe filter test --exclude '**/*.md' --with-glaze-output --output json`

## Step 13: Port `session show` to dual-mode Glazed output (preserve `--yaml`)

This step ports `prescribe session show` to dual-mode. The classic output remains the default, and `--yaml/-y` still produces the same YAML serialization as before. When `--with-glaze-output` is requested, the command instead emits a single structured “session summary” row (which can be formatted as JSON/YAML/CSV via Glazed output flags).

To avoid ambiguous combinations, `--yaml` is treated as **classic-only**: using it together with `--with-glaze-output` returns a clear error telling users to use `--output yaml` instead.

**Commit (code):** 425af79cd76e4642dd3900b3279ad9bd2dc2b0c2 — "prescribe: dual-mode session show"

### What I did
- Reworked `cmd/prescribe/cmds/session/show.go` to build `ShowCmd` via the Glazed dual-mode Cobra builder.
- Added a small `session-show` layer carrying the legacy `--yaml/-y` flag so it stays available for classic mode.
- Implemented Glaze mode output as a single row containing:
  - branch info, file counts, filter/context counts, token count
  - preset metadata (if set) or a prompt preview (if no preset)
- Added an explicit error when `--yaml` is combined with `--with-glaze-output`.

### Why
- `session show` is a “query command”; structured output makes it easy to introspect and automate.
- Preserving `--yaml` avoids breaking existing usage while we migrate the command plumbing.

### What worked
- `go test ./... -count=1` passed.
- Smoke tests:
  - `cd prescribe && go run ./cmd/prescribe session show --yaml`
  - `cd prescribe && go run ./cmd/prescribe session show --with-glaze-output --output json`
  - `cd prescribe && go run ./cmd/prescribe session show --with-glaze-output --yaml --output json` (expected error)

### What didn't work
- N/A.

### What I learned
- Treating “classic-only” flags as a small command-specific layer is a clean way to preserve compatibility even when the Cobra command itself is built by Glazed.

### What was tricky to build
- Choosing a stable row schema for a “session summary” without exploding into many row types; keeping it to one row keeps it predictable for scripting.

### What warrants a second pair of eyes
- Confirm the chosen row fields are the right long-term contract for structured output (especially prompt/preset fields).

### What should be done in the future
- If we later add a structured “detailed mode” (files, filters, context items), decide whether to expose it as separate subcommands or additional flags (but keep row schema stable).

### Code review instructions
- Start in `cmd/prescribe/cmds/session/show.go` and review:
  - `--yaml` classic-only enforcement
  - row schema
- Validate with:
  - `cd prescribe && go run ./cmd/prescribe session show --with-glaze-output --output json`

## Step 14: Cleanup — remove a no-op init() to keep initialization explicit

This step removes an empty `init()` function that lived in `prescribe/pkg/doc.go`. It didn’t do anything, but given our migration goal (predictable, explicit initialization rather than implicit side effects), removing no-op inits helps keep the codebase consistent and avoids future “why is this here?” questions.

**Commit (code):** fc233a3718eca45b09e5d51f9af1fa0bc9a8e6b0 — "prescribe: remove no-op init"

### What I did
- Deleted `prescribe/pkg/doc.go`, which contained an empty `init()` function.

### Why
- Align the codebase with the explicit initialization approach (and avoid implicit side-effect hooks that can later become ordering footguns).

### What worked
- `go test ./... -count=1` still passes after the deletion.

### What didn't work
- N/A.

### What I learned
- N/A.

### What was tricky to build
- N/A.

### What warrants a second pair of eyes
- N/A (mechanical deletion).

### What should be done in the future
- N/A.

### Code review instructions
- Verify the deletion is safe and unused:
  - `cd prescribe && go test ./... -count=1`

## Step 15: Port `filter add` to a Glazed-built BareCommand (preserve legacy flags)

This step ports `prescribe filter add` to the Glazed command builder, while keeping the existing CLI flags intact (`--name/-n`, `--description/-d`, `--exclude/-e`, `--include/-i`). The command remains “classic output only” (BareCommand): it prints confirmation text and saves the session, just like before.

The main change is that the command no longer reads Cobra flags directly; it uses `*layers.ParsedLayers` and the shared `NewInitializedControllerFromParsedLayers` helper. This keeps the migration consistent with the other command ports and reduces future flag-plumbing.

**Commit (code):** 451d28b025f658c5b38453e4548173098ecb559a — "prescribe: glazed barecommand filter add"

### What I did
- Reworked `cmd/prescribe/cmds/filter/add.go` to:
  - build `AddFilterCmd` via `cli.BuildCobraCommand` from a `cmds.BareCommand`
  - define a dedicated layer (`filter-add`) that exposes the legacy flags and requiredness
  - use repository settings from inherited root flags (`--repo`, `--target`) via the existing-flags wrapper
  - initialize controller via `helpers.NewInitializedControllerFromParsedLayers`

### Why
- Make `filter add` consistent with the “Glazed parsing + explicit init” architecture.
- Keep backwards compatibility at the CLI surface while migrating internals.

### What worked
- `go test ./... -count=1` passed.
- Manual smoke test:
  - `cd prescribe && go run ./cmd/prescribe filter add --name \"Exclude docs\" --exclude \"**/*.md\"`

### What didn't work
- N/A.

### What I learned
- BareCommands are a good fit for state-modifying commands: we still benefit from Glazed layer parsing without forcing structured output.

### What was tricky to build
- Ensuring we don’t re-add `--repo/--target` flags (they’re already persistent on the root) while still allowing Glazed parsing for those values.

### What warrants a second pair of eyes
- Confirm the `filter-add` layer schema is a good long-term contract (especially requiredness of `--name`) and that error messages remain user-friendly vs the old Cobra-required behavior.

### What should be done in the future
- Consider extracting a shared “legacy filter flags layer” for `filter add` + `filter test` to reduce duplication, if we find ourselves repeating these definitions.

### Code review instructions
- Start in `cmd/prescribe/cmds/filter/add.go`.
- Validate:
  - `cd prescribe && go run ./cmd/prescribe filter add --help`
  - `cd prescribe && go run ./cmd/prescribe filter add --name test --exclude '**/*.md'`

## Step 16: Simplify — drop backwards compatibility and remove dual-mode toggle plumbing

This step intentionally breaks backwards compatibility to simplify the migration: instead of maintaining “classic output vs Glazed output” dual-mode and a `--with-glaze-output` toggle, the ported query commands now **always run as Glazed commands**. This significantly reduces code and removes a whole class of branching and subtle behavior differences.

It also means we stop carrying “classic-only” flags like `session show --yaml`. If users want YAML now, they should use Glazed’s formatter flags: `--output yaml`.

**Commit (code):** 9860e32be94852851326d48e42a35936ced75c3d — "prescribe: drop dual-mode / compatibility glue"

### What I did
- Removed dual-mode Cobra builders and classic-mode run functions from:
  - `filter list`
  - `filter show`
  - `filter test`
  - `session show`
- Removed the `--with-glaze-output` toggle flag entirely (ported commands always expose Glazed output flags directly).
- Simplified `session show` by removing the `--yaml` flag and classic output implementation.

### Why
- You explicitly don’t want backwards compatibility; dual-mode was pure complexity.
- Single-mode Glazed commands are easier to reason about and standardize (output formatting via Glazed, no custom printing).

### What worked
- `go test ./... -count=1` still passes.
- `prescribe filter list --output json` works without requiring a mode toggle.

### What didn't work
- When piping to `head` during smoke tests, Glazed JSON output can end with “broken pipe”; this is expected when the consumer closes early.

### What I learned
- Once you commit to “Glazed-first”, the command implementations become dramatically smaller and more uniform.

### What was tricky to build
- Deciding what to do with classic-only flags (`--yaml`). The cleanest approach is to remove them and rely on Glazed format flags.

### What warrants a second pair of eyes
- Review the UX change: help output no longer shows a “classic mode”. Confirm this aligns with intended usage and docs.

### What should be done in the future
- Update any scripts and docs that referenced `--with-glaze-output` or `session show --yaml`.

### Code review instructions
- Start with:
  - `cmd/prescribe/cmds/filter/list.go`
  - `cmd/prescribe/cmds/session/show.go`
- Validate:
  - `cd prescribe && go run ./cmd/prescribe filter list --output json`
  - `cd prescribe && go run ./cmd/prescribe session show --output yaml`

## Step 17: Create onboarding playbook for porting Cobra verbs to Glazed

This step creates a single “how to port commands” playbook meant for a new developer joining the project with no prior Glazed context. It captures the repo-specific conventions we learned during the ticket: explicit command initialization (no `init()`), how to think about BareCommand vs GlazeCommand, how to wire schema layers, and common pitfalls like re-adding root persistent flags.

The goal is to make future ports consistent and fast, and to reduce the need to reverse-engineer patterns from previous commits.

### What I did
- Added a playbook document:
  - `playbook/01-playbook-port-existing-cobra-verbs-to-glazed-no-back-compat.md`
- Included:
  - environment assumptions + smoke test commands
  - copy/paste porting recipe (schema → implementation → InitXxxCmd → group Init → tests → commits)
  - pitfalls / gotchas (init ordering, root flags, broken pipe during head)
- Related key implementation files directly to the playbook with `docmgr doc relate`.

### Why
- We want the next developer to be productive immediately without reading the entire ticket diary or Glazed docs.
- Codifying patterns also prevents drift (especially important now that we explicitly do not preserve backwards compatibility).

### What warrants a second pair of eyes
- Ensure the playbook aligns with the current “no back-compat” direction and doesn’t suggest dual-mode patterns.

## Step 18: Port `filter remove` and `filter clear` to Glazed BareCommands (positional args supported)

This step finishes the `filter` command family migration by porting the remaining “session mutators” (`remove`, `clear`) from plain Cobra to Glazed-built `BareCommand`s. The primary goal was consistency: all filter subcommands now follow the same Glazed plumbing (ParsedLayers + controller init from layers + explicit Init wiring) instead of mixing parsing styles.

The only potentially tricky aspect was `filter remove`’s positional argument (`<index|name>`). We verified Glazed positional argument support works here via a `schema.DefaultSlug` section using `schema.WithArguments(...)`, so we didn’t need to fall back to a flag.

**Commit (code):** be520636d10f84c77fcfef7bc240f06889fa88d2 — "prescribe: port filter remove/clear to glazed"

### What I did
- Converted `cmd/prescribe/cmds/filter/remove.go` from a `*cobra.Command` to a Glazed `cmds.BareCommand` built with `cli.BuildCobraCommand(...)`.
- Implemented the positional `<index-or-name>` argument using:
  - `schema.NewSection(schema.DefaultSlug, ..., schema.WithArguments(fields.New(...)))`
  - `parsedLayers.InitializeStruct(schema.DefaultSlug, &settings)`
- Converted `cmd/prescribe/cmds/filter/clear.go` to a Glazed `cmds.BareCommand`.
- Updated `cmd/prescribe/cmds/filter/filter.go` to explicitly initialize and register `remove` and `clear` via `InitRemoveFilterCmd()` / `InitClearFiltersCmd()`.

### Why
- Keep the CLI surface consistent: all filter subcommands now use the same Glazed parsing + controller initialization path.
- Avoid init-order footguns: explicit init wiring ensures the command tree is deterministic and safe.

### What worked
- `cd prescribe && go test ./... -count=1` passed after the port.
- Positional argument parsing for `filter remove` works as intended (no need to replace it with flags).

### What was tricky to build
- Ensuring `filter remove` kept its “index-or-name” selector behavior while moving from Cobra’s `Args` handling to Glazed argument decoding.
- Avoiding `--repo/--target` flag redefinition: still handled via `WrapAsExistingCobraFlagsLayer(...)`.

### What warrants a second pair of eyes
- Confirm the positional arg contract is acceptable long-term (`filter remove <index|name>` with **0-based** index).
- Validate error messages and UX remain clear when:
  - no session exists,
  - no filters exist,
  - name is not found,
  - index is out of range.

### What should be done in the future
- Add small integration tests around the command wiring + argument parsing (especially `filter remove`), so we don’t regress on positional args.

### Code review instructions
- Start in:
  - `cmd/prescribe/cmds/filter/remove.go`
  - `cmd/prescribe/cmds/filter/clear.go`
  - `cmd/prescribe/cmds/filter/filter.go`
- Validate with:
  - `cd prescribe && go run ./cmd/prescribe filter remove --help`
  - `cd prescribe && go run ./cmd/prescribe filter clear --help`

## Step 19: Port `session load` and `session save` to Glazed BareCommands (optional positional path)

This step ports the remaining “simple” session file operations (`load`, `save`) from plain Cobra commands to Glazed-built `BareCommand`s. This keeps the `session` command group consistent: Glazed parsing is now used everywhere, and these commands no longer read Cobra flags/args directly.

Both commands take an optional positional `[path]`. We verified Glazed supports this cleanly by defining a `schema.DefaultSlug` section with a single non-required argument field (`path`), which automatically generates the correct `Usage` (`[path]`) and Cobra arg validation.

**Commit (code):** a2e2bca985e03aa855afeed54ef27155d4ded227 — "prescribe: port session load/save to glazed"

### What I did
- Converted `cmd/prescribe/cmds/session/load.go` into a Glazed `cmds.BareCommand`:
  - argument parsing via `schema.WithArguments(fields.New("path", ...))`
  - controller init via `helpers.NewInitializedControllerFromParsedLayers`
- Converted `cmd/prescribe/cmds/session/save.go` into a Glazed `cmds.BareCommand`:
  - same optional `[path]` handling
  - still loads existing session if present before saving (`LoadDefaultSessionIfExists`)
- Updated `cmd/prescribe/cmds/session/session.go` to explicitly initialize and register both commands via `InitLoadCmd()` / `InitSaveCmd()`.

### Why
- Eliminate remaining direct Cobra arg parsing in the session group; standardize on Glazed ParsedLayers.
- Keep the migration pattern consistent across command families (filter/session).

### What worked
- `cd prescribe && go test ./... -count=1` passed after the port.
- Help output correctly shows:
  - `prescribe session load [path] [flags]`
  - `prescribe session save [path] [flags]`

### What was tricky to build
- Making sure “optional arg” maps to “empty string means default session path” without losing the old UX.
- Keeping the session-save behavior of loading an existing session first so save reflects current state.

### What warrants a second pair of eyes
- Confirm the optional-arg semantics match expectations:
  - empty `[path]` uses `ctrl.GetDefaultSessionPath()`
  - provided `[path]` is used verbatim

### What should be done in the future
- Add a lightweight command-level integration test that exercises `session save` and `session load` argument parsing (0 args / 1 arg) to avoid regressions.

### Code review instructions
- Start in:
  - `cmd/prescribe/cmds/session/load.go`
  - `cmd/prescribe/cmds/session/save.go`
  - `cmd/prescribe/cmds/session/session.go`
- Validate with:
  - `cd prescribe && go run ./cmd/prescribe session load --help`
  - `cd prescribe && go run ./cmd/prescribe session save --help`

## Step 20: Port `session init` to a Glazed BareCommand (keep `--save/--path`)

This step ports `prescribe session init` from a plain Cobra command (with package-level flag variables) to a Glazed-built `BareCommand`. The goal is to remove the last “Cobra-only” parsing from the `session` command family and keep all session subcommands using ParsedLayers + the shared controller init helper.

We kept the existing CLI flags (`--save` and `--path/-p`) for now to avoid unnecessary churn, even though the ticket doesn’t promise backwards compatibility. This keeps the UX familiar while still adopting the new parsing model.

**Commit (code):** b28b057cf6ebc7921011b063eedcca75894471bf — "prescribe: port session init to glazed"

### What I did
- Replaced the old `InitCmd = &cobra.Command{...}` implementation with:
  - `NewSessionInitCommand()` returning a Glazed `cmds.BareCommand`
  - a dedicated section (`session-init`) defining:
    - `--save` (bool)
    - `--path/-p` (string)
- Updated `InitInitCmd()` to build the Cobra command via `cli.BuildCobraCommand(...)`.
- Switched controller initialization to `helpers.NewInitializedControllerFromParsedLayers`.

### Why
- Eliminate global flag variables and direct Cobra flag reading.
- Standardize session command plumbing, matching the already-ported filter commands.

### What worked
- `cd prescribe && go test ./... -count=1` passed.
- `prescribe session init --help` shows the Glazed section with `--save` and `--path/-p`.

### What was tricky to build
- Ensuring the `--path/-p` flag remains a command flag (not a positional arg) and still defaults to `ctrl.GetDefaultSessionPath()` when empty.

### What warrants a second pair of eyes
- Confirm the flag naming should remain `--save/--path` long-term vs adopting the shared SessionLayer (`session-path`, `auto-save`) for consistency across future ports.

### What should be done in the future
- If we decide to standardize on SessionLayer, do it in one intentional breaking change across all relevant commands rather than drifting per-command.

### Code review instructions
- Start in `cmd/prescribe/cmds/session/init.go`.
- Validate with:
  - `cd prescribe && go run ./cmd/prescribe session init --help`

## Step 21: Port `file toggle` to a Glazed BareCommand (positional `<path>`)

This step ports `prescribe file toggle` from plain Cobra to a Glazed-built `BareCommand`. Like the other “session mutators”, it remains classic text output, but it now uses ParsedLayers for consistent parsing and controller initialization.

We kept the positional `<path>` argument (required) and implemented it via a `schema.DefaultSlug` section with `schema.WithArguments(...)`, which means usage/help and Cobra arg validation remain correct without custom `cobra.Args` logic.

**Commit (code):** 8909c028a1f4ce2a034895acd4db2be111fffa14 — "prescribe: port file toggle to glazed"

### What I did
- Converted `cmd/prescribe/cmds/file/toggle.go` to a Glazed `cmds.BareCommand` built with `cli.BuildCobraCommand(...)`.
- Implemented the required positional `<path>` argument as a default section argument:
  - `schema.NewSection(schema.DefaultSlug, ..., schema.WithArguments(fields.New("path", ...)))`
- Updated `cmd/prescribe/cmds/file/file.go` to explicitly initialize the command via `InitToggleFileCmd()` (consistent with the explicit init pattern used in `filter` and `session`).

### Why
- Remove another Cobra-only parsing island; standardize all ported commands on ParsedLayers and the controller-from-layers helper.
- Keep initialization deterministic and avoid future init-order surprises.

### What worked
- `cd prescribe && go test ./... -count=1` passed.
- `prescribe file toggle --help` shows:
  - `prescribe file toggle <path> [flags]`

### What was tricky to build
- Ensuring the “new state” printed after toggling reflects the updated controller data (we re-read `ctrl.GetData()` after toggling).

### What warrants a second pair of eyes
- Confirm the path matching logic is acceptable (exact match against `data.ChangedFiles[i].Path`).
- Confirm error messages remain user-friendly when the path is not found.

### What should be done in the future
- Consider supporting a more forgiving path match (or listing candidates) if users often pass slightly different relative paths.

### Code review instructions
- Start in:
  - `cmd/prescribe/cmds/file/toggle.go`
  - `cmd/prescribe/cmds/file/file.go`
- Validate with:
  - `cd prescribe && go run ./cmd/prescribe file toggle --help`

## Step 22: Port `context add` to a Glazed BareCommand (flag `--note` + optional positional `[file-path]`)

This step ports `prescribe context add` from plain Cobra to a Glazed-built `BareCommand`. This command is slightly special because it has a mutually-exclusive input model: either you add a note (`--note`) or you add a context file (positional `[file-path]`). The port keeps that behavior exactly, but moves parsing and validation into ParsedLayers-based settings initialization.

We modeled the two inputs as:
- a command section (`context-add`) containing the `--note` flag, and
- a default section (`schema.DefaultSlug`) containing an optional positional `file-path` argument.

**Commit (code):** 519a9593a87a112b1dc5b718990234776af64372 — "prescribe: port context add to glazed"

### What I did
- Converted `cmd/prescribe/cmds/context/add.go` into a Glazed `cmds.BareCommand`.
- Implemented mutually-exclusive input validation after decoding:
  - error if both `--note` and `[file-path]` are empty
  - error if both are provided
- Kept behavior the same:
  - loads default session if it exists
  - adds note or file to context
  - saves session and prints token count

### Why
- Remove the last Cobra-only parsing in the `context` command group.
- Keep command plumbing consistent across the app (ParsedLayers + controller init from layers).

### What worked
- `cd prescribe && go test ./... -count=1` passed.
- `prescribe context add --help` shows correct usage and flags.

### What was tricky to build
- Mapping a mutually-exclusive “either arg or flag” contract cleanly into Glazed’s section model (default section for args + a dedicated section for flags).

### What warrants a second pair of eyes
- Confirm the UX for invalid combinations is clear enough (especially the error text).

### What should be done in the future
- Consider adding an integration test for the mutual-exclusion behavior so we don’t regress on the CLI contract.

### Code review instructions
- Start in `cmd/prescribe/cmds/context/add.go`.
- Validate with:
  - `cd prescribe && go run ./cmd/prescribe context add --help`

## Step 23: Port `generate` to a Glazed BareCommand (use GenerationLayer)

This step ports the root-level `prescribe generate` command from plain Cobra to a Glazed-built `BareCommand`. The goal is to eliminate another Cobra-only parsing path and to actually start using the already-created `GenerationLayer` (prompt/preset/load-session/output-file).

One intentional behavior tweak: we now load the **default session if it exists** before generating, so `generate` reflects the current session state created/modified by other commands (filters, file toggles, context items). If `--load-session` is provided, it overrides the default session.

**Commit (code):** 2184c6237fe68f647f4d2a78f62407e1867b9a1d — "prescribe: port generate to glazed"

### What I did
- Converted `cmd/prescribe/cmds/generate.go` into a Glazed `cmds.BareCommand` built with `cli.BuildCobraCommand(...)`.
- Replaced legacy Cobra flags with the shared `GenerationLayer` fields:
  - `--prompt/-p`
  - `--preset`
  - `--load-session/-s`
  - `--output-file/-o`
- Switched controller initialization to `helpers.NewInitializedControllerFromParsedLayers`.
- Added `helpers.LoadDefaultSessionIfExists(ctrl)` before generation (then override with `--load-session` if set).

### Why
- Standardize parsing and wiring for root-level commands as well (not just subcommand groups).
- Reuse the shared generation schema definitions so future tests/docs can treat these flags as a stable contract.

### What worked
- `cd prescribe && go test ./... -count=1` passed.
- `prescribe generate --help` shows the Generation layer flags.

### What was tricky to build
- Choosing whether to preserve the legacy `--session` flag name vs adopting the `GenerationLayer` field name `--load-session`. We went with `GenerationLayer` for consistency.

### What warrants a second pair of eyes
- Review the behavior change (default session is now loaded if present) to confirm it matches intended UX.

### What should be done in the future
- Consider adding a short doc section noting the flag rename (`--session` → `--load-session`) since we’re intentionally not preserving backwards compatibility.

### Code review instructions
- Start in `cmd/prescribe/cmds/generate.go`.
- Validate with:
  - `cd prescribe && go run ./cmd/prescribe generate --help`

## Step 24: Port `tui` to a Glazed BareCommand and wire it via explicit init

This step ports the root-level `prescribe tui` command to a Glazed-built `BareCommand`. The big goal here is consistency: even interactive commands should initialize the controller from ParsedLayers (instead of reading Cobra flags directly) so the overall CLI has a single, predictable parsing path.

Because `tuiCmd` is registered at the root level (not inside a command group `Init()`), we also added an explicit `InitTuiCmd()` call in `cmds.InitRootCmd` to ensure the Cobra command is built deterministically before it’s added to the root command.

**Commit (code):** b0e58f853c1d3fa75f8107c3b5edd0fd3f985b47 — "prescribe: port tui to glazed"

### What I did
- Converted `cmd/prescribe/cmds/tui.go` into a Glazed `cmds.BareCommand` built with `cli.BuildCobraCommand(...)`.
- Switched controller initialization to `helpers.NewInitializedControllerFromParsedLayers`.
- Kept existing behavior of loading the default session if it exists before launching the UI.
- Updated `cmd/prescribe/cmds/root.go` to call `InitTuiCmd()` during `InitRootCmd`.

### Why
- Reduce “special casing”: root-level commands should follow the same Glazed initialization pattern as subcommands.
- Avoid accidental nil command registration by ensuring `tuiCmd` is built before it’s added to the root command.

### What worked
- `cd prescribe && go test ./... -count=1` passed.
- `prescribe tui --help` works and no longer depends on a pre-constructed Cobra command literal.

### What was tricky to build
- Remembering that `tuiCmd` is not initialized anywhere unless we explicitly call `InitTuiCmd()` in `InitRootCmd`.

### What warrants a second pair of eyes
- Verify we didn’t introduce any UX regressions for TUI startup (flags, session loading, alt-screen behavior).

### What should be done in the future
- Consider whether we want to attach additional layers (session/filter) to TUI to support config-driven initialization (not necessary for the migration, but now possible).

### Code review instructions
- Start in:
  - `cmd/prescribe/cmds/tui.go`
  - `cmd/prescribe/cmds/root.go`
- Validate with:
  - `cd prescribe && go run ./cmd/prescribe tui --help`

## Step 25: Update smoke scripts + docs for the new Glazed CLI surface (no back-compat)

This step cleans up the practical fallout from intentionally dropping backwards compatibility: we had multiple smoke scripts and docs still using the old CLI surface (`show --yaml`, `generate --session`, and flat verbs like `add-filter`/`toggle-file`). Those would now fail and confuse future devs.

We updated the tracked smoke scripts under `prescribe/test/` to use the new grouped command tree (`session …`, `filter …`, `file …`, `context …`) and the new Glazed output conventions (eg. `session show --output yaml`). We also made the scripts more portable by building a local `prescribe` binary in `/tmp` instead of pointing at a hardcoded `/home/ubuntu/...` path.

Finally, we updated the top-level docs (`README.md`, `PROJECT-SUMMARY.md`, and the Bubbletea TUI playbook) so examples match the new CLI.

**Commit (code/docs):** fe689b50bdec6c5590cc4281d9c23587e17da864 — "prescribe: update smoke scripts + docs for glazed cli"

### What I did
- Updated `prescribe/test/*.sh` to:
  - build a local binary (`/tmp/prescribe`) for consistent execution
  - switch commands to the new grouped layout:
    - `session init/show/save/load`
    - `filter add/list/test/show/remove/clear`
    - `file toggle`
    - `context add`
    - `generate` flags via GenerationLayer (`--load-session`, `--output-file`)
  - replace legacy YAML export (`--yaml`) with Glazed formatting (`--output yaml`)
  - gate `generate` behind `PRESCRIBE_RUN_GENERATE=1` so smoke tests don’t require API credentials by default
- Updated `prescribe/README.md`, `prescribe/PROJECT-SUMMARY.md`, and `prescribe/PLAYBOOK-Bubbletea-TUI-Development.md` to match the new command names and output flags.

### Why
- Keep smoke tests runnable after the no-back-compat migration.
- Prevent docs/scripts from teaching a CLI interface that no longer exists.

### What worked
- The updated smoke scripts run successfully with `PRESCRIBE_RUN_GENERATE` unset (generation tests skipped).

### What was tricky to build
- The earlier docs/scripts assumed a “flat” CLI (`show`, `add-filter`, etc.); mapping those to grouped commands required systematic updates across multiple files.

### What warrants a second pair of eyes
- Confirm the doc examples are now consistent everywhere we care about (especially `--load-session` vs the old `--session`).

### What should be done in the future
- If we later re-introduce a compatibility layer (unlikely for this ticket), we should do it intentionally and update scripts accordingly. For now, the no-back-compat contract is reflected in smoke scripts and docs.

### Code review instructions
- Start in:
  - `prescribe/test/test-session-cli.sh`
  - `prescribe/test/test-all.sh`
  - `prescribe/test/test-filters.sh`
  - `prescribe/README.md`
- Validate with:
  - `cd prescribe && bash test/test-cli.sh`
