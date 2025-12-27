---
Title: Token counting in prescribe
Ticket: 003-QOL-IMPROVEMENTS
Status: active
Topics:
    - prescribe
    - qol
    - docs
    - tokens
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
    - Path: prescribe/internal/api/api.go
      Note: Mock API response TokensUsed = len(output)/4
    - Path: prescribe/internal/controller/controller.go
      Note: Token estimate for additional context items
    - Path: prescribe/internal/domain/domain.go
      Note: PRData.GetTotalTokens + token recomputation for diff/full modes
    - Path: prescribe/internal/git/git.go
      Note: Initial token estimate for each changed file from len(diff)/4
    - Path: prescribe/internal/session/session.go
      Note: Session ApplyToData recomputes context tokens but not file tokens
    - Path: prescribe/internal/tui/app/view.go
      Note: TUI displays data.GetTotalTokens in stats line
ExternalSources: []
Summary: How token counts are estimated in the current prescribe TUI + mock API implementation (not a real tokenizer).
LastUpdated: 2025-12-27T14:24:24.057140988-05:00
WhatFor: To understand what the TUI 'Tokens:' number represents today, where it is computed, and what it does (and does not) correspond to.
WhenToUse: When reasoning about context size, inclusion/exclusion effects, or when replacing the rough estimate with a real tokenizer.
---


## Summary

`prescribe` currently does **not** use a real model tokenizer. It uses a **rough heuristic**:

- **Estimated tokens** = `len(text) / 4` (≈ 1 token per 4 characters)

This estimate is used for:

- the **TUI “Tokens:”** context total (sum of included file diffs / full files + added context items)
- the mock “TokensUsed” in the mock API description generator (based on generated output length)

## Two separate “token count” concepts

### 1) Context token estimate shown in the TUI

The TUI renders:

- visible files count
- filtered files count
- **`data.GetTotalTokens()`**

`GetTotalTokens()` sums:

- `Tokens` for **visible** files where `Included == true`, plus
- `Tokens` for **all** `AdditionalContext` items

### 2) Mock API “TokensUsed” estimate

The mock API response reports `TokensUsed` as `len(generatedText)/4`. This is not tied to a real provider response.

## Where `Tokens` values are assigned / recomputed

### Changed files (initial load)

When loading changed files, tokens are estimated from the **file diff**:

- `Tokens = len(diff) / 4`

Source: `prescribe/internal/git/git.go` (`GetChangedFiles`)

### Switching file mode (diff ⇄ full file)

When a file is switched to “full file” mode, tokens are recomputed from the stored full contents:

- before: `len(FullBefore)/4`
- after: `len(FullAfter)/4`
- both: `(len(FullBefore)+len(FullAfter))/4`

Switching back to diff mode recomputes as `len(Diff)/4`.

Source: `prescribe/internal/domain/domain.go` (`ReplaceWithFullFile`, `RestoreToDiff`)

### Additional context items

Added context items estimate tokens from their content:

- context file: `len(fileContent)/4`
- context note: `len(note)/4`

Source: `prescribe/internal/controller/controller.go` (`AddContextFile`, `AddContextNote`)

### Session load nuance (potential mismatch)

When applying a saved session, file `Type` / `Version` is set, but **file `Tokens` are not recomputed** as part of `ApplyToData` (it only recomputes tokens for context items from `len(cc.Content)/4`).

This means after loading a session, the displayed token total can remain based on the original diff-token estimate until another action triggers a recompute (e.g. using the TUI actions that call `ReplaceWithFullFile` / `RestoreToDiff`).

Source: `prescribe/internal/session/session.go` (`ApplyToData`)

## Key code pointers

- `prescribe/internal/tui/app/view.go`: renders the stats line with `data.GetTotalTokens()`
- `prescribe/internal/domain/domain.go`: `PRData.GetTotalTokens()`, `ReplaceWithFullFile()`, `RestoreToDiff()`
- `prescribe/internal/git/git.go`: initial token estimate from `len(diff)/4`
- `prescribe/internal/controller/controller.go`: context item token estimates
- `prescribe/internal/api/api.go`: mock “TokensUsed” as `len(output)/4`

## Implications / QOL opportunities

- **Accuracy**: The current value is a heuristic and will deviate from provider tokenizers.
- **Provider awareness**: Different models tokenize differently; the current approach is model-agnostic.
- **Session correctness**: Consider recomputing file tokens in `session.ApplyToData` based on the requested mode to keep totals consistent after loading.

