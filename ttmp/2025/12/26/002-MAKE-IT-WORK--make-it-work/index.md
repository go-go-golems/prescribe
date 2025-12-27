---
Title: Make it work
Ticket: 002-MAKE-IT-WORK
Status: active
Topics:
    - tui
    - bubbletea
    - ux
    - refactoring
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: prescribe/cmd/prescribe/cmds/tui.go
      Note: CLI launches app root TUI
    - Path: prescribe/internal/tui/app/boot.go
      Note: Boot session load cmd
    - Path: prescribe/internal/tui/app/default_deps.go
      Note: Deps implementation (clipboard/time)
    - Path: prescribe/internal/tui/app/model.go
      Note: Phase 2 app root model (Update/Init)
    - Path: prescribe/internal/tui/app/view.go
      Note: Phase 2 app views (main/filters/generating/result)
    - Path: prescribe/internal/tui/components/status/model.go
      Note: Phase 1 scaffolding - status footer model
    - Path: prescribe/internal/tui/components/status/toast.go
      Note: Phase 1 scaffolding - ID-safe toast state
    - Path: prescribe/internal/tui/components/status/toast_test.go
      Note: Phase 1 scaffolding - toast tests
    - Path: prescribe/internal/tui/events/events.go
      Note: Phase 1 scaffolding - shared typed TUI messages
    - Path: prescribe/internal/tui/keys/keymap.go
      Note: Phase 1 scaffolding - centralized keymap for help
    - Path: prescribe/internal/tui/layout/layout.go
      Note: Phase 1 scaffolding - layout computation helper
    - Path: prescribe/internal/tui/layout/layout_test.go
      Note: Phase 1 scaffolding - layout tests
    - Path: prescribe/internal/tui/styles/styles.go
      Note: Phase 1 scaffolding - Styles struct
    - Path: prescribe/ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/tasks.md
      Note: Detailed phased refactor task breakdown
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-26T18:32:05.980662011-05:00
WhatFor: ""
WhenToUse: ""
---




# Make it work

## Overview

<!-- Provide a brief overview of the ticket, its goals, and current status -->

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- tui
- bubbletea
- ux
- refactoring

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
