---
Title: 'Token count mismatch: prescribe session show vs rendered payload'
Ticket: TOKEN-COUNT-DISCREPANCY
Status: review
Topics:
    - prescribe
    - tokenization
    - bug
    - inference
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/prescribe/cmds/session/show.go
      Note: Exposes token_count field in session show output
    - Path: internal/api/prompt.go
      Note: CompilePrompt and template var mapping (.diff/.code/.context)
    - Path: internal/domain/domain.go
      Note: token_count semantics via PRData.GetTotalTokens (sum included file.Tokens + context item tokens)
    - Path: internal/export/context.go
      Note: Rendered payload exporter used for external token counting
    - Path: internal/tokens/tokens.go
      Note: Tokenizer implementation (tiktoken-go tokenizer; default cl100k_base; PRESCRIBE_TOKEN_ENCODING override)
    - Path: ttmp/2025/12/28/GENERATE-LLM-INFERENCE--llm-inference-streaming-and-parsing/reference/02-pr-run-diary-draft-pr-description-150k-tokens.md
      Note: Concrete session commands + token tuning steps used during report
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-28T11:01:59.563730233-05:00
WhatFor: ""
WhenToUse: ""
---



# Token count mismatch: prescribe session show vs rendered payload

## Overview

This ticket tracks an observed discrepancy between:
- `prescribe session show` **token_count** (preflight budget based on included diffs + additional context), and
- an external token counting result (Pinocchio, also “tiktoken-based”) for the *rendered payload*.

Goal: determine whether the mismatch comes from different input text, different encoding selection, stale accounting, or a bug—and then either fix it or document the intended semantics clearly.

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- prescribe
- tokenization
- bug
- inference

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
