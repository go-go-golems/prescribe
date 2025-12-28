---
Title: Diary
Ticket: 008-GENERATE
Status: active
Topics:
    - prescribe
    - geppetto
    - inference
    - pr-generation
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-27T19:54:31-05:00
WhatFor: ""
WhenToUse: ""
---

# Diary

## Goal

Document the step-by-step research and analysis process for implementing AI-powered PR description generation in prescribe using geppetto inference engines. This diary captures what was explored, what was learned, and what decisions were made during the analysis phase.

## Step 1: Initial Exploration and Ticket Creation

**Commit (code):** N/A — Research phase

### What I did
- Created ticket `008-GENERATE` using docmgr
- Explored the catter implementation in `pinocchio/cmd/pinocchio/cmds/catter/`
- Read catter's README, TODO, and main command files
- Examined fileprocessor.go and stats.go to understand the architecture

### Why
- Need to understand how catter exports file contents for LLM contexts
- Catter provides a similar pattern to what prescribe needs for exporting diff data
- Understanding catter's filter and token counting will inform prescribe's export design

### What worked
- Found comprehensive catter implementation with clear separation of concerns
- Identified key components: FileProcessor, Stats, filefilter integration
- Discovered support for multiple output formats (text, zip, tar.gz) and token counting

### What I learned
- Catter uses `tiktoken-go` with `cl100k_base` encoding for token counting
- FileProcessor uses options pattern for configuration
- Stats are pre-computed using `filewalker.Walker` before processing
- Filter system integrates with `clay/pkg/filefilter` package

### What was tricky to build
- Understanding the relationship between FileProcessor, Stats, and filewalker
- Distinguishing between text output, archive output, and glazed structured output modes

### What warrants a second pair of eyes
- The token counting approach (cl100k_base) matches OpenAI's tokenization — verify this is appropriate for all providers
- FileProcessor's limit checking (max-tokens, max-total-size) happens during processing — ensure this matches prescribe's needs

### What should be done in the future
- Verify token counting encoding compatibility with different LLM providers (Claude uses different tokenization)
- Consider whether prescribe needs archive export functionality or just Turn block generation

### Code review instructions
- Review `pinocchio/cmd/pinocchio/cmds/catter/pkg/fileprocessor.go` lines 144-238 (ProcessPaths and processFileContent)
- Review `pinocchio/cmd/pinocchio/cmds/catter/pkg/stats.go` lines 94-137 (ComputeStats)

### Technical details
- FileProcessor tracks: TotalSize, TotalTokens, FileCount, TokenCounts map
- Uses `tiktoken.GetEncoding("cl100k_base")` for token counting
- Supports delimiter types: default, xml, markdown, simple, begin-end
- Archive formats: zip, tar.gz with optional prefix

## Step 2: Exploring Prescribe's Current Export and API Structure

**Commit (code):** N/A — Research phase

### What I did
- Read `prescribe/internal/tui/export/export.go` to understand current export format
- Examined `prescribe/internal/api/api.go` to see mock API service implementation
- Reviewed `prescribe/internal/domain/domain.go` for PRData structure
- Analyzed `prescribe/internal/controller/controller.go` for how GenerateDescription is called

### Why
- Need to understand prescribe's current architecture before integrating geppetto
- Export format shows how data is currently structured for LLM consumption
- Mock API service shows the interface that needs to be replaced

### What worked
- Found `BuildGenerationContextText()` which formats PR data into markdown
- Identified `GenerateDescriptionRequest` structure that needs to be converted to Turn
- Discovered `PRData` contains all necessary information (files, context, prompts)

### What I learned
- Prescribe already has a well-structured export format (`BuildGenerationContextText`)
- Current mock API service returns structured response with Description, TokensUsed, Model
- `FileChange` struct supports both diff and full file content (FullBefore, FullAfter)
- Prompt is stored as combined system+user string in `PRData.CurrentPrompt`

### What was tricky to build
- Understanding the relationship between `domain.PRData`, `api.GenerateDescriptionRequest`, and `export.BuildGenerationContextText`
- Noting that prompt is combined system+user but geppetto needs them separated

### What warrants a second pair of eyes
- The combined prompt format (system+user) needs to be split for Turn blocks — verify splitting strategy
- `BuildGenerationContextText` uses markdown formatting — ensure this aligns with geppetto's expectations

### What should be done in the future
- Create prompt splitting utility to separate system/user prompts
- Consider whether to maintain markdown format in Turn blocks or use structured blocks
- Verify token counting matches between prescribe's current approach and geppetto's

### Code review instructions
- Review `prescribe/internal/tui/export/export.go` lines 14-80 (BuildGenerationContextText)
- Review `prescribe/internal/api/api.go` lines 38-127 (GenerateDescription mock implementation)
- Review `prescribe/internal/domain/domain.go` lines 104-123 (PRData structure)

### Technical details
- `GenerateDescriptionRequest` contains: SourceBranch, TargetBranch, Files, AdditionalContext, Prompt
- `FileChange` has: Path, Included, Additions, Deletions, Tokens, Type (diff/full), Version (before/after/both), Diff, FullBefore, FullAfter
- Export format uses markdown with code blocks for diffs and file content

## Step 3: Analyzing Pinocchio's Create Pull Request Pattern

**Commit (code):** N/A — Research phase

### What I did
- Read `pinocchio/cmd/pinocchio/prompts/code/create-pull-request.yaml` template
- Examined `prescribe/internal/prompts/assets/create-pull-request.yaml` (embedded version)
- Searched for how pinocchio executes prompts with geppetto
- Found `pinocchio/pkg/cmds/cmd.go` showing `runEngineAndCollectMessages` pattern

### Why
- Pinocchio's create-pull-request template is the reference implementation
- Need to understand how pinocchio integrates prompts with geppetto engines
- Template structure shows what variables are available and how they're used

### What worked
- Found comprehensive prompt template with system prompt and structured user prompt
- Identified template variables: commits, diff, code, context, description, issue
- Discovered output format is YAML with title, body, changelog, release_notes

### What I learned
- Pinocchio uses Go template syntax (`{{ .variable }}`)
- Templates are rendered before building Turn blocks
- Pinocchio's `buildInitialTurn()` converts rendered template to Turn
- Output format is structured YAML, not just markdown text

### What was tricky to build
- Understanding the relationship between YAML template, template rendering, and Turn building
- Noting that pinocchio loads prompts from files while prescribe embeds them

### What warrants a second pair of eyes
- The YAML output format (title, body, changelog, release_notes) — should prescribe adopt this or use simpler format?
- Template rendering happens before Turn building — verify this is the right approach for prescribe

### What should be done in the future
- Consider whether prescribe should support YAML output format or simpler text
- Evaluate template rendering approach — use Go templates or simpler string substitution?
- Determine how to map prescribe's PRData to template variables

### Code review instructions
- Review `pinocchio/cmd/pinocchio/prompts/code/create-pull-request.yaml` lines 57-153 (full template)
- Review `pinocchio/pkg/cmds/cmd.go` lines 374-395 (runEngineAndCollectMessages)

### Technical details
- Template has system-prompt and prompt sections
- Uses conditional rendering: `{{ if .diff }}...{{ end }}`
- Output format specified in prompt: YAML with title, body, changelog, release_notes
- Pinocchio renders template with variables, then builds Turn from rendered text

## Step 4: Deep Dive into Geppetto Inference Engine Architecture

**Commit (code):** N/A — Research phase

### What I did
- Read `geppetto/pkg/doc/topics/06-inference-engines.md` comprehensive guide
- Examined `geppetto/pkg/inference/engine/engine.go` for Engine interface
- Reviewed `geppetto/pkg/turns/types.go` for Turn and Block structures
- Studied `geppetto/pkg/turns/builders.go` for Turn building patterns
- Analyzed `geppetto/pkg/conversation/builder/builder.go` (legacy approach)
- Found example implementations in `geppetto/cmd/examples/simple-inference/main.go`

### Why
- Need to understand geppetto's Turn-based architecture to integrate with prescribe
- Engine interface shows the API prescribe needs to call
- Turn structure shows how to format data for inference
- Examples show practical usage patterns

### What worked
- Found clear Engine interface: `RunInference(ctx, *Turn) (*Turn, error)`
- Identified Turn structure with Blocks array and metadata
- Discovered TurnBuilder pattern for constructing Turns
- Found factory pattern for creating engines from configuration layers

### What I learned
- Engines operate on Turn (ordered Blocks + metadata), not conversation.Manager
- Recommended approach is to use Turns directly, not conversation.Manager
- Blocks have Kind (system_text, user_text, assistant_text, tool_call, etc.)
- Factory creates engines from parsed layers (API keys, models, etc.)
- Engines can emit streaming events via sinks

### What was tricky to build
- Understanding the relationship between Turns, Blocks, and conversation.Manager
- Distinguishing between recommended (Turns) and legacy (conversation.Manager) approaches
- Understanding how tools are attached (via context.Context registry, not Turn.Data)

### What warrants a second pair of eyes
- The recommendation to use Turns directly vs conversation.Manager — verify this is the right choice for prescribe
- Tool attachment via context.Context — ensure this pattern works for prescribe's use case

### What should be done in the future
- Implement Turn building from PRData
- Create engine factory integration with prescribe's configuration
- Add streaming support (optional, via event sinks)
- Handle tool calling if needed (though PR generation likely doesn't need tools)

### Code review instructions
- Review `geppetto/pkg/inference/engine/engine.go` lines 9-16 (Engine interface)
- Review `geppetto/pkg/turns/types.go` lines 80-100 (Turn and Block structures)
- Review `geppetto/pkg/doc/topics/06-inference-engines.md` lines 149-184 (simple inference example)

### Technical details
- Turn contains: ID, RunID, Blocks array, Metadata map, Data map
- Block contains: ID, TurnID, Kind, Role, Payload map, Metadata map
- Block kinds: BlockKindSystemText, BlockKindUserText, BlockKindAssistantText, BlockKindToolCall, etc.
- Factory: `factory.NewEngineFromParsedLayers(parsedLayers, options...)`
- Engine options: `engine.WithSink(sink)` for streaming

## Step 5: Creating Comprehensive Analysis Document

**Commit (code):** N/A — Documentation phase

### What I did
- Created analysis document with three main sections:
  1. Catter implementation analysis
  2. Create pull request pattern analysis
  3. Geppetto inference engine architecture
- Included pseudocode for key integration points
- Documented implementation plan with phases
- Listed key design decisions and open questions

### Why
- Need comprehensive reference document for implementation
- Pseudocode helps clarify integration approach
- Implementation plan provides roadmap for development
- Design decisions document rationale for choices

### What worked
- Structured analysis by component (catter, pinocchio, geppetto)
- Included file paths and function names for reference
- Provided pseudocode showing integration patterns
- Documented open questions for future resolution

### What I learned
- Catter's architecture is well-suited for adaptation to prescribe's export needs
- Pinocchio's prompt template structure can be reused with minor adaptations
- Geppetto's Turn-based architecture is simpler than expected
- Key challenge is prompt splitting (combined system+user → separate blocks)

### What was tricky to build
- Balancing detail level — too much vs too little
- Organizing information logically across three different systems
- Writing pseudocode that's clear but not overly prescriptive

### What warrants a second pair of eyes
- The pseudocode implementations — verify they align with actual geppetto APIs
- The implementation plan phases — ensure they're in the right order and complete
- The design decisions — validate rationale and consider alternatives

### What should be done in the future
- Resolve open questions (prompt splitting, template variables, error handling, etc.)
- Begin Phase 1 implementation (export infrastructure)
- Create unit tests for Turn building from PRData
- Integrate geppetto dependencies and test basic inference

### Code review instructions
- Review analysis document structure and completeness
- Verify pseudocode matches geppetto API patterns
- Check implementation plan for feasibility and completeness

### Technical details
- Analysis document (moved to inference ticket): `prescribe/ttmp/2025/12/28/GENERATE-LLM-INFERENCE--llm-inference-streaming-and-parsing/analysis/01-analysis-export-prescribe-diff-data-and-generate-pr-descriptions-with-geppetto-inference.md`
- Document includes: architecture analysis, pseudocode, implementation plan, design decisions, open questions
- Next steps: resolve open questions, begin Phase 1 implementation

## Step 6: Resume work — validate export milestone state (TUI/CLI parity)

**Commit (code):** N/A — Investigation

### What I did
- Read the ticket `tasks.md` and cross-checked the current code paths for:
  - CLI: `prescribe generate --export-context --separator ...` implementation
  - TUI: CopyContext hotkey path (clipboard export)
  - Exporter implementations (canonical vs duplicated)
- Verified docs/playbook references for `--export-context`, `--separator`, and `--output-file`.

### Why
- The “Next (handoff-ready)” tasks are mostly about correctness and avoiding drift:
  TUI/CLI should share the same exporter, and we need tests that lock down the flag behavior.

### What I learned
- The shared exporter already exists in `internal/export/context.go` and supports multiple separators.
- The TUI CopyContext path already uses the shared exporter with `SeparatorMarkdown` (clipboard behavior preserved).
- There is still a duplicate markdown exporter in `internal/tui/export/` that can now be removed (or replaced with tests in `internal/export` to prevent drift).
- The “standard CLI testing playbook” and the doc topic already contain `generate --export-context` examples; the task list appears slightly stale.

### Open questions / next steps
- Add automated CLI coverage: exercise `generate --export-context` for all separators and `--output-file`.
- Move the existing exporter test to `internal/export` and delete the duplicate `internal/tui/export` package to avoid drift.

### Commits (if any)
- N/A

## Step 7: Lock down export-context behavior (tests) + remove exporter drift

**Commit (code):** `1b25b00` — test: cover generate --export-context and dedupe exporter

### What I did
- Extended `prescribe/test/test-cli.sh` and `prescribe/test/test-all.sh` to exercise:
  - `prescribe generate --export-context` (default xml)
  - `--separator` in all modes: xml/markdown/simple/begin-end/default
  - `--output-file` in export-only mode (non-empty file output)
- Fixed an issue where those shell tests could accidentally run a stale `/tmp/prescribe` binary:
  they now rebuild a per-`git rev-parse --short HEAD` binary each run.
- Moved the markdown exporter unit test from the old `internal/tui/export` package into `internal/export`
  (test now targets `BuildGenerationContext(..., SeparatorMarkdown)`).
- Deleted the duplicate exporter implementation under `internal/tui/export` to avoid format drift.

### Why
- The export-only milestone is a stable seam we’ll rely on for future inference integration; it needs regression coverage.
- Keeping a single exporter implementation avoids subtle TUI vs CLI differences over time.

### What worked
- `go test ./...` in the `prescribe` module passes.
- Shell smoke tests (`bash prescribe/test/test-cli.sh` and `bash prescribe/test/test-all.sh`) now pass and validate the new flags.

### Open questions / next steps
- None for the export milestone; next work is the “Later” inference path (templating + deterministic output parsing).

### Commits (if any)
- `1b25b00` - test: cover generate --export-context and dedupe exporter

## Step 8: Start “real prompt” support for inference (split + render templates)

**Commit (code):** N/A — Implementation (in progress)

### What I did
- Reviewed the current inference path (`internal/api/api.go`) and confirmed we currently feed the *entire* stored prompt string as a system prompt.
- Noted that the default prompt template is pinocchio-style and contains Go template directives (`{{ define "context" }}`, `{{ .diff }}`, etc.) which are not currently rendered before inference.

### Why
- Without rendering, the model sees raw template syntax and missing variables, which hurts output quality and makes presets misleading.
- The ticket’s “Later” milestone explicitly calls out pinocchio-style prompt templating as the next functional step after export-context.

### Next steps
- Implement a small prompt compiler:
  - split combined prompt into **system template** + **user template** (default marker: `{{ define "context"`),
  - render both using `text/template` + a minimal FuncMap (e.g. `join`),
  - populate `.diff`, `.code`, and `.context` from `GenerateDescriptionRequest`,
  - fall back to the existing system+userContext behavior when the prompt isn’t template-shaped.

### Commits (if any)
- N/A

## Step 9: Switch to Glazed templating helpers (sprig + TemplateFuncs) like Pinocchio

**Commit (code):** `fd6eeed` — feat: render pinocchio-style prompt templates via glazed templating

### What I did
- Located Glazed’s canonical template helper: `glazed/pkg/helpers/templating` (`CreateTemplate` wires in sprig + `TemplateFuncs`).
- Confirmed Pinocchio’s runtime prompt rendering uses the same helper (`templating.CreateTemplate(...).Parse(...).Execute(...)`).
- Implemented prompt compilation for prescribe inference:
  - detect pinocchio-style combined prompt (contains `{{ define "context"`),
  - split into system vs user template,
  - render both using Glazed templating helpers,
  - map prescribe data into template vars (`diff`, `code`, `context`, `description`, …),
  - fall back to the previous behavior for non-template prompts.
- Added unit tests validating both the templated path (no raw `{{ ... }}` left; diff/context appear) and the fallback path.

### Why
- The default prompt preset embedded in prescribe is pinocchio-style and contains Go template directives; without rendering, the model sees raw template syntax.
- Using the shared Glazed helper keeps parity with Pinocchio and gives us sprig functions “for free”.

### What worked
- `go test ./...` passes in the `prescribe` module after wiring the new compilation step.

### Commits (if any)
- `fd6eeed` - feat: render pinocchio-style prompt templates via glazed templating

## Step 10: Research + write the streaming/template/render/parsing analysis doc

**Commit (code):** N/A — Documentation/analysis

### What I did
- Reviewed Geppetto’s streaming reference (`geppetto/cmd/examples/simple-streaming-inference/main.go`) to capture the canonical `EventRouter` + `WatermillSink` + `errgroup` pattern.
- Reviewed Geppetto’s structured streaming extraction design (`geppetto/pkg/doc/topics/11-structured-data-event-sinks.md`) to understand tag-only sinks and extractor-owned YAML parsing (`parsehelpers`).
- Wrote a detailed end-to-end analysis doc in the ticket describing:
  - template rendering pipeline (Glazed templating),
  - streaming inference wiring for the TUI,
  - deterministic extraction/parsing of PR YAML output into structured fields,
  - an optional “structured tag” approach for robust streaming typed extraction.

### Why
- We need a concrete blueprint before implementing streaming in the TUI and before locking down a stable output format/parser for PR data.

### Output
- (moved) `ttmp/2025/12/28/GENERATE-LLM-INFERENCE--llm-inference-streaming-and-parsing/analysis/02-analysis-template-rendering-streaming-and-prdata-extraction.md`

### Commits (if any)
- N/A

## Step 11: Refine analysis docs (XML diff boundaries, robust final parse, split TUI doc)

**Commit (code):** N/A — Documentation/analysis

### What I did
- Updated the streaming/template/render/parsing analysis doc to:
  - make it explicit that “concat all diffs” means using the **XML separator exporter style** (per-file `<file><diff>...</diff></file>` boundaries) rather than naive string joins,
  - incorporate robust final-output normalization using existing Geppetto helpers:
    - `geppetto/pkg/steps/parse.ExtractYAMLBlocks` (multi-block fenced YAML extraction), and
    - `geppetto/pkg/events/structuredsink/parsehelpers` (fence stripping + YAML parsing helpers),
  - focus “streaming” on **stdio terminal output** for the current milestone.
- Split the TUI streaming design notes into a separate document for later implementation.

### Why
- The exporter already gives us a deterministic boundary format (XML default); the analysis should reflect that to avoid drift.
- We want robust extraction from the final Turn even without structured sinks (which mainly help with streaming).
- The near-term deliverable is CLI/stdio streaming output; the TUI wiring can follow later without polluting the core pipeline doc.

### Output
- Updated (moved): `ttmp/2025/12/28/GENERATE-LLM-INFERENCE--llm-inference-streaming-and-parsing/analysis/02-analysis-template-rendering-streaming-and-prdata-extraction.md`
- Added (moved): `ttmp/2025/12/28/GENERATE-LLM-INFERENCE--llm-inference-streaming-and-parsing/analysis/03-analysis-tui-streaming-integration.md`

### Commits (if any)
- N/A

## Step 12: Add `generate --export-rendered` (render prompt templates, no inference)

**Commit (code):** N/A — Implementation (in progress)

### What I did
- Added a new export-only mode for `prescribe generate` to print the **rendered** LLM payload `(system,user)` without running inference.
- Reused the same Glazed templating helpers as Pinocchio to render the prompt template.

### Why
- The existing `--export-context` is great for inspecting the canonical request + file/context payload, but it prints the prompt template “as stored”.
- For debugging prompt templating, we also want “what the model actually sees”, without spending tokens.

### Next steps
- Run `go test ./...` and the shell smoke tests.
- Commit once everything is green.

## Step 13: Validate `generate --export-rendered` (green tests) + close ticket tasks

This step was about making sure the `--export-rendered` workflow is actually shippable: it should render the pinocchio-style prompt template into the concrete `(system,user)` payload, support the same separator styles as `--export-context`, and honor `--output-file`—all without running inference.

After confirming the implementation exists end-to-end (CLI flag wiring → exporter → prompt compiler) I ran the `prescribe` module unit tests and the CLI smoke suites to lock down the behavior in automation. With that green, I’m updating the ticket bookkeeping so the open items reflect reality.

**Commit (docs):** `aa7c015` — "docs: close out 008 export-rendered tasks"

**Commit (code):** N/A — verification step

### What I did
- Verified the CLI export-only path in `cmd/prescribe/cmds/generate.go` supports:
  - `--export-rendered` (mutually exclusive with `--export-context`)
  - `--separator` selection
  - `--output-file` writes for export-only mode
- Verified the exporter implementation in `internal/export/context.go` emits `(system,user)` in all separator modes (XML/markdown/simple/begin-end/default).
- Ran tests:
  - `go test ./... -count=1`
  - `bash test/test-cli.sh`
  - `bash test/test-all.sh`

### Why
- `--export-rendered` is a debugging seam: it lets us confirm “what the model sees” without spending tokens. It needs to be deterministic and well-tested.

### What worked
- Unit tests passed for `internal/api` and `internal/export`.
- Both shell smoke suites passed, including checks that:
  - XML output contains `<llm_payload>`
  - markdown output contains `# Prescribe LLM payload (rendered)`
  - `--output-file` produces a non-empty rendered file

### What didn't work
- N/A

### What I learned
- The existing smoke test suite is already a great guardrail for keeping export-only modes stable across refactors (especially `--separator` output shape expectations).

### What was tricky to build
- N/A (verification step)

### What warrants a second pair of eyes
- Confirm that `--separator` help text (“Separator format for export flags…”) is clear that the rendered payload uses its own envelope (not the same XML as `--export-context`).
- Confirm the pinocchio prompt-splitting heuristic (`{{ define "context" ... }}` boundary) is “strict enough” for our embedded presets without surprising edge-cases.

### What should be done in the future
- Consider adding branch/commit metadata to the rendered payload envelope (similar to the “Later” task for exporter commit refs).

### Code review instructions
- Start in `cmd/prescribe/cmds/generate.go` and review the export-only branch (`--export-context` / `--export-rendered`).
- Review `internal/export/context.go` (`BuildRenderedLLMPayload`) for delimiter formats.
- Run:
  - `go test ./... -count=1`
  - `bash test/test-cli.sh`

### Technical details
- `--export-rendered` output is built by `internal/export.BuildRenderedLLMPayload`, which calls `internal/api.CompilePrompt` to produce the concrete system/user strings (template rendering when applicable).

## Step 14: Update help docs in `pkg/doc` for export-only modes

This step tightens the user-facing help topic so it matches current CLI behavior: `--export-context` prints the canonical “stored prompt + included files + additional context” blob, while `--export-rendered` prints the rendered `(system,user)` payload that seeds the LLM Turn.

The key nuance is templating: if the stored prompt is pinocchio-style Go templates, `--export-context` will still show the raw template, and `--export-rendered` is the one that shows what the model actually sees. I also made it explicit that `--export-context` and `--export-rendered` are mutually exclusive, and that `--separator` only affects export modes.

**Commit (docs):** `bcab6e1` — "docs: clarify export-context vs export-rendered"

### What I did
- Updated the `how-to-generate-pr-description` help topic to clarify:
  - `--export-context` vs `--export-rendered`
  - mutual exclusion of export flags
  - `--separator` applies to export modes

### Why
- Avoid confusion when prompts contain Go template directives: users need a reliable “show me what the model sees” switch.

### Code review instructions
- Review `pkg/doc/topics/02-how-to-generate-pr-description.md` around the Step 5 export sections.

## Step 15: Reconcile branch names with commit SHAs in export payloads

This step makes exports reproducible. Previously, export payloads only recorded branch names (e.g. `feature/foo` vs `master`). Since branches move, that meant an exported context could become ambiguous later: “feature/foo” might point to a different commit tomorrow than it did when you exported.

I fixed that by resolving the branch refs to concrete commit SHAs at request-build time and emitting those SHAs in both `--export-context` and `--export-rendered` outputs. I also attach the source commit ref to context-file items so you can tell which exact tree the context file content came from.

**Commit (code):** N/A — implementation in progress

### What I did
- Added ref-to-SHA resolution to the git layer (`git rev-parse`) and threaded:
  - `SourceCommit`
  - `TargetCommit`
  through the canonical `GenerateDescriptionRequest`.
- Updated exporters:
  - `--export-context`: add `<commits>` to XML and commit lines to markdown/simple/begin-end/default
  - `--export-rendered`: add `<commits>` to the rendered XML envelope and commit info to other separators
  - add `commit="..."` attribute to XML context-file items (sourced from the source commit)
- Tightened shell smoke tests to assert commit tags are present and non-empty in XML exports.

### Why
- Export payloads should be self-contained and reproducible; branch names alone are not stable identifiers.

### What worked
- `go test ./...` passes.
- `bash test/test-cli.sh` and `bash test/test-all.sh` pass and now validate `<source_commit>`/`<target_commit>` tags in the XML export outputs.

### What warrants a second pair of eyes
- Confirm the choice of `git diff target...source` (merge-base semantics) matches how reviewers expect “target commit” to be interpreted. We currently record the ref’s resolved commit SHA, not the merge-base SHA.

### Code review instructions
- Start at `internal/controller/controller.go` (`BuildGenerateDescriptionRequest`) and verify commit resolution is best-effort and does not break non-git unit tests.
- Review `internal/export/context.go` for the additive output changes (`<commits>` blocks and context-file `commit` attribute).
- Validate via:
  - `go test ./... -count=1`
  - `bash test/test-cli.sh`
