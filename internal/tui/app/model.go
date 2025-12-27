package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/prescribe/internal/controller"
	"github.com/go-go-golems/prescribe/internal/tui/components/status"
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
	// Phase 2 will grow this Update into the real mode machine.
	// For now, keep this a compiling placeholder.
	var cmd tea.Cmd
	m.status, cmd = m.status.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	return m.view()
}


