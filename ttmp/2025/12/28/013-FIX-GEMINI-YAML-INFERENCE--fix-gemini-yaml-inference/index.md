---
Title: Fix Gemini YAML inference (partial/truncated YAML output)
Ticket: 013-FIX-GEMINI-YAML-INFERENCE
Status: active
Topics:
    - inference
    - gemini
    - yaml
    - parsing
DocType: index
Intent: short-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../geppetto/pkg/steps/parse/yaml_blocks.go
      Note: YAML fence extraction helper used for robustness
    - Path: ../../../../../../pinocchio/pkg/middlewares/agentmode/middleware.go
      Note: 'Reference pattern: scan YAML blocks + validate required fields'
    - Path: cmd/prescribe/cmds/generate.go
      Note: Loads ~/.pinocchio/config.yaml as defaults overlay so ai-max-response-tokens applies (commits d290191c523e...
    - Path: internal/api/api.go
      Note: Debug logging for seed prompt + assistant raw output (preview+hash)
    - Path: internal/api/prdata_parse.go
      Note: Structured PR YAML parsing + robustness/repair logic
    - Path: internal/prompts/assets/create-pull-request.yaml
      Note: Prompt contract for YAML output
    - Path: ttmp/2025/12/28/012-USE-PINOCCHIO-PROFILES--use-pinocchio-profiles-for-generate-command/scripts/05-smoke-test-prescribe-generate-gemini-profile.sh
      Note: Repro script for Gemini profile inference
    - Path: ttmp/2025/12/28/012-USE-PINOCCHIO-PROFILES--use-pinocchio-profiles-for-generate-command/scripts/06-compare-provider-profiles-generate.sh
      Note: Compare Gemini vs OpenAI vs Claude via profiles on identical repo/session
    - Path: ttmp/2025/12/28/013-FIX-GEMINI-YAML-INFERENCE--fix-gemini-yaml-inference/scripts/01-compare-gemini-streaming-vs-nonstreaming.sh
      Note: Repro script comparing stream vs non-stream and capturing stop_reason/usage
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-28T00:00:00Z
WhatFor: ""
WhenToUse: ""
---


# 013 — Fix Gemini YAML inference (partial/truncated YAML output)

## Overview

Gemini profile runs (`PINOCCHIO_PROFILE=gemini-2.5-pro`) frequently return **invalid/partial YAML**
(e.g. `title: ...` followed by a bare `body` key with no `:` or value). This breaks our structured
PR parsing contract and results in empty `body/changelog` (or parse errors) despite an apparently
correct prompt.

This ticket captures current state, repro scripts, and the recommended next steps to fix the root
cause (likely Gemini streaming / response assembly or model refusal/truncation).

## Key docs

- Bug report analysis: `analysis/01-bug-report-gemini-yaml-partial-output.md`
- Tasks: `tasks.md`
- Changelog: `changelog.md`


