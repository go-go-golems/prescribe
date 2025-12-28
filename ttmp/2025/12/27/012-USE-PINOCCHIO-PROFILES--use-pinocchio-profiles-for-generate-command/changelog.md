# Changelog

## 2025-12-27

- Initial workspace created


## 2025-12-27

Created analysis document and research diary. Analyzed current generate.go implementation, profile system architecture, appconfig package, and bootstrap pattern. Identified need to add WithProfile() option to appconfig and refactor generate.go to use appconfig.Parser.

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/ttmp/2025/12/27/012-USE-PINOCCHIO-PROFILES--use-pinocchio-profiles-for-generate-command/analysis/01-analysis-integrating-pinocchio-profiles-with-generate-command-using-appconfig.md — Comprehensive analysis document
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/ttmp/2025/12/27/012-USE-PINOCCHIO-PROFILES--use-pinocchio-profiles-for-generate-command/reference/01-diary.md — Research diary documenting findings


## 2025-12-28

Implemented appconfig.WithProfile (circularity-safe bootstrap pre-parse of profile-settings + profiles.yaml application), added unit tests and a small example program, and committed in glazed worktree (9c37e8d).

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/glazed/cmd/examples/appconfig-profiles/main.go — Small standalone example for debugging WithProfile
- /home/manuel/workspaces/2025-12-26/prescribe-import/glazed/pkg/appconfig/options.go — Add WithProfile + selection/bootstrap logic
- /home/manuel/workspaces/2025-12-26/prescribe-import/glazed/pkg/appconfig/profile_test.go — Tests for selection/precedence/error behavior


## 2025-12-28

Prescribe generate now loads Pinocchio profiles via bootstrap profile-settings parse; added small-repo smoke test script under ticket scripts/. Commits: ca8da4e (generate wiring), 7d37195 (smoke test).

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/cmd/prescribe/cmds/generate.go — Enable profile flags + bootstrap selection + apply GatherFlagsFromProfiles
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/ttmp/2025/12/27/012-USE-PINOCCHIO-PROFILES--use-pinocchio-profiles-for-generate-command/scripts/01-smoke-test-prescribe-generate-profiles.sh — Reproducible small-repo smoke test


## 2025-12-28

Tightened the embedded `create-pull-request` prompt contract to be YAML-only and to avoid rendering an empty description marker when no note-based description exists. Added a unit test to lock the behavior down. Commit: fbfb180.

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/internal/prompts/assets/create-pull-request.yaml — Conditional description rendering + YAML-only output instructions
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/internal/api/prompt_test.go — Regression test for empty-description rendering


## 2025-12-28

Made PR title/description first-class session state and plumbed them into request building and prompt template variables. Commit: da26af2.

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/internal/domain/domain.go — Add `PRData.Title` / `PRData.Description`
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/internal/session/session.go — Persist `title`/`description` in session.yaml
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/internal/controller/controller.go — Thread into `GenerateDescriptionRequest`
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/internal/api/prompt.go — Map `.title` / `.description` template vars
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/internal/api/prompt_test.go — Tests for title/description rendering


## 2025-12-28

Added `generate --title/--description` overrides and `session init --title/--description` for persisting the fields during session creation. Also exposed `title` and `description_preview` in `session show`. Commit: 46a2c0f.

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/pkg/layers/generation.go — Add `title`/`description` flags to Generation layer
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/cmd/prescribe/cmds/generate.go — Apply flag overrides after session load
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/cmd/prescribe/cmds/session/init.go — Persist title/description when initializing + saving
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/cmd/prescribe/cmds/session/show.go — Show title + description preview


## 2025-12-28

Added a small-repo smoke test to assert session title/description are present in the rendered payload, and fixed `TEST_REPO_DIR` propagation in the existing profiles smoke test. Commit: c4f7a31.

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/ttmp/2025/12/27/012-USE-PINOCCHIO-PROFILES--use-pinocchio-profiles-for-generate-command/scripts/03-smoke-test-prescribe-generate-title-description.sh — New smoke test for title/description rendering
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/ttmp/2025/12/27/012-USE-PINOCCHIO-PROFILES--use-pinocchio-profiles-for-generate-command/scripts/01-smoke-test-prescribe-generate-profiles.sh — Fix test repo directory propagation


## 2025-12-28

Fixed `appconfig.WithProfile` bootstrap selection to only consult env when env parsing is enabled (or explicitly requested), and added a regression test. Commit: 15c63ab.

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/glazed/pkg/appconfig/options.go — Gate bootstrap `UpdateFromEnv` behind env-enabled check
- /home/manuel/workspaces/2025-12-26/prescribe-import/glazed/pkg/appconfig/profile_test.go — Regression test for env-disabled behavior


## 2025-12-28

Improved robustness when models emit prose-wrapped YAML by adding a heuristic salvage parse path (parse from last `title:` block when strict parsing fails). Commit: b7e89b0.

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/internal/api/prdata_parse.go — Heuristic salvage parsing for prose-wrapped YAML
- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/internal/api/prdata_parse_test.go — Test covering prose-wrapped YAML parsing


## 2025-12-28

Added an integration script that exercises `glazed`’s `appconfig.WithProfile` end-to-end via the `cmd/examples/appconfig-profiles` program. Commit: 1a1169c.

### Related Files

- /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/ttmp/2025/12/27/012-USE-PINOCCHIO-PROFILES--use-pinocchio-profiles-for-generate-command/scripts/04-integration-test-glazed-withprofile-example.sh — Validates defaults, env selection, and config override behavior

