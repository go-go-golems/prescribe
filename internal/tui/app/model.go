package app

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/key"
	"github.com/go-go-golems/prescribe/internal/controller"
	"github.com/go-go-golems/prescribe/internal/tui/components/status"
	"github.com/go-go-golems/prescribe/internal/tui/events"
	"github.com/go-go-golems/prescribe/internal/tui/keys"
	"github.com/go-go-golems/prescribe/internal/tui/styles"
)

// New creates a new app root model.
//
// NOTE: Phase 2 will wire this into `prescribe tui`. For now we keep this model
// compiling while we build out behavior incrementally.
func New(ctrl *controller.Controller, deps Deps) Model {
	km := keys.Default()
	st := styles.Default()
	sm := statusModel(km, st)

	return Model{
		ctrl:   ctrl,
		deps:   deps,
		mode:   ModeMain,
		keymap: km,
		styles: st,
		status: sm,
	}
}

func statusModel(km keys.KeyMap, st styles.Styles) (m status.Model) {
	m = status.New(km, st)
	return m
}

func (m Model) Init() tea.Cmd {
	return bootCmd(m.ctrl)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.status.SetSize(m.width)
		m.status.SetShowFullHelp(m.showFullHelp)

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keymap.Help):
			m.showFullHelp = !m.showFullHelp
			m.status.SetShowFullHelp(m.showFullHelp)
		}

	case events.SessionLoadedMsg:
		m.status, cmd = m.status.Update(events.ShowToastMsg{
			Text:     "Session loaded",
			Level:    events.ToastSuccess,
			Duration: 2 * time.Second,
		})
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	case events.SessionLoadFailedMsg:
		m.status, cmd = m.status.Update(events.ShowToastMsg{
			Text:     "Failed to load session: " + msg.Err.Error(),
			Level:    events.ToastWarning,
			Duration: 5 * time.Second,
		})
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	case events.SessionLoadSkippedMsg:
		// No toast for missing session by default; keep quiet.
	}

	// Let status model consume messages too (toast expiry, etc.).
	m.status, cmd = m.status.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	return m.view()
}


