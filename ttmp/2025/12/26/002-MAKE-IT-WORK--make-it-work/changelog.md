# Changelog

## 2025-12-26

- Initial workspace created


## 2025-12-26

Seeded ticket tasks and wrote go-go-golems Bubbletea developer guide (bobatea-derived patterns for resize, selection, clipboard, help/toasts).

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/reference/01-go-go-golems-bubbletea-application-guide.md — Developer guide (bobatea/go-go-golems style)
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/tasks.md — Initial task list for requested TUI features


## 2025-12-26

Docs: add thorough prescribe TUI structure analysis (models/screens/messages/state/controller side-effects) + start ticket diary

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/analysis/01-prescribe-tui-structure-models-messages-and-control-flow.md — New analysis doc
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/reference/02-diary.md — Ticket 002 diary


## 2025-12-26

Design: propose modular bobatea-style prescribe TUI (root orchestrator + child models + typed messages + layout propagation + side-effect boundary)

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/design-doc/01-prescribe-tui-modularization-proposal-bobatea-style.md — New modularization design proposal


## 2025-12-26

Docs: add didactic deep-dive of core architecture (Controller + domain + git/session/api flows, invariants, failure modes)

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/analysis/02-core-architecture-controller-domain-model-git-session-api-subsystems.md — New core-architecture analysis doc


## 2025-12-26

Design: clarify controller lifetime (single long-lived controller per TUI run; default session load once at boot)

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/design-doc/01-prescribe-tui-modularization-proposal-bobatea-style.md — Added explicit controller lifetime + session boot-load section


## 2025-12-26

Created comprehensive CLI testing playbook and validated all commands work correctly with hierarchical structure. Documented glob pattern behavior (filename vs path matching).

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/playbooks/01-cli-testing-playbook.md — Testing playbook
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/reference/02-diary.md — Diary entry documenting testing


## 2025-12-26

Refined TUI modularization proposal: added explicit package boundaries, shared events message taxonomy, layout API, Deps interface, controller helper suggestions, and a phased implementation plan with exit criteria.

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/design-doc/01-prescribe-tui-modularization-proposal-bobatea-style.md — Updated design doc
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/reference/02-diary.md — Diary Step 5


## 2025-12-26

Smoke-tested  launch under a pseudo-tty (timeout); UI renders main screen without crashing.

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/reference/02-diary.md — Updated Step 4 with TUI smoke test


## 2025-12-26

Expanded refactor work into a detailed phased task breakdown (scaffolding → app root → controller helpers → components → features → cleanup + testing). Marked Bubbletea guide task complete.

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/tasks.md — Added detailed tasks


## 2025-12-26

Phase 1 scaffolding started: added shared events, layout helper + tests, centralized keymap, Styles struct, and status/toast component with ID-safe expiry.

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/internal/tui/components/status/model.go — Status footer model
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/internal/tui/events/events.go — Shared typed messages
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/internal/tui/keys/keymap.go — Central keymap
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/internal/tui/layout/layout.go — Layout Compute()
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/internal/tui/styles/styles.go — Styles struct

