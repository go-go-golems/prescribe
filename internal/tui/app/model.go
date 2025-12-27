package app

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/prescribe/internal/controller"
	"github.com/go-go-golems/prescribe/internal/domain"
	"github.com/go-go-golems/prescribe/internal/tui/components/status"
	"github.com/go-go-golems/prescribe/internal/tui/events"
	"github.com/go-go-golems/prescribe/internal/tui/keys"
	"github.com/go-go-golems/prescribe/internal/tui/layout"
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
		m.layout = layout.Compute(m.width, m.height, 0, lipgloss.Height(m.status.View()))

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keymap.Help):
			m.showFullHelp = !m.showFullHelp
			m.status.SetShowFullHelp(m.showFullHelp)
			m.layout = layout.Compute(m.width, m.height, 0, lipgloss.Height(m.status.View()))

		case key.Matches(msg, m.keymap.Back):
			// Global "back" semantics.
			switch m.mode {
			case ModeFilters, ModeResult:
				m.mode = ModeMain
			}

		case m.mode == ModeMain && key.Matches(msg, m.keymap.OpenFilters):
			m.mode = ModeFilters
			m.filterIndex = 0

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

		case m.mode == ModeMain && key.Matches(msg, m.keymap.Generate):
			m.mode = ModeGenerating
			cmds = append(cmds, generateCmd(m.ctrl))

		case m.mode == ModeFilters && key.Matches(msg, m.keymap.Up):
			if m.filterIndex > 0 {
				m.filterIndex--
			}

		case m.mode == ModeFilters && key.Matches(msg, m.keymap.Down):
			filters := m.ctrl.GetFilters()
			if len(filters) > 0 && m.filterIndex < len(filters)-1 {
				m.filterIndex++
			}

		case m.mode == ModeFilters && key.Matches(msg, m.keymap.DeleteFilter):
			filters := m.ctrl.GetFilters()
			if len(filters) > 0 && m.filterIndex >= 0 && m.filterIndex < len(filters) {
				_ = m.ctrl.RemoveFilter(m.filterIndex)
				cmds = append(cmds, saveSessionCmd(m.ctrl))
				// Clamp after deletion.
				if m.filterIndex >= len(m.ctrl.GetFilters()) {
					m.filterIndex = max(0, len(m.ctrl.GetFilters())-1)
				}
			}

		case m.mode == ModeFilters && key.Matches(msg, m.keymap.ClearFilters):
			m.ctrl.ClearFilters()
			cmds = append(cmds, saveSessionCmd(m.ctrl))
			m.filterIndex = 0

		case m.mode == ModeFilters && key.Matches(msg, m.keymap.Preset1):
			m.ctrl.AddFilter(domain.Filter{
				Name:        "Exclude Tests",
				Description: "Exclude test files",
				Rules: []domain.FilterRule{
					{Type: domain.FilterTypeExclude, Pattern: "**/*test*"},
					{Type: domain.FilterTypeExclude, Pattern: "**/*spec*"},
				},
			})
			cmds = append(cmds, saveSessionCmd(m.ctrl))

		case m.mode == ModeFilters && key.Matches(msg, m.keymap.Preset2):
			m.ctrl.AddFilter(domain.Filter{
				Name:        "Exclude Docs",
				Description: "Exclude documentation files",
				Rules: []domain.FilterRule{
					{Type: domain.FilterTypeExclude, Pattern: "**/*.md"},
					{Type: domain.FilterTypeExclude, Pattern: "**/docs/**"},
				},
			})
			cmds = append(cmds, saveSessionCmd(m.ctrl))

		case m.mode == ModeFilters && key.Matches(msg, m.keymap.Preset3):
			m.ctrl.AddFilter(domain.Filter{
				Name:        "Only Source",
				Description: "Include only source code files",
				Rules: []domain.FilterRule{
					{Type: domain.FilterTypeInclude, Pattern: "**/*.go"},
					{Type: domain.FilterTypeInclude, Pattern: "**/*.ts"},
					{Type: domain.FilterTypeInclude, Pattern: "**/*.js"},
					{Type: domain.FilterTypeInclude, Pattern: "**/*.py"},
				},
			})
			cmds = append(cmds, saveSessionCmd(m.ctrl))
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

	case events.DescriptionGeneratedMsg:
		m.generatedDesc = msg.Text
		m.err = nil
		m.mode = ModeResult

	case events.DescriptionGenerationFailedMsg:
		m.generatedDesc = ""
		m.err = msg.Err
		m.mode = ModeResult
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

func generateCmd(ctrl *controller.Controller) tea.Cmd {
	return func() tea.Msg {
		desc, err := ctrl.GenerateDescription()
		if err != nil {
			return events.DescriptionGenerationFailedMsg{Err: err}
		}
		return events.DescriptionGeneratedMsg{Text: desc}
	}
}

func (m Model) View() string {
	return m.view()
}


