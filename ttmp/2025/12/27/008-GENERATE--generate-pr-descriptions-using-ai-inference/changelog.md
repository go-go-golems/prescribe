# Changelog

## 2025-12-27

- Initial workspace created


## 2025-12-27

Completed comprehensive analysis of catter implementation, pinocchio create-pull-request pattern, and geppetto inference engine architecture. Created detailed analysis document with pseudocode and implementation plan. Maintained research diary throughout exploration process.

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/ttmp/2025/12/27/008-GENERATE--generate-pr-descriptions-using-ai-inference/analysis/01-analysis-export-prescribe-diff-data-and-generate-pr-descriptions-with-geppetto-inference.md — Comprehensive analysis document
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/ttmp/2025/12/27/008-GENERATE--generate-pr-descriptions-using-ai-inference/reference/01-diary.md — Research diary documenting exploration process


## 2025-12-27

Updated analysis document to include separator/delimiter approaches for formatting output data. Added section on XML (default), markdown, simple, and begin-end separators. Documented how different content types (diffs, full files, prompts, context) should be formatted with separators. Updated pseudocode to include separator handling.


## 2025-12-27

Added commit version information to separator formatting examples. Diffs now include source-commit and target-commit attributes. Full files include commit attribute. Context files include optional commit attribute.


## 2025-12-27

Added design guide and task list. Implemented export-first milestone:  outputs full prompt/context payload (catter-style). Refactored API service to accept geppetto StepSettings (parsed higher up) and use Turns directly.


## 2025-12-27

Implemented export-first milestone: prescribe generate --export-context --separator xml prints the full prompt/context payload (no inference). Added design guide and task list. Refactored API service to accept geppetto StepSettings (parsed in CLI/TUI) and use Turns directly.


## 2025-12-27

CLI: --export-context now honors --output-file. Documented --export-context/--separator in help topic (how-to-generate-pr-description).

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/cmd/prescribe/cmds/generate.go — Export context to stdout or --output-file
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/pkg/doc/topics/02-how-to-generate-pr-description.md — Docs for export-context and separator


## 2025-12-27

Updated standard CLI testing playbook to include generate --export-context and corrected generate flag names. Refreshed ticket tasks for handoff (Next vs Later).

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/playbooks/01-cli-testing-playbook.md — Added export-context section and fixed flags
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/ttmp/2025/12/27/008-GENERATE--generate-pr-descriptions-using-ai-inference/tasks.md — Handoff-ready task list


## 2025-12-28

Added regression coverage for `prescribe generate --export-context` in `test/test-cli.sh` and `test/test-all.sh` (all separators + `--output-file`). Fixed those shell tests to rebuild a per-HEAD binary so they don’t accidentally exercise a stale `/tmp/prescribe`. Removed the duplicate markdown exporter under `internal/tui/export` and moved its unit test to `internal/export`. (commit `1b25b00`)


## 2025-12-28

Implemented Pinocchio-style prompt templating for inference using Glazed’s templating helpers (`sprig` + `TemplateFuncs`). The combined default prompt is now split into system vs user templates at `{{ define "context" ... }}` and rendered with variables derived from the generation request (diff/full files/context notes/files). Added unit tests for templated and fallback behavior. (commit `fd6eeed`)


## 2025-12-28

Added a detailed analysis doc covering the end-to-end pipeline: render templates via Glazed helpers, run inference in streaming mode (Watermill sink + EventRouter), and extract/parse structured PR YAML output (including an optional structured-tag + extractor approach). (doc: `analysis/02-analysis-template-rendering-streaming-and-prdata-extraction.md`)

