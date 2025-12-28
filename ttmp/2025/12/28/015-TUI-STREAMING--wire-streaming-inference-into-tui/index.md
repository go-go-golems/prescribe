---
Title: Wire streaming inference into Prescribe TUI
Ticket: 015-TUI-STREAMING
Status: active
Topics:
  - prescribe
  - tui
  - streaming
  - inference
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
  - Path: ../GENERATE-LLM-INFERENCE--llm-inference-streaming-and-parsing/reference/01-diary.md
    Note: Original diary with implementation history and links
  - Path: ../GENERATE-LLM-INFERENCE--llm-inference-streaming-and-parsing/analysis/03-analysis-tui-streaming-integration.md
    Note: Design notes for Bubble Tea streaming integration (message flow, cancellation, event mapping)
  - Path: ../../../../../../prescribe/internal/tui/app/model.go
    Note: TUI entrypoint for generation; target for wiring streaming generation
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-28T00:00:00Z
WhatFor: ""
WhenToUse: ""
---

# 015 — Wire streaming inference into Prescribe TUI

## Overview

This ticket carries forward the **TUI streaming integration** task originally tracked in
`GENERATE-LLM-INFERENCE`. The goal is to show live deltas in the Bubble Tea UI while inference is
running, with proper cancellation and a final structured summary when available.

## Origin

Split from:
- `ttmp/2025/12/28/GENERATE-LLM-INFERENCE--llm-inference-streaming-and-parsing`

Start there for context:
- `reference/01-diary.md`
- `analysis/03-analysis-tui-streaming-integration.md`

## Tasks

See `tasks.md`.


