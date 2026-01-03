# Tasks

## TODO

- [x] Decide representation: first-class request field vs new `ContextType` for git history
- [x] Add git service support: commit range + (optional) numstat + (optional) patch extraction
- [x] Wire `.commits` template variable to actual history (prompt contract)
- [x] Add export/debug output section for Git history (and rename “Commits” -> “Commit refs”)
- [x] Add token-count coverage for Git history (if not modeled as `AdditionalContext`)
- [x] Decide persistence: `session.yaml` config block vs computed-only history
