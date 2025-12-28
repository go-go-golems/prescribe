---
Title: Tasks
Ticket: 014-STRUCTUREDSINK-STREAMING
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
RelatedFiles:
  - Path: ../GENERATE-LLM-INFERENCE--llm-inference-streaming-and-parsing/tasks.md
    Note: Source ticket where these tasks originated
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-28T00:00:00Z
WhatFor: ""
WhenToUse: ""
---

# Tasks — 014-STRUCTUREDSINK-STREAMING

## Carried forward from GENERATE-LLM-INFERENCE

- [ ] Update prompt to optionally emit a tagged block (e.g. `<prescribe:prdata:v1>...`) for structuredsink extraction
- [ ] Implement a `prdata` extractor session using `parsehelpers.NewDebouncedYAML` to emit `prdata-update` / `prdata-completed` events

## Implementation details (suggested)

- [ ] Define the tag contract:
  - Start tag: `<prescribe:prdata:v1>`
  - End tag: `</prescribe:prdata:v1>`
  - Inner payload: YAML matching `domain.GeneratedPRData`
- [ ] Extend `create-pull-request` prompt to optionally include this tagged block when a flag is enabled.
- [ ] Implement a `structuredsink` extractor:
  - Input: stream deltas / partial completions
  - Output events:
    - `prdata-update` (best-effort partial YAML)
    - `prdata-completed` (final parsed object)
- [ ] Wire extractor events into the existing `EventRouter` in streaming mode:
  - ensure these events can be forwarded to TUI later

## Tests

- [ ] Unit tests for extractor:
  - Handles incremental chunks
  - Debounces correctly
  - Produces `completed` only when YAML is valid and stable


