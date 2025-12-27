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
