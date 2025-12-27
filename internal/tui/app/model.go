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

		case m.mode == ModeMain && key.Matches(msg, m.keymap.Up):
			if m.selectedIndex > 0 {
				m.selectedIndex--
			}

		case m.mode == ModeMain && key.Matches(msg, m.keymap.Down):
			limit := len(m.ctrl.GetData().GetVisibleFiles())
			if m.showFiltered {
				limit = len(m.ctrl.GetData().GetFilteredFiles())
			}
			if limit > 0 && m.selectedIndex < limit-1 {
				m.selectedIndex++
			}

		case m.mode == ModeMain && key.Matches(msg, m.keymap.ToggleFilteredView):
			m.showFiltered = !m.showFiltered
			// Clamp selection to new list length.
			limit := len(m.ctrl.GetData().GetVisibleFiles())
			if m.showFiltered {
				limit = len(m.ctrl.GetData().GetFilteredFiles())
			}
			if limit == 0 {
				m.selectedIndex = 0
			} else if m.selectedIndex >= limit {
				m.selectedIndex = limit - 1
			}

		case m.mode == ModeMain && key.Matches(msg, m.keymap.ToggleIncluded):
			if m.showFiltered {
				// Filtered view is read-only for now.
				m.status, cmd = m.status.Update(events.ShowToastMsg{
					Text:     "Filtered view is read-only",
					Level:    events.ToastInfo,
					Duration: 2 * time.Second,
				})
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
				break
			}

			visible := m.ctrl.GetData().GetVisibleFiles()
			if len(visible) == 0 || m.selectedIndex < 0 || m.selectedIndex >= len(visible) {
				break
			}

			selected := visible[m.selectedIndex]
			for i, f := range m.ctrl.GetData().ChangedFiles {
				if f.Path == selected.Path {
					_ = m.ctrl.ToggleFileInclusion(i)
					cmds = append(cmds, saveSessionCmd(m.ctrl))
					break
				}
			}
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

	case events.SessionSavedMsg:
		m.status, cmd = m.status.Update(events.ShowToastMsg{
			Text:     "Session saved",
			Level:    events.ToastSuccess,
			Duration: 2 * time.Second,
		})
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case events.SessionSaveFailedMsg:
		m.status, cmd = m.status.Update(events.ShowToastMsg{
			Text:     "Failed to save session: " + msg.Err.Error(),
			Level:    events.ToastError,
			Duration: 5 * time.Second,
		})
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	// Let status model consume messages too (toast expiry, etc.).
	m.status, cmd = m.status.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func saveSessionCmd(ctrl *controller.Controller) tea.Cmd {
	return func() tea.Msg {
		path := ctrl.GetDefaultSessionPath()
		if err := ctrl.SaveSession(path); err != nil {
			return events.SessionSaveFailedMsg{Path: path, Err: err}
		}
		return events.SessionSavedMsg{Path: path}
	}
}

func (m Model) View() string {
	return m.view()
}


