---
Title: Tasks
Ticket: 015-TUI-STREAMING
Status: active
Topics:
  - prescribe
  - tui
  - streaming
  - inference
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
  - Path: ../GENERATE-LLM-INFERENCE--llm-inference-streaming-and-parsing/tasks.md
    Note: Source ticket where this task originated
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-28T00:00:00Z
WhatFor: ""
WhenToUse: ""
---

# Tasks — 015-TUI-STREAMING

## Carried forward from GENERATE-LLM-INFERENCE

- [ ] Wire streaming generation into `prescribe tui` (see analysis doc in the source ticket)

## Implementation plan (suggested)

- [ ] Create a Bubble Tea command that starts streaming generation:
  - create `EventRouter`, attach `WatermillSink`, run `router.Run(ctx)` + `engine.RunInference(ctx, seed)` in an `errgroup`
- [ ] Map streaming events into Bubble Tea messages:
  - delta events → append to result viewport
  - completion event → store final text and structured YAML (if available)
- [ ] Cancellation:
  - store `cancel()` in the model and call it on user action
- [ ] UX:
  - keep UI responsive while streaming
  - keep stdout-like semantics inside the UI (final output displayed and copyable)

## Tests

- [ ] Add deterministic test harness for streaming:
  - mock event router handler → feed deltas
  - verify model updates result state correctly


