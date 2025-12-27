package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/prescribe/internal/controller"
	"github.com/go-go-golems/prescribe/internal/domain"
	"github.com/go-go-golems/prescribe/internal/tui/components/filelist"
	"github.com/go-go-golems/prescribe/internal/tui/components/result"
	"github.com/go-go-golems/prescribe/internal/tui/components/status"
	"github.com/go-go-golems/prescribe/internal/tui/events"
	"github.com/go-go-golems/prescribe/internal/tui/export"
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
	rm := result.New()
	fl := filelist.New(km, st)

	return Model{
		ctrl:     ctrl,
		deps:     deps,
		mode:     ModeMain,
		keymap:   km,
		styles:   st,
		status:   sm,
		result:   rm,
		filelist: fl,
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
		m.layout = layout.Compute(m.width, m.height, m.headerHeight(), lipgloss.Height(m.status.View()))
		m.result.SetSize(m.layout.BodyW, m.layout.BodyH)
		m.filelist.SetSize(m.layout.BodyW, m.layout.BodyH)
		m.syncFilelist()

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keymap.Help):
			m.showFullHelp = !m.showFullHelp
			m.status.SetShowFullHelp(m.showFullHelp)
			m.layout = layout.Compute(m.width, m.height, m.headerHeight(), lipgloss.Height(m.status.View()))
			m.result.SetSize(m.layout.BodyW, m.layout.BodyH)
			m.filelist.SetSize(m.layout.BodyW, m.layout.BodyH)
			m.syncFilelist()

		case key.Matches(msg, m.keymap.Back):
			// Global "back" semantics.
			switch m.mode {
			case ModeFilters, ModeResult:
				m.mode = ModeMain
				m.syncFilelist()
			}

		case m.mode == ModeMain && key.Matches(msg, m.keymap.OpenFilters):
			m.mode = ModeFilters
			m.filterIndex = 0

		case m.mode == ModeMain && key.Matches(msg, m.keymap.ToggleFilteredView):
			m.showFiltered = !m.showFiltered
			m.syncFilelist()

		case m.mode == ModeMain && key.Matches(msg, m.keymap.Generate):
			m.mode = ModeGenerating
			cmds = append(cmds, generateCmd(m.ctrl))

		case (m.mode == ModeMain || m.mode == ModeResult) && key.Matches(msg, m.keymap.CopyContext):
			cmds = append(cmds, copyContextCmd(m.ctrl, m.deps))

		case m.mode == ModeMain:
			m.filelist, cmd = m.filelist.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}

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
		m.result.SetContent(m.generatedDesc)
		m.err = nil
		m.mode = ModeResult

	case events.DescriptionGenerationFailedMsg:
		m.generatedDesc = ""
		m.result.SetContent("")
		m.err = msg.Err
		m.mode = ModeResult

	case events.ToggleFileIncludedRequested:
		if m.showFiltered {
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

		current, ok := m.currentIncludedByPath(msg.Path)
		if !ok {
			m.status, cmd = m.status.Update(events.ShowToastMsg{
				Text:     "File not found: " + msg.Path,
				Level:    events.ToastError,
				Duration: 5 * time.Second,
			})
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			break
		}

		if err := m.ctrl.SetFileIncludedByPath(msg.Path, !current); err != nil {
			m.status, cmd = m.status.Update(events.ShowToastMsg{
				Text:     "Failed to toggle file: " + err.Error(),
				Level:    events.ToastError,
				Duration: 5 * time.Second,
			})
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			break
		}

		m.syncFilelist()
		cmds = append(cmds, saveSessionCmd(m.ctrl))

	case events.SetAllVisibleIncludedRequested:
		if m.showFiltered {
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

		n, err := m.ctrl.SetAllVisibleIncluded(msg.Included)
		if err != nil {
			m.status, cmd = m.status.Update(events.ShowToastMsg{
				Text:     "Bulk update failed: " + err.Error(),
				Level:    events.ToastError,
				Duration: 5 * time.Second,
			})
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			break
		}

		if n == 0 {
			m.status, cmd = m.status.Update(events.ShowToastMsg{
				Text:     "No visible files",
				Level:    events.ToastInfo,
				Duration: 2 * time.Second,
			})
		} else {
			verb := "Selected"
			if !msg.Included {
				verb = "Unselected"
			}
			m.status, cmd = m.status.Update(events.ShowToastMsg{
				Text:     fmt.Sprintf("%s %d files", verb, n),
				Level:    events.ToastSuccess,
				Duration: 2 * time.Second,
			})
		}
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		m.syncFilelist()
		cmds = append(cmds, saveSessionCmd(m.ctrl))

	case events.ClipboardCopiedMsg:
		m.status, cmd = m.status.Update(events.ShowToastMsg{
			Text:     fmt.Sprintf("Copied %s (%d bytes)", msg.What, msg.Bytes),
			Level:    events.ToastSuccess,
			Duration: 2 * time.Second,
		})
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case events.ClipboardCopyFailedMsg:
		level := events.ToastError
		text := "Copy failed: " + msg.Err.Error()
		if strings.Contains(msg.Err.Error(), "no files included") {
			level = events.ToastWarning
			text = msg.Err.Error()
		}
		m.status, cmd = m.status.Update(events.ShowToastMsg{
			Text:     text,
			Level:    level,
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

	// Let result model consume messages too (viewport scrolling / internal state),
	// but only while in result mode to avoid stealing navigation keys.
	if m.mode == ModeResult {
		m.result, cmd = m.result.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
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

func copyContextCmd(ctrl *controller.Controller, deps Deps) tea.Cmd {
	return func() tea.Msg {
		req, err := ctrl.BuildGenerateDescriptionRequest()
		if err != nil {
			return events.ClipboardCopyFailedMsg{Err: err}
		}

		text := export.BuildGenerationContextText(req)
		if err := deps.ClipboardWriteAll(text); err != nil {
			return events.ClipboardCopyFailedMsg{Err: err}
		}
		return events.ClipboardCopiedMsg{What: "context", Bytes: len(text)}
	}
}

var _ tea.Model = Model{}
var _ tea.Model = (*Model)(nil)

func (m Model) View() string {
	return m.view()
}

func (m Model) headerHeight() int {
	// Only the Result screen currently needs explicit "body vs header" layout separation.
	// We keep this intentionally conservative until Phase 4/5 introduce real list components.
	switch m.mode {
	case ModeMain:
		// renderMain writes a fixed header before the file list:
		// title + blank + branch + blank + stats + blank + section header + separator.
		return 8
	case ModeResult:
		// renderResult writes: title line + "\n\n" (=> 3 lines total before the viewport)
		return 3
	default:
		return 0
	}
}

func (m *Model) syncFilelist() {
	idx := m.filelist.SelectedIndex()
	files := m.ctrl.GetData().GetVisibleFiles()
	if m.showFiltered {
		files = m.ctrl.GetData().GetFilteredFiles()
	}
	m.filelist.SetFiles(files)
	m.filelist.SetSelectedIndex(idx)
	m.selectedIndex = m.filelist.SelectedIndex()
}

func (m Model) currentIncludedByPath(path string) (bool, bool) {
	for _, f := range m.ctrl.GetData().ChangedFiles {
		if f.Path == path {
			return f.Included, true
		}
	}
	return false, false
}
