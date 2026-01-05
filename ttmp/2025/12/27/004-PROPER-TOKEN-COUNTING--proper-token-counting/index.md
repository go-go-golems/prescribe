---
Title: Proper token counting
Ticket: 004-PROPER-TOKEN-COUNTING
Status: complete
Topics:
    - prescribe
    - tokens
    - qol
    - tokenizer
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: geppetto/pkg/events/metadata.go
      Note: events.Usage schema for provider token usage
    - Path: geppetto/pkg/steps/ai/claude/engine_claude.go
      Note: Extracts Claude usage into events.Usage
    - Path: geppetto/pkg/steps/ai/gemini/engine_gemini.go
      Note: Does not currently populate metadata.Usage
    - Path: geppetto/pkg/steps/ai/openai/engine_openai.go
      Note: Extracts OpenAI stream usage into events.Usage
    - Path: prescribe/internal/api/api.go
      Note: Mock TokensUsed now uses tokenizer
    - Path: prescribe/internal/controller/controller.go
      Note: Context file/note token counts now use tokenizer
    - Path: prescribe/internal/domain/domain.go
      Note: Full/diff mode token recompute now uses tokenizer
    - Path: prescribe/internal/git/git.go
      Note: Changed file token counts now use tokenizer
    - Path: prescribe/internal/session/session.go
      Note: Session load recomputes file/context tokens via tokenizer
    - Path: prescribe/internal/tokens/tokens.go
      Note: Tokenizer-based token counter (replaces len/4 heuristic)
ExternalSources: []
Summary: ""
LastUpdated: 2026-01-04T22:40:20.621365508-05:00
WhatFor: ""
WhenToUse: ""
---



# Proper token counting

## Overview

<!-- Provide a brief overview of the ticket, its goals, and current status -->

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- prescribe
- tokens
- qol
- tokenizer

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
