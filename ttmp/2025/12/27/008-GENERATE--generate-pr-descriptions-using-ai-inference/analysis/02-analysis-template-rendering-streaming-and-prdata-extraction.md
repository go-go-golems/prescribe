---
Title: 'Analysis: Template rendering, stdio streaming inference, and structured PR data extraction'
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
WhatFor: "Detailed blueprint for: render prompt templates → run inference (stdio streaming) → extract structured PR output."
WhenToUse: "When implementing terminal (stdio) streaming output and robust extraction/parsing of PR output from the final Turn."
---

# Analysis: Template rendering, stdio streaming inference, and structured PR data extraction

## Executive Summary

This document describes an end-to-end “inference pipeline” for `prescribe`:

1. **Render the prompt template** using Glazed’s templating helpers (sprig + `TemplateFuncs`), following Pinocchio’s pattern.
2. **Build a seed Turn** (system/user blocks) and run the Geppetto engine.
3. **Stream inference events** to the terminal (stdio) using Watermill sinks + `events.EventRouter`.
4. **Extract deterministic result data** from the assistant output:
   - minimal path: “last assistant LLMText block” → parse YAML into PR fields
   - advanced path: structured streaming extraction using Geppetto’s tag-only `FilteringSink` + YAML extractor.

The goal is to make the pipeline:
- deterministic (clear boundaries, stable parsing),
- debuggable (export-context, printable final Turn, structured event logs),
- stream-friendly (partial updates in the terminal without blocking),
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
- `diff`: **do not** naively concatenate diffs; instead, reuse the same **separator/exporter approach** we implemented for `prescribe generate --export-context`:
  - for the XML default, represent each diff as a `<file ...><diff>...</diff></file>` fragment (per included file)
  - “concat all diffs” means “concatenate *well-delimited per-file diff fragments*”, not “smash unified diffs together without boundaries”
  - implementation detail: produce this via a dedicated helper in `internal/export` (preferred) so the `.diff` formatting stays consistent with `BuildGenerationContext(..., SeparatorXML)`
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
For the current milestone, we focus on **stdio streaming output** (not the TUI).

**Reference implementation**:
`geppetto/cmd/examples/simple-streaming-inference/main.go`

Canonical pattern:
- `router, _ := events.NewEventRouter(...)`
- `sink := middleware.NewWatermillSink(router.Publisher, "chat")`
- `engine := factory.NewEngineFromParsedLayers(..., engine.WithSink(sink))`
- add a router handler to print events to stdout/stderr (examples):
  - `events.StepPrinterFunc("", os.Stdout)` (human-friendly streaming)
  - `events.NewStructuredPrinter(os.Stdout, events.PrinterOptions{Format: "json"|"yaml"|"text", ...})` (structured event stream)
- run `router.Run(ctx)` and `engine.RunInference(ctx, seed)` concurrently in an `errgroup`

### Stage F: Extract deterministic “PR data”

Two options:

#### Option 1 (minimal): parse YAML from assistant text after completion

1) Extract assistant text:
- Prefer: last `turns.BlockKindLLMText` where `RoleAssistant`
- Fallback: if empty, treat as error and keep raw

2) Normalize + select a parse target (robust final extraction):
- Prefer extracting fenced YAML blocks from the assistant text using:
  - `geppetto/pkg/steps/parse.ExtractYAMLBlocks(markdownText)` (Goldmark-based; handles multiple fenced YAML blocks)
- If multiple YAML blocks are present, **parse the last one** (nearest to the end) to reduce “analysis + final” ambiguity.
- If no fenced YAML blocks exist, fall back to:
  - `geppetto/pkg/events/structuredsink/parsehelpers.StripCodeFenceBytes([]byte(raw))` (best-effort single fence stripping),
  - or parse the entire raw text as YAML as a last resort (often fails, but harmless if handled).
- Always trim whitespace before parsing and keep the original `raw` for debugging.

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

## Prescribe TUI streaming (later)

The design for wiring the streaming event router into Bubble Tea (forwarding partial deltas and final structured parse results into the UI)
is intentionally separated into its own ticket document:

- `analysis/03-analysis-tui-streaming-integration.md`

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
2) `blocks, _ := parse.ExtractYAMLBlocks(raw)`; if `len(blocks) > 0`, set `body := []byte(blocks[len(blocks)-1])`
3) else `_, body := parsehelpers.StripCodeFenceBytes([]byte(raw))`
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
  - terminal printer receives deltas in correct order
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
- YAML block extraction helper (robust final parsing):
  - `geppetto/pkg/steps/parse/yaml_blocks.go` (`parse.ExtractYAMLBlocks`)
- Fence stripping + debounced YAML parsing helper (usable for final parsing too):
  - `geppetto/pkg/events/structuredsink/parsehelpers/helpers.go`
- Structured streaming extraction:
  - `geppetto/pkg/doc/topics/11-structured-data-event-sinks.md`
- Prescribe prompt pack:
  - `prescribe/internal/prompts/assets/create-pull-request.yaml`


