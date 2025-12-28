---
Title: "Pinocchio YAML robustification parsing: patterns and what to port to prescribe"
Ticket: 012-USE-PINOCCHIO-PROFILES
Status: active
Topics:
  - configuration
  - profiles
  - appconfig
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
  - Path: ../../../../../../pinocchio/pkg/middlewares/agentmode/middleware.go
    Note: Pinocchio pattern: extract YAML blocks, try decoding each, validate required fields, scan from end for most-recent intent
  - Path: ../../../../../../geppetto/pkg/steps/parse/yaml_blocks.go
    Note: Shared helper used by pinocchio to extract fenced YAML blocks (goldmark AST)
  - Path: ../../../../../../prescribe/internal/api/prdata_parse.go
    Note: Prescribe structured PR YAML parsing; target for incorporating pinocchio patterns
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-28T00:00:00Z
WhatFor: ""
WhenToUse: ""
---

# Pinocchio YAML robustification parsing: patterns and what to port to prescribe

## Context

Pinocchio and prescribe both depend on a “YAML contract” for structured model output (e.g. PR
title/body/changelog/release_notes). In practice, models sometimes emit:
- multiple YAML examples before the “real” one
- prose before/after YAML
- multiple fenced YAML blocks

Pinocchio’s approach isn’t “make YAML parsing magically tolerant”; instead it is:
**extract all fenced YAML blocks, then attempt to decode each until a block matches the expected
shape**, preferring the most recent block.

Prescribe already uses `ExtractYAMLBlocks` and a small prose-salvage heuristic, but it historically
preferred “the last YAML block” without validating whether it actually matches the expected schema.

## What Pinocchio does (robustification patterns)

### Pattern 1: Extract fenced YAML blocks (shared helper)

Pinocchio uses `geppetto/pkg/steps/parse.ExtractYAMLBlocks`, which parses markdown via `goldmark`
and returns the inner content of fenced code blocks whose language is `yaml` or `yml`.

- This is implemented in `geppetto/pkg/steps/parse/yaml_blocks.go`.
- Important property: it is **fence-based**; it does not parse “naked YAML” without fences.

### Pattern 2: Try-decoding multiple candidates + validate fields

In `pinocchio/pkg/middlewares/agentmode/middleware.go`, the mode-switch detector:
- iterates YAML blocks
- attempts `yaml.Unmarshal` into a typed struct
- **validates required fields** (e.g. `analysis` must be non-empty)
- returns the first match

There is also a variant that scans blocks from the end (`DetectYamlModeSwitchInBlocks`), which is
important when the model emits multiple YAML snippets and we want the most recent intent.

### Pattern 3: Prefer “nearest-to-the-end” behavior

Pinocchio’s “in blocks” detector scans from the back and returns the first match. This is a simple
and effective disambiguation rule:
- earlier YAML blocks are often examples or intermediate thoughts
- the last matching block tends to be the final instruction

## What to port to prescribe

### 1) Candidate validation (schema check) on extracted YAML blocks

Instead of always returning “last YAML block”, prescribe should:
- scan YAML blocks from last → first
- parse each candidate
- validate required PR fields (at least: `title`, `body` non-empty)
- return the first candidate that passes validation

This matches the Pinocchio robustness pattern and prevents selecting the wrong YAML when the last
block is an example or incomplete.

### 2) Keep existing prose-salvage, but apply validation

Prescribe already has a heuristic salvage for prose-wrapped YAML: parse from the last `title:` line.
This should remain, but we should validate the parsed output before accepting it, to avoid accidental
matches (e.g. `release_notes.title:`).

### 3) Optional future: “repair loop” (explicit retry)

Pinocchio’s robustification is entirely deterministic; it doesn’t “re-ask the model”. If we want
stronger guarantees across providers, we can add an explicit “repair” retry step at the inference
layer (not in pure parsing) when parsing fails.

For now, porting the candidate-selection+validation pattern gives most of the benefit with minimal
behavior risk.


