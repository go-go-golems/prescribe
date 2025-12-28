---
Title: 'Bug report: token_count mismatch vs rendered payload token count'
Ticket: TOKEN-COUNT-DISCREPANCY
Status: active
Topics:
    - prescribe
    - tokenization
    - bug
    - inference
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-27T20:56:46.175955097-05:00
WhatFor: ""
WhenToUse: ""
---

# Bug report: token_count mismatch vs rendered payload token count

## Goal

Capture and investigate a mismatch between:
- `prescribe session show` **token_count** (preflight budget), and
- token counts computed externally (Pinocchio token counting, also “tiktoken-based”) for the *rendered prompt payload*.

## Context

### What we observed

- In `prescribe`, `session show` reported ~**146,878** tokens for a curated session (base `origin/main`).
- The user reports that when counting the “rendered version” using Pinocchio’s token count (also using tiktoken), it showed ~**248** tokens.

This is a huge discrepancy and likely means **we are counting different text blobs**, or using **different encodings**, or there is a **bug in the counting flow**.

### Tokenizer used by prescribe

`prescribe` uses `github.com/tiktoken-go/tokenizer`:
- Default encoding: **cl100k_base**
- Override with env var: `PRESCRIBE_TOKEN_ENCODING` (supported: `cl100k_base`, `o200k_base`, `r50k_base`, `p50k_base`, `p50k_edit`)

Implementation:
- `prescribe/internal/tokens/tokens.go` → `tokens.Count(s)` → `tokenizer.Get(enc)` → `codec.Count(s)`

Important: token counting is documented as **preflight** (not billing-authoritative). It should still be *roughly consistent* with other cl100k tokenizers for the same input text.

### Where `session show token_count` comes from

`session show` uses `domain.PRData.GetTotalTokens()`:
- sums `file.Tokens` for included files in `GetVisibleFiles()`
- plus `ctx.Tokens` for all `AdditionalContext` items

Relevant code:
- `cmd/prescribe/cmds/session/show.go` (field `token_count`)
- `internal/domain/domain.go` (`GetTotalTokens`)
- `internal/git/git.go` (initial token computation for diff via `tokens.Count(diff)`)
- `internal/session/session.go` (recomputes file tokens on load for diff/full modes)

### Where “rendered payload” comes from

Exported rendered payload is created by:
- `cmd/prescribe/cmds/generate.go` → `internal/export.BuildRenderedLLMPayload(req, sep)`
- which calls `internal/api.CompilePrompt(req)` (templating split + render)

So “rendered payload tokens” should correspond to tokenizing the concatenated `(system,user)` strings (or the XML envelope if you tokenize the exported XML file).

### Environment (redacted)

- Timestamp captured: 2025-12-27T20:57:36-05:00
- `PRESCRIBE_TOKEN_ENCODING`: (not set in environment snapshot)
- Inference run attempted earlier failed with OpenAI 401 (missing API key), which is unrelated to token counting except for reproduction convenience.

## Quick Reference

### Reproduction snapshot from current workspace

Repo: `/home/manuel/workspaces/2025-12-26/prescribe-import/prescribe`

Session state (captured):
- Base: `origin/main`
- `token_count`: **146,878**
- `visible_files`: 97
- `included_files`: 97
- `additional_context_items`: 12

Captured outputs:
- Session JSON: `/tmp/prescribe-session-show.json` (433 bytes)
- Exported payload sizes:
  - `/tmp/prescribe-rendered.xml`: 914,522 bytes
  - `/tmp/prescribe-context.xml`: 607,230 bytes

Commands used:

```bash
git fetch --all --prune
/tmp/prescribe-self -r . -t origin/main session init --save

# Apply exclude filter (zsh-safe quoting)
/tmp/prescribe-self -r . -t origin/main filter add --name "Trim huge docs" \
  --exclude 'ttmp/**' \
  --exclude 'TUI-SCREENSHOTS*' \
  --exclude 'FILTER-*.md' \
  --exclude 'PLAYBOOK-*.md' \
  --exclude 'dev-diary.md' \
  --exclude 'PROJECT-SUMMARY.md' \
  --exclude 'TUI-DEMO.md' \
  --exclude '*.pdf'

# Add context docs (selected)
/tmp/prescribe-self -r . -t origin/main context add PROJECT-SUMMARY.md
/tmp/prescribe-self -r . -t origin/main context add README.md
/tmp/prescribe-self -r . -t origin/main context add ttmp/2025/12/27/008-GENERATE--generate-pr-descriptions-using-ai-inference/reference/01-diary.md
/tmp/prescribe-self -r . -t origin/main context add ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/reference/02-diary.md

/tmp/prescribe-self -r . -t origin/main session show --output json > /tmp/prescribe-session-show.json

# Export for external counting / debugging
/tmp/prescribe-self -r . -t origin/main generate --export-context --separator xml --output-file /tmp/prescribe-context.xml
/tmp/prescribe-self -r . -t origin/main generate --export-rendered --separator xml --output-file /tmp/prescribe-rendered.xml
```

### Hypotheses (most likely first)

- **H1: Counting different text**: Pinocchio is counting only the YAML *output* (or only a subset of the rendered payload), not the full `(system,user)` prompt content.
- **H2: Counting with a different encoding**: Pinocchio may default to `o200k_base` or a model-specific encoding, while prescribe defaults to `cl100k_base`.
- **H3: Session token_count includes content that is not in rendered payload**:
  - e.g. context items counted in session, but not injected into the rendered prompt (template does not include `.context` or logic differs).
- **H4: Session token_count is stale**: file tokens are computed from diffs/full-file content but the rendered payload includes something else (e.g., only summary or only paths), creating a mismatch.

### What we should add to the product (to make this diagnosable)

- A command to count tokens of the *exact rendered payload*, e.g.:
  - `prescribe generate --export-rendered --token-count`
- Per-file token breakdown for session:
  - `prescribe session show --tokens-by-file`

## Usage Examples

### How to validate if both tools tokenize the same text

1) Tokenize **exact same bytes** in both tools:
- use `/tmp/prescribe-rendered.xml` as the input to both tokenizers.
2) Ensure the same encoding:
- prescribe: default `cl100k_base` unless `PRESCRIBE_TOKEN_ENCODING` is set
- pinocchio: confirm its encoding (and force it if possible)

### Where to start in code

- Token counting:
  - `prescribe/internal/tokens.Count`
  - `prescribe/internal/domain.PRData.GetTotalTokens`
  - `prescribe/internal/git.Service.GetChangedFiles` (diff tokenization)
- Rendered prompt:
  - `prescribe/internal/api.CompilePrompt`
  - `prescribe/internal/export.BuildRenderedLLMPayload`

## Related

- PR run diary with concrete numbers and commands:
  - `ttmp/2025/12/28/GENERATE-LLM-INFERENCE--llm-inference-streaming-and-parsing/reference/02-pr-run-diary-draft-pr-description-150k-tokens.md`
- Generate inference ticket:
  - `ttmp/2025/12/28/GENERATE-LLM-INFERENCE--llm-inference-streaming-and-parsing/index.md`
