---
Title: Bug report — Gemini returns partial/invalid YAML for create-pull-request
Ticket: 013-FIX-GEMINI-YAML-INFERENCE
Status: active
Topics:
  - inference
  - gemini
  - yaml
  - parsing
DocType: analysis
Intent: short-term
Owners: []
RelatedFiles:
  - Path: ../../../../../../../prescribe/internal/api/api.go
    Note: Logs seed prompt + assistant raw output (preview+hash) for debugging
  - Path: ../../../../../../../prescribe/internal/api/prdata_parse.go
    Note: PR YAML parsing, block selection, salvage, and repair logic
  - Path: ../../../../../../../prescribe/internal/prompts/assets/create-pull-request.yaml
    Note: YAML-only contract prompt (embedded preset)
  - Path: ../../../../../../../prescribe/cmd/prescribe/cmds/generate.go
    Note: `--stream` mode prints parsed PR YAML summary to stderr
  - Path: ../../../../../../../prescribe/ttmp/2025/12/27/012-USE-PINOCCHIO-PROFILES--use-pinocchio-profiles-for-generate-command/scripts/05-smoke-test-prescribe-generate-gemini-profile.sh
    Note: Repro script for Gemini profile runs
  - Path: ../../../../../../../prescribe/ttmp/2025/12/27/012-USE-PINOCCHIO-PROFILES--use-pinocchio-profiles-for-generate-command/scripts/06-compare-provider-profiles-generate.sh
    Note: Compare providers on same repo/session
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-28T00:00:00Z
WhatFor: ""
WhenToUse: ""
---

# Bug report — Gemini returns partial/invalid YAML for create-pull-request

## Summary

When running `prescribe generate` against Gemini via Pinocchio profiles (e.g. `PINOCCHIO_PROFILE=gemini-2.5-pro`),
the model often returns **partial / invalid YAML**, typically:

- a valid `title: ...` line
- followed by a bare `body` key **without** `:` or a value
- sometimes fenced as ```yaml, sometimes not

This violates the YAML-only contract and prevents structured PR parsing from producing `body/changelog/release_notes`.

## Expected

Assistant returns valid YAML with at least:

```yaml
title: ...
body: |
  ...
changelog: |
  ...
release_notes:
  title: ...
  body: |
    ...
```

## Actual (evidence)

### Raw assistant output (Gemini profile)

Observed outputs in a small-repo run include forms like:

```text
title: '...'
body
```

or:

```text
```yaml
title: '...'
body
```
```

### Streaming vs non-streaming: both truncate due to MaxTokens

We ran the ticket comparison script to execute the same tiny repo/session twice (Gemini):
- once with `generate --stream`
- once without `--stream`

Artifacts (Dec 28, 2025):
- Base: `/tmp/prescribe-gemini-stream-vs-nonstream-20251228-152947`
- Streaming:
  - `/tmp/prescribe-gemini-stream-vs-nonstream-20251228-152947.stream.out.txt`
  - `/tmp/prescribe-gemini-stream-vs-nonstream-20251228-152947.stream.err.txt`
- Non-streaming:
  - `/tmp/prescribe-gemini-stream-vs-nonstream-20251228-152947.nonstream.out.txt`
  - `/tmp/prescribe-gemini-stream-vs-nonstream-20251228-152947.nonstream.err.txt`

Key finding: **both runs report** `stop_reason=FinishReasonMaxTokens` with very small output token counts (tens of tokens).
That strongly indicates the YAML is being cut off due to **max response tokens**, not due to streaming output parsing/assembly.

### Root cause: Pinocchio config defaults not applied to prescribe

Using `--print-parsed-parameters`, we confirmed that the `gemini-2.5-pro` profile does **not** set `ai-max-response-tokens`,
so `prescribe` initially fell back to the geppetto `ai-chat` layer default (1000). In our environment, Pinocchio additionally
has a global config file at `~/.pinocchio/config.yaml` that sets `ai-chat.ai-max-response-tokens: 4096`, but `prescribe`
did not load Pinocchio’s config because it only discovers `prescribe` config paths via `ResolveAppConfigPath("prescribe", "")`.

Fix: `prescribe generate` now loads `~/.pinocchio/config.yaml` as a *defaults overlay* (lower precedence than profiles),
so provider/model selection still comes from `PINOCCHIO_PROFILE`, while token defaults are inherited.

### Seed prompt correctness

We instrumented debug logs to print:
- prompt lengths + SHA256 hashes + previews
- assistant raw output length + hash + preview

Those logs show the user prompt contains explicit “output YAML only” instructions and includes the full YAML schema.
So the prompt being *sent* appears structurally correct.

## What we changed so far (mitigations)

### Prompt contract tightening
- Updated `internal/prompts/assets/create-pull-request.yaml` to:
  - explicitly require YAML-only output and forbid code fences
  - avoid emitting empty description markers

### Parser robustness
- `ParseGeneratedPRDataFromAssistantText` now:
  - scans fenced YAML blocks from last→first and selects the first candidate that parses and matches minimal schema
  - has a prose-salvage fallback (parse from last `title:` line)
  - repairs a common formatting failure (`body` as a bare key line) into `body: ""` so parsing doesn’t hard-fail

### Debug logging
- Added debug logs to capture:
  - the seed Turn’s system/user prompt (preview+hash)
  - the raw assistant output (preview+hash)

## Repro procedure (recommended)

### 1) Gemini-only smoke repro

Run:
- `prescribe/ttmp/2025/12/27/012-USE-PINOCCHIO-PROFILES--use-pinocchio-profiles-for-generate-command/scripts/05-smoke-test-prescribe-generate-gemini-profile.sh`

It will:
- create a tiny repo
- init a session
- run `generate --stream`
- store stdout/stderr logs in `/tmp/…` and print tail(stderr)

### 2) Provider compare

Run:
- `prescribe/ttmp/2025/12/27/012-USE-PINOCCHIO-PROFILES--use-pinocchio-profiles-for-generate-command/scripts/06-compare-provider-profiles-generate.sh`

It runs the same session across:
- `gemini-2.5-pro`
- `o4-mini`
- `sonnet-4.5`

This helps determine whether the issue is Gemini-specific.

## Hypotheses

1) **Max response tokens too low** (most likely): both streaming and non-streaming runs ended with `FinishReasonMaxTokens`.
2) **Gemini model behavior**: even with sufficient tokens, the model may still violate “YAML-only” (code fences, partial keys) sometimes.
3) **Prompt interaction**: despite “YAML-only”, some models may output “schema header” and stop.

## Next steps (recommended)

1) Compare `--stream` vs non-streaming (`prescribe generate` without `--stream`) on Gemini:
   - If non-streaming yields full YAML → likely streaming assembly bug.
2) Inspect `geppetto/pkg/steps/ai/gemini/engine_gemini.go`:
   - verify how streaming chunks are collected and how final assistant text is extracted.
3) Add a single retry/repair loop when YAML is invalid or incomplete:
   - re-ask once: “Your YAML was invalid; output full YAML only”, include prior output for correction.


