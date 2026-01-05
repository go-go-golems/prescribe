---
Title: Add branch selection for context file addition
Ticket: 009-CONTEXT-BRANCH
Status: archived
Topics:
    - prescribe
    - cli
    - context
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/prescribe/cmds/context/add.go
      Note: CLI command for adding context files
    - Path: internal/controller/controller.go
      Note: AddContextFile method implementation
    - Path: internal/git/git.go
      Note: GetFileContent method that supports any ref
ExternalSources: []
Summary: ""
LastUpdated: 2026-01-04T22:40:24.528446932-05:00
WhatFor: ""
WhenToUse: ""
---



# Add branch selection for context file addition

## Overview

<!-- Provide a brief overview of the ticket, its goals, and current status -->

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- prescribe
- cli
- context

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
