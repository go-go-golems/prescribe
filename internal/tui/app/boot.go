package app

import (
	"errors"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/prescribe/internal/controller"
	"github.com/go-go-golems/prescribe/internal/tui/events"
)

// bootCmd attempts to load the default session at startup.
//
// Semantics:
// - missing file: SessionLoadFailedMsg (TUI requires an initialized session)
// - other errors: SessionLoadFailedMsg (the app should toast this)
// - success: SessionLoadedMsg
func bootCmd(ctrl *controller.Controller) tea.Cmd {
	return func() tea.Msg {
		path := ctrl.GetDefaultSessionPath()
		if err := ctrl.LoadSession(path); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return events.SessionLoadFailedMsg{
					Path: path,
					Err:  fmt.Errorf("no session found; run 'prescribe session init --save' first"),
				}
			}
			return events.SessionLoadFailedMsg{Path: path, Err: err}
		}
		return events.SessionLoadedMsg{Path: path}
	}
}
