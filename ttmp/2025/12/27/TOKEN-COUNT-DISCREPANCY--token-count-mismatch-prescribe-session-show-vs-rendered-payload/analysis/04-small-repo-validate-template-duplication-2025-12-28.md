---
Title: 'Small repo validation: template duplication balloons rendered payload (2025-12-28)'
Ticket: TOKEN-COUNT-DISCREPANCY
Status: active
Topics:
    - prescribe
    - tokenization
    - prompts
    - debugging
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/prompts/assets/create-pull-request.yaml
      Note: Default prompt template contains a second `{{ template "context" . }}` under `.bracket`.
    - Path: internal/api/prompt.go
      Note: `buildTemplateVars` sets `.bracket=true` by default.
    - Path: ../scripts/repro-small-repo-balloon.sh
      Note: Script used to reproduce and quantify duplication on a tiny repo.
ExternalSources: []
Summary: On a tiny test repo, removing the bracketed second context render halves rendered user tokens and flips a stable duplication marker from 2 → 1.
LastUpdated: 2025-12-28T00:00:00Z
WhatFor: Provide concrete evidence that the default prompt template duplicates the context and materially inflates rendered token counts.
WhenToUse: When deciding whether to change `.bracket` defaults, adjust token-count semantics, or document expected behavior.
---

# Small repo validation: template duplication balloons rendered payload (2025-12-28)

## Setup

We use the existing smoke-test repo generator (`prescribe/test-scripts/setup-test-repo.sh`) to create a tiny repo with a feature branch and a small diff.

We then run two exports:
- **Default prompt**: current behavior (`.bracket=true` in template vars; template renders `context` twice)
- **No-dup prompt**: same prompt but with the bracketed duplication block removed (by patching a copy of session.yaml)

Script:
- `scripts/repro-small-repo-balloon.sh`

## Results (script stdout)

```text
session_show.token_count=861
dup_marker="The description of the pull request is:" default_count=2 nodup_count=1

DEFAULT rendered payload token counts: system=98 user=2511 total=2609
DEFAULT rendered xml export token count: 2733

NO-DUP rendered payload token counts: system=98 user=1255 total=1353
NO-DUP rendered xml export token count: 1477
```

## Findings

- **H1 confirmed**: the default rendered user prompt contains the `context` block twice (marker count 2).
- Removing the duplication reduces rendered **user tokens by 1256** (2511 → 1255), and reduces the XML envelope total similarly (2733 → 1477).
- `session_show.token_count` remains **861** in both cases because it counts only session context elements (diffs + additional context), not the prompt scaffolding.

## Implication

The “ballooning” is primarily explained by **template duplication** plus prompt scaffolding. It is not a tokenizer mismatch (confirmed previously by same-bytes comparisons on the big repro).

Next decision is product semantics:
- If duplication is intentional prompting strategy, document `session show token_count` as **context-only** and treat rendered payload tokens as a separate number.
- If duplication is not intended, consider changing defaults (e.g., `.bracket=false` or remove the second render) and/or making it configurable.


