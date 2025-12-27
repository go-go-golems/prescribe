---
Title: Filters and Glob Syntax
Slug: filters-and-glob-syntax
Short: Playbook for creating file filters in prescribe and writing correct doublestar glob patterns.
Topics:
- prescribe
- filters
- glob
- tui
- cli
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---

# Prescribe Filters and Glob Syntax

## Overview

Filters in `prescribe` let you hide (or show) changed files by applying **include/exclude rules** against each file’s repo-relative path. This is most useful to focus your PR description context on relevant code while excluding noise (generated files, docs, vendored code, tests, etc.). This playbook explains how to create and manage filters from the CLI/TUI and—most importantly—how the underlying **doublestar glob syntax** works so your patterns behave as intended.

## How filters work in prescribe

A filter is a named set of rules. Each rule has a **type** (`exclude` or `include`) and a **glob pattern** that is matched against the file path (using `doublestar.Match`, with `/` as the path separator).

### Rule evaluation (current behavior)

Rules are evaluated with “fail fast” logic:

- **Exclude rule**: if the file path matches the rule’s pattern, the file is filtered out.
- **Include rule**: if the file path does *not* match the rule’s pattern, the file is filtered out.

This has an important implication: **multiple include rules are combined with AND**, not OR. In other words, if you add multiple `--include` patterns, a file must match *all* of them to remain visible.

### Multiple filters (current behavior)

If you have multiple active filters, **all filters must pass** for a file to remain visible. Practically, that means multiple filters are also combined with AND.

### Invalid patterns

If a pattern is malformed, `prescribe` falls back to a simple substring check (effectively: “does the path contain this text?”). This can make broken glob patterns look like they “sort of work”, but it’s usually better to fix the glob.

## Creating and managing filters (CLI)

The `prescribe filter` command group manages filters in the current session. The common workflow is:

- create a filter (`filter add`)
- test patterns without applying them (`filter test`)
- list active filters (`filter list`)
- remove or clear filters (`filter remove`, `filter clear`)

### Add a filter

This creates a named filter with one or more patterns and saves it to the session:

```bash
prescribe filter add --name "Exclude docs" \
  --exclude "**/*.md" \
  --exclude "**/docs/**"
```

### Test a filter (recommended)

Use this to validate globs and understand their impact before you save/apply them:

```bash
prescribe filter test --name "Test exclude tests" \
  --exclude "**/*test*" \
  --exclude "**/*spec*"
```

### List filters and impact

```bash
prescribe filter list
```

### Avoid multi-include AND traps

Because include rules are ANDed, this is usually **not** what you want:

```bash
# Likely matches nothing (a file can’t be both *.go AND *.ts)
prescribe filter add --name "Only source (broken)" \
  --include "**/*.go" \
  --include "**/*.ts"
```

Instead, use a **single pattern** with doublestar alternatives (see “Glob syntax”):

```bash
prescribe filter add --name "Only source" \
  --include "**/*.{go,ts,js,py}"
```

## Filter presets and repo defaults

Filter presets let you save common filter definitions as YAML files and apply them later. Repo defaults allow you to automatically apply one or more presets when starting the TUI in a repo that does not yet have a `.pr-builder/session.yaml`.

### Filter preset locations

- Project: `<repo>/.pr-builder/filters/*.yaml`
- Global: `~/.pr-builder/filters/*.yaml`

### Minimal preset schema

```yaml
name: Exclude tests
description: Exclude common test files
rules:
  - type: exclude
    pattern: "**/*test*"
  - type: exclude
    pattern: "**/*spec*"
```

### Preset CLI commands

```bash
# List presets
prescribe filter preset list --all

# Save a preset
prescribe filter preset save --project --name "Exclude tests" \
  --exclude "**/*test*" --exclude "**/*spec*"

# Apply a preset into the current session
prescribe filter preset apply exclude_tests.yaml
```

### Repo defaults (TUI “new session” behavior)

Create `<repo>/.pr-builder/config.yaml`:

```yaml
defaults:
  filter_presets:
    - exclude_tests.yaml
```

When `prescribe tui` starts and `<repo>/.pr-builder/session.yaml` is missing, these defaults are applied (session state still wins when it exists).

## Creating and managing filters (TUI)

The TUI provides a filter management mode (and some quick-add presets) to add and remove filters interactively. The semantics are the same as CLI-defined filters, so the glob rules below apply unchanged.

## Glob syntax (doublestar) in prescribe

`prescribe` uses `github.com/bmatcuk/doublestar/v4` for matching patterns. Patterns are split on forward slash (`/`), and matching requires the pattern to match the *entire* path (not just a substring).

### Special terms

These are the supported pattern primitives:

- `*`: matches any sequence of **non-path-separators** (does not cross `/`)
- `?`: matches any single **non-path-separator** character
- `/**/`: matches zero or more **directories** (this is the “globstar” / `**` feature)
- `[class]`: character classes (see below)
- `{alt1,...}`: alternatives (one of the comma-separated alternatives must match)

Any special character can be escaped with a backslash (`\`).

### The `**` (globstar) rule you must remember

`**` only works as a full **path component**. That means:

- **Good**: `path/to/**/*.txt`
- **Not what you want**: `path/to/**.txt` (behaves like `path/to/*.txt`)

### Character classes

Character classes match exactly one character within a single path segment:

- `[abc123]` matches one character in the set
- `[a-z0-9]` matches one character in a range
- `[^class]` / `[!class]` negates the class

### Examples you can copy/paste

```bash
# All Markdown files anywhere
--exclude "**/*.md"

# Everything under any docs/ directory anywhere
--exclude "**/docs/**"

# Any path segment containing "test" or "spec"
--exclude "**/*test*"
--exclude "**/*spec*"

# Only these extensions anywhere (OR via {alts})
--include "**/*.{go,ts,js,py}"

# Only files directly under a directory (no recursion)
--include "internal/*"
```

## Usage examples

This section provides complete, copy/paste-ready filter recipes that are known to work with `prescribe`’s current filter semantics.

### Example: exclude tests + docs

```bash
prescribe filter add --name "Exclude tests+docs" \
  --exclude "**/*test*" \
  --exclude "**/*spec*" \
  --exclude "**/*.md" \
  --exclude "**/docs/**"
```

### Example: “only source files” (using `{...}` alternatives)

```bash
prescribe filter add --name "Only source" \
  --include "**/*.{go,ts,js,py}"
```

## Troubleshooting

When a filter behaves unexpectedly, the issue is almost always (a) AND semantics across include rules, or (b) a misunderstanding of how `*` vs `**` interacts with `/`. These quick checks help you converge quickly.

### “My include patterns match nothing”

This usually happens because include rules are ANDed. Put your OR logic into a **single glob** using `{alt1,...}`.

### “My `**` pattern doesn’t recurse”

Ensure `**` appears as its own path component, typically as `**/` (or `/**/` in the middle of a pattern), like `**/*.go` or `path/**/file.txt`.

### “My pattern seems to work even though it’s wrong”

If doublestar reports a bad pattern, `prescribe` falls back to substring matching. Treat this as a sign to fix the glob rather than relying on the fallback.

## References

The glob semantics documented here come directly from doublestar’s pattern specification and behavior.

- `https://pkg.go.dev/github.com/bmatcuk/doublestar/v4`


