---
Title: Command Mapping Analysis
Ticket: 005-PORT-GLAZED
Status: active
Topics:
    - glazed
    - prescribe
    - porting
    - cli
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Comprehensive analysis mapping Prescribe CLI commands to Glazed framework patterns, including layer design and command structure
LastUpdated: 2025-12-27T15:08:54.054377699-05:00
WhatFor: Design document for porting Prescribe commands to Glazed
WhenToUse: Reference during implementation to understand command structure and layer design
---

# Command Mapping Analysis: Prescribe → Glazed

## Executive Summary

This document provides a systematic analysis of all Prescribe CLI commands and maps them to Glazed framework patterns. The goal is to migrate Prescribe from its current Cobra-based implementation to Glazed, which provides structured output capabilities, reusable parameter layers, and a more consistent command interface.

Prescribe currently has 14 commands organized into 5 command groups (filter, session, file, context, plus root-level generate and tui). Most commands operate on a session state that tracks PR diffs, filters, and context. The current implementation uses Cobra directly with persistent flags for repository configuration, and each command manually handles session loading and saving.

By porting to Glazed, we gain several benefits: commands that display data can output structured formats (JSON, YAML, CSV) automatically, configuration can be organized into reusable layers, and the help system becomes more powerful. The migration also provides an opportunity to improve consistency across commands and make the codebase more maintainable.

**Key Findings:**
- 14 commands across 5 command groups
- 2 persistent flags (repo, target) used by all commands that are perfect candidates for a RepositoryLayer
- 4 reusable layers identified: Repository, Session, Filter, and Generation
- 5 commands are excellent candidates for structured output (filter list, filter show, filter test, session show, generate)
- 9 commands will remain as simple text-output commands (BareCommand)

## Table of Contents

1. [Understanding the Glazed Framework](#understanding-the-glazed-framework)
2. [Prescribe Command Inventory](#prescribe-command-inventory)
3. [Layer Design Strategy](#layer-design-strategy)
4. [Command-by-Command Mapping](#command-by-command-mapping)
5. [Migration Strategy](#migration-strategy)

## Understanding the Glazed Framework

Before mapping Prescribe commands to Glazed, it's important to understand how Glazed works and why it's beneficial for this migration. Glazed is built on top of Cobra but adds structured data processing capabilities. Instead of commands writing directly to stdout with `fmt.Printf`, they create structured `types.Row` objects that Glazed can automatically format into JSON, YAML, CSV, or formatted tables.

The framework introduces the concept of **parameter layers**—reusable groups of related command-line flags that can be shared across multiple commands. This eliminates the need to duplicate flag definitions and provides consistent interfaces. For example, if multiple commands need repository configuration, you create a RepositoryLayer once and reuse it everywhere.

### Core Concepts

Glazed commands follow a consistent pattern. Each command embeds a `CommandDescription` that contains metadata (name, help text, parameters), and uses a settings struct with `glazed.parameter` tags to map command-line flags to Go fields. The framework handles parsing, validation, and help text generation automatically.

There are two main command interfaces: `BareCommand` for simple text output (like `fmt.Printf`) and `GlazeCommand` for structured data output. Commands can also implement both interfaces to support dual-mode operation—users can get human-readable text by default or structured data with a flag like `--with-glaze-output`.

### Key API Components

The Glazed framework is organized into several packages, each serving a specific purpose. The newer API uses clearer terminology: "sections" instead of "layers" and "fields" instead of "parameters", though they're aliases for the same underlying types.

**Command Definition (`github.com/go-go-golems/glazed/pkg/cmds`):**
- `cmds.CommandDescription` - Container for command metadata (name, help, parameters)
- `cmds.GlazeCommand` - Interface for commands that output structured data
- `cmds.BareCommand` - Interface for commands that output plain text
- `cmds.NewCommandDescription()` - Constructor for creating command descriptions

**Parameter Layers (`github.com/go-go-golems/glazed/pkg/cmds/schema`):**
- `schema.Section` - Interface for a group of related parameters (alias for `layers.ParameterLayer`)
- `schema.NewSection()` - Creates a new parameter section with clearer naming
- `schema.WithFields()` - Attaches field definitions to a section (clearer than `WithParameterDefinitions`)
- `layers.ParsedLayers` - Container holding parsed parameter values from all layers
- `parsedLayers.InitializeStruct(slug, &settings)` - Extracts settings from parsed layers into a struct

**Field Definitions (`github.com/go-go-golems/glazed/pkg/cmds/fields`):**
- `fields.Definition` - Defines a single command-line parameter (alias for `parameters.ParameterDefinition`)
- `fields.New()` - Creates a field definition with clearer naming
- `fields.Type*` - Type constants (TypeString, TypeInteger, TypeBool, TypeChoice, TypeStringList, etc.)
- Options like `fields.WithDefault()`, `fields.WithHelp()`, `fields.WithShortFlag()` configure the field

**Note:** The `schema` and `fields` packages provide clearer aliases for the underlying `layers` and `parameters` packages. The newer API uses "section" and "field" terminology which is more intuitive than "layer" and "parameter".

**CLI Integration (`github.com/go-go-golems/glazed/pkg/cli`):**
- `cli.BuildCobraCommand()` - Converts a Glazed command into a Cobra command
- `cli.WithDualMode()` - Enables dual-mode (both BareCommand and GlazeCommand)
- `cli.WithGlazeToggleFlag()` - Sets the flag name for switching to structured output
- `cli.CobraParserConfig` - Configures how parameters are parsed from Cobra flags

**Data Processing (`github.com/go-go-golems/glazed/pkg/middlewares` and `github.com/go-go-golems/glazed/pkg/types`):**
- `middlewares.Processor` - Interface for processing structured data rows
- `types.Row` - A single row of structured data (key-value pairs)
- `types.NewRow()` - Creates a row from key-value pairs
- `types.MRP()` - Helper to create key-value pairs for row construction

**Built-in Layers (`github.com/go-go-golems/glazed/pkg/settings`):**
- `settings.NewGlazedParameterLayers()` - Provides standard output formatting flags (`--output`, `--fields`, `--sort-columns`)

### Command Structure Pattern

Every Glazed command follows this pattern. Here's a complete example showing how all the pieces fit together:

```go
// 1. Command struct embeds CommandDescription
type MyCommand struct {
    *cmds.CommandDescription
}

// 2. Settings struct with glazed.parameter tags
// These tags map command-line flags to struct fields
type MySettings struct {
    Param1 string `glazed.parameter:"param1"`
    Param2 int    `glazed.parameter:"param2"`
}

// 3. Implement GlazeCommand interface
// This method receives parsed parameters and a processor for structured output
func (c *MyCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    parsedLayers *layers.ParsedLayers,
    gp middlewares.Processor,
) error {
    // Extract settings from parsed layers
    settings := &MySettings{}
    if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
        return err
    }
    
    // Business logic here
    // ...
    
    // Output structured data instead of using fmt.Printf
    row := types.NewRow(
        types.MRP("key", value),
    )
    return gp.AddRow(ctx, row)
}

// 4. Constructor creates command with layers
func NewMyCommand() (*MyCommand, error) {
    // Get built-in Glazed layer for output formatting
    glazedLayer, _ := settings.NewGlazedParameterLayers()
    
    // Get custom layer (e.g., RepositoryLayer)
    customLayer, _ := NewCustomLayer()
    
    // Create command description with fields and layers
    cmdDesc := cmds.NewCommandDescription(
        "my-command",
        cmds.WithShort("Description"),
        cmds.WithFlags(
            // Use the newer fields API for clearer naming
            fields.New(
                "param1",
                fields.TypeString,
                fields.WithDefault("default"),
                fields.WithHelp("Help text"),
            ),
        ),
        // Layers (sections) provide reusable parameter groups
        cmds.WithLayersList(glazedLayer, customLayer),
    )
    
    return &MyCommand{CommandDescription: cmdDesc}, nil
}

// 5. Interface compliance check (compile-time verification)
var _ cmds.GlazeCommand = &MyCommand{}

// 6. Cobra integration in main()
cobraCmd, _ := cli.BuildCobraCommand(glazedCmd,
    cli.WithParserConfig(cli.CobraParserConfig{
        ShortHelpLayers: []string{layers.DefaultSlug},
        MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
    }),
)
```

The key insight is that instead of reading Cobra flags directly, you use `parsedLayers.InitializeStruct()` to populate your settings struct. This ensures defaults, validation, and help text stay consistent across the application.

### Using the Newer Schema/Fields API

Glazed provides newer API aliases that use clearer terminology. Instead of `layers.ParameterLayer` and `parameters.ParameterDefinition`, you can use `schema.Section` and `fields.New()`. These aliases make the code more readable and are the recommended approach for new code.

The newer API uses:
- `schema.NewSection()` instead of `layers.NewParameterLayer()` - Creates a parameter section
- `schema.WithFields()` instead of `layers.WithParameterDefinitions()` - Attaches fields to a section
- `fields.New()` instead of `parameters.NewParameterDefinition()` - Creates a field definition
- `fields.TypeString`, `fields.TypeInteger`, etc. instead of `parameters.ParameterTypeString`, etc.
- `fields.WithDefault()`, `fields.WithHelp()`, etc. instead of `parameters.WithDefault()`, etc.

All examples in this document use the newer schema/fields API for clarity. The underlying functionality is identical—these are just clearer aliases.

**Alternative Pattern: AppConfig API**

Glazed also provides an AppConfig API that uses a different pattern for grouped settings. Instead of individual layers, you create a grouped settings struct and register layers with binder functions. This pattern is more declarative but may be overkill for simple commands. The layer pattern (shown above) is more flexible for command composition and is what we'll use for Prescribe.

See `glazed/cmd/examples/appconfig-parser/main.go` for an example of the AppConfig pattern.

### Documentation Resources

The Glazed framework has excellent documentation:
- `glazed/pkg/doc/tutorials/05-build-first-command.md` - Complete tutorial on building commands
- `glazed/pkg/doc/tutorials/custom-layer.md` - Tutorial on creating reusable layers

These tutorials provide detailed examples and explain the design rationale behind the framework's patterns.

## Prescribe Command Inventory

Prescribe is a CLI tool for building GitHub PR descriptions using LLMs. It operates on a session concept—a stateful workspace that tracks changed files, filters, context, and prompt configuration. Users can filter files, add context, and generate PR descriptions based on the session state.

The application currently has 14 commands organized into a clear hierarchy. Understanding this structure is essential for mapping to Glazed, as it helps identify which commands share configuration and which could benefit from structured output.

### Command Hierarchy

The command structure follows a logical grouping:

```
prescribe (root)
├── filter          # File filtering operations
│   ├── add         # Add a filter to session
│   ├── list        # List active filters
│   ├── remove      # Remove a filter
│   ├── clear       # Clear all filters
│   ├── test        # Test filter without applying
│   └── show        # Show filtered files
├── session         # Session management
│   ├── init        # Initialize new session
│   ├── load        # Load session from file
│   ├── save        # Save session to file
│   └── show        # Show session state
├── file            # File operations
│   └── toggle      # Toggle file inclusion
├── context         # Context management
│   └── add         # Add file or note as context
├── generate        # Generate PR description (root-level)
└── tui             # Launch interactive TUI (root-level)
```

### Persistent Configuration

All commands inherit two persistent flags from the root command:
- `--repo, -r` (string, default: ".") - Path to git repository
- `--target, -t` (string, default: "") - Target branch (defaults to main or master)

These flags are used by every command because they determine which git repository and branch comparison to operate on. This makes them perfect candidates for a RepositoryLayer that can be reused across all commands.

### Command Categories

Commands fall into three categories based on their behavior:

**Query Commands** (display data, good for structured output):
- `filter list` - Lists active filters
- `filter show` - Shows filtered files
- `filter test` - Tests filter patterns
- `session show` - Shows session state

**State Modification Commands** (change session, simple text output):
- `filter add` - Adds a filter
- `filter remove` - Removes a filter
- `filter clear` - Clears all filters
- `session init` - Initializes session
- `session load` - Loads session
- `session save` - Saves session
- `file toggle` - Toggles file inclusion
- `context add` - Adds context

**Special Commands**:
- `generate` - Generates PR description (could be dual-mode: text output + metadata)
- `tui` - Interactive TUI (no changes needed)

### Detailed Command Analysis

Each command has specific flags and behavior that need to be preserved during migration. The following sections detail each command's current implementation and how it maps to Glazed patterns.

#### Filter Commands

The filter commands manage file filtering rules that control which files appear in the PR description. Filters use glob patterns and can include or exclude files based on path matching.

**`filter add`** - Adds a filter to the current session. Requires a name and at least one pattern (exclude or include). The command creates the filter, adds it to the session, saves the session, and prints an impact summary showing how many files are now filtered.

**`filter list`** - Displays all active filters in a formatted table. This is a perfect candidate for structured output because users might want to parse filter information programmatically. Currently outputs a human-readable table, but could output JSON/YAML rows.

**`filter remove`** - Removes a filter by index or name. Takes the filter identifier as a positional argument. This is a simple state modification command that doesn't need structured output.

**`filter clear`** - Removes all filters from the session. Simple state modification with text confirmation.

**`filter test`** - Tests filter patterns without actually applying them to the session. This is useful for previewing what a filter would do. Currently outputs formatted text, but structured output would be valuable for scripting and automation.

**`filter show`** - Shows which files are currently filtered out by active filters. Displays file paths with diff statistics. This could benefit from structured output for programmatic processing.

#### Session Commands

Session commands manage the session state—loading, saving, and inspecting the current session.

**`session init`** - Initializes a new session from the current git state. Can optionally auto-save the session after initialization. This is a simple initialization command.

**`session load`** - Loads a session from a YAML file. Takes an optional path argument (defaults to the app's default session path). Prints a summary of what was loaded.

**`session save`** - Saves the current session to a YAML file. Takes an optional path argument. Simple state persistence command.

**`session show`** - Displays the current session state. Has a `--yaml` flag to output raw YAML. This is an excellent candidate for dual-mode: human-readable summary by default, structured data with `--with-glaze-output`.

#### File and Context Commands

**`file toggle`** - Toggles whether a file is included in the PR description context. Takes a file path as a positional argument. Simple state modification.

**`context add`** - Adds additional context to the session—either a file path or a text note via `--note` flag. Saves the session and prints token count. Simple state modification.

#### Root-Level Commands

**`generate`** - Generates a PR description using AI based on the current session. Supports custom prompts, presets, and session loading. Writes output to a file or stdout. This could be dual-mode: text output (the description) by default, structured data (description + metadata) with `--with-glaze-output`.

**`tui`** - Launches an interactive Terminal User Interface using Bubbletea. No changes needed—this is purely interactive and doesn't output structured data.

## Layer Design Strategy

One of Glazed's key benefits is the ability to create reusable parameter layers. Instead of duplicating flag definitions across commands, you create a layer once and reuse it everywhere. This provides consistency, reduces maintenance overhead, and ensures all commands share the same interface for common configuration.

For Prescribe, we've identified four layers that encapsulate related configuration:

### Layer 1: Repository Layer

**Purpose:** Encapsulate repository and branch configuration used by every command.

Every Prescribe command needs to know which git repository to operate on and which branch to compare against. Currently, these are persistent flags on the root command. By creating a RepositoryLayer, we make this configuration explicit and reusable, while also enabling better help text and validation.

The layer provides two parameters:
- `repo` - Path to git repository (defaults to current directory)
- `target` - Target branch for comparison (defaults to main or master)

**Implementation:**

```go
import (
    "github.com/go-go-golems/glazed/pkg/cmds/fields"
    "github.com/go-go-golems/glazed/pkg/cmds/layers"
    "github.com/go-go-golems/glazed/pkg/cmds/schema"
)

const RepositorySlug = "repository"

type RepositorySettings struct {
    RepoPath     string `glazed.parameter:"repo"`
    TargetBranch string `glazed.parameter:"target"`
}

func NewRepositoryLayer() (schema.Section, error) {
    return schema.NewSection(
        RepositorySlug,
        "Repository Configuration",
        schema.WithFields(
            fields.New(
                "repo",
                fields.TypeString,
                fields.WithDefault("."),
                fields.WithHelp("Path to git repository"),
                fields.WithShortFlag("r"),
            ),
            fields.New(
                "target",
                fields.TypeString,
                fields.WithDefault(""),
                fields.WithHelp("Target branch (default: main or master)"),
                fields.WithShortFlag("t"),
            ),
        ),
    )
}

func GetRepositorySettings(parsedLayers *layers.ParsedLayers) (*RepositorySettings, error) {
    settings := &RepositorySettings{}
    if err := parsedLayers.InitializeStruct(RepositorySlug, settings); err != nil {
        return nil, fmt.Errorf("failed to initialize repository settings: %w", err)
    }
    return settings, nil
}
```

**Usage:** All 14 commands will use this layer.

### Layer 2: Session Layer

**Purpose:** Encapsulate session file management configuration.

Many Prescribe commands need to load or save session files. Some commands have flags for custom session paths or auto-save behavior. By creating a SessionLayer, we provide a consistent interface for session management across all commands that need it.

The layer provides:
- `session-path` - Custom path to session file (defaults to app's default)
- `auto-save` - Automatically save session after operations

**Implementation:**

```go
const SessionSlug = "session"

type SessionSettings struct {
    SessionPath string `glazed.parameter:"session-path"`
    AutoSave    bool   `glazed.parameter:"auto-save"`
}

func NewSessionLayer() (schema.Section, error) {
    return schema.NewSection(
        SessionSlug,
        "Session Configuration",
        schema.WithFields(
            fields.New(
                "session-path",
                fields.TypeString,
                fields.WithDefault(""),
                fields.WithHelp("Path to session file (default: app default session path)"),
                fields.WithShortFlag("p"),
            ),
            fields.New(
                "auto-save",
                fields.TypeBool,
                fields.WithDefault(false),
                fields.WithHelp("Automatically save session after operations"),
            ),
        ),
    )
}
```

**Usage:** Commands that read/write sessions (filter, session, file, context, generate, tui).

### Layer 3: Filter Layer

**Purpose:** Encapsulate filter creation parameters.

The `filter add` and `filter test` commands share the same parameters for defining filters. By creating a FilterLayer, we avoid duplication and ensure consistency. The layer provides:
- `filter-name` - Name of the filter
- `filter-description` - Optional description
- `exclude-patterns` - List of exclude patterns (glob syntax)
- `include-patterns` - List of include patterns (glob syntax)

**Implementation:**

```go
const FilterSlug = "filter"

type FilterSettings struct {
    Name            string   `glazed.parameter:"filter-name"`
    Description     string   `glazed.parameter:"filter-description"`
    ExcludePatterns []string `glazed.parameter:"exclude-patterns"`
    IncludePatterns []string `glazed.parameter:"include-patterns"`
}

func NewFilterLayer() (schema.Section, error) {
    return schema.NewSection(
        FilterSlug,
        "Filter Configuration",
        schema.WithFields(
            fields.New(
                "filter-name",
                fields.TypeString,
                fields.WithDefault(""),
                fields.WithHelp("Filter name"),
                fields.WithShortFlag("n"),
            ),
            fields.New(
                "filter-description",
                fields.TypeString,
                fields.WithDefault(""),
                fields.WithHelp("Filter description"),
                fields.WithShortFlag("d"),
            ),
            fields.New(
                "exclude-patterns",
                fields.TypeStringList,
                fields.WithDefault([]string{}),
                fields.WithHelp("Exclude patterns (glob syntax)"),
                fields.WithShortFlag("e"),
            ),
            fields.New(
                "include-patterns",
                fields.TypeStringList,
                fields.WithDefault([]string{}),
                fields.WithHelp("Include patterns (glob syntax)"),
                fields.WithShortFlag("i"),
            ),
        ),
    )
}
```

**Usage:** `filter add` and `filter test` commands.

### Layer 4: Generation Layer

**Purpose:** Encapsulate PR generation parameters.

The `generate` command has several parameters for controlling how PR descriptions are generated. These are specific to generation and don't need to be shared with other commands, but organizing them into a layer provides consistency and makes the command structure clearer.

The layer provides:
- `prompt` - Custom prompt text
- `preset` - Prompt preset ID
- `load-session` - Load session file before generating
- `output-file` - Output file path (defaults to stdout)

**Implementation:**

```go
const GenerationSlug = "generation"

type GenerationSettings struct {
    Prompt      string `glazed.parameter:"prompt"`
    Preset      string `glazed.parameter:"preset"`
    LoadSession string `glazed.parameter:"load-session"`
    OutputFile  string `glazed.parameter:"output-file"`
}

func NewGenerationLayer() (schema.Section, error) {
    return schema.NewSection(
        GenerationSlug,
        "Generation Configuration",
        schema.WithFields(
            fields.New(
                "prompt",
                fields.TypeString,
                fields.WithDefault(""),
                fields.WithHelp("Custom prompt text"),
                fields.WithShortFlag("p"),
            ),
            fields.New(
                "preset",
                fields.TypeString,
                fields.WithDefault(""),
                fields.WithHelp("Prompt preset ID"),
            ),
            fields.New(
                "load-session",
                fields.TypeString,
                fields.WithDefault(""),
                fields.WithHelp("Load session file before generating"),
                fields.WithShortFlag("s"),
            ),
            fields.New(
                "output-file",
                fields.TypeString,
                fields.WithDefault(""),
                fields.WithHelp("Output file (default: stdout)"),
                fields.WithShortFlag("o"),
            ),
        ),
    )
}
```

**Usage:** `generate` command only.

## Command-by-Command Mapping

This section provides detailed mappings for each command, showing how the current Cobra implementation translates to Glazed patterns. For each command, we show the command type (BareCommand, GlazeCommand, or Dual), the settings struct, layers used, and pseudocode for the implementation.

### Filter Commands

The filter commands are a good example of the migration pattern. Some commands are simple state modifications (add, remove, clear) that stay as BareCommand, while others display data (list, show, test) and benefit from structured output.

#### `filter add` → `FilterAddCommand`

**Type:** BareCommand (state modification with simple text confirmation)

**Rationale:** This command modifies session state and outputs a simple confirmation message. There's no need for structured output here—users just need to know the filter was added successfully.

**Settings:**
```go
type FilterAddSettings struct {
    // From FilterLayer
    Name            string   `glazed.parameter:"filter-name"`
    Description     string   `glazed.parameter:"filter-description"`
    ExcludePatterns []string `glazed.parameter:"exclude-patterns"`
    IncludePatterns []string `glazed.parameter:"include-patterns"`
}
```

**Layers:** RepositoryLayer, SessionLayer, FilterLayer

**Implementation Pseudocode:**
```
1. Extract settings from parsed layers (repository, session, filter)
2. Create controller with repository settings
3. Initialize controller with target branch
4. Load default session if it exists
5. Build filter rules from exclude/include patterns
6. Create filter domain object
7. Add filter to controller
8. Save session (use session-path if provided, else default)
9. Print confirmation message with impact summary
```

The implementation follows the same logic as the current command but uses layers for configuration instead of reading Cobra flags directly.

#### `filter list` → `FilterListCommand`

**Type:** Dual Command (BareCommand + GlazeCommand)

**Rationale:** This command displays data, making it a good candidate for structured output. Users might want to parse filter information programmatically (e.g., in scripts or CI/CD pipelines). Dual mode provides the best of both worlds: human-readable tables by default, structured data when needed.

**Settings:** None (reads from session state)

**Layers:** RepositoryLayer, SessionLayer, GlazedLayer (for output format)

**BareCommand Implementation:**
The text mode outputs a formatted table showing each filter with its rules and impact, matching the current behavior.

**GlazeCommand Implementation:**
The structured mode outputs one row per filter with columns:
- `name` - Filter name
- `description` - Filter description
- `rule_count` - Number of rules in the filter
- `visible_files` - Number of files visible after filtering
- `filtered_files` - Number of files filtered out

This allows users to run `prescribe filter list --with-glaze-output --output json` to get machine-parseable data.

**Cobra Integration:**
```go
cobraCmd, _ := cli.BuildCobraCommand(filterListCmd,
    cli.WithDualMode(true),
    cli.WithGlazeToggleFlag("with-glaze-output"),
    cli.WithParserConfig(cli.CobraParserConfig{
        ShortHelpLayers: []string{layers.DefaultSlug},
        MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
    }),
)
```

#### `filter test` → `FilterTestCommand`

**Type:** Dual Command

**Rationale:** Testing filters produces structured data (which files match, which are filtered) that's valuable for scripting and automation. Structured output makes it easy to process test results programmatically.

**Settings:**
```go
type FilterTestSettings struct {
    Name            string   `glazed.parameter:"filter-name"`
    ExcludePatterns []string `glazed.parameter:"exclude-patterns"`
    IncludePatterns []string `glazed.parameter:"include-patterns"`
}
```

**Layers:** RepositoryLayer, FilterLayer, GlazedLayer

**GlazeCommand Implementation:**
Outputs one row per file with columns:
- `file_path` - Path to the file
- `status` - "matched" or "filtered"
- `additions` - Number of additions in diff
- `deletions` - Number of deletions in diff
- `tokens` - Token count for the file

This allows users to test filters and process results in scripts, making it easy to validate filter patterns before applying them.

#### `filter show` → `FilterShowCommand`

**Type:** Dual Command

**Rationale:** Similar to `filter test`, this displays structured data (which files are filtered) that benefits from machine-readable output.

**GlazeCommand Implementation:**
Outputs one row per filtered file with columns:
- `file_path` - Path to the file
- `additions` - Number of additions
- `deletions` - Number of deletions
- `tokens` - Token count
- `filtered_by` - List of filter names that filtered this file

### Session Commands

Session commands manage session state. Most are simple state operations, but `session show` displays data and benefits from structured output.

#### `session show` → `SessionShowCommand`

**Type:** Dual Command

**Rationale:** This command already has a `--yaml` flag, showing that users want structured output. Dual mode provides a better interface: human-readable summary by default, structured data with `--with-glaze-output`.

**Settings:**
```go
type SessionShowSettings struct {
    YAML bool `glazed.parameter:"yaml"` // Keep for backward compatibility
}
```

**Layers:** RepositoryLayer, SessionLayer, GlazedLayer

**GlazeCommand Implementation:**
Outputs a single row with all session state:
- `source_branch` - Source branch name
- `target_branch` - Target branch name
- `total_files` - Total changed files
- `visible_files` - Files visible after filtering
- `included_files` - Files included in context
- `filtered_files` - Files filtered out
- `active_filters` - Number of active filters
- `context_items` - Number of context items
- `token_count` - Total token count
- `current_preset` - Current preset name (if using preset)
- `current_prompt_preview` - Preview of current prompt (if custom)

This provides a complete snapshot of session state in a machine-parseable format.

### Generate Command

#### `generate` → `GenerateCommand`

**Type:** Dual Command

**Rationale:** The generate command produces text output (the PR description), but it could also output metadata about the generation (tokens used, preset, files included). Dual mode allows users to get just the description (text mode) or description plus metadata (glaze mode).

**Settings:**
```go
type GenerateSettings struct {
    Prompt      string `glazed.parameter:"prompt"`
    Preset      string `glazed.parameter:"preset"`
    LoadSession string `glazed.parameter:"load-session"`
    OutputFile  string `glazed.parameter:"output-file"`
}
```

**Layers:** RepositoryLayer, SessionLayer, GenerationLayer, GlazedLayer

**BareCommand Implementation:**
Generates the PR description and writes it to the output file (or stdout). This matches current behavior.

**GlazeCommand Implementation:**
Outputs a single row with:
- `description` - The generated PR description text
- `tokens_used` - Total tokens used
- `preset_used` - Preset ID if used
- `prompt_preview` - Preview of prompt if custom
- `files_included` - Number of files included
- `context_items` - Number of context items

This allows users to get both the description and metadata about how it was generated, useful for tracking and optimization.

## Glazed Program Initialization

Before porting individual commands, we need to transform Prescribe into a proper Glazed program. This involves setting up the root command with logging initialization and integrating the Glazed help system. This foundational work enables all the benefits of the Glazed framework and should be done first.

### Root Command Setup

A Glazed program requires three key components in its main entry point:

1. **Logging Initialization** - Add logging flags and initialize the logger before command execution
2. **Help System** - Set up the enhanced help system for better documentation and help text
3. **Command Registration** - Register Glazed commands using `cli.BuildCobraCommand()`

### Logging Setup Pattern

Glazed provides a logging layer that adds standard logging flags (`--log-level`, `--log-format`, `--log-file`, etc.) to the root command. The logger must be initialized in `PersistentPreRunE` so it's active before any command logic runs.

The pattern is:

```go
var rootCmd = &cobra.Command{
    Use:   "prescribe",
    Short: "A TUI for building GitHub PR descriptions",
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        // Initialize logger after Cobra has parsed flags
        return logging.InitLoggerFromCobra(cmd)
    },
}

func main() {
    // Add logging flags to root command
    err := logging.AddLoggingLayerToRootCommand(rootCmd, "prescribe")
    cobra.CheckErr(err)
    
    // ... rest of initialization
}
```

**Key Points:**
- `logging.AddLoggingLayerToRootCommand()` adds persistent flags for logging configuration
- `logging.InitLoggerFromCobra()` reads those flags and initializes the global logger
- Initialization happens in `PersistentPreRunE` so logging is active for all commands
- The app name ("prescribe") is used for Logstash integration if enabled

### Help System Setup

The Glazed help system provides enhanced help functionality, including contextual help, topic-based documentation, and better formatting. Setting it up is straightforward:

```go
import (
    "github.com/go-go-golems/glazed/pkg/help"
    help_cmd "github.com/go-go-golems/glazed/pkg/help/cmd"
)

func main() {
    // ... logging setup ...
    
    // Create help system
    helpSystem := help.NewHelpSystem()
    
    // Optionally load documentation from embedded files or filesystem
    // err := helpSystem.LoadSectionsFromFS(docsFS, ".")
    // cobra.CheckErr(err)
    
    // Set up help system with root command
    help_cmd.SetupCobraRootCommand(helpSystem, rootCmd)
    
    // ... register commands ...
}
```

**What `SetupCobraRootCommand` provides:**
- Enhanced help templates with better formatting
- `help` command for topic-based documentation
- Contextual help that shows relevant parameters
- Support for `--long-help` flag for detailed help
- UI mode support (`help --ui`) for interactive help browsing

### Complete Main Function Pattern

Here's how the complete main function should look for Prescribe:

```go
package main

import (
    "github.com/go-go-golems/glazed/pkg/cli"
    "github.com/go-go-golems/glazed/pkg/cmds/logging"
    "github.com/go-go-golems/glazed/pkg/help"
    help_cmd "github.com/go-go-golems/glazed/pkg/help/cmd"
    "github.com/go-go-golems/prescribe/cmd/prescribe/cmds"
    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:     "prescribe",
    Short:   "A TUI for building GitHub PR descriptions",
    Version: "0.1.0",
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        return logging.InitLoggerFromCobra(cmd)
    },
}

func main() {
    // 1. Add logging flags
    err := logging.AddLoggingLayerToRootCommand(rootCmd, "prescribe")
    cobra.CheckErr(err)
    
    // 2. Set up help system
    helpSystem := help.NewHelpSystem()
    help_cmd.SetupCobraRootCommand(helpSystem, rootCmd)
    
    // 3. Register Glazed commands
    // Commands will be registered here using cli.BuildCobraCommand()
    // This replaces the current registerCommands() function
    
    // 4. Execute
    if err := rootCmd.Execute(); err != nil {
        cobra.CheckErr(err)
    }
}
```

### Migration from Current Structure

Prescribe currently has a simple structure:
- `main.go` just calls `cmds.Execute()`
- `cmds/root.go` defines the root command and registers subcommands
- Persistent flags are added in `init()`

**Migration Steps:**
1. Move root command definition to `main.go` (or keep in `cmds/root.go` but update it)
2. Add `PersistentPreRunE` for logging initialization
3. Add logging layer setup in main
4. Add help system setup in main
5. Update command registration to use `cli.BuildCobraCommand()` for Glazed commands
6. Remove old persistent flag definitions (they'll come from RepositoryLayer)

### Benefits of This Setup

**Logging Benefits:**
- Consistent logging interface across all commands
- Configurable log levels for debugging
- Support for structured logging (JSON format)
- File logging and Logstash integration options
- Logging active before command execution (useful for debugging command loading)

**Help System Benefits:**
- Better formatted help text
- Topic-based documentation (`prescribe help <topic>`)
- Contextual help showing relevant parameters
- Interactive help UI (`prescribe help --ui`)
- Easier to maintain and extend documentation

**Integration Benefits:**
- Foundation for all Glazed features
- Consistent command interface
- Better error handling and validation
- Enhanced debugging capabilities

This initialization work is foundational and should be completed in Phase 1 before porting individual commands.

## Migration Strategy

Migrating Prescribe to Glazed is a significant undertaking that should be done incrementally to minimize risk and allow for testing at each stage. The migration strategy is organized into four phases, each building on the previous one.

### Phase 1: Infrastructure Setup

Before porting any commands, we need to set up the infrastructure that all commands will use. This includes creating the layer packages and helper functions.

**Tasks:**
1. Create layer packages in `prescribe/pkg/layers/`:
   - `repository.go` - RepositoryLayer implementation
   - `session.go` - SessionLayer implementation
   - `filter.go` - FilterLayer implementation
   - `generation.go` - GenerationLayer implementation

2. Create helper functions for each layer:
   - Settings extraction functions (e.g., `GetRepositorySettings()`)
   - Controller initialization helpers that use layers

3. Update root command:
   - Integrate Glazed help system
   - Add Glazed layers to root command
   - Ensure backward compatibility with existing flags

**Success Criteria:**
- All layers can be created without errors
- Settings can be extracted from parsed layers
- Root command still works with existing flags
- Help system displays layer information correctly

### Phase 2: Query Commands (Structured Output)

Port commands that primarily display data. These commands benefit most from structured output and are good candidates for dual-mode operation.

**Commands to Port:**
1. `filter list` - Dual command
2. `filter show` - Dual command
3. `filter test` - Dual command
4. `session show` - Dual command

**Benefits:**
- Users can get JSON/YAML output for scripting
- Consistent output format across commands
- Easy to test and validate
- Demonstrates Glazed capabilities

**Testing:**
- Verify text output matches current behavior
- Verify structured output (JSON/YAML) works correctly
- Test dual-mode flag (`--with-glaze-output`)
- Ensure backward compatibility

### Phase 3: State Modification Commands

Port commands that modify state. These stay as BareCommand but use layers for configuration.

**Commands to Port:**
1. `filter add` - BareCommand (uses FilterLayer)
2. `generate` - Dual command (text output + metadata)

**Rationale:**
- `filter add` is a simple state modification that doesn't need structured output
- `generate` could benefit from dual-mode to provide metadata alongside the description

### Phase 4: Simple Commands

Port remaining simple commands. These are straightforward state operations that stay as BareCommand.

**Commands to Port:**
1. `filter remove` - BareCommand
2. `filter clear` - BareCommand
3. `session init` - BareCommand
4. `session load` - BareCommand
5. `session save` - BareCommand
6. `file toggle` - BareCommand
7. `context add` - BareCommand
8. `tui` - BareCommand (no changes needed, just uses layers)

**Rationale:**
These commands are simple state modifications that don't need structured output. Porting them completes the migration and ensures all commands use the same layer infrastructure.

### Testing Strategy

Testing is critical for ensuring the migration doesn't break existing functionality. The testing strategy covers multiple levels:

**Unit Tests:**
- Test layer creation and parameter definitions
- Test settings extraction from parsed layers
- Test helper functions

**Integration Tests:**
- Test commands with mock controllers
- Test layer composition
- Test dual-mode switching

**E2E Tests:**
- Test full command execution with real git repositories
- Test session loading and saving
- Test filter operations
- Test generation

**Backward Compatibility Tests:**
- Ensure existing scripts still work
- Verify flag names and behavior match current implementation
- Test positional arguments
- Verify text output format matches current behavior

### Backward Compatibility

Maintaining backward compatibility is essential. Users have existing scripts and workflows that depend on current behavior. The migration should be transparent to users who don't use new features.

**Compatibility Requirements:**
- Keep existing flag names where possible
- Maintain same default behavior
- Preserve positional argument handling
- Keep text output format identical for BareCommand mode
- Ensure existing scripts continue to work

**Migration Path:**
- Phase 1 can be done without breaking changes (infrastructure only)
- Phases 2-4 can be done incrementally, testing each command
- New features (structured output) are opt-in via flags
- Old behavior remains the default

## Summary

This analysis provides a comprehensive roadmap for migrating Prescribe commands to the Glazed framework. The migration will provide structured output capabilities, reusable configuration layers, and a more consistent command interface, while maintaining backward compatibility with existing scripts and workflows.

**Commands to Port:**
- 14 total commands
- 5 dual commands (filter list, filter show, filter test, session show, generate)
- 9 bare commands (state modification, simple operations)
- 1 interactive command (tui - no changes needed)

**Layers to Create:**
- RepositoryLayer (used by all commands)
- SessionLayer (used by most commands)
- FilterLayer (used by filter commands)
- GenerationLayer (used by generate)

**Key Benefits:**
- Structured output for scripting and automation (JSON, YAML, CSV)
- Reusable configuration layers reduce duplication
- Consistent command interface across all commands
- Better help system integration
- Easier testing and validation
- Maintainable codebase with clear separation of concerns

The migration strategy is incremental and low-risk, allowing for testing at each phase while maintaining backward compatibility throughout the process.
