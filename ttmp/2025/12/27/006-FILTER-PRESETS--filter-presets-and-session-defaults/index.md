---
Title: Filter presets and session defaults
Ticket: 006-FILTER-PRESETS
Status: active
Topics:
    - prescribe
    - filters
    - qol
    - session
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/controller/controller.go
      Note: Prompt preset load/save precedent; add analogous filter preset APIs here
    - Path: internal/domain/domain.go
      Note: Filter/FilterRule domain types; ActiveFilters behavior
    - Path: internal/session/session.go
      Note: Session YAML includes filters; default session path under .pr-builder/session.yaml
    - Path: internal/tui/app/boot.go
      Note: Startup session load; hook in default-filter application when session missing
    - Path: internal/tui/app/model.go
      Note: Hardcoded quick filter presets; future UI integration point
    - Path: ttmp/2025/12/27/003-QOL-IMPROVEMENTS--qol-improvements/analysis/02-filter-presets-and-default-filters-repo-user-session-defaults.md
      Note: Background analysis and requirements
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-27T16:39:29.640611745-05:00
WhatFor: ""
WhenToUse: ""
---


# Filter presets and session defaults

## Overview

<!-- Provide a brief overview of the ticket, its goals, and current status -->

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- prescribe
- filters
- qol
- session

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
