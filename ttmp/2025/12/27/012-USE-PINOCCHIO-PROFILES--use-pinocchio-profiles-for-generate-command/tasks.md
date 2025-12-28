# Tasks

## TODO

- [x] Add tasks here

- [x] Implement glazed/pkg/appconfig.WithProfile (bootstrap-parse profile-settings, then load profiles.yaml at correct precedence)
- [x] Add unit tests covering WithProfile selection via env/config/cobra + error behavior when profiles.yaml missing
- [x] Run go test ./... and commit WithProfile work; update diary with commands + commit hash (glazed commit: 9c37e8d)
- [x] Create/reuse small test git repo via TOKEN-COUNT-DISCREPANCY/scripts/repro-small-repo-balloon.sh; use it for integration testing (we reuse the same underlying helper `test-scripts/setup-test-repo.sh` in ticket smoke tests; prescribe commit: c4f7a31)
- [x] Fix appconfig.WithProfile bootstrap selection so it truly uses only configured sources (env/config/cobra) and not unconditionally env (glazed commit: 15c63ab)
- [x] Add tests proving profile selection resolution from config/env/cobra and correct precedence defaults < profiles < config/env/flags (glazed commit: 15c63ab)
- [x] Integration test in small repo: run a minimal Go snippet or command to validate WithProfile end-to-end, then go test and commit (glazed example + script; prescribe commit: 1a1169c)
- [x] Prescribe: wire generate command to load pinocchio profiles.yaml (bootstrap-parse profile-settings, apply GatherFlagsFromProfiles in middleware chain) (prescribe commit: ca8da4e)
- [x] Prescribe: add small-repo integration script to validate generate + --print-parsed-parameters shows profiles step (prescribe commit: 7d37195)
- [x] Design+impl: Persist PR Title/Description in session.yaml (internal/session.Session + internal/domain.PRData) and plumb into BuildGenerateDescriptionRequest + template vars (.title/.description) (prescribe commit: da26af2)
- [x] CLI: Add generate --title/--description flags (override session values) and session init --title/--description (persist when --save) (prescribe commit: 46a2c0f)
- [x] Prompt: Update internal/prompts/assets/create-pull-request.yaml with clearer upfront framing (PR description task + YAML-only contract) and avoid emitting empty 'The description of the pull request is: .' (prescribe commit: fbfb180)
- [x] Tests: Add/extend unit tests for prompt rendering to assert .title/.description are rendered and no empty description marker is emitted (prescribe commit: fbfb180)
- [x] Small-repo scripts: Add/extend ticket scripts to set session title/description and export-rendered to verify the exact prompt sent contains them (prescribe commit: c4f7a31)
- [x] Gemini robustness: Decide whether to add retry/repair for non-YAML model outputs (detect parse failure; optionally re-ask for YAML-only) (implemented heuristic salvage; prescribe commit: b7e89b0)
