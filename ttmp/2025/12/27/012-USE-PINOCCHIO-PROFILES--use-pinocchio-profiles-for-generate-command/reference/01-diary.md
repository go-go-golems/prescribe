---
Title: Diary
Ticket: 012-USE-PINOCCHIO-PROFILES
Status: active
Topics:
    - configuration
    - profiles
    - appconfig
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-27T20:54:08.169281838-05:00
WhatFor: ""
WhenToUse: ""
---

# Diary

## Goal

Document the step-by-step implementation of integrating Pinocchio profiles (`~/.config/pinocchio/profiles.yaml`) into the `prescribe generate` command using the `appconfig` package. This includes refactoring the command initialization to use `appconfig.Parser` and adding profile support to the appconfig package itself.

## Step 1: Initial Research and Analysis

This step involved understanding the current architecture, the profile system, and the appconfig package to determine the integration approach.

### What I did
- Read `prescribe/cmd/prescribe/cmds/generate.go` to understand current implementation
- Explored `glazed/pkg/appconfig/` package (parser.go, options.go, doc.go)
- Searched codebase for profile loading patterns (`GatherFlagsFromProfiles`, `ProfileSettings`)
- Reviewed `geppetto/pkg/layers/layers.go` for bootstrap pattern example
- Read `glazed/pkg/cmds/middlewares/profiles.go` for profile middleware implementation
- Examined `glazed/pkg/cli/cobra-parser.go` for cobra integration patterns
- Created ticket `012-USE-PINOCCHIO-PROFILES` and analysis document

### Why
- Need to understand how profiles work before implementing
- Need to understand appconfig API to use it correctly
- Need to identify integration points and challenges
- Need to document findings for implementation planning

### What worked
- Found clear examples of profile bootstrap pattern in `geppetto/pkg/layers/layers.go`
- Identified `ProfileSettings` layer in `glazed/pkg/cli/cli.go`
- Understood middleware precedence model (defaults < profiles < config < env < flags)
- Found `GatherFlagsFromProfiles` middleware that handles profile loading
- Identified that appconfig currently lacks `WithProfile()` option

### What didn't work
- Initial assumption that appconfig had built-in profile support was incorrect
- Need to add profile support to appconfig first

### What I learned
- Profile loading requires bootstrap parsing: ProfileSettings must be parsed from defaults/config/env/flags before profile file can be loaded
- appconfig.Parser uses middleware chain with precedence ordering
- Profile middleware (`GatherFlagsFromProfiles`) takes defaultProfileFile, profileFile, and profile name
- Current generate.go uses `CobraCommandDefaultMiddlewares` which doesn't include profile support
- Geppetto layers are complex and may need special handling

### What was tricky to build
- Understanding the bootstrap pattern: ProfileSettings needs to be parsed before profile middleware can run, but profile middleware is part of the main parsing chain
- Determining where appconfig.Parser should be instantiated: InitGenerateCmd() vs Run() method
- Figuring out how to integrate cobra command building with appconfig.Parser

### What warrants a second pair of eyes
- The bootstrap pattern for ProfileSettings - ensure we're handling precedence correctly
- The decision on where to instantiate appconfig.Parser (InitGenerateCmd vs Run)
- The approach for handling Geppetto layers which may not map cleanly to struct fields

### What should be done in the future
- Add `WithProfile()` option to appconfig package
- Consider adding helper function `WithPinocchioProfile()` for common use case
- Document profile bootstrap pattern for other commands
- Consider adding `BuildCobraCommandWithParser()` helper in cli package

### Code review instructions
- Review analysis document: `analysis/01-analysis-integrating-pinocchio-profiles-with-generate-command-using-appconfig.md`
- Check understanding of bootstrap pattern in `geppetto/pkg/layers/layers.go:181-271`
- Verify profile middleware behavior in `glazed/pkg/cmds/middlewares/profiles.go:13-75`

### Technical details

**Key Files Explored:**
- `prescribe/cmd/prescribe/cmds/generate.go` - Current command implementation
- `glazed/pkg/appconfig/parser.go` - Parser implementation
- `glazed/pkg/appconfig/options.go` - Parser options
- `glazed/pkg/cmds/middlewares/profiles.go` - Profile middleware
- `glazed/pkg/cli/cli.go` - ProfileSettings layer
- `geppetto/pkg/layers/layers.go` - Bootstrap pattern example

**Key Functions Identified:**
- `appconfig.NewParser[T](options...)` - Creates parser
- `parser.Register(slug, layer, binder)` - Registers layer
- `parser.Parse()` - Executes middleware chain, returns T
- `cli.NewProfileSettingsLayer()` - Creates profile settings layer
- `middlewares.GatherFlagsFromProfiles(defaultFile, file, profile)` - Profile middleware
- `cli.BuildCobraCommand(cmd, opts...)` - Builds cobra command

**Profile Bootstrap Pattern (from geppetto):**
```go
// 1. Bootstrap parse ProfileSettings
profileSettings := &cli.ProfileSettings{}
bootstrapProfileLayers := layers.NewParameterLayers(layers.WithLayers(profileSettingsLayer))
bootstrapProfileParsed := layers.NewParsedLayers()
middlewares.ExecuteMiddlewares(
    bootstrapProfileLayers,
    bootstrapProfileParsed,
    ParseFromCobraCommand(cmd),
    UpdateFromEnv("PINOCCHIO"),
    SetFromDefaults(),
)
bootstrapProfileParsed.InitializeStruct(cli.ProfileSettingsSlug, profileSettings)

// 2. Resolve profile file and name
defaultProfileFile := fmt.Sprintf("%s/pinocchio/profiles.yaml", xdgConfigPath)
if profileSettings.ProfileFile == "" {
    profileSettings.ProfileFile = defaultProfileFile
}
if profileSettings.Profile == "" {
    profileSettings.Profile = "default"
}

// 3. Add profile middleware to main chain
middlewares_ = append(middlewares_,
    middlewares.GatherFlagsFromProfiles(
        defaultProfileFile,
        profileSettings.ProfileFile,
        profileSettings.Profile,
        parameters.WithParseStepSource("profiles"),
    ),
)
```

### What I'd do differently next time
- Start with reading the profile documentation (`glazed/pkg/doc/topics/15-profiles.md`) earlier
- Look for existing examples of appconfig usage in the codebase
- Check if there are any open issues or TODOs related to profile support in appconfig

## Step 2: Deep Dive into Bootstrap Pattern and appconfig Examples

This step involved examining the bootstrap pattern more closely and understanding how appconfig is actually used in practice.

### What I did
- Read `geppetto/pkg/layers/layers.go:140-293` in detail to understand bootstrap pattern
- Read `glazed/cmd/examples/appconfig-parser/main.go` to see appconfig usage example
- Reviewed profile documentation (`glazed/pkg/doc/topics/15-profiles.md`)
- Analyzed how `CobraParserConfig` handles profile settings layer

### Why
- Need to understand the exact bootstrap sequence for ProfileSettings
- Need to see how appconfig.Parser is used with cobra commands
- Need to understand the precedence ordering and middleware execution order

### What worked
- Found clear bootstrap pattern: parse ProfileSettings first, then use resolved values for profile middleware
- Understood that middleware execution order is reversed: last middleware in slice executes first (highest precedence)
- Confirmed that ProfileSettings layer must be registered for flags to exist
- Found that `CobraParserConfig` has `EnableProfileSettingsLayer` flag but doesn't handle profile loading automatically

### What didn't work
- appconfig.Parser doesn't have built-in profile support - need to add it
- Current `CobraParserConfig` doesn't integrate profile loading into middleware chain

### What I learned
- Bootstrap pattern requires two-phase parsing:
  1. Parse ProfileSettings from defaults + config + env + cobra
  2. Use resolved ProfileSettings to load profile middleware
  3. Add profile middleware to main chain at correct precedence (after defaults, before config/env/flags)
- Middleware execution is reversed: slice order is low→high precedence, but execution is high→low
- Profile middleware should be placed AFTER config middleware in slice (so config executes first, then profiles)
- But profiles should override config, so profiles middleware must call `next()` first, then update parsedLayers
- `GatherFlagsFromProfiles` takes three parameters: defaultProfileFile, profileFile, profileName
- Default profile file for pinocchio: `~/.config/pinocchio/profiles.yaml`
- Default profile name: "default"

### What was tricky to build
- Understanding middleware execution order vs slice ordering
- Figuring out how to bootstrap ProfileSettings within appconfig.Parser when cobra isn't available yet
- Determining where profile middleware should be placed in the middleware chain

### What warrants a second pair of eyes
- The middleware execution order semantics - ensure we understand the reversal correctly
- The bootstrap approach within appconfig.Parser - is two-phase parsing the right approach?
- The integration point between appconfig.Parser and cobra command building

### What should be done in the future
- Add `WithProfile()` option to appconfig that handles bootstrap internally
- Consider adding `WithPinocchioProfile()` convenience function
- Document the profile bootstrap pattern for appconfig users
- Consider adding profile support to `CobraParserConfig` middleware builder

### Code review instructions
- Review bootstrap pattern: `geppetto/pkg/layers/layers.go:181-271`
- Check middleware execution order: `glazed/pkg/appconfig/parser.go:96-101`
- Verify profile middleware behavior: `glazed/pkg/cmds/middlewares/profiles.go:13-75`

### Technical details

**Bootstrap Pattern Sequence (from geppetto):**
1. Parse CommandSettings (for config file resolution)
2. Resolve config files (low → high precedence)
3. Bootstrap parse ProfileSettings:
   - Create ProfileSettingsLayer
   - Execute middlewares: cobra → env → config → defaults
   - Extract ProfileSettings struct
4. Resolve profile file and name (with defaults)
5. Add profile middleware to main chain:
   - Place AFTER config middleware in slice
   - But profile middleware calls next() first, then updates
   - So execution order ensures: defaults → profiles → config → env → flags

**Middleware Execution Order:**
- Options are applied in low→high precedence order
- Middlewares are collected in same order
- But ExecuteMiddlewares reverses the order (high→low precedence)
- So: last option = highest precedence middleware = executes first

**Profile Middleware Placement:**
- In slice: [flags, args, env, config, profiles, defaults] (low→high precedence)
- Execution: defaults → profiles → config → env → args → flags (low→high precedence)
- Profile middleware calls next() first, then updates parsedLayers
- This ensures profiles override defaults, but config/env/flags override profiles

**appconfig.Parser Usage Pattern:**
```go
parser, err := appconfig.NewParser[AppSettings](
    appconfig.WithDefaults(),        // Lowest precedence
    appconfig.WithConfigFiles(...),  // Config files
    appconfig.WithEnv("APP"),        // Environment
    appconfig.WithCobra(cmd, args),  // Highest precedence
)
// Register layers
parser.Register("layer-slug", layer, binder)
// Parse
settings, err := parser.Parse()
```

**Key Insight:** appconfig.Parser needs cobra command/args for `WithCobra()`, but command is built in `InitGenerateCmd()`. This suggests we need to either:
- Parse in `Run()` method where cmd/args are available
- Use two-phase approach: bootstrap in InitGenerateCmd(), full parse in Run()
- Store parser, add cobra option dynamically in Run()

### What I'd do differently next time
- Read the middleware execution order documentation more carefully upfront
- Look at the actual middleware implementations to understand the next() pattern
- Check if there are helper functions for profile bootstrap in glazed

## Step 3: Implement `appconfig.WithProfile` + add tests + add a tiny example program

This step made profile selection “circularity-safe” for `appconfig.Parser`: we can now resolve `profile-settings.profile` and `profile-settings.profile-file` via a bootstrap pre-parse (config/env/cobra + defaults), then apply `profiles.yaml` at the correct precedence level. I also added a small Glazed example program so we can debug the behavior without involving the large `prescribe` repo, plus unit tests that lock down selection + precedence + error behavior.

**Commit (code):** 9c37e8d950654e410c4b0fac2c0e02a8a7ad50ba — "feat(appconfig): add WithProfile bootstrap for profiles.yaml"

### What I did
- Implemented `appconfig.WithProfile(appName, ...)` in `glazed/pkg/appconfig/options.go`
  - Bootstraps `profile-settings` selection via a mini middleware chain
  - Applies the selected profile using `middlewares.GatherFlagsFromProfiles(...)` with parse-step metadata
- Fixed default env-prefix behavior for selection:
  - defaults to `strings.ToUpper(appName)` (so `WithProfile(\"pinocchio\")` naturally uses `PINOCCHIO_PROFILE*`)
- Added unit tests in `glazed/pkg/appconfig/profile_test.go`:
  - selection from env/config/cobra
  - precedence (flags > env > config > profiles > defaults)
  - missing-file error behavior
- Added a small example program: `glazed/cmd/examples/appconfig-profiles/main.go`
  - Creates a temp `profiles.yaml` (self-contained)
  - Shows how to drive selection via env + `--profile`
- Ran formatting + tests:

```bash
cd /home/manuel/workspaces/2025-12-26/prescribe-import && gofmt -w glazed/pkg/appconfig/options.go glazed/pkg/appconfig/profile_test.go glazed/cmd/examples/appconfig-profiles/main.go
cd /home/manuel/workspaces/2025-12-26/prescribe-import/glazed && go test ./... -count=1
```

### Why
- `GatherFlagsFromProfiles(...)` needs constructor arguments, so profile selection must be resolved *before* profile middleware is instantiated (bootstrap pre-parse pattern).
- The example program makes it easy to debug and iterate without having to run `prescribe generate` on a large repo.

### What worked
- `go test ./...` in the `glazed` module is green, and the new example compiles.
- The tests confirm the expected precedence and failure modes.

### What didn't work
- I initially thought git commits were blocked because `.git` isn’t a directory in a worktree. In reality, `glazed/.git` is a **file** pointing at the worktree gitdir, so committing from `glazed/` works fine.

### What I learned
- Defaulting profile selection env vars to `strings.ToUpper(appName)` is the most predictable behavior for apps that want to load `~/.config/<appName>/profiles.yaml` and select via `<APP>_PROFILE*`.
- A tiny example binary is a better debugging surface than attempting to reproduce via big real repos early.

### What was tricky to build
- Getting the bootstrap selection to use the right “selection env prefix” without accidentally coupling it to whatever env prefix the app uses for non-profile settings.
- Ensuring the profile application happens at the right precedence layer (after defaults, before config/env/flags).

### What warrants a second pair of eyes
- Confirm the intended default env prefix for selection should be `strings.ToUpper(appName)` (not “last `WithEnv` prefix”), especially for the Pinocchio use-case.
- Confirm the error semantics we inherit from `GatherFlagsFromProfiles` match expectations for non-default profiles.

### What should be done in the future
- Once we have an actual git checkout, commit the changes and add the commit hashes here.
- Add an integration script (small repo via `prescribe/test-scripts/setup-test-repo.sh`) once `prescribe generate` is migrated to use `appconfig.WithProfile`.

## Step 4: Tighten `create-pull-request` prompt contract and avoid empty description markers

This step tightened the default `create-pull-request` prompt so it doesn’t emit a dangling “The description of the pull request is: .” line when no note-based description exists. While doing that, I also made the “YAML-only output” contract more explicit (no markdown / no code fences) and added a unit test to lock the behavior down.

**Commit (code):** fbfb180f08081d1b7ee423e24d5b5793b009face — "prompt: tighten create-pull-request YAML contract and omit empty description"

### What I did
- Updated `internal/prompts/assets/create-pull-request.yaml`:
  - Added explicit “output only YAML” framing
  - Wrapped the description section in `{{ if .description }}` and rendered it as a block when present
  - Removed the example’s triple-backtick code fence and explicitly disallowed fences
- Added a unit test in `internal/api/prompt_test.go` proving that when no note-based description exists, we do not render any empty description marker.
- Ran formatting + tests:

```bash
cd /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe && gofmt -w internal/api/prompt_test.go
cd /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe && go test ./... -count=1
```

### Why
- In prescribe, `.description` is currently derived from note context entries; when there are no notes, the template would still render a sentence with an empty placeholder, which is confusing and can degrade model behavior.
- A stronger “YAML-only” contract reduces the chance of non-parseable or multi-part assistant output.

### What worked
- The prompt still renders correctly for the “notes present” case, and now cleanly omits the description section when empty.
- The new unit test guards against regressions.

### What didn't work
- I initially tried committing from the workspace root and got `fatal: not a git repository`. This repo is split: the git repo for this change lives under `prescribe/`, so commits must be executed there.

### What I learned
- The “Pinocchio-style” variable mapping in `internal/api/prompt.go` uses `AdditionalContext` notes as the `.description` template var, so templates must treat `.description` as optional.

### What was tricky to build
- Avoiding template whitespace/formatting issues while still keeping the rendered prompt readable and stable for tests.

### What warrants a second pair of eyes
- Whether we want to be strict about “no code fences” in the prompt contract, given the parser can handle fenced YAML; confirm this aligns with desired model behavior across providers.

### What should be done in the future
- Consider renaming `.description` to something like `.notes` in our variable mapping (and updating the template accordingly) to avoid semantic confusion between “PR description” vs “user notes”.

### Code review instructions
- Start with `internal/prompts/assets/create-pull-request.yaml` (look for the conditional description block and YAML-only output rules).
- Then review `internal/api/prompt_test.go` for the regression test.

## Step 5: Persist PR title/description in session.yaml and plumb into prompt rendering

This step added explicit PR `title` and `description` fields to the core domain model and to `session.yaml`, then threaded them through the controller request builder and the prompt template variables. The practical outcome: `prescribe generate` can now render `.title` / `.description` even when the description isn’t coming from “note” context.

**Commit (code):** da26af267e4f0687adef932fb7e1ad99c9b9e0a7 — "feat: plumb PR title/description into session and prompt vars"

### What I did
- Added `Title`/`Description` to `internal/domain.PRData`
- Persisted them in `internal/session.Session` (`title`, `description`)
- Extended `internal/api.GenerateDescriptionRequest` and `Controller.BuildGenerateDescriptionRequest`
- Updated `internal/api/prompt.go` variable mapping:
  - `.title` now comes from `req.Title`
  - `.description` now comes from `req.Description` + appended note-context (if present)
- Added/updated unit tests for request building and prompt rendering
- Ran formatting + tests:

```bash
cd /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe && gofmt -w internal/domain/domain.go internal/session/session.go internal/api/api.go internal/api/prompt.go internal/api/prompt_test.go internal/controller/controller.go internal/controller/controller_test.go
cd /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe && go test ./... -count=1
```

### Why
- We want PR title/description to be first-class session state (persisted and controllable), not an accidental byproduct of “note” context concatenation.

### What warrants a second pair of eyes
- The decision to append note-context onto `req.Description` for `.description` (keeps backward-compat, but mixes semantics).

### Code review instructions
- Start at `internal/domain/domain.go`, then `internal/session/session.go`, then `internal/api/prompt.go`.
- Validate with `go test ./... -count=1`.

## Step 6: Add `generate --title/--description` and `session init --title/--description`

This step added CLI flags so users can override the session’s title/description at generation time, and set/persist them when initializing a session. I also surfaced the values in `session show` (as `title` and `description_preview`) for quick inspection.

**Commit (code):** 46a2c0f0018399a8e53e66c4e4196ae42a908117 — "feat: add --title/--description flags for generate and session init"

### What I did
- Extended `pkg/layers.GenerationSettings` with `title` and `description`
- In `generate`, applied the flags as overrides after loading session.yaml
- In `session init`, added flags and persisted them when `--save` is used
- In `session show`, added `title` + `description_preview` to the output row
- Ran formatting + tests:

```bash
cd /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe && gofmt -w pkg/layers/generation.go cmd/prescribe/cmds/generate.go cmd/prescribe/cmds/session/init.go cmd/prescribe/cmds/session/show.go
cd /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe && go test ./... -count=1
```

### Why
- These fields need to be easy to set without manually editing session.yaml, and `generate` needs a clean “flags override session” contract.

### What warrants a second pair of eyes
- Flag naming/UX and whether we want short flags (`-t`, etc.) given potential collisions across commands.

### Code review instructions
- Start at `pkg/layers/generation.go` to see the flag definitions, then review `cmd/prescribe/cmds/generate.go` and `cmd/prescribe/cmds/session/init.go`.

## Step 7: Backfill smoke tests for small-repo scripts (and fix TEST_REPO_DIR propagation)

This step started as a “quick confidence check” after adding title/description plumbing: I wrote a small-repo smoke test that initializes a session with `--title/--description`, then exports the rendered payload and asserts those strings are present. The first run failed in a surprising way (controller couldn’t initialize git), which turned out to be a script wiring bug: we weren’t propagating `TEST_REPO_DIR` into the shared `setup-test-repo.sh` helper.

### What I did
- Added a new ticket smoke test:
  - `scripts/03-smoke-test-prescribe-generate-title-description.sh`
  - Flow: setup tiny repo → `session init --save --title/--description` → `generate --export-rendered` → `grep` asserts.
- Ran the smoke test and captured the failure:

```bash
bash /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/ttmp/2025/12/27/012-USE-PINOCCHIO-PROFILES--use-pinocchio-profiles-for-generate-command/scripts/03-smoke-test-prescribe-generate-title-description.sh
```

### What didn't work (exact error)
- The helper created the repo at `/tmp/prescribe-test-repo`, but `prescribe` was invoked with `--repo /tmp/prescribe-generate-title-desc-test-repo`, so controller init failed:
  - `Error: failed to create controller: failed to initialize git service: not a git repository: /tmp/prescribe-generate-title-desc-test-repo`

### Why it happened
- `test-scripts/setup-test-repo.sh` reads `TEST_REPO_DIR` from the environment.
- The smoke scripts were calling the helper without exporting `TEST_REPO_DIR`, so it always defaulted to `/tmp/prescribe-test-repo`, drifting from the script’s own `TEST_REPO_DIR`.

### What I changed
- Fixed `scripts/03-smoke-test-prescribe-generate-title-description.sh` to call:
  - `env TEST_REPO_DIR="$TEST_REPO_DIR" bash "$PRESCRIBE_ROOT/test-scripts/setup-test-repo.sh"`
- Also fixed the same bug in the earlier profiles smoke test `scripts/01-smoke-test-prescribe-generate-profiles.sh` for consistency.

### What worked
- After the fix, the title/description smoke test passes and confirms the rendered payload contains both strings:

```bash
bash /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/ttmp/2025/12/27/012-USE-PINOCCHIO-PROFILES--use-pinocchio-profiles-for-generate-command/scripts/03-smoke-test-prescribe-generate-title-description.sh
```

### Commit (code)
- **c4f7a31c3278d7bae1ec4dcfdd2daa1599309fbf** — "test(012): add title/description smoke test and fix TEST_REPO_DIR propagation"

### What warrants a second pair of eyes
- Sanity check that all ticket smoke scripts consistently propagate env vars into shared helpers (to avoid “works on my machine” / wrong repo path).

### What should be done in the future
- Consider making `setup-test-repo.sh` print (or export) the actual repo path it created in a machine-readable way, to remove this entire class of mismatch.

## Step 8: Make `WithProfile` bootstrap selection respect configured sources (don’t consult env unless enabled)

This step fixed an important subtlety in `appconfig.WithProfile`: the bootstrap pre-parse of `profile-settings` was always consulting environment variables (defaulting to `strings.ToUpper(appName)`), even if the parser never enabled env parsing. That made profile selection “leak in” from env unexpectedly. I changed it so env is only consulted when env parsing is configured on the parser (or explicitly requested via `WithProfileEnvPrefix`), and added a unit test that would have caught the bug.

**Commit (code):** 15c63ab2816e06f4711045d3bf9408be7c3dff29 — "fix(appconfig): only consult env for profile selection when enabled"

### What I did
- Updated `glazed/pkg/appconfig/options.go`:
  - Gate bootstrap `UpdateFromEnv(...)` behind `envEnabled := len(o.envPrefixes) > 0 || pcfg.envPrefix != ""`
- Updated tests in `glazed/pkg/appconfig/profile_test.go`:
  - Adjusted env-selection tests to enable env parsing (`WithEnv("MYAPP")`)
  - Added `TestWithProfile_ProfileSelection_FromEnv_requiresEnvEnabled`
- Ran formatting + tests:

```bash
cd /home/manuel/workspaces/2025-12-26/prescribe-import/glazed && gofmt -w pkg/appconfig/options.go pkg/appconfig/profile_test.go
cd /home/manuel/workspaces/2025-12-26/prescribe-import/glazed && go test ./... -count=1
```

### Why
- Profile selection should not silently depend on env unless the caller explicitly enabled env parsing (principle of least surprise / “only configured sources”).

### What warrants a second pair of eyes
- Confirm the intended semantics: env selection should be enabled when any env parsing is configured for the parser (even if the main env prefix differs from the profile-selection prefix).

### Code review instructions
- Review `glazed/pkg/appconfig/options.go` inside `WithProfile`, then check `glazed/pkg/appconfig/profile_test.go` for the new regression test.

## Step 9: Robustness for prose-wrapped YAML outputs (heuristic salvage)

This step improves resilience when a model (often Gemini-style) wraps the YAML in prose like “Here is the YAML: …”. We already prefer the last fenced ```yaml``` block when available, but when there are no fences and YAML is preceded by prose, a strict YAML unmarshal fails. I added a conservative salvage path: if parsing fails, we look for the last `title:` block and attempt to parse YAML from there.

**Commit (code):** b7e89b03371f76db58e0915dd41966eecd04eb1a — "fix(api): salvage YAML blocks from prose-wrapped outputs"

### What I did
- Updated `prescribe/internal/api/prdata_parse.go`:
  - After the fence-stripping fallback fails, try `trySalvageYAMLFromTitleBlock` (regex `(?m)^[ \\t]*title:`) and re-parse from that point.
- Added a unit test in `prescribe/internal/api/prdata_parse_test.go` covering “Sure — here is the YAML:” + YAML body.
- Ran formatting + tests:

```bash
cd /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe && gofmt -w internal/api/prdata_parse.go internal/api/prdata_parse_test.go
cd /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe && go test ./... -count=1
```

### Why
- We want the structured `GeneratedPRData` view to work even when the model doesn’t follow the YAML-only contract perfectly.

### What warrants a second pair of eyes
- Confirm the heuristic (last `title:` block) is conservative enough and won’t accidentally parse unrelated YAML-ish snippets in long outputs.
