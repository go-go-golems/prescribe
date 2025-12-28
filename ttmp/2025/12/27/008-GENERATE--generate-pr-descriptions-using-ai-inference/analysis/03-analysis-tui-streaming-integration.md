---
Title: 'Analysis: Streaming inference integration into the Prescribe TUI (later milestone)'
Ticket: 008-GENERATE
Status: active
Topics:
    - prescribe
    - tui
    - bubbletea
    - streaming
    - geppetto
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-28T00:00:00.000000000Z
WhatFor: "Design notes for wiring Geppetto streaming (EventRouter + WatermillSink) into Bubble Tea models and messages."
WhenToUse: "When implementing live streaming generation inside `prescribe tui`."
---

# Analysis: Streaming inference integration into the Prescribe TUI (later milestone)

## Scope

This document is intentionally **TUI-only**. The stdio streaming + final PR-data parsing pipeline is covered in:

- `analysis/02-analysis-template-rendering-streaming-and-prdata-extraction.md`

## Requirements

- Show **live partial output** on the Result screen while generation is running.
- Keep the UI responsive (scroll, quit, back).
- Cancellation should stop both the event router and inference.
- If structured PR output is available (YAML parse, or structuredsink extractor), show it (optionally) without breaking the plain text stream.

## Proposed message/event flow (Bubble Tea)

1) `GenerateRequested` triggers a “start streaming generation” command:
- create router, sink, engine
- install a router handler that forwards events into a channel
- run router + inference in goroutines (errgroup)

2) The command emits Bubble Tea messages:
- `GenerationStartedMsg`
- `GenerationDeltaMsg{Delta string}` on partial completions
- `GenerationCompletedMsg{FinalText string, Parsed PRData?}` when done
- `GenerationFailedMsg{Err error}`

3) The Result model consumes these:
- append deltas to viewport
- on completed: show final parsed summary + allow copy/export

## Mapping Watermill events to UI messages

From the Watermill sink, you’ll receive Geppetto events (serialized).
A handler should:
- parse JSON into typed events (via Geppetto `events` helpers),
- switch on event type:
  - partial completion delta → `GenerationDeltaMsg`
  - final event → `GenerationCompletedMsg`
  - errors → `GenerationFailedMsg`

If we adopt `structuredsink` + extractor for PR YAML:
- the router will also receive typed extractor events:
  - `prdata-update` (structured snapshots)
  - `prdata-completed` (final parsed object)

Those can update a “structured preview” pane in the TUI (optional).

## Cancellation model

- Use `ctx, cancel := context.WithCancel(parentCtx)`
- Run both `router.Run(ctx)` and `engine.RunInference(ctx, seed)` in an `errgroup`
- On user cancel:
  - call `cancel()`
  - let both goroutines exit

## Testing strategy (TUI-specific)

- Unit test message reduction in the Result model:
  - deltas append correctly
  - completion transitions state
  - cancellation stops streaming
- Integration test with a fake stream:
  - replay partial/final events into the router handler channel
  - ensure the UI receives ordered messages

## Key references

- Streaming reference (EventRouter + Watermill sink + errgroup):
  - `geppetto/cmd/examples/simple-streaming-inference/main.go`
- Structured streaming extraction (tag-only sink + extractor-owned parsing):
  - `geppetto/pkg/doc/topics/11-structured-data-event-sinks.md`


