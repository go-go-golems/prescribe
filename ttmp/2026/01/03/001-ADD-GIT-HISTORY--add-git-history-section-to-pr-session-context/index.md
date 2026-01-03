---
Title: Add Git history section to PR session context
Ticket: 001-ADD-GIT-HISTORY
Status: active
Topics:
    - git
    - pr
    - context
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../go.work
      Note: Align workspace go version with module requirements so smoke tests can build/run
    - Path: cmd/prescribe/cmds/session/token_count.go
      Note: Counts derived git history tokens via request builder for parity with generate
    - Path: go.mod
      Note: Align go version with available toolchain so smoke tests can build/run
    - Path: internal/api/api.go
      Note: Fallback user context now includes Commit refs + Git history sections
    - Path: internal/api/prompt.go
      Note: Maps git_history AdditionalContext to .commits prompt variable
    - Path: internal/controller/controller.go
      Note: Injects git_history context item into GenerateDescriptionRequest
    - Path: internal/domain/domain.go
      Note: Adds ContextTypeGitHistory constant
    - Path: internal/export/context.go
      Note: Renders Git history section in export-context and export-rendered outputs
    - Path: internal/git/git.go
      Note: Builds git commit history text via git log --numstat
    - Path: internal/prompts/assets/create-pull-request.yaml
      Note: Embedded prompt pack; defines the .commits variable contract
    - Path: internal/session/session.go
      Note: Session YAML schema; only supports file/note context items today
    - Path: test-scripts/setup-test-repo.sh
      Note: Mock repo now creates multiple commits/authors for history coverage
    - Path: test-scripts/test-all.sh
      Note: Smoke suite asserts exported rendered payload includes history
    - Path: test-scripts/test-cli.sh
      Note: Smoke test asserts export-rendered includes commit history
    - Path: test/setup-test-repo.sh
      Note: Mock repo now creates multiple commits/authors for history coverage
    - Path: test/test-all.sh
      Note: Full suite asserts git_history and BEGIN COMMITS
    - Path: test/test-cli.sh
      Note: Export context/rendered tests assert git_history and BEGIN COMMITS
ExternalSources: []
Summary: ""
LastUpdated: 2026-01-03T16:00:15.725599907-05:00
WhatFor: ""
WhenToUse: ""
---




# Add Git history section to PR session context

## Overview

<!-- Provide a brief overview of the ticket, its goals, and current status -->

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- git
- pr
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
