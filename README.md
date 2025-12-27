# Prescribe (PR Builder)

A CLI/TUI application for generating pull request descriptions using AI, built with Go and Bubbletea.

## Features

PR Builder provides a powerful, session-based workflow for creating high-quality PR descriptions:

- **Session Management**: Save and load PR configurations as YAML files
- **File Filtering**: Include/exclude files using glob patterns
- **Context Control**: Add additional files or notes as context
- **Prompt Customization**: Use presets or custom prompts
- **Interactive TUI**: Navigate and edit with keyboard shortcuts
- **CLI Commands**: Scriptable interface for automation
- **Token Counting**: Track context window usage
- **Git Integration**: Automatically detects changes from git

## Installation

### Prerequisites

- Go 1.25+ (download from https://go.dev/dl/)
- Git repository with changes

### Build from Source

```bash
cd prescribe
go build -o prescribe ./cmd/prescribe
```

## Quick Start

### Initialize a Session

```bash
# Navigate to your git repository
cd /path/to/your/repo

# Initialize PR builder session
prescribe session init --save

# This creates .pr-builder/session.yaml with your current changes
```

### View Session State

```bash
# Human-readable format
prescribe session show

# YAML format
prescribe session show --output yaml
```

### Generate PR Description

```bash
# Generate with default settings
prescribe generate

# Generate with custom prompt
prescribe generate --prompt "Write a concise 3-sentence description"

# Generate with preset
prescribe generate --preset concise

# Save to file
prescribe generate -o pr-description.md
```

### Launch Interactive TUI

```bash
prescribe tui
```

## CLI Commands

### Session Management

#### `session init`
Initialize a new session from current git state.

```bash
prescribe session init [--save] [--path PATH]
```

Options:
- `--save`: Automatically save session after init
- `--path, -p PATH`: Custom session file path

#### `session save`
Save current session to YAML file.

```bash
prescribe session save [PATH]
```

Default path: `.pr-builder/session.yaml`

#### `session load`
Load session from YAML file.

```bash
prescribe session load [PATH]
```

#### `session show`
Display current session state.

```bash
prescribe session show [--output yaml|json|csv|table]
```

This command is implemented as a Glazed query command, so it supports structured output formats via `--output ...`.

### File Management

#### `file toggle`
Toggle whether a file is included in the context.

```bash
prescribe file toggle <file-path>
```

Example:
```bash
prescribe file toggle src/auth/login.ts
```

### Filters

#### `filter add`
Add a glob-based filter to exclude or include files.

```bash
prescribe filter add --name NAME [--description DESC] [--exclude PATTERN]... [--include PATTERN]...
```

Options:
- `--name, -n NAME`: Filter name (required)
- `--description, -d DESC`: Filter description
- `--exclude, -e PATTERN`: Exclude pattern (can specify multiple)
- `--include, -i PATTERN`: Include pattern (can specify multiple)

Examples:
```bash
# Exclude all test files
prescribe filter add --name "Exclude tests" --exclude "*test*" --exclude "*spec*"

# Include only TypeScript files
prescribe filter add --name "Only TS" --include "*.ts"

# Complex filter
prescribe filter add \
  --name "Backend only" \
  --description "Only backend code" \
  --include "src/backend/**" \
  --exclude "*test*"
```

### Context

#### `context add`
Add additional context for PR generation.

```bash
# Add a file
prescribe context add <file-path>

# Add a note
prescribe context add --note "This PR is part of the Q1 security improvements"
```

### Generation

#### `generate`
Generate PR description using AI.

```bash
prescribe generate [--output-file PATH] [--prompt TEXT] [--preset ID] [--load-session PATH]
```

Options:
- `--output-file, -o PATH`: Output file (default: stdout)
- `--prompt, -p TEXT`: Custom prompt text
- `--preset ID`: Prompt preset ID (detailed, concise, technical)
- `--load-session, -s PATH`: Load session file before generating

Examples:
```bash
# Generate with default settings
prescribe generate

# Generate with custom prompt
prescribe generate --prompt "Write a technical PR description focusing on architecture changes"

# Generate using a specific session
prescribe generate --load-session /path/to/session.yaml -o pr.md

# Generate with preset
prescribe generate --preset concise
```

### TUI

#### `tui`
Launch interactive Terminal User Interface.

```bash
prescribe tui
```

Keyboard shortcuts:
- `↑/↓` or `j/k`: Navigate file list
- `Space`: Toggle file inclusion
- `g`: Generate PR description
- `Esc`: Go back (from result screen)
- `q`: Quit

## Session File Format

Sessions are stored as YAML files with the following structure:

```yaml
version: "1.0"
source_branch: feature/user-auth
target_branch: master

files:
  - path: src/auth/login.ts
    included: true
    mode: diff  # or full_before, full_after, full_both
  - path: src/auth/middleware.ts
    included: true
    mode: diff
  - path: tests/auth.test.ts
    included: false
    mode: diff

filters:
  - name: Exclude tests
    description: Hide test files from context
    rules:
      - type: exclude
        pattern: '*test*'
      - type: exclude
        pattern: '*spec*'

context:
  - type: file
    path: README.md
  - type: note
    content: "This PR is part of the auth refactor epic"

prompt:
  preset: detailed  # or use 'template' for custom
  # template: "Custom prompt text here"
```

## Use Cases

### Team Workflow

Share session configurations across your team:

```bash
# Create a session template
pr-builder init --save
pr-builder add-filter --name "Exclude generated" --exclude "dist/**" --exclude "build/**"
pr-builder save .pr-builder/team-template.yaml

# Commit the template
git add .pr-builder/team-template.yaml
git commit -m "Add PR builder template"

# Team members can load it
pr-builder load .pr-builder/team-template.yaml
pr-builder generate
```

### CI/CD Integration

Automate PR description generation in your CI pipeline:

```bash
#!/bin/bash
# .github/scripts/generate-pr-description.sh

# Initialize from current branch
pr-builder init --save

# Add filters for your project
pr-builder add-filter --name "Exclude tests" --exclude "*test*"
pr-builder add-filter --name "Exclude config" --exclude "*.config.js"

# Generate description
pr-builder generate -o pr-description.md

# Use with gh CLI
gh pr create --title "$(git log -1 --pretty=%s)" --body-file pr-description.md
```

### Interactive Refinement

Use the TUI for interactive refinement:

```bash
# Start with CLI to set up filters
pr-builder init --save
pr-builder add-filter --name "Exclude tests" --exclude "*test*"

# Switch to TUI for fine-tuning
pr-builder tui

# Navigate with j/k, toggle files with Space
# Press 'g' to generate when ready
```

## Architecture

### Core Components

The application follows a clean architecture with clear separation of concerns:

#### Domain Layer (`internal/domain`)
Contains pure business logic and data structures:
- `PRData`: Core data model for PR information
- `FileChange`: Represents a changed file with stats
- `Filter`: File filtering logic with glob patterns
- `ContextItem`: Additional context (files or notes)
- `PromptPreset`: Predefined prompt templates

#### Service Layer
External interactions:
- `internal/git`: Git repository operations
- `internal/api`: Mock API for PR description generation

#### Controller Layer (`internal/controller`)
Coordinates between domain and services:
- Initializes from git state
- Applies filters and transformations
- Manages session persistence
- Orchestrates generation

#### Session Layer (`internal/session`)
Handles YAML serialization/deserialization:
- Converts between domain model and YAML
- Saves/loads session files
- Validates session integrity

#### Presentation Layer
Two interfaces to the same core:
- **CLI** (`cmd/`): Cobra commands for scripting
- **TUI** (`internal/tui`): Bubbletea interface for interaction

### Design Patterns

**MVC Pattern**: Bubbletea TUI follows Model-View-Controller:
- Model: Contains state, implements Init/Update/View
- Update: Handles messages (events)
- View: Renders current state

**Repository Pattern**: Git service abstracts repository access

**Strategy Pattern**: Prompt presets allow different generation strategies

**Command Pattern**: Cobra CLI commands encapsulate operations

## Development

### Project Structure

```
pr-builder/
├── cmd/                    # CLI commands (Cobra)
│   ├── root.go
│   ├── init.go
│   ├── save.go
│   ├── load.go
│   ├── show.go
│   ├── generate.go
│   ├── add_filter.go
│   ├── toggle_file.go
│   ├── add_context.go
│   └── tui.go
├── internal/
│   ├── domain/            # Business logic
│   │   └── domain.go
│   ├── git/               # Git operations
│   │   └── git.go
│   ├── api/               # API client (mock)
│   │   └── api.go
│   ├── controller/        # Orchestration
│   │   ├── controller.go
│   │   └── session.go
│   ├── session/           # YAML persistence
│   │   └── session.go
│   └── tui/               # Bubbletea UI
│       ├── model.go
│       └── styles.go
├── test/                  # Test scripts
│   ├── setup-test-repo.sh
│   ├── test-session-cli.sh
│   └── test-all.sh
├── main.go
├── go.mod
└── README.md
```

### Running Tests

```bash
# Run all tests
./test/test-all.sh

# Test session management
./test/test-session-cli.sh

# Set up test repository
./test/setup-test-repo.sh
```

### Building

```bash
# Build binary
go build -o pr-builder .

# Build with specific version
go build -ldflags "-X main.version=1.0.0" -o pr-builder .

# Cross-compile for different platforms
GOOS=darwin GOARCH=amd64 go build -o pr-builder-darwin .
GOOS=linux GOARCH=amd64 go build -o pr-builder-linux .
GOOS=windows GOARCH=amd64 go build -o pr-builder.exe .
```

## Contributing

Contributions are welcome! Areas for improvement:

- Additional TUI screens (filter editor, context editor, prompt editor)
- Real LLM integration (OpenAI, Anthropic, etc.)
- More sophisticated file analysis
- Template system for PR descriptions
- Git hosting platform integration (GitHub, GitLab, Bitbucket)
- Diff visualization in TUI
- Undo/redo for session changes

## License

MIT License - see LICENSE file for details.

## Acknowledgments

Built with:
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [go-yaml](https://github.com/go-yaml/yaml) - YAML parsing
