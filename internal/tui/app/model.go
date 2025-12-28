package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/prescribe/internal/controller"
	"github.com/go-go-golems/prescribe/internal/domain"
	pexport "github.com/go-go-golems/prescribe/internal/export"
	"github.com/go-go-golems/prescribe/internal/tui/components/filelist"
	"github.com/go-go-golems/prescribe/internal/tui/components/filterpane"
	"github.com/go-go-golems/prescribe/internal/tui/components/result"
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
	rm := result.New()
	fl := filelist.New(km, st)
	fp := filterpane.New(km, st)

	return Model{
		ctrl:       ctrl,
		deps:       deps,
		mode:       ModeMain,
		keymap:     km,
		styles:     st,
		status:     sm,
		result:     rm,
		filelist:   fl,
		filterpane: fp,
	}
}

func statusModel(km keys.KeyMap, st styles.Styles) status.Model {
	return status.New(km, st)
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		bootCmd(m.ctrl),
		loadFilterPresetsCmd(m.ctrl),
	)
}

func (m Model) contentWH() (int, int) {
	frameW, frameH := m.styles.BorderBox.GetFrameSize()

	// Content size is the terminal size minus the border+padding frame.
	// This is the size we should pass to child components and PlaceHorizontal() calls.
	// Keep a 1-row slack to prevent terminal scrolling when something accidentally
	// over-produces a line (tmux captures then “lose” the top border).
	return maxInt(0, m.width-frameW), maxInt(0, m.height-frameH-1)
}

func (m Model) boxWH() (int, int) {
	// BorderBox.Width/Height apply to the content block *before* the border is drawn.
	// So to make the overall rendered box fit the terminal:
	//   boxW = terminalW - borderW
	//   boxH = terminalH - borderH
	b, top, right, bottom, left := m.styles.BorderBox.GetBorder()
	borderW := 0
	borderH := 0
	if left {
		borderW += b.GetLeftSize()
	}
	if right {
		borderW += b.GetRightSize()
	}
	if top {
		borderH += b.GetTopSize()
	}
	if bottom {
		borderH += b.GetBottomSize()
	}
	// Keep a 1-row slack to avoid scrolling in tmux captures.
	return maxInt(0, m.width-borderW), maxInt(0, m.height-borderH-1)
}

func (m *Model) footerHeight() int {
	// Most screens put a blank line before the status footer for readability.
	base := 1 + lipgloss.Height(m.status.View())
	switch m.mode {
	case ModeFilters:
		// Filters screen renders a fixed "Quick Add Presets" block before the status footer.
		// Height breakdown:
		// - blank line (1)
		// - header line (1)
		// - separator line (1)
		// - presets line (1)
		// - blank line (1)
		const presetsBlockH = 5
		return presetsBlockH + base
	case ModeMain, ModeGenerating, ModeResult:
		return base
	}
	return base
}

func (m *Model) recomputeLayout() {
	contentW, contentH := m.contentWH()
	m.layout = layout.Compute(contentW, contentH, m.headerHeight(), m.footerHeight())
	m.result.SetSize(m.layout.BodyW, m.layout.BodyH)
	m.filelist.SetSize(m.layout.BodyW, m.layout.BodyH)
	m.filterpane.SetSize(m.layout.BodyW, m.layout.BodyH)
	m.syncFilelist()
	m.syncFilterpane()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		contentW, _ := m.contentWH()
		m.status.SetSize(contentW)
		m.status.SetShowFullHelp(m.showFullHelp)
		m.recomputeLayout()

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keymap.Help):
			m.showFullHelp = !m.showFullHelp
			m.status.SetShowFullHelp(m.showFullHelp)
			m.recomputeLayout()

		case key.Matches(msg, m.keymap.Back):
			// Global "back" semantics.
			switch m.mode {
			case ModeFilters, ModeResult:
				m.mode = ModeMain
				m.recomputeLayout()
			case ModeMain, ModeGenerating:
				// no-op
			}

		case m.mode == ModeMain && key.Matches(msg, m.keymap.OpenFilters):
			m.mode = ModeFilters
			m.recomputeLayout()

		case m.mode == ModeFilters && key.Matches(msg, m.keymap.Preset1):
			m = m.applyQuickPresetByIndex(0)
			m.syncFilelist()
			m.syncFilterpane()
			cmds = append(cmds, saveSessionCmd(m.ctrl))
		case m.mode == ModeFilters && key.Matches(msg, m.keymap.Preset2):
			m = m.applyQuickPresetByIndex(1)
			m.syncFilelist()
			m.syncFilterpane()
			cmds = append(cmds, saveSessionCmd(m.ctrl))
		case m.mode == ModeFilters && key.Matches(msg, m.keymap.Preset3):
			m = m.applyQuickPresetByIndex(2)
			m.syncFilelist()
			m.syncFilterpane()
			cmds = append(cmds, saveSessionCmd(m.ctrl))

		case m.mode == ModeMain && key.Matches(msg, m.keymap.ToggleFilteredView):
			m.showFiltered = !m.showFiltered
			m.syncFilelist()

		case m.mode == ModeMain && key.Matches(msg, m.keymap.Generate):
			m.mode = ModeGenerating
			m.recomputeLayout()
			cmds = append(cmds, generateCmd(m.ctrl))

		case (m.mode == ModeMain || m.mode == ModeResult) && key.Matches(msg, m.keymap.CopyContext):
			cmds = append(cmds, copyContextCmd(m.ctrl, m.deps))

		case m.mode == ModeMain:
			m.filelist, cmd = m.filelist.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}

		case m.mode == ModeFilters:
			m.filterpane, cmd = m.filterpane.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}

	case events.SessionLoadedMsg:
		m.status, cmd = m.status.Update(events.ShowToastMsg{
			Text:     "Session loaded",
			Level:    events.ToastSuccess,
			Duration: 2 * time.Second,
		})
		// Session load may change included bits and filters.
		m.syncFilelist()
		m.syncFilterpane()
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
		m.recomputeLayout()

	case events.DescriptionGenerationFailedMsg:
		m.generatedDesc = ""
		m.result.SetContent("")
		m.err = msg.Err
		m.mode = ModeResult
		m.recomputeLayout()

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

	case events.RemoveFilterRequested:
		if err := m.ctrl.RemoveFilter(msg.Index); err != nil {
			m.status, cmd = m.status.Update(events.ShowToastMsg{
				Text:     "Failed to remove filter: " + err.Error(),
				Level:    events.ToastError,
				Duration: 5 * time.Second,
			})
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			break
		}
		m.syncFilterpane()
		cmds = append(cmds, saveSessionCmd(m.ctrl))

	case events.ClearFiltersRequested:
		m.ctrl.ClearFilters()
		m.syncFilterpane()
		cmds = append(cmds, saveSessionCmd(m.ctrl))

	case events.FilterPresetsLoadedMsg:
		m.filterPresets = msg.Presets
	case events.FilterPresetsLoadFailedMsg:
		m.status, cmd = m.status.Update(events.ShowToastMsg{
			Text:     "Failed to load filter presets: " + msg.Err.Error(),
			Level:    events.ToastWarning,
			Duration: 5 * time.Second,
		})
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	// Let status model consume messages too (toast expiry, etc.).
	beforeFooterH := lipgloss.Height(m.status.View())
	m.status, cmd = m.status.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	afterFooterH := lipgloss.Height(m.status.View())
	if beforeFooterH != afterFooterH {
		m.recomputeLayout()
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

func (m Model) applyQuickPresetByIndex(i int) Model {
	if i < 0 || i >= len(m.filterPresets) {
		m.status, _ = m.status.Update(events.ShowToastMsg{
			Text:     "No preset bound to that key",
			Level:    events.ToastInfo,
			Duration: 2 * time.Second,
		})
		return m
	}

	p := m.filterPresets[i]
	preset, err := m.ctrl.LoadFilterPresetByID(p.ID)
	if err != nil {
		m.status, _ = m.status.Update(events.ShowToastMsg{
			Text:     "Failed to apply preset: " + err.Error(),
			Level:    events.ToastError,
			Duration: 5 * time.Second,
		})
		return m
	}

	m.ctrl.AddFilter(domain.Filter{
		Name:        preset.Name,
		Description: preset.Description,
		Rules:       preset.Rules,
	})
	m.status, _ = m.status.Update(events.ShowToastMsg{
		Text:     "Applied preset: " + preset.ID,
		Level:    events.ToastSuccess,
		Duration: 2 * time.Second,
	})
	return m
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
		desc, err := ctrl.GenerateDescription(context.Background())
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

		// Keep clipboard output markdown-ish for human usability, but reuse the shared exporter.
		text := pexport.BuildGenerationContext(req, pexport.SeparatorMarkdown)
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
	case ModeFilters:
		// renderFilters writes: title + blank + stats + blank + header + separator.
		return 6
	case ModeResult:
		// renderResult writes: title line + "\n\n" (=> 3 lines total before the viewport)
		return 3
	case ModeGenerating:
		// renderGenerating uses the same overall structure as main for now.
		return 8
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

func (m *Model) syncFilterpane() {
	idx := m.filterpane.SelectedIndex()
	filters := m.ctrl.GetFilters()
	m.filterpane.SetFilters(filters)
	m.filterpane.SetSelectedIndex(idx)
	m.filterIndex = m.filterpane.SelectedIndex()
}

// (Quick presets are now loaded from preset dirs at runtime; see loadFilterPresetsCmd.)
