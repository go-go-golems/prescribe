---
Title: Code Review Issues Analysis
Ticket: 016-FIX-CODE-REVIEW
Status: active
Topics:
    - bugfix
    - code-quality
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Analysis of three code review issues: prompt preset loading, full_both file handling, and potential git.go issues
LastUpdated: 2025-12-28T17:19:50.843728702-05:00
WhatFor: Understanding the root causes and solution approaches for code review findings
WhenToUse: Reference when implementing fixes
---

# Code Review Issues Analysis

## Overview

This document analyzes three code review issues identified in the prescribe codebase:

1. **P2: Restore non-builtin prompt presets on session load** (`internal/session/session.go`)
2. **P2: Include both versions for full_both files** (`internal/api/prompt.go`)
3. **P1: Potential issue in `internal/git/git.go`** (to be investigated)

## Issue 1: Restore Non-Builtin Prompt Presets on Session Load

### Problem Statement

**Location:** `internal/session/session.go:230-238`

**Issue:** Session loading only attempts to resolve `s.Prompt.Preset` against built-in presets. Project/global presets saved in `session.yaml` will never be found, and the prompt silently falls back to whatever was already in PRData (usually the default). In repos that use custom prompt presets, reopening a session will change the generation behavior without warning.

### Current Implementation

```go
// Apply prompt
if s.Prompt.Preset != "" {
    // Find and apply preset
    builtins := domain.GetBuiltinPresets()
    for _, preset := range builtins {
        if preset.ID == s.Prompt.Preset {
            data.SetPrompt(preset.Template, &preset)
            return nil
        }
    }
}
```

### Root Cause

The `Session.ApplyToPRData()` method only checks built-in presets. It doesn't check:
- Project presets (`.pr-builder/prompts/` in repo root)
- Global presets (`~/.pr-builder/prompts/`)

### Reference Implementation

The correct pattern exists in `internal/controller/controller.go:176-209` in the `LoadPromptPreset()` method:

```go
func (c *Controller) LoadPromptPreset(presetID string) error {
    // Check built-in presets
    builtins := domain.GetBuiltinPresets()
    for _, preset := range builtins {
        if preset.ID == presetID {
            c.data.SetPrompt(preset.Template, &preset)
            return nil
        }
    }

    // Check project presets
    projectPresets, err := c.LoadProjectPresets()
    if err == nil {
        for _, preset := range projectPresets {
            if preset.ID == presetID {
                c.data.SetPrompt(preset.Template, &preset)
                return nil
            }
        }
    }

    // Check global presets
    globalPresets, err := c.LoadGlobalPresets()
    if err == nil {
        for _, preset := range globalPresets {
            if preset.ID == presetID {
                c.data.SetPrompt(preset.Template, &preset)
                return nil
            }
        }
    }

    return fmt.Errorf("preset not found: %s", presetID)
}
```

### Solution Approach

1. **Option A: Add Controller dependency to Session** (Recommended)
   - Pass a `Controller` instance to `ApplyToPRData()`
   - Call `controller.LoadPromptPreset(s.Prompt.Preset)` instead of manual lookup
   - Pros: Reuses existing logic, consistent behavior
   - Cons: Adds dependency, requires refactoring call sites

2. **Option B: Extract preset loading logic**
   - Create a shared function `ResolvePromptPreset(presetID string, repoPath string) (*domain.PromptPreset, error)`
   - Use it in both `Controller.LoadPromptPreset()` and `Session.ApplyToPRData()`
   - Pros: No dependency injection needed
   - Cons: Need to pass `repoPath` to session methods

3. **Option C: Duplicate logic in Session**
   - Copy the preset loading logic from Controller to Session
   - Pros: No refactoring needed
   - Cons: Code duplication, maintenance burden

**Recommendation:** Option B - Extract shared logic into a utility function.

### Files to Modify

- `internal/session/session.go` - Update `ApplyToPRData()` method
- `internal/controller/controller.go` - Extract shared preset resolution logic
- Potentially create `internal/presets/resolver.go` for shared logic

### Testing Considerations

- Test session load with builtin preset
- Test session load with project preset
- Test session load with global preset
- Test session load with non-existent preset (should error or fallback gracefully)
- Test that session save/load roundtrip preserves custom presets

---

## Issue 2: Include Both Versions for full_both Files

### Problem Statement

**Location:** `internal/api/prompt.go:45-55`

**Issue:** For full-file mode, the code always prefers `FullAfter` and only falls back to `FullBefore` if it's empty, without checking `FileVersionBoth`. That means selecting "full_both" (or loading a session with that mode) still sends only one side to the model, and the token counts in the session UI can be higher than what the prompt actually contains. This occurs for any file that has both `FullBefore` and `FullAfter` populated.

### Current Implementation

```go
case domain.FileTypeFull:
    content := f.FullAfter
    if content == "" {
        content = f.FullBefore
    }
    if content == "" {
        content = f.Diff
    }
    if strings.TrimSpace(content) != "" {
        codeFiles = append(codeFiles, templateFile{Path: f.Path, Content: strings.TrimRight(content, "\n")})
    }
```

### Root Cause

The code doesn't check `f.Version` to determine if both versions should be included. When `f.Version == domain.FileVersionBoth`, both `FullBefore` and `FullAfter` should be included in the prompt.

### Domain Model

From `internal/domain/domain.go`:

```go
type FileChange struct {
    Path       string
    Included   bool
    Additions  int
    Deletions  int
    Tokens     int
    Type       FileType
    Version    FileVersion  // <-- This field exists but isn't checked
    Diff       string
    FullBefore string
    FullAfter  string
}

type FileVersion string

const (
    FileVersionBefore FileVersion = "before"
    FileVersionAfter  FileVersion = "after"
    FileVersionBoth   FileVersion = "both"
)
```

### Solution Approach

When `f.Version == domain.FileVersionBoth`, include both versions in the prompt. Options:

1. **Option A: Add two entries to codeFiles** (Recommended)
   - Add `FullBefore` as one entry with path suffix `:before`
   - Add `FullAfter` as another entry with path suffix `:after`
   - Pros: Clear separation, matches template expectations
   - Cons: Two entries per file

2. **Option B: Concatenate with separator**
   - Combine both versions with a clear separator (e.g., `\n\n--- BEFORE ---\n\n` / `\n\n--- AFTER ---\n\n`)
   - Pros: Single entry, preserves context
   - Cons: Less structured, harder to parse

3. **Option C: Use structured format**
   - Wrap in XML-like tags similar to diff format
   - Pros: Consistent with diff format
   - Cons: More complex

**Recommendation:** Option A - Add two entries with path suffixes.

### Implementation Details

```go
case domain.FileTypeFull:
    if f.Version == domain.FileVersionBoth {
        // Include both versions
        if strings.TrimSpace(f.FullBefore) != "" {
            codeFiles = append(codeFiles, templateFile{
                Path: f.Path + ":before",
                Content: strings.TrimRight(f.FullBefore, "\n"),
            })
        }
        if strings.TrimSpace(f.FullAfter) != "" {
            codeFiles = append(codeFiles, templateFile{
                Path: f.Path + ":after",
                Content: strings.TrimRight(f.FullAfter, "\n"),
            })
        }
    } else {
        // Single version logic (existing behavior)
        content := f.FullAfter
        if f.Version == domain.FileVersionBefore || content == "" {
            content = f.FullBefore
        }
        if content == "" {
            content = f.Diff
        }
        if strings.TrimSpace(content) != "" {
            codeFiles = append(codeFiles, templateFile{
                Path: f.Path,
                Content: strings.TrimRight(content, "\n"),
            })
        }
    }
```

### Files to Modify

- `internal/api/prompt.go` - Update `buildTemplateVars()` function

### Testing Considerations

- Test `FileVersionBoth` includes both versions
- Test `FileVersionBefore` includes only before
- Test `FileVersionAfter` includes only after
- Test fallback to diff when full content is empty
- Verify token counts match actual prompt content

### Related Code

- `internal/session/session.go:88-95` - Session loading handles `FileVersionBoth`
- `internal/export/context.go` - May need updates if export format changes

---

## Issue 3: Potential Issue in internal/git/git.go

### Investigation Needed

The code review mentioned a P1 issue in `internal/git/git.go`, but no specific details were provided. Reviewing the file shows it's a straightforward git wrapper with no obvious bugs, but we should:

1. Check for error handling issues
2. Verify command execution safety
3. Check for race conditions or concurrency issues
4. Verify path handling (security concerns)

### Areas to Review

1. **Error handling in `GetFileContent()`** - Returns empty string on error, which might mask issues
2. **Command injection** - All commands use `exec.Command()` with proper arguments, looks safe
3. **Path validation** - No explicit validation that paths don't escape repo root
4. **Concurrent access** - No locking, but Service is typically used per-repo

### Action Items

- Review git.go with focus on security and error handling
- Check if there are any known issues or TODOs
- Verify error propagation is appropriate

---

## Implementation Plan

### Phase 1: Fix Prompt Preset Loading
1. Extract preset resolution logic to shared utility
2. Update `Session.ApplyToPRData()` to use shared logic
3. Add tests for all preset types
4. Commit changes

### Phase 2: Fix full_both File Handling
1. Update `buildTemplateVars()` to check `FileVersion`
2. Implement both-versions inclusion logic
3. Add tests for all version types
4. Commit changes

### Phase 3: Review git.go
1. Review file for security/error handling issues
2. Document findings
3. Fix any issues found
4. Commit changes

### Phase 4: Integration Testing
1. Test session save/load roundtrip with custom presets
2. Test full_both mode end-to-end
3. Verify token counts are accurate
4. Commit final changes

---

## Related Files

- `internal/session/session.go` - Session loading logic
- `internal/controller/controller.go` - Controller preset loading (reference implementation)
- `internal/api/prompt.go` - Prompt template rendering
- `internal/domain/domain.go` - Domain model definitions
- `internal/git/git.go` - Git operations (to be reviewed)
