---
Title: Diary
Ticket: 003-QOL-IMPROVEMENTS
Status: active
Topics:
    - prescribe
    - qol
    - docs
    - tokens
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: prescribe/ttmp/2025/12/27/003-QOL-IMPROVEMENTS--qol-improvements/analysis/01-token-counting-in-prescribe.md
      Note: Deep-dive findings referenced by diary
ExternalSources: []
Summary: "Research + implementation diary for 003-QOL-IMPROVEMENTS (starting with token-counting investigation)."
LastUpdated: 2025-12-27T14:24:24.108864185-05:00
WhatFor: "Track what we learned, what we changed, and how to validate improvements for prescribe."
WhenToUse: "Read this first when resuming work on 003-QOL-IMPROVEMENTS."
---

# Diary

## Goal

Keep a step-by-step narrative of research and QOL changes for `prescribe`, with enough detail to resume later without re-deriving context.

## Step 1: Research — How are tokens counted in prescribe?

This step answered the question “How are tokens counted in prescribe?” and established that the app currently uses a character-length heuristic rather than a real tokenizer. This matters because the TUI “Tokens:” number is used for user decisions (include/exclude / diff vs full file), so accuracy and consistency are important.

### What I did
- Searched the codebase for token-related symbols (`Tokens`, `GetTotalTokens`, `TokensUsed`).
- Read the core paths where `Tokens` values are assigned and where the TUI total is computed.

### Why
- To confirm whether counts were computed via model-specific tokenizers or estimates.
- To identify the exact “source of truth” for the token number shown in the TUI.

### What worked
- Found that token counting is consistently implemented as `len(text)/4` in multiple places.
- Identified the computation path for the TUI total: `PRData.GetTotalTokens()`.

### What didn't work
- N/A (this was pure research).

### What I learned
- `prescribe` currently uses a **rough estimate**: **1 token ≈ 4 characters**.
- There are **two concepts**:
  - context token total shown in the TUI (sum of included visible files + added context)
  - mock “TokensUsed” returned by the mock API generator

### What was tricky to build
- There is a subtle consistency risk on session load: `session.ApplyToData` sets file modes but does **not** recompute file token counts; only context item tokens are recomputed from their content.

### What warrants a second pair of eyes
- Confirm intended behavior for token totals after loading sessions (should file tokens be recomputed immediately?).

### What should be done in the future
- If accuracy matters, replace `len(text)/4` with a tokenizer aligned to the target model/provider.
- Ensure session load applies the same token computation rules as interactive mode switches.

### Code review instructions
- Start with `prescribe/internal/domain/domain.go`:
  - `PRData.GetTotalTokens`
  - `ReplaceWithFullFile`, `RestoreToDiff`
- Then check assignment sites:
  - `prescribe/internal/git/git.go` (diff tokens)
  - `prescribe/internal/controller/controller.go` (context tokens)
  - `prescribe/internal/api/api.go` (mock `TokensUsed`)
  - `prescribe/internal/session/session.go` (`ApplyToData` nuance)

### Technical details
- Analysis doc: `../analysis/01-token-counting-in-prescribe.md`

## Step 2: Bookkeeping — Create ticket workspace + docs

This step created the `003-QOL-IMPROVEMENTS` ticket workspace and added two initial documents: an analysis doc capturing the token-counting findings, and this diary for ongoing research notes.

### What I did
- Ran:

```bash
cd /home/manuel/workspaces/2025-12-26/prescribe-import && docmgr ticket create-ticket --ticket 003-QOL-IMPROVEMENTS --title "QOL improvements" --topics prescribe,qol,docs,tokens && docmgr doc add --ticket 003-QOL-IMPROVEMENTS --doc-type analysis --title "Token counting in prescribe" && docmgr doc add --ticket 003-QOL-IMPROVEMENTS --doc-type reference --title "Diary" && docmgr doc list --ticket 003-QOL-IMPROVEMENTS --with-glaze-output --output json
```

### What worked
- Ticket created at: `prescribe/ttmp/2025/12/27/003-QOL-IMPROVEMENTS--qol-improvements/`
- Docs created:
  - `analysis/01-token-counting-in-prescribe.md`
  - `reference/01-diary.md`

## Step 3: Bookkeeping — Relate files + update changelog

This step connected the new docs to the key implementation files (for reverse-lookup via `docmgr doc search --file ...`) and added a changelog entry so the ticket has an audit trail.

### What I did
- Related the token-counting implementation files to the analysis doc (and linked docs in the ticket index).
- Updated the ticket changelog with a short summary and file links.

### Commands

```bash
cd /home/manuel/workspaces/2025-12-26/prescribe-import && docmgr doc relate --doc /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/ttmp/2025/12/27/003-QOL-IMPROVEMENTS--qol-improvements/analysis/01-token-counting-in-prescribe.md --file-note "/home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/internal/domain/domain.go:PRData.GetTotalTokens + token recomputation for diff/full modes" --file-note "/home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/internal/git/git.go:Initial token estimate for each changed file from len(diff)/4" --file-note "/home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/internal/controller/controller.go:Token estimate for additional context items" --file-note "/home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/internal/tui/app/view.go:TUI displays data.GetTotalTokens in stats line" --file-note "/home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/internal/api/api.go:Mock API response TokensUsed = len(output)/4" --file-note "/home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/internal/session/session.go:Session ApplyToData recomputes context tokens but not file tokens" && docmgr doc relate --ticket 003-QOL-IMPROVEMENTS --file-note "/home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/ttmp/2025/12/27/003-QOL-IMPROVEMENTS--qol-improvements/analysis/01-token-counting-in-prescribe.md:Token-counting analysis" --file-note "/home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/ttmp/2025/12/27/003-QOL-IMPROVEMENTS--qol-improvements/reference/01-diary.md:Implementation/research diary"
cd /home/manuel/workspaces/2025-12-26/prescribe-import && docmgr changelog update --ticket 003-QOL-IMPROVEMENTS --entry "Documented current token counting: len(text)/4 heuristic for TUI totals and mock API; added analysis + diary docs; linked key implementation files." --file-note "/home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/ttmp/2025/12/27/003-QOL-IMPROVEMENTS--qol-improvements/analysis/01-token-counting-in-prescribe.md:Token-counting analysis" --file-note "/home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/ttmp/2025/12/27/003-QOL-IMPROVEMENTS--qol-improvements/reference/01-diary.md:Research diary" --file-note "/home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/internal/domain/domain.go:GetTotalTokens + recompute logic"
```

## Step 4: Docs — Add a filters + glob syntax playbook

This step added a dedicated documentation page that explains how to create filters in `prescribe`, including the exact glob syntax supported by the `doublestar` matcher and the current include/exclude semantics (including the “multiple includes are ANDed” gotcha).

### What I did
- Created: `prescribe/pkg/doc/topics/01-filters-and-glob-syntax.md`
- Documented:
  - CLI usage (`prescribe filter add|test|list|show|remove|clear`)
  - Filter evaluation semantics (AND behavior for include rules and multiple filters)
  - Doublestar glob primitives (`*`, `**`, `?`, character classes, `{alts}`)
  - Known pitfalls + troubleshooting recipes


