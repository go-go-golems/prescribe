---
Title: 'Analysis: Integrating Pinocchio Profiles with Generate Command using appconfig'
Ticket: 012-USE-PINOCCHIO-PROFILES
Status: active
Topics:
    - configuration
    - profiles
    - appconfig
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-27T20:54:07.142247695-05:00
WhatFor: ""
WhenToUse: ""
---

# Analysis: Integrating Pinocchio Profiles with Generate Command using appconfig

## Executive Summary

This analysis examines how to integrate Pinocchio profiles (`~/.config/pinocchio/profiles.yaml`) into the `prescribe generate` command using the `appconfig` package. The goal is to enable profile-based configuration loading while refactoring the command initialization to use `appconfig.Parser` for cleaner, more maintainable configuration parsing.

## Current State Analysis

### Current Implementation (`generate.go`)

**File:** `prescribe/cmd/prescribe/cmds/generate.go`

**Current Architecture:**
- Uses `cli.BuildCobraCommand()` with `CobraParserConfig`
- Relies on `CobraCommandDefaultMiddlewares` for middleware chain
- Manually constructs layers and registers them with `CommandDescription`
- Parses settings using `parsedLayers.InitializeStruct()` in `Run()` method

**Key Functions:**
- `NewGenerateCommand()` (lines 37-99): Constructs command description with layers
- `Run()` (lines 101-220): Executes command, parses layers manually
- `InitGenerateCmd()` (lines 222-239): Builds cobra command using `cli.BuildCobraCommand()`

**Current Layers:**
1. `RepositoryLayer` (via `prescribe_layers.NewRepositoryLayer()`)
2. `GenerationLayer` (via `prescribe_layers.NewGenerationLayer()`)
3. `GeppettoLayers` (via `geppettolayers.CreateGeppettoLayers()`)

**Current Parsing Flow:**
```
InitGenerateCmd() 
  → NewGenerateCommand() 
    → Builds CommandDescription with layers
  → cli.BuildCobraCommand(glazedCmd, WithParserConfig(...))
    → Creates CobraParser with CobraCommandDefaultMiddlewares
  → Run() method
    → parsedLayers.InitializeStruct() for each settings struct
```

### Profile System Architecture

**Profile File Location:** `~/.config/pinocchio/profiles.yaml`

**Profile File Format:**
```yaml
profile-name:
  layer-slug:
    parameter-name: parameter-value
  another-layer:
    parameter-name: parameter-value
```

**Profile Loading Components:**

1. **ProfileSettings Layer** (`glazed/pkg/cli/cli.go`):
   - Provides `--profile` and `--profile-file` flags
   - Struct: `ProfileSettings{Profile, ProfileFile}`
   - Slug: `cli.ProfileSettingsSlug = "profile-settings"`
   - Function: `cli.NewProfileSettingsLayer()`

2. **Profile Middleware** (`glazed/pkg/cmds/middlewares/profiles.go`):
   - `GatherFlagsFromProfiles(defaultProfileFile, profileFile, profile, ...options)`
   - Loads YAML, extracts profile map, updates `ParsedLayers`
   - Precedence: profiles override defaults, but are overridden by config/env/flags

3. **Bootstrap Pattern** (used in `geppetto/pkg/layers/layers.go`):
   - Parse `ProfileSettings` from defaults + config + env + flags first
   - Then use resolved profile settings to load profile middleware
   - Prevents circular dependency between profile selection and profile loading

### appconfig Package Architecture

**Location:** `glazed/pkg/appconfig/`

**Core Components:**

1. **Parser[T]** (`parser.go`):
   - Generic parser for typed settings struct `T`
   - Registers layers with binders to struct fields
   - Executes middleware chain to populate `T`

2. **ParserOptions** (`options.go`):
   - `WithDefaults()`: Adds defaults middleware
   - `WithEnv(prefix)`: Adds environment variable parsing
   - `WithConfigFiles(...files)`: Adds config file loading
   - `WithCobra(cmd, args)`: Adds cobra flag/arg parsing
   - `WithMiddlewares(...)`: Escape hatch for custom middlewares
   - `WithValuesForLayers(...)`: Programmatic values

3. **Middleware Ordering:**
   - Options are applied in low→high precedence order
   - Internally reversed for execution (high→low precedence)
   - Standard order: defaults < config < env < args < flags

**Current Limitations:**
- No built-in `WithProfile()` option
- Profile loading requires manual middleware construction
- ProfileSettings layer must be registered separately

## Requirements

### Functional Requirements

1. **Profile Loading:**
   - Load profiles from `~/.config/pinocchio/profiles.yaml` by default
   - Support `--profile` flag to select profile name (default: "default")
   - Support `--profile-file` flag to override profile file path
   - Support `PRESCRIBE_PROFILE` and `PRESCRIBE_PROFILE_FILE` environment variables

2. **appconfig Integration:**
   - Refactor `InitGenerateCmd()` to use `appconfig.Parser`
   - Create grouped settings struct for all layers
   - Register all layers with appropriate binders
   - Use appconfig options for profile, config, env, cobra parsing

3. **Backward Compatibility:**
   - Existing flags and behavior must continue to work
   - No breaking changes to command interface
   - Existing layer structures remain unchanged

### Non-Functional Requirements

1. **Precedence Order:**
   - Defaults < Profiles < Config Files < Environment < CLI Flags
   - Must match existing Glazed precedence model

2. **Error Handling:**
   - Clear errors when profile file doesn't exist (if explicitly requested)
   - Graceful handling when default profile file doesn't exist
   - Validation errors for invalid profile names

## Proposed Architecture

### Settings Struct Design

```go
type GenerateAppSettings struct {
    // Profile settings (for profile selection)
    ProfileSettings cli.ProfileSettings
    
    // Command-specific layers
    Repository  prescribe_layers.RepositorySettings
    Generation  prescribe_layers.GenerationSettings
    Extra       GenerateExtraSettings
    
    // Geppetto layers (AI inference settings)
    // Note: Geppetto layers are complex, may need separate struct
    // or use map[string]interface{} with manual extraction
}
```

### Parser Initialization Flow

```go
func InitGenerateCmd() error {
    // 1. Create appconfig.Parser with options
    parser, err := appconfig.NewParser[GenerateAppSettings](
        appconfig.WithDefaults(),
        appconfig.WithProfile("pinocchio", "default"), // NEW: profile support
        appconfig.WithConfigFiles(...), // Optional config files
        appconfig.WithEnv("PRESCRIBE"),
        appconfig.WithCobra(cmd, args), // Applied in Run()
    )
    
    // 2. Register all layers
    parser.Register("profile-settings", profileLayer, func(s *GenerateAppSettings) any {
        return &s.ProfileSettings
    })
    parser.Register("repository", repoLayer, func(s *GenerateAppSettings) any {
        return &s.Repository
    })
    // ... etc
    
    // 3. Build cobra command with parser integration
    // This requires new integration point in cli.BuildCobraCommand
}
```

### Profile Integration Options

**Option A: Add `WithProfile()` to appconfig (Recommended)**

Add new option to `appconfig/options.go`:

```go
// WithProfile configures profile loading from pinocchio profiles.yaml
func WithProfile(appName string, defaultProfile string) ParserOption {
    return func(o *parserOptions) error {
        xdgConfigPath, err := os.UserConfigDir()
        if err != nil {
            return errors.Wrap(err, "failed to get user config directory")
        }
        defaultProfileFile := filepath.Join(xdgConfigPath, appName, "profiles.yaml")
        
        // Bootstrap parse ProfileSettings first
        // Then add profile middleware using resolved settings
        // This requires two-phase parsing or middleware composition
        
        return nil
    }
}
```

**Option B: Manual Profile Middleware (Current Approach)**

Use `WithMiddlewares()` to add profile loading:

```go
parser, err := appconfig.NewParser[GenerateAppSettings](
    appconfig.WithDefaults(),
    appconfig.WithMiddlewares(
        // Profile middleware (after defaults, before config)
        buildProfileMiddleware("pinocchio", "default"),
    ),
    appconfig.WithConfigFiles(...),
    appconfig.WithEnv("PRESCRIBE"),
    appconfig.WithCobra(cmd, args),
)
```

**Option C: Hybrid - Profile Helper Function**

Create helper that returns profile-aware parser options:

```go
func WithPinocchioProfile(defaultProfile string) []appconfig.ParserOption {
    return []appconfig.ParserOption{
        appconfig.WithDefaults(),
        // Bootstrap profile settings
        // Add profile middleware
        // Add config/env/cobra
    }
}
```

## Implementation Plan

### Phase 1: Add Profile Support to appconfig

**File:** `glazed/pkg/appconfig/options.go`

**Changes:**
1. Add `WithProfile(appName, defaultProfile)` function
2. Implement bootstrap parsing of ProfileSettings
3. Add profile middleware to chain at correct precedence

**Pseudocode:**
```go
func WithProfile(appName string, defaultProfile string) ParserOption {
    return func(o *parserOptions) error {
        // 1. Resolve default profile file path
        xdgConfigPath, err := os.UserConfigDir()
        defaultProfileFile := filepath.Join(xdgConfigPath, appName, "profiles.yaml")
        
        // 2. Create bootstrap parser for ProfileSettings
        bootstrapParser := NewParser[cli.ProfileSettings](
            WithDefaults(),
            WithEnv(strings.ToUpper(appName)),
            // Note: Can't use WithCobra here as cmd not available yet
        )
        
        // 3. Parse profile settings (will be re-parsed in full chain)
        // Actually, we need cmd/args for cobra parsing...
        // This suggests we need a two-phase approach or callback
        
        // Alternative: Add profile middleware that bootstraps internally
        o.middlewares = append(o.middlewares,
            buildProfileMiddleware(appName, defaultProfileFile, defaultProfile),
        )
        
        return nil
    }
}

func buildProfileMiddleware(appName, defaultProfileFile, defaultProfile string) cmd_middlewares.Middleware {
    return func(next cmd_middlewares.HandlerFunc) cmd_middlewares.HandlerFunc {
        return func(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
            // Bootstrap parse ProfileSettings
            // Load profile from file
            // Update parsedLayers with profile values
            return next(layers_, parsedLayers)
        }
    }
}
```

**Challenge:** Profile settings need to be parsed from cobra/env/config before profile can be loaded, but `WithCobra()` is called later. This requires either:
- Two-phase parsing (bootstrap then full)
- Middleware that bootstraps internally
- Callback pattern for profile resolution

### Phase 2: Refactor generate.go to use appconfig

**File:** `prescribe/cmd/prescribe/cmds/generate.go`

**Changes:**

1. **Create Settings Struct:**
```go
type GenerateAppSettings struct {
    ProfileSettings cli.ProfileSettings `glazed.parameter:"profile-settings"`
    Repository      prescribe_layers.RepositorySettings `glazed.parameter:"repository"`
    Generation      prescribe_layers.GenerationSettings `glazed.parameter:"generation"`
    Extra           GenerateExtraSettings `glazed.parameter:"default"`
    // Geppetto settings - may need manual extraction
}
```

2. **Refactor InitGenerateCmd():**
```go
func InitGenerateCmd() error {
    // Create layers
    repoLayer, err := prescribe_layers.NewRepositoryLayer()
    // ... create other layers
    
    // Create parser with profile support
    parser, err := appconfig.NewParser[GenerateAppSettings](
        appconfig.WithDefaults(),
        appconfig.WithProfile("pinocchio", "default"),
        appconfig.WithEnv("PRESCRIBE"),
        // WithCobra will be added in Run()
    )
    
    // Register layers
    parser.Register("profile-settings", profileLayer, ...)
    parser.Register("repository", repoLayer, ...)
    // ... register all layers
    
    // Build cobra command
    // Need integration point: cli.BuildCobraCommandWithParser()
    // OR: Use parser in Run() method instead of InitGenerateCmd()
}
```

3. **Refactor Run() Method:**
```go
func (c *GenerateCommand) Run(ctx context.Context, parsedLayers *glazed_layers.ParsedLayers) error {
    // Option A: Use appconfig.Parser in Run()
    parser, err := buildParser(cmd, args) // cmd/args from cobra context
    settings, err := parser.Parse()
    
    // Option B: Keep existing parsedLayers, extract to settings struct
    settings := &GenerateAppSettings{}
    // ... extract from parsedLayers
    
    // Use settings...
}
```

**Decision Point:** Where should appconfig.Parser be used?
- **Option A:** In `InitGenerateCmd()` - requires cobra command available
- **Option B:** In `Run()` - parser created per execution
- **Option C:** Hybrid - parser created in `InitGenerateCmd()`, `Parse()` called in `Run()`

### Phase 3: Integration with Cobra Command Building

**File:** `glazed/pkg/cli/cobra.go`

**Changes:**

Add new builder function that accepts appconfig.Parser:

```go
func BuildCobraCommandWithParser[T any](
    s cmds.Command,
    parser *appconfig.Parser[T],
    opts ...CobraOption,
) (*cobra.Command, error) {
    // 1. Extract layers from parser registrations
    // 2. Build cobra command with those layers
    // 3. Wire parser.Parse() into Run() method
    // 4. Handle profile settings layer registration
}
```

**Alternative:** Modify existing `BuildCobraCommand()` to accept parser option:

```go
type CobraOption func(cfg *commandBuildConfig)

func WithAppConfigParser[T any](parser *appconfig.Parser[T]) CobraOption {
    // Store parser in config
    // Use in command building
}
```

## File-by-File Analysis

### Files to Modify

1. **`glazed/pkg/appconfig/options.go`**
   - Add `WithProfile()` function
   - Implement profile bootstrap logic
   - Add profile middleware builder

2. **`glazed/pkg/appconfig/parser.go`**
   - Possibly extend `Parser` to support profile bootstrap
   - May need two-phase parsing support

3. **`prescribe/cmd/prescribe/cmds/generate.go`**
   - Create `GenerateAppSettings` struct
   - Refactor `InitGenerateCmd()` to use appconfig
   - Refactor `Run()` to use parsed settings struct
   - Register all layers with parser

4. **`glazed/pkg/cli/cobra.go`** (Optional)
   - Add `BuildCobraCommandWithParser()` or `WithAppConfigParser()` option
   - Integrate parser with cobra command building

### Files to Reference

1. **`glazed/pkg/cmds/middlewares/profiles.go`**
   - `GatherFlagsFromProfiles()` - profile loading middleware
   - `GatherFlagsFromCustomProfiles()` - alternative profile loader
   - `ProfileConfig` - profile configuration struct

2. **`glazed/pkg/cli/cli.go`**
   - `ProfileSettings` struct
   - `NewProfileSettingsLayer()` function
   - `ProfileSettingsSlug` constant

3. **`geppetto/pkg/layers/layers.go`**
   - `GetGeppettoMiddlewares()` - example of profile bootstrap pattern
   - Shows two-phase parsing approach

4. **`glazed/pkg/cli/cobra-parser.go`**
   - `CobraParserConfig` - current parser configuration
   - `ParseCommandSettingsLayer()` - bootstrap parsing example

## Pseudocode Implementation

### Complete Flow Pseudocode

```go
// Phase 1: Add WithProfile to appconfig
func WithProfile(appName string, defaultProfile string) ParserOption {
    return func(o *parserOptions) error {
        // Resolve default profile file
        xdgConfigPath, _ := os.UserConfigDir()
        defaultProfileFile := filepath.Join(xdgConfigPath, appName, "profiles.yaml")
        
        // Add profile middleware that bootstraps ProfileSettings internally
        o.middlewares = append(o.middlewares,
            buildBootstrapProfileMiddleware(appName, defaultProfileFile, defaultProfile),
        )
        
        return nil
    }
}

func buildBootstrapProfileMiddleware(appName, defaultProfileFile, defaultProfile string) cmd_middlewares.Middleware {
    return func(next cmd_middlewares.HandlerFunc) cmd_middlewares.HandlerFunc {
        return func(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
            // 1. Bootstrap parse ProfileSettings
            profileLayer, _ := cli.NewProfileSettingsLayer()
            bootstrapLayers := layers.NewParameterLayers(layers.WithLayers(profileLayer))
            bootstrapParsed := layers.NewParsedLayers()
            
            // Parse from current parsedLayers (already has defaults/env/cobra)
            // Extract ProfileSettings
            profileSettings := &cli.ProfileSettings{}
            if err := parsedLayers.InitializeStruct(cli.ProfileSettingsSlug, profileSettings); err == nil {
                // Use resolved settings
                profileFile := profileSettings.ProfileFile
                if profileFile == "" {
                    profileFile = defaultProfileFile
                }
                profileName := profileSettings.Profile
                if profileName == "" {
                    profileName = defaultProfile
                }
                
                // 2. Load profile from file
                if profileMap, err := loadProfileFromFile(profileFile, profileName); err == nil && profileMap != nil {
                    // 3. Update parsedLayers with profile values
                    return updateFromMap(layers_, parsedLayers, profileMap,
                        parameters.WithParseStepSource("profiles"),
                    )
                }
            }
            
            return next(layers_, parsedLayers)
        }
    }
}

// Phase 2: Refactor generate.go
type GenerateAppSettings struct {
    ProfileSettings cli.ProfileSettings
    Repository      prescribe_layers.RepositorySettings
    Generation      prescribe_layers.GenerationSettings
    Extra           GenerateExtraSettings
    // Geppetto settings handled separately
}

func InitGenerateCmd() error {
    // Create layers
    profileLayer, _ := cli.NewProfileSettingsLayer()
    repoLayer, _ := prescribe_layers.NewRepositoryLayer()
    genLayer, _ := prescribe_layers.NewGenerationLayer()
    geppettoLayers, _ := geppettolayers.CreateGeppettoLayers()
    
    // Create parser (cobra will be added in Run)
    parser, err := appconfig.NewParser[GenerateAppSettings](
        appconfig.WithDefaults(),
        appconfig.WithProfile("pinocchio", "default"),
        appconfig.WithEnv("PRESCRIBE"),
    )
    
    // Register layers
    parser.Register("profile-settings", profileLayer, func(s *GenerateAppSettings) any {
        return &s.ProfileSettings
    })
    parser.Register("repository", repoLayer, func(s *GenerateAppSettings) any {
        return &s.Repository
    })
    parser.Register("generation", genLayer, func(s *GenerateAppSettings) any {
        return &s.Generation
    })
    // Geppetto layers registered separately or extracted manually
    
    // Store parser in command or global variable
    generateParser = parser
    
    // Build cobra command (needs integration)
    glazedCmd, _ := NewGenerateCommand()
    cobraCmd, err := cli.BuildCobraCommand(glazedCmd, ...)
    
    generateCmd = cobraCmd
    return nil
}

func (c *GenerateCommand) Run(ctx context.Context, parsedLayers *glazed_layers.ParsedLayers) error {
    // Get cobra command and args from context or closure
    cmd := getCobraCommand(ctx) // Need to pass this
    args := getArgs(ctx) // Need to pass this
    
    // Add cobra parsing to parser
    parserWithCobra := appconfig.NewParser[GenerateAppSettings](
        appconfig.WithDefaults(),
        appconfig.WithProfile("pinocchio", "default"),
        appconfig.WithEnv("PRESCRIBE"),
        appconfig.WithCobra(cmd, args),
    )
    // Re-register layers...
    
    // Parse settings
    settings, err := parserWithCobra.Parse()
    
    // Use settings...
    ctrl, err := helpers.NewInitializedControllerFromParsedLayers(parsedLayers)
    // ... rest of Run() logic
}
```

## Key Challenges and Solutions

### Challenge 1: Profile Bootstrap Timing

**Problem:** Profile settings need to be parsed from cobra/env/config before profile can be loaded, but `WithCobra()` is typically called after profile middleware.

**Solution:** Profile middleware bootstraps ProfileSettings internally from already-parsed layers, then loads profile.

### Challenge 2: Cobra Command Availability

**Problem:** `appconfig.Parser` needs cobra command for `WithCobra()`, but command is built in `InitGenerateCmd()`.

**Solution Options:**
- **A:** Parse in `Run()` method where cmd/args are available
- **B:** Two-phase parsing: bootstrap in `InitGenerateCmd()`, full parse in `Run()`
- **C:** Store parser, add cobra option in `Run()`

**Recommendation:** Option C - store parser, add cobra in `Run()`.

### Challenge 3: Geppetto Layers Complexity

**Problem:** Geppetto layers are complex and may not map cleanly to struct fields.

**Solution:** Extract Geppetto settings manually from `parsedLayers` using existing `gepsettings.NewStepSettingsFromParsedLayers()`.

### Challenge 4: Backward Compatibility

**Problem:** Existing code uses `parsedLayers.InitializeStruct()` directly.

**Solution:** Keep `parsedLayers` available, extract to settings struct for new code paths. Gradually migrate.

## Testing Strategy

1. **Unit Tests:**
   - Test `WithProfile()` option behavior
   - Test profile bootstrap middleware
   - Test precedence ordering

2. **Integration Tests:**
   - Test profile loading from `~/.config/pinocchio/profiles.yaml`
   - Test `--profile` and `--profile-file` flags
   - Test environment variable overrides
   - Test precedence: defaults < profiles < config < env < flags

3. **End-to-End Tests:**
   - Test `prescribe generate` with profile
   - Test profile values override defaults
   - Test flags override profile values

## Migration Path

1. **Phase 1:** Add `WithProfile()` to appconfig (non-breaking)
2. **Phase 2:** Refactor `generate.go` to use appconfig (keep old path working)
3. **Phase 3:** Test thoroughly
4. **Phase 4:** Remove old parsing code
5. **Phase 5:** Apply pattern to other commands

## References

- `glazed/pkg/appconfig/` - appconfig package implementation
- `glazed/pkg/cmds/middlewares/profiles.go` - profile middleware
- `geppetto/pkg/layers/layers.go` - example profile bootstrap pattern
- `glazed/pkg/cli/cli.go` - ProfileSettings layer
- `glazed/pkg/doc/topics/15-profiles.md` - profile documentation
