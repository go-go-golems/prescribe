---
Title: PR creation - codebase analysis
Ticket: PR-CREATION
Status: active
Topics:
    - cli
    - git
    - prescribe
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-29T08:33:58.021624745-05:00
WhatFor: ""
WhenToUse: ""
---

# PR creation - codebase analysis

## Goal

Map the **files and symbols** in `prescribe/` that are relevant to implementing **end-to-end PR creation**
(not just PR description generation), so we can later turn this into concrete docmgr tasks.

## Context

`prescribe` today is a CLI/TUI for **generating PR descriptions** (and a structured YAML payload: `title`, `body`, `changelog`, `release_notes`).
It already has:

- git plumbing to compute `target...source` diffs
- a canonical “generation request” builder (controller → api)
- a default prompt pack adapted from Pinocchio (“create-pull-request”)
- best-effort parsing of the assistant output into a structured `domain.GeneratedPRData`

It does **not** appear to have an implementation that actually **creates** a GitHub PR (no `gh` / GitHub API call sites found in `prescribe/internal/**`).
The README explicitly suggests pairing output with the GitHub CLI (`gh pr create --body-file ...`), which is likely the simplest integration surface.

## Quick Reference

### “Where is the PR description produced?”

- **CLI entry point**: `prescribe/cmd/prescribe/cmds/generate.go`
  - Builds an initialized controller, loads the default session if present, then runs generation.
- **Canonical request builder**: `prescribe/internal/controller/controller.go`
  - `(*Controller).BuildGenerateDescriptionRequest() (api.GenerateDescriptionRequest, error)`
- **Inference + parse**: `prescribe/internal/api/api.go`
  - `(*api.Service).GenerateDescription(...)` and `GenerateDescriptionStreaming(...)`
  - parses assistant output via `ParseGeneratedPRDataFromAssistantText`
- **Parse logic**: `prescribe/internal/api/prdata_parse.go`
  - fenced-yaml extraction + salvage heuristics

### “Where does the prompt come from?”

- **Embedded prompt pack**: `prescribe/internal/prompts/assets/create-pull-request.yaml`
- **Default prompt loader**: `prescribe/internal/prompts/default.go`

### “Where does branch/diff/file context come from?”

- **Git subprocess wrapper**: `prescribe/internal/git/git.go` (`type Service`)
  - `GetCurrentBranch()`, `GetDefaultBranch()`, `GetChangedFiles()`, `GetDiff()`, `ResolveCommit()`, …
- **Controller initialization**: `prescribe/internal/controller/controller.go`
  - `(*Controller).Initialize(targetBranch string) error` fills `domain.PRData` with branches + `[]FileChange`

### “Where is persistent state?”

- **Session model**: `prescribe/internal/session/session.go` (`type Session`)
- **Default session path**: `<repo>/.pr-builder/session.yaml` (see `session.GetDefaultSessionPath`)

### “What structured PR data do we already have?”

- `prescribe/internal/domain/domain.go`
  - `type PRData`
  - `type GeneratedPRData` (+ `GeneratedPRDataRN`)

### Symbol map (high-signal)

- **Prompt contract**
  - `prescribe/internal/prompts/assets/create-pull-request.yaml`: expects YAML keys `title`, `body`, `changelog`, `release_notes`
  - `prescribe/internal/prompts.DefaultPrompt() string`: loads the embedded prompt pack as a single combined template string
- **PR data model**
  - `prescribe/internal/domain.PRData`: carries `SourceBranch`, `TargetBranch`, optional `Title`/`Description`, files/context, and generated output
  - `prescribe/internal/domain.GeneratedPRData`: structured parse target for PR YAML (`Title`, `Body`, `Changelog`, `ReleaseNotes`)
- **Generation pipeline**
  - `prescribe/internal/controller.(*Controller).BuildGenerateDescriptionRequest() (api.GenerateDescriptionRequest, error)`
  - `prescribe/internal/controller.(*Controller).GenerateDescription(ctx) (string, error)`
  - `prescribe/internal/controller.(*Controller).GenerateDescriptionStreaming(ctx, w) (string, error)`
  - `prescribe/internal/api.(*Service).GenerateDescription(ctx, req) (*GenerateDescriptionResponse, error)`
  - `prescribe/internal/api.ParseGeneratedPRDataFromAssistantText(text) (*domain.GeneratedPRData, error)`
- **Git plumbing**
  - `prescribe/internal/git.NewService(repoPath) (*git.Service, error)`
  - `prescribe/internal/git.(*Service).GetCurrentBranch() (string, error)`
  - `prescribe/internal/git.(*Service).GetDefaultBranch() (string, error)`
  - `prescribe/internal/git.(*Service).GetChangedFiles(source, target) ([]domain.FileChange, error)`
  - `prescribe/internal/git.(*Service).ResolveCommit(ref) (string, error)`
- **Session plumbing**
  - `prescribe/internal/controller.(*Controller).GetDefaultSessionPath() string`
  - `prescribe/cmd/prescribe/cmds/helpers.LoadDefaultSessionIfExists(ctrl)`
  - `prescribe/internal/session.GetDefaultSessionPath(repoPath) string` (default: `<repo>/.pr-builder/session.yaml`)

### Likely integration surface for PR creation

There are two obvious options:

1. **Shell out to GitHub CLI** (`gh pr create …`)
2. Use the **GitHub API** (new dependency + auth story)

Given the repository docs already reference `gh pr create`, option (1) is likely the shortest path.

## Usage Examples

### Helpful searches while implementing PR creation

```bash
# PR prompt contract + parsing
rg -n "create-pull-request|GeneratedPRData|ParseGeneratedPRDataFromAssistantText" prescribe/

# Generation entrypoint (to reuse in a "create PR" command)
rg -n "NewGenerateCommand\\(|Generating PR description|BuildGenerateDescriptionRequest" prescribe/cmd/ prescribe/internal/

# Git plumbing (push/branch/base/head will likely extend this surface)
rg -n "exec\\.Command\\(\"git\"|GetDefaultBranch\\(|GetCurrentBranch\\(" prescribe/internal/git/
```

### README hint (existing workflow)

`prescribe/README.md` suggests a workflow like:
- generate a description to a file
- then run `gh pr create --title ... --body-file ...`

This is not implemented in code yet; it’s just documentation of intended usage.

## Related

- Ticket diary: `reference/01-diary.md`
- Prescribe end-to-end description workflow: `prescribe/pkg/doc/topics/02-how-to-generate-pr-description.md`
