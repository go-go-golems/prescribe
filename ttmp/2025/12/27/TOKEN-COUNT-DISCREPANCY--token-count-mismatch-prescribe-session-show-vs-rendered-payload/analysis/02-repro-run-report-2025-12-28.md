---
Title: 'Repro run report: token discrepancy (2025-12-28)'
Ticket: TOKEN-COUNT-DISCREPANCY
Status: active
Topics:
    - prescribe
    - tokenization
    - debugging
    - pinocchio
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/prescribe/cmds/session/token_count.go
      Note: Per-element session token breakdown command used in repro.
    - Path: cmd/prescribe/cmds/generate.go
      Note: `--print-rendered-token-count` output used in repro.
    - Path: cmd/prescribe/cmds/tokens/count_xml.go
      Note: Post-hoc XML-ish counter used in repro.
    - Path: internal/tokens/tokens.go
      Note: Tokenizer/encoding selection (cl100k_base by default).
    - Path: ../scripts/repro-token-count-discrepancy.sh
      Note: Single low-noise repro script that generates the numbers in this report.
ExternalSources: []
Summary: Repro run shows pinocchio and prescribe agree on token count for the exact same rendered.xml bytes; discrepancy is due to counting different text (session context-only vs rendered payload), not tokenizer mismatch.
LastUpdated: 2025-12-28T00:00:00Z
WhatFor: Provide a minimal, repeatable run and concrete numbers to explain the token count mismatch.
WhenToUse: When validating whether a token discrepancy is a tokenizer mismatch vs input-text mismatch.
---

# Repro run report: token discrepancy (2025-12-28)

## Goal

Answer: “Are the extra tokens in pinocchio due to a tokenizer/codec mismatch, or are we counting different input text?”

## Repro method (low-noise)

Script:
- `scripts/repro-token-count-discrepancy.sh`

It:
- initializes a deterministic session state (same filter + context adds as the bug report),
- exports `/tmp/...context.xml` and `/tmp/...rendered.xml`,
- runs prescribe’s debug tools (`session token-count`, `generate --print-rendered-token-count`, `tokens count-xml`),
- and runs `pinocchio tokens count` on the *exact same bytes* (rendered.xml).

## Results (from script stdout)

Run output summary:

```text
encoding(prescribe)=cl100k_base
session_show.token_count=128798
session_token_count.stored_total=128798 effective_total=128790 delta=8
context_xml.tokens=150160
rendered_xml.tokens=242372
Rendered payload token counts (encoding=cl100k_base): system=98 user=242096 total=242194
Rendered payload export token count (separator=xml): 242372
```

Pinocchio comparison (from script log):

```text
Model:
Codec: cl100k_base
Total tokens: 242372
```

## Findings

- **Not a tokenizer mismatch**:
  - `prescribe tokens count-xml` document total for `rendered.xml` is **242,372**.
  - `pinocchio tokens count --model '' --codec cl100k_base` on the *exact same bytes* is also **242,372**.
  - Conclusion: prescribe and pinocchio agree when given the same input + codec.

- **The discrepancy is “different text”**:
  - `session show token_count` (and `session token-count stored_total`) is **128,798** — this is a *context-only* preflight budget (included files + additional context items).
  - The **rendered payload** (system+user) totals **242,194**, and the XML export envelope totals **242,372**.
  - The gap is therefore explained by prompt/template framing (system/user instructions) + how the prompt template includes context (potentially duplicating it).

- **Sanity checks**:
  - `session show token_count` == `session token-count stored_total` (**128,798**) → stored accounting is consistent.
  - `rendered payload export token count` == `rendered_xml.tokens` (**242,372**) → our exported XML-ish counter matches the exporter output.

## What this implies (next narrowing step)

If we want to explain “where did the extra ~113k tokens come from?” precisely, the next step is to attribute rendered user-prompt tokens to *template structure*:
- confirm whether `.bracket=true` causes the context block to render twice (as suspected in the ticket bug report),
- quantify token share of “prompt scaffolding” vs “diff/context content” inside the rendered user prompt.

The tools added in this ticket make that follow-up straightforward (we can iterate by tweaking template vars / prompt presets and rerunning the script).


