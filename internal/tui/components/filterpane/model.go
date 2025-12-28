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

	return Model{list: l, keymap: km, styles: st}.recomputeLayout()
}

func (m Model) View() string {
	var b strings.Builder

	// List
	listView := m.list.View()
	b.WriteString(listView)
	if !strings.HasSuffix(listView, "\n") {
		b.WriteString("\n")
	}

	// Rule preview (bounded to component height; list height is reduced accordingly)
	if m.previewHeight() > 0 {
		f, ok := m.SelectedFilter()
		if ok {
			b.WriteString("\n")
			b.WriteString(m.styles.Header.Render("RULES"))
			b.WriteString("\n")

			rulesShown := minInt(len(f.Rules), m.maxRulesToShow())
			for i := 0; i < rulesShown; i++ {
				r := f.Rules[i]
				b.WriteString(m.styles.MutedText.Render(fmt.Sprintf("  %s: %s", r.Type, r.Pattern)))
				b.WriteString("\n")
			}
			if rulesShown < len(f.Rules) {
				b.WriteString(m.styles.MutedText.Render("  …"))
				b.WriteString("\n")
			}
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
			return m.recomputeLayout(), nil
		case key.Matches(msg, m.keymap.Down):
			m.list.CursorDown()
			return m.recomputeLayout(), nil
		case key.Matches(msg, m.keymap.DeleteFilter):
			if len(m.list.Items()) == 0 {
				return m, nil
			}
			idx := m.list.Index()
			return m, func() tea.Msg { return events.RemoveFilterRequested{Index: idx} }
		case key.Matches(msg, m.keymap.ClearFilters):
			return m, func() tea.Msg { return events.ClearFiltersRequested{} }
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m.recomputeLayout(), cmd
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
	*m = m.recomputeLayout()
}

func (m *Model) SetFilters(filters []domain.Filter) {
	items := make([]list.Item, 0, len(filters))
	for _, f := range filters {
		items = append(items, item{filter: f})
	}
	m.list.SetItems(items)
	*m = m.recomputeLayout()
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
	*m = m.recomputeLayout()
}

func (m Model) SelectedIndex() int { return m.list.Index() }

func (m Model) recomputeLayout() Model {
	listH := maxInt(0, m.height-m.previewHeight())
	m.list.SetSize(m.width, listH)
	return m
}

func (m Model) previewHeight() int {
	if m.height <= 0 {
		return 0
	}
	f, ok := m.SelectedFilter()
	if !ok || len(f.Rules) == 0 {
		return 0
	}

	// 1 blank + 1 header + rules (bounded) + optional "…"
	rulesShown := minInt(len(f.Rules), m.maxRulesToShow())
	h := 1 + 1 + rulesShown
	if rulesShown < len(f.Rules) {
		h++
	}
	if h > m.height {
		return m.height
	}
	return h
}

func (m Model) maxRulesToShow() int {
	// Keep at least a few rows for the list itself.
	const minListH = 3
	available := m.height - minListH - 2 // 2 = blank + header
	if available < 0 {
		return 0
	}
	// Leave room for the ellipsis line if needed.
	return available
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
