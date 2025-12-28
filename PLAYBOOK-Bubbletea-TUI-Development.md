# Playbook: Building CLI/TUI Applications with Go and Bubbletea

This playbook documents the methodology for building professional CLI/TUI applications using Go, Cobra, and Bubbletea, based on the PR Builder project.

## Overview

This methodology creates applications with dual interfaces: a scriptable CLI for automation and an interactive TUI for manual use. Both interfaces operate on the same core logic, ensuring consistency and maintainability.

## Architecture Principles

### 1. Session-First Design

The core insight is to build around a **session configuration file** rather than just imperative commands. This provides:

- **Declarative configuration**: Describe what you want, not how to get there
- **Reproducibility**: Same session = same result
- **Shareability**: Team members can use the same configurations
- **Versionability**: Track configuration changes over time
- **Scriptability**: Easy to generate and modify programmatically

**Implementation**:
- Use YAML for human-readable configuration
- Store all application state in the session
- Make session the single source of truth
- Provide load/save commands
- Auto-save on changes in TUI

### 2. Layered Architecture

Build in layers from core to presentation:

```
┌─────────────────────────────────┐
│     Presentation Layer          │
│  ┌──────────┐    ┌──────────┐  │
│  │   CLI    │    │   TUI    │  │
│  │ (Cobra)  │    │(Bubbletea)│  │
│  └──────────┘    └──────────┘  │
├─────────────────────────────────┤
│     Controller Layer            │
│  (Orchestration & Coordination) │
├─────────────────────────────────┤
│     Session Layer               │
│  (Persistence & Serialization)  │
├─────────────────────────────────┤
│     Domain Layer                │
│  (Business Logic & Data Models) │
├─────────────────────────────────┤
│     Service Layer               │
│  (External Interactions)        │
└─────────────────────────────────┘
```

**Benefits**:
- Clear separation of concerns
- Easy to test each layer independently
- CLI and TUI share the same core
- Can add new interfaces without changing core

### 3. Test-First Development

Build and test the core functionality before adding UI:

1. **Domain layer**: Pure business logic, no dependencies
2. **Service layer**: Mock external services for testing
3. **Controller layer**: Orchestration logic
4. **CLI commands**: Test with scripts
5. **Session system**: Test save/load round-trips
6. **TUI**: Add last as a thin presentation layer

**Why this order**:
- Catch logic errors early
- Faster iteration (no UI overhead)
- Better architecture (UI doesn't drive design)
- Easier debugging
- More maintainable code

## Implementation Steps

### Step 1: Project Setup

#### 1.1 Install Go

```bash
# Download latest Go (1.25+)
wget https://go.dev/dl/go1.25.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.25.5.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

#### 1.2 Initialize Project

```bash
mkdir myapp
cd myapp
go mod init github.com/user/myapp

# Install dependencies
go get github.com/spf13/cobra@latest
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/lipgloss@latest
go get gopkg.in/yaml.v3
```

#### 1.3 Create Directory Structure

```bash
mkdir -p cmd internal/{domain,controller,session,tui} test
```

### Step 2: Domain Layer

#### 2.1 Define Data Models

Create `internal/domain/domain.go`:

```go
package domain

// Core data model
type AppData struct {
    // Your application state
    Items []Item
    Filters []Filter
    Config Config
}

type Item struct {
    ID string
    Name string
    // ... fields
}

// Methods on AppData
func (d *AppData) GetVisibleItems() []Item {
    // Apply filters, return visible items
}
```

**Key principles**:
- Pure data structures
- No external dependencies
- Business logic as methods
- Immutable where possible

#### 2.2 Implement Business Logic

Add methods for all operations:
- Filtering
- Sorting
- Validation
- Calculations
- Transformations

**Example**:
```go
func (d *AppData) AddFilter(f Filter) {
    d.Filters = append(d.Filters, f)
}

func (d *AppData) ApplyFilters() {
    // Filter logic
}
```

### Step 3: Service Layer

#### 3.1 Define Service Interfaces

Create `internal/service/interface.go`:

```go
package service

type DataService interface {
    Load() ([]Item, error)
    Save(items []Item) error
}

type APIService interface {
    Process(data string) (string, error)
}
```

#### 3.2 Implement Services

**Real implementation**:
```go
type FileService struct {
    path string
}

func (s *FileService) Load() ([]Item, error) {
    // Read from file
}
```

**Mock for testing**:
```go
type MockService struct {
    data []Item
}

func (s *MockService) Load() ([]Item, error) {
    return s.data, nil
}
```

### Step 4: Session System

#### 4.1 Define Session Structure

Create `internal/session/session.go`:

```go
package session

type Session struct {
    Version string `yaml:"version"`
    
    // Your configuration
    Items []ItemConfig `yaml:"items"`
    Filters []FilterConfig `yaml:"filters"`
    Settings SettingsConfig `yaml:"settings"`
}

type ItemConfig struct {
    ID string `yaml:"id"`
    Enabled bool `yaml:"enabled"`
}
```

#### 4.2 Implement Serialization

```go
func NewSession(data *domain.AppData) *Session {
    // Convert domain model to session
}

func (s *Session) ApplyToData(data *domain.AppData) error {
    // Apply session to domain model
}

func (s *Session) Save(path string) error {
    data, err := yaml.Marshal(s)
    if err != nil {
        return err
    }
    return os.WriteFile(path, data, 0644)
}

func Load(path string) (*Session, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    var session Session
    err = yaml.Unmarshal(data, &session)
    return &session, err
}
```

### Step 5: Controller Layer

#### 5.1 Create Controller

Create `internal/controller/controller.go`:

```go
package controller

type Controller struct {
    data *domain.AppData
    services map[string]interface{}
    // ... other state
}

func NewController() *Controller {
    return &Controller{
        data: &domain.AppData{},
        services: make(map[string]interface{}),
    }
}

func (c *Controller) Initialize() error {
    // Load data, initialize state
}

func (c *Controller) GetData() *domain.AppData {
    return c.data
}

// Operation methods
func (c *Controller) AddItem(item domain.Item) error {
    // Validate, add, update state
}

func (c *Controller) RemoveItem(id string) error {
    // Find, remove, update state
}
```

#### 5.2 Add Session Methods

```go
func (c *Controller) SaveSession(path string) error {
    sess := session.NewSession(c.data)
    return sess.Save(path)
}

func (c *Controller) LoadSession(path string) error {
    sess, err := session.Load(path)
    if err != nil {
        return err
    }
    return sess.ApplyToData(c.data)
}
```

### Step 6: CLI Commands

#### 6.1 Create Root Command

Create `cmd/root.go`:

```go
package cmd

import (
    "github.com/spf13/cobra"
)

var (
    configPath string
)

var rootCmd = &cobra.Command{
    Use:   "myapp",
    Short: "My application",
    Long:  `Detailed description of my application.`,
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}

func init() {
    rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", ".", "Config path")
}
```

#### 6.2 Create Commands

**Init command** (`cmd/init.go`):
```go
var initCmd = &cobra.Command{
    Use:   "init",
    Short: "Initialize new session",
    RunE: func(cmd *cobra.Command, args []string) error {
        ctrl := controller.NewController()
        if err := ctrl.Initialize(); err != nil {
            return err
        }
        return ctrl.SaveSession(getDefaultPath())
    },
}

func init() {
    rootCmd.AddCommand(initCmd)
}
```

**Show command** (`cmd/show.go`):
```go
var showCmd = &cobra.Command{
    Use:   "show",
    Short: "Show current state",
    RunE: func(cmd *cobra.Command, args []string) error {
        ctrl := controller.NewController()
        ctrl.Initialize()
        data := ctrl.GetData()
        // Print data
        return nil
    },
}
```

**Session commands**:
- `save [path]`: Save session
- `load [path]`: Load session
- `session show [--output yaml]`: Display state (structured output via Glazed)

**Operation commands**:
- `add <item>`: Add item
- `remove <id>`: Remove item
- `list`: List items
- `filter <pattern>`: Add filter

### Step 7: Test Scripts

#### 7.1 Create Test Repository/Data

Create `test/setup-test-data.sh`:

```bash
#!/bin/bash
set -e

# Create test data
mkdir -p /tmp/myapp-test
cd /tmp/myapp-test

# Initialize with test data
# ...

echo "Test data created"
```

#### 7.2 Create Comprehensive Test Script

Create `test/test-all.sh`:

```bash
#!/bin/bash
set -e

APP="/path/to/myapp"

echo "=== Test 1: Initialize ==="
$APP init --save
echo "✓ Init works"

echo "=== Test 2: Show state ==="
$APP show
echo "✓ Show works"

echo "=== Test 3: Add item ==="
$APP add "test-item"
echo "✓ Add works"

echo "=== Test 4: Save/Load ==="
$APP save /tmp/session.yaml
$APP init  # Reset
$APP load /tmp/session.yaml
echo "✓ Session persistence works"

# ... more tests

echo "✓ All tests passed!"
```

#### 7.3 Run Tests

```bash
chmod +x test/*.sh
./test/test-all.sh
```

### Step 8: Bubbletea TUI

#### 8.1 Create Styles

Create `internal/tui/styles.go`:

```go
package tui

import "github.com/charmbracelet/lipgloss"

var (
    ColorPrimary = lipgloss.Color("63")
    ColorSuccess = lipgloss.Color("42")
    ColorError = lipgloss.Color("196")
    
    TitleStyle = lipgloss.NewStyle().
        Bold(true).
        Foreground(ColorPrimary).
        Padding(0, 1)
    
    SelectedStyle = lipgloss.NewStyle().
        Foreground(ColorPrimary).
        Bold(true).
        PaddingLeft(2)
    
    // ... more styles
)
```

#### 8.2 Create Model

Create `internal/tui/model.go`:

```go
package tui

import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/user/myapp/internal/controller"
)

type Model struct {
    controller *controller.Controller
    selectedIndex int
    width int
    height int
    // ... state
}

func NewModel(ctrl *controller.Controller) *Model {
    return &Model{
        controller: ctrl,
        selectedIndex: 0,
    }
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
    return nil
}

// Update handles messages
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", "q":
            return m, tea.Quit
        case "up", "k":
            if m.selectedIndex > 0 {
                m.selectedIndex--
            }
        case "down", "j":
            m.selectedIndex++
        case " ":
            // Toggle selection
        case "enter":
            // Perform action
        }
    }
    
    return m, nil
}

// View renders the UI
func (m *Model) View() string {
    var b strings.Builder
    
    // Title
    b.WriteString(TitleStyle.Render("MY APPLICATION"))
    b.WriteString("\n\n")
    
    // Content
    data := m.controller.GetData()
    for i, item := range data.Items {
        if i == m.selectedIndex {
            b.WriteString(SelectedStyle.Render("▶ " + item.Name))
        } else {
            b.WriteString("  " + item.Name)
        }
        b.WriteString("\n")
    }
    
    // Help
    b.WriteString("\n[↑↓] Navigate  [Space] Select  [Q] Quit\n")
    
    return b.String()
}
```

#### 8.3 Add TUI Command

Create `cmd/tui.go`:

```go
var tuiCmd = &cobra.Command{
    Use:   "tui",
    Short: "Launch interactive TUI",
    RunE: func(cmd *cobra.Command, args []string) error {
        ctrl := controller.NewController()
        if err := ctrl.Initialize(); err != nil {
            return err
        }
        
        // Try to load existing session
        _ = ctrl.LoadSession(getDefaultPath())
        
        // Run TUI
        p := tea.NewProgram(tui.NewModel(ctrl), tea.WithAltScreen())
        _, err := p.Run()
        return err
    },
}

func init() {
    rootCmd.AddCommand(tuiCmd)
}
```

### Step 9: Advanced TUI Patterns

#### 9.1 Multiple Screens

Use a state machine:

```go
type Screen int

const (
    ScreenMain Screen = iota
    ScreenEdit
    ScreenResult
)

type Model struct {
    screen Screen
    // ... sub-models
}

func (m *Model) View() string {
    switch m.screen {
    case ScreenMain:
        return m.renderMain()
    case ScreenEdit:
        return m.renderEdit()
    case ScreenResult:
        return m.renderResult()
    }
    return ""
}
```

#### 9.2 Async Operations

Use commands and custom messages:

```go
type processCompleteMsg struct {
    result string
    err error
}

func (m *Model) processCmd() tea.Cmd {
    return func() tea.Msg {
        result, err := m.controller.Process()
        return processCompleteMsg{result, err}
    }
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case processCompleteMsg:
        if msg.err != nil {
            m.err = msg.err
        } else {
            m.result = msg.result
        }
        return m, nil
    }
    // ...
}
```

#### 9.3 Using Bubbles

Integrate reusable components:

```go
import "github.com/charmbracelet/bubbles/list"

type Model struct {
    list list.Model
    // ...
}

func NewModel() *Model {
    items := []list.Item{
        // ... your items
    }
    l := list.New(items, list.NewDefaultDelegate(), 0, 0)
    return &Model{list: l}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd
    m.list, cmd = m.list.Update(msg)
    return m, cmd
}

func (m *Model) View() string {
    return m.list.View()
}
```

## Best Practices

### Code Organization

**Do**:
- One file per command
- One file per screen/model
- Group related functionality
- Use meaningful package names

**Don't**:
- Mix presentation and business logic
- Put everything in main.go
- Create circular dependencies

### Error Handling

**CLI**:
```go
if err != nil {
    return fmt.Errorf("failed to load session: %w", err)
}
```

**TUI**:
```go
type Model struct {
    err error
}

func (m *Model) View() string {
    if m.err != nil {
        return ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
    }
    // ... normal view
}
```

### Session Management

**Auto-save in TUI**:
```go
case " ":
    // Toggle item
    m.controller.ToggleItem(m.selectedIndex)
    // Auto-save
    m.controller.SaveSession(getDefaultPath())
```

**Validation**:
```go
func (s *Session) Validate() error {
    if s.Version != "1.0" {
        return fmt.Errorf("unsupported version: %s", s.Version)
    }
    // ... more validation
    return nil
}
```

### Testing

**Unit tests for domain**:
```go
func TestAddItem(t *testing.T) {
    data := &domain.AppData{}
    item := domain.Item{ID: "1", Name: "Test"}
    data.AddItem(item)
    
    if len(data.Items) != 1 {
        t.Errorf("expected 1 item, got %d", len(data.Items))
    }
}
```

**Integration tests with scripts**:
```bash
# Test full workflow
$APP init
$APP add "item1"
$APP show | grep "item1" || exit 1
echo "✓ Add item works"
```

## Common Patterns

### Pattern 1: List with Selection

```go
type Model struct {
    items []Item
    selected int
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "up", "k":
            if m.selected > 0 {
                m.selected--
            }
        case "down", "j":
            if m.selected < len(m.items)-1 {
                m.selected++
            }
        }
    }
    return m, nil
}

func (m *Model) View() string {
    var b strings.Builder
    for i, item := range m.items {
        if i == m.selected {
            b.WriteString("▶ " + item.Name)
        } else {
            b.WriteString("  " + item.Name)
        }
        b.WriteString("\n")
    }
    return b.String()
}
```

### Pattern 2: Loading State

```go
type Model struct {
    loading bool
    result string
}

func (m *Model) View() string {
    if m.loading {
        return "Loading..."
    }
    return m.result
}

func (m *Model) loadCmd() tea.Cmd {
    return func() tea.Msg {
        // Do work
        return loadCompleteMsg{result}
    }
}
```

### Pattern 3: Form Input

```go
import "github.com/charmbracelet/bubbles/textinput"

type Model struct {
    input textinput.Model
}

func NewModel() *Model {
    ti := textinput.New()
    ti.Placeholder = "Enter value"
    ti.Focus()
    return &Model{input: ti}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd
    m.input, cmd = m.input.Update(msg)
    
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.String() == "enter" {
            // Process input
            value := m.input.Value()
        }
    }
    
    return m, cmd
}
```

## Troubleshooting

### Issue: TUI flickers

**Solution**: Use `tea.WithAltScreen()`:
```go
p := tea.NewProgram(model, tea.WithAltScreen())
```

### Issue: Colors don't show

**Solution**: Check terminal support:
```go
import "github.com/muesli/termenv"

if termenv.ColorProfile() == termenv.Ascii {
    // Disable colors
}
```

### Issue: Session not persisting

**Solution**: Ensure directory exists:
```go
func (s *Session) Save(path string) error {
    dir := filepath.Dir(path)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return err
    }
    // ... save
}
```

## Deployment

### Build for Multiple Platforms

```bash
#!/bin/bash
# build.sh

VERSION="1.0.0"
LDFLAGS="-X main.version=$VERSION"

# Linux
GOOS=linux GOARCH=amd64 go build -ldflags "$LDFLAGS" -o dist/myapp-linux-amd64

# macOS
GOOS=darwin GOARCH=amd64 go build -ldflags "$LDFLAGS" -o dist/myapp-darwin-amd64
GOOS=darwin GOARCH=arm64 go build -ldflags "$LDFLAGS" -o dist/myapp-darwin-arm64

# Windows
GOOS=windows GOARCH=amd64 go build -ldflags "$LDFLAGS" -o dist/myapp-windows-amd64.exe
```

### Installation Script

```bash
#!/bin/bash
# install.sh

INSTALL_DIR="/usr/local/bin"
BINARY="myapp"

# Download latest release
curl -L "https://github.com/user/myapp/releases/latest/download/myapp-$(uname -s)-$(uname -m)" -o "$BINARY"

# Make executable
chmod +x "$BINARY"

# Move to install directory
sudo mv "$BINARY" "$INSTALL_DIR/"

echo "✓ Installed to $INSTALL_DIR/$BINARY"
```

## Summary

This playbook provides a complete methodology for building professional CLI/TUI applications with Go and Bubbletea. The key insights are:

1. **Session-first design**: Build around a configuration file
2. **Layered architecture**: Separate concerns cleanly
3. **Test-first development**: Validate core before adding UI
4. **Dual interfaces**: CLI for automation, TUI for interaction
5. **Bubbletea patterns**: Use MVC with message passing

Following this methodology results in applications that are:
- Maintainable (clear architecture)
- Testable (separated layers)
- Scriptable (CLI commands)
- User-friendly (interactive TUI)
- Reproducible (session files)
- Professional (polished UX)

## References

- [Bubbletea Documentation](https://github.com/charmbracelet/bubbletea)
- [Cobra Documentation](https://github.com/spf13/cobra)
- [Lipgloss Documentation](https://github.com/charmbracelet/lipgloss)
- [Bubbles Components](https://github.com/charmbracelet/bubbles)
- [Example: PR Builder](https://github.com/user/pr-builder)
