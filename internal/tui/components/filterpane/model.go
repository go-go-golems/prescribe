package filterpane

import (
	"fmt"
	"strings"

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
	filter domain.Filter
}

func (i item) Title() string { return i.filter.Name }
func (i item) Description() string {
	if i.filter.Description == "" {
		return ""
	}
	return i.filter.Description
}
func (i item) FilterValue() string { return i.filter.Name }

type Model struct {
	list   list.Model
	keymap keys.KeyMap
	styles styles.Styles

	width  int
	height int
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

func (m Model) View() string {
	var b strings.Builder

	// List
	listView := m.list.View()
	b.WriteString(listView)
	if !strings.HasSuffix(listView, "\n") {
		b.WriteString("\n")
	}

	// Rule preview (bounded-ish)
	if f, ok := m.SelectedFilter(); ok {
		b.WriteString("\n")
		b.WriteString(m.styles.Header.Render("RULES"))
		b.WriteString("\n")
		for _, r := range f.Rules {
			b.WriteString(m.styles.MutedText.Render(fmt.Sprintf("  %s: %s", r.Type, r.Pattern)))
			b.WriteString("\n")
		}
	}

	return b.String()
}

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
		case key.Matches(msg, m.keymap.DeleteFilter):
			if len(m.list.Items()) == 0 {
				return m, nil
			}
			idx := m.list.Index()
			return m, func() tea.Msg { return events.RemoveFilterRequested{Index: idx} }
		case key.Matches(msg, m.keymap.ClearFilters):
			return m, func() tea.Msg { return events.ClearFiltersRequested{} }
		case key.Matches(msg, m.keymap.Preset1):
			return m, func() tea.Msg { return events.AddFilterPresetRequested{PresetID: "exclude-tests"} }
		case key.Matches(msg, m.keymap.Preset2):
			return m, func() tea.Msg { return events.AddFilterPresetRequested{PresetID: "exclude-docs"} }
		case key.Matches(msg, m.keymap.Preset3):
			return m, func() tea.Msg { return events.AddFilterPresetRequested{PresetID: "only-source"} }
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
	m.width = w
	m.height = h
	m.list.SetSize(w, h)
}

func (m *Model) SetFilters(filters []domain.Filter) {
	items := make([]list.Item, 0, len(filters))
	for _, f := range filters {
		items = append(items, item{filter: f})
	}
	m.list.SetItems(items)
}

func (m Model) SelectedFilter() (domain.Filter, bool) {
	it, ok := m.list.SelectedItem().(item)
	if !ok {
		return domain.Filter{}, false
	}
	return it.filter, true
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
