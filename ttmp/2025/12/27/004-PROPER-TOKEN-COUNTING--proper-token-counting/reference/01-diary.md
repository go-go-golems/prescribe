---
Title: Diary
Ticket: 004-PROPER-TOKEN-COUNTING
Status: active
Topics:
    - prescribe
    - tokens
    - qol
    - tokenizer
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Research + implementation diary for 004-PROPER-TOKEN-COUNTING.
LastUpdated: 2025-12-27T15:06:05.102160435-05:00
WhatFor: Track research and implementation steps for replacing prescribe’s heuristic token estimate with a tokenizer-based counter.
WhenToUse: Read this first when resuming work on proper token counting in prescribe.
---

# Diary

## Goal

Track research and implementation steps for replacing `prescribe`’s heuristic token estimate (`len(text)/4`) with a real tokenizer-based counter.

## Step 1: Ticket setup

This step created the dedicated ticket workspace and seeded it with an analysis doc for research and this diary for ongoing progress.

### What I did
- Created ticket: `004-PROPER-TOKEN-COUNTING`
- Created docs:
  - `analysis/01-token-counting-geppetto-prescribe.md`
  - `reference/01-diary.md`

### Commands

```bash
cd /home/manuel/workspaces/2025-12-26/prescribe-import && docmgr ticket create-ticket --ticket 004-PROPER-TOKEN-COUNTING --title "Proper token counting" --topics prescribe,tokens,qol,tokenizer && docmgr doc add --ticket 004-PROPER-TOKEN-COUNTING --doc-type analysis --title "Token counting: geppetto + prescribe" && docmgr doc add --ticket 004-PROPER-TOKEN-COUNTING --doc-type reference --title "Diary" && docmgr doc list --ticket 004-PROPER-TOKEN-COUNTING --with-glaze-output --output json
```

## Step 2: Research — how geppetto handles token usage

This step established that `geppetto` primarily treats token counts as **provider-reported usage**, not local tokenization. This matters because `prescribe` needs a preflight count in the TUI before any provider call exists.

### What I did
- Located the common usage container type: `geppetto/pkg/events/metadata.go` (`events.Usage`).
- Found OpenAI usage extraction from streaming chunks in `geppetto/pkg/steps/ai/openai/engine_openai.go`.
- Found Claude usage extraction from the final merged response in `geppetto/pkg/steps/ai/claude/engine_claude.go`.
- Noted that Gemini usage is not currently extracted/populated in `geppetto/pkg/steps/ai/gemini/engine_gemini.go`.

### What I learned
- **Geppetto is provider-usage-first**: `events.Usage` is populated from SDK responses (OpenAI/Claude).
- For preflight counts (like `prescribe` TUI), we still need a **local tokenizer** aligned with the model/encoding.

### Technical details
- Analysis doc: `../analysis/01-token-counting-geppetto-prescribe.md`

## Step 3: Implementation — replace len(text)/4 with tokenizer-based counting

This step replaced the rough `len(text)/4` estimate with a real tokenizer-backed counter, so the TUI “Tokens:” total reflects actual tokenization rather than character length. It also fixes the session-load consistency issue by recomputing file tokens based on the saved mode immediately when applying a session.

**Commit (code):** 59f3c73b839009dce7d2627b93f47208434b7a97 — "prescribe: replace heuristic token estimate with tokenizer"

### What I did
- Added `prescribe/internal/tokens/tokens.go` using `github.com/tiktoken-go/tokenizer`.
- Replaced token calculation call sites:
  - changed file diffs (`internal/git`)
  - context items (`internal/controller`)
  - diff/full-file mode recompute (`internal/domain` and `internal/model`)
  - session apply recompute (`internal/session`)
  - mock API `TokensUsed` (`internal/api`)
- Added dependency and verified with `go test`.

### Commands

```bash
cd /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe && go mod tidy && go test ./...
```

### Notes
- Encoding defaults to `cl100k_base`, override via `PRESCRIBE_TOKEN_ENCODING` (e.g. `o200k_base`).
