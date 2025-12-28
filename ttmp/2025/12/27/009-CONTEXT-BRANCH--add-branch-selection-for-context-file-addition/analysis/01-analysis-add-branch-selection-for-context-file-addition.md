---
Title: 'Analysis: Add branch selection for context file addition'
Ticket: 009-CONTEXT-BRANCH
Status: active
Topics:
    - prescribe
    - cli
    - context
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-27T18:30:00.000000000-05:00
WhatFor: ""
WhenToUse: ""
---

# Analysis: Add branch selection for context file addition

## Executive Summary

The `prescribe context add` command currently only supports adding files from the current/source branch. However, the underlying git service (`GetFileContent`) already supports reading files from any branch or commit. This analysis documents the current limitation and proposes adding a `--branch` or `--from` flag to allow users to specify which branch/commit to read context files from.

## Current Implementation

### CLI Command

**Location**: `prescribe/cmd/prescribe/cmds/context/add.go`

The `context add` command supports two mutually exclusive modes:
- Adding a note: `prescribe context add --note "text"`
- Adding a file: `prescribe context add <file-path>`

**Current Usage**:
```bash
# Add file from current branch (only option)
prescribe context add src/config.yaml

# Add note
prescribe context add --note "This PR is part of Q1 improvements"
```

### Controller Implementation

**Location**: `prescribe/internal/controller/controller.go`

```go
// AddContextFile adds a file from the repository as context
func (c *Controller) AddContextFile(path string) error {
	// Get file content from current branch
	content, err := c.gitService.GetFileContent(c.data.SourceBranch, path)
	if err != nil {
		return fmt.Errorf("failed to get file content: %w", err)
	}

	tokens_ := tokens.Count(content)

	c.data.AddContextItem(domain.ContextItem{
		Type:    domain.ContextTypeFile,
		Path:    path,
		Content: content,
		Tokens:  tokens_,
	})

	return nil
}
```

**Limitation**: Always uses `c.data.SourceBranch` (current branch), cannot specify target/base branch.

### Git Service Support

**Location**: `prescribe/internal/git/git.go`

The git service already supports reading from any branch/commit:

```go
// GetFileContent returns the content of a file at a specific branch/commit
func (s *Service) GetFileContent(ref, filePath string) (string, error) {
	cmd := exec.Command("git", "show", fmt.Sprintf("%s:%s", ref, filePath))
	cmd.Dir = s.repoPath
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get file content: %w", err)
	}
	return string(output), nil
}
```

**Capability**: The `ref` parameter can be:
- Branch name (e.g., `main`, `feature-branch`)
- Commit SHA (e.g., `abc123def`)
- Tag name (e.g., `v1.0.0`)
- Any valid git reference

## Use Cases

### Use Case 1: Compare with Base Branch

When generating a PR description, users may want to include a file from the base/target branch to show what changed:

```bash
# Add config file from base branch for comparison
prescribe context add --from main src/config.yaml
```

### Use Case 2: Include Related Files from Different Branch

Users may want to include files from a related branch that's not part of the current PR:

```bash
# Add file from another feature branch
prescribe context add --from feature/auth-improvements src/auth/config.go
```

### Use Case 3: Historical Context

Users may want to include files from a specific commit for historical context:

```bash
# Add file from specific commit
prescribe context add --from abc123def src/legacy-code.go
```

## Proposed Solution

### Option 1: Add `--from` Flag (Recommended)

Add a `--from` flag that accepts a branch/commit reference:

```bash
# Default behavior (current branch)
prescribe context add src/config.yaml

# From target branch
prescribe context add --from main src/config.yaml

# From specific commit
prescribe context add --from abc123def src/config.yaml
```

**Implementation**:
1. Add `--from` flag to `ContextAddSettings` in `add.go`
2. Update `AddContextFile` to accept optional branch parameter
3. Default to `SourceBranch` if `--from` not specified (backward compatible)

### Option 2: Add `--branch` Flag

Similar to Option 1, but using `--branch` name:

```bash
prescribe context add --branch main src/config.yaml
```

**Consideration**: `--branch` might imply only branches, not commits. `--from` is more generic.

### Option 3: Positional Branch Argument

Add branch as optional second positional argument:

```bash
prescribe context add src/config.yaml main
```

**Consideration**: Less discoverable, harder to parse, conflicts with existing positional argument pattern.

## Implementation Plan

### Phase 1: Update Command Definition

**File**: `prescribe/cmd/prescribe/cmds/context/add.go`

```go
type ContextAddSettings struct {
	Note string `glazed.parameter:"note"`
	From string `glazed.parameter:"from"`  // New: branch/commit reference
}

// In NewContextAddCommand, add field:
fields.New(
	"from",
	fields.TypeString,
	fields.WithDefault(""),
	fields.WithHelp("Branch, commit, or tag to read file from (default: current branch)"),
),
```

### Phase 2: Update Controller Method

**File**: `prescribe/internal/controller/controller.go`

```go
// AddContextFile adds a file from the repository as context
func (c *Controller) AddContextFile(path string, branch string) error {
	// Use provided branch or default to source branch
	ref := branch
	if ref == "" {
		ref = c.data.SourceBranch
	}

	content, err := c.gitService.GetFileContent(ref, path)
	if err != nil {
		return fmt.Errorf("failed to get file content from %s: %w", ref, err)
	}

	tokens_ := tokens.Count(content)

	c.data.AddContextItem(domain.ContextItem{
		Type:    domain.ContextTypeFile,
		Path:    path,
		Content: content,
		Tokens:  tokens_,
	})

	return nil
}
```

**Breaking Change**: Method signature changes from `AddContextFile(path string)` to `AddContextFile(path string, branch string)`. Need to update all call sites.

### Phase 3: Update Command Handler

**File**: `prescribe/cmd/prescribe/cmds/context/add.go`

```go
func (c *ContextAddCommand) Run(ctx context.Context, parsedLayers *glazed_layers.ParsedLayers) error {
	// ... existing validation ...

	if settings.Note != "" {
		ctrl.AddContextNote(settings.Note)
		fmt.Printf("Added note to context\n")
	} else {
		branch := settings.From
		if err := ctrl.AddContextFile(defaultSettings.FilePath, branch); err != nil {
			return errors.Wrap(err, "failed to add file")
		}
		if branch != "" {
			fmt.Printf("Added file '%s' from '%s' to context\n", defaultSettings.FilePath, branch)
		} else {
			fmt.Printf("Added file '%s' to context\n", defaultSettings.FilePath)
		}
	}

	// ... rest of function ...
}
```

### Phase 4: Update Domain Model (Optional)

**File**: `prescribe/internal/domain/domain.go`

Consider adding branch information to `ContextItem`:

```go
type ContextItem struct {
	Type    ContextType
	Path    string
	Content string
	Tokens  int
	Branch  string  // New: branch/commit reference (optional)
}
```

**Benefit**: Preserves which branch the file came from in session data.

**Consideration**: May not be necessary if branch is only needed at add time.

## Design Decisions

### Decision 1: Flag Name

**Choice**: Use `--from` instead of `--branch`

**Rationale**:
- More generic (supports branches, commits, tags)
- Clearer intent ("from this reference")
- Consistent with git terminology (`git show ref:path`)

### Decision 2: Default Behavior

**Choice**: Default to `SourceBranch` if `--from` not specified

**Rationale**:
- Maintains backward compatibility
- Matches current behavior
- Most common use case is current branch

### Decision 3: Method Signature Change

**Choice**: Add `branch` parameter to `AddContextFile`

**Rationale**:
- Clean API (explicit parameter)
- Type-safe (string parameter)
- Easy to extend (could add more options later)

**Alternative Considered**: Keep single parameter, use `PRData.TargetBranch` as default. Rejected because:
- Less flexible (can't specify arbitrary branch)
- Unclear behavior (why target branch?)

### Decision 4: Store Branch in ContextItem

**Choice**: Optional enhancement, not required for MVP

**Rationale**:
- Branch is metadata, not core content
- Can be added later if needed
- Keeps initial implementation simple

## Testing Strategy

### Test Cases

1. **Default behavior** (backward compatibility):
   ```bash
   prescribe context add src/config.yaml
   # Should use SourceBranch
   ```

2. **Explicit current branch**:
   ```bash
   prescribe context add --from feature-branch src/config.yaml
   # Should use feature-branch
   ```

3. **Target branch**:
   ```bash
   prescribe context add --from main src/config.yaml
   # Should use main branch
   ```

4. **Invalid branch**:
   ```bash
   prescribe context add --from nonexistent src/config.yaml
   # Should return error
   ```

5. **File doesn't exist in branch**:
   ```bash
   prescribe context add --from main src/nonexistent.yaml
   # Should return error
   ```

## Open Questions

1. **Should we support relative refs?** (e.g., `HEAD~1`, `main~2`)
   - Git service already supports this via `git show`
   - Just need to document it

2. **Should we validate branch exists?**
   - Git service will error if invalid
   - Could add pre-validation for better UX

3. **Should we show branch in TUI?**
   - If storing branch in ContextItem, could display it
   - Low priority for MVP

4. **Should we support multiple files from same branch?**
   - Current implementation is per-file
   - Could add `--from` that applies to multiple files
   - Out of scope for initial implementation

## References

- **Current Implementation**: `prescribe/cmd/prescribe/cmds/context/add.go`
- **Controller Method**: `prescribe/internal/controller/controller.go` (AddContextFile)
- **Git Service**: `prescribe/internal/git/git.go` (GetFileContent)
- **Domain Model**: `prescribe/internal/domain/domain.go` (ContextItem)
