# Tasks

## TODO

- [ ] Reproduce mismatch reliably (capture: session show json, exported rendered payload file, exact pinocchio command/token output)
- [ ] Confirm tokenizer/encoding on both sides (prescribe vs pinocchio) and document differences
- [ ] Determine whether mismatch is “different input text” vs “different encoding” vs “bug in counting”
- [ ] Add a diagnostic command or output to show per-file token contributions (or token_count for rendered payload)
- [ ] Fix or document expected semantics (session token_count vs rendered payload token count)

### Debugging tools to build (this ticket)

- [ ] Add **verbose per-element session context token breakdown** ("show token count"):
  - [ ] One row/entry per included file (diff/full) with `path`, `type`, `included`, `tokens`, and small summary fields (additions/deletions if available)
  - [ ] One row/entry per additional context item with `type`, `path` (if file), `tokens`
  - [ ] Include `PRESCRIBE_TOKEN_ENCODING` / `tokens.EncodingName()` in output
  - [ ] Output should be stable + machine-readable (JSON-friendly)
- [ ] Add **post-hoc “XML-ish” token counter** utility:
  - [ ] Input: a rendered/exported `.xml` file (not strict XML)
  - [ ] Output: token counts per top-level-ish section/tag (best-effort), plus totals
  - [ ] Must use the same tokenizer as prescribe (`internal/tokens`)
- [ ] Add **rendered payload token count** output:
  - [ ] When a flag is passed to `generate`, compute token counts of the rendered payload:
    - [ ] total tokens
    - [ ] system prompt tokens
    - [ ] user prompt tokens
    - [ ] (optional) exported envelope tokens when `--separator xml` is used
  - [ ] Ensure this works both when exporting rendered payload to file and when generating normally (no inference required)

