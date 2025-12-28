---
Title: Tasks
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
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-28T00:00:00Z
WhatFor: ""
WhenToUse: ""
---

# Tasks — 013-FIX-GEMINI-YAML-INFERENCE

## Repro + evidence gathering

- [ ] Run the Gemini repro script and attach the artifacts paths to the bug report:
  - `prescribe/ttmp/2025/12/27/012-USE-PINOCCHIO-PROFILES--use-pinocchio-profiles-for-generate-command/scripts/05-smoke-test-prescribe-generate-gemini-profile.sh`
- [ ] Run provider compare on the same repo/session (Gemini vs OpenAI vs Claude):
  - `prescribe/ttmp/2025/12/27/012-USE-PINOCCHIO-PROFILES--use-pinocchio-profiles-for-generate-command/scripts/06-compare-provider-profiles-generate.sh`
  - Confirm whether the “bare `body`” issue is Gemini-specific.

## Debugging: streaming vs non-streaming

- [ ] Add a dedicated ticket script to compare Gemini streaming vs non-streaming on the same repo/session:
  - New script location: `scripts/01-compare-gemini-streaming-vs-nonstreaming.sh`
  - Run twice with `PINOCCHIO_PROFILE=gemini-2.5-pro`: once with `generate --stream`, once without.
  - Capture stdout/stderr artifacts under `/tmp/...` and compare assistant raw output length/hash and parse results.
- [ ] Re-run Gemini generation **without** `--stream` and compare raw assistant output length/hash:
  - If non-streaming returns full YAML, the issue is likely streaming aggregation in the Gemini engine.
  - If non-streaming also returns partial YAML, it’s likely model behavior / max tokens / safety refusal.
- [ ] Inspect geppetto Gemini engine response assembly:
  - `geppetto/pkg/steps/ai/gemini/engine_gemini.go`
  - Confirm how streaming chunks are concatenated and how final text is chosen.

## Fix options

- [ ] Implement a “repair retry” on invalid/partial YAML:
  - Detect when parsed YAML is invalid or missing required fields (`title/body`).
  - Re-ask the model once with the original prompt + “Your YAML was invalid; output full YAML only”.
  - Make retry provider-agnostic but primarily to unblock Gemini.
- [ ] Tighten prompt contract for Gemini:
  - Reduce ambiguity; add “do not stop early; include body/changelog/release_notes”.
  - Consider adding a short “minimal valid YAML example” (but ensure it doesn’t get copied as output).

## Regression tests

- [ ] Add unit tests for parsing/repair/retry triggers:
  - “bare `body` key” case
  - prose + YAML fence case
  - multiple YAML blocks (example + final)

## Handoff docs

- [ ] Keep `analysis/01-bug-report-gemini-yaml-partial-output.md` updated with:
  - latest repro artifacts paths
  - outcome per provider
  - what fix was chosen


