---
Title: 'Analysis: Export Prescribe Diff Data and Generate PR Descriptions with Geppetto Inference'
Ticket: 008-GENERATE
Status: active
Topics:
    - prescribe
    - geppetto
    - inference
    - pr-generation
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-27T18:17:16.177493457-05:00
WhatFor: ""
WhenToUse: ""
---

# Analysis: Export Prescribe Diff Data and Generate PR Descriptions with Geppetto Inference

## Executive Summary

This document analyzes three key components needed to implement AI-powered PR description generation within the `prescribe` binary:

1. **Catter Implementation**: How `catter` exports file contents and statistics for LLM context preparation
2. **Create Pull Request Pattern**: How pinocchio's `create-pull-request` prompt template works and can be adapted
3. **Geppetto Inference Engine**: How to perform real inference from within prescribe using geppetto's Turn-based architecture

The goal is to replace prescribe's mock API service (`internal/api/api.go`) with a real geppetto inference engine that generates PR descriptions using the same prompt template structure as pinocchio.

## 1. Catter Implementation Analysis

### 1.1 Architecture Overview

**Location**: `pinocchio/cmd/pinocchio/cmds/catter/`

Catter is a CLI tool designed to prepare codebase content for LLM contexts. It processes files recursively, applies filters, and provides token counting statistics.

**Key Components**:

- **Command Layer** (`cmds/print.go`, `cmds/stats.go`): Glazed-based command definitions
- **Processing Layer** (`pkg/fileprocessor.go`): Core file processing logic
- **Statistics Layer** (`pkg/stats.go`): Token counting and statistics computation

### 1.2 FileProcessor Architecture

**File**: `pinocchio/cmd/pinocchio/cmds/catter/pkg/fileprocessor.go`

The `FileProcessor` struct handles:
- File filtering via `filefilter.FileFilter`
- Token counting using `tiktoken-go` (cl100k_base encoding)
- Multiple output formats: text, zip, tar.gz
- Size and token limits (per-file and total)
- Line truncation

**Key Functions**:

```go
// Constructor with options pattern
func NewFileProcessor(options ...FileProcessorOption) *FileProcessor

// Main processing entry point
func (fp *FileProcessor) ProcessPaths(paths []string) error

// Process individual file
func (fp *FileProcessor) processFileContent(filePath string, fileInfo os.FileInfo) error

// Apply size/token limits
func (fp *FileProcessor) applyLimits(contentBytes []byte) string
```

**Output Modes**:
- **Text**: Prints to stdout with configurable delimiters (default, xml, markdown, simple, begin-end)
- **Archive**: Creates zip or tar.gz archives with optional prefix
- **Glazed**: Uses `middlewares.Processor` for structured output

### 1.3 Statistics Computation

**File**: `pinocchio/cmd/pinocchio/cmds/catter/pkg/stats.go`

The `Stats` struct tracks:
- Per-file statistics (tokens, lines, size)
- Per-file-type aggregations
- Per-directory aggregations
- Total statistics

**Key Functions**:

```go
// Pre-compute stats for all files
func (s *Stats) ComputeStats(paths []string, filter *filefilter.FileFilter) error

// Get stats for a specific file
func (s *Stats) GetStats(path string) (FileStats, bool)
```

**Computation Flow**:
1. Uses `filewalker.Walker` to traverse paths
2. Applies `filefilter.FileFilter` during traversal
3. Reads file content and counts tokens using tiktoken
4. Aggregates statistics by file, type, and directory

### 1.4 Filter System

Catter uses the `clay/pkg/filefilter` package which provides:
- Extension include/exclude patterns
- Filename/path regex matching
- Directory exclusion
- Gitignore support
- Binary file filtering
- YAML configuration profiles

### 1.5 Separator/Delimiter Approaches

**Location**: `pinocchio/cmd/pinocchio/cmds/catter/pkg/fileprocessor.go` (lines 402-413)

Catter supports multiple delimiter types for formatting file output:

1. **XML** (default for prescribe):
   ```xml
   <file name="path/to/file.go">
   <content>
   ...file content...
   </content>
   </file>
   ```

2. **Markdown**:
   ```markdown
   ## File: path/to/file.go
   
   ```
   ...file content...
   ```
   ```

3. **Simple**:
   ```
   --- START FILE: path/to/file.go ---
   ...file content...
   --- END FILE: path/to/file.go ---
   ```

4. **Begin-End**:
   ```
   --- BEGIN FILE: path/to/file.go ---
   ...file content...
   --- END FILE: path/to/file.go ---
   ```

5. **Default** (plain):
   ```
   File: path/to/file.go
   ...file content...
   ```

**Implementation Pattern**:
```go
switch delimiterType {
case "xml":
    fmt.Printf("<file name=\"%s\">\n<content>\n%s\n</content>\n</file>\n", path, content)
case "markdown":
    fmt.Printf("## File: %s\n\n```\n%s\n```\n\n", path, content)
case "simple":
    fmt.Printf("--- START FILE: %s ---\n%s\n--- END FILE: %s ---\n", path, content, path)
case "begin-end":
    fmt.Printf("--- BEGIN FILE: %s ---\n%s\n--- END FILE: %s ---\n", path, content, path)
default:
    fmt.Printf("File: %s\n%s\n", path, content)
}
```

### 1.6 Adaptation for Prescribe

**What prescribe needs**:
- Export `domain.PRData` (changed files, context, prompts) in a format suitable for LLM inference
- Similar to `export.BuildGenerationContextText()` but structured for geppetto Turn format
- Token counting for context size management
- **Configurable separator/delimiter approach** (default: XML) for formatting different content types:
  - Diffs (unified diff format)
  - Full file content (before/after/both)
  - Manual prompts
  - Additional context files
  - Context notes

**Key Differences**:
- Catter processes files from filesystem; prescribe needs to export in-memory `FileChange` structs
- Catter focuses on file content; prescribe needs diffs, full files, and additional context
- Catter outputs text/archives; prescribe needs to build Turn blocks for geppetto
- Prescribe needs to handle multiple content types (diffs, full files, prompts, context) with appropriate separators

**Separator Strategy for Prescribe**:

Prescribe should support configurable separators (default: XML) for different content types. All file-related elements should include commit version information:

1. **Diffs**: Use XML wrapper with diff-specific attributes and commit versions
   ```xml
   <file name="path/to/file.go" type="diff" source-commit="abc123def" target-commit="xyz789ghi">
   <diff>
   ...unified diff content...
   </diff>
   </file>
   ```
   - `source-commit`: Commit hash of source branch (e.g., `git rev-parse --short HEAD`)
   - `target-commit`: Commit hash of target branch (e.g., `git rev-parse --short main`)

2. **Full Files**: Use XML wrapper with version attribute and commit version
   ```xml
   <file name="path/to/file.go" type="full" version="after" commit="abc123def">
   <content>
   ...file content...
   </content>
   </file>
   ```
   - `commit`: Commit hash corresponding to the version (before = target commit, after = source commit, both = both commits)

3. **Manual Prompts**: Use XML wrapper for prompt sections
   ```xml
   <prompt>
   <text>
   ...prompt text...
   </text>
   </prompt>
   ```

4. **Context Files**: Use XML wrapper for additional context with optional commit version
   ```xml
   <context type="file" path="path/to/context.go" commit="abc123def">
   <content>
   ...context content...
   </content>
   </context>
   ```

5. **Context Notes**: Use XML wrapper for notes
   ```xml
   <context type="note">
   <text>
   ...note text...
   </text>
   </context>
   ```

**Pseudocode for Prescribe Export with Separators**:

```go
// Similar to catter's FileProcessor but adapted for PRData
type PRDataExporter struct {
    data *domain.PRData
    tokenCounter *tiktoken.Tiktoken
    maxTokens int
    maxTotalSize int64
}

func (e *PRDataExporter) ExportToTurnBlocks() ([]turns.Block, error) {
    blocks := []turns.Block{}
    
    // Add system prompt block
    blocks = append(blocks, turns.NewSystemTextBlock(e.data.CurrentPrompt))
    
    // Add user prompt with context
    userPrompt := e.buildUserPrompt()
    blocks = append(blocks, turns.NewUserTextBlock(userPrompt))
    
    return blocks, nil
}

func (e *PRDataExporter) buildUserPrompt() string {
    var b strings.Builder
    
    // Add branch info
    b.WriteString(fmt.Sprintf("Source: %s\nTarget: %s\n\n", 
        e.data.SourceBranch, e.data.TargetBranch))
    
    // Add included files (similar to export.BuildGenerationContextText)
    for _, file := range e.data.GetVisibleFiles() {
        if !file.Included {
            continue
        }
        
        b.WriteString(fmt.Sprintf("### %s\n\n", file.Path))
        switch file.Type {
        case domain.FileTypeDiff:
            b.WriteString("```diff\n")
            b.WriteString(file.Diff)
            b.WriteString("\n```\n\n")
        case domain.FileTypeFull:
            // Include full file content based on Version
            // ...
        }
    }
    
    // Add additional context
    for _, ctx := range e.data.AdditionalContext {
        // ...
    }
    
    return b.String()
}
```

## 2. Create Pull Request Pattern Analysis

### 2.1 Pinocchio Prompt Template

**File**: `pinocchio/cmd/pinocchio/prompts/code/create-pull-request.yaml`

The template defines:
- **System prompt**: Sets role as experienced software engineer
- **User prompt**: Structured template with sections for commits, issue, description, diff, code files, context
- **Output format**: YAML with `title`, `body`, `changelog`, `release_notes`

**Key Template Variables**:
- `{{ .commits }}`: Commit history
- `{{ .diff }}`: Unified diff of changes
- `{{ .code }}`: List of code files with content
- `{{ .context }}`: Additional context files
- `{{ .description }}`: User-provided description
- `{{ .issue }}`: Related issue description

**Output Structure**:
```yaml
title: ...
body: |
  ...
changelog: |
  ...
release_notes:
  title: ...
  body: |
    ...
```

### 2.2 Prescribe's Current Prompt

**File**: `prescribe/internal/prompts/assets/create-pull-request.yaml`

Prescribe embeds the same template but combines system + prompt into a single string:

```go
// From internal/prompts/default.go
func DefaultPrompt() string {
    // Combines system-prompt + prompt from YAML
    return strings.TrimSpace(p.SystemPrompt) + "\n\n" + strings.TrimSpace(p.Prompt)
}
```

**Current Usage**:
- Stored in `domain.PRData.CurrentPrompt`
- Used by `api.Service.GenerateDescription(ctx, req)` (real geppetto inference via Turns; configured via StepSettings)
- Formatted by `export.BuildGenerationContextText()` for display

### 2.3 Pinocchio's Execution Flow

**File**: `pinocchio/pkg/cmds/cmd.go`

Pinocchio executes prompts using geppetto:

```go
// From runEngineAndCollectMessages
func (g *PinocchioCommand) runEngineAndCollectMessages(...) error {
    // Create engine from parsed layers
    engine, err := rc.EngineFactory.CreateEngine(rc.StepSettings, options...)
    
    // Build seed Turn from system + messages + prompt (rendered)
    seed, err := g.buildInitialTurn(rc.Variables, rc.ImagePaths)
    
    // Run inference
    updatedTurn, err := engine.RunInference(ctx, seed)
    
    // Store result
    rc.ResultTurn = updatedTurn
    return nil
}
```

**Template Rendering**:
- Pinocchio uses Go template syntax (`{{ .variable }}`)
- Variables come from flags, files, and context
- Templates are rendered before building Turn blocks

### 2.4 Adaptation for Prescribe

**What prescribe needs**:
- Use the same prompt template structure
- Render template with prescribe's data (files, context, branches)
- Execute via geppetto engine instead of mock API

**Key Differences**:
- Pinocchio loads prompts from YAML files; prescribe embeds them
- Pinocchio has flag-based variables; prescribe has `PRData` struct
- Pinocchio uses `buildInitialTurn()`; prescribe needs custom Turn builder

**Pseudocode for Prescribe Integration (StepSettings injected; Turns only)**:

```go
// In internal/api/api.go
type Service struct {
    stepSettings *settings.StepSettings
}

// Parsing StepSettings happens higher up (CLI/TUI). Service just consumes it.
func (s *Service) SetStepSettings(ss *settings.StepSettings) { s.stepSettings = ss }

func (s *Service) GenerateDescription(ctx context.Context, req GenerateDescriptionRequest) (*GenerateDescriptionResponse, error) {
    // Build a seed Turn directly (no conversation.Manager).
    seed := turns.NewTurnBuilder().
        WithSystemPrompt(strings.TrimSpace(req.Prompt)).
        WithUserPrompt(BuildExportedContext(req, "xml")). // exporter decides separators; default xml
        Build()

    eng, err := factory.NewEngineFromStepSettings(s.stepSettings)
    if err != nil { return nil, err }

    updated, err := eng.RunInference(ctx, seed)
    if err != nil { return nil, err }

    description := ExtractLastAssistantText(updated)
    return &GenerateDescriptionResponse{ Description: description }, nil
}
```

**Where StepSettings come from**:
- `geppetto/pkg/steps/ai/settings/settings-step.go`: `settings.StepSettings`
- `settings.NewStepSettingsFromParsedLayers(parsedLayers)` (used in `prescribe generate` + `prescribe tui`)
- Engine creation: `geppetto/pkg/inference/engine/factory/helpers.go`: `factory.NewEngineFromStepSettings(stepSettings, ...)`

**Streaming pattern reference**:
- `geppetto/cmd/examples/simple-streaming-inference/main.go`: `events.NewEventRouter()` + `middleware.NewWatermillSink(...)` + `engine.WithSink(...)` + `errgroup` to run router and inference concurrently.

## 3. Geppetto Inference Engine Architecture

### 3.1 Core Concepts

**Engine Interface** (`geppetto/pkg/inference/engine/engine.go`):

```go
type Engine interface {
    RunInference(ctx context.Context, t *turns.Turn) (*turns.Turn, error)
}
```

**Key Principles**:
- Engines operate on `Turn` (ordered `Block`s + metadata)
- Engines handle provider-specific API calls
- Engines emit streaming events via sinks
- Tool orchestration handled by helpers/middleware

### 3.2 Turn Structure

**File**: `geppetto/pkg/turns/types.go`

```go
type Turn struct {
    ID     string
    RunID  string
    Blocks []Block
    Metadata map[TurnMetadataKey]interface{}
    Data map[TurnDataKey]interface{}
}

type Block struct {
    ID      string
    TurnID  string
    Kind    BlockKind  // llm_text, tool_call, system_text, user_text, etc.
    Role    string
    Payload map[string]any
    Metadata map[BlockMetadataKey]interface{}
}
```

**Block Kinds**:
- `BlockKindSystemText`: System prompt
- `BlockKindUserText`: User message
- `BlockKindAssistantText`: LLM response text
- `BlockKindToolCall`: Tool invocation request
- `BlockKindToolUse`: Tool execution result

### 3.3 Building Turns

**File**: `geppetto/pkg/turns/builders.go`

```go
// Builder pattern for creating Turns
seed := turns.NewTurnBuilder().
    WithSystemPrompt("You are a helpful assistant.").
    WithUserPrompt("What is the weather?").
    Build()

// Or manually append blocks
turn := &turns.Turn{}
turns.AppendBlock(turn, turns.NewSystemTextBlock("..."))
turns.AppendBlock(turn, turns.NewUserTextBlock("..."))
```

### 3.4 Conversation Builder (Legacy)

**File**: `geppetto/pkg/conversation/builder/builder.go`

The `ManagerBuilder` creates `conversation.Manager` instances, but the recommended approach is to use Turns directly:

```go
// From 06-inference-engines.md example
mgr, err := builder.NewManagerBuilder().
    WithSystemPrompt("You are a helpful assistant.").
    WithPrompt(prompt).
    Build()

// Convert conversation to Turn
seed := &turns.Turn{}
turns.AppendBlocks(seed, turns.BlocksFromConversationDelta(mgr.GetConversation(), 0)...)

// Run inference
updated, err := engine.RunInference(ctx, seed)

// Convert back to conversation for display
for _, m := range turns.BuildConversationFromTurn(updated) {
    fmt.Println(m.Content.String())
}
```

**Note**: The documentation recommends using Turns directly rather than conversation.Manager for new code.

### 3.5 Engine Factory

**File**: `geppetto/pkg/inference/engine/factory/factory.go`

```go
// Create engine from configuration layers
engine, err := factory.NewEngineFromParsedLayers(parsedLayers, engineOptions...)
```

**Engine Options**:
- `engine.WithSink(sink)`: Add event sink for streaming
- Provider-specific options via layers

**Configuration Layers**:
- API keys, model selection, base URLs
- Provider-specific settings (OpenAI, Claude, Gemini)
- Parsed from YAML/CLI flags via glazed layers

### 3.6 Streaming Events

**File**: `geppetto/pkg/inference/middleware/sink.go`

Engines can emit events via sinks:

```go
// Create Watermill sink for event publishing
watermillSink := middleware.NewWatermillSink(publisher, "chat")

engineOptions := []engine.Option{
    engine.WithSink(watermillSink),
}

engine, err := factory.NewEngineFromParsedLayers(parsedLayers, engineOptions...)
```

**Event Types**:
- `StepStart`: Inference started
- `StepPartial`: Partial response (streaming)
- `StepFinal`: Final response
- `ToolCall`: Tool invocation
- `ToolResult`: Tool execution result

### 3.7 Integration Pattern for Prescribe

**Pseudocode**:

```go
// In internal/api/service.go
type Service struct {
    parsedLayers *layers.ParsedLayers
    eventSink middleware.Sink  // optional
}

func NewService(parsedLayers *layers.ParsedLayers) *Service {
    return &Service{
        parsedLayers: parsedLayers,
    }
}

func (s *Service) GenerateDescription(ctx context.Context, req GenerateDescriptionRequest) (*GenerateDescriptionResponse, error) {
    // 1. Create engine
    options := []engine.Option{}
    if s.eventSink != nil {
        options = append(options, engine.WithSink(s.eventSink))
    }
    
    e, err := factory.NewEngineFromParsedLayers(s.parsedLayers, options...)
    if err != nil {
        return nil, fmt.Errorf("failed to create engine: %w", err)
    }
    
    // 2. Build Turn from request
    turn := s.buildTurnFromRequest(req)
    
    // 3. Run inference
    updatedTurn, err := e.RunInference(ctx, turn)
    if err != nil {
        return nil, fmt.Errorf("inference failed: %w", err)
    }
    
    // 4. Extract description from Turn
    description := s.extractDescriptionFromTurn(updatedTurn)
    
    // 5. Extract metadata (tokens, model)
    tokensUsed := s.extractTokensFromTurn(updatedTurn)
    model := s.extractModelFromTurn(updatedTurn)
    
    return &GenerateDescriptionResponse{
        Description: description,
        TokensUsed: tokensUsed,
        Model: model,
    }, nil
}

func (s *Service) buildTurnFromRequest(req GenerateDescriptionRequest) *turns.Turn {
    turn := &turns.Turn{}
    
    // Extract system prompt from combined prompt
    systemPrompt, userPrompt := s.splitPrompt(req.Prompt)
    
    // Add system block
    turns.AppendBlock(turn, turns.NewSystemTextBlock(systemPrompt))
    
    // Build user prompt with context (similar to export.BuildGenerationContextText)
    userContent := s.buildUserPromptContent(req)
    
    // Add user block
    turns.AppendBlock(turn, turns.NewUserTextBlock(userContent))
    
    return turn
}

func (s *Service) extractDescriptionFromTurn(turn *turns.Turn) string {
    // Find last assistant text block
    blocks := turns.FindLastBlocksByKind(*turn, turns.BlockKindAssistantText)
    if len(blocks) == 0 {
        return ""
    }
    
    // Extract text from payload
    lastBlock := blocks[len(blocks)-1]
    if text, ok := lastBlock.Payload[turns.PayloadKeyText].(string); ok {
        return text
    }
    
    return ""
}
```

## 4. Separator Implementation Strategy

### 4.1 Separator Types and Use Cases

**Default: XML** (recommended for LLM consumption)

XML separators provide:
- Structured, parseable format
- Clear boundaries between content sections
- Metadata via attributes (type, version, path)
- Escaping support for special characters
- Consistent with many LLM training formats

**Alternative: Markdown** (human-readable)

Markdown separators provide:
- Human-readable format
- Good for copy/paste into PRs or documentation
- Familiar syntax for developers
- Less structured than XML

**Alternative: Simple/Begin-End** (minimal)

Simple separators provide:
- Minimal overhead
- Easy to parse programmatically
- Less metadata support

### 4.2 Content Type Handling

Different content types need different formatting:

1. **Diffs**: Unified diff format with `+`/`-` prefixes
   - XML: Wrap in `<diff>` tags with `type="diff"` attribute, include `source-commit` and `target-commit` attributes
   - Markdown: Use ````diff` code fence with commit info in header
   - Simple: Plain diff with file boundaries and commit info

2. **Full Files**: Complete file content
   - XML: Wrap in `<content>` tags with `version` attribute (before/after/both) and `commit` attribute
   - Markdown: Use ````text` code fence with commit info in header
   - Simple: Plain content with file boundaries and commit info

3. **Prompts**: User-provided prompt text
   - XML: Wrap in `<prompt><text>` tags
   - Markdown: Use ````text` code fence with "Prompt:" header
   - Simple: Plain text with "PROMPT:" prefix

4. **Context Files**: Additional files added as context
   - XML: Wrap in `<context type="file">` with `path` attribute
   - Markdown: Use ````text` code fence with file path header
   - Simple: Plain content with "CONTEXT FILE:" prefix

5. **Context Notes**: Text notes added as context
   - XML: Wrap in `<context type="note"><text>` tags
   - Markdown: Use bullet point or blockquote
   - Simple: Plain text with "NOTE:" prefix

### 4.3 Configuration

Separator type should be configurable via:
- CLI flag: `--separator xml|markdown|simple|begin-end|default`
- Configuration file: `separator: xml` in session/config
- Default: `xml` (as specified)

**Immediate milestone (export-first)**:
- `prescribe generate --export-context --separator xml` prints the full “what we would send to the model” payload and exits (no inference).

## 5. Implementation Plan

### 5.1 Phase 1: Export Infrastructure

**Goal**: Create exporter similar to catter that converts `PRData` to Turn blocks with configurable separators

**Tasks**:
1. Create `internal/export/turn.go` with `PRDataToTurn()` function
2. Create `internal/export/separator.go` with separator formatting functions
3. Adapt catter's file processing logic for `FileChange` structs
4. Implement token counting using tiktoken
5. Handle diff vs full file content based on `FileChange.Type`
6. Support all separator types (xml, markdown, simple, begin-end, default)
7. Format prompts, context files, and context notes with separators

**Files to Create**:
- `prescribe/internal/export/turn.go` (Turn building)
- `prescribe/internal/export/separator.go` (Separator formatting)
- `prescribe/internal/export/prompt.go` (Prompt splitting/rendering)

### 5.2 Phase 2: Geppetto Integration

**Goal**: Replace mock API service with real geppetto engine

**Tasks**:
1. Add geppetto dependencies to `go.mod`
2. Create engine factory in `internal/api/service.go`
3. Implement `GenerateDescription()` using geppetto engine
4. Add configuration layers for API keys/models
5. Handle Turn → description extraction

**Files to Modify**:
- `prescribe/internal/api/api.go` (replace mock with real)
- `prescribe/cmd/prescribe/main.go` (add glazed layers for engine config)
- `prescribe/pkg/layers/` (add inference engine layers)

### 5.3 Phase 3: Prompt Template Integration

**Goal**: Use pinocchio-style prompt template with prescribe data

**Tasks**:
1. Extract system/user prompts from combined prompt string
2. Render template with prescribe variables (files, context, branches)
3. Support template variables like `{{ .diff }}`, `{{ .code }}`, etc.
4. Maintain compatibility with existing prompt presets

**Files to Modify**:
- `prescribe/internal/prompts/default.go` (add template rendering)
- `prescribe/internal/export/prompt.go` (template variable mapping)

### 5.4 Phase 4: Streaming Support (Optional)

**Goal**: Add real-time streaming of generation progress

**Tasks**:
1. Integrate Watermill event router
2. Add event sink to engine options
3. Stream partial responses to TUI
4. Handle completion events

**Files to Modify**:
- `prescribe/internal/tui/` (add streaming UI updates)
- `prescribe/internal/api/service.go` (add event sink)

## 6. Key Design Decisions

### 6.1 Turn vs Conversation

**Decision**: Use Turns directly (not conversation.Manager)

**Rationale**:
- Geppetto documentation recommends Turns for new code
- Turns are simpler and more flexible
- Direct block manipulation matches prescribe's needs

### 6.2 Prompt Template Format

**Decision**: Keep combined system+user prompt but split for Turn blocks

**Rationale**:
- Maintains compatibility with existing prompt presets
- Allows gradual migration
- Can extract system/user prompts via heuristics or markers

### 6.3 Export Format

**Decision**: Build Turn blocks directly (not intermediate text format)

**Rationale**:
- More efficient (no string concatenation)
- Preserves structure for better LLM understanding
- Aligns with geppetto architecture

### 6.4 Separator Selection

**Decision**: Default to XML separator

**Rationale**:
- XML provides structured, parseable format ideal for LLM consumption
- Supports metadata via attributes (type, version, path)
- Consistent with many LLM training formats
- Can be easily converted to other formats if needed

**Alternative Considered**: Markdown. Rejected as default because:
- Less structured than XML
- Harder to parse programmatically
- Better suited for human-readable output

**Implementation**: Configurable via CLI flag and config file, with XML as default.

### 6.5 Configuration Management

**Decision**: Use glazed layers for engine configuration

**Rationale**:
- Consistent with pinocchio/geppetto patterns
- Supports multiple providers (OpenAI, Claude, Gemini)
- Handles API keys, models, base URLs via layers

## 7. Open Questions

1. **Prompt Splitting**: How to reliably split combined prompt into system/user? Use markers? Heuristics?
2. **Template Variables**: Should prescribe support full Go template syntax or subset?
3. **Error Handling**: How to handle API failures, rate limits, timeouts?
4. **Token Limits**: How to handle contexts exceeding model limits? Truncation strategy?
5. **Streaming UI**: How to integrate streaming into existing TUI without major refactor?

## 8. References

- **Catter Implementation**: `pinocchio/cmd/pinocchio/cmds/catter/`
- **Geppetto Inference Guide**: `geppetto/pkg/doc/topics/06-inference-engines.md`
- **Turns Documentation**: `geppetto/pkg/doc/topics/08-turns.md`
- **Pinocchio Prompt Template**: `pinocchio/cmd/pinocchio/prompts/code/create-pull-request.yaml`
- **Prescribe Export**: `prescribe/internal/tui/export/export.go`
- **Prescribe API Service**: `prescribe/internal/api/api.go`
