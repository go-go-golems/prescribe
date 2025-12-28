# Changelog

## 2025-12-28

Created ticket `GENERATE-LLM-INFERENCE` by splitting inference-related tasks and analysis out of `008-GENERATE`.

## 2025-12-28

Implemented robust final YAML extraction/parsing for PR output (title/body/changelog/release_notes). The inference service now best-effort parses the assistant output into a structured result (preferring the last fenced YAML block) and stores it on `domain.PRData` alongside the raw assistant text.

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/internal/api/prdata_parse.go — YAML extraction + parsing helpers
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/internal/api/api.go — Wire parsing into `GenerateDescription`
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/internal/domain/domain.go — Add `GeneratedPRData` to `PRData`
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/internal/controller/controller.go — Store parsed fields on controller data

## 2025-12-28

Added stdio streaming mode for inference: `prescribe generate --stream` streams partial completions/events to stderr via `events.EventRouter` + `middleware.NewWatermillSink` while still producing a final result.

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/internal/api/api.go — `GenerateDescriptionStreaming` (engine sink + router)
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/internal/controller/controller.go — `GenerateDescriptionStreaming` wrapper
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/cmd/prescribe/cmds/generate.go — `--stream` flag wiring

## 2025-12-28

Improved streaming UX by printing a final parsed PR-data summary (YAML) to stderr at the end of `--stream` runs (or a clear parse-failed marker). This keeps stdout semantics stable for the final description output.

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/cmd/prescribe/cmds/generate.go — End-of-run parsed summary printing for `--stream`

## 2025-12-28

Updated user-facing documentation for the new `generate` options (including `--stream`) and added a playbook for writing a strong PR description from scratch in a Go repo with many commits.

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/pkg/doc/topics/02-how-to-generate-pr-description.md — Document `--stream` semantics
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/README.md — Update generate usage/options
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/ttmp/2025/12/28/GENERATE-LLM-INFERENCE--llm-inference-streaming-and-parsing/playbook/01-playbook-write-a-great-pr-description-from-scratch-go-repo-many-commits.md — New playbook


