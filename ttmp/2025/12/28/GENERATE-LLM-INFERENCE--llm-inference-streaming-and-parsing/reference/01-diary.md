---
Title: Diary
Ticket: GENERATE-LLM-INFERENCE
Status: active
Topics:
    - prescribe
    - geppetto
    - inference
    - streaming
    - parsing
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-27T20:18:36-05:00
WhatFor: ""
WhenToUse: ""
---

# Diary

## Step 1: Ticket split (from 008-GENERATE)

### What I did
- Created a new ticket workspace `GENERATE-LLM-INFERENCE`.
- Moved inference-focused tasks and documents out of `008-GENERATE` so `008` can stay export-focused.

### Why
- Keep `008-GENERATE` scoped to deterministic “export-context” functionality.
- Keep inference (templating, streaming, output parsing) in a dedicated ticket.

### Next steps
- Migrate the inference-heavy analysis/design docs from `008-GENERATE` into this ticket.
- Continue implementation work here (stdio streaming + robust PR YAML parsing).

## Step 2: Move inference-heavy docs into the new ticket

### What I did
- Moved the inference-heavy analysis/design documents (via `git mv`) from:
  - `ttmp/2025/12/27/008-GENERATE--.../analysis/*`
  - `ttmp/2025/12/27/008-GENERATE--.../design-doc/*`
  into this ticket’s `analysis/` and `design-doc/` directories.

### Why
- Keep `008-GENERATE` focused on exporter/export-only CLI workflows.
- Keep all inference architecture, streaming, and parsing guidance in the inference ticket.

### Result
- This ticket now contains:
  - `analysis/01-analysis-export-prescribe-diff-data-and-generate-pr-descriptions-with-geppetto-inference.md`
  - `analysis/02-analysis-template-rendering-streaming-and-prdata-extraction.md`
  - `analysis/03-analysis-tui-streaming-integration.md`
  - `design-doc/01-design-guide-generation-pipeline-exporters-and-geppetto-stepsettings.md`

## Step 3: Implement robust final YAML extraction/parsing for PR output

This step starts the “real inference” portion by making the output contract deterministic and machine-usable. The default prompt asks the model to output a YAML document with `title`, `body`, `changelog`, and `release_notes`; without a parser we can only treat the result as an opaque blob.

I’m implementing a best-effort parser that works for both non-streaming and future streaming modes: it extracts the last fenced YAML block if present (to avoid “analysis + final” ambiguity), and falls back to fence stripping + YAML unmarshal. Parsed fields are stored on the in-memory `domain.PRData` alongside the raw assistant text.

**Commit (code):** N/A — implementation in progress

### What I did
- Added a structured result type for PR YAML output (`domain.GeneratedPRData`) and stored it on `domain.PRData` as `GeneratedPRData` (best-effort; not persisted in session YAML).
- Implemented `api.ParseGeneratedPRDataFromAssistantText`:
  - prefer last fenced YAML block via `geppetto/pkg/steps/parse.ExtractYAMLBlocks`
  - fallback via `geppetto/pkg/events/structuredsink/parsehelpers.StripCodeFenceBytes`
- Wired parsing into `internal/api.Service.GenerateDescription` and stored parse output + parse error string in the response.
- Updated `Controller.GenerateDescription` to store parsed fields on `c.data` for later UI use.
- Added unit tests covering “prefer last YAML block” and “fence stripping fallback”.

### Why
- We need a deterministic “final output parsing” seam before streaming: it locks down the output contract and gives the TUI a structured payload to render.

### What warrants a second pair of eyes
- Whether “parse error string” should be stored even when there is no fenced YAML (i.e. treat as expected-unstructured vs actual-error).
- Whether the parser should validate required fields (title/body/changelog) vs allowing partial structs.

### Code review instructions
- Start at `internal/api/prdata_parse.go` and its tests.
- Review `internal/api/api.go` for where parsing is applied and how failures are carried.

## Step 4: Add stdio streaming mode (`prescribe generate --stream`)

This step adds a “live” UX for `prescribe generate`: while the model is producing output, we stream partial completions and events to the terminal. This is the same plumbing we’ll later reuse for TUI streaming, but for now it’s purely stdio-driven.

Implementation-wise, we follow Geppetto’s canonical pattern: create an `events.EventRouter`, attach a `middleware.NewWatermillSink` to the engine (`engine.WithSink`), register an `events.StepPrinterFunc` handler, and run `router.Run` in parallel with `engine.RunInference` using `errgroup`.

**Commit (code):** N/A — implementation in progress

### What I did
- Added `api.Service.GenerateDescriptionStreaming(ctx, req, w)` which:
  - streams events to `w` via `events.StepPrinterFunc`
  - still returns a final `GenerateDescriptionResponse` (including parsed PR YAML fields)
- Added `Controller.GenerateDescriptionStreaming(ctx, w)` to store the final raw and parsed results on `domain.PRData`.
- Added CLI flag `prescribe generate --stream` to enable streaming (events to stderr), while keeping final output behavior unchanged.
- Added a minimal unit test asserting streaming requires StepSettings (no provider configured).

### Why
- Streaming makes generation feel responsive and enables quick iteration on prompts without waiting for a full “silent” run.

### What warrants a second pair of eyes
- Decide whether the final “parsed PR summary” should be printed explicitly at the end of streaming mode (vs leaving that to future TUI integration).

### Code review instructions
- Start in `internal/api/api.go` (`GenerateDescriptionStreaming`).
- Review `cmd/prescribe/cmds/generate.go` for the `--stream` switch.

## Step 5: Print a final parsed PR-data summary at the end of streaming runs

Streaming is great for responsiveness, but it can be hard to “see the final structure” when the output arrives as deltas. This step adds a deterministic end-of-run summary for `--stream` runs: after inference completes, we print either the parsed PR YAML fields (preferred) or a clear parse-failed marker.

This keeps stdout semantics stable (final description output still goes to stdout or `--output-file`) while stderr carries both the live stream and the final structured summary.

**Commit (code):** N/A — implementation in progress

### What I did
- Updated the `prescribe generate --stream` path to print:
  - `--- Parsed PR data (YAML) ---` followed by a marshaled YAML struct when parsing succeeds, or
  - a parse error marker when parsing fails.

### Why
- Makes the “structured output contract” visible and debuggable during streaming runs without requiring a separate parsing step.

### Code review instructions
- Review `cmd/prescribe/cmds/generate.go` streaming branch for the stderr summary printing.

