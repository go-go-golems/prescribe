package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/user/pr-builder/internal/controller"
)

// Model is the main Bubbletea model
type Model struct {
	controller     *controller.Controller
	width          int
	height         int
	selectedIndex  int
	err            error
	generating     bool
	generated      bool
	generatedDesc  string
}

// NewModel creates a new TUI model
func NewModel(ctrl *controller.Controller) *Model {
	return &Model{
		controller:    ctrl,
		selectedIndex: 0,
	}
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
		
	case tea.KeyMsg:
		// If generating, ignore input
		if m.generating {
			return m, nil
		}
		
		// If showing generated result
		if m.generated {
			switch msg.String() {
			case "q", "esc":
				m.generated = false
				return m, nil
			}
			return m, nil
		}
		
		// Normal navigation
		data := m.controller.GetData()
		visibleFiles := data.GetVisibleFiles()
		
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
			
		case "up", "k":
			if m.selectedIndex > 0 {
				m.selectedIndex--
			}
			
		case "down", "j":
			if m.selectedIndex < len(visibleFiles)-1 {
				m.selectedIndex++
			}
			
		case " ":
			// Toggle file inclusion
			if m.selectedIndex < len(visibleFiles) {
				selectedFile := visibleFiles[m.selectedIndex]
				for i, f := range data.ChangedFiles {
					if f.Path == selectedFile.Path {
						m.controller.ToggleFileInclusion(i)
						// Auto-save session
						m.controller.SaveSession(m.controller.GetDefaultSessionPath())
						break
					}
				}
			}
			
		case "g":
			// Generate description
			m.generating = true
			return m, m.generateCmd()
		}
	
	case generateCompleteMsg:
		m.generating = false
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.generated = true
			m.generatedDesc = msg.description
		}
		return m, nil
	}
	
	return m, nil
}

// View renders the current screen
func (m *Model) View() string {
	if m.err != nil {
		return ErrorStyle.Render(fmt.Sprintf("Error: %v\n\nPress q to quit", m.err))
	}
	
	if m.generating {
		return m.renderGenerating()
	}
	
	if m.generated {
		return m.renderResult()
	}
	
	return m.renderMain()
}

func (m *Model) renderMain() string {
	data := m.controller.GetData()
	visibleFiles := data.GetVisibleFiles()
	
	var b strings.Builder
	
	// Title
	title := TitleStyle.Render("PR DESCRIPTION GENERATOR")
	b.WriteString(lipgloss.PlaceHorizontal(80, lipgloss.Center, title))
	b.WriteString("\n\n")
	
	// Branch info
	branchInfo := fmt.Sprintf("%s → %s",
		SuccessStyle.Render(data.SourceBranch),
		SuccessStyle.Render(data.TargetBranch))
	b.WriteString(BaseStyle.Render(branchInfo))
	b.WriteString("\n\n")
	
	// Stats
	stats := fmt.Sprintf("Files: %d | Tokens: %s",
		len(visibleFiles),
		TokenStyle.Render(fmt.Sprintf("%d", data.GetTotalTokens())))
	b.WriteString(BaseStyle.Render(stats))
	b.WriteString("\n\n")
	
	// Files list
	b.WriteString(HeaderStyle.Render("CHANGED FILES"))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", 78))
	b.WriteString("\n")
	
	for i, file := range visibleFiles {
		included := " "
		if file.Included {
			included = "✓"
		}
		
		fileLine := fmt.Sprintf("[%s] %-45s +%-3d -%-3d (%dt)",
			included,
			file.Path,
			file.Additions,
			file.Deletions,
			file.Tokens)
		
		if i == m.selectedIndex {
			b.WriteString(SelectedItemStyle.Render("▶ " + fileLine))
		} else {
			b.WriteString(UnselectedItemStyle.Render(fileLine))
		}
		b.WriteString("\n")
	}
	
	// Help
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", 78))
	b.WriteString("\n")
	
	helpText := []string{
		HelpKeyStyle.Render("[↑↓/jk]") + " Navigate",
		HelpKeyStyle.Render("[Space]") + " Toggle",
		HelpKeyStyle.Render("[G]") + " Generate",
		HelpKeyStyle.Render("[Q]") + " Quit",
	}
	b.WriteString(HelpStyle.Render(strings.Join(helpText, "  ")))
	b.WriteString("\n")
	
	return BorderStyle.Render(b.String())
}

func (m *Model) renderGenerating() string {
	var b strings.Builder
	
	title := TitleStyle.Render("GENERATING PR DESCRIPTION")
	b.WriteString(lipgloss.PlaceHorizontal(80, lipgloss.Center, title))
	b.WriteString("\n\n")
	
	b.WriteString(BaseStyle.Render("Analyzing changes and generating description..."))
	b.WriteString("\n\n")
	
	spinner := "⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏"
	b.WriteString(lipgloss.PlaceHorizontal(80, lipgloss.Center, 
		SuccessStyle.Render(string(spinner[0]))))
	b.WriteString("\n\n")
	
	b.WriteString(MutedStyle.Render("This may take a few seconds..."))
	b.WriteString("\n")
	
	return BorderStyle.Render(b.String())
}

func (m *Model) renderResult() string {
	var b strings.Builder
	
	title := TitleStyle.Render("GENERATED PR DESCRIPTION")
	b.WriteString(lipgloss.PlaceHorizontal(80, lipgloss.Center, title))
	b.WriteString("\n\n")
	
	b.WriteString(SuccessStyle.Render("✓ Description generated successfully!"))
	b.WriteString("\n\n")
	
	// Description box
	descBox := BoxStyle.Width(74).Render(m.generatedDesc)
	b.WriteString(descBox)
	b.WriteString("\n\n")
	
	// Help
	helpText := []string{
		HelpKeyStyle.Render("[Esc]") + " Back",
		HelpKeyStyle.Render("[Q]") + " Quit",
	}
	b.WriteString(HelpStyle.Render(strings.Join(helpText, "  ")))
	b.WriteString("\n")
	
	return BorderStyle.Render(b.String())
}

// generateCmd is a command that generates the PR description
func (m *Model) generateCmd() tea.Cmd {
	return func() tea.Msg {
		description, err := m.controller.GenerateDescription()
		return generateCompleteMsg{
			description: description,
			err:         err,
		}
	}
}

// Messages

type generateCompleteMsg struct {
	description string
	err         error
}
