# Changelog

## 2025-12-27

- Initial workspace created


## 2025-12-27

Created comprehensive analysis of Prescribe commands and Glazed framework patterns. Documented all 14 commands, identified 4 reusable layers, and created detailed mapping strategy. Research diary tracks learning process and API symbols.

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/ttmp/2025/12/27/005-PORT-GLAZED--port-prescribe-commands-to-glazed-framework/analysis/01-command-mapping-analysis.md — Comprehensive command mapping analysis
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/ttmp/2025/12/27/005-PORT-GLAZED--port-prescribe-commands-to-glazed-framework/reference/01-research-diary.md — Research diary documenting analysis process


## 2025-12-27

Rewrote command mapping analysis document to be more readable with narrative prose, explanatory paragraphs, and better flow. Document now follows technical writing guidelines with context before structure, explanations of 'why' behind decisions, and bridging concepts with narrative.

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/ttmp/2025/12/27/005-PORT-GLAZED--port-prescribe-commands-to-glazed-framework/analysis/01-command-mapping-analysis.md — Rewritten with narrative flow and prose paragraphs


## 2025-12-27

Updated document to use newer schema/fields API aliases instead of older layers/parameters API. Changed all examples to use schema.NewSection() and fields.New() for clearer naming. Updated API reference section to emphasize the newer API.

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/ttmp/2025/12/27/005-PORT-GLAZED--port-prescribe-commands-to-glazed-framework/analysis/01-command-mapping-analysis.md — Updated to use schema/fields API


## 2025-12-27

Created comprehensive task list for porting all 14 commands to Glazed. Tasks organized by migration phases: Phase 1 (infrastructure/layers), Phase 2 (query commands with structured output), Phase 3 (state modification commands), Phase 4 (simple commands), plus testing and documentation tasks.

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/ttmp/2025/12/27/005-PORT-GLAZED--port-prescribe-commands-to-glazed-framework/tasks.md — Task list with 26 tasks covering all migration work


## 2025-12-27

Added section on Glazed program initialization explaining how to set up root command with logging and help system. Added tasks for initializing Prescribe as a Glazed program.

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/ttmp/2025/12/27/005-PORT-GLAZED--port-prescribe-commands-to-glazed-framework/analysis/01-command-mapping-analysis.md — Added Glazed Program Initialization section


## 2025-12-27

Step 1: Initialize Prescribe with Glazed logging + help system and embed Prescribe help topics (commit 90d7951)

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/cmd/prescribe/cmds/root.go — Logging init in PersistentPreRunE
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/cmd/prescribe/main.go — Glazed init wiring
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/pkg/doc/embed.go — Embedded docs source


## 2025-12-27

Step 2: Add Phase 1 Glazed parameter layers package (repository/session/filter/generation) with settings helpers (commit cb59b50)

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/pkg/layers/filter.go — New schema section
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/pkg/layers/generation.go — New schema section
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/pkg/layers/repository.go — New schema section
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/pkg/layers/session.go — New schema section


## 2025-12-27

Step 3: Add controller initialization helper based on Glazed parsed layers (commit 8e294d1)

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/cmd/prescribe/cmds/helpers/controller_from_layers.go — New helper for Glazed command ports


## 2025-12-27

Step 4: Remove init()-based Cobra wiring; add explicit Init() funcs + first dual-mode Glazed command (filter list) (commit da425db)

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/cmd/prescribe/cmds/filter/list.go — Dual-mode Glazed output
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/cmd/prescribe/cmds/root.go — Deterministic initialization


## 2025-12-27

Step 5: Port filter show to dual-mode (classic + Glaze structured output) (commit 3f05fca)

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/cmd/prescribe/cmds/filter/show.go — Dual-mode command wiring + row schema


## 2025-12-27

Step 6: Port filter test to dual-mode (classic + Glaze structured output) (commit 3a028b5)

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/cmd/prescribe/cmds/filter/test.go — Dual-mode command + row schema


## 2025-12-27

Step 7: Port session show to dual-mode (classic + Glaze structured output) (commit 425af79)

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/cmd/prescribe/cmds/session/show.go — Dual-mode command + --yaml handling


## 2025-12-27

Step 8: Remove no-op init() (explicit init preference) (commit fc233a3)

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/pkg/doc.go — Deleted empty init file


## 2025-12-27

Step 9: Port filter add to Glazed-built BareCommand (legacy flags preserved) (commit 451d28b)

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/cmd/prescribe/cmds/filter/add.go — BareCommand + layer schema + controller init from parsed layers


## 2025-12-27

Step 10: Drop dual-mode/back-compat glue; ported query commands now always use Glazed output (commit 9860e32)

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/cmd/prescribe/cmds/session/show.go — Removed classic + yaml flag


## 2025-12-27

Docs: Add onboarding playbook for porting Cobra verbs to Glazed (no back-compat)

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/ttmp/2025/12/27/005-PORT-GLAZED--port-prescribe-commands-to-glazed-framework/playbook/01-playbook-port-existing-cobra-verbs-to-glazed-no-back-compat.md — New playbook


## 2025-12-27

Step 11: Port `filter remove` and `filter clear` to Glazed BareCommands (commit be52063)

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/cmd/prescribe/cmds/filter/remove.go — Glazed BareCommand with positional arg parsing via schema.DefaultSlug
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/cmd/prescribe/cmds/filter/clear.go — Glazed BareCommand for clearing filters and saving session
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/cmd/prescribe/cmds/filter/filter.go — Explicit init wiring for remove/clear


## 2025-12-27

Step 12: Port `session load` and `session save` to Glazed BareCommands (commit a2e2bca)

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/cmd/prescribe/cmds/session/load.go — Glazed BareCommand with optional positional `[path]`
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/cmd/prescribe/cmds/session/save.go — Glazed BareCommand with optional positional `[path]`
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/cmd/prescribe/cmds/session/session.go — Explicit init wiring for load/save


## 2025-12-27

Step 13: Port `session init` to a Glazed BareCommand (commit b28b057)

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/cmd/prescribe/cmds/session/init.go — Glazed BareCommand + `session-init` section for `--save/--path`


## 2025-12-27

Step 14: Port `file toggle` to a Glazed BareCommand (commit 8909c02)

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/cmd/prescribe/cmds/file/toggle.go — Glazed BareCommand with required positional `<path>`
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/cmd/prescribe/cmds/file/file.go — Explicit init wiring for toggle


## 2025-12-27

Step 15: Port `context add` to a Glazed BareCommand (commit 519a959)

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/cmd/prescribe/cmds/context/add.go — Glazed BareCommand with `--note` + optional positional `[file-path]` and mutual exclusion validation

