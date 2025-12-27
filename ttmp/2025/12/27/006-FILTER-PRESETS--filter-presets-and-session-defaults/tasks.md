# Tasks

## TODO

- [x] **Define filter preset YAML schema + locations**
- [x] Decide schema (mirror prompt presets): `name`, `description`, `rules[{type,pattern}]`
- [x] Decide directories:
- [x] Project: `<repo>/.pr-builder/filters/*.yaml`
- [x] Global: `~/.pr-builder/filters/*.yaml`
- [x] Add an example preset file in docs/test fixtures

- [x] **Implement filter preset load/save in controller**
- [x] Add `LoadProjectFilterPresets()` and `LoadGlobalFilterPresets()` (modeled after prompt presets)
- [x] Add `SaveFilterPreset(name, description, rules, location)` writing YAML
- [x] Decide preset ID strategy (filename vs slug) and keep it stable

- [x] **Implement “default filters for new sessions in current repo”**
- [x] Define per-repo config file (proposed): `<repo>/.pr-builder/config.yaml`
- [x] Add `defaults.filter_presets: [ ... ]` list (IDs or filenames)
- [x] On TUI boot (`internal/tui/app/boot.go`):
- [x] If session missing, load repo defaults and apply them
- [x] Decide whether to auto-save a new session after applying defaults
- [x] CLI parity: decide whether CLI commands should also apply defaults when session is missing

- [x] **Expose preset management via CLI**
  - [ ] Add subcommands:
- [x] `prescribe filter preset list [--project|--global|--all]`
- [x] `prescribe filter preset save --name ... [--project|--global] [--from-active|--exclude/--include ...]`
- [x] `prescribe filter preset apply PRESET_ID`
- [x] Ensure commands do not clobber existing sessions (load default session if exists before mutating)

- [x] **TUI integration**
- [x] Decide whether “quick presets” remain hardcoded or become data-driven from preset dirs
- [x] Add UI affordance to “save current active filters as preset” (project/global)
- [x] Add UI affordance to apply a saved preset

- [x] **Tests + docs**
- [x] Unit tests: YAML round-trip for presets (load/save)
- [x] Unit tests: boot behavior (missing session + defaults applied)
- [x] Update docs: explain preset locations and default behavior

