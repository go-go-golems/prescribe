# Changelog

## 2025-12-27

- Initial workspace created
- Added implementation tasks and related key files for filter presets + default filters


## 2025-12-27

Step 1: add controller-level filter preset load/save + tests (code commit bc3149d)

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/internal/controller/filter_presets.go — New controller APIs and YAML schema
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/internal/controller/filter_presets_test.go — New unit tests
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/ttmp/2025/12/27/006-FILTER-PRESETS--filter-presets-and-session-defaults/reference/01-diary.md — Recorded Step 1 diary


## 2025-12-27

Step 2: apply repo default filter presets on boot when session missing (code commit cc52899)

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/internal/controller/repo_defaults.go — Repo config + default preset application
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/internal/tui/app/boot.go — Boot integration
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/internal/tui/app/boot_test.go — Regression test
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/ttmp/2025/12/27/006-FILTER-PRESETS--filter-presets-and-session-defaults/reference/01-diary.md — Recorded Step 2 diary


## 2025-12-27

Step 3: add CLI filter preset list/save/apply (code commit 4880311)

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/cmd/prescribe/cmds/filter/preset_apply.go — Preset apply command
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/cmd/prescribe/cmds/filter/preset_list.go — Preset list command
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/cmd/prescribe/cmds/filter/preset_save.go — Preset save command
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/ttmp/2025/12/27/006-FILTER-PRESETS--filter-presets-and-session-defaults/reference/01-diary.md — Recorded Step 3 diary


## 2025-12-27

Closeout: added example preset fixture; TUI quick presets now loaded from preset dirs; deferred full TUI affordances to 007; decided no auto-save on defaults + no CLI-defaults parity

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/internal/tui/app/filter_presets.go — Load quick presets from project/global preset dirs (task 25)
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/pkg/doc/fixtures/filter-presets/exclude_tests.yaml — Example preset fixture (task 6)
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/ttmp/2025/12/27/007-TUI-FILTER-PRESETS--tui-affordances-for-filter-presets/tasks.md — Follow-up ticket for richer TUI affordances

