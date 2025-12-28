---
Title: Generate PR descriptions using AI inference
Ticket: 008-GENERATE
Status: active
Topics:
    - prescribe
    - geppetto
    - inference
    - pr-generation
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../geppetto/cmd/examples/simple-streaming-inference/main.go
      Note: Reference implementation for streaming inference via Watermill sink + errgroup
    - Path: ../../../../../../geppetto/pkg/doc/topics/06-inference-engines.md
      Note: Comprehensive guide to inference engine architecture
    - Path: ../../../../../../geppetto/pkg/inference/engine/engine.go
      Note: Engine interface for inference
    - Path: ../../../../../../geppetto/pkg/turns/types.go
      Note: Turn and Block structure definitions
    - Path: ../../../../../../pinocchio/cmd/pinocchio/cmds/catter/pkg/fileprocessor.go
      Note: Core file processing logic for exporting file contents
    - Path: ../../../../../../pinocchio/cmd/pinocchio/cmds/catter/pkg/stats.go
      Note: Token counting and statistics computation
    - Path: ../../../../../../pinocchio/cmd/pinocchio/prompts/code/create-pull-request.yaml
      Note: Reference prompt template structure
    - Path: cmd/prescribe/cmds/generate.go
      Note: --export-context + --separator flags
    - Path: internal/api/api.go
      Note: |-
        Mock API service to be replaced with geppetto engine
        API service now consumes StepSettings and runs geppetto inference via Turns
    - Path: internal/controller/controller.go
      Note: Controller injects StepSettings into API service; GenerateDescription now takes context
    - Path: internal/export/context.go
      Note: Catter-style exporter for full generation context (xml default)
    - Path: internal/tui/export/export.go
      Note: Current export format for PR generation context
    - Path: pkg/doc/topics/02-how-to-generate-pr-description.md
      Note: Document export-only context generation flags
    - Path: ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/playbooks/01-cli-testing-playbook.md
      Note: Standard CLI playbook now includes export-context
    - Path: analysis/02-analysis-template-rendering-streaming-and-prdata-extraction.md
      Note: End-to-end blueprint: Glazed templating + streaming inference + structured PR data parsing/extraction
    - Path: internal/api/prompt.go
      Note: Current prompt compiler (split+render pinocchio-style prompt using glazed templating)
    - Path: ../../../../../../geppetto/pkg/doc/topics/11-structured-data-event-sinks.md
      Note: Tag-only structured sink + extractor-owned parsing for robust streaming structured extraction
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-27T18:17:03.937288008-05:00
WhatFor: ""
WhenToUse: ""
---





# Generate PR descriptions using AI inference

## Overview

<!-- Provide a brief overview of the ticket, its goals, and current status -->

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- prescribe
- geppetto
- inference
- pr-generation

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- design/ - Architecture and design documents
- reference/ - Prompt packs, API contracts, context summaries
- playbooks/ - Command sequences and test procedures
- scripts/ - Temporary code and tooling
- various/ - Working notes and research
- archive/ - Deprecated or reference-only artifacts
