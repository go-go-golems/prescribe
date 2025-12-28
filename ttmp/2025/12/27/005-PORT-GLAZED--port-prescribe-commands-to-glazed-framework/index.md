---
Title: Port Prescribe Commands to Glazed Framework
Ticket: 005-PORT-GLAZED
Status: review
Topics:
    - glazed
    - prescribe
    - porting
    - cli
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: glazed/cmd/examples/appconfig-parser/main.go
      Note: Example of new appconfig API pattern
    - Path: glazed/pkg/doc/tutorials/05-build-first-command.md
      Note: Primary tutorial for building Glazed commands
    - Path: glazed/pkg/doc/tutorials/custom-layer.md
      Note: Tutorial for creating custom parameter layers
    - Path: prescribe/cmd/prescribe/cmds/filter/add.go
      Note: Port filter add to Glazed-built BareCommand (legacy flags preserved) (commit 451d28b)
    - Path: prescribe/cmd/prescribe/cmds/filter/list.go
      Note: |-
        Dual-mode filter list (classic + Glaze structured output) (commit da425db)
        Dropped dual-mode toggle; always Glazed output (commit 9860e32)
    - Path: prescribe/cmd/prescribe/cmds/filter/show.go
      Note: |-
        Dual-mode filter show (classic + Glaze output) (commit 3f05fca)
        Dropped dual-mode toggle; always Glazed output (commit 9860e32)
    - Path: prescribe/cmd/prescribe/cmds/filter/test.go
      Note: |-
        Dual-mode filter test with legacy flags + Glaze row output (commit 3a028b5)
        Dropped dual-mode toggle; always Glazed output (commit 9860e32)
    - Path: prescribe/cmd/prescribe/cmds/helpers/controller.go
      Note: Controller pattern used by all commands
    - Path: prescribe/cmd/prescribe/cmds/helpers/controller_from_layers.go
      Note: Create+init controller from Glazed parsed layers (commit 8e294d1)
    - Path: prescribe/cmd/prescribe/cmds/root.go
      Note: |-
        Root command structure with persistent flags
        Root command now initializes logging in PersistentPreRunE (commit 90d7951)
        NewRootCmd/InitRootCmd for deterministic init (commit da425db)
    - Path: prescribe/cmd/prescribe/cmds/session/show.go
      Note: |-
        Dual-mode session show with --yaml preserved and Glaze summary row output (commit 425af79)
        Dropped classic output + --yaml; always Glazed output (commit 9860e32)
    - Path: prescribe/cmd/prescribe/main.go
      Note: |-
        Glazed logging/help init and loading Prescribe help topics (commit 90d7951)
        Explicit command initialization via cmds.InitRootCmd (commit da425db)
    - Path: prescribe/pkg/doc/embed.go
      Note: Embeds Prescribe markdown help sections for Glazed help system (commit 90d7951)
    - Path: prescribe/pkg/doc/topics/01-filters-and-glob-syntax.md
      Note: Prescribe help topic loaded into Glazed help system (commit 90d7951)
    - Path: prescribe/pkg/layers/existing_cobra_flags_layer.go
      Note: Wrap schema sections when flags already exist on root (commit da425db)
    - Path: prescribe/pkg/layers/filter.go
      Note: FilterLayer + GetFilterSettings helper (commit cb59b50)
    - Path: prescribe/pkg/layers/generation.go
      Note: GenerationLayer + GetGenerationSettings helper (commit cb59b50)
    - Path: prescribe/pkg/layers/repository.go
      Note: RepositoryLayer + GetRepositorySettings helper (commit cb59b50)
    - Path: prescribe/pkg/layers/session.go
      Note: SessionLayer + GetSessionSettings helper (commit cb59b50)
    - Path: prescribe/ttmp/2025/12/27/005-PORT-GLAZED--port-prescribe-commands-to-glazed-framework/playbook/01-playbook-port-existing-cobra-verbs-to-glazed-no-back-compat.md
      Note: Onboarding playbook for porting commands
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-27T16:54:22.815236356-05:00
WhatFor: ""
WhenToUse: ""
---














# Port Prescribe Commands to Glazed Framework

## Overview

<!-- Provide a brief overview of the ticket, its goals, and current status -->

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- glazed
- prescribe
- porting
- cli

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
