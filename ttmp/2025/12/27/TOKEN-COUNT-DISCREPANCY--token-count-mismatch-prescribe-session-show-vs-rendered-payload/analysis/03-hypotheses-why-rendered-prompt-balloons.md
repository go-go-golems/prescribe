---
Title: 'Hypotheses: why rendered prompt balloons vs session token_count'
Ticket: TOKEN-COUNT-DISCREPANCY
Status: active
Topics:
    - prescribe
    - tokenization
    - prompts
    - debugging
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/prompts/assets/create-pull-request.yaml
      Note: Default prompt template includes a second `{{ template "context" . }}` under `.bracket`, likely duplicating context.
    - Path: internal/api/prompt.go
      Note: `buildTemplateVars` sets `.bracket=true` by default and constructs `.diff` with XML-ish wrappers.
    - Path: internal/domain/domain.go
      Note: `session show token_count` uses `PRData.GetTotalTokens()` (included file tokens + additional context tokens only).
    - Path: cmd/prescribe/cmds/generate.go
      Note: `--print-rendered-token-count` reports token counts of rendered system/user prompts.
ExternalSources: []
Summary: Working hypotheses for why rendered system/user payload can be ~2x+ larger than session context-only token budgeting.
LastUpdated: 2025-12-28T00:00:00Z
WhatFor: Provide a checklist of plausible causes for token ballooning and the concrete observations/tests to confirm or reject each.
WhenToUse: When comparing `session show token_count` to rendered payload token counts and needing to localize the difference.
---

# Hypotheses: why rendered prompt balloons vs session token_count

## Context

`session show token_count` is a **context-only budget**: it sums tokens for included file diffs/full content + additional context items. It does **not** include the system/user prompt template text.

The rendered LLM payload is the **actual prompt**: system + user, including instructions, schema, wrappers, and whatever the template chooses to render from `.diff`, `.code`, `.context`, `.description`, etc.

So ballooning can be “expected” even without a bug—but we still want to localize exactly *which* parts cause it.

## Hypotheses (ordered by likelihood)

### H1: Template duplicates the main context (bracketed re-render)

- **Mechanism**: The default template defines `context` and renders it once, then renders it **again** when `.bracket` is true:
  - `{{ template "context" . }}` appears twice (second time guarded by `{{ if .bracket }}`).
- **Evidence in repo**:
  - `internal/prompts/assets/create-pull-request.yaml` contains:
    - first render: `{{ template "context" . }}`
    - second render: `{{ if .bracket }} {{ template "context" . }} {{ end }}`
  - `internal/api/prompt.go` sets `"bracket": true` by default in `buildTemplateVars`.
- **Prediction**:
  - The rendered user prompt should contain duplicated markers (e.g. “The description of the pull request is …” twice).
  - Rendered token count should be roughly: `base_instructions + 2*(context block)` + optional additional context.
- **Test** (small repo):
  - Export rendered user prompt and count occurrences of a stable phrase from the `context` template; expect **2**.

### H2: Rendered prompt includes large instruction/scaffolding text not counted in session token_count

- **Mechanism**: The user prompt includes long guidance + YAML schema block + formatting rules.
- **Prediction**:
  - Even with no duplication, rendered user tokens should be materially larger than session context-only tokens.
- **Test**:
  - Compare `session token-count stored_total` vs `generate --print-rendered-token-count total`.
  - The delta should be mostly in user prompt tokens (system prompt is small).

### H3: `.diff` is wrapped in XML-ish per-file envelopes, increasing tokens

- **Mechanism**: `buildTemplateVars` creates `.diff` by joining per-file blocks like:
  - `<file name="..." type="diff">\n<diff>\n...\n</diff>\n</file>`
- **Prediction**:
  - Rendered user prompt contains additional wrapper tokens beyond raw diffs; this overhead scales with number of files.
- **Test**:
  - On a small repo, compare the raw `git diff` bytes to the rendered `.diff` portion (or count wrapper occurrences).

### H4: Additional context is included in rendered prompt in addition to the main context block

- **Mechanism**: The template renders `.context` separately (“Additional Context: …”) and may also include it inside the duplicated `context` block depending on future changes.
- **Prediction**:
  - Adding context files should increase both `session token_count` and rendered user prompt tokens.
- **Test**:
  - Run small repo repro with and without `context add` and compare deltas.

### H5: “Envelope overhead” from exported XML separator is counted by pinocchio but not by the rendered system/user token count

- **Mechanism**: `--separator xml` wraps system/user into an XML-ish `<prescribe><llm_payload>...` structure.
- **Prediction**:
  - `rendered.xml` document token count is slightly larger than `system+user` total (small constant overhead).
- **Test**:
  - Compare `generate --print-rendered-token-count total` vs `Rendered payload export token count`.

### H6: Tokenizer mismatch (encoding/model) inflates counts

- **Mechanism**: using different encoding (e.g. `o200k_base` vs `cl100k_base`) or different model defaults.
- **Prediction**:
  - For the same input bytes, prescribe and pinocchio counts diverge.
- **Current status**:
  - Already rejected for the large-repo repro: prescribe + pinocchio match exactly on `rendered.xml` with `cl100k_base`.

## Minimal next experiment (small repo)

Use the existing smoke-test repo generator (`test-scripts/setup-test-repo.sh`) and run:
- `session show --output json`
- `session token-count --output json`
- `generate --export-rendered --separator xml --print-rendered-token-count --output-file /tmp/...rendered.xml`
- grep the rendered user prompt for duplicated phrases to confirm H1 quickly.


