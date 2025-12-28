---
Title: 'Analysis: Add git history with stat to context'
Ticket: 010-GIT-HISTORY
Status: active
Topics:
    - prescribe
    - cli
    - context
    - git-history
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-27T18:45:00.000000000-05:00
WhatFor: ""
WhenToUse: ""
---

# Analysis: Add git history with stat to context

## Executive Summary

Currently, `prescribe context add` supports adding files and notes as context. This analysis proposes adding the ability to include git commit history (with statistics) as additional context for PR description generation. This would allow users to provide commit messages, author information, and file change statistics to help LLMs understand the evolution of changes.

## Current Implementation

### Context Addition

**Location**: `prescribe/cmd/prescribe/cmds/context/add.go`

The `context add` command currently supports:
- Adding files: `prescribe context add <file-path>`
- Adding notes: `prescribe context add --note "text"`

**Context Types** (`prescribe/internal/domain/domain.go`):
```go
type ContextType string

const (
    ContextTypeFile ContextType = "file"
    ContextTypeNote ContextType = "note"
)
```

### Git Service Capabilities

**Location**: `prescribe/internal/git/git.go`

The git service provides:
- `GetDiff()`: Unified diff between branches
- `GetChangedFiles()`: List of changed files with stats
- `GetFileContent()`: File content at specific branch/commit
- `GetFileDiff()`: Diff for specific file

**Missing**: Methods to get commit history with statistics.

## Use Cases

### Use Case 1: Include Commit Messages

Users may want to include commit messages to provide context about the reasoning behind changes:

```bash
# Add commit history from feature branch
prescribe context add --git-history feature-branch --commits 10
```

### Use Case 2: Include Commit Statistics

Users may want to include file change statistics (additions/deletions) per commit:

```bash
# Add commit history with file stats
prescribe context add --git-history feature-branch --with-stat
```

### Use Case 3: Include Author Information

Users may want to include commit author information to understand who made changes:

```bash
# Add commit history with author info
prescribe context add --git-history feature-branch --with-author
```

### Use Case 4: Filtered Commit History

Users may want to include only commits that affect specific files:

```bash
# Add commit history for specific files
prescribe context add --git-history feature-branch --files "src/**/*.go"
```

## Proposed Solution

### Option 1: New Context Type `git-history` (Recommended)

Add a new context type specifically for git history:

```bash
# Add git history as context
prescribe context add --git-history <branch> [--commits N] [--with-stat] [--with-author] [--files PATTERN]
```

**Implementation**:
- Add `ContextTypeGitHistory` to domain model
- Store branch, commit count, and options in `ContextItem`
- Generate git history content on-demand or at add time
- Format using separator approach (XML, markdown, etc.)

### Option 2: Extend Existing Context Types

Add git history as a special case of `ContextTypeFile` or `ContextTypeNote`:

```bash
# Add git history as note-like context
prescribe context add --git-history <branch> --as-note
```

**Consideration**: Less structured, harder to distinguish from regular notes.

## Git History Format

### Commit History with Stats

Example output format (XML separator):

```xml
<context type="git-history" branch="feature-branch" commits="5">
<history>
  <commit hash="abc123def" author="John Doe" date="2025-12-27T10:30:00Z">
    <message>Add authentication middleware</message>
    <stats>
      <file path="src/auth/middleware.go" additions="45" deletions="12"/>
      <file path="src/auth/login.go" additions="23" deletions="5"/>
    </stats>
  </commit>
  <commit hash="xyz789ghi" author="Jane Smith" date="2025-12-26T14:20:00Z">
    <message>Fix login validation</message>
    <stats>
      <file path="src/auth/login.go" additions="8" deletions="3"/>
    </stats>
  </commit>
</history>
</context>
```

### Git Commands Needed

1. **Get commit list**: `git log --oneline <branch> -n <count>`
2. **Get commit details**: `git log --format="%H|%an|%ae|%ad|%s" --date=iso <branch> -n <count>`
3. **Get commit stats**: `git log --stat <branch> -n <count>`
4. **Get commit stats for files**: `git log --stat -- <files> <branch> -n <count>`

## Implementation Plan

### Phase 1: Git Service Extensions

**File**: `prescribe/internal/git/git.go`

Add methods to retrieve commit history:

```go
// CommitInfo represents a single commit with metadata
type CommitInfo struct {
    Hash      string
    Author    string
    Email     string
    Date      time.Time
    Message   string
    Stats     []FileStat
}

// FileStat represents file change statistics for a commit
type FileStat struct {
    Path      string
    Additions int
    Deletions int
}

// GetCommitHistory returns commit history for a branch
func (s *Service) GetCommitHistory(branch string, count int, withStat bool) ([]CommitInfo, error)

// GetCommitHistoryForFiles returns commit history filtered by file patterns
func (s *Service) GetCommitHistoryForFiles(branch string, patterns []string, count int, withStat bool) ([]CommitInfo, error)
```

### Phase 2: Domain Model Updates

**File**: `prescribe/internal/domain/domain.go`

Add new context type:

```go
type ContextType string

const (
    ContextTypeFile       ContextType = "file"
    ContextTypeNote       ContextType = "note"
    ContextTypeGitHistory ContextType = "git_history"  // New
)

type ContextItem struct {
    Type    ContextType
    Path    string
    Content string
    Tokens  int
    // New fields for git history
    GitHistoryBranch string   // Branch name
    GitHistoryCount  int      // Number of commits
    GitHistoryOptions map[string]interface{}  // Options (with-stat, with-author, files, etc.)
}
```

### Phase 3: CLI Command Updates

**File**: `prescribe/cmd/prescribe/cmds/context/add.go`

Add flags for git history:

```go
type ContextAddSettings struct {
    Note        string `glazed.parameter:"note"`
    GitHistory  string `glazed.parameter:"git-history"`      // Branch name
    Commits     int    `glazed.parameter:"commits"`          // Number of commits (default: 10)
    WithStat    bool   `glazed.parameter:"with-stat"`        // Include file stats
    WithAuthor  bool   `glazed.parameter:"with-author"`      // Include author info
    Files       []string `glazed.parameter:"files"`          // Filter by file patterns
}
```

### Phase 4: Controller Updates

**File**: `prescribe/internal/controller/controller.go`

Add method to add git history:

```go
// AddContextGitHistory adds git commit history as context
func (c *Controller) AddContextGitHistory(branch string, count int, withStat, withAuthor bool, filePatterns []string) error {
    commits, err := c.gitService.GetCommitHistory(branch, count, withStat)
    if err != nil {
        return fmt.Errorf("failed to get commit history: %w", err)
    }
    
    // Format commits using separator approach
    content := c.formatGitHistory(commits, withStat, withAuthor)
    
    tokens_ := tokens.Count(content)
    
    c.data.AddContextItem(domain.ContextItem{
        Type:              domain.ContextTypeGitHistory,
        Path:              branch,
        Content:           content,
        Tokens:            tokens_,
        GitHistoryBranch:  branch,
        GitHistoryCount:   count,
        GitHistoryOptions: map[string]interface{}{
            "with-stat":   withStat,
            "with-author": withAuthor,
            "files":       filePatterns,
        },
    })
    
    return nil
}
```

## Design Decisions

### Decision 1: Context Type vs Special File

**Choice**: New context type `ContextTypeGitHistory`

**Rationale**:
- Clear semantic distinction from files and notes
- Allows specialized formatting and handling
- Easier to filter/display in TUI
- Supports structured metadata (branch, count, options)

### Decision 2: On-Demand vs Cached Generation

**Choice**: Generate content at add time (cached)

**Rationale**:
- Token counting requires content
- Avoids regenerating on every use
- Can be refreshed if needed
- Simpler implementation

**Alternative Considered**: Generate on-demand. Rejected because:
- Token counting becomes complex
- Performance concerns for large histories
- Harder to preview in TUI

### Decision 3: Default Commit Count

**Choice**: Default to 10 commits

**Rationale**:
- Reasonable default for most PRs
- Configurable via `--commits` flag
- Balances context size vs completeness

### Decision 4: Stat Format

**Choice**: Include per-file stats in structured format

**Rationale**:
- Provides detailed change information
- Useful for LLM understanding
- Can be formatted using separator approach
- Optional via `--with-stat` flag

## Integration with Separator System

Git history should use the same separator approach as other content types:

- **XML**: Structured format with `<commit>`, `<stats>`, `<file>` tags
- **Markdown**: Human-readable format with headers and code blocks
- **Simple**: Plain text with clear delimiters

See ticket `008-GENERATE` for separator implementation details.

## Open Questions

1. **Should we support commit range?** (e.g., `--from-commit abc123 --to-commit xyz789`)
   - Useful for specific commit ranges
   - More complex to implement
   - Can be added later

2. **Should we support merge commits?** (e.g., `--no-merges`)
   - Filter out merge commits for cleaner history
   - Common git log option
   - Easy to add

3. **Should we support commit message filtering?** (e.g., `--grep "pattern"`)
   - Filter commits by message content
   - Useful for focused context
   - Can be added later

4. **How to handle very long commit histories?**
   - Token limit considerations
   - Truncation strategy
   - Warning when approaching limits

5. **Should git history be included in session files?**
   - Current session format supports context items
   - Git history is dynamic (may change)
   - Consider whether to store or regenerate

## References

- **Current Context Implementation**: `prescribe/cmd/prescribe/cmds/context/add.go`
- **Domain Model**: `prescribe/internal/domain/domain.go`
- **Git Service**: `prescribe/internal/git/git.go`
- **Separator System**: Ticket `008-GENERATE` analysis document
