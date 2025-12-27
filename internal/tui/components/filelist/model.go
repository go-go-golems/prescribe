package filelist

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/prescribe/internal/domain"
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

type Model struct {
	list   list.Model
	keymap keys.KeyMap
	styles styles.Styles
}

func New(km keys.KeyMap, st styles.Styles) Model {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().Inherit(st.SelectedItem)
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().Inherit(st.SelectedItem)
	delegate.Styles.NormalTitle = lipgloss.NewStyle().Inherit(st.UnselectedItem)
	delegate.Styles.NormalDesc = lipgloss.NewStyle().Inherit(st.UnselectedItem)

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
