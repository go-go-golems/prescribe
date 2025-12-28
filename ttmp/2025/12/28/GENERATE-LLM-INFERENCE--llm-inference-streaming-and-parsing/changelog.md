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


