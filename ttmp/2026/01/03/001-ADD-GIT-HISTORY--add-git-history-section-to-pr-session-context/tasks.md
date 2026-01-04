# Tasks

## Done

- [x] Decide representation: first-class request field vs new `ContextType` for git history
- [x] Add git service support: commit range + (optional) numstat + (optional) patch extraction
- [x] Wire `.commits` template variable to actual history (prompt contract)
- [x] Add export/debug output section for Git history (and rename “Commits” -> “Commit refs”)
- [x] Add token-count coverage for Git history (if not modeled as `AdditionalContext`)
- [x] Decide persistence: `session.yaml` config block vs computed-only history

## Next: Session-Controlled Git History (`git_history`)

- [x] Add `GitHistoryConfig` to `internal/domain` and persist it in `internal/session` as `git_history:` (enabled/max_commits/include_merges/first_parent/include_numstat)
- [x] Decide compatibility: if `git_history` missing in session, treat as enabled defaults (recommended) vs disabled
- [x] Parameterize `internal/git` history builder with config knobs (merges/first-parent/max commits/detail level)
- [x] Make history injection conditional in `internal/controller/controller.go` based on session config
- [x] Add `prescribe context git history show` to print effective config + derived range (`target..source`)
- [x] Add `prescribe context git history enable|disable` (mutate session, save)
- [x] Add `prescribe context git history set` (mutate session fields, save)
- [x] Add smoke coverage: disabling history removes `BEGIN COMMITS` in `generate --export-rendered`

## Next: Explicit Git Context Items (`git_context`)

- [x] Add `GitContextItem` list to `internal/domain` and persist it in `internal/session` as `git_context:` (reference-based; no embedded blobs)
- [x] Define supported kinds + schema fields (commit, commit_patch, file_at_ref, file_diff)
- [x] Extend `internal/git` with helpers to materialize items:
- [x] commit metadata + numstat summary for a ref
- [x] commit patch (optionally path-filtered) for a ref
- [x] file content at ref
- [x] file diff between refs
- [x] Materialize `git_context` into `GenerateDescriptionRequest.AdditionalContext` in `internal/controller/controller.go` with:
- [x] stable, strongly delimited formatting (xml-ish or begin/end blocks)
- [x] per-item token/byte caps + explicit truncation markers
- [x] new context `Type` values for each kind (so exporters can format cleanly)
- [x] Add `prescribe context git list` (show configured items with indices + summary)
- [x] Add `prescribe context git remove <index>` and `prescribe context git clear`
- [x] Add `prescribe context git add commit <ref>` (+ flags for patch/paths/numstat)
- [x] Add `prescribe context git add commit-patch <ref>` (+ `--path ...` repeatable)
- [x] Add `prescribe context git add file-at <ref> <path>`
- [x] Add `prescribe context git add file-diff --from <ref> --to <ref> --path <path>`
- [x] Add smoke coverage: a git_context item shows up in `generate --export-context` and `generate --export-rendered`

## Docs

- [x] Update `README.md` session format example to include `git_history:` and `git_context:` (after implementation)
- [x] Add `README.md` usage examples for `prescribe context git history ...` and `prescribe context git add ...`

## Next: CLI Refactor (Glazed-first, root.go registration)

### Design + Prep

- [ ] Confirm command type rule: use Glazed `BareCommand` for non-table output; use `GlazeCommand` for tabular/structured output.
- [ ] Finalize target directory layout (`cmd/prescribe/cmds/<group>/<subgroup...>/root.go` + one file per verb).
- [ ] Enumerate current commands → target file paths (mapping table in design doc).
- [ ] Decide standard “leaf constructor” naming: `New<Verb>CobraCommand()` returning `*cobra.Command`.
- [ ] Add/extend smoke test that walks `prescribe --help` and key subgroup `--help` outputs (tree visibility).

### Phase 1: Root wiring (no Init methods)

- [x] Migrate `cmd/prescribe/cmds/root.go` to use group constructors (no `Init()` pattern); keep behavior identical.
- [ ] For each group, replace `Init()` with `New<Group>Cmd()` in `<group>/root.go` and update root imports:
- [x] `context`
  - [ ] `filter`
  - [ ] `session`
  - [ ] `file`
  - [ ] `tokens`
- [ ] For root-level verbs, decide whether they live under `cmd/prescribe/cmds/root/` subgroup or remain in `cmd/prescribe/cmds/` with constructors:
  - [ ] `generate`
  - [ ] `create`
  - [ ] `tui`

### Phase 2: Context group (split subtrees, one file per verb)

- [x] Replace `cmd/prescribe/cmds/context/*.go` with:
- [x] `cmd/prescribe/cmds/context/root.go` (registers `add`, `git`)
- [x] `cmd/prescribe/cmds/context/add.go` (BareCommand)
- [ ] Split `context git` tree into subpackages:
- [x] `cmd/prescribe/cmds/context/git/root.go` (registers `list/remove/clear/add/history`)
- [x] `cmd/prescribe/cmds/context/git/list.go`
- [x] `cmd/prescribe/cmds/context/git/remove.go`
- [x] `cmd/prescribe/cmds/context/git/clear.go`
  - [ ] `cmd/prescribe/cmds/context/git/add/root.go`
  - [ ] `cmd/prescribe/cmds/context/git/add/commit.go`
  - [ ] `cmd/prescribe/cmds/context/git/add/commit_patch.go`
  - [ ] `cmd/prescribe/cmds/context/git/add/file_at.go`
  - [ ] `cmd/prescribe/cmds/context/git/add/file_diff.go`
  - [ ] `cmd/prescribe/cmds/context/git/history/root.go`
  - [ ] `cmd/prescribe/cmds/context/git/history/show.go`
  - [ ] `cmd/prescribe/cmds/context/git/history/enable.go`
  - [ ] `cmd/prescribe/cmds/context/git/history/disable.go`
  - [ ] `cmd/prescribe/cmds/context/git/history/set.go`
- [ ] Ensure all `context` verbs are Glazed commands (BareCommand unless/where we deliberately return rows).
- [x] Remove old `cmd/prescribe/cmds/context/context.go` and monolithic `cmd/prescribe/cmds/context/git.go`.

### Phase 3: Other groups

- [ ] Migrate `filter` group to directory-per-subgroup (including `preset` subtree) with one file per verb.
- [ ] Migrate `session` group to one file per verb and root.go registration.
- [ ] Migrate `file` group to one file per verb and root.go registration.
- [ ] Migrate `tokens` group to one file per verb and root.go registration.

### Phase 4: Cleanup + Validation

- [ ] Ensure `GOWORK=off go test ./...` passes after each group migration.
- [ ] Run smoke scripts after major milestones:
  - [ ] `bash test-scripts/test-cli.sh`
  - [ ] `bash test-scripts/test-all.sh`
- [ ] Ensure no CLI behavior regressions (help text, flag names, default behavior).
- [ ] Update README CLI layout references if any paths/usage changed.
