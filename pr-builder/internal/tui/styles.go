package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	ColorPrimary   = lipgloss.Color("63")  // Blue
	ColorSecondary = lipgloss.Color("141") // Purple
	ColorSuccess   = lipgloss.Color("42")  // Green
	ColorWarning   = lipgloss.Color("220") // Yellow
	ColorError     = lipgloss.Color("196") // Red
	ColorMuted     = lipgloss.Color("240") // Gray
	ColorBorder    = lipgloss.Color("238") // Dark gray
	
	// Base styles
	BaseStyle = lipgloss.NewStyle().
			Padding(0, 1)
	
	// Title styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			Padding(0, 1)
	
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorSecondary).
			Padding(0, 1)
	
	// Border styles
	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(1, 2)
	
	BoxStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(ColorBorder).
		Padding(0, 1)
	
	// Status styles
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Padding(0, 1)
	
	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)
	
	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true)
	
	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorWarning).
			Bold(true)
	
	MutedStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)
	
	// List item styles
	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true).
				PaddingLeft(2)
	
	UnselectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(4)
	
	// Help styles
	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Padding(1, 0)
	
	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)
	
	// File stats styles
	AdditionStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess)
	
	DeletionStyle = lipgloss.NewStyle().
			Foreground(ColorError)
	
	TokenStyle = lipgloss.NewStyle().
			Foreground(ColorWarning)
)
