# Changelog

## 2025-12-28

- Ticket created for investigating Gemini partial/invalid YAML outputs in `prescribe generate`



## 2025-12-28

Add Gemini stop_reason/usage observability: engine stores finish reason + token counts on Turn metadata; prescribe logs them in debug output. Code commits: geppetto 59a19acf2dbecd209218cca73ce53572560634f2, prescribe 2d8e0a159041bf6ac29e1631be30aeb925a80ba2.

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/geppetto/pkg/steps/ai/gemini/engine_gemini.go — Turn metadata now includes stop_reason + usage
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/internal/api/api.go — Debug log prints stop_reason + usage when present

