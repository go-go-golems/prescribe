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

