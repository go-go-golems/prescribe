# PR Builder

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
cd pr-builder
go build -o pr-builder .
```

## Quick Start

### Initialize a Session

```bash
# Navigate to your git repository
cd /path/to/your/repo

# Initialize PR builder session
pr-builder init --save

# This creates .pr-builder/session.yaml with your current changes
```

### View Session State

```bash
# Human-readable format
pr-builder show

# YAML format
pr-builder show --yaml
```

### Generate PR Description

```bash
# Generate with default settings
pr-builder generate

# Generate with custom prompt
pr-builder generate --prompt "Write a concise 3-sentence description"

# Generate with preset
pr-builder generate --preset concise

# Save to file
pr-builder generate -o pr-description.md
```

### Launch Interactive TUI

```bash
pr-builder tui
```

## CLI Commands

### Session Management

#### `init`
Initialize a new session from current git state.

```bash
pr-builder init [--save] [--output PATH]
```

Options:
- `--save, -s`: Automatically save session after init
- `--output, -o PATH`: Custom session file path

#### `save`
Save current session to YAML file.

```bash
pr-builder save [PATH]
```

Default path: `.pr-builder/session.yaml`

#### `load`
Load session from YAML file.

```bash
pr-builder load [PATH]
```

#### `show`
Display current session state.

```bash
pr-builder show [--yaml]
```

Options:
- `--yaml, -y`: Output as YAML instead of human-readable format

### File Management

#### `toggle-file`
Toggle whether a file is included in the context.

```bash
pr-builder toggle-file <file-path>
```

Example:
```bash
pr-builder toggle-file src/auth/login.ts
```

### Filters

#### `add-filter`
Add a glob-based filter to exclude or include files.

```bash
pr-builder add-filter --name NAME [--description DESC] [--exclude PATTERN]... [--include PATTERN]...
```

Options:
- `--name, -n NAME`: Filter name (required)
- `--description, -d DESC`: Filter description
- `--exclude, -e PATTERN`: Exclude pattern (can specify multiple)
- `--include, -i PATTERN`: Include pattern (can specify multiple)

Examples:
```bash
# Exclude all test files
pr-builder add-filter --name "Exclude tests" --exclude "*test*" --exclude "*spec*"

# Include only TypeScript files
pr-builder add-filter --name "Only TS" --include "*.ts"

# Complex filter
pr-builder add-filter \
  --name "Backend only" \
  --description "Only backend code" \
  --include "src/backend/**" \
  --exclude "*test*"
```

### Context

#### `add-context`
Add additional context for PR generation.

```bash
# Add a file
pr-builder add-context <file-path>

# Add a note
pr-builder add-context --note "This PR is part of the Q1 security improvements"
```

### Generation

#### `generate`
Generate PR description using AI.

```bash
pr-builder generate [--output PATH] [--prompt TEXT] [--preset ID] [--session PATH]
```

Options:
- `--output, -o PATH`: Output file (default: stdout)
- `--prompt, -p TEXT`: Custom prompt text
- `--preset ID`: Prompt preset ID (detailed, concise, technical)
- `--session, -s PATH`: Load session file before generating

Examples:
```bash
# Generate with default settings
pr-builder generate

# Generate with custom prompt
pr-builder generate --prompt "Write a technical PR description focusing on architecture changes"

# Generate using a specific session
pr-builder generate --session /path/to/session.yaml -o pr.md

# Generate with preset
pr-builder generate --preset concise
```

### TUI

#### `tui`
Launch interactive Terminal User Interface.

```bash
pr-builder tui
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
