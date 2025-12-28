# Tasks

## TODO

- [ ] Reproduce mismatch reliably (capture: session show json, exported rendered payload file, exact pinocchio command/token output)
- [ ] Confirm tokenizer/encoding on both sides (prescribe vs pinocchio) and document differences
- [ ] Determine whether mismatch is “different input text” vs “different encoding” vs “bug in counting”
- [ ] Add a diagnostic command or output to show per-file token contributions (or token_count for rendered payload)
- [ ] Fix or document expected semantics (session token_count vs rendered payload token count)

### Debugging tools to build (this ticket)

- [x] Add **verbose per-element session context token breakdown** ("show token count"):
  - [x] One row/entry per included file (diff/full) with `path`, `type`, `included`, `tokens`, and small summary fields (additions/deletions if available)
  - [x] One row/entry per additional context item with `type`, `path` (if file), `tokens`
  - [x] Include `PRESCRIBE_TOKEN_ENCODING` / `tokens.EncodingName()` in output
  - [x] Output should be stable + machine-readable (JSON-friendly)
- [x] Add **post-hoc “XML-ish” token counter** utility:
  - [x] Input: a rendered/exported `.xml` file (not strict XML)
  - [x] Output: token counts per top-level-ish section/tag (best-effort), plus totals
  - [x] Must use the same tokenizer as prescribe (`internal/tokens`)
- [x] Add **rendered payload token count** output:
  - [x] When a flag is passed to `generate`, compute token counts of the rendered payload:
    - [x] total tokens
    - [x] system prompt tokens
    - [x] user prompt tokens
    - [x] (optional) exported envelope tokens when `--separator xml` is used
  - [x] Ensure this works both when exporting rendered payload to file and when generating normally (no inference required)

