---
Title: Structured streaming extraction for PR YAML (structuredsink)
Ticket: 014-STRUCTUREDSINK-STREAMING
Status: active
Topics:
  - prescribe
  - geppetto
  - inference
  - streaming
  - parsing
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
  - Path: ../GENERATE-LLM-INFERENCE--llm-inference-streaming-and-parsing/reference/01-diary.md
    Note: Original diary with implementation history and links
  - Path: ../GENERATE-LLM-INFERENCE--llm-inference-streaming-and-parsing/analysis/02-analysis-template-rendering-streaming-and-prdata-extraction.md
    Note: End-to-end inference pipeline (templating + streaming + parsing) context
  - Path: ../../../../../../geppetto/pkg/doc/topics/11-structured-data-event-sinks.md
    Note: structuredsink approach for tagged blocks and incremental extraction
  - Path: ../../../../../../geppetto/pkg/events/structuredsink/parsehelpers/helpers.go
    Note: parsehelpers.NewDebouncedYAML (candidate for incremental YAML extraction)
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-28T00:00:00Z
WhatFor: ""
WhenToUse: ""
---

# 014 — Structured streaming extraction for PR YAML (structuredsink)

## Overview

This ticket carries forward the **structured streaming extraction** work originally tracked in
`GENERATE-LLM-INFERENCE`. The goal is to emit **incremental structured PR data** (title/body/etc.)
while streaming, instead of waiting for a final parse of the full assistant output.

## Origin

Split from:
- `ttmp/2025/12/28/GENERATE-LLM-INFERENCE--llm-inference-streaming-and-parsing`

Start there for context:
- `reference/01-diary.md`
- `analysis/02-analysis-template-rendering-streaming-and-prdata-extraction.md`

## Tasks

See `tasks.md`.


