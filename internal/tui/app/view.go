package app

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) view() string {
	var b strings.Builder

	title := m.styles.Title.Render("PRESCRIBE (modular TUI - WIP)")
	b.WriteString(lipgloss.PlaceHorizontal(max(0, m.width), lipgloss.Center, title))
	b.WriteString("\n\n")

	// For Phase 2 this will render the main/filter/result body.
	b.WriteString(m.styles.MutedText.Render("Phase 2 in progress: wiring root model + layout + boot/session loading"))
	b.WriteString("\n\n")

	// Footer (help + toast)
	b.WriteString(m.status.View())
	b.WriteString("\n")

	return m.styles.BorderBox.Render(b.String())
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}


