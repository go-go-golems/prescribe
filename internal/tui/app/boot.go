package app

import (
	"errors"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/prescribe/internal/controller"
	"github.com/go-go-golems/prescribe/internal/tui/events"
)

// bootCmd attempts to load the default session at startup.
//
// Semantics:
// - missing file: ignored (SessionLoadSkippedMsg)
// - other errors: SessionLoadFailedMsg (the app should toast this)
// - success: SessionLoadedMsg
func bootCmd(ctrl *controller.Controller) tea.Cmd {
	return func() tea.Msg {
		path := ctrl.GetDefaultSessionPath()
		if err := ctrl.LoadSession(path); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				// New session: apply repo defaults (if configured).
				n, err2 := ctrl.ApplyDefaultFilterPresetsFromRepoConfig()
				if err2 != nil {
					return events.DefaultFiltersApplyFailedMsg{Err: err2}
				}
				if n > 0 {
					return events.DefaultFiltersAppliedMsg{Count: n}
				}
				return events.SessionLoadSkippedMsg{Path: path}
			}
			return events.SessionLoadFailedMsg{Path: path, Err: err}
		}
		return events.SessionLoadedMsg{Path: path}
	}
}
