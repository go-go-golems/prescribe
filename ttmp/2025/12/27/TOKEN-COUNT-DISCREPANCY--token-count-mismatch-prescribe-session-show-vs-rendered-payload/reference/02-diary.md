---
Title: Diary
Ticket: TOKEN-COUNT-DISCREPANCY
Status: active
Topics:
    - prescribe
    - tokenization
    - debugging
    - cli
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Implementation diary for debugging tools to diagnose token count discrepancies (session show vs rendered payload).
LastUpdated: 2025-12-28T00:00:00Z
WhatFor: Track step-by-step progress building token-count debugging tools (verbose breakdown, XML-ish posthoc counter, rendered payload token counts).
WhenToUse: Read this first when resuming work on TOKEN-COUNT-DISCREPANCY.
---

# Diary

## Goal

Build practical debugging tooling so we can precisely answer “where did the tokens go?” across:
- `session show token_count` (per-file + additional context budgeting), and
- the exported/rendered payload (system/user prompts + XML-ish envelope).

## Step 1: Tasking + diary setup

This step formalized the debugging tools as explicit ticket tasks and created a dedicated diary document to capture progress and failures as we implement. The intent is to keep the investigation reproducible: once the tooling exists, reproducing and explaining the mismatch should be a matter of running a few commands and comparing breakdown outputs.

### What I did
- Updated ticket `tasks.md` with three concrete debugging-tool deliverables:
  - verbose per-element session token breakdown
  - XML-ish posthoc token counting utility
  - rendered payload token counts behind a flag
- Created this diary doc: `reference/02-diary.md`

### What worked
- N/A (setup step)

### What didn't work
- N/A

### What I learned
- N/A

### What was tricky to build
- N/A

### What warrants a second pair of eyes
- N/A

### What should be done in the future
- N/A

### Code review instructions
- Start in `tasks.md` and verify the deliverables match the ticket goals.

## Step 2: Add `session token-count` breakdown command

This step added a dedicated CLI command to emit a machine-readable per-element token breakdown for the current session context. The key design choice is to show both the “stored” counts (what `session show token_count` currently sums) and an “effective” recomputed count (trimmed + best-effort content selection) to spot drift from how the generation context is actually assembled.

**Commit (code):** e39d6ea38ca4715f53344b08491d7422caadcae6 — "prescribe: add session token-count breakdown"

### What I did
- Added `prescribe session token-count`:
  - per-file rows (visible by default; can include filtered + non-included with flags)
  - per-additional-context rows
  - a summary `total` row with `stored_total` vs `effective_total`
- Included tokenizer encoding name (`tokens.EncodingName()`) in output.

### What worked
- `go test ./...` passes after adding the new command.

### What was tricky to build
- Defining a useful “effective” count without pretending it perfectly matches the prompt template rendering (we use trimmed content + best-effort fallback selection to help debug drift).

### What warrants a second pair of eyes
- Whether `effectiveFileContent` should treat `full_both` as two separately-formatted blocks rather than a concatenation with `\n` (this only impacts diagnostics, but should match real formatting as closely as feasible).

### Code review instructions
- Start at `cmd/prescribe/cmds/session/token_count.go`.
- Run:
  - `prescribe session token-count --output json`
  - `prescribe session token-count --include-filtered --all --output json`

## Step 3: Add `generate --print-rendered-token-count` flag

This step added a small but high-signal debug flag to `prescribe generate` that prints token counts for the **rendered** LLM payload (system + user). This makes it easy to compare `session show token_count` (preflight budgeting) with the exact rendered prompt strings we will send to inference, without having to export a file or use an external counter.

**Commit (code):** ae455d6d7ff12dc0de28a8157ffc69ee4445a3cc — "prescribe: optionally print rendered payload token counts"

### What I did
- Added `--print-rendered-token-count` to `prescribe generate`.
- When enabled, prints to stderr:
  - encoding name (`PRESCRIBE_TOKEN_ENCODING` / `tokens.EncodingName()`)
  - system prompt token count
  - user prompt token count
  - total (system + user)
  - token count of the exported envelope for the selected `--separator` (best-effort)

### What worked
- `go test ./...` passes.

### What warrants a second pair of eyes
- Whether we want to always print both “raw system/user” and “export envelope” counts, or gate the envelope count to export modes only (right now it prints best-effort in both paths).

## Step 4: Add `tokens count-xml` post-hoc counter

This step added a best-effort utility command to analyze exported “XML-ish” files (both context export and rendered payload export) and compute token counts per section. The goal is not perfect XML parsing—just enough structure to answer questions like “how many tokens are in `<llm_payload>` vs `<files>`?” and “which `<file>` blocks are the biggest?” using the same tokenizer/encoding as prescribe.

**Commit (code):** 9b796be270ac3352abbeadfbeefae3b02fa1814a — "prescribe: add tokens count-xml utility"

### What I did
- Added `prescribe tokens count-xml --file /path/to/export.xml` which emits rows for:
  - entire document total
  - common top-level-ish sections (`branches`, `commits`, `prompt`, `files`, `context`, `llm_payload`)
  - `system` / `user` CDATA content (if present)
  - optional per-`<file>` and per-`<item>` breakdowns (best-effort; enabled by default)

### What worked
- `go test ./...` passes.

### What was tricky to build
- Keeping the parsing logic intentionally dumb but still useful (simple tag matching + attribute extraction; no strict XML requirements).

### Code review instructions
- Start at `cmd/prescribe/cmds/tokens/count_xml.go`.
- Example usage:
  - `prescribe tokens count-xml --file /tmp/prescribe-rendered.xml --output json`

## Step 5: Repro script plan (low-noise) + first run

This step adds a single low-noise script that recreates the “mismatch” setup from the ticket and then runs *only* the debug commands we need: session breakdown, rendered payload token counts, export files, and post-hoc section counts. The goal is to quickly answer whether the discrepancy is a **tokenizer mismatch** (different encoding/codecs) or a **text mismatch** (e.g. context duplication + envelope overhead).

### Debugging plan (updated)
- Run the repro script to generate deterministic artifacts (session JSON + exported XML).
- Compare:
  - `session show token_count` vs `session token-count stored_total` (should match)
  - `generate --print-rendered-token-count` (system/user/total) vs `tokens count-xml` section counts (`<llm_payload>`, `<system>`, `<user>`)
  - `prescribe tokens count-xml` document total vs `pinocchio tokens count` on the *exact same bytes* (rendered.xml)
- If prescribe + pinocchio match on rendered.xml with the same codec, the discrepancy is **not** tokenizer mismatch.

### What I did
- Added script: `scripts/repro-token-count-discrepancy.sh`

### What worked
- The repro run shows `pinocchio` and `prescribe` match exactly when counting the *same bytes* with `cl100k_base`:
  - `rendered.xml` token count: **242,372** (both tools)

### What I learned
- The discrepancy is **not** a tokenizer mismatch; it’s a **different-input-text** mismatch:
  - `session show token_count`: **128,798**
  - rendered payload (system+user): **242,194**
  - rendered XML export envelope: **242,372**

### What should be done in the future
- Use this script + the existing tools to quantify whether the default template is duplicating context (e.g. `.bracket=true` rendering the context block twice), and how much token overhead comes from scaffolding vs diff/context content.

### Code review instructions
- Read the script, then run it from anywhere:
  - `bash scripts/repro-token-count-discrepancy.sh`

## Step 6: Hypotheses for ballooning + plan to validate on a small test repo

This step documented concrete hypotheses for why the rendered system/user payload can be much larger than `session show token_count`. The main suspected culprit is **template duplication**: the default prompt renders the `context` template twice when `.bracket=true`, and prescribe sets `.bracket=true` by default. The next move is to validate this on a small synthetic repo so we can compare actual rendered outputs directly without huge files.

### What I did
- Added hypothesis doc: `analysis/03-hypotheses-why-rendered-prompt-balloons.md`

### What warrants a second pair of eyes
- Confirm we should treat context duplication as a “bug” vs “intended prompting strategy”; if it’s intended, we should rename/clarify semantics of `session show token_count` as *context-only*.


