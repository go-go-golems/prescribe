---
Title: 'Analysis: Template rendering, streaming inference, and structured PR data extraction'
Ticket: 008-GENERATE
Status: active
Topics:
    - prescribe
    - geppetto
    - streaming
    - templating
    - inference
    - pr-generation
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-28T00:00:00.000000000Z
WhatFor: "Detailed blueprint for: render prompt templates → run inference (streaming) → extract structured PR output."
WhenToUse: "When implementing streaming generation in TUI/CLI and parsing the model output into structured PR fields."
---

# Analysis: Template rendering, streaming inference, and structured PR data extraction

## Executive Summary

This document describes an end-to-end “inference pipeline” for `prescribe`:

1. **Render the prompt template** using Glazed’s templating helpers (sprig + `TemplateFuncs`), following Pinocchio’s pattern.
2. **Build a seed Turn** (system/user blocks) and run the Geppetto engine.
3. **Stream inference events** into the TUI using Watermill sinks + `events.EventRouter`.
4. **Extract deterministic result data** from the assistant output:
   - minimal path: “last assistant LLMText block” → parse YAML into PR fields
   - advanced path: structured streaming extraction using Geppetto’s tag-only `FilteringSink` + YAML extractor.

The goal is to make the pipeline:
- deterministic (clear boundaries, stable parsing),
- debuggable (export-context, printable final Turn, structured event logs),
- stream-friendly (partial updates in the UI without blocking),
- provider-agnostic (StepSettings chooses engine/provider).

## Current code state (as of 2025-12-28)

### Prompt rendering (templating)

Prescribe now renders the embedded pinocchio-style prompt template using:
`glazed/pkg/helpers/templating.CreateTemplate(...).Parse(...).Execute(...)`.

This is intentionally the same mechanism used in Pinocchio (see `pinocchio/pkg/cmds/cmd.go`), so:
- sprig functions are available (eg `join`)
- TemplateFuncs are available (Cobra-lifted helpers + extra Glazed helpers)

### Non-streaming inference

The service currently:
- compiles prompt into `(systemPrompt, userPrompt)`
- runs `engine.RunInference(ctx, seedTurn)`
- extracts the “last assistant LLMText” for the description

## Pipeline overview (recommended architecture)

### Stage A: Build canonical generation request

Boundary stays:
`Controller.BuildGenerateDescriptionRequest() -> api.GenerateDescriptionRequest`

This request is the canonical snapshot of:
- included files (diff or full content)
- additional context items (notes and context files)
- prompt string (combined/system+template text)
- source/target branches

### Stage B: Compile prompt (split + render templates)

**Inputs**:
- `req.Prompt` (combined prompt)
- `req.Files`, `req.AdditionalContext` (data for variables)

**Outputs**:
- `systemPrompt` (rendered string)
- `userPrompt` (rendered string; contains the final instruction payload)

**Pinocchio-style prompt split**

The embedded preset uses a Go template definition:

```text
{{ define "context" -}}
...
{{- end }}
```

We treat:
- `systemTemplate`: text before `{{ define "context" ...`
- `userTemplate`: text starting at `{{ define "context" ...` through the remainder

**Rendering engine**

Use Glazed’s templating helper:
- `templating.CreateTemplate(name)` (adds `sprig.TxtFuncMap()` and `templating.TemplateFuncs`)
- `Parse(text)`
- `Execute(writer, vars)`

**Template variables**

The pinocchio prompt expects (subset):
- `.diff` (string) — unified diff text
- `.code` (list) — each entry has `.Path`, `.Content`
- `.context` (list) — each entry has `.Path`, `.Content`
- `.description` (string) — human description (we map context notes to this by default)
- plus other optional vars (`.commits`, `.issue`, `.title`, `.additional_system`, `.additional`, ...)

Mapping rules (recommended):
- `diff`: concatenate all `FileTypeDiff` diffs (already filtered to included files by the controller)
- `code`: include any `FileTypeFull` content (prefer FullAfter/FullBefore then fallback to Diff)
- `context`: include `ContextTypeFile` items (path + content)
- `description`: concatenate `ContextTypeNote` items (newline-separated)

### Stage C: Build seed Turn

Use Turns (Engine-first) rather than conversation manager:

- `systemPrompt` becomes `turns.NewSystemTextBlock(systemPrompt)`
- `userPrompt` becomes `turns.NewUserTextBlock(userPrompt)`

For future extensions (tools, multimodal), add additional blocks between them.

### Stage D: Run inference (non-streaming)

Non-streaming path is:
- `engine := factory.NewEngineFromStepSettings(stepSettings)`
- `updatedTurn, err := engine.RunInference(ctx, seed)`
- `assistantText := extractLastAssistantText(updatedTurn)`

### Stage E: Run inference (streaming)

Streaming requires attaching a sink, and running an event router in parallel.

**Reference implementation**:
`geppetto/cmd/examples/simple-streaming-inference/main.go`

Canonical pattern:
- `router, _ := events.NewEventRouter(...)`
- `sink := middleware.NewWatermillSink(router.Publisher, "chat")`
- `engine := factory.NewEngineFromParsedLayers(..., engine.WithSink(sink))`
- add router handlers to print/forward events
- run `router.Run(ctx)` and `engine.RunInference(ctx, seed)` concurrently in an `errgroup`

### Stage F: Extract deterministic “PR data”

Two options:

#### Option 1 (minimal): parse YAML from assistant text after completion

1) Extract assistant text:
- Prefer: last `turns.BlockKindLLMText` where `RoleAssistant`
- Fallback: if empty, treat as error and keep raw

2) Normalize for parsing:
- strip code fences like:
  - ```yaml ... ```
  - ``` ... ```
- trim whitespace
- optionally: take the last YAML-ish document if multiple are present

3) Parse YAML into a Go struct (proposed):

```yaml
title: ...
body: |
  ...
changelog: |
  ...
release_notes:
  title: ...
  body: |
    ...
```

4) Store:
- Raw output (always)
- Parsed fields (when parse succeeds)

#### Option 2 (robust + streaming-friendly): structured tag extraction + YAML extractor

Use Geppetto structured sinks:
`geppetto/pkg/events/structuredsink` (tag-only sink, extractor-owned parsing) and
`geppetto/pkg/events/structuredsink/parsehelpers` (debounced YAML parsing; fence stripping).

**How it works**
- Update the prompt to emit a tagged block:

```text
<prescribe:prdata:v1>
```yaml
title: ...
...
```
</prescribe:prdata:v1>
```

- Wrap the engine sink:
  - base sink: Watermill sink (publishes events to router)
  - filtering sink: intercepts streamed partials/final, extracts the tagged payload, forwards filtered text, publishes typed events

**Benefits**
- You get:
  - normal “chat text” streaming to the UI (filtered)
  - typed `prdata-update` / `prdata-completed` events with parsed YAML, even before completion (snapshots)
- Parsing is deterministic and robust across partial boundaries.

## Streaming into the Prescribe TUI (design)

### Requirements
- TUI should show live partial output on the Result screen while generation is running.
- TUI should remain responsive (scroll/quit/back).
- Cancellation should stop both router and inference.

### Proposed message/event flow (Bubble Tea)

1) `GenerateRequested` triggers a “start streaming generation” cmd:
- creates router, sink, engine
- installs a handler that forwards events into a channel
- runs router + inference in goroutines

2) The cmd emits Bubble Tea messages:
- `GenerationStartedMsg`
- `GenerationDeltaMsg{Delta string}` on partial completions
- `GenerationCompletedMsg{FinalText string, Parsed PRData?}` when done
- `GenerationFailedMsg{Err error}`

3) The Result model consumes these:
- append deltas to viewport
- on completed: show final parsed summary + allow copy/export

### Mapping Watermill events to UI messages

From the Watermill sink, you’ll receive Geppetto events (serialized).
A handler can:
- parse JSON via `events.NewEventFromJson` (or helper)
- switch on type:
  - partial delta events → `GenerationDeltaMsg`
  - final event → `GenerationCompletedMsg`
  - errors → `GenerationFailedMsg`

If using structuredsink + extractor, you’ll also receive typed events:
- `prdata-update` (structured snapshots)
- `prdata-completed` (final parsed object)

Those can update a “structured preview” pane in the TUI (optional).

## Deterministic output parsing (YAML → PR fields)

### Proposed PR result type (internal)

Even if the TUI currently stores only `GeneratedDescription` as a string, we should parse into a struct for:
- displaying title/body separately
- later output formats (GitHub API, changelog tooling)

Proposed schema:
- `Title string`
- `Body string`
- `Changelog string`
- `ReleaseNotes.Title string`
- `ReleaseNotes.Body string`
- `Raw string` (original assistant output)

### Parsing algorithm (minimal path)

1) `raw := assistantText`
2) `lang, body := parsehelpers.StripCodeFenceBytes([]byte(raw))`
3) if `lang` is empty or not yaml:
   - still try YAML parse on `body` (best-effort)
4) `yaml.Unmarshal(body, &result)`
5) validate required fields (at least `body` or `title`)
6) on failure: keep raw text

### Parsing algorithm (streaming structured sink)

Let the extractor own parsing:
- use `parsehelpers.NewDebouncedYAML[PRResult](DebounceConfig{...})`
- call `FeedBytes(chunk)` in `OnRaw`
- call `FinalBytes(raw)` in `OnCompleted`

Emit typed events with:
- partial snapshots (if parse succeeds)
- final snapshot + success status

## Testing strategy (recommended)

### Unit tests
- Prompt compilation:
  - pinocchio-style combined prompt renders (no `{{` remains)
  - plain prompt falls back
- YAML parsing:
  - parse valid YAML
  - parse fenced YAML
  - handle malformed YAML gracefully (returns raw)

### Integration tests (no real provider)
- Use a fake engine implementation or a test sink that replays:
  - partial completion events
  - final event with known YAML payload
- Assert:
  - UI receives deltas in correct order
  - final parsed PR result matches expected

## Open questions / decisions to make

1) **Should we switch the default prompt output to structured tags?**
   - If yes, we should adopt the structuredsink approach early—it will pay off for deterministic parsing.
2) **Do we want `.description` to be separate from context notes?**
   - Today, prescribe has no separate “description” field; we can keep mapping notes into `.description`.
3) **Where should structured PR fields live?**
   - `domain.PRData` currently only stores `GeneratedDescription string`. We likely want a structured field (or a parallel result type) once we parse YAML.

## Key references

- Glazed templating helpers:
  - `glazed/pkg/helpers/templating/templating.go` (`CreateTemplate` adds sprig + TemplateFuncs)
- Pinocchio prompt rendering pattern:
  - `pinocchio/pkg/cmds/cmd.go` (`renderTemplateString(...)`)
- Streaming reference:
  - `geppetto/cmd/examples/simple-streaming-inference/main.go`
- Structured streaming extraction:
  - `geppetto/pkg/doc/topics/11-structured-data-event-sinks.md`
- Prescribe prompt pack:
  - `prescribe/internal/prompts/assets/create-pull-request.yaml`


