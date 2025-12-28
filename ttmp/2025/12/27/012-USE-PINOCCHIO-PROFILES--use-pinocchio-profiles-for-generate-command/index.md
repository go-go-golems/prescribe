---
Title: Use Pinocchio Profiles for Generate Command
Ticket: 012-USE-PINOCCHIO-PROFILES
Status: closed
Topics:
    - configuration
    - profiles
    - appconfig
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../geppetto/pkg/layers/layers.go
      Note: Example bootstrap pattern for profiles
    - Path: ../../../../../../glazed/cmd/examples/appconfig-profiles/main.go
      Note: Standalone WithProfile example program for debugging
    - Path: ../../../../../../glazed/pkg/appconfig/options.go
      Note: Parser options
    - Path: ../../../../../../glazed/pkg/appconfig/parser.go
      Note: appconfig.Parser implementation
    - Path: ../../../../../../glazed/pkg/appconfig/profile_test.go
      Note: Unit tests for WithProfile selection/precedence/error semantics
    - Path: ../../../../../../glazed/pkg/cli/cli.go
      Note: ProfileSettings layer definition
    - Path: ../../../../../../glazed/pkg/cmds/middlewares/profiles.go
      Note: Profile loading middleware
    - Path: cmd/prescribe/cmds/generate.go
      Note: Current generate command implementation to be refactored
    - Path: ttmp/2025/12/27/012-USE-PINOCCHIO-PROFILES--use-pinocchio-profiles-for-generate-command/scripts/01-smoke-test-prescribe-generate-profiles.sh
      Note: Reusable small-repo smoke test for generate profile loading and provenance
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-27T20:54:06.132997053-05:00
WhatFor: ""
WhenToUse: ""
---




# Use Pinocchio Profiles for Generate Command

## Overview

This ticket implements Pinocchio profile support (`~/.config/pinocchio/profiles.yaml`) for the `prescribe generate` command and refactors the command to use `appconfig.Parser` for cleaner configuration parsing.

**Goals:**
1. Enable loading profiles from `~/.config/pinocchio/profiles.yaml` for the generate command
2. Refactor `generate.go` to use `appconfig.Parser` instead of manual layer parsing
3. Add `WithProfile()` option to `appconfig` package for profile support

**Current Status:** Analysis complete. Ready for implementation.

**Key Findings:**
- Profile loading requires bootstrap pattern: ProfileSettings must be parsed before profile middleware can run
- appconfig.Parser currently lacks profile support - needs `WithProfile()` option
- Bootstrap pattern requires two-phase parsing: parse ProfileSettings first, then use resolved values
- Middleware execution order is reversed: slice order is low→high precedence, execution is high→low

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **closed**

## Topics

- configuration
- profiles
- appconfig

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
