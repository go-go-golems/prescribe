---
Title: 'Playbook: Write a great PR description from scratch (Go repo + many commits)'
Ticket: GENERATE-LLM-INFERENCE
Status: active
Topics:
    - prescribe
    - geppetto
    - inference
    - streaming
    - templating
    - parsing
DocType: playbook
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/prescribe/cmds/generate.go
      Note: Documents how to use generate flags (export/stream/output-file)
    - Path: pkg/doc/topics/02-how-to-generate-pr-description.md
      Note: End-to-end topic referenced by this playbook
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-27T20:21:06.249976688-05:00
WhatFor: ""
WhenToUse: ""
---


# Playbook: Write a great PR description from scratch (Go repo + many commits)

## Purpose

Create a high-quality PR description for a Go codebase with many commits by:
- understanding the change set (commits + diff + tests),
- curating the context that should be described,
- generating a structured PR YAML (title/body/changelog/release notes) using `prescribe`,
- and doing a final human editing pass.

## Environment Assumptions

- You are in a git repository with a feature branch checked out.
- You know (or can determine) the target branch (typically `main` or `master`).
- You have `go` installed for running tests.
- You have `prescribe` available (either built from this repo or installed).
- For inference (non-export): you have provider/model config available via the Geppetto StepSettings flags supported by `prescribe generate`.

## Commands

This playbook is split into phases. Run the commands from your repo root unless specified.

```bash
# Set target branch explicitly to avoid surprises
TARGET=main

# 1) Quick sanity: what changed?
git status --porcelain && git rev-parse --abbrev-ref HEAD && git log --oneline --decorate -n 20

# 2) Review commits against target (good for “many commits” PRs)
git fetch --all --prune && git log --oneline --no-decorate "$TARGET..HEAD"

# 3) Review the diff and stats (helps decide what to include/exclude)
git diff --stat "$TARGET...HEAD" && git diff "$TARGET...HEAD" | sed -n '1,120p'

# 4) Run Go checks so your PR description can honestly say what was tested
go test ./... -count=1
# Optional (if relevant):
# go test ./... -race -count=1
# go test ./... -run TestSomething -count=1
```

### Use `prescribe` to curate context and generate output

If you don’t already have a session:

```bash
prescribe -r . -t "$TARGET" session init --save
prescribe -r . -t "$TARGET" session show
```

#### Apply filters (reduce noise)

Start by hiding obvious noise (generated/vendor/build artifacts), then narrow further.

```bash
# Examples only — adjust to your repo
prescribe -r . -t "$TARGET" filter add --name "Hide vendor/generated" \
  --exclude "**/vendor/**" \
  --exclude "**/*generated*" \
  --exclude "**/*.pb.go"

prescribe -r . -t "$TARGET" session show
```

#### Ensure inclusion matches what you want to describe

Remember: filters control what you *see*, inclusion controls what gets sent to generation.

```bash
# Toggle a file in/out by path
prescribe -r . -t "$TARGET" file toggle path/to/file.go
```

#### Add context notes/files (optional but powerful)

Use notes for intent/constraints, and files for contracts/specs.

```bash
prescribe -r . -t "$TARGET" context add --note "Intent: ..."
prescribe -r . -t "$TARGET" context add --note "Constraints: ..."
prescribe -r . -t "$TARGET" context add README.md
```

#### (Recommended) Inspect what will be sent (no inference)

```bash
# Canonical generation context blob (xml is default)
prescribe -r . -t "$TARGET" generate --export-context --separator xml --output-file /tmp/prescribe-context.xml

# Rendered LLM payload (system+user), no inference
prescribe -r . -t "$TARGET" generate --export-rendered --separator xml --output-file /tmp/prescribe-rendered.xml
```

#### Generate (non-streaming)

```bash
prescribe -r . -t "$TARGET" generate --output-file /tmp/pr.yaml
```

#### Generate (streaming stdio)

Use streaming when you want responsiveness. Streaming events + a final parsed summary print to stderr; the final description still goes to stdout or `--output-file`.

```bash
prescribe -r . -t "$TARGET" generate --stream --output-file /tmp/pr.yaml
```

### Final human editing pass (what to check)

- **Title**: should summarize the “why/what” in one line; avoid trivialities (“update imports”).
- **Body**: focus on behavior/architecture changes, invariants, migration steps, and testing performed.
- **Changelog**: single-line, present tense, technical.
- **Release notes**: user-facing impact, any breaking changes or rollout notes.

## Exit Criteria

- You have a PR description (or YAML PR payload) that:
  - accurately reflects the actual diff and commit intent,
  - includes what was tested (or explicitly notes what wasn’t),
  - calls out risks/rollout notes when relevant,
  - is readable by someone who didn’t author the changes.
- `go test ./...` is green (or failures are understood and documented).

## Notes

- For “many commits” PRs, a good description usually has a short “commit narrative” first: group commits into 2–5 themes and describe each theme.
- If the repo has lots of mechanical changes, aggressively filter/hide noise and include only the meaningful files in the generation context.
- If generation output is not valid YAML, use the `--stream` run’s **parsed PR data summary** to see what the parser could (or could not) extract.
