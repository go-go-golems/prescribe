---
Title: Generate PR descriptions using AI inference
Ticket: 008-GENERATE
Status: review
Topics:
    - prescribe
    - geppetto
    - inference
    - pr-generation
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../2025/12/28/GENERATE-LLM-INFERENCE--llm-inference-streaming-and-parsing/index.md
      Note: Inference (templating/streaming/parsing) has moved to GENERATE-LLM-INFERENCE
    - Path: ../../../../../../pinocchio/cmd/pinocchio/cmds/catter/pkg/fileprocessor.go
      Note: Core file processing logic for exporting file contents
    - Path: ../../../../../../pinocchio/cmd/pinocchio/cmds/catter/pkg/stats.go
      Note: Token counting and statistics computation
    - Path: cmd/prescribe/cmds/generate.go
      Note: --export-context + --separator + export-only functionality
    - Path: internal/export/context.go
      Note: Catter-style exporter for full generation context (xml default)
    - Path: pkg/doc/topics/02-how-to-generate-pr-description.md
      Note: Document export-only context generation flags
    - Path: ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/playbooks/01-cli-testing-playbook.md
      Note: Standard CLI playbook now includes export-context
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-27T20:00:48.529265643-05:00
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
