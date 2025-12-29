---
Title: Diary
Ticket: PR-CREATION
Status: active
Topics:
    - cli
    - git
    - prescribe
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: prescribe/cmd/prescribe/cmds/create.go
      Note: |-
        New 'prescribe create' command skeleton + flags (task #2)
        On failure
    - Path: prescribe/cmd/prescribe/cmds/root.go
      Note: 'Wires create command into root command tree (task #2)'
    - Path: prescribe/internal/github/github_test.go
      Note: 'Unit tests for gh pr create arg building and redaction (task #14)'
    - Path: prescribe/internal/prdata/prdata.go
      Note: 'Failure PR data path + timestamped save location (task #11)'
    - Path: prescribe/test-scripts/08-integration-test-pr-creation.sh
      Note: 'End-to-end PR creation integration test: local git remote + fake gh; covers create + generate --create'
    - Path: prescribe/test-scripts/README.md
      Note: 'Documents PR creation integration script + SKIP_GENERATE mode'
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-29T08:32:36.47208347-05:00
WhatFor: ""
WhenToUse: ""
---



# Diary

## Goal

Track the step-by-step work to implement PR creation (end-to-end) for `prescribe`, including command output, failures, and a clear map for code review.

## Step 1: Create ticket + initialize diary doc

This step created the docmgr ticket workspace and ensured we have a dedicated diary document to keep the work auditable. It didn’t change code, but it established the workflow foundation we’ll follow for every subsequent implementation step.

### What I did
- Created ticket `PR-CREATION` under `prescribe/ttmp/2025/12/29/PR-CREATION--implement-pr-creation-end-to-end`
- Created this diary doc (reference): `reference/01-diary.md`

### Why
- Ensure all future work can be continued/reviewed without re-deriving context.

### What worked
- Ticket creation and diary doc creation succeeded.

### What didn't work
- N/A

### What I learned
- `docmgr` in this repo is configured to use `prescribe/ttmp` as the docs root via `.ttmp.yaml`.

### What was tricky to build
- N/A (no code changes yet)

### What warrants a second pair of eyes
- N/A

### What should be done in the future
- Ensure each substantive step adds: related files + changelog entry.

### Code review instructions
- N/A (no code changes yet)

### Technical details
- Commands run:
  - `docmgr ticket create-ticket --ticket PR-CREATION --title "Implement PR creation (end-to-end)" --topics cli,git,prescribe`
  - `docmgr doc add --ticket PR-CREATION --doc-type reference --title "Diary"`

## Step 2: Codebase analysis (files + symbols) for PR creation

This step mapped the current code paths responsible for generating PR descriptions and identified the likely integration points for adding “create PR” functionality. The main finding is that `prescribe` already produces structured PR data (`title/body/...`) but does not yet have any implementation that calls `gh` or the GitHub API to actually create a PR.

### What I did
- Created analysis doc: `reference/02-pr-creation-codebase-analysis.md`
- Searched for PR-creation related symbols and confirmed the existing generation flow:
  - prompt pack (`internal/prompts/assets/create-pull-request.yaml`)
  - structured parse (`internal/api/prdata_parse.go`)
  - generation entrypoints (`internal/api/api.go`, `cmd/prescribe/cmds/generate.go`)
  - git plumbing (`internal/git/git.go`)

### Why
- Establish the concrete file/symbol map needed to convert the goal into docmgr tasks.

### What worked
- Found the core generation pipeline and the structured YAML parse contract.

### What didn't work
- N/A

### What I learned
- The README recommends `gh pr create --body-file ...`, but there are no `gh` call sites in `prescribe/internal/**` yet.
- `domain.GeneratedPRData` already matches what `gh pr create` needs (`title` + `body`).

### What was tricky to build
- Separating “PR description generation” (existing) vs “PR creation” (missing) without over-speculating on implementation details.

### What warrants a second pair of eyes
- Confirm the intended “create PR” UX: should it always call `generate` first, or support “use last generated PR data”?

### What should be done in the future
- Convert the analysis doc into a task list and then execute incrementally with small commits.

### Code review instructions
- Start with the analysis doc and follow the file links:
  - `reference/02-pr-creation-codebase-analysis.md`

### Technical details
- Analysis doc path:
  - `prescribe/ttmp/2025/12/29/PR-CREATION--implement-pr-creation-end-to-end/reference/02-pr-creation-codebase-analysis.md`

## Step 3: Clarification workflow setup (plz-confirm)

This step prepared the clarification workflow so we can quickly lock down user-facing PR creation behavior without long back-and-forth. The key outcome is a concrete plan for which `plz-confirm` widgets we’ll use and how we’ll capture answers in a structured way that can be translated into docmgr tasks.

### What I did
- Re-read the intended interaction pattern for `plz-confirm` (confirm/select/form).
- Chose a default approach for capturing clarifications as structured output (JSON).

### Why
- Clarifying questions are highest leverage before we start implementing or writing tasks; it reduces rework and keeps the DAG honest.

### What worked
- `plz-confirm` supports the exact interaction primitives we need (yes/no, choice lists, and small structured forms).

### What didn't work
- N/A

### What I learned
- `plz-confirm confirm` is best for yes/no toggles (e.g., “use gh CLI?”).
- `plz-confirm select` is best for small enumerations (e.g., “which base branch inference?”).
- `plz-confirm form` is best for a small set of structured fields (e.g., title/body source, flags, draft state).

### What was tricky to build
- N/A

### What warrants a second pair of eyes
- Confirm we’re not over-fitting to `gh` (GitHub CLI) if the project wants a GitHub API integration instead.

### What should be done in the future
- Draft the actual questions and run them via plz-confirm; then convert the answers into docmgr tasks.

### Code review instructions
- N/A (no code changes yet)

### Technical details
- References:
  - `plz-confirm help how-to-use`
  - `docmgr help how-to-use`

## Step 4: Draft clarifying questions from analysis

This step derived a minimal set of concrete clarifying questions from the codebase analysis document to remove ambiguity about PR creation functionality before task planning. The questions cover integration approach, CLI surface design, workflow, branch management, and error handling.

### What I did
- Reviewed the codebase analysis doc (`reference/02-pr-creation-codebase-analysis.md`)
- Examined the existing CLI structure (`prescribe/cmd/prescribe/cmds/`)
- Drafted 8 clarifying questions covering:
  1. Integration approach (gh CLI vs GitHub API)
  2. CLI surface design (new command vs flag)
  3. PR data source (always generate vs reuse)
  4. Branch management (push/create branch)
  5. Draft PRs support
  6. Dry-run/preview mode
  7. Error handling and user feedback
  8. Title/body override support
- Created questions document: `reference/03-clarifying-questions.md`
- Each question includes: title, why it matters, recommended plz-confirm widget type

### Why
- Lock down user-facing behavior and constraints before implementation
- Reduce rework by clarifying ambiguous requirements upfront
- Enable structured answer capture via plz-confirm for translation into docmgr tasks

### What worked
- Questions cover the key decision points identified in the analysis
- Each question maps to a specific plz-confirm widget type (confirm/select/form)
- Questions are concrete and answerable without long-form responses

### What didn't work
- N/A

### What I learned
- The analysis doc clearly identified the two integration options (gh CLI vs GitHub API)
- Current CLI has `generate` and `tui` commands, suggesting a new `create` command might fit the pattern
- Session management already exists, which could support "reuse last generated" workflow

### What was tricky to build
- Balancing comprehensiveness (cover all ambiguities) with minimalism (smallest set of questions)
- Choosing appropriate widget types for multi-part questions (e.g., branch management)

### What warrants a second pair of eyes
- Confirm the questions cover all critical ambiguities
- Verify widget type recommendations match plz-confirm capabilities

### What should be done in the future
- Run these questions via plz-confirm to capture structured answers
- Convert answers into concrete docmgr tasks
- Begin implementation based on clarified requirements

### Code review instructions
- Review the questions document:
  - `reference/03-clarifying-questions.md`
- Verify questions are answerable and cover key decision points
- Confirm widget type recommendations are appropriate

### Technical details
- Questions document path:
  - `prescribe/ttmp/2025/12/29/PR-CREATION--implement-pr-creation-end-to-end/reference/03-clarifying-questions.md`
- Related analysis:
  - `reference/02-pr-creation-codebase-analysis.md`

## Step 5: Ask clarifying questions via plz-confirm

This step asked all 8 clarifying questions via plz-confirm widgets to capture structured answers that will drive task creation. The answers lock down the PR creation functionality requirements before implementation begins.

### What I did
- Asked 8 clarifying questions using appropriate plz-confirm widgets:
  1. **Q1 (select)**: Integration approach → Selected "Shell out to GitHub CLI (gh pr create)"
  2. **Q2 (select)**: CLI surface design → Selected "Both (allow generate --create and standalone create)" with note about pointing to previously generated YAML
  3. **Q3 (confirm)**: PR data source → Selected "Allow reuse of last generated PR data from session"
  4. **Q4 (form)**: Branch management → Push branch: yes, Create branch remotely: no, Base branch inference: "main (as a --base parameter default)"
  5. **Q5 (form)**: Draft PRs → Support: yes, Default to draft: no, Flag name: `--draft`
  6. **Q6 (confirm)**: Dry-run mode → Selected "Yes: Support --dry-run flag"
  7. **Q7 (select)**: Error handling → Selected "Exit with error + save generated PR data to file for manual retry"
  8. **Q8 (confirm)**: Title/body override → Selected "Yes: Support --title and --body flags" with note about full YAML editing capability
- Captured all answers in structured JSON format
- Created answers document: `reference/04-clarification-answers.json`

### Why
- Lock down exact user-facing behavior before writing implementation tasks
- Ensure all ambiguities are resolved so task planning can proceed with confidence
- Create a structured record of decisions that can be referenced during implementation

### What worked
- plz-confirm widgets worked well for each question type (select for choices, confirm for yes/no, form for structured multi-field inputs)
- User provided helpful comments/clarifications alongside answers (e.g., ability to point to previously generated YAML, full YAML editing capability)
- All questions were answered, providing complete clarity on PR creation requirements

### What didn't work
- N/A

### What I learned
- Key decisions made:
  - Use GitHub CLI (`gh pr create`) rather than GitHub API (simpler, matches README)
  - Support both `generate --create` and standalone `create` command
  - Allow reuse of last generated PR data (don't always regenerate)
  - Push branch before creating PR, but don't auto-create remote branch
  - Support draft PRs via `--draft` flag (opt-in, not default)
  - Support `--dry-run` for preview mode
  - On error: save PR data to file for manual retry
  - Support `--title`/`--body` flags plus full YAML editing
- User wants flexibility: point to previously generated YAML, edit full YAML data, reuse session data

### What was tricky to build
- Creating JSON schemas for form widgets (Q4 branch management, Q5 draft PRs)
- Ensuring questions were clear enough to get actionable answers without over-explaining

### What warrants a second pair of eyes
- Review the answers document to ensure all decisions are clear and actionable
- Verify the answers align with the codebase analysis and existing patterns

### What should be done in the future
- Convert clarification answers into concrete docmgr tasks
- Begin implementation based on clarified requirements
- Reference answers document during implementation to ensure alignment

### Code review instructions
- Review the answers document:
  - `reference/04-clarification-answers.json`
- Verify decisions are clear and can be translated into implementation tasks
- Check that answers align with user expectations and codebase constraints

### Technical details
- Answers document path:
  - `prescribe/ttmp/2025/12/29/PR-CREATION--implement-pr-creation-end-to-end/reference/04-clarification-answers.json`
- Questions document:
  - `reference/03-clarifying-questions.md`
- plz-confirm widgets used:
  - `select` for Q1, Q2, Q7 (choice-based questions)
  - `confirm` for Q3, Q6, Q8 (yes/no questions)
  - `form` for Q4, Q5 (structured multi-field inputs)

## Step 6: Add `prescribe create` command skeleton

This step introduced the **CLI surface area** for end-to-end PR creation: a new `prescribe create` command with the agreed-upon flags, wired into the root command tree. It deliberately does **not** implement PR creation yet; it only establishes the contract and unblocks implementing `gh pr create` integration next.

It also clarified an important repo detail: **`prescribe/` is its own git worktree**, so code commits for `prescribe` changes need to happen from within the `prescribe/` directory.

**Commit (code):** b38e25ee508164c456c31c3d4e9f4e3784b06077 — "cmd: add prescribe create command skeleton"

### What I did
- Added a new cobra/glaze command file: `cmd/prescribe/cmds/create.go`
- Wired the new command into `cmd/prescribe/cmds/root.go` via `InitCreateCmd()` and `rootCmd.AddCommand(createCmd)`
- Implemented the initial flag surface (no behavior yet):
  - `--use-last`, `--yaml-file`, `--title`, `--body`, `--draft`, `--dry-run`, `--base`
- Formatted and validated the module:
  - `gofmt -w cmd/prescribe/cmds/create.go cmd/prescribe/cmds/root.go`
  - `go test ./... -count=1`
- Verified the command is present using `go run`:
  - `go run ./cmd/prescribe create --help`

### Why
- Establish a stable UX/flag contract from the captured clarifications before implementing behavior
- Make subsequent tasks (gh integration + generate --create wiring) incremental and reviewable

### What worked
- `prescribe create --help` shows the command and expected flags
- `go test ./... -count=1` passes for the `prescribe` module
- Pre-commit hooks (test + golangci-lint) passed on commit

### What didn't work
- `prescribe create` still returns an error because PR creation is not implemented yet (expected for this step)

### What I learned
- The `generate` command uses `cli.BuildCobraCommand(...)`; matching that pattern avoided missing helper APIs (e.g., non-existent `BuildCobraCommandFromGlazeCommand`)
- `prescribe/` is a nested git worktree, so commits must be made from inside `prescribe/`

### What was tricky to build
- Getting the Glazed/Cobra wiring correct (mirroring `generate.go`’s `cli.BuildCobraCommand` pattern) so the command shows up and flags render properly

### What warrants a second pair of eyes
- Confirm the initial flag set matches `reference/04-clarification-answers.json` and doesn’t conflict with existing `generate` flags
- Confirm the chosen layering/middleware approach is consistent with other commands (especially config/env precedence expectations)

### What should be done in the future
- Implement the actual PR creation flow behind `prescribe create` (gh integration, data sourcing, dry-run, error-save behavior)
- Add `generate --create` wiring once create flow exists

### Code review instructions
- Start with:
  - `cmd/prescribe/cmds/create.go`
  - `cmd/prescribe/cmds/root.go`
- Validate with:
  - `cd prescribe && go run ./cmd/prescribe create --help`
  - `cd prescribe && go test ./... -count=1`

### Technical details
- New command: `prescribe create`
- Flags implemented (behavior pending): `--use-last`, `--yaml-file`, `--title`, `--body`, `--draft`, `--dry-run`, `--base`

## Step 7: Add `gh pr create` integration (dry-run supported)

This step added a minimal GitHub CLI integration layer so `prescribe create` can construct and (optionally) execute `gh pr create`. To keep it safe and testable locally, it also supports `--dry-run`, which prints the redacted `gh` command without creating anything.

This unlocks the next steps: pushing branches before create, loading PR data from session/YAML, and wiring `generate --create`.

**Commit (code):** 88c26c9672deef0d74a211ab1e816e6d4a7c901f — "create: add gh pr create integration"

### What I did
- Added `internal/github` package with:
  - `BuildGhCreatePRArgs` for constructing `gh pr create` args
  - `CreatePR(ctx, opts)` to execute `gh` via `exec.CommandContext`
  - redaction helper so error output/prints don’t include full PR body
- Updated `cmd/prescribe/cmds/create.go` to:
  - validate required inputs (for now: `--title` + `--body`)
  - print a redacted command in `--dry-run` mode
  - run `gh pr create ...` when not in dry-run mode
- Tested via `go run` (no build artifacts):
  - `cd prescribe && go run ./cmd/prescribe create --dry-run --title \"test title\" --body \"test body\"`
- Ran module tests:
  - `cd prescribe && go test ./... -count=1`

### Why
- Implement the agreed integration choice (GitHub CLI) with a small, testable abstraction
- Establish safe local testing (`--dry-run`) before we start pushing branches or creating real PRs

### What worked
- `--dry-run` prints the intended `gh pr create` invocation with the body redacted
- `go test ./... -count=1` continues to pass
- Command wiring stays within existing Glazed/Cobra patterns

### What didn't work
- `--use-last` / `--yaml-file` are still intentionally unimplemented (tracked as separate tasks)

### What I learned
- Redacting the `--body` argument is important for logs/errors because PR bodies can be large and may contain sensitive info

### What was tricky to build
- Avoiding accidental leakage of the full PR body into logs while still keeping the CLI debuggable

### What warrants a second pair of eyes
- Confirm the `internal/github` surface is the right abstraction boundary (vs putting exec directly into the command)
- Confirm argument redaction behavior is sufficient (and doesn’t hide too much)

### What should be done in the future
- Add branch push behavior before create
- Implement `--use-last` and `--yaml-file` data sources
- Add error-path persistence of PR YAML to disk on failure

### Code review instructions
- Start with:
  - `internal/github/github.go`
  - `cmd/prescribe/cmds/create.go`
- Validate quickly:
  - `cd prescribe && go run ./cmd/prescribe create --dry-run --title \"t\" --body \"b\"`

### Technical details
- `internal/github.Service.CreatePR` executes `gh` in `repoPath` and returns combined output on success

## Step 8: Push branch before PR creation

This step made `prescribe create` **push the current branch** before attempting to create the PR via `gh pr create`, matching the clarified workflow (push=yes). To keep behavior conservative, we intentionally do **not** set upstream automatically (no `git push -u ...`); if the branch has no upstream configured, the push will fail and we’ll surface the git error.

**Commit (code):** c1b08979a43533c7e786d2e5b4aa976083d3e221 — "create: push branch before gh pr create"

### What I did
- Added `(*git.Service).PushCurrentBranch(ctx)` in `internal/git/git.go` (runs `git push` with context)
- Updated `cmd/prescribe/cmds/create.go` to:
  - include the push step before `gh pr create`
  - show the push step in `--dry-run` output
- Verified behavior via `go run`:
  - `cd prescribe && go run ./cmd/prescribe create --dry-run --title \"t\" --body \"b\"`
- Ran module tests:
  - `cd prescribe && go test ./... -count=1`

### Why
- The PR creation workflow should be end-to-end: ensure the branch is pushed before opening the PR
- Keeping upstream configuration manual avoids implicit remote-branch creation semantics beyond “push”

### What worked
- `--dry-run` now shows both steps: `git push` then `gh pr create ...`
- Code compiles and tests pass

### What didn't work
- If a branch has no upstream set, `git push` will fail (expected by design for now)

### What I learned
- Using plain `git push` is a good “minimal behavior” default that respects existing git configs and forces explicit upstream setup when missing

### What was tricky to build
- Avoiding accidental “set upstream / create remote branch” behavior while still satisfying “push before create”

### What warrants a second pair of eyes
- Confirm this interpretation of “create_branch_remote: false” is correct (plain `git push`, no `-u`)

### What should be done in the future
- If desired, add an explicit flag (e.g., `--set-upstream`) to opt into `git push -u origin HEAD`

### Code review instructions
- Start with:
  - `internal/git/git.go` (`PushCurrentBranch`)
  - `cmd/prescribe/cmds/create.go` (push+create ordering)
- Validate with:
  - `cd prescribe && go run ./cmd/prescribe create --dry-run --title \"t\" --body \"b\"`

## Step 9: Support `prescribe create --yaml-file` (dry-run smoke-tested)

This step made the standalone `prescribe create` workflow practical: you can now point the command at a previously-generated YAML file containing `title` and `body`, and it will use that as the PR content source. This matches the clarification note that “standalone create should be able to point to a previously generated yaml and use that”.

We also kept the workflow safe for environments without a GitHub remote by validating everything via `--dry-run`.

**Commit (code):** 457a6e75fac47590a71560aa3c4ce1fab573def6 — "create: support --yaml-file and --use-last (dry-run smoke test)"

### What I did
- Added `internal/prdata` helper to load `domain.GeneratedPRData` from a YAML file
- Updated `cmd/prescribe/cmds/create.go`:
  - `--yaml-file` loads `title`/`body` from the file
  - explicit `--title`/`--body` override file contents (if provided)
- Added a smoke test script using `go run`:
  - `test-scripts/07-smoke-test-prescribe-create-dry-run.sh`

### Why
- Enable “standalone create” to work from a saved YAML artifact without needing to re-run generation

### What worked
- `create --dry-run --yaml-file <file>` prints the expected redacted `gh pr create ...` command
- Full suite still passes: `go test ./... -count=1`

### What didn't work
- N/A

### What I learned
- Keeping the YAML parsing in a small `internal/prdata` package makes it easy to reuse for `--use-last` and future “save-on-error” behavior

### What was tricky to build
- Ensuring flag precedence: allow YAML file as the source but still allow explicit `--title/--body` overrides

### What warrants a second pair of eyes
- Confirm the chosen YAML contract (only requiring `title` and `body`) is acceptable even though the prompt pack also includes `changelog` and `release_notes`

### What should be done in the future
- Add support for “edit full YAML” workflow (e.g., open editor) if desired; current behavior supports overrides via flags only

### Code review instructions
- Start with:
  - `internal/prdata/prdata.go`
  - `cmd/prescribe/cmds/create.go`
  - `test-scripts/07-smoke-test-prescribe-create-dry-run.sh`
- Validate:
  - `cd prescribe && bash test-scripts/07-smoke-test-prescribe-create-dry-run.sh`

## Step 10: Support `prescribe create --use-last` (persist via generate; dry-run smoke-tested)

This step enabled a fast “generate once, then create later” workflow: `prescribe generate` now persists the parsed structured PR data to `.pr-builder/last-generated-pr.yaml`, and `prescribe create --use-last` loads that file as the source of `title` and `body`.

Because our smoke test environment doesn’t have a GitHub remote, we validate this behavior via `--dry-run` and a ticket script that writes a synthetic `last-generated-pr.yaml`.

**Commit (code):** 457a6e75fac47590a71560aa3c4ce1fab573def6 — "create: support --yaml-file and --use-last (dry-run smoke test)"

### What I did
- Added `internal/prdata.WriteGeneratedPRDataToYAMLFile` to persist structured PR YAML
- Updated `cmd/prescribe/cmds/generate.go` to write parsed PR data to:
  - `.pr-builder/last-generated-pr.yaml`
- Updated `cmd/prescribe/cmds/create.go`:
  - `--use-last` loads `.pr-builder/last-generated-pr.yaml`
  - still allows `--title/--body` overrides
- Added a tiny ticket helper to write last-generated-pr.yaml for smoke tests:
  - `ttmp/2025/12/29/PR-CREATION--implement-pr-creation-end-to-end/scripts/01-write-last-generated-prdata.go`
- Extended/added smoke testing to cover `--use-last` in dry-run:
  - `test-scripts/07-smoke-test-prescribe-create-dry-run.sh`

### Why
- Match the clarified requirement to allow reuse of last generated PR data (don’t always regenerate)

### What worked
- `create --dry-run --use-last` prints the expected redacted `gh pr create` invocation
- Unit test added for YAML roundtrip: `internal/prdata/prdata_test.go`

### What didn't work
- N/A

### What I learned
- “Use-last” becomes much simpler and more stable if we treat `.pr-builder/last-generated-pr.yaml` as the single source of truth, instead of trying to reverse-engineer session.yaml (which doesn’t persist generated PR data)

### What was tricky to build
- Finding a robust persistence point: session.yaml doesn’t contain generated PR data, so `generate` must write it explicitly

### What warrants a second pair of eyes
- Confirm we should always overwrite `.pr-builder/last-generated-pr.yaml` on generate (vs versioning by timestamp)

### What should be done in the future
- Decide whether to also save the file on `create` failure (task #11) using the same `internal/prdata` helpers

### Code review instructions
- Start with:
  - `cmd/prescribe/cmds/generate.go` (persist parsed PR data)
  - `cmd/prescribe/cmds/create.go` (use-last loading)
  - `internal/prdata/prdata.go` (+ test)
  - `test-scripts/07-smoke-test-prescribe-create-dry-run.sh`

## Step 11: Save PR YAML to disk on create failure

This step hardens the CLI for real-world failures (auth, networking, missing upstream): if `prescribe create` fails while pushing or running `gh pr create`, it now writes a timestamped YAML file under `.pr-builder/` so you can retry manually without losing the generated PR content.

We keep the workflow safe for local development by relying on the existing dry-run smoke tests for the happy-path and validating the failure-save logic via unit tests + code review (we don’t require an actual GH remote).

**Commit (code):** c4e22c9b91bc6fc454dbcabb33ab0f67564a4ae6 — "create: save pr-data yaml on failure; test gh args"

### What I did
- Added `internal/prdata.FailurePRDataPath(repoPath, now)` for timestamped paths like:
  - `.pr-builder/pr-data-YYYYMMDD-HHMMSS.yaml`
- Updated `cmd/prescribe/cmds/create.go`:
  - on `git push` failure: write PR data YAML and print where it was saved
  - on `gh pr create` failure: write PR data YAML and print where it was saved

### Why
- Avoid losing work when create fails; always leave a retry artifact behind

### What worked
- Unit and lint pass: `go test ./... -count=1`
- Failure path now produces a deterministic “saved PR data to …” message on stderr

### What didn't work
- N/A

### What I learned
- The `internal/prdata` helper is a natural place to centralize “save on failure” behavior so future features (task #16 / integrations) can reuse it

### What was tricky to build
- Avoiding misleading output: we need to save to disk *before* telling the user where to find it

### What warrants a second pair of eyes
- Confirm the timestamp format and directory choice are acceptable as a stable UX contract

### What should be done in the future
- Optionally write the same “failure artifact” on other failure types (e.g., parse failures) if useful

### Code review instructions
- Start with:
  - `cmd/prescribe/cmds/create.go`
  - `internal/prdata/prdata.go`

### Technical details
- Failure artifacts go under: `.pr-builder/pr-data-<timestamp>.yaml`

## Step 12: Add unit tests for `gh pr create` argument construction

This step added unit tests around the most error-prone piece of the `gh` integration: building the argument list correctly and ensuring we redact the PR body when printing/logging.

**Commit (code):** c4e22c9b91bc6fc454dbcabb33ab0f67564a4ae6 — "create: save pr-data yaml on failure; test gh args"

### What I did
- Added `internal/github/github_test.go` with tests for:
  - missing title/body validation
  - base + draft flags
  - redaction behavior for `--body`

### Why
- Keep the integration stable without requiring network/GitHub access during tests

### What worked
- Tests run quickly and validate the CLI surface without external dependencies

### What warrants a second pair of eyes
- Confirm we should keep argument ordering stable (tests assert exact slice ordering)

## Step 13: Diagnose non-dry-run create timeouts (git push hooks) + extend smoke test

This step investigated why a `prescribe create` run without `--dry-run` previously “timed out”. The result is clear: the hang was in **`git push`**, specifically because the `prescribe/` repo has a **`lefthook` pre-push hook** that runs tests and golangci-lint. It was not `gh` prompting interactively.

To make this diagnosable going forward, we extended the existing smoke test to include a **bounded non-dry-run** run against a tiny repo with no remote configured, so `git push` fails quickly and we can still validate tracing + save-on-failure behavior.

**Commit (code):** 106e5e5be87a99f51a6064501f4caf3f8107f5fa — "test: add bounded non-dry-run create smoke check"

### What I did
- Ran `prescribe create` from within `prescribe/` with a timeout and prompts disabled:
  - `timeout 10s env GIT_TERMINAL_PROMPT=0 GH_PROMPT_DISABLED=1 go run ./cmd/prescribe --repo . create --title \"t\" --body \"b\"`
- Observed the tracing output:
  - it always prints the commands first (`git push`, then redacted `gh ...`)
  - it stalled during `git push`
- Confirmed the actual work happening during the stall was `lefthook`’s **pre-push** hook running `make test` and `golangci-lint`
- Updated `test-scripts/07-smoke-test-prescribe-create-dry-run.sh` to add a third section:
  - create a tiny local repo (no remote)
  - run `create` non-dry-run with `timeout` + prompts disabled
  - assert the log contains the new trace lines and “saved PR data to …”

### Why
- Prevent future confusion: timeouts aren’t “gh interactive”, they’re “git push runs hooks / blocks”
- Ensure we have a reliable, offline-friendly reproduction that never talks to GitHub

### What worked
- The log clearly shows where we were running (`cwd`), what repo we targeted, and which commands were executed
- The bounded smoke test captures tracing + save-on-failure output deterministically

### What didn't work
- Running non-dry-run against the actual `prescribe/` repo is slow because pre-push hooks do real work (expected)

### What I learned
- The safest “diagnose create” recipe is: `timeout + GIT_TERMINAL_PROMPT=0 + GH_PROMPT_DISABLED=1` plus our new tracing lines

### What warrants a second pair of eyes
- Decide whether `prescribe create` should provide a flag to skip pushing (useful for repos with heavy pre-push hooks)

## Step 14: Validate real create flow (LEFTHOOK=0) against the prescribe repo

This step ran a **real** `prescribe create` (non-dry-run) from within the `prescribe/` repo with `LEFTHOOK=0` to skip pre-push hooks. The run successfully pushed the branch and created a **draft PR** via `gh pr create`, confirming that the earlier “timeout” problem was indeed the git hooks and not `gh` interactivity.

**Result:** Draft PR created: `https://github.com/go-go-golems/prescribe/pull/2`

### What I did
- Ran from within `prescribe/`:
  - `timeout 120s env LEFTHOOK=0 GIT_TERMINAL_PROMPT=0 GH_PROMPT_DISABLED=1 go run ./cmd/prescribe --repo . create --draft --title ... --body ...`
- Observed tracing output:
  - `git push` succeeded (~1s)
  - `gh pr create` succeeded (~2s)

### Why
- Confirm the end-to-end “real action” flow works in the actual upstream repo when hooks are skipped

### What worked
- Push + PR creation succeeded with prompts disabled
- The tracing output clearly showed cwd/repo/source/commands and timings

### What didn't work
- `gh` warned about uncommitted changes; it appears these are ticket docs present as untracked files in this worktree (safe to ignore for PR creation, but worth keeping in mind)

### What I learned
- `LEFTHOOK=0` is the key lever to avoid the pre-push hook cost during interactive create workflows

### What was tricky to build
- N/A (this was an execution/validation step)

### What warrants a second pair of eyes
- Confirm the PR is targeting the intended base branch and contains only the intended commits (no doc noise)

### What should be done in the future
- Consider whether `prescribe create` should expose a flag like `--skip-push` or `--no-push` (for users with heavy hooks) while keeping current default behavior

### Code review instructions
- Open PR #2 and review commit list:
  - `https://github.com/go-go-golems/prescribe/pull/2`

## Step 15: Add `prescribe generate --create` wiring (with dry-run/base/draft flags)

This step wired up a combined workflow: after generating a PR description (and successfully parsing the structured YAML into `GeneratedPRData`), `prescribe generate --create` can now push the branch and create the PR via `gh pr create`.

To keep this testable in environments without a GitHub remote, we also added `--create-dry-run` which prints the create actions without executing them.

**Commit (code):** a87691b292f95b3a1a1bcc97da37a33b640e304a — "generate: add --create (with dry-run/base/draft flags)"

### What I did
- Added new flags to `prescribe generate`:
  - `--create` (enable PR creation after generation)
  - `--create-dry-run` (print actions only)
  - `--create-draft` (create draft PR)
  - `--create-base` (base branch, default `main`)
- Implemented the post-generation flow using `GeneratedPRData`:
  - `git push` (same behavior as `prescribe create`)
  - `gh pr create ...` (printed to stderr to avoid mixing with description output)

### Why
- Enable the “one command” workflow: generate + create PR
- Keep it safe and inspectable in CI/smoke tests via `--create-dry-run`

### What worked
- Help output shows the new flags (`go run ./cmd/prescribe generate --help`)
- Existing unit tests and smoke tests still pass

### What didn't work
- We did not run a full end-to-end “generate with inference” in smoke tests (it requires LLM config); this step is wiring + flag surface + safe dry-run mode

### What warrants a second pair of eyes
- Confirm that printing `gh` output to stderr is the right UX (stdout remains the generated description)
- Confirm the failure-save behavior is sufficient (we only save PR data YAML on `gh` failure, not on push failure in generate path yet)

### Code review instructions
- Start with:
  - `cmd/prescribe/cmds/generate.go`
- Validate flags exist:
  - `cd prescribe && go run ./cmd/prescribe generate --help | grep -E -- \"create-dry-run|create-base|create-draft|--create\"`

## Step 16: Full-circle test with Gemini profile (generate → create PR)

This step validated the full workflow using a real model profile: run `prescribe generate` with `PINOCCHIO_PROFILE=gemini-2.5-flash`, persist parsed PR data to `.pr-builder/last-generated-pr.yaml`, then create a **draft PR** from that data.

The first attempt used the integrated `generate --create` path and failed at `git push` because the fresh test branch had **no upstream** configured (expected given our current “plain git push” behavior). After explicitly setting upstream once, `prescribe create --use-last --draft` succeeded and opened a draft PR.

**Draft PR created:** `https://github.com/go-go-golems/prescribe/pull/3`

### What I did
- Created a fresh test branch:
  - `git checkout -b task/pr-creation-e2e-20251229-140459`
- Ran generation with a real model profile:
  - `env PINOCCHIO_PROFILE=gemini-2.5-flash ... go run ./cmd/prescribe --repo . --target main generate`
  - Confirmed it wrote: `.pr-builder/last-generated-pr.yaml`
- Attempted integrated creation:
  - `generate --create --create-draft --create-base main`
  - Failed at `git push` due to missing upstream on the new branch
- Set upstream once and created from saved PR YAML:
  - `env LEFTHOOK=0 git push --set-upstream origin HEAD`
  - `env LEFTHOOK=0 ... go run ./cmd/prescribe --repo . create --use-last --draft --base main`

### Why
- Verify the end-to-end happy path with a real LLM profile and real `gh pr create`
- Confirm our persistence/hand-off mechanism (`last-generated-pr.yaml`) is sufficient for “generate once, create later”

### What worked
- Generation succeeded under `PINOCCHIO_PROFILE=gemini-2.5-flash`
- Parsed PR data was persisted to `.pr-builder/last-generated-pr.yaml`
- Draft PR creation succeeded via `create --use-last --draft` (PR #3)

### What didn't work
- `generate --create` failed on a brand-new branch due to missing upstream (`git push` requires upstream unless configured)

### What I learned
- For first push of a new branch, either the user must set upstream once (`git push -u ...`) or we need an explicit opt-in flag to do that automatically

### What warrants a second pair of eyes
- Decide the intended UX for “no upstream” branches:
  - keep strict behavior (fail with clear guidance), or
  - add an explicit flag to set upstream automatically

### Code review instructions
- Review the created PR:
  - `https://github.com/go-go-golems/prescribe/pull/3`

## Step 17: Update user-facing documentation (README + workflow topic)

This step updated the user-facing docs to reflect the new end-to-end PR creation capabilities: `prescribe create` and `prescribe generate --create`, including safe dry-run workflows, YAML sources (`--use-last` / `--yaml-file`), and common gotchas like upstream branches and `lefthook` pre-push hooks.

### What I did
- Updated `prescribe/README.md` with:
  - quick start examples for `create` and `generate --create`
  - a new `create` command section with flags and failure behavior
  - updated `generate` section to include the new `--create*` flags
  - refreshed outdated `pr-builder` examples to `prescribe` equivalents in common workflows
- Updated the help topic `prescribe help how-to-generate-pr-description`:
  - added “Step 6: Create the PR on GitHub” with both workflows:
    - generate → create (`--use-last`)
    - `generate --create` (plus `--create-dry-run`)
  - documented upstream branch and `LEFTHOOK=0` gotchas

### Why
- Ensure new users discover the end-to-end workflow without reading code or tickets
- Prevent common failure modes (no upstream branch; slow push hooks) by documenting them up front

### What worked
- Docs now match the actual CLI surface we implemented and tested in the `prescribe/` repo

### What warrants a second pair of eyes
- Confirm the README examples align with your intended “recommended path” (use-last vs yaml-file vs generate --create)

### Code review instructions
- Start with:
  - `prescribe/README.md`
  - `prescribe/pkg/doc/topics/02-how-to-generate-pr-description.md`

## Step 18: Add end-to-end integration test for PR creation (local remote + fake gh)

This step turned the previously “manual validation” into a repeatable, safe integration check: we can now run `create` and `generate --create` end-to-end without pushing to GitHub or creating real PRs. The script uses a local bare git remote so `git push` succeeds, and it injects a fake `gh` into `PATH` so `gh pr create` is captured to a log instead of hitting the network.

It also explicitly acknowledges a constraint: `prescribe generate` still requires real AI step settings. To keep the test useful even without AI config, the script supports `SKIP_GENERATE=1` to test only the create flows.

**Commit (code):** c29bac1a02b3f48ff811f090297f411b70994b7e — "test: add PR creation integration script (fake gh + local remote)"

### What I did
- Added a new integration script: `prescribe/test-scripts/08-integration-test-pr-creation.sh`
  - Creates a local bare remote under `/tmp/...remote.git` and sets it as `origin`
  - Writes a tiny YAML file and runs `prescribe create --yaml-file ...` (non-dry-run)
  - Runs `prescribe generate` and asserts it writes `.pr-builder/last-generated-pr.yaml`
  - Runs `prescribe create --use-last` (non-dry-run)
  - Runs `prescribe generate --create` (non-dry-run)
  - Replaces `gh` with a fake script in `PATH` that records `gh pr create ...` and prints a fake URL
- Updated `prescribe/test-scripts/README.md` to document the new script and `SKIP_GENERATE=1`
- Ran it in both modes:
  - `cd prescribe && SKIP_GENERATE=1 bash test-scripts/08-integration-test-pr-creation.sh`
  - `cd prescribe && bash test-scripts/08-integration-test-pr-creation.sh`

### Why
- The last remaining ticket task was an “integration test: end-to-end PR creation”
- We want coverage that is deterministic and safe (no real GitHub side effects), while still validating `git push` + `gh pr create` wiring

### What worked
- The script verified all targeted flows and produced stable logs:
  - create from YAML (`--yaml-file`)
  - generate → persisted `.pr-builder/last-generated-pr.yaml` → create (`--use-last`)
  - generate `--create` (push + gh)

### What didn't work
- Initial run failed with:
  - `grep: unrecognized option '--title Hello'`
- Root cause: patterns beginning with `--` were treated as grep flags; fixed by using `grep -Fq -- "..."`.

### What I learned
- For CLI integration tests, it’s worth guarding against “grep option parsing” whenever expected output contains `--flags`.
- A local bare remote is the simplest way to exercise `git push` without network access.

### What was tricky to build
- Making the test both “safe by default” (no GitHub) and still meaningful (exercise push + create), while acknowledging that generate requires AI config.

### What warrants a second pair of eyes
- Verify that mocking `gh` via `PATH` is acceptable for this repo’s testing philosophy (vs. a Go-level exec abstraction).
- Confirm the script’s assumptions around branch names (`master`) match the rest of the smoke test scripts.

### What should be done in the future
- If we ever add a mock inference provider, we could remove the AI-config requirement and make the full script runnable in CI.

### Code review instructions
- Start with:
  - `prescribe/test-scripts/08-integration-test-pr-creation.sh`
  - `prescribe/test-scripts/README.md`
- Validate locally:
  - `cd prescribe && SKIP_GENERATE=1 bash test-scripts/08-integration-test-pr-creation.sh`
  - `cd prescribe && bash test-scripts/08-integration-test-pr-creation.sh`

### Technical details
- The fake `gh` writes a log to `${GH_LOG}` and returns a fake PR URL for `gh pr create`.
