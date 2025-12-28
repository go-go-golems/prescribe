---
Title: Generate LLM inference (streaming + parsing)
Ticket: GENERATE-LLM-INFERENCE
Status: active
Topics:
    - prescribe
    - geppetto
    - inference
    - streaming
    - templating
    - parsing
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../geppetto/cmd/examples/simple-streaming-inference/main.go
      Note: Canonical stdio streaming pattern (EventRouter + Watermill sink + errgroup)
    - Path: ../../../../../../geppetto/pkg/doc/topics/11-structured-data-event-sinks.md
      Note: Tag-only structured sink + extractor-owned parsing (useful for streaming + typed extraction)
    - Path: ../../../../../../geppetto/pkg/steps/parse/yaml_blocks.go
      Note: Robust fenced-YAML extraction from assistant markdown (final output parsing)
    - Path: ../../../../../../geppetto/pkg/events/structuredsink/parsehelpers/helpers.go
      Note: Fence stripping + YAML parsing helpers (usable for final parse too)
    - Path: internal/api/api.go
      Note: Prescribe inference service (Turn seed + engine inference + extraction)
    - Path: internal/api/prompt.go
      Note: Prompt compilation (split + render pinocchio-style combined prompt via glazed templating)
    - Path: internal/prompts/assets/create-pull-request.yaml
      Note: Default pinocchio-style prompt pack requiring template rendering
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-28T00:00:00.000000000Z
WhatFor: ""
WhenToUse: ""
---

# Generate LLM inference (streaming + parsing)

## Overview

This ticket owns the **actual LLM inference path** behind `prescribe generate`:
- streaming output to terminal (stdio),
- (later) streaming integration into the TUI,
- robust extraction/parsing of structured PR output (YAML) from the final Turn,
- optional structured streaming extraction via Geppetto `structuredsink`.

## Tasks

See [tasks.md](./tasks.md).

## Changelog

See [changelog.md](./changelog.md).


