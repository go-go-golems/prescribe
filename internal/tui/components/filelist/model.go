package filelist

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/prescribe/internal/domain"
	"github.com/go-go-golems/prescribe/internal/tui/events"
	"github.com/go-go-golems/prescribe/internal/tui/keys"
	"github.com/go-go-golems/prescribe/internal/tui/styles"
)

type item struct {
	file domain.FileChange
}

func (i item) Title() string { return i.file.Path }
func (i item) Description() string {
	included := " "
	if i.file.Included {
		included = "✓"
	}
	return fmt.Sprintf("[%s] +%d -%d (%dt)", included, i.file.Additions, i.file.Deletions, i.file.Tokens)
}
func (i item) FilterValue() string { return i.file.Path }

type singleLineDelegate struct {
	styles styles.Styles
}

func (d singleLineDelegate) Height() int  { return 1 }
func (d singleLineDelegate) Spacing() int { return 0 }

func (d singleLineDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d singleLineDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	it, ok := listItem.(item)
	if !ok {
		return
	}

	included := " "
	if it.file.Included {
		included = "✓"
	}

	// Single-line, old-style summary. Keep it compact; truncate if terminal is narrow.
	line := fmt.Sprintf("[%s] %s +%d -%d (%dt)",
		included,
		it.file.Path,
		it.file.Additions,
		it.file.Deletions,
		it.file.Tokens,
	)

	prefix := "  "
	style := lipgloss.NewStyle()
	if index == m.Index() {
		prefix = "▶ "
		style = lipgloss.NewStyle().Foreground(d.styles.Primary).Bold(true)
	}

	available := m.Width()
	// Leave a tiny buffer so we don't wrap in tight terminals.
	if available > 0 {
		available = maxInt(0, available-1)
	}
	out := prefix + line
	out = truncate(out, available)

	_, _ = fmt.Fprint(w, style.Render(out))
}

type Model struct {
	list   list.Model
	keymap keys.KeyMap
	styles styles.Styles
}

func New(km keys.KeyMap, st styles.Styles) Model {
	delegate := singleLineDelegate{styles: st}

	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = ""
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.SetShowPagination(false)

	return Model{list: l, keymap: km, styles: st}
}

func (m Model) View() string { return m.list.View() }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.Up):
			m.list.CursorUp()
			return m, nil
		case key.Matches(msg, m.keymap.Down):
			m.list.CursorDown()
			return m, nil
		case key.Matches(msg, m.keymap.ToggleIncluded):
			path, ok := m.SelectedPath()
			if !ok {
				return m, nil
			}
			return m, func() tea.Msg { return events.ToggleFileIncludedRequested{Path: path} }
		case key.Matches(msg, m.keymap.SelectAllVisible):
			return m, func() tea.Msg { return events.SetAllVisibleIncludedRequested{Included: true} }
		case key.Matches(msg, m.keymap.UnselectAllVisible):
			return m, func() tea.Msg { return events.SetAllVisibleIncludedRequested{Included: false} }
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *Model) SetSize(w, h int) {
	if w < 0 {
		w = 0
	}
	if h < 0 {
		h = 0
	}
	m.list.SetSize(w, h)
}

func (m *Model) SetFiles(files []domain.FileChange) {
	items := make([]list.Item, 0, len(files))
	for _, f := range files {
		items = append(items, item{file: f})
	}
	m.list.SetItems(items)
}

func (m Model) SelectedPath() (string, bool) {
	it, ok := m.list.SelectedItem().(item)
	if !ok {
		return "", false
	}
	return it.file.Path, true
}

func (m Model) SelectedIncluded() (bool, bool) {
	it, ok := m.list.SelectedItem().(item)
	if !ok {
		return false, false
	}
	return it.file.Included, true
}

func (m *Model) SetSelectedIndex(i int) {
	if i < 0 {
		i = 0
	}
	if i >= len(m.list.Items()) && len(m.list.Items()) > 0 {
		i = len(m.list.Items()) - 1
	}
	m.list.Select(i)
}

func (m Model) SelectedIndex() int { return m.list.Index() }

func (m *Model) CursorUp()   { m.list.CursorUp() }
func (m *Model) CursorDown() { m.list.CursorDown() }

func truncate(s string, w int) string {
	if w <= 0 {
		return s
	}
	// If already short enough, return as-is.
	if len([]rune(s)) <= w {
		return s
	}
	if w == 1 {
		return "…"
	}
	rs := []rune(s)
	return string(rs[:w-1]) + "…"
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
