# Changelog

## 2025-12-27

- Initial workspace created


## 2025-12-27

Implemented tokenizer-based token counting in prescribe (replacing len(text)/4); session load now recomputes tokens consistently; documented geppetto provider-usage approach. (code: 59f3c73)

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/internal/session/session.go — Recompute tokens on session apply
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/internal/tokens/tokens.go — Tokenizer counter
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/ttmp/2025/12/27/004-PROPER-TOKEN-COUNTING--proper-token-counting/analysis/01-token-counting-geppetto-prescribe.md — Research + decision


## 2026-01-04

Closed (ticket hygiene): tokenizer-based token counting implemented; remaining open task was placeholder.

