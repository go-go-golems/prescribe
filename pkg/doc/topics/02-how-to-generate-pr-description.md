---
Title: End-to-End PR Description Workflow
Slug: how-to-generate-pr-description
Short: End-to-end guide for using prescribe (TUI or CLI) to generate a pull request description from a repo diff.
Topics:
- prescribe
- workflow
- pr
- tui
- cli
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---

# Prescribe: End-to-End PR Description Workflow

## Overview

`prescribe` builds a PR description from your repo’s current diff by letting you (a) narrow the set of relevant files with filters, (b) choose exactly which visible files are included in the generation context, and (c) generate a markdown description via `prescribe generate` or the interactive `prescribe tui`. This guide walks through the entire flow and shows both a recommended TUI workflow and an all-CLI workflow.

## Prerequisites

To generate a useful PR description, `prescribe` needs a git repository with a meaningful diff between your current branch and a target branch, plus a working tree where you can inspect the changed paths you want to describe.

- **You are in a git repo** (or pass `--repo /path/to/repo`)
- **You have a diff vs a target branch** (defaults to `origin/HEAD`’s default branch, then `main`, then `master`)
- **You know your “story”**: what should be included vs ignored (tests, docs, generated code, etc.)

## Step 1: Initialize or load the session (recommended)

A session is the on-disk state that makes your work repeatable: selected files, filters, and additional context are stored under the repo’s `.pr-builder/` directory, so subsequent runs reuse the same view of the PR.

### Initialize and persist a new session

```bash
# Initialize based on current git state and save it to the default path:
prescribe session init --save
```

### Inspect current session state

```bash
prescribe session show
```

Expected output fields (example):

```bash
source_branch  target_branch  total_files  visible_files  included_files  filtered_files  active_filters  token_count
...
```

## Step 2: Narrow the diff with filters (optional, but high leverage)

Filters let you shrink the “visible files” set before you decide what to include. This reduces noise, improves focus, and usually improves the generated description.

### Quick start: add a filter

```bash
prescribe filter add --name "Exclude docs" \
  --exclude "**/*.md" \
  --exclude "**/docs/**"
```

### Verify active filters

```bash
prescribe filter list
```

### Use filter presets (project/global)

Filter presets are named YAML files stored in:

- Project: `<repo>/.pr-builder/filters/*.yaml`
- Global: `~/.pr-builder/filters/*.yaml`

List presets:

```bash
prescribe filter preset list --all
```

Save a preset (project):

```bash
prescribe filter preset save --project --name "Exclude tests" \
  --exclude "**/*test*" \
  --exclude "**/*spec*"
```

Apply a preset into the current session (adds it as an active filter and saves the session):

```bash
prescribe filter preset apply exclude_tests.yaml
```

### Learn the glob semantics (important)

Filter rules use doublestar globs, and multiple `--include` patterns are ANDed. For a deep dive:

```bash
prescribe help filters-and-glob-syntax
```

## Step 3: Choose which files are included in generation

Filtering controls what you *see*; inclusion controls what gets sent into the generation context. `prescribe generate` requires **at least one included file**, so this step is the one that turns a curated view into a real input payload.

### Recommended: use the TUI to toggle inclusion

```bash
prescribe tui
```

The TUI requires a saved session (see Step 1). It loads the default session and lets you:
- toggle included/excluded files,
- add/remove filters,
- generate and copy the result.

### CLI-only: toggle inclusion by path

```bash
prescribe file toggle path/to/file.go
prescribe file toggle internal/controller/controller.go
```

Confirm you have included files:

```bash
prescribe session show
```

## Step 4: (Optional) Adjust the prompt used for generation

The prompt determines the shape of the final description (tone, structure, required sections). In the CLI, prompt selection is a generation-time override: you can provide custom prompt text or choose a prompt preset ID.

### Generate with a one-off prompt

```bash
prescribe generate --prompt "Write a concise PR description with a Summary, Changes, and Testing section."
```

### Generate with a prompt preset

```bash
prescribe generate --preset some-preset-id.yaml
```

Prompt preset locations (mirrors the controller implementation):

- Project: `<repo>/.pr-builder/prompts/*.yaml`
- Global: `~/.pr-builder/prompts/*.yaml`

Minimal prompt preset schema:

```yaml
name: My PR description style
description: Short, structured PR descriptions
template: |
  Write a PR description with:
  - Summary
  - Changes (bulleted)
  - Testing
```

## Step 5: Generate the PR description

Generation uses the current session state (filters + included files + context) to build a canonical request, validates that at least one file is included, then prints the result to stdout (or writes it to a file).

### Generate to stdout

```bash
prescribe generate
```

### Generate to a file

```bash
prescribe generate --output-file pr-description.md
```

### Export the exact generation context payload (no inference)

This is the “catter-style” export: it prints the **exact text blob** that `prescribe generate` would send to the model (prompt + included files + additional context), without running inference.

```bash
# Default separator is xml
prescribe generate --export-context

# Explicit separator selection
prescribe generate --export-context --separator xml
prescribe generate --export-context --separator markdown
prescribe generate --export-context --separator simple
prescribe generate --export-context --separator begin-end
prescribe generate --export-context --separator default

# Write export to a file
prescribe generate --export-context --separator xml --output-file context.xml
```

### Export the rendered LLM payload (no inference)

This renders the prompt template (pinocchio-style Go templates) and outputs the **exact** `(system,user)` payload
that would seed the LLM Turn, without running inference.

```bash
# Default separator is xml
prescribe generate --export-rendered

# Explicit separator selection
prescribe generate --export-rendered --separator xml
prescribe generate --export-rendered --separator markdown
prescribe generate --export-rendered --separator simple
prescribe generate --export-rendered --separator begin-end
prescribe generate --export-rendered --separator default

# Write export to a file
prescribe generate --export-rendered --separator xml --output-file rendered.xml
```

## Troubleshooting

When generation fails, it’s usually a session state issue (no included files, wrong branch session, over-filtering) rather than a “model problem”.

### “At least one file must be included”

This means your visible files are all marked excluded. Fix by toggling one file:

```bash
prescribe file toggle path/to/file.go
prescribe generate
```

### “Session source branch doesn’t match current branch”

Sessions are tied to a source branch. If you switched branches, initialize a new session:

```bash
prescribe session init --save
```

### “My filter hides everything”

Most often this is an AND trap with multiple include patterns. Use a single glob with `{alt1,...}` instead. See:

```bash
prescribe help filters-and-glob-syntax
```

## Reference: Where things live on disk

These locations are the durable “project configuration surface” of `prescribe`:

- **Session state**: `<repo>/.pr-builder/session.yaml`
- **Repo defaults (new session via TUI)**: `<repo>/.pr-builder/config.yaml`
- **Filter presets**:
  - `<repo>/.pr-builder/filters/*.yaml`
  - `~/.pr-builder/filters/*.yaml`
- **Prompt presets**:
  - `<repo>/.pr-builder/prompts/*.yaml`
  - `~/.pr-builder/prompts/*.yaml`


