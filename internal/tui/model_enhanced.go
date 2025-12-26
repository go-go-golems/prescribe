package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/user/pr-builder/internal/controller"
	"github.com/user/pr-builder/internal/domain"
)

// Screen represents different TUI screens
type Screen int

const (
	ScreenMain Screen = iota
	ScreenFilters
	ScreenGenerating
	ScreenResult
)

// EnhancedModel is the main Bubbletea model with filter support
type EnhancedModel struct {
	controller      *controller.Controller
	width           int
	height          int
	currentScreen   Screen
	selectedIndex   int
	filterIndex     int
	err             error
	generatedDesc   string
	showFilteredFiles bool
}

// NewEnhancedModel creates a new enhanced TUI model
func NewEnhancedModel(ctrl *controller.Controller) *EnhancedModel {
	return &EnhancedModel{
		controller:    ctrl,
		currentScreen: ScreenMain,
		selectedIndex: 0,
		filterIndex:   0,
	}
}

// Init initializes the model
func (m *EnhancedModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m *EnhancedModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
		
	case tea.KeyMsg:
		switch m.currentScreen {
		case ScreenMain:
			return m.updateMain(msg)
		case ScreenFilters:
			return m.updateFilters(msg)
		case ScreenGenerating:
			// Ignore input while generating
			return m, nil
		case ScreenResult:
			return m.updateResult(msg)
		}
	
	case generateCompleteMsg:
		if msg.err != nil {
			m.err = msg.err
			m.currentScreen = ScreenMain
		} else {
			m.generatedDesc = msg.description
			m.currentScreen = ScreenResult
		}
		return m, nil
	}
	
	return m, nil
}

func (m *EnhancedModel) updateMain(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		
	case "f":
		// Switch to filter screen
		m.currentScreen = ScreenFilters
		m.filterIndex = 0
		
	case "v":
		// Toggle showing filtered files
		m.showFilteredFiles = !m.showFilteredFiles
		
	case "g":
		// Generate description
		m.currentScreen = ScreenGenerating
		return m, m.generateCmd()
	}
	
	return m, nil
}

func (m *EnhancedModel) updateFilters(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	filters := m.controller.GetFilters()
	
	switch msg.String() {
	case "ctrl+c", "q", "esc":
		// Return to main screen
		m.currentScreen = ScreenMain
		
	case "up", "k":
		if m.filterIndex > 0 {
			m.filterIndex--
		}
		
	case "down", "j":
		if len(filters) > 0 && m.filterIndex < len(filters)-1 {
			m.filterIndex++
		}
		
	case "d", "x":
		// Delete selected filter
		if len(filters) > 0 && m.filterIndex < len(filters) {
			m.controller.RemoveFilter(m.filterIndex)
			m.controller.SaveSession(m.controller.GetDefaultSessionPath())
			if m.filterIndex >= len(m.controller.GetFilters()) {
				m.filterIndex = max(0, len(m.controller.GetFilters())-1)
			}
		}
		
	case "c":
		// Clear all filters
		m.controller.ClearFilters()
		m.controller.SaveSession(m.controller.GetDefaultSessionPath())
		m.filterIndex = 0
		
	case "1":
		// Add preset filter: Exclude tests
		filter := domain.Filter{
			Name:        "Exclude Tests",
			Description: "Exclude test files",
			Rules: []domain.FilterRule{
				{Type: domain.FilterTypeExclude, Pattern: "**/*test*"},
				{Type: domain.FilterTypeExclude, Pattern: "**/*spec*"},
			},
		}
		m.controller.AddFilter(filter)
		m.controller.SaveSession(m.controller.GetDefaultSessionPath())
		
	case "2":
		// Add preset filter: Exclude docs
		filter := domain.Filter{
			Name:        "Exclude Docs",
			Description: "Exclude documentation files",
			Rules: []domain.FilterRule{
				{Type: domain.FilterTypeExclude, Pattern: "**/*.md"},
				{Type: domain.FilterTypeExclude, Pattern: "**/docs/**"},
			},
		}
		m.controller.AddFilter(filter)
		m.controller.SaveSession(m.controller.GetDefaultSessionPath())
		
	case "3":
		// Add preset filter: Only source
		filter := domain.Filter{
			Name:        "Only Source",
			Description: "Include only source code files",
			Rules: []domain.FilterRule{
				{Type: domain.FilterTypeInclude, Pattern: "**/*.go"},
				{Type: domain.FilterTypeInclude, Pattern: "**/*.ts"},
				{Type: domain.FilterTypeInclude, Pattern: "**/*.js"},
				{Type: domain.FilterTypeInclude, Pattern: "**/*.py"},
			},
		}
		m.controller.AddFilter(filter)
		m.controller.SaveSession(m.controller.GetDefaultSessionPath())
	}
	
	return m, nil
}

func (m *EnhancedModel) updateResult(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "esc":
		m.currentScreen = ScreenMain
	}
	return m, nil
}

// View renders the current screen
func (m *EnhancedModel) View() string {
	if m.err != nil {
		return ErrorStyle.Render(fmt.Sprintf("Error: %v\n\nPress q to quit", m.err))
	}
	
	switch m.currentScreen {
	case ScreenMain:
		return m.renderMain()
	case ScreenFilters:
		return m.renderFilters()
	case ScreenGenerating:
		return m.renderGenerating()
	case ScreenResult:
		return m.renderResult()
	}
	
	return ""
}

func (m *EnhancedModel) renderMain() string {
	data := m.controller.GetData()
	visibleFiles := data.GetVisibleFiles()
	filteredFiles := data.GetFilteredFiles()
	
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
	stats := fmt.Sprintf("Files: %d visible, %d filtered | Tokens: %s | Filters: %d",
		len(visibleFiles),
		len(filteredFiles),
		TokenStyle.Render(fmt.Sprintf("%d", data.GetTotalTokens())),
		len(data.ActiveFilters))
	b.WriteString(BaseStyle.Render(stats))
	b.WriteString("\n\n")
	
	// Files list
	if m.showFilteredFiles && len(filteredFiles) > 0 {
		b.WriteString(HeaderStyle.Render("FILTERED FILES"))
	} else {
		b.WriteString(HeaderStyle.Render("CHANGED FILES"))
	}
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", 78))
	b.WriteString("\n")
	
	filesToShow := visibleFiles
	if m.showFilteredFiles {
		filesToShow = filteredFiles
	}
	
	for i, file := range filesToShow {
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
		HelpKeyStyle.Render("[F]") + " Filters",
		HelpKeyStyle.Render("[V]") + " View Filtered",
		HelpKeyStyle.Render("[G]") + " Generate",
		HelpKeyStyle.Render("[Q]") + " Quit",
	}
	b.WriteString(HelpStyle.Render(strings.Join(helpText, "  ")))
	b.WriteString("\n")
	
	return BorderStyle.Render(b.String())
}

func (m *EnhancedModel) renderFilters() string {
	filters := m.controller.GetFilters()
	data := m.controller.GetData()
	
	var b strings.Builder
	
	// Title
	title := TitleStyle.Render("FILTER MANAGEMENT")
	b.WriteString(lipgloss.PlaceHorizontal(80, lipgloss.Center, title))
	b.WriteString("\n\n")
	
	// Stats
	stats := fmt.Sprintf("Active Filters: %d | Filtered Files: %d",
		len(filters),
		len(data.GetFilteredFiles()))
	b.WriteString(BaseStyle.Render(stats))
	b.WriteString("\n\n")
	
	// Filters list
	b.WriteString(HeaderStyle.Render("ACTIVE FILTERS"))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", 78))
	b.WriteString("\n")
	
	if len(filters) == 0 {
		b.WriteString(MutedStyle.Render("No active filters"))
		b.WriteString("\n")
	} else {
		for i, filter := range filters {
			filterLine := fmt.Sprintf("[%d] %s", i, filter.Name)
			if filter.Description != "" {
				filterLine += fmt.Sprintf(" - %s", filter.Description)
			}
			
			if i == m.filterIndex {
				b.WriteString(SelectedItemStyle.Render("▶ " + filterLine))
			} else {
				b.WriteString(UnselectedItemStyle.Render(filterLine))
			}
			b.WriteString("\n")
			
			// Show rules for selected filter
			if i == m.filterIndex {
				for _, rule := range filter.Rules {
					ruleText := fmt.Sprintf("    %s: %s", rule.Type, rule.Pattern)
					b.WriteString(MutedStyle.Render(ruleText))
					b.WriteString("\n")
				}
			}
		}
	}
	
	// Presets
	b.WriteString("\n")
	b.WriteString(HeaderStyle.Render("QUICK ADD PRESETS"))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", 78))
	b.WriteString("\n")
	b.WriteString(BaseStyle.Render("[1] Exclude Tests  [2] Exclude Docs  [3] Only Source"))
	b.WriteString("\n")
	
	// Help
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", 78))
	b.WriteString("\n")
	
	helpText := []string{
		HelpKeyStyle.Render("[↑↓/jk]") + " Navigate",
		HelpKeyStyle.Render("[D/X]") + " Delete",
		HelpKeyStyle.Render("[C]") + " Clear All",
		HelpKeyStyle.Render("[1-3]") + " Add Preset",
		HelpKeyStyle.Render("[Esc]") + " Back",
	}
	b.WriteString(HelpStyle.Render(strings.Join(helpText, "  ")))
	b.WriteString("\n")
	
	return BorderStyle.Render(b.String())
}

func (m *EnhancedModel) renderGenerating() string {
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

func (m *EnhancedModel) renderResult() string {
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
func (m *EnhancedModel) generateCmd() tea.Cmd {
	return func() tea.Msg {
		description, err := m.controller.GenerateDescription()
		return generateCompleteMsg{
			description: description,
			err:         err,
		}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
