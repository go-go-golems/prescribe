package styles

import "github.com/charmbracelet/lipgloss"

// Styles groups all lipgloss styles used by the new modular TUI.
//
// Goal: stop expanding global style vars in internal/tui/styles.go (package tui).
// The modular TUI will pass a Styles value down into components.
type Styles struct {
	// Colors
	Primary   lipgloss.Color
	Secondary lipgloss.Color
	Success   lipgloss.Color
	Warning   lipgloss.Color
	Error     lipgloss.Color
	Muted     lipgloss.Color
	Border    lipgloss.Color

	// Base styles
	Base      lipgloss.Style
	BorderBox lipgloss.Style
	Box       lipgloss.Style

	// Text styles
	Title       lipgloss.Style
	Header      lipgloss.Style
	SuccessText lipgloss.Style
	WarningText lipgloss.Style
	ErrorText   lipgloss.Style
	MutedText   lipgloss.Style

	// List styles
	SelectedItem   lipgloss.Style
	UnselectedItem lipgloss.Style

	// Help/status
	Help    lipgloss.Style
	HelpKey lipgloss.Style
}

// Default returns the initial style palette. This is intentionally close to the current
// internal/tui/styles.go look so the Phase 2 switch is behavior-compatible.
func Default() Styles {
	s := Styles{
		Primary:   lipgloss.Color("63"),
		Secondary: lipgloss.Color("141"),
		Success:   lipgloss.Color("42"),
		Warning:   lipgloss.Color("220"),
		Error:     lipgloss.Color("196"),
		Muted:     lipgloss.Color("240"),
		Border:    lipgloss.Color("238"),
	}

	// Keep Base style width-neutral so it doesn't push content wider than the frame.
	s.Base = lipgloss.NewStyle()

	s.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(s.Primary).
		Padding(0, 1)

	s.Header = lipgloss.NewStyle().
		Bold(true).
		Foreground(s.Secondary).
		Padding(0, 1)

	s.BorderBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true).
		BorderForeground(s.Border).
		Padding(1, 2)

	s.Box = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true).
		BorderForeground(s.Border).
		Padding(0, 1)

	s.SuccessText = lipgloss.NewStyle().Foreground(s.Success).Bold(true)
	s.WarningText = lipgloss.NewStyle().Foreground(s.Warning).Bold(true)
	s.ErrorText = lipgloss.NewStyle().Foreground(s.Error).Bold(true)
	s.MutedText = lipgloss.NewStyle().Foreground(s.Muted)

	s.SelectedItem = lipgloss.NewStyle().
		Foreground(s.Primary).
		Bold(true).
		PaddingLeft(2)

	s.UnselectedItem = lipgloss.NewStyle().
		PaddingLeft(4)

	s.Help = lipgloss.NewStyle().
		Foreground(s.Muted).
		Padding(1, 0)

	s.HelpKey = lipgloss.NewStyle().
		Foreground(s.Primary).
		Bold(true)

	return s
}
