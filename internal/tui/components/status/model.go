package status

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/help"
	"github.com/go-go-golems/prescribe/internal/tui/events"
	"github.com/go-go-golems/prescribe/internal/tui/keys"
	"github.com/go-go-golems/prescribe/internal/tui/styles"
)

// Model is a small footer model that renders help and a transient toast.
//
// Phase 1 goal: have a compiling, testable toast state machine and a place to render help.
// Phase 2+ will wire this into the new app root model.
type Model struct {
	help help.Model

	keymap keys.KeyMap
	styles styles.Styles

	showFullHelp bool
	width        int

	toast ToastState
}

func New(km keys.KeyMap, st styles.Styles) Model {
	return Model{
		help:   help.New(),
		keymap: km,
		styles: st,
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) SetSize(width int) {
	m.width = width
}

func (m Model) SetShowFullHelp(v bool) {
	m.showFullHelp = v
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case events.ShowToastMsg:
		id := m.toast.Show(msg.Text, msg.Level)
		if msg.Duration <= 0 {
			// 0 duration means "show until replaced" (no auto-expire).
			return m, nil
		}
		return m, tea.Tick(msg.Duration, func(time.Time) tea.Msg {
			return events.ToastExpiredMsg{ID: id}
		})

	case events.ToastExpiredMsg:
		_ = m.toast.Expire(msg.ID)
		return m, nil
	}

	return m, nil
}

func (m Model) View() string {
	var toastLine string
	if t := m.toast.Current(); t != nil {
		toastLine = m.renderToast(*t)
	}

	h := m.help
	h.Width = m.width
	h.ShowAll = m.showFullHelp
	helpView := h.View(m.keymap)

	if toastLine == "" {
		return helpView
	}
	return toastLine + "\n" + helpView
}

func (m Model) renderToast(t Toast) string {
	switch t.Level {
	case events.ToastSuccess:
		return m.styles.SuccessText.Render(t.Text)
	case events.ToastWarning:
		return m.styles.WarningText.Render(t.Text)
	case events.ToastError:
		return m.styles.ErrorText.Render(t.Text)
	case events.ToastInfo:
		fallthrough
	default:
		return m.styles.MutedText.Render(t.Text)
	}
}


