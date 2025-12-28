---
Title: Diary
Ticket: 013-FIX-GEMINI-YAML-INFERENCE
Status: active
Topics:
    - inference
    - gemini
    - yaml
    - parsing
DocType: reference
Intent: short-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../geppetto/pkg/steps/ai/gemini/engine_gemini.go
      Note: Extracts Gemini finish reason + UsageMetadata into Turn metadata (commit 59a19acf2dbecd209218cca73ce53572560634f2)
    - Path: internal/api/api.go
      Note: Logs assistant preview/hash plus stop_reason+usage metadata (commit 2d8e0a159041bf6ac29e1631be30aeb925a80ba2)
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-28T15:16:54.745584272-05:00
WhatFor: ""
WhenToUse: ""
---


# Diary

## Goal

Track the investigation and fix for **Gemini producing partial/invalid YAML** for the create-pull-request prompt, with enough detail (commands, artifacts, code pointers) that we can resume quickly and review confidently.

## Context

- Ticket: `013-FIX-GEMINI-YAML-INFERENCE`
- Symptom: Gemini outputs partial YAML like `title: ...` followed by a bare `body` key line (`body` with no `:`), which violates our structured YAML contract and breaks PR data extraction.
- Key code:
  - `prescribe/internal/api/api.go`: runs inference and parses from the final assistant text block on the returned Turn
  - `prescribe/internal/api/prdata_parse.go`: YAML extraction/repair/validation
  - `geppetto/pkg/steps/ai/gemini/engine_gemini.go`: Gemini engine streaming assembly

## Quick Reference

## Step 1: Confirm parsing uses the final inference result (not streamed deltas)

This step clarified the control-flow: streaming is purely for **displaying incremental deltas**, while the structured PR YAML parsing is done from the **final assistant text** extracted from the `*turns.Turn` returned by `RunInference`. That means any “partial YAML” issue is not caused by our CLI parsing the streamed bytes; it’s either the model output itself or how the engine constructs the final Turn text.

This also highlights a subtle risk: our parser can “repair” malformed YAML into something that unmarshals but still yields an empty `body`, which can look like success even when the output is unusable. We’ll treat “parsable but missing required fields” as a failure trigger for a retry in later steps.

### What I did
- Read `api.Service.GenerateDescriptionStreaming` and verified it waits for inference completion, then parses from `extractLastAssistantText(updatedTurn)`.
- Verified Gemini engine accumulates all streaming deltas into a single `message` and appends it as an assistant text block to the Turn.

### Why
- We must parse the **final**, complete assistant output returned by the inference call, not anything printed during streaming, otherwise we risk truncation/missing content.

### What I learned
- Parsing is already using the inference result (the final Turn). The “bare `body`” issue is not explained by “parsing from streaming output”.

### Code review instructions
- Start in `prescribe/internal/api/api.go`:
  - `GenerateDescriptionStreaming` → `description := extractLastAssistantText(updatedTurn)` → `ParseGeneratedPRDataFromAssistantText(description)`
- Then check `geppetto/pkg/steps/ai/gemini/engine_gemini.go`:
  - `message += delta` loop → `turns.AppendBlock(t, turns.NewAssistantTextBlock(message))`

## Step 2: Add debug metadata (stop reason + token usage) to diagnose cut-off outputs

This step adds best-effort logging for Gemini response metadata so we can quickly tell whether the assistant output was likely **cut off** (e.g. due to max tokens), stopped normally, or ended for safety/refusal reasons. The main goal is observability: we don’t change prompt or parsing behavior yet; we just log more structured facts.

**Commit (code / geppetto):** 59a19acf2dbecd209218cca73ce53572560634f2 — "Gemini: record stop reason + token usage in Turn metadata"

**Commit (code / prescribe):** 2d8e0a159041bf6ac29e1631be30aeb925a80ba2 — "API: log stop reason + token usage when available"

### What I did
- Updated the Gemini engine to extract:
  - finish/stop reason (from candidates)
  - token usage (prompt/output counts)
- Stored these on the returned Turn’s metadata using standard keys (`stop_reason`, `usage`).
- Extended `prescribe` debug logging to print these metadata values next to the existing assistant preview/hash.

### Why
- Without stop reason/usage, “partial YAML” is ambiguous: it could be truncation (`MAX_TOKENS`), refusal/safety, or genuine model noncompliance.

### What didn't work
- The first attempt to commit the Gemini engine change failed due to `golangci-lint` (`exhaustive`) complaining about an intentionally-partial `switch` over `reflect.Kind`:

  - `pkg/steps/ai/gemini/engine_gemini.go:... missing cases in switch of type reflect.Kind (exhaustive)`

  We fixed it by adding `//nolint:exhaustive` on that numeric-only switch.

### What warrants a second pair of eyes
- The reflection-based extraction in the Gemini engine: confirm it matches the SDK fields in the current `generative-ai-go` version.
- The mapping of Gemini token counts to our generic `Usage` struct (input=prompt tokens, output=candidate tokens).

## Step 3: Reproduce and confirm truncation cause (MaxTokens) for stream vs non-stream

This step ran the new ticket script that executes Gemini twice on the same tiny repo/session: once with `generate --stream` and once without. Both runs produced truncated YAML and, critically, the captured metadata shows the **stop reason is MaxTokens** in both cases. That strongly suggests the “bare `body`/partial YAML” symptom is primarily a **token budget issue**, not a streaming-assembly bug.

### What I did
- Ran:
  - `prescribe/ttmp/2025/12/28/013-FIX-GEMINI-YAML-INFERENCE--fix-gemini-yaml-inference/scripts/01-compare-gemini-streaming-vs-nonstreaming.sh`
- Artifacts:
  - Base: `/tmp/prescribe-gemini-stream-vs-nonstream-20251228-152947`
  - Streaming:
    - `/tmp/prescribe-gemini-stream-vs-nonstream-20251228-152947.stream.out.txt`
    - `/tmp/prescribe-gemini-stream-vs-nonstream-20251228-152947.stream.err.txt`
  - Non-streaming:
    - `/tmp/prescribe-gemini-stream-vs-nonstream-20251228-152947.nonstream.out.txt`
    - `/tmp/prescribe-gemini-stream-vs-nonstream-20251228-152947.nonstream.err.txt`

### What worked
- We now have deterministic evidence (logs + stop_reason + token usage) to explain the truncation.

### Key observations
- Streaming run:
  - `stop_reason=FinishReasonMaxTokens`
  - `input_tokens=1718 output_tokens=22`
  - assistant output stops right after `body:` (truncated YAML)
- Non-streaming run:
  - `stop_reason=FinishReasonMaxTokens`
  - `input_tokens=1718 output_tokens=19`
  - assistant output stops right after the title line (also truncated)

### What I learned
- The issue reproduces **with and without** `--stream`; streaming display is not the cause.
- The model is being capped at a very small output token budget (tens of tokens), causing YAML truncation.

### What should be done next
- Implement a single retry path for invalid/partial YAML when `stop_reason` indicates max tokens:
  - bump `ai-max-response-tokens` for the retry (provider-agnostic; especially helps Gemini profiles)
  - add a corrective retry instruction (“Output complete YAML only; no fences; include all required keys”)
- Also consider documenting/adjusting recommended `ai-max-response-tokens` for Gemini profiles used with `prescribe generate`.

## Step 4: Trace the low max token limit back to config defaults (Pinocchio vs Prescribe)

This step answered “why is Gemini so capped?” by using `--print-parsed-parameters` to inspect provenance for `ai-max-response-tokens`. The key discovery is that the Gemini profile (`gemini-2.5-pro`) does **not** set `ai-max-response-tokens`, so we fall back to geppetto’s `ai-chat` defaults unless a config file overrides it. In my environment, Pinocchio has a global config file (`~/.pinocchio/config.yaml`) that sets `ai-chat.ai-max-response-tokens: 4096`, but `prescribe` (as a separate app name) did not load that file, causing the unexpectedly small output budget.

Important process note: `--print-parsed-parameters` can include **secret values** when your config/profile contains API keys; treat its output as sensitive.

**Commit (code / prescribe):** d290191c523ebc7907bab1b3cd5365f32cd707a5 — "Generate: apply ~/.pinocchio config as defaults overlay"

**Commit (code / prescribe):** 0e5fce96e8300a8a65655b8bb26e6bdffddaafb7 — "Generate: inherit ai-chat defaults from ~/.pinocchio/config.yaml"

### What I did
- Ran `prescribe generate --print-parsed-parameters` with `PINOCCHIO_PROFILE=gemini-2.5-pro` and confirmed:
  - `ai-max-response-tokens` initially came from defaults (1000) and not from the Gemini profile.
  - After loading Pinocchio config as a defaults overlay, `ai-max-response-tokens` resolves to 4096 from `~/.pinocchio/config.yaml`.
- Updated `prescribe generate` middleware chain to load Pinocchio config as a *defaults overlay* (lower precedence than profiles), while filtering non-layer keys like `repositories: [...]`.

### What worked
- With the overlay in place, Gemini runs complete without truncation and the retry path is not triggered.

### What warrants a second pair of eyes
- The config loader mapper: confirm we correctly ignore non-layer keys (e.g. `repositories`) and don’t accidentally load unrelated keys into existing layers.

## Usage Examples

N/A (diary document)

## Related

- Bug report: `../analysis/01-bug-report-gemini-yaml-partial-output.md`
- Tasks: `../tasks.md`
