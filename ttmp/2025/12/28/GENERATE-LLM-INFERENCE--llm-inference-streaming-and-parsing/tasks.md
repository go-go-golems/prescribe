# Tasks

## TODO

### Streaming (stdio)
- [x] Add `prescribe generate --stream` (or similar) to stream LLM output/events to the terminal using `events.EventRouter` + `middleware.NewWatermillSink`
- [x] Ensure streaming mode still returns a final result (and exit code) and prints the final parsed PR data summary (or raw output) at the end

### Robust final extraction / parsing (non-streaming too)
- [x] Parse assistant output YAML into structured PR result (title/body/changelog/release_notes) using:
  - `geppetto/pkg/steps/parse.ExtractYAMLBlocks` (prefer last fenced YAML block)
  - fallback: `geppetto/pkg/events/structuredsink/parsehelpers.StripCodeFenceBytes`
- [x] Decide where structured PR fields live (extend `domain.PRData` vs parallel result type)

### Structured streaming extraction (optional, later)
- [ ] (moved) Update prompt to optionally emit a tagged block (e.g. `<prescribe:prdata:v1>...`) for structuredsink extraction — moved to `014-STRUCTUREDSINK-STREAMING`
- [ ] (moved) Implement a `prdata` extractor session using `parsehelpers.NewDebouncedYAML` to emit `prdata-update` / `prdata-completed` events — moved to `014-STRUCTUREDSINK-STREAMING`

### (Later) TUI streaming integration
- [ ] (moved) Wire streaming generation into `prescribe tui` — moved to `015-TUI-STREAMING`

## Completed

- [x] Pinocchio-style prompt templating for inference (split combined prompt, render via Glazed templating helpers)
- [x] Deterministic assistant text extraction: extract last assistant `BlockKindLLMText` from final Turn

## Notes

- This ticket is split out of `008-GENERATE` so that `008` can remain focused on deterministic export-context tooling.


