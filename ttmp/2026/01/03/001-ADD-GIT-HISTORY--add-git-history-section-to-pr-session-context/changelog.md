# Changelog

## 2026-01-03

- Initial workspace created


## 2026-01-03

Created research analysis and diary; identified prompt pipeline touchpoints and .commits contract for adding Git history to PR session context.

### Related Files

- /home/manuel/workspaces/2026-01-03/add-git-history-prescribe/prescribe/ttmp/2026/01/03/001-ADD-GIT-HISTORY--add-git-history-section-to-pr-session-context/analysis/01-research-git-history-section-for-pr-session-context.md — Design/research document
- /home/manuel/workspaces/2026-01-03/add-git-history-prescribe/prescribe/ttmp/2026/01/03/001-ADD-GIT-HISTORY--add-git-history-section-to-pr-session-context/reference/01-diary.md — Research diary entries


## 2026-01-03

Expanded diary and analysis with design options (first-class .commits field recommended) and token-budget defaults; updated ticket tasks for follow-up implementation.

### Related Files

- /home/manuel/workspaces/2026-01-03/add-git-history-prescribe/prescribe/ttmp/2026/01/03/001-ADD-GIT-HISTORY--add-git-history-section-to-pr-session-context/reference/01-diary.md — Added design decision notes (Step 3)
- /home/manuel/workspaces/2026-01-03/add-git-history-prescribe/prescribe/ttmp/2026/01/03/001-ADD-GIT-HISTORY--add-git-history-section-to-pr-session-context/tasks.md — Implementation task breakdown


## 2026-01-03

Cleaned up RelatedFiles notes after an initial zsh backtick substitution issue; ensured .commits references render correctly in ticket metadata.

### Related Files

- /home/manuel/workspaces/2026-01-03/add-git-history-prescribe/prescribe/ttmp/2026/01/03/001-ADD-GIT-HISTORY--add-git-history-section-to-pr-session-context/analysis/01-research-git-history-section-for-pr-session-context.md — Fixed truncated RelatedFiles note
- /home/manuel/workspaces/2026-01-03/add-git-history-prescribe/prescribe/ttmp/2026/01/03/001-ADD-GIT-HISTORY--add-git-history-section-to-pr-session-context/index.md — Fixed duplicated/blank notes for RelatedFiles


## 2026-01-03

Implemented git history section in generation context and rendered payload; updated token-count and augmented mock-repo smoke scripts to assert history is present.

### Related Files

- /home/manuel/workspaces/2026-01-03/add-git-history-prescribe/prescribe/internal/controller/controller.go — Inject git_history context item during request build
- /home/manuel/workspaces/2026-01-03/add-git-history-prescribe/prescribe/internal/git/git.go — BuildCommitHistoryText implementation
- /home/manuel/workspaces/2026-01-03/add-git-history-prescribe/prescribe/test-scripts/test-cli.sh — Smoke test coverage for BEGIN COMMITS + author


## 2026-01-03

Cleaned up ticket index RelatedFiles notes for the new git history implementation.

### Related Files

- /home/manuel/workspaces/2026-01-03/add-git-history-prescribe/prescribe/ttmp/2026/01/03/001-ADD-GIT-HISTORY--add-git-history-section-to-pr-session-context/index.md — Removed duplicated RelatedFiles notes


## 2026-01-03

Committed git history feature implementation (commit 362c0f6).

### Related Files

- /home/manuel/workspaces/2026-01-03/add-git-history-prescribe/prescribe/internal/git/git.go — Commit history extraction
- /home/manuel/workspaces/2026-01-03/add-git-history-prescribe/prescribe/test-scripts/test-all.sh — Smoke coverage for history


## 2026-01-03

Added design doc proposing session.yaml `git_history` config and `prescribe context git history ...` verbs for explicit, controllable history inclusion.

### Related Files

- /home/manuel/workspaces/2026-01-03/add-git-history-prescribe/prescribe/ttmp/2026/01/03/001-ADD-GIT-HISTORY--add-git-history-section-to-pr-session-context/design-doc/01-session-git-history-config-and-context-git-verbs.md — Schema + CLI UX proposal
- /home/manuel/workspaces/2026-01-03/add-git-history-prescribe/prescribe/ttmp/2026/01/03/001-ADD-GIT-HISTORY--add-git-history-section-to-pr-session-context/tasks.md — Added follow-up implementation tasks

## 2026-01-03

Extended design: add session.yaml git_context list and prescribe context git verbs for adding specific commits, commit patches, file-at-ref snapshots, and file diffs.

### Related Files

- /home/manuel/workspaces/2026-01-03/add-git-history-prescribe/prescribe/ttmp/2026/01/03/001-ADD-GIT-HISTORY--add-git-history-section-to-pr-session-context/design-doc/01-session-git-history-config-and-context-git-verbs.md — Added git_context schema + verb proposals
- /home/manuel/workspaces/2026-01-03/add-git-history-prescribe/prescribe/ttmp/2026/01/03/001-ADD-GIT-HISTORY--add-git-history-section-to-pr-session-context/reference/01-diary.md — Added Step 6 design notes
- /home/manuel/workspaces/2026-01-03/add-git-history-prescribe/prescribe/ttmp/2026/01/03/001-ADD-GIT-HISTORY--add-git-history-section-to-pr-session-context/tasks.md — Added git_context follow-up tasks

## 2026-01-04

Step 7: Implement session git_history + git_context controls (commit 53272bb)

### Related Files

- /home/manuel/workspaces/2026-01-03/add-git-history-prescribe/prescribe/cmd/prescribe/cmds/context/git.go — CLI verbs for context git history + git_context items
- /home/manuel/workspaces/2026-01-03/add-git-history-prescribe/prescribe/internal/controller/controller.go — Conditional history injection + materialize git_context at request build
- /home/manuel/workspaces/2026-01-03/add-git-history-prescribe/prescribe/internal/git/context_items.go — Materialize commit/patch/file-at/file-diff with caps and truncation markers
- /home/manuel/workspaces/2026-01-03/add-git-history-prescribe/prescribe/internal/session/session.go — Persist git_history and git_context in session.yaml
- /home/manuel/workspaces/2026-01-03/add-git-history-prescribe/prescribe/test-scripts/test-cli.sh — Smoke coverage for disabling history and explicit git_context items


## 2026-01-04

Docs: architecture analysis + plugin-based context provider design

### Related Files

- /home/manuel/workspaces/2026-01-03/add-git-history-prescribe/prescribe/ttmp/2026/01/03/001-ADD-GIT-HISTORY--add-git-history-section-to-pr-session-context/analysis/02-architecture-current-structure-and-modularization-opportunities.md — Document current architecture and modularization seams
- /home/manuel/workspaces/2026-01-03/add-git-history-prescribe/prescribe/ttmp/2026/01/03/001-ADD-GIT-HISTORY--add-git-history-section-to-pr-session-context/design-doc/02-plugin-based-context-providers-proposed-architecture-and-migration-plan.md — Propose provider/registry architecture and migration plan


## 2026-01-04

Closed during ticket hygiene cleanup: all tasks complete; implementation + docs landed. See 002-CLEANUP-OLD-TICKETS.


## 2026-01-04

Docs: design CLI refactor to Glazed-first command layout

### Related Files

- /home/manuel/workspaces/2026-01-03/add-git-history-prescribe/prescribe/ttmp/2026/01/03/001-ADD-GIT-HISTORY--add-git-history-section-to-pr-session-context/design-doc/03-refactor-cli-migrate-cobra-verbs-to-glazed-and-reorganize-command-packages.md — Plan to migrate remaining Cobra-only verbs and reorganize cmd layout


## 2026-01-04

Docs: refine Glazed-first CLI refactor (root.go registration; no Init methods)

### Related Files

- /home/manuel/workspaces/2026-01-03/add-git-history-prescribe/prescribe/ttmp/2026/01/03/001-ADD-GIT-HISTORY--add-git-history-section-to-pr-session-context/design-doc/03-refactor-cli-migrate-cobra-verbs-to-glazed-and-reorganize-command-packages.md — Update registration plan to root.go ownership and remove Init pattern
- /home/manuel/workspaces/2026-01-03/add-git-history-prescribe/prescribe/ttmp/2026/01/03/001-ADD-GIT-HISTORY--add-git-history-section-to-pr-session-context/reference/01-diary.md — Record Step 10 design update

