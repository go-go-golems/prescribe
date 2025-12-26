---
Title: Current Structure Analysis and Transformation Plan
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
LastUpdated: 2025-12-26T16:54:03.740878109-05:00
WhatFor: ""
WhenToUse: ""
---

# Current Structure Analysis and Transformation Plan

## Executive Summary

This document analyzes the current structure of `prescribe` (formerly `pr-builder`) and provides a comprehensive plan to transform it into a proper go-go-golems application following the patterns established in `pinocchio/`. The transformation involves module renaming, CI/CD setup, Makefile standardization, and structural alignment with go-go-golems conventions.

## Current State Analysis

### Module and Package Structure

**Current Module Name:** `github.com/user/pr-builder`  
**Target Module Name:** `github.com/go-go-golems/prescribe`

**Current Issues:**
- Module path uses placeholder `github.com/user/pr-builder` (31 occurrences across codebase)
- Binary name is `pr-builder` but should be `prescribe`
- Root command uses `pr-builder` instead of `prescribe`
- Makefile references `XXX` placeholder (3 occurrences)
- `cmd/XXX/main.go` exists but is empty/unused

**Current Package Structure:**
```
prescribe/
├── cmd/                    # CLI commands (Cobra)
│   ├── root.go            # Root command (uses "pr-builder")
│   ├── init.go
│   ├── save.go
│   ├── load.go
│   ├── show.go
│   ├── generate.go
│   ├── add_filter.go
│   ├── toggle_file.go
│   ├── add_context.go
│   ├── tui.go
│   └── XXX/               # Empty/unused directory
│       └── main.go        # Empty file
├── internal/
│   ├── domain/            # Business logic
│   ├── git/               # Git operations
│   ├── api/               # API client (mock)
│   ├── controller/        # Orchestration
│   ├── session/           # YAML persistence
│   └── tui/               # Bubbletea UI
├── main.go                # Entry point (imports cmd)
├── go.mod                 # Module: github.com/user/pr-builder
├── Makefile               # References XXX placeholder
├── lefthook.yml           # Git hooks (simpler than pinocchio)
├── .golangci.yml          # Linter config (matches pinocchio)
└── LICENSE                # MIT License (matches pinocchio)
```

### Comparison with Pinocchio

**Pinocchio Structure:**
```
pinocchio/
├── cmd/
│   └── pinocchio/         # Main command directory
│       ├── main.go        # Entry point
│       ├── cmds/          # Subcommands
│       └── prompts/       # Embedded prompts
├── pkg/                   # Library code
│   ├── cmds/
│   ├── middlewares/
│   └── ...
├── go.mod                 # Module: github.com/go-go-golems/pinocchio
├── Makefile               # References "pinocchio" (not XXX)
└── lefthook.yml           # More comprehensive hooks
```

**Key Differences:**
1. **Command Structure:** Pinocchio uses `cmd/pinocchio/main.go` as entry point, prescribe uses `main.go` + `cmd/` package
2. **Package Organization:** Pinocchio has `pkg/` for library code, prescribe uses `internal/`
3. **Makefile:** Pinocchio has no `XXX` placeholders, uses actual binary name
4. **Lefthook:** Pinocchio has more comprehensive pre-push hooks (includes gosec, govulncheck)
5. **Module Path:** Pinocchio uses proper go-go-golems path, prescribe uses placeholder

### Dependencies Analysis

**Current Dependencies (prescribe):**
- `github.com/charmbracelet/bubbletea` v1.3.10
- `github.com/charmbracelet/lipgloss` v1.1.0
- `github.com/spf13/cobra` v1.10.2
- `gopkg.in/yaml.v3` v3.0.1

**Pinocchio Dependencies (for reference):**
- Uses go-go-golems ecosystem packages (`glazed`, `geppetto`, `clay`, `bobatea`, `prompto`)
- More comprehensive dependency set
- Uses `github.com/pkg/errors` for error wrapping
- Uses `github.com/rs/zerolog` for logging

**Note:** Prescribe is currently standalone and doesn't use go-go-golems packages. This is acceptable for now, but we should consider integration opportunities.

### CI/CD and Build Infrastructure

**Current State:**
- ✅ `.golangci.yml` exists and matches pinocchio configuration
- ✅ `lefthook.yml` exists but simpler than pinocchio
- ✅ `Makefile` exists but has `XXX` placeholders
- ❌ No `.github/workflows/` directory (no CI/CD)
- ❌ Makefile doesn't include geppetto-lint (pinocchio has this)

**Pinocchio Makefile Features:**
- `geppetto-lint-build` and `geppetto-lint` targets
- `codeql-local` target for security scanning
- More comprehensive `bump-glazed` target (includes multiple packages)
- Uses actual binary name (`pinocchio`) not placeholder

**Pinocchio Lefthook:**
- Pre-commit: `lint` (with `lintmax`) and `test`
- Pre-push: `release` (goreleaser), `lint` (with gosec, govulncheck), `test`

**Current Lefthook:**
- Pre-commit: `lint` (with `lint`) and `test`
- Pre-push: `release`, `lint`, `test` (simpler)

### Documentation Structure

**Current State:**
- ✅ `README.md` exists (comprehensive)
- ✅ `AGENT.md` exists (has `XXX` placeholder reference)
- ✅ Multiple markdown docs (FILTER-SYSTEM.md, TUI-SCREENSHOTS.md, etc.)
- ✅ `ttmp/` directory exists (for temporary docs)

**Pinocchio:**
- ✅ `AGENT.md` exists (no placeholders)
- ✅ `README.md` exists
- ✅ `ttmp/` directory exists
- ✅ `changelog.md` exists

## Transformation Requirements

### 1. Module Renaming

**Priority: CRITICAL**

**Changes Required:**
1. Update `go.mod`: `github.com/user/pr-builder` → `github.com/go-go-golems/prescribe`
2. Update all import statements (31 occurrences):
   - `github.com/user/pr-builder/cmd` → `github.com/go-go-golems/prescribe/cmd`
   - `github.com/user/pr-builder/internal/*` → `github.com/go-go-golems/prescribe/internal/*`
3. Update root command name: `pr-builder` → `prescribe`
4. Update binary name references in Makefile: `XXX` → `prescribe`
5. Update binary name references in documentation

**Files to Update:**
- `go.mod`
- `main.go`
- All files in `cmd/` (12 files)
- All files in `internal/` (7+ files)
- `Makefile`
- `AGENT.md`
- `README.md` (if it references binary name)
- Test scripts in `test/` directory

### 2. Command Structure Alignment

**Priority: MEDIUM**

**Current:** `main.go` imports `cmd` package, `cmd/root.go` defines root command  
**Target:** Consider moving to `cmd/prescribe/main.go` pattern like pinocchio

**Decision:** Keep current structure for now (simpler), but document pinocchio pattern for future consideration.

**Changes Required:**
- Remove `cmd/XXX/main.go` (empty/unused)
- Update root command name in `cmd/root.go`
- Ensure command structure is consistent

### 3. Makefile Standardization

**Priority: HIGH**

**Changes Required:**
1. Replace `XXX` with `prescribe`:
   - Line 49: `GOPROXY=proxy.golang.org go list -m github.com/go-go-golems/XXX@$(shell svu current)`
   - Line 56: `XXX_BINARY=$(shell which XXX)`
   - Line 58: `go build -o ./dist/XXX ./cmd/XXX`
2. Consider adding geppetto-lint targets (like pinocchio)
3. Consider adding codeql-local target (optional)
4. Update `bump-glazed` if prescribe starts using glazed packages

**Current Makefile Issues:**
```makefile
# Line 49 - WRONG
GOPROXY=proxy.golang.org go list -m github.com/go-go-golems/XXX@$(shell svu current)

# Line 56-59 - WRONG
XXX_BINARY=$(shell which XXX)
install:
	go build -o ./dist/XXX ./cmd/XXX && \
		cp ./dist/XXX $(XXX_BINARY)
```

**Target:**
```makefile
# Line 49 - CORRECT
GOPROXY=proxy.golang.org go list -m github.com/go-go-golems/prescribe@$(shell svu current)

# Line 56-59 - CORRECT
prescribe_BINARY=$(shell which prescribe)
install:
	go build -o ./dist/prescribe . && \
		cp ./dist/prescribe $(prescribe_BINARY)
```

**Note:** Prescribe builds from root (`.`), not `./cmd` or `./cmd/prescribe`. The `main.go` is at the root and imports the `cmd` package. The current Makefile incorrectly references `./cmd/XXX` which doesn't exist. The correct build path should be `.` (root directory).

**Additional Findings:**
- `pr-builder/` directory exists but appears unused (empty subdirectories)
- `pkg/doc.go` exists but is empty (just empty init function)
- `cmd/XXX/main.go` is empty and unused
- Test scripts contain hardcoded paths: `/home/ubuntu/pr-builder/pr-builder` and `/tmp/pr-builder-test-repo`
- Test scripts reference `.pr-builder/` directory (session storage)

### 4. Lefthook Enhancement

**Priority: MEDIUM**

**Current:** Basic hooks  
**Target:** Match pinocchio's comprehensive hooks

**Changes Required:**
- Pre-commit: Use `lintmax` instead of `lint` (like pinocchio)
- Pre-push: Add `gosec` and `govulncheck` to lint step
- Ensure parallel execution is configured

### 5. CI/CD Setup

**Priority: HIGH**

**Current:** No CI/CD  
**Target:** Add GitHub Actions workflows

**Required Workflows:**
1. **CI Workflow** (`.github/workflows/ci.yml`):
   - Run on push to main/master and PRs
   - Run tests
   - Run linting
   - Build binary
   - (Optional) Run security scans

2. **Release Workflow** (`.github/workflows/release.yml`):
   - Run on tags
   - Build and release using goreleaser
   - (Optional) Publish to package managers

**Reference:** Check pinocchio for workflow examples (if they exist).

### 6. Documentation Updates

**Priority: MEDIUM**

**Changes Required:**
1. Update `AGENT.md`: Remove `XXX` placeholder reference
2. Update `README.md`: Replace `pr-builder` with `prescribe` where appropriate
3. Update command examples in all markdown files
4. Update test scripts: Replace hardcoded paths/binary names

**Files to Review:**
- `AGENT.md` (line 5: `XXX/YYY/FOOO` reference)
- `README.md` (multiple `pr-builder` references)
- `FILTER-SYSTEM.md` (multiple `pr-builder` references)
- `test/*.sh` scripts (hardcoded paths: `/home/ubuntu/pr-builder/pr-builder`, `/tmp/pr-builder-test-repo`)
- All markdown documentation files with command examples

### 7. Package Structure Considerations

**Priority: LOW**

**Current:** Uses `internal/` for all code  
**Target:** Consider `pkg/` for library code if prescribe becomes a library

**Decision:** Keep `internal/` for now. If prescribe becomes a library in the future, extract reusable code to `pkg/`.

## Step-by-Step Transformation Plan

### Phase 1: Critical Renaming (Must Do First)

1. **Update go.mod**
   - Change module path to `github.com/go-go-golems/prescribe`
   - Run `go mod tidy`

2. **Update all import statements**
   - Use find/replace: `github.com/user/pr-builder` → `github.com/go-go-golems/prescribe`
   - Verify all files compile: `go build ./...`

3. **Update root command name**
   - `cmd/root.go`: Change `Use: "pr-builder"` to `Use: "prescribe"`
   - Update command description if needed

4. **Update Makefile**
   - Replace `XXX` with `prescribe`
   - Fix build path (prescribe builds from `./cmd`, not `./cmd/prescribe`)

5. **Remove unused files**
   - Delete `cmd/XXX/main.go` and `cmd/XXX/` directory
   - Consider removing `pr-builder/` directory if unused (verify first)
   - Consider removing or populating `pkg/doc.go` if empty

6. **Test compilation**
   - `go build ./...`
   - `make build`

### Phase 2: Build Infrastructure

1. **Enhance Makefile**
   - Add geppetto-lint targets (if prescribe uses geppetto)
   - Add codeql-local target (optional)
   - Verify all targets work

2. **Enhance lefthook.yml**
   - Update pre-commit to use `lintmax`
   - Add `gosec` and `govulncheck` to pre-push
   - Test hooks: `lefthook run pre-commit`

3. **Create CI/CD workflows**
   - Create `.github/workflows/ci.yml`
   - Create `.github/workflows/release.yml` (if using goreleaser)
   - Test workflows on a test branch

### Phase 3: Documentation and Cleanup

1. **Update documentation**
   - Update `AGENT.md` (remove XXX reference)
   - Update `README.md` (replace pr-builder with prescribe)
   - Update all markdown files with command examples
   - Update test scripts

2. **Verify everything works**
   - Run all tests: `make test`
   - Run linting: `make lint`
   - Build binary: `make build`
   - Test binary: `./dist/prescribe --help`

3. **Final validation**
   - Check for any remaining `pr-builder` references
   - Check for any remaining `XXX` references
   - Verify module path is correct everywhere

## Risk Assessment

### High Risk
- **Import path changes:** Could break if not done atomically
- **Module rename:** Requires careful coordination with go.mod

### Medium Risk
- **Makefile changes:** Build paths might differ from pinocchio
- **CI/CD setup:** New workflows need testing

### Low Risk
- **Documentation updates:** Can be done incrementally
- **Lefthook changes:** Easy to test locally

## Validation Checklist

After transformation, verify:

- [ ] `go mod tidy` succeeds
- [ ] `go build ./...` succeeds
- [ ] `make build` succeeds
- [ ] `make test` passes
- [ ] `make lint` passes
- [ ] Binary name is `prescribe` (not `pr-builder`)
- [ ] Root command is `prescribe` (not `pr-builder`)
- [ ] No `XXX` placeholders remain
- [ ] No `github.com/user/pr-builder` imports remain
- [ ] Makefile build path is correct (`.` not `./cmd/XXX`)
- [ ] Unused directories/files removed (`cmd/XXX/`, possibly `pr-builder/`)
- [ ] Test scripts updated (remove hardcoded paths or make them configurable)
- [ ] CI/CD workflows run successfully
- [ ] Lefthook hooks work correctly
- [ ] Documentation is updated

## Future Considerations

1. **Integration with go-go-golems ecosystem:**
   - Consider using `glazed` for CLI output formatting
   - Consider using `geppetto` for LLM integration
   - Consider using `clay` for command management

2. **Package structure:**
   - If prescribe becomes a library, extract `pkg/` from `internal/`
   - Consider making TUI components reusable

3. **Command structure:**
   - Consider moving to `cmd/prescribe/main.go` pattern (like pinocchio)
   - Consider subcommand organization

4. **Testing:**
   - Add more comprehensive tests
   - Consider integration tests
   - Add CI test coverage reporting

## Related Files

This analysis covers:
- `go.mod` - Module definition
- `main.go` - Entry point
- `cmd/` - All command files
- `internal/` - All internal packages
- `Makefile` - Build configuration
- `lefthook.yml` - Git hooks
- `.golangci.yml` - Linter configuration
- `AGENT.md` - Agent guidelines
- `README.md` - Project documentation

## Summary of Required Changes

### Critical Changes (Must Do)
1. **Module Rename:** `github.com/user/pr-builder` → `github.com/go-go-golems/prescribe` (31 import statements)
2. **Command Name:** `pr-builder` → `prescribe` (root command and binary)
3. **Makefile:** Replace `XXX` with `prescribe` (3 occurrences, fix build path)
4. **Remove Unused:** Delete `cmd/XXX/` directory

### High Priority Changes
1. **Lefthook:** Enhance hooks to match pinocchio (lintmax, gosec, govulncheck)
2. **CI/CD:** Add GitHub Actions workflows
3. **Test Scripts:** Update hardcoded paths to use environment variables

### Medium Priority Changes
1. **Documentation:** Update all markdown files with new command name
2. **AGENT.md:** Remove XXX placeholder reference
3. **Cleanup:** Verify and remove `pr-builder/` directory if unused

### Low Priority / Future Considerations
1. **Session Directory:** Consider renaming `.pr-builder/` to `.prescribe/` (breaking change)
2. **Package Structure:** Consider `pkg/` extraction if prescribe becomes a library
3. **Command Structure:** Consider `cmd/prescribe/main.go` pattern
4. **Ecosystem Integration:** Consider using glazed, geppetto, clay packages

## File Change Summary

| File/Directory | Change Type | Priority | Notes |
|----------------|-------------|----------|-------|
| `go.mod` | Rename module | CRITICAL | Change module path |
| All `cmd/*.go` | Update imports | CRITICAL | 12 files |
| All `internal/*/*.go` | Update imports | CRITICAL | 7+ files |
| `main.go` | Update import | CRITICAL | 1 file |
| `cmd/root.go` | Rename command | CRITICAL | Change Use field |
| `Makefile` | Replace XXX | CRITICAL | 3 occurrences, fix build path |
| `cmd/XXX/` | Delete directory | CRITICAL | Unused |
| `lefthook.yml` | Enhance hooks | HIGH | Match pinocchio |
| `.github/workflows/` | Create CI/CD | HIGH | New directory |
| `test/*.sh` | Update paths | HIGH | 5 scripts |
| `AGENT.md` | Remove XXX | MEDIUM | 1 reference |
| `README.md` | Update examples | MEDIUM | Multiple references |
| All `*.md` | Update examples | MEDIUM | Multiple files |
| `pr-builder/` | Verify/remove | MEDIUM | If unused |

## References

- Pinocchio repository structure (reference implementation)
- go-go-golems organization patterns
- Go module migration best practices
- [Go Modules Documentation](https://go.dev/ref/mod)
