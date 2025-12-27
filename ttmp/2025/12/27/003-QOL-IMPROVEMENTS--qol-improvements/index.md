---
Title: QOL improvements
Ticket: 003-QOL-IMPROVEMENTS
Status: active
Topics:
    - prescribe
    - qol
    - docs
    - tokens
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: prescribe/pkg/doc/topics/01-filters-and-glob-syntax.md
      Note: Playbook for creating filters + doublestar glob syntax
    - Path: prescribe/ttmp/2025/12/27/003-QOL-IMPROVEMENTS--qol-improvements/analysis/01-token-counting-in-prescribe.md
      Note: Token-counting analysis
    - Path: prescribe/ttmp/2025/12/27/003-QOL-IMPROVEMENTS--qol-improvements/reference/01-diary.md
      Note: Implementation/research diary
ExternalSources: []
Summary: 'QOL improvements for prescribe (initial focus: clarify/fix token counting and related UX/consistency).'
LastUpdated: 2025-12-27T14:24:24.002193975-05:00
WhatFor: Home base for small-but-high-impact improvements to prescribe.
WhenToUse: Start here to find the current analysis/diary and linked implementation files.
---



# QOL improvements

## Overview

Initial focus: document and improve how `prescribe` reports token counts (currently a `len(text)/4` estimate) and ensure consistency across flows (e.g., session load vs interactive mode changes).

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- prescribe
- qol
- docs
- tokens

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
