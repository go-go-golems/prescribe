---
Title: 'Token counting: geppetto + prescribe'
Ticket: 004-PROPER-TOKEN-COUNTING
Status: active
Topics:
    - prescribe
    - tokens
    - qol
    - tokenizer
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Token counting in geppetto is provider-reported usage; prescribe currently uses len(text)/4 and should switch to a tokenizer-based counter for preflight UI totals.
LastUpdated: 2025-12-27T15:06:05.05049541-05:00
WhatFor: Decide how prescribe should count tokens (preflight estimate vs provider usage) and document the relevant geppetto APIs/call sites.
WhenToUse: When implementing accurate token counts in prescribe UI and when integrating real LLM backends.
---

## Overview

There are two different needs that get conflated as “token counting”:

- **Provider usage (authoritative billing usage)**: returned by the LLM provider as part of the response. This is what `geppetto` primarily uses.
- **Preflight token estimation (before you call the provider)**: useful for UIs (like `prescribe` TUI) to stay under context limits. This requires a **local tokenizer** aligned with the target model/encoding.

Today:

- `geppetto` is **provider-usage-first**: it extracts token usage from provider SDK responses and stores it in `events.Usage`.
- `prescribe` is **heuristic-only**: it estimates tokens as `len(text)/4` for diffs, full files, context notes, and even the mock generated output.

## How geppetto counts tokens

### Common representation: `events.Usage`

`geppetto/pkg/events/metadata.go` defines a provider-agnostic usage struct:

- `InputTokens`, `OutputTokens`
- `CachedTokens` (OpenAI prompt caching detail)
- `CacheCreationInputTokens`, `CacheReadInputTokens` (Claude cache details)

These are attached to `events.EventMetadata.LLMInferenceData.Usage` so they can be rendered in UIs and persisted consistently.

### OpenAI: stream usage via `response.Usage`

In `geppetto/pkg/steps/ai/openai/engine_openai.go`, usage is extracted from streaming chunks:

- `PromptTokens` → `events.Usage.InputTokens`
- `CompletionTokens` → `events.Usage.OutputTokens`
- `PromptTokensDetails.CachedTokens` → `events.Usage.CachedTokens`

Important API detail: OpenAI streaming only includes usage when the request opts into it (`StreamOptions.IncludeUsage = true` in the request builder).

### Claude: final usage from the response

In `geppetto/pkg/steps/ai/claude/engine_claude.go`, usage is populated from the final merged response:

- `response.Usage.InputTokens`
- `response.Usage.OutputTokens`

Claude also exposes cache read/creation token fields (represented on `events.Usage`), but the specific fields depend on the client wrapper response type.

### Gemini: not currently wired to usage in geppetto

In `geppetto/pkg/steps/ai/gemini/engine_gemini.go`, the engine publishes start/partial/final events but does **not** currently extract provider usage into `metadata.Usage`.

So, as-is, `geppetto` has:

- strong usage extraction for **OpenAI** and **Claude**
- missing usage extraction for **Gemini** (follow-up opportunity)

## What this means for prescribe

`prescribe` needs token counts in two places:

- **TUI “Tokens:”**: this is a *preflight* number that helps you decide what to include in the LLM context. This should be a **local tokenizer count**, not `len(text)/4`.
- **Post-inference usage** (future): once `prescribe` uses a real provider, it should prefer **provider usage** fields (like geppetto does) for reporting/billing/telemetry.

Because current `prescribe` generation is a mock, the only meaningful “accurate” improvement right now is to replace the heuristic with a tokenizer-based counter for all content types.

## API options to count tokens “properly”

### Option A (authoritative): provider usage fields (preferred when calling a provider)

- OpenAI: `usage.prompt_tokens`, `usage.completion_tokens`, plus optional caching details (streaming requires `IncludeUsage`)
- Claude: `usage.input_tokens`, `usage.output_tokens`, plus cache read/creation token fields when enabled
- Gemini: provider SDKs expose usage/metadata depending on model + endpoint; geppetto currently doesn’t plumb it through

### Option B (preflight): local tokenizer library

For preflight UI totals, use a tokenizer matching the model family (e.g. OpenAI encodings like `cl100k_base` / `o200k_base`).

This is the approach we will take inside `prescribe` for now, because `prescribe` needs counts *before* any provider call.

