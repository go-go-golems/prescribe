---
Title: 'PR Run Diary: draft PR description (~150k tokens)'
Ticket: GENERATE-LLM-INFERENCE
Status: active
Topics:
    - prescribe
    - geppetto
    - inference
    - streaming
    - templating
    - parsing
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-27T20:47:58-05:00
WhatFor: ""
WhenToUse: ""
---

# PR Run Diary: draft PR description (~150k tokens)

## Goal

Produce a PR description for the recent `prescribe generate` inference work (parsing + streaming), using the ticket playbook as the operational checklist. Keep the curated context token count **around 150k** by intentionally managing which diffs/full files/context files are included.

## Context

- Repo: `prescribe/` (this repo)
- We will generate a PR draft *locally* (no GitHub push/PR creation in this workflow).
- Token target: ~150k tokens as reported by `prescribe session show` (`token_count`).
- Token count is an approximate preflight metric (not provider-exact).

## Quick Reference

### Commands we’ll use (copy/paste)

```bash
TARGET=main
git rev-parse --abbrev-ref HEAD && git fetch --all --prune && git log --oneline "$TARGET..HEAD"
git diff --stat "$TARGET...HEAD"

prescribe -r . -t "$TARGET" session init --save
prescribe -r . -t "$TARGET" session show --output json

# Export-only inspection (no inference)
prescribe -r . -t "$TARGET" generate --export-context --separator xml --output-file /tmp/prescribe-context.xml
prescribe -r . -t "$TARGET" generate --export-rendered --separator xml --output-file /tmp/prescribe-rendered.xml

# Generation (may require provider/model flags + API keys)
prescribe -r . -t "$TARGET" generate --stream --output-file /tmp/pr.yaml
```

### Token-count tuning knobs (how we reach ~150k)

- **Exclude noise**:
  - apply filters for generated, vendored, large irrelevant docs, etc.
- **Inclusion vs visibility**:
  - filters change what’s visible; `file toggle` changes what is *included* in the request.
- **Increase tokens intentionally**:
  - add context files with `prescribe context add path/to/file.md` (large docs/specs/ADR) when the diff alone is too small.
  - switch important diffs to full files (TUI) if you need more surrounding code (and to increase token count).
- **Decrease tokens intentionally**:
  - toggle off low-signal files (generated code, lockfiles, large refactors not relevant to PR narrative).
  - avoid adding huge binary-ish files (PDF, images).

## Usage Examples

## Step 1: Establish PR scope and base branch

Record:
- base branch (`TARGET`)
- current branch
- commit list and diffstat

### Log

- **Timestamp**: 2025-12-27T20:35:11-05:00
- **Current branch**: `task/prescribe-import`
- **Base**: `origin/main`

**Correction:** The base for this run is `origin/main` (not a local commit SHA). The goal is to draft a PR as if we were opening it against the upstream default branch.

## Step 2: Create/refresh session and inspect token_count

Record:
- `token_count` from `prescribe session show`
- included/visible/filtered counts

### Log

Reinitialized the session against `origin/main` (correct base).

Commands:

```bash
git fetch --all --prune
/tmp/prescribe-self -r . -t origin/main session init --save
/tmp/prescribe-self -r . -t origin/main session show --output json
```

Observed session state:
- **target_branch**: `origin/main`
- **total_files / visible_files**: 224 / 224
- **included_files**: 224
- **token_count**: 331384

Decision (token target): the raw diff is **too large** (>150k), so we’ll reduce token_count by filtering/excluding low-signal paths and only keeping the subset relevant to the inference work we want to describe.
## Step 3: Tune token_count toward ~150k

Record each adjustment:
- what was included/excluded
- context files added
- resulting `token_count`

### Log

#### Step 3.1 — Start from origin/main baseline (too large)

- **Before**: token_count = 331,384 (224 files included)
- **Goal**: bring token_count down to ~150k by excluding low-signal/huge docs and “ticket archive” paths.

What I did:
- Applied an exclude-based filter (quoted globs for zsh):

```bash
/tmp/prescribe-self -r . -t origin/main filter add --name "Trim huge docs" \
  --exclude 'ttmp/**' \
  --exclude 'TUI-SCREENSHOTS*' \
  --exclude 'FILTER-*.md' \
  --exclude 'PLAYBOOK-*.md' \
  --exclude 'dev-diary.md' \
  --exclude 'PROJECT-SUMMARY.md' \
  --exclude 'TUI-DEMO.md' \
  --exclude '*.pdf'
```

Result:
- **After**: token_count = 101,289
- visible_files/included_files: 97/97
- filtered_files: 127

Decision: we overshot “too small” (101k), so we’ll *add back* tokens using **high-signal context files** rather than re-including huge low-signal diffs.

What would have helped me do a better job:
- A built-in `prescribe session show --tokens-by-file` (or similar) to see top token contributors in the diff.
- A “filter dry-run” view that shows “top N files removed” + token deltas per rule.

#### Step 3.2 — Add high-signal context to approach ~150k

Guiding principle: keep the diff focused; use context files to add *narrative/architecture* and *prompt contracts*.

Commands (context note + files):

```bash
/tmp/prescribe-self -r . -t origin/main context add --note "Goal: draft PR description for inference work (PR YAML: title/body/changelog/release_notes). Target ~150k tokens: exclude low-signal paths, then add only high-signal docs as context."
/tmp/prescribe-self -r . -t origin/main context add PROJECT-SUMMARY.md
/tmp/prescribe-self -r . -t origin/main context add README.md
/tmp/prescribe-self -r . -t origin/main context add ttmp/2025/12/28/GENERATE-LLM-INFERENCE--llm-inference-streaming-and-parsing/analysis/02-analysis-template-rendering-streaming-and-prdata-extraction.md
/tmp/prescribe-self -r . -t origin/main context add ttmp/2025/12/28/GENERATE-LLM-INFERENCE--llm-inference-streaming-and-parsing/analysis/03-analysis-tui-streaming-integration.md
/tmp/prescribe-self -r . -t origin/main context add ttmp/2025/12/28/GENERATE-LLM-INFERENCE--llm-inference-streaming-and-parsing/design-doc/01-design-guide-generation-pipeline-exporters-and-geppetto-stepsettings.md
```

Token progression (observed):
- 101,289 → 111,473

Then added additional internal diaries/playbooks to increase “why/how” coverage:

```bash
/tmp/prescribe-self -r . -t origin/main context add ttmp/2025/12/27/008-GENERATE--generate-pr-descriptions-using-ai-inference/reference/01-diary.md
/tmp/prescribe-self -r . -t origin/main context add ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/reference/02-diary.md
/tmp/prescribe-self -r . -t origin/main context add ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/playbooks/01-cli-testing-playbook.md
```

Token progression (observed):
- 111,473 → 132,242

Finally, added two larger context docs to get close to the target:

```bash
/tmp/prescribe-self -r . -t origin/main context add ttmp/2025/12/26/001-INITIAL-IMPORT--initial-import-transform-prescribe-into-go-go-golems-application/design/claude-session-TUI-simulation.md
/tmp/prescribe-self -r . -t origin/main context add ttmp/2025/12/27/011-TUI-UPDATE--tui-update-reconcile-initial-screen-design-with-implemented-tui-wiring/analysis/01-tui-wiring-analysis-screens-models-messages-cli-entrypoints.md
/tmp/prescribe-self -r . -t origin/main context add ttmp/2025/12/28/GENERATE-LLM-INFERENCE--llm-inference-streaming-and-parsing/playbook/01-playbook-write-a-great-pr-description-from-scratch-go-repo-many-commits.md
```

Final state:
- token_count = **146,878** (≈150k target)
- additional_context_items = 12

What would have helped me do a better job:
- A “token budget target” mode: `prescribe tune --target-tokens 150000` that suggests which files to include/exclude or add as context.
- A command to list current context items and their individual token contributions (so we can pick the best “bang for buck” without guesswork).

## Step 4: Generate PR draft + final structured summary

Record:
- whether inference ran successfully (provider config)
- output file path
- final PR YAML (title/body/changelog/release_notes)

### Log

#### Step 4.1 — Export-only inspection (works without API keys)

```bash
/tmp/prescribe-self -r . -t origin/main generate --export-context --separator xml --output-file /tmp/prescribe-context.xml
/tmp/prescribe-self -r . -t origin/main generate --export-rendered --separator xml --output-file /tmp/prescribe-rendered.xml
```

Result:
- Wrote `/tmp/prescribe-context.xml` (canonical request blob)
- Wrote `/tmp/prescribe-rendered.xml` (rendered system+user payload)

#### Step 4.2 — Inference attempt (failed: missing API key)

```bash
/tmp/prescribe-self -r . -t origin/main generate --stream --output-file /tmp/pr.yaml
```

Observed failure:
- OpenAI request returned **401 Unauthorized** (“You didn't provide an API key…”)
- No `/tmp/pr.yaml` generated

What would have helped me do a better job:
- A clear “provider config check” command like `prescribe doctor --ai` that validates StepSettings + env vars before running inference.
- A `--dry-run-inference` mode that fails fast with a friendly message listing which env vars/flags are missing.

#### Step 4.3 — Manual PR YAML draft (since inference is blocked)

Since inference requires an API key, I drafted the PR YAML manually below. This is the shape we’re aiming for once we can run `prescribe generate` end-to-end.

```yaml
title: Add streaming PR generation and structured output parsing
body: |
  Summary
  - Introduce robust parsing of the assistant’s YAML output into structured PR fields
    (title/body/changelog/release_notes), preferring the last fenced YAML block.
  - Add stdio streaming mode for `prescribe generate` so output/events stream to stderr and
    a deterministic parsed summary is printed at the end of streaming runs.

  Changes
  - Inference pipeline
    - Parse PR YAML output into a structured result type and store it on session state
      (`GeneratedPRData` + parse error string) alongside the raw assistant text.
    - Add `--stream` to run inference with an event router + Watermill sink and stream partial
      completions/events to the terminal.
  - Export/debug tooling
    - Keep export-only workflows (`--export-context`, `--export-rendered`) as stable debugging
      seams for prompt/context inspection.
  - Docs/playbooks
    - Update docs for generate options and add an end-to-end playbook for writing a good PR
      description in a Go repo with many commits.

  Testing
  - go test ./... -count=1
  - bash test/test-cli.sh
  - bash test/test-all.sh

  Notes
  - This branch contains a large number of commits across multiple tickets; review is easier if
    you start with generate/inference-related commits and then expand outward.
changelog: |
  Add streaming PR generation and structured PR YAML parsing
release_notes:
  title: Streaming PR description generation
  body: |
    Prescribe can now stream generation output to the terminal while still producing a final
    result, and it can parse the model’s structured YAML output into dedicated PR fields.
    This improves both responsiveness and the reliability of downstream tooling that wants
    to consume PR metadata programmatically.
```

## Related

- Ticket playbook: `playbook/01-playbook-write-a-great-pr-description-from-scratch-go-repo-many-commits.md`
- End-to-end guide: `pkg/doc/topics/02-how-to-generate-pr-description.md`
