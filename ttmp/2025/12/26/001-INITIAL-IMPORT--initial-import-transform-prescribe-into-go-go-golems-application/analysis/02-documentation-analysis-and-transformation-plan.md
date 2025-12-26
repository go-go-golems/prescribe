---
Title: Documentation Analysis and Transformation Plan
Ticket: 001-INITIAL-IMPORT
Status: active
Topics:
    - migration
    - refactoring
    - ci-cd
    - go-module
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-26T17:30:00.000000000-05:00
WhatFor: ""
WhenToUse: ""
---

# Documentation Analysis and Transformation Plan

## Executive Summary

This document analyzes all markdown files in the prescribe repository and provides a comprehensive plan for:
1. **Archiving** historical/development diaries and outdated content
2. **Transforming** user-facing documentation to match go-go-golems documentation standards
3. **Identifying** missing documentation that should be created

The analysis is based on the documentation guidelines from `glazed/pkg/doc/topics/how-to-write-good-documentation-pages.md` and the help system structure used in glazed/pinocchio.

## Current Documentation Inventory

### Files Analyzed

| File | Lines | Type | Status | Recommendation |
|------|-------|------|--------|----------------|
| `README.md` | 463 | User-facing | Needs update | Transform to proper README + help docs |
| `AGENT.md` | 70 | Developer | Needs update | Update references, keep as-is |
| `dev-diary.md` | 374 | Historical | Archive | Move to archive |
| `FILTER-DEVELOPMENT-DIARY.md` | 492 | Historical | Archive | Move to archive |
| `FILTER-SYSTEM.md` | 658 | User-facing | Transform | Convert to help system docs |
| `FILTER-SYSTEM-SUMMARY.md` | 413 | Summary | Archive/Transform | Extract useful content, archive |
| `FILTER-TUI-SCREENSHOTS.md` | 579 | Visual guide | Transform | Convert to help system with images |
| `TUI-DEMO.md` | 143 | User-facing | Transform | Convert to help system tutorial |
| `TUI-SCREENSHOTS.md` | 832 | Visual guide | Transform | Convert to help system with images |
| `PLAYBOOK-Bubbletea-TUI-Development.md` | 980 | Reference | Transform | Convert to help system reference |
| `PROJECT-SUMMARY.md` | 354 | Summary | Archive | Move to archive |

**Total:** 11 files, ~5,358 lines

## Detailed Analysis

### 1. Files to Archive (Historical/Development Content)

These files document the development process and are valuable for historical reference but not for end users.

#### 1.1. `dev-diary.md` (374 lines)
**Content:** Development diary for TUI implementation  
**Status:** Historical development log  
**Action:** Archive to `ttmp/archive/` or `doc/archive/`  
**Reason:** Documents development process, not user-facing documentation

#### 1.2. `FILTER-DEVELOPMENT-DIARY.md` (492 lines)
**Content:** Step-by-step development diary for filter system  
**Status:** Historical development log  
**Action:** Archive to `ttmp/archive/` or `doc/archive/`  
**Reason:** Documents implementation process, useful for future developers but not end users

#### 1.3. `PROJECT-SUMMARY.md` (354 lines)
**Content:** High-level project summary, architecture overview, testing results  
**Status:** Project summary/documentation  
**Action:** Archive or extract useful architecture diagrams to new docs  
**Reason:** Contains useful architecture info but is more of a project report than user docs

**Recommendation:** Create `doc/archive/` directory and move these files there. They're valuable for historical reference but clutter the root directory.

### 2. Files to Transform (User-Facing Documentation)

These files contain valuable user-facing content but need to be restructured according to go-go-golems documentation standards.

#### 2.1. `README.md` (463 lines)
**Current State:**
- Comprehensive user guide
- Installation instructions
- CLI command reference
- Examples and use cases
- References `pr-builder` (needs update to `prescribe`)

**Issues:**
- Too long for a README (should be concise overview)
- Mixes installation, usage, and reference
- No YAML frontmatter
- Doesn't follow help system structure

**Transformation Plan:**
1. **Keep README.md** as a concise overview (50-100 lines):
   - What prescribe is
   - Quick start (3-5 commands)
   - Link to help system
   - Installation
   - Basic usage example

2. **Create help system docs:**
   - `doc/topics/getting-started.md` - Installation and first steps
   - `doc/topics/commands-reference.md` - Complete command reference
   - `doc/examples/basic-workflow.md` - Basic usage example
   - `doc/examples/team-workflow.md` - Team collaboration example
   - `doc/examples/ci-integration.md` - CI/CD integration example

**Priority:** HIGH - This is the main entry point

#### 2.2. `FILTER-SYSTEM.md` (658 lines)
**Current State:**
- Comprehensive filter system documentation
- Architecture explanation
- Pattern matching details
- CLI command examples
- Use cases and workflows
- References `pr-builder` (needs update)

**Issues:**
- No YAML frontmatter
- Mixes architecture with usage
- Too detailed for a single document
- Should be split into topics

**Transformation Plan:**
1. **Create `doc/topics/filter-system.md`** (GeneralTopic):
   - Overview of filter system
   - Core concepts (filters, rules, patterns)
   - Architecture overview
   - Pattern matching reference

2. **Create `doc/examples/filter-basics.md`** (Example):
   - Basic filter usage
   - Common patterns
   - Quick examples

3. **Create `doc/tutorials/filter-workflow.md`** (Tutorial):
   - Step-by-step filter workflow
   - Team template creation
   - Advanced patterns

**Priority:** HIGH - Core feature documentation

#### 2.3. `FILTER-SYSTEM-SUMMARY.md` (413 lines)
**Current State:**
- Implementation summary
- Feature checklist
- Architecture diagrams
- Code examples

**Issues:**
- Mixes implementation details with user docs
- Some content overlaps with FILTER-SYSTEM.md
- More of a project report

**Transformation Plan:**
- Extract useful architecture diagrams → `doc/topics/filter-system.md`
- Extract feature checklist → Update FILTER-SYSTEM.md
- Archive the rest

**Priority:** MEDIUM - Extract useful content, then archive

#### 2.4. `FILTER-TUI-SCREENSHOTS.md` (579 lines)
**Current State:**
- Visual guide to filter system in TUI
- ASCII art screenshots
- Screen-by-screen walkthrough
- Keyboard shortcuts

**Issues:**
- No YAML frontmatter
- Should be integrated with filter system docs
- ASCII art is good but could use real screenshots

**Transformation Plan:**
- Integrate into `doc/topics/filter-system.md` as a section
- Create `doc/tutorials/tui-filters.md` (Tutorial) with visual guide
- Keep ASCII art, add real screenshots if available

**Priority:** MEDIUM - Visual guides are valuable

#### 2.5. `TUI-DEMO.md` (143 lines)
**Current State:**
- Brief TUI overview
- Main screen description
- Basic keyboard shortcuts
- Generating workflow

**Issues:**
- Too brief
- No YAML frontmatter
- Should be expanded and structured

**Transformation Plan:**
- Create `doc/topics/tui-overview.md` (GeneralTopic):
   - TUI architecture
   - Screen overview
   - Navigation patterns

- Create `doc/tutorials/tui-basics.md` (Tutorial):
   - Step-by-step TUI usage
   - Keyboard shortcuts reference
   - Common workflows

**Priority:** HIGH - TUI is a core feature

#### 2.6. `TUI-SCREENSHOTS.md` (832 lines)
**Current State:**
- Comprehensive visual guide
- All screens documented
- Detailed explanations
- Architecture overview
- MVC pattern explanation

**Issues:**
- Very long (832 lines)
- Mixes user docs with architecture
- No YAML frontmatter
- Should be split into multiple docs

**Transformation Plan:**
1. **Create `doc/topics/tui-architecture.md`** (GeneralTopic):
   - MVC pattern in TUI
   - Screen structure
   - State management

2. **Create `doc/tutorials/tui-complete-guide.md`** (Tutorial):
   - Complete visual walkthrough
   - All screens explained
   - Keyboard shortcuts reference

3. **Create `doc/examples/tui-workflows.md`** (Example):
   - Common TUI workflows
   - Tips and tricks

**Priority:** HIGH - Most comprehensive TUI documentation

#### 2.7. `PLAYBOOK-Bubbletea-TUI-Development.md` (980 lines)
**Current State:**
- Comprehensive playbook for building CLI/TUI apps
- Architecture principles
- Implementation steps
- Code examples
- Testing strategies
- Best practices

**Issues:**
- Very long (980 lines)
- More of a general development guide than prescribe-specific
- Could be valuable as a reference
- No YAML frontmatter

**Transformation Plan:**
**Option A: Keep as reference (recommended)**
- Create `doc/topics/development-playbook.md` (GeneralTopic)
- Add YAML frontmatter
- Update references from `pr-builder` to `prescribe`
- Mark as developer reference

**Option B: Extract prescribe-specific content**
- Extract prescribe architecture → `doc/topics/architecture.md`
- Extract testing → `doc/topics/testing.md`
- Archive general playbook content

**Recommendation:** Option A - This is valuable reference material for developers working on prescribe or similar tools.

**Priority:** MEDIUM - Valuable but not user-facing

### 3. Files to Update (Minor Changes)

#### 3.1. `AGENT.md` (70 lines)
**Current State:**
- Developer guidelines for go-go-golems projects
- Build commands
- Project structure
- Contains reference to `XXX` placeholder (line 5)

**Issues:**
- References `XXX/YYY/FOOO` pattern (should be updated)
- Otherwise good developer reference

**Action:**
- Update line 5: Remove `XXX` reference or update to `prescribe`
- Keep as-is otherwise

**Priority:** LOW - Quick fix

## Missing Documentation

Based on the analysis and go-go-golems documentation standards, the following documentation should be created:

### 3.1. Core User Documentation

1. **`doc/topics/getting-started.md`** (GeneralTopic)
   - Installation
   - First session
   - Basic workflow
   - Common use cases

2. **`doc/topics/commands-reference.md`** (GeneralTopic)
   - Complete command reference
   - All flags and options
   - Command combinations
   - Examples for each command

3. **`doc/topics/session-management.md`** (GeneralTopic)
   - Session file format
   - Save/load workflows
   - Team templates
   - Session sharing

4. **`doc/topics/context-management.md`** (GeneralTopic)
   - Adding context files
   - Adding notes
   - Token management
   - Context best practices

5. **`doc/topics/prompt-customization.md`** (GeneralTopic)
   - Prompt presets
   - Custom prompts
   - Prompt templates
   - Best practices

### 3.2. Examples

1. **`doc/examples/basic-workflow.md`** (Example)
   - Simple PR description generation
   - Step-by-step CLI workflow

2. **`doc/examples/filter-workflow.md`** (Example)
   - Using filters to exclude tests
   - Creating team templates

3. **`doc/examples/team-collaboration.md`** (Example)
   - Sharing session templates
   - Team workflows

4. **`doc/examples/ci-integration.md`** (Example)
   - Integrating with CI/CD
   - Automated PR descriptions

### 3.3. Tutorials

1. **`doc/tutorials/complete-workflow.md`** (Tutorial)
   - End-to-end workflow
   - From init to generation
   - Best practices

2. **`doc/tutorials/advanced-filters.md`** (Tutorial)
   - Complex filter patterns
   - Filter combinations
   - Performance considerations

3. **`doc/tutorials/tui-mastery.md`** (Tutorial)
   - Advanced TUI usage
   - Keyboard shortcuts
   - Power user tips

### 3.4. Developer Documentation

1. **`doc/topics/architecture.md`** (GeneralTopic)
   - System architecture
   - Layer structure
   - Design patterns
   - Extension points

2. **`doc/topics/contributing.md`** (GeneralTopic)
   - Development setup
   - Code structure
   - Testing
   - Contribution guidelines

## Transformation Strategy

### Phase 1: Cleanup (Immediate)
1. Create `doc/archive/` directory
2. Move historical diaries:
   - `dev-diary.md` → `doc/archive/dev-diary.md`
   - `FILTER-DEVELOPMENT-DIARY.md` → `doc/archive/filter-development-diary.md`
   - `PROJECT-SUMMARY.md` → `doc/archive/project-summary.md` (or extract useful content first)

### Phase 2: Update Existing Files (High Priority)
1. **README.md:**
   - Condense to 50-100 lines
   - Update `pr-builder` → `prescribe`
   - Add link to help system
   - Keep quick start

2. **AGENT.md:**
   - Fix `XXX` reference
   - Update if needed

### Phase 3: Transform to Help System (High Priority)
1. **Create help system structure:**
   ```
   doc/
   ├── topics/
   │   ├── getting-started.md
   │   ├── commands-reference.md
   │   ├── filter-system.md
   │   ├── tui-overview.md
   │   ├── session-management.md
   │   └── ...
   ├── examples/
   │   ├── basic-workflow.md
   │   ├── filter-workflow.md
   │   └── ...
   ├── tutorials/
   │   ├── complete-workflow.md
   │   ├── tui-basics.md
   │   └── ...
   └── archive/
       └── ...
   ```

2. **Transform FILTER-SYSTEM.md:**
   - Split into topics/examples/tutorials
   - Add YAML frontmatter
   - Update references

3. **Transform TUI docs:**
   - Integrate TUI-DEMO.md and TUI-SCREENSHOTS.md
   - Create structured topics and tutorials
   - Add YAML frontmatter

### Phase 4: Create Missing Documentation (Medium Priority)
1. Create core user documentation topics
2. Create examples
3. Create tutorials
4. Create developer documentation

### Phase 5: Integrate Help System (Future)
1. Implement help system integration (like glazed)
2. Add `prescribe help` command
3. Load docs from `doc/` directory
4. Enable querying and filtering

## Documentation Standards Compliance

### YAML Frontmatter Template

All transformed documentation should follow this structure:

```yaml
---
Title: Clear, Descriptive Title
Slug: url-friendly-identifier
Short: One-line description
Topics:
  - relevant
  - topic
  - tags
Commands:
  - RelatedCommands (if applicable)
Flags:
  - relevant-flags (if applicable)
IsTemplate: false
IsTopLevel: true/false
ShowPerDefault: true/false
SectionType: GeneralTopic|Example|Application|Tutorial
---
```

### Writing Style Checklist

- [ ] Topic-focused introductory paragraph (explains "what" and "why")
- [ ] Clear section hierarchy (H2 for major sections, H3 for subsections)
- [ ] Minimal, focused code examples
- [ ] Comments explain "why" not "what"
- [ ] Runnable code examples where possible
- [ ] Expected output shown
- [ ] Links use `prescribe help <slug>` format (when help system is integrated)
- [ ] Active voice, present tense
- [ ] Consistent terminology
- [ ] Audience-centric (developer-user perspective)

## Priority Matrix

| Priority | Task | Effort | Impact |
|----------|------|--------|--------|
| P0 | Update README.md (condense, fix references) | Low | High |
| P0 | Fix AGENT.md (XXX reference) | Low | Low |
| P0 | Archive historical diaries | Low | Medium |
| P1 | Transform FILTER-SYSTEM.md | Medium | High |
| P1 | Transform TUI docs | Medium | High |
| P1 | Create getting-started.md | Medium | High |
| P1 | Create commands-reference.md | Medium | High |
| P2 | Create examples | Medium | Medium |
| P2 | Create tutorials | High | Medium |
| P2 | Transform PLAYBOOK | Low | Low |
| P3 | Create missing topics | High | Low |
| P3 | Integrate help system | High | High (future) |

## File-by-File Recommendations

### Immediate Actions (Do First)

1. **README.md** → Condense, update references, add help system link
2. **AGENT.md** → Fix XXX reference
3. **dev-diary.md** → Move to `doc/archive/`
4. **FILTER-DEVELOPMENT-DIARY.md** → Move to `doc/archive/`

### High Priority Transformations

1. **FILTER-SYSTEM.md** → Split into:
   - `doc/topics/filter-system.md`
   - `doc/examples/filter-basics.md`
   - `doc/tutorials/filter-workflow.md`

2. **TUI-SCREENSHOTS.md + TUI-DEMO.md** → Combine and split into:
   - `doc/topics/tui-overview.md`
   - `doc/tutorials/tui-complete-guide.md`
   - `doc/examples/tui-workflows.md`

### Medium Priority

1. **FILTER-SYSTEM-SUMMARY.md** → Extract useful content, archive
2. **FILTER-TUI-SCREENSHOTS.md** → Integrate into filter system docs
3. **PLAYBOOK-Bubbletea-TUI-Development.md** → Add frontmatter, mark as developer reference

### Low Priority / Future

1. Create missing documentation topics
2. Integrate help system
3. Add real screenshots (replace ASCII art)

## Implementation Notes

### Help System Integration

Currently, prescribe doesn't have a help system integrated. The transformation should:
1. Create documentation in the expected structure
2. Add YAML frontmatter (even if help system isn't integrated yet)
3. Plan for future help system integration (similar to glazed)

### Reference Updates

All documentation needs to be updated:
- `pr-builder` → `prescribe`
- `pr-builder init` → `prescribe init`
- `.pr-builder/` → `.prescribe/` (if session directory is renamed)
- Update all command examples
- Update all file paths

### Code Examples

All code examples should be:
- Updated to use `prescribe` command
- Tested and verified
- Include expected output
- Minimal and focused

## Success Criteria

Documentation transformation is complete when:

- [ ] All historical diaries archived
- [ ] README.md is concise and links to help system
- [ ] All user-facing docs have YAML frontmatter
- [ ] Core topics documented (getting-started, commands, filters, TUI)
- [ ] Examples created for common workflows
- [ ] Tutorials created for complex workflows
- [ ] All references updated (`pr-builder` → `prescribe`)
- [ ] Documentation follows style guide
- [ ] Help system integration planned (if not implemented)

## Related Files

This analysis covers:
- All markdown files in prescribe root directory
- Documentation structure in glazed (reference)
- Documentation guidelines from glazed

## Next Steps

1. Review and approve this analysis
2. Create `doc/` directory structure
3. Begin Phase 1 (cleanup)
4. Begin Phase 2 (update existing files)
5. Begin Phase 3 (transform to help system)
6. Create missing documentation (Phase 4)
7. Plan help system integration (Phase 5)
