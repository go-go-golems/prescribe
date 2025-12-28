---
Title: 'Design guide: generation pipeline, exporters, and geppetto StepSettings'
Ticket: 008-GENERATE
Status: active
Topics:
    - prescribe
    - geppetto
    - inference
    - pr-generation
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-27T18:32:34.656282312-05:00
WhatFor: ""
WhenToUse: ""
---

# Design guide: generation pipeline, exporters, and geppetto StepSettings

## Executive Summary

We are splitting “generation” into **two independent layers**:

- **Export layer (catter-like)**: deterministic “what would be sent to the model” string export with a configurable separator (default: **xml**). This is what you can use immediately to debug and iterate on prompt/context selection.
- **Inference layer (geppetto)**: real LLM inference using `geppetto` engines created from **already-parsed** `settings.StepSettings` (parsing happens in CLI/TUI).

The near-term milestone is: `prescribe generate --export-context --separator xml` prints the full prompt/context payload and exits (no inference).

## Problem Statement

Right now, we need a clear path to:

- Inspect the exact prompt/context payload (diffs/full files/context notes/manual prompt) in a **single blob** (catter-style), without triggering inference.
- Run real inference via geppetto, but **without** pushing config parsing down into `internal/api` (so the CLI/TUI can own parsing, profiles, config files, etc.).
- Keep the system composable so that later we can add streaming (Watermill sink) without rewriting everything.

## Proposed Solution

### 1) Canonical request stays the boundary

We keep `Controller.BuildGenerateDescriptionRequest()` as the canonical “inputs to generation”. Everything builds off that.

### 2) Add a catter-like exporter package

New package: `prescribe/internal/export`:

- `export.BuildGenerationContext(req, sep)` produces a single string from `api.GenerateDescriptionRequest`.
- `sep` supports: `xml` (default), `markdown`, `simple`, `begin-end`, `default`.

This exporter is the shared formatter for:
- CLI “export context” (now)
- TUI “copy context” (next)

### 3) `generate` has an export-only path

`prescribe generate` gains:
- `--export-context` (bool): prints exported payload and exits (no inference)
- `--separator` (string): selects separator, default `xml`

This is deliberately “catter-ish”: it’s about producing a stable input blob you can feed to other tools or inspect.

### 4) API service uses geppetto StepSettings (no parsing in api)

`prescribe/internal/api.Service`:
- gets configured via `SetStepSettings(*settings.StepSettings)`
- `GenerateDescription(ctx, req)` builds a seed `*turns.Turn` using `turns.NewTurnBuilder()` and runs `engine.RunInference`

Parsing is owned by:
- `prescribe generate` (CLI): `settings.NewStepSettingsFromParsedLayers(parsedLayers)`
- `prescribe tui` (TUI): same

### 5) Streaming later, following `simple-streaming-inference`

When we wire streaming into the TUI, we will follow the exact pattern from:
`geppetto/cmd/examples/simple-streaming-inference/main.go`:

- create `events.NewEventRouter()`
- create `middleware.NewWatermillSink(router.Publisher, "chat")`
- pass `engine.WithSink(watermillSink)` when creating engine
- run `router.Run(ctx)` and `engine.RunInference(ctx, seed)` in an `errgroup`

## Design Decisions

### Decision: Export-first milestone

We’re prioritizing the exporter because it is deterministic and immediately useful, even without API keys or provider configuration.

### Decision: XML default separator

Defaulting to **xml** makes the payload:
- more structured than markdown fences
- easier to parse later
- less likely to confuse the model about boundaries (prompt vs diff vs file)

### Decision: StepSettings parsing happens in CLI/TUI

This matches geppetto’s shape:
- `settings.NewStepSettingsFromParsedLayers(parsedLayers)`
- `factory.NewEngineFromStepSettings(stepSettings, options...)`

It also keeps `internal/api` testable and not responsible for config merging/profile selection.

## Alternatives Considered

### Alternative: Keep using the TUI’s markdown exporter only

Rejected as the only exporter because we want catter-like separators (xml default) and we want to reuse formatting for both CLI and TUI.

### Alternative: Parse geppetto config inside `internal/api`

Rejected because config parsing (files/env/profiles) is a CLI concern; `internal/api` should be “given settings, do inference”.

## Implementation Plan

### Phase 1 (Now): Export-only path

- Add `prescribe/internal/export.BuildGenerationContext(...)`
- Add `prescribe generate --export-context --separator ...`
- (Next) switch TUI CopyContext to use the new exporter

### Phase 2 (Now): StepSettings-based inference plumbing

- Parse `settings.StepSettings` in CLI/TUI
- Inject into controller via `Controller.SetStepSettings(...)`
- Use Turns directly in `internal/api` (no conversation manager)

### Phase 3 (Later): Streaming in TUI

- Follow the `simple-streaming-inference` pattern, wiring Watermill sink to a TUI event handler.

## Open Questions

- How should we split the combined prompt into system/user for the Turn? (currently we treat `req.Prompt` as system and exporter/context as user)
- How do we want to represent included vs excluded files in XML? (right now exporter prints included files from the request)
- Should exporter include token counts and commit hashes in the XML payload by default?

## References

- `geppetto/cmd/examples/simple-streaming-inference/main.go`: canonical streaming pattern
- `geppetto/pkg/steps/ai/settings/settings-step.go`: `settings.StepSettings`
- `geppetto/pkg/inference/engine/factory/helpers.go`: `factory.NewEngineFromStepSettings`
- `prescribe/internal/export/context.go`: exporter implementation (xml default)
- `prescribe/cmd/prescribe/cmds/generate.go`: `--export-context` / `--separator`
