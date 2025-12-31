---
Title: Task planning context refresh
Ticket: PR-CREATION
Status: active
Topics:
    - cli
    - git
    - prescribe
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-29T12:05:00.000000000-05:00
WhatFor: ""
WhenToUse: ""
---

# Task planning context refresh

This document confirms the exact docmgr commands and workflows I'll use when creating tasks from the analysis and clarification answers.

## docmgr task commands

### List tasks
```bash
docmgr task list --ticket PR-CREATION
```

### Add a task
```bash
docmgr task add --ticket PR-CREATION --text "Task description here"
```

### Check off completed tasks
```bash
docmgr task check --ticket PR-CREATION --id 1,2,3
```

### Edit a task
```bash
docmgr task edit --ticket PR-CREATION --id 1 --text "Updated task description"
```

### Remove a task
```bash
docmgr task remove --ticket PR-CREATION --id 1
```

### Uncheck a task (mark as not done)
```bash
docmgr task uncheck --ticket PR-CREATION --id 1
```

## Workflow pattern for task creation

1. **Extract work items** from analysis doc + clarification answers
2. **Add tasks** using `docmgr task add` for each concrete work item
3. **Relate files** to tasks using `docmgr doc relate` when tasks involve specific files
4. **Update changelog** after adding tasks: `docmgr changelog update --ticket PR-CREATION --entry "Added tasks from analysis and clarifications"`
5. **Update diary** with the task creation step

## Diary workflow (confirmed)

- Use absolute paths in `--file-note` flags
- Include commit hashes when code changes are involved
- Follow the structured format: What I did, Why, What worked, What didn't work, What I learned, etc.
- Update changelog after each significant step

## Git commit workflow (confirmed)

- Stage specific files: `git add path/to/file.go`
- Commit with message: `git commit -m "Short summary"`
- Get commit hash: `git rev-parse HEAD` (for diary entries)
- Never commit noise (node_modules, build artifacts, logs, etc.)

## Related workflow commands

### Update changelog
```bash
docmgr changelog update --ticket PR-CREATION \
  --entry "Description of change" \
  --file-note "/abs/path/to/file.go:Why this file matters"
```

### Relate files to docs
```bash
docmgr doc relate --ticket PR-CREATION \
  --doc-type reference \
  --file-note "/abs/path/to/file.go:Why this file matters"
```

### Update metadata
```bash
docmgr meta update --ticket PR-CREATION --field Status --value active
```

