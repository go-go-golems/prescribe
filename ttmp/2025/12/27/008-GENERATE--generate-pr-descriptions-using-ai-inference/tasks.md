# Tasks

## TODO

- [x] (cleanup) Remove placeholder “Add tasks here”

- [x] Add catter-style context exporter for GenerateDescriptionRequest (supports separator types; default xml)
- [x] Wire 'prescribe generate' flag to output exported context string (no inference)
- [x] Add 'prescribe generate' flag for separator selection (xml/markdown/simple/begin-end/default)
- [x] Refactor TUI CopyContext to reuse the same exporter (keep clipboard behavior)
- [x] Update docs/help for new generate export flag and separator flag

## Next (handoff-ready)

- [x] Update “standard CLI testing playbook” to include `generate --export-context` and correct flag names
- [x] Extend `prescribe/test/test-all.sh` and/or `prescribe/test/test-cli.sh` to exercise `generate --export-context` (all separators; file output)
- [x] Update `prescribe/pkg/doc/topics/02-how-to-generate-pr-description.md` to mention `--output-file` works with `--export-context` (already documented) and keep examples consistent
- [x] Remove duplicate markdown exporter (`internal/tui/export`) or explicitly document why we keep both (avoid drift)

## Later (post-export milestone)

- [x] Reconcile commit/branch metadata in exporter (include source/target commit hashes; include commit ref for context files)

## Next

- [x] Add `prescribe generate --export-rendered` to export the **rendered** (templated) prompt payload as text (no inference), mirroring Pinocchio-style template rendering
- [x] Document and test `--export-rendered` (works with `--output-file`; clarify interaction with `--separator` and `--export-context`)
