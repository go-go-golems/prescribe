package helpers

import (
	"fmt"

	"github.com/go-go-golems/prescribe/internal/controller"
)

// LoadDefaultSessionIfExists attempts to load the default session and ignores errors
// (useful for commands that should work even if no session exists yet).
func LoadDefaultSessionIfExists(ctrl *controller.Controller) {
	sessionPath := ctrl.GetDefaultSessionPath()
	_ = ctrl.LoadSession(sessionPath)
}

// LoadDefaultSession loads the default session and returns an error if it fails.
func LoadDefaultSession(ctrl *controller.Controller) error {
	sessionPath := ctrl.GetDefaultSessionPath()
	if err := ctrl.LoadSession(sessionPath); err != nil {
		return fmt.Errorf("failed to load session: %w", err)
	}
	return nil
}
