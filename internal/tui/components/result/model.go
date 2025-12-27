package result

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// Model renders scrollable result content using bubbles/viewport.
type Model struct {
	vp viewport.Model
}

func New() Model {
	return Model{vp: viewport.New(0, 0)}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.vp, cmd = m.vp.Update(msg)
	return m, cmd
}

func (m Model) View() string { return m.vp.View() }

func (m *Model) SetSize(w, h int) {
	if w < 0 {
		w = 0
	}
	if h < 0 {
		h = 0
	}
	m.vp.Width = w
	m.vp.Height = h
}

func (m *Model) SetContent(s string) {
	m.vp.SetContent(s)
}
