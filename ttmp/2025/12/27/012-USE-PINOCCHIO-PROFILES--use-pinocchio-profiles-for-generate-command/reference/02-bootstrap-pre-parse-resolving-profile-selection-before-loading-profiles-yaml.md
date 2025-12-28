---
Title: 'Bootstrap pre-parse: resolving profile selection before loading profiles.yaml'
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
LastUpdated: 2025-12-27T21:04:47.175298322-05:00
WhatFor: ""
WhenToUse: ""
---

## Bootstrap pre-parse: resolving profile selection before loading `profiles.yaml`

This document explains the “bootstrap pre-parse” pattern we use to make profiles work *predictably* when the profile selection itself can be configured via **config files**, **environment variables**, and **CLI flags**. The core idea is simple: before we can load a profile from `profiles.yaml`, we must first decide *which* profile to load and *which file* to load it from — and those decisions must follow the same precedence rules as everything else.

In this ticket, we’re adding this behavior to `glazed/pkg/appconfig` as `appconfig.WithProfile(...)`, so any app using `appconfig.Parser` can opt into profile loading without copy/pasting the Geppetto bootstrap logic.

### Goal

Make `profiles.yaml` loading **circularity-safe** and **debuggable**:

- **Circularity-safe**: `profile-settings.profile` / `profile-settings.profile-file` can come from config/env/flags, and we still load the correct profile.
- **Correct precedence**: \( \text{defaults} < \text{profiles} < \text{config} < \text{env} < \text{flags} \).
- **Debuggable**: profile application shows up as a parse step (source = `profiles`, metadata includes `profile` and `profileFile`).

### Context: the “why” (what breaks without bootstrap)

The profile loading middleware is `middlewares.GatherFlagsFromProfiles(defaultProfileFile, profileFile, profileName, ...)` in `glazed/pkg/cmds/middlewares/profiles.go`.

The critical detail: it needs **constructor arguments** (`profileFile`, `profileName`). If we compute those early (using defaults) and then later apply env/config/flags, we’ve already made the decision and will load the wrong profile.

This is the “profile selection circularity” mentioned in `glazed/pkg/doc/topics/15-profiles.md`:

- We want config/env/flags to influence `profile-settings.*`.
- But the profile middleware must be instantiated using the resolved `profile-settings.*`.
- Therefore we need a **bootstrap pre-parse** that resolves `profile-settings.*` first.

### Mental model

Think of it as a two-phase parse:

1. **Bootstrap parse (tiny, focused)**: parse only the `profile-settings` layer to determine `(profileName, profileFile)`.
2. **Main parse (full settings)**: parse all layers, and insert a “load profile values” step between defaults and higher-precedence sources.

### Quick Reference (what we will do in `appconfig.WithProfile`)

#### Inputs we care about

- **Where profile selection can come from** (selection stage):
  - Config files: `profile-settings: { profile, profile-file }`
  - Env vars: `<ENV_PREFIX>_PROFILE`, `<ENV_PREFIX>_PROFILE_FILE`
  - Cobra flags: `--profile`, `--profile-file` (only if those flags exist on the cobra command)
  - Defaults: `profile="default"`, `profileFile=~/.config/<appName>/profiles.yaml` (XDG via `os.UserConfigDir()`)

- **Where profile values come from** (application stage):
  - `profiles.yaml` file (map: profile → layer → param → value)
  - Applied as parse step source `profiles` and metadata `{profile, profileFile}`

#### Bootstrap pre-parse chain (selection stage)

In `appconfig.WithProfile`, we construct a mini middleware list for **only** the `profile-settings` layer:

- **Cobra** (if `WithCobra(cmd,args)` was configured): `ParseFromCobraCommand(cmd)` (source = `cobra`)
- **Env** (from a prefix): `UpdateFromEnv(prefix)` (source = `env`)
- **Config files** (if `WithConfigFiles(...)` was configured): `LoadParametersFromFiles(files...)` (source = `config`)
- **Defaults**: `SetFromDefaults()` (source = `defaults`)

Then we hydrate:

- `cli.ProfileSettings` via `InitializeStruct(cli.ProfileSettingsSlug, &ps)`

And resolve final selection:

- `profileName = ps.Profile` or default (`"default"`)
- `profileFile = ps.ProfileFile` or default profile file

#### Main parse integration (application stage)

`WithProfile` itself is implemented as a middleware inserted into the `appconfig.Parser` middleware chain. It does:

1. `next(...)` to run **lower precedence** sources first (typically defaults and provided-values).
2. Apply the selected profile values using `GatherFlagsFromProfiles(...)`, updating `parsedLayers`.
3. Return control to the remaining higher-precedence middlewares (config/env/cobra) that will override profile values as needed.

This results in:

\[
\text{defaults} \;\rightarrow\; \text{profiles} \;\rightarrow\; \text{config} \;\rightarrow\; \text{env} \;\rightarrow\; \text{flags}
\]

### How this maps to code (what exists today)

#### Reference implementation example (Geppetto)

Geppetto already does this manually in `geppetto/pkg/layers/layers.go` (see the “Option A bootstrap parse” block). The key steps are:

- Bootstrap parse `command-settings` (for config discovery)
- Resolve config file list
- Bootstrap parse `profile-settings`
- Instantiate `GatherFlagsFromProfiles(defaultProfileFile, resolvedProfileFile, resolvedProfileName, ...)`
- Build main chain with the profile step at the correct precedence

#### Our `appconfig` plan (what we actually implement)

In `glazed/pkg/appconfig/options.go`, we implement:

- `WithProfile(appName string, opts ...ProfileOption) ParserOption`

To make bootstrap possible inside `appconfig`, we also record “what sources were configured” when other options are applied:

- `WithEnv(prefix)` records `prefix` in internal bookkeeping
- `WithConfigFiles(files...)` records `files`
- `WithCobra(cmd,args)` records `(cmd,args)`

Then the `WithProfile` middleware uses those recorded values to bootstrap-parse selection.

### Practical pseudocode (what happens at runtime)

This pseudocode is intentionally “written out” to show ordering and intent.

```go
// During parser construction:
o.envPrefixes += WithEnv(...)
o.configFiles += WithConfigFiles(...)
o.cobraCmd/Args = WithCobra(...)
o.middlewares += WithProfile(appName)

// During parsing (ExecuteMiddlewares):
WithProfileMiddleware(next):
    // phase 1: bootstrap selection
    ps = bootstrapParseProfileSettings(
        cobra=o.cobraCmd,
        envPrefix=chosenPrefix,
        configFiles=o.configFiles,
        defaults=true,
    )
    (profileName, profileFile) = resolve(ps, defaultProfileName, defaultProfileFile)

    // phase 2: run lower-precedence chain first
    next(layers, parsedLayers)

    // phase 3: apply profile values (source="profiles", metadata includes selection)
    GatherFlagsFromProfiles(defaultProfileFile, profileFile, profileName)(noOp)(layers, parsedLayers)

    return nil
```

### Why the precedence works (and the subtle `next()` detail)

Many Glazed “value-setting” middlewares use this pattern:

- call `next(...)` first
- then update `parsedLayers`

`GatherFlagsFromProfiles` works exactly like that (`glazed/pkg/cmds/middlewares/profiles.go`), which is why it can be inserted at the “profiles precedence” layer while still allowing later config/env/flags to override it.

### Usage examples (how an app would use it)

#### Example: basic appconfig parser with profiles + config + env + cobra

```go
parser, err := appconfig.NewParser[MySettings](
    appconfig.WithDefaults(),
    appconfig.WithProfile("pinocchio"),
    appconfig.WithConfigFiles("/etc/myapp/config.yaml"),
    appconfig.WithEnv("MYAPP"),
    appconfig.WithCobra(cmd, args),
)
```

#### Example: selecting profile via config file

`config.yaml`:

```yaml
profile-settings:
  profile: dev
  profile-file: /home/me/.config/pinocchio/profiles.yaml
```

This will:

- bootstrap-parse `profile-settings` from config/env/flags
- choose `dev` + that file
- apply `dev` profile values to all layers

#### Example: selecting profile via env vars

```bash
export MYAPP_PROFILE=prod
export MYAPP_PROFILE_FILE=/etc/shared/profiles.yaml
```

Env will override config selection, but CLI flags will override env selection.

### Common failure modes (what we want to avoid)

- **Naive construction**: instantiate `GatherFlagsFromProfiles(...)` with defaults before parsing env/config. This silently loads the wrong profile.
- **Missing profile settings flags**: `--profile` / `--profile-file` only work if the cobra command has those flags (usually by adding the ProfileSettings layer). Without flags, bootstrap still works via env/config.
- **Missing default profile file**: error semantics depend on the “well-known default” file path (see `GatherFlagsFromProfiles` behavior).

### Related

- `glazed/pkg/appconfig/options.go`: `WithProfile`, `WithCobra`, `WithEnv`, `WithConfigFiles`
- `glazed/pkg/cmds/middlewares/profiles.go`: `GatherFlagsFromProfiles` error behavior + YAML format
- `glazed/pkg/doc/topics/15-profiles.md`: conceptual overview + circularity explanation
- `geppetto/pkg/layers/layers.go`: reference “Option A bootstrap parse” implementation
