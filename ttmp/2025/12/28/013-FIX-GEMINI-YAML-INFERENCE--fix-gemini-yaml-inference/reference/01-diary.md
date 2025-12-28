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

### What warrants a second pair of eyes
- The reflection-based extraction in the Gemini engine: confirm it matches the SDK fields in the current `generative-ai-go` version.
- The mapping of Gemini token counts to our generic `Usage` struct (input=prompt tokens, output=candidate tokens).

## Usage Examples

N/A (diary document)

## Related

- Bug report: `../analysis/01-bug-report-gemini-yaml-partial-output.md`
- Tasks: `../tasks.md`
