package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) view() string {
	switch m.mode {
	case ModeMain:
		return m.renderMain()
	case ModeFilters:
		return m.renderFilters()
	case ModeGenerating:
		return m.renderGenerating()
	case ModeResult:
		return m.renderResult()
	default:
		return m.renderMain()
	}
}

func (m Model) renderMain() string {
	data := m.ctrl.GetData()
	files := data.GetVisibleFiles()
	if m.showFiltered {
		files = data.GetFilteredFiles()
	}

	var b strings.Builder

	title := m.styles.Title.Render("PRESCRIBE")
	b.WriteString(lipgloss.PlaceHorizontal(max(0, m.layout.Width), lipgloss.Center, title))
	b.WriteString("\n\n")

	branchInfo := fmt.Sprintf("%s → %s", data.SourceBranch, data.TargetBranch)
	b.WriteString(m.styles.Base.Render(branchInfo))
	b.WriteString("\n\n")

	stats := fmt.Sprintf("Files: %d visible, %d filtered | Tokens: %d | Filters: %d",
		len(data.GetVisibleFiles()),
		len(data.GetFilteredFiles()),
		data.GetTotalTokens(),
		len(data.ActiveFilters),
	)
	b.WriteString(m.styles.Base.Render(stats))
	b.WriteString("\n\n")

	if m.showFiltered {
		b.WriteString(m.styles.Header.Render("FILTERED FILES"))
	} else {
		b.WriteString(m.styles.Header.Render("CHANGED FILES"))
	}
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", max(0, m.layout.Width)))
	b.WriteString("\n")

	if len(files) == 0 {
		b.WriteString(m.styles.MutedText.Render("No files to show"))
		b.WriteString("\n")
	} else {
		listView := m.filelist.View()
		b.WriteString(listView)
		if !strings.HasSuffix(listView, "\n") {
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(m.status.View())

	boxW, boxH := m.boxWH()
	return strings.TrimRight(
		m.styles.BorderBox.
			Width(boxW).Height(boxH).
			Render(b.String()),
		"\n",
	)
}

func (m Model) renderFilters() string {
	data := m.ctrl.GetData()
	filters := m.ctrl.GetFilters()

	var b strings.Builder

	title := m.styles.Title.Render("FILTER MANAGEMENT")
	b.WriteString(lipgloss.PlaceHorizontal(max(0, m.layout.Width), lipgloss.Center, title))
	b.WriteString("\n\n")

	stats := fmt.Sprintf("Active Filters: %d | Filtered Files: %d | Files: %d visible",
		len(filters),
		len(data.GetFilteredFiles()),
		len(data.GetVisibleFiles()),
	)
	b.WriteString(m.styles.Base.Render(stats))
	b.WriteString("\n\n")

	b.WriteString(m.styles.Header.Render("ACTIVE FILTERS"))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", max(0, m.layout.Width)))
	b.WriteString("\n")

	if len(filters) == 0 {
		b.WriteString(m.styles.MutedText.Render("No active filters"))
		b.WriteString("\n")
	} else {
		paneView := m.filterpane.View()
		b.WriteString(paneView)
		if !strings.HasSuffix(paneView, "\n") {
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(m.styles.Header.Render("QUICK ADD PRESETS"))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", max(0, m.layout.Width)))
	b.WriteString("\n")
	b.WriteString(m.styles.Base.Render("[1] Exclude Tests  [2] Exclude Docs  [3] Only Source"))
	b.WriteString("\n\n")

	b.WriteString(m.status.View())

	boxW, boxH := m.boxWH()
	return strings.TrimRight(
		m.styles.BorderBox.
			Width(boxW).Height(boxH).
			Render(b.String()),
		"\n",
	)
}

func (m Model) renderGenerating() string {
	var b strings.Builder
	title := m.styles.Title.Render("GENERATING")
	b.WriteString(lipgloss.PlaceHorizontal(max(0, m.layout.Width), lipgloss.Center, title))
	b.WriteString("\n\n")
	b.WriteString(m.styles.Base.Render("Generating PR description..."))
	b.WriteString("\n\n")
	b.WriteString(m.status.View())
	boxW, boxH := m.boxWH()
	return strings.TrimRight(
		m.styles.BorderBox.
			Width(boxW).Height(boxH).
			Render(b.String()),
		"\n",
	)
}

func (m Model) renderResult() string {
	var b strings.Builder

	title := m.styles.Title.Render("RESULT")
	b.WriteString(lipgloss.PlaceHorizontal(max(0, m.layout.Width), lipgloss.Center, title))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(m.styles.ErrorText.Render("Error: " + m.err.Error()))
		b.WriteString("\n\n")
	} else {
		b.WriteString(m.result.View())
		b.WriteString("\n")
	}

	b.WriteString(m.status.View())

	boxW, boxH := m.boxWH()
	return strings.TrimRight(
		m.styles.BorderBox.
			Width(boxW).Height(boxH).
			Render(b.String()),
		"\n",
	)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
