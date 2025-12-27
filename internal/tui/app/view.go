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
	b.WriteString(lipgloss.PlaceHorizontal(max(0, m.width), lipgloss.Center, title))
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
	b.WriteString(strings.Repeat("─", max(0, m.width-6)))
	b.WriteString("\n")

	if len(files) == 0 {
		b.WriteString(m.styles.MutedText.Render("No files to show"))
		b.WriteString("\n")
	} else {
		for i, f := range files {
			included := " "
			if f.Included {
				included = "✓"
			}
			line := fmt.Sprintf("[%s] %s +%d -%d (%dt)", included, f.Path, f.Additions, f.Deletions, f.Tokens)
			if i == m.selectedIndex {
				b.WriteString(m.styles.SelectedItem.Render("▶ " + line))
			} else {
				b.WriteString(m.styles.UnselectedItem.Render(line))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(m.status.View())
	b.WriteString("\n")

	return m.styles.BorderBox.Render(b.String())
}

func (m Model) renderFilters() string {
	var b strings.Builder

	title := m.styles.Title.Render("FILTERS (WIP)")
	b.WriteString(lipgloss.PlaceHorizontal(max(0, m.width), lipgloss.Center, title))
	b.WriteString("\n\n")

	b.WriteString(m.styles.MutedText.Render("Filter pane will be extracted in Phase 5. Press Esc to go back."))
	b.WriteString("\n\n")
	b.WriteString(m.status.View())
	b.WriteString("\n")

	return m.styles.BorderBox.Render(b.String())
}

func (m Model) renderGenerating() string {
	var b strings.Builder
	title := m.styles.Title.Render("GENERATING")
	b.WriteString(lipgloss.PlaceHorizontal(max(0, m.width), lipgloss.Center, title))
	b.WriteString("\n\n")
	b.WriteString(m.styles.Base.Render("Generating PR description..."))
	b.WriteString("\n\n")
	b.WriteString(m.status.View())
	b.WriteString("\n")
	return m.styles.BorderBox.Render(b.String())
}

func (m Model) renderResult() string {
	var b strings.Builder

	title := m.styles.Title.Render("RESULT")
	b.WriteString(lipgloss.PlaceHorizontal(max(0, m.width), lipgloss.Center, title))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(m.styles.ErrorText.Render("Error: " + m.err.Error()))
		b.WriteString("\n\n")
	} else {
		b.WriteString(m.styles.Box.Render(m.generatedDesc))
		b.WriteString("\n\n")
	}

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


