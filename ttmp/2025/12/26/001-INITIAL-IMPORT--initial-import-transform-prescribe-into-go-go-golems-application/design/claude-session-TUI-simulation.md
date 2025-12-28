# PR Builder TUI - Technical Specification

## Table of Contents
1. [Overview](#overview)
2. [Screen Reference](#screen-reference)
3. [Action Reference](#action-reference)
4. [Data Schemas](#data-schemas)

---

## Overview

PR Builder is a CLI TUI application for generating pull request descriptions using LLMs. It allows users to:
- View and filter PR diffs
- Toggle file inclusion and replace diffs with full files
- Apply filters with glob patterns
- Customize prompts with presets
- Generate AI-powered PR descriptions

---

## Screen Reference

### 1. Main Screen

```
╔════════════════════════════════════════════════════════════════════════════╗
║                         PR DESCRIPTION GENERATOR                           ║
╠════════════════════════════════════════════════════════════════════════════╣
║                                                                            ║
║  Current Branch: feature/user-auth → main                                 ║
║  Files Changed: 2 (1 filtered out)                                         ║
║  Token Count: 2,146 tokens                                                 ║
║                                                                            ║
╠════════════════════════════════════════════════════════════════════════════╣
║  CHANGED FILES                                                             ║
╠════════════════════════════════════════════════════════════════════════════╣
║  [✓] src/auth/login.ts                                    +89 -12  (342t) ║
║  [✓] src/auth/middleware.ts                               +156 -3  (1.8k) ║
╠════════════════════════════════════════════════════════════════════════════╣
║  FILTERS: Exclude tests (1 file hidden) [H to view]                       ║
╠════════════════════════════════════════════════════════════════════════════╣
║  ADDITIONAL CONTEXT                                                        ║
╠════════════════════════════════════════════════════════════════════════════╣
║  No additional files or context added                                      ║
╠════════════════════════════════════════════════════════════════════════════╣
║  PROMPT TEMPLATE                                                           ║
╠════════════════════════════════════════════════════════════════════════════╣
║  "Generate a clear PR description with: summary of changes, motivation,   ║
║   key changes, testing notes, and breaking changes if any."               ║
╠════════════════════════════════════════════════════════════════════════════╣
║                                                                            ║
║  [C] Edit Context  [F] Edit Filters  [H] View Hidden  [E] Edit Prompt     ║
║  [A] Add Files  [T] Add Notes  [G] Generate  [S] Save  [L] Load  [Q] Quit ║
║                                                                            ║
╚════════════════════════════════════════════════════════════════════════════╝

> _
```

**Actions:**
- `C` - Edit Context (go to Edit Context Window)
- `F` - Edit Filters (go to Edit File Filters)
- `H` - View Hidden files (go to Filtered Out Files)
- `E` - Edit Prompt (go to Edit Prompt Template)
- `A` - Add Files (go to Add Files from Repo)
- `T` - Add Notes (go to Add Text Notes)
- `G` - Generate Description (go to Generate Screen)
- `S` - Save Session (show save dialog)
- `L` - Load Session (show load dialog)
- `Q` - Quit application

**Data Schema:**
```typescript
interface MainScreenData {
  branch: {
    source: string;
    target: string;
  };
  files: {
    total: number;
    visible: number;
    filtered: number;
  };
  tokenCount: number;
  changedFiles: Array<{
    path: string;
    included: boolean;
    additions: number;
    deletions: number;
    tokens: number;
    type: 'diff' | 'full_file';
    version?: 'before' | 'after' | 'both';
  }>;
  activeFilters: string[];
  additionalContext: Array<{
    type: 'file' | 'note';
    path?: string;
    content?: string;
  }>;
  promptTemplate: string;
}
```

---

### 2. Edit Context Window

```
╔════════════════════════════════════════════════════════════════════════════╗
║                        EDIT CONTEXT WINDOW                                 ║
╠════════════════════════════════════════════════════════════════════════════╣
║  FILES (↑↓ navigate, Space toggle, Enter view full)          2,847 tokens ║
╠═══════════════════════════════════╦════════════════════════════════════════╣
║                                   ║                                        ║
║  [✓] src/auth/login.ts            ║  @@ -23,7 +23,18 @@                   ║
║      +89 -12  (342t) [DIFF]       ║   export async function login(        ║
║                                   ║     email: string,                     ║
║  [✓] src/auth/middleware.ts       ║     password: string                   ║
║      +156 -3  (1.8k) [DIFF]       ║   ) {                                  ║
║                                   ║  +  // Validate input                  ║
║ ▶[✓] tests/auth.test.ts           ║  +  if (!email || !password) {        ║
║      (701t) [FULL:AFTER]          ║  +    throw new Error('Missing...');  ║
║                                   ║  +  }                                  ║
║                                   ║  +                                     ║
║                                   ║  +  // Hash password before compare    ║
║                                   ║  +  const hash = await bcrypt.hash... ║
║                                   ║     ...                                ║
║                                   ║                                        ║
║                                   ║  [Showing first 10 lines of full file] ║
║                                   ║                                        ║
╠═══════════════════════════════════╩════════════════════════════════════════╣
║                                                                            ║
║  ↑↓ Navigate  Space Toggle  Enter Full View  R Replace Options            ║
║  D Restore Diff  F Filter  A Add Other Files  Esc Back                    ║
║                                                                            ║
╚════════════════════════════════════════════════════════════════════════════╝
```

**Actions:**
- `↑/↓` - Navigate through files
- `Space` - Toggle file inclusion
- `Enter` - View full diff/file content
- `R` - Replace with full file (show Replace Dialog)
- `D` - Restore to diff (if currently showing full file)
- `F` - Filter files (go to Edit File Filters)
- `A` - Add other files (go to Add Files from Repo)
- `Esc` - Back to Main Screen

**Data Schema:**
```typescript
interface EditContextData {
  totalTokens: number;
  files: Array<{
    path: string;
    included: boolean;
    additions: number;
    deletions: number;
    tokens: number;
    type: 'diff' | 'full_file';
    version?: 'before' | 'after' | 'both';
    diffPreview: string; // First 10 lines
  }>;
  selectedIndex: number;
}
```

---

### 3. Replace with Full File Dialog

```
╔════════════════════════════════════════════════════════════════════════════╗
║                   REPLACE WITH FULL FILE                                   ║
╠════════════════════════════════════════════════════════════════════════════╣
║                                                                            ║
║  File: tests/auth.test.ts                                                  ║
║                                                                            ║
║  Select version to include:                                                ║
║                                                                            ║
║    ▶ [ ] Before (original version)                          ~650 tokens   ║
║      [ ] After (final version)                              ~701 tokens   ║
║      [✓] Both (before + after)                              ~1,351 tokens ║
║                                                                            ║
║  This will replace the diff (701t) with full file content.                ║
║                                                                            ║
╠════════════════════════════════════════════════════════════════════════════╣
║                                                                            ║
║  ↑↓ Select  Space Toggle  Enter Confirm  Esc Cancel                       ║
║                                                                            ║
╚════════════════════════════════════════════════════════════════════════════╝
```

**Actions:**
- `↑/↓` - Navigate options
- `Space` - Toggle selection
- `Enter` - Confirm and replace
- `Esc` - Cancel

**Data Schema:**
```typescript
interface ReplaceFileDialogData {
  filePath: string;
  currentTokens: number;
  options: Array<{
    id: 'before' | 'after' | 'both';
    label: string;
    tokens: number;
    selected: boolean;
  }>;
  selectedIndex: number;
}
```

---

### 4. Filtered Out Files View

```
╔════════════════════════════════════════════════════════════════════════════╗
║                         FILTERED OUT FILES                                 ║
╠════════════════════════════════════════════════════════════════════════════╣
║                                                                            ║
║  Filter: Exclude tests                                                     ║
║  1 file hidden from context                                                ║
║                                                                            ║
╠════════════════════════════════════════════════════════════════════════════╣
║                                                                            ║
║  [ ] tests/auth.test.ts                                   (701t) [FULL:AFTER] ║
║                                                                            ║
╠════════════════════════════════════════════════════════════════════════════╣
║                                                                            ║
║  Space Toggle to include  F Edit Filters  Esc Back                        ║
║                                                                            ║
╚════════════════════════════════════════════════════════════════════════════╝
```

**Actions:**
- `Space` - Toggle file to include (removes from filtered list)
- `F` - Edit Filters (go to Edit File Filters)
- `Esc` - Back to Main Screen

**Data Schema:**
```typescript
interface FilteredOutFilesData {
  activeFilter: string;
  files: Array<{
    path: string;
    tokens: number;
    type: 'diff' | 'full_file';
    version?: 'before' | 'after' | 'both';
  }>;
}
```

---

### 5. Edit File Filters Screen

```
╔════════════════════════════════════════════════════════════════════════════╗
║                          EDIT FILE FILTERS                                 ║
╠════════════════════════════════════════════════════════════════════════════╣
║                                                                            ║
║  Active Filters: Backend only                                              ║
║  Files Matched: 7 of 12 total changed files (5 filtered out)              ║
║                                                                            ║
╠════════════════════════════════════════════════════════════════════════════╣
║  QUICK FILTERS                                                             ║
╠════════════════════════════════════════════════════════════════════════════╣
║                                                                            ║
║    [ ] Exclude tests          (matches: tests/, **/*.test.*, **/*.spec.*) ║
║    [ ] Exclude docs           (matches: docs/, *.md, README*)             ║
║    [ ] Only Go files          (matches: **/*.go)                          ║
║    [ ] Only Python files      (matches: **/*.py)                          ║
║    [ ] Only TypeScript/JS     (matches: **/*.ts, **/*.tsx, **/*.js)       ║
║    [ ] Exclude config         (matches: *.json, *.yaml, *.toml, .*)       ║
║                                                                            ║
╠════════════════════════════════════════════════════════════════════════════╣
║  SAVED PRESETS                                                             ║
╠════════════════════════════════════════════════════════════════════════════╣
║                                                                            ║
║ ▶ [✓] Backend only                                              [PROJECT]  ║
║       Only backend Go code and proto files, excluding tests                ║
║       (4 rules, 7 files matched)                                           ║
║                                                                            ║
╠════════════════════════════════════════════════════════════════════════════╣
║  SESSION FILTERS (unsaved - will be lost on quit)                          ║
╠════════════════════════════════════════════════════════════════════════════╣
║                                                                            ║
║    [ ] Frontend focus                                            [TEMP] 💾  ║
║        React components and TypeScript only                                ║
║        (2 rules, 4 files matched)                                          ║
║                                                                            ║
║  [N] New Custom Filter                                                     ║
║                                                                            ║
╠════════════════════════════════════════════════════════════════════════════╣
║                                                                            ║
║  ↑↓ Navigate  Space Toggle  N New  E Edit  D Delete  W Save Preset        ║
║  Enter Apply  Esc Cancel                                                   ║
║                                                                            ║
╚════════════════════════════════════════════════════════════════════════════╝
```

**Actions:**
- `↑/↓` - Navigate filters
- `Space` - Toggle filter on/off
- `N` - New custom filter (go to Create Custom Filter)
- `E` - Edit selected filter (go to Create Custom Filter with data)
- `D` - Delete selected filter
- `W` - Save as preset (show Save Filter Preset dialog)
- `Enter` - Apply filters and return to previous screen
- `Esc` - Cancel and return to previous screen

**Data Schema:**
```typescript
interface EditFiltersData {
  activeFilters: string[];
  matchedFiles: number;
  totalFiles: number;
  filteredOutCount: number;
  quickFilters: Array<{
    id: string;
    name: string;
    patterns: string[];
    active: boolean;
  }>;
  savedPresets: Array<{
    id: string;
    name: string;
    description: string;
    location: 'project' | 'global';
    ruleCount: number;
    matchedFiles: number;
    active: boolean;
    filePath: string;
  }>;
  sessionFilters: Array<{
    id: string;
    name: string;
    description: string;
    ruleCount: number;
    matchedFiles: number;
    active: boolean;
    temporary: true;
  }>;
  selectedIndex: number;
}
```

---

### 6. Create/Edit Custom Filter Screen

```
╔════════════════════════════════════════════════════════════════════════════╗
║                       CREATE CUSTOM FILTER                                 ║
╠════════════════════════════════════════════════════════════════════════════╣
║                                                                            ║
║  Filter Name:                                                              ║
║  > Backend only_                                                           ║
║                                                                            ║
║  Description:                                                              ║
║  > Only backend Go code and proto files, excluding tests_                 ║
║                                                                            ║
╠════════════════════════════════════════════════════════════════════════════╣
║  RULES (evaluated in order, ↑↓ to reorder)                                ║
╠════════════════════════════════════════════════════════════════════════════╣
║                                                                            ║
║  1. [INCLUDE] src/backend/**/*                                             ║
║  2. [EXCLUDE] **/*.test.*                                                  ║
║  3. [INCLUDE] src/api/**/*.go                                              ║
║ ▶4. [INCLUDE] **/*.proto                                                   ║
║                                                                            ║
║  [+] Add Rule                                                              ║
║                                                                            ║
╠════════════════════════════════════════════════════════════════════════════╣
║  Preview: 7 files would match                                              ║
║    ✓ src/backend/auth.go                                                   ║
║    ✓ src/backend/db/connection.go                                          ║
║    ✗ src/backend/auth_test.go         (excluded by rule 2)                ║
║    ✓ src/api/handlers.go                                                   ║
║    ✓ src/proto/user.proto                                                  ║
║    ✗ frontend/app.tsx                   (no matching include rule)        ║
║                                                                            ║
╠════════════════════════════════════════════════════════════════════════════╣
║                                                                            ║
║  ↑↓ Navigate  Enter Edit  +Add  -Delete  Shift+↑↓ Reorder                 ║
║  S Save & Use  W Save as Preset  Esc Cancel                                ║
║                                                                            ║
╚════════════════════════════════════════════════════════════════════════════╝
```

**Actions:**
- `↑/↓` - Navigate rules
- `Enter` - Edit selected rule (go to Add Filter Rule)
- `+` - Add new rule (go to Add Filter Rule)
- `-` - Delete selected rule
- `Shift+↑/↓` - Reorder rules
- `S` - Save and use filter (return to previous screen)
- `W` - Save as preset (show Save Filter Preset dialog)
- `Esc` - Cancel

**Data Schema:**
```typescript
interface CreateFilterData {
  name: string;
  description: string;
  rules: Array<{
    id: string;
    order: number;
    type: 'include' | 'exclude';
    pattern: string;
  }>;
  preview: {
    matchedCount: number;
    examples: Array<{
      path: string;
      included: boolean;
      reason?: string;
    }>;
  };
  selectedRuleIndex: number;
  isEditing: boolean; // true if editing existing filter
  filterId?: string; // present if editing
}
```

---

### 7. Add Filter Rule Dialog

```
╔════════════════════════════════════════════════════════════════════════════╗
║                         ADD FILTER RULE                                    ║
╠════════════════════════════════════════════════════════════════════════════╣
║                                                                            ║
║  Rule Type:                                                                ║
║                                                                            ║
║  ▶ ( ) Include files matching pattern                                     ║
║    ( ) Exclude files matching pattern                                     ║
║                                                                            ║
╠════════════════════════════════════════════════════════════════════════════╣
║  Glob Pattern:                                                             ║
║                                                                            ║
║  > **/*.proto_                                                             ║
║                                                                            ║
╠════════════════════════════════════════════════════════════════════════════╣
║  Examples:                                                                 ║
║    src/**/*.go          - All Go files in src/ and subdirs                ║
║    **/*_test.py         - All Python test files anywhere                  ║
║    frontend/components/ - Everything in that directory                    ║
║    *.{json,yaml,toml}   - Config files with these extensions              ║
║    !vendor/             - Negation (exclude vendor/)                      ║
║                                                                            ║
╠════════════════════════════════════════════════════════════════════════════╣
║                                                                            ║
║  Tab Toggle Type  Enter Confirm  Esc Cancel                               ║
║                                                                            ║
╚════════════════════════════════════════════════════════════════════════════╝
```

**Actions:**
- `Tab` - Toggle between Include/Exclude
- `Enter` - Confirm and add rule
- `Esc` - Cancel

**Data Schema:**
```typescript
interface AddRuleDialogData {
  ruleType: 'include' | 'exclude';
  pattern: string;
  isEditing: boolean; // true if editing existing rule
  ruleId?: string; // present if editing
}
```

---

### 8. Save Filter Preset Dialog

```
╔════════════════════════════════════════════════════════════════════════════╗
║                       SAVE FILTER AS PRESET                                ║
╠════════════════════════════════════════════════════════════════════════════╣
║                                                                            ║
║  Filter: Backend only                                                      ║
║                                                                            ║
║  Save Location:                                                            ║
║                                                                            ║
║  ▶ ( ) Project (.pr-builder/filters/backend_only.yaml)                    ║
║       Available only in this repository                                    ║
║                                                                            ║
║    ( ) Global (~/.pr-builder/filters/backend_only.yaml)                   ║
║       Available across all repositories                                    ║
║                                                                            ║
╠════════════════════════════════════════════════════════════════════════════╣
║                                                                            ║
║  ↑↓ Select Location  Enter Confirm  Esc Cancel                            ║
║                                                                            ║
╚════════════════════════════════════════════════════════════════════════════╝
```

**Actions:**
- `↑/↓` - Select location
- `Enter` - Confirm and save
- `Esc` - Cancel

**Data Schema:**
```typescript
interface SaveFilterPresetData {
  filterName: string;
  filter: {
    name: string;
    description: string;
    rules: Array<{
      type: 'include' | 'exclude';
      pattern: string;
      order: number;
    }>;
  };
  locations: Array<{
    id: 'project' | 'global';
    label: string;
    path: string;
  }>;
  selectedLocationIndex: number;
}
```

---

### 9. Edit Prompt Template Screen

```
╔════════════════════════════════════════════════════════════════════════════╗
║                          EDIT PROMPT TEMPLATE                              ║
╠════════════════════════════════════════════════════════════════════════════╣
║                                                                            ║
║  Current: Default prompt                                        [SESSION]  ║
║                                                                            ║
╠════════════════════════════════════════════════════════════════════════════╣
║  PROMPT TEXT                                                               ║
╠════════════════════════════════════════════════════════════════════════════╣
║                                                                            ║
║  Generate a clear PR description with: summary of changes, motivation,    ║
║  key changes, testing notes, and breaking changes if any.                 ║
║                                                                            ║
║                                                                            ║
║                                                                            ║
║                                                                            ║
║                                                                            ║
║                                                                            ║
║                                                                            ║
║                                                                            ║
║                                                                            ║
╠════════════════════════════════════════════════════════════════════════════╣
║  OPTIONS                                                                   ║
╠════════════════════════════════════════════════════════════════════════════╣
║                                                                            ║
║  [P] Load Preset    [X] Open in $EDITOR    [W] Save as Preset             ║
║  [R] Reset to Default                                                      ║
║                                                                            ║
╠════════════════════════════════════════════════════════════════════════════╣
║                                                                            ║
║  Type to edit  Ctrl+S Save & Use  Esc Cancel                               ║
║                                                                            ║
╚════════════════════════════════════════════════════════════════════════════╝
```

**Actions:**
- `Type` - Edit prompt text directly
- `P` - Load preset (go to Select Prompt Preset)
- `X` - Open in external editor (show External Editor screen)
- `W` - Save as preset (show Save Prompt Preset dialog)
- `R` - Reset to default prompt
- `Ctrl+S` - Save and use prompt
- `Esc` - Cancel

**Data Schema:**
```typescript
interface EditPromptData {
  currentName: string;
  promptText: string;
  isSession: boolean; // true if not saved as preset
  presetId?: string; // present if loaded from preset
}
```

---

### 10. Select Prompt Preset Screen

```
╔════════════════════════════════════════════════════════════════════════════╗
║                        SELECT PROMPT PRESET                                ║
╠════════════════════════════════════════════════════════════════════════════╣
║                                                                            ║
║  BUILT-IN PRESETS                                                          ║
╠════════════════════════════════════════════════════════════════════════════╣
║                                                                            ║
║ ▶ Default                                                         [BUILTIN]║
║   Generate a clear PR description with: summary of changes...             ║
║                                                                            ║
║   Detailed                                                        [BUILTIN]║
║   Create a comprehensive PR description including: Executive summary,     ║
║   detailed changes by component, rationale, testing strategy...           ║
║                                                                            ║
║   Concise                                                         [BUILTIN]║
║   Write a brief PR description: What changed, why, and how to test.       ║
║                                                                            ║
║   Conventional Commits                                            [BUILTIN]║
║   Generate PR description following conventional commits format with      ║
║   type, scope, breaking changes, and footer.                              ║
║                                                                            ║
╠════════════════════════════════════════════════════════════════════════════╣
║  PROJECT PRESETS                                                           ║
╠════════════════════════════════════════════════════════════════════════════╣
║                                                                            ║
║   Security Review                                                [PROJECT] ║
║   Focus on security implications, auth changes, data access...            ║
║   (.pr-builder/prompts/security_review.yaml)                              ║
║                                                                            ║
╠════════════════════════════════════════════════════════════════════════════╣
║  GLOBAL PRESETS                                                            ║
╠════════════════════════════════════════════════════════════════════════════╣
║                                                                            ║
║   API Changes                                                     [GLOBAL] ║
║   Emphasize API contract changes, versioning, backwards compatibility     ║
║   (~/.pr-builder/prompts/api_changes.yaml)                                ║
║                                                                            ║
║   Refactoring                                                     [GLOBAL] ║
║   Highlight code quality improvements, technical debt addressed...        ║
║   (~/.pr-builder/prompts/refactoring.yaml)                                ║
║                                                                            ║
╠════════════════════════════════════════════════════════════════════════════╣
║                                                                            ║
║  ↑↓ Navigate  Enter Select  V View Full  E Edit  D Delete  Esc Cancel     ║
║                                                                            ║
╚════════════════════════════════════════════════════════════════════════════╝
```

**Actions:**
- `↑/↓` - Navigate presets
- `Enter` - Select preset and load
- `V` - View full prompt text
- `E` - Edit preset (go to Edit Prompt Template with preset data)
- `D` - Delete preset (not available for built-in)
- `Esc` - Cancel

**Data Schema:**
```typescript
interface SelectPromptPresetData {
  builtinPresets: Array<{
    id: string;
    name: string;
    preview: string;
    fullText: string;
  }>;
  projectPresets: Array<{
    id: string;
    name: string;
    description: string;
    filePath: string;
    fullText: string;
  }>;
  globalPresets: Array<{
    id: string;
    name: string;
    description: string;
    filePath: string;
    fullText: string;
  }>;
  selectedIndex: number;
  selectedCategory: 'builtin' | 'project' | 'global';
}
```

---

### 11. Save Prompt Preset Dialog

```
╔════════════════════════════════════════════════════════════════════════════╗
║                      SAVE PROMPT AS PRESET                                 ║
╠════════════════════════════════════════════════════════════════════════════╣
║                                                                            ║
║  Preset Name:                                                              ║
║  > My Detailed Format_                                                     ║
║                                                                            ║
║  Description:                                                              ║
║  > Comprehensive format with executive summary and component breakdown_   ║
║                                                                            ║
╠════════════════════════════════════════════════════════════════════════════╣
║  Save Location:                                                            ║
║                                                                            ║
║  ▶ ( ) Project (.pr-builder/prompts/my_detailed_format.yaml)              ║
║       Available only in this repository                                    ║
║                                                                            ║
║    ( ) Global (~/.pr-builder/prompts/my_detailed_format.yaml)             ║
║       Available across all repositories                                    ║
║                                                                            ║
╠════════════════════════════════════════════════════════════════════════════╣
║                                                                            ║
║  Tab Next Field  ↑↓ Select Location  Enter Confirm  Esc Cancel            ║
║                                                                            ║
╚════════════════════════════════════════════════════════════════════════════╝
```

**Actions:**
- `Tab` - Move to next field
- `↑/↓` - Select location (when in location field)
- `Enter` - Confirm and save
- `Esc` - Cancel

**Data Schema:**
```typescript
interface SavePromptPresetData {
  promptText: string;
  name: string;
  description: string;
  locations: Array<{
    id: 'project' | 'global';
    label: string;
    path: string;
  }>;
  selectedLocationIndex: number;
}
```

---

### 12. External Editor Screen

```
╔════════════════════════════════════════════════════════════════════════════╗
║                          EDIT PROMPT TEMPLATE                              ║
╠════════════════════════════════════════════════════════════════════════════╣
║                                                                            ║
║  Opening in $EDITOR (vim)...                                               ║
║                                                                            ║
║  Temporary file: /tmp/pr-builder-prompt-a3f9d2.txt                         ║
║                                                                            ║
║  Save and close the editor to continue.                                    ║
║  Changes will be loaded automatically.                                     ║
║                                                                            ║
╠════════════════════════════════════════════════════════════════════════════╣
║                                                                            ║
║  Waiting for editor to close...                                            ║
║                                                                            ║
╚════════════════════════════════════════════════════════════════════════════╝
```

**Actions:**
- Wait for external editor process to complete
- On completion, load edited text and return to Edit Prompt Template

**Data Schema:**
```typescript
interface ExternalEditorData {
  editor: string; // from $EDITOR env variable
  tempFilePath: string;
  originalText: string;
}
```

---

## Action Reference

### Global Actions
- `Esc` - Return to previous screen / Cancel current operation
- `Q` - Quit application (from Main Screen only)

### Main Screen Actions
| Key | Action | Destination |
|-----|--------|-------------|
| `C` | Edit Context | Edit Context Window |
| `F` | Edit Filters | Edit File Filters |
| `H` | View Hidden Files | Filtered Out Files View |
| `E` | Edit Prompt | Edit Prompt Template |
| `A` | Add Files | Add Files from Repo |
| `T` | Add Notes | Add Text Notes |
| `G` | Generate Description | Generate Screen |
| `S` | Save Session | Save Session Dialog |
| `L` | Load Session | Load Session Dialog |

### Edit Context Window Actions
| Key | Action | Description |
|-----|--------|-------------|
| `↑/↓` | Navigate | Move through file list |
| `Space` | Toggle | Include/exclude file |
| `Enter` | View Full | Show complete diff/file |
| `R` | Replace Options | Open Replace with Full File dialog |
| `D` | Restore Diff | Convert full file back to diff |
| `F` | Filter | Go to Edit File Filters |
| `A` | Add Files | Go to Add Files from Repo |

### Filter Management Actions
| Key | Action | Description |
|-----|--------|-------------|
| `↑/↓` | Navigate | Move through filters/rules |
| `Space` | Toggle | Enable/disable filter |
| `N` | New | Create new custom filter |
| `E` | Edit | Edit selected filter |
| `D` | Delete | Delete selected filter |
| `W` | Save Preset | Save filter as preset |
| `+` | Add Rule | Add new rule to filter |
| `-` | Delete Rule | Remove selected rule |
| `Shift+↑/↓` | Reorder | Change rule execution order |
| `S` | Save & Use | Save filter and apply |
| `Enter` | Confirm | Apply/select/confirm action |

### Prompt Template Actions
| Key | Action | Description |
|-----|--------|-------------|
| `Type` | Edit | Direct text editing |
| `P` | Load Preset | Open preset selection |
| `X` | External Editor | Open in $EDITOR |
| `W` | Save Preset | Save as preset |
| `R` | Reset | Reset to default |
| `Ctrl+S` | Save & Use | Save and return |
| `V` | View Full | View complete prompt text |

---

## Data Schemas

### Core Data Types

```typescript
// File representation
interface FileInfo {
  path: string;
  included: boolean;
  additions: number;
  deletions: number;
  tokens: number;
  type: 'diff' | 'full_file';
  version?: 'before' | 'after' | 'both';
  content?: string; // full content if loaded
  diffPreview?: string; // first N lines for preview
}

// Branch information
interface BranchInfo {
  source: string;
  target: string;
}

// Filter rule
interface FilterRule {
  id: string;
  order: number;
  type: 'include' | 'exclude';
  pattern: string;
}

// Filter definition
interface Filter {
  id: string;
  name: string;
  description: string;
  rules: FilterRule[];
  location?: 'project' | 'global';
  filePath?: string;
  temporary?: boolean;
}

// Prompt preset
interface PromptPreset {
  id: string;
  name: string;
  description?: string;
  text: string;
  location: 'builtin' | 'project' | 'global';
  filePath?: string;
}

// Additional context item
interface ContextItem {
  type: 'file' | 'note';
  path?: string; // for files
  content: string;
  tokens: number;
}

// Session state (for save/load)
interface SessionState {
  branch: BranchInfo;
  files: FileInfo[];
  activeFilters: string[];
  additionalContext: ContextItem[];
  promptTemplate: string;
  promptPresetId?: string;
  timestamp: string;
}
```

### YAML File Formats

**Filter Preset (.pr-builder/filters/*.yaml)**
```yaml
name: Backend only
description: Only backend Go code and proto files, excluding tests
rules:
  - order: 1
    type: include
    pattern: src/backend/**/*
  - order: 2
    type: exclude
    pattern: "**/*.test.*"
  - order: 3
    type: include
    pattern: src/api/**/*.go
  - order: 4
    type: include
    pattern: "**/*.proto"
```

**Prompt Preset (.pr-builder/prompts/*.yaml)**
```yaml
name: Security Review
description: Focus on security implications, auth changes, data access
text: |
  Generate a PR description focusing on security aspects:
  
  ## Security Impact
  - Authentication/Authorization changes
  - Data access modifications
  - New dependencies and their security status
  
  ## Changes
  [Detailed changes here]
  
  ## Security Testing
  - What security tests were added/modified
  - Manual security review checklist
```

**Session State (.pr-builder/sessions/*.yaml)**
```yaml
timestamp: "2024-01-15T10:30:00Z"
branch:
  source: feature/user-auth
  target: main
files:
  - path: src/auth/login.ts
    included: true
    type: diff
    additions: 89
    deletions: 12
    tokens: 342
  - path: tests/auth.test.ts
    included: true
    type: full_file
    version: after
    tokens: 701
active_filters:
  - exclude_tests
additional_context:
  - type: note
    content: "This PR implements OAuth 2.0 authentication"
    tokens: 15
prompt_template: "Generate a clear PR description..."
prompt_preset_id: detailed
```

---

## Implementation Notes

### Token Counting
- Token count should be calculated using the target LLM's tokenizer
- Display as human-readable format (e.g., "342t", "1.8k", "12.5k")
- Update dynamically when files are toggled or filters applied

### File Paths
- Project-specific: `.pr-builder/` in repository root
- Global: `~/.pr-builder/` in user home directory
- Subdirectories: `filters/`, `prompts/`, `sessions/`

### Navigation State
- Each screen maintains its own cursor/selection state
- Selection should be visually distinct (▶ marker)
- Multi-select uses checkboxes `[✓]` / `[ ]`

### Text Input
- Support standard editing keys (Backspace, Delete, Arrow keys)
- For multi-line text, support Up/Down arrow navigation
- Tab key moves between form fields

### External Editor
- Use `$EDITOR` environment variable (fallback to `vi`)
- Write content to temporary file
- Block until editor process completes
- Load modified content on exit

### Preview Updates
- Filter previews should update in real-time as rules change
- Token counts should update when files are toggled
- File match indicators should reflect current filter state