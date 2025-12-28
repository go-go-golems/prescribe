# Changelog

## 2025-12-28

- Ticket created for investigating Gemini partial/invalid YAML outputs in `prescribe generate`



## 2025-12-28

Add Gemini stop_reason/usage observability: engine stores finish reason + token counts on Turn metadata; prescribe logs them in debug output. Code commits: geppetto 59a19acf2dbecd209218cca73ce53572560634f2, prescribe 2d8e0a159041bf6ac29e1631be30aeb925a80ba2.

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/geppetto/pkg/steps/ai/gemini/engine_gemini.go — Turn metadata now includes stop_reason + usage
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/internal/api/api.go — Debug log prints stop_reason + usage when present


## 2025-12-28

Root cause: Gemini truncation due to max tokens; prescribe wasn't inheriting ai-max-response-tokens from ~/.pinocchio/config.yaml. Fix: load pinocchio config as defaults overlay + add stream-vs-nonstream repro script. Commits: prescribe d290191c523e..., 0e5fce96e830..., geppetto 0f70eaf2f0c7..., pinocchio 3a4e5947a75c....

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/geppetto/pkg/layers/layers.go — Fix profiles middleware signature
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/cmd/prescribe/cmds/generate.go — Load pinocchio config defaults overlay
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/ttmp/2025/12/28/013-FIX-GEMINI-YAML-INFERENCE--fix-gemini-yaml-inference/scripts/01-compare-gemini-streaming-vs-nonstreaming.sh — Repro script artifacts

