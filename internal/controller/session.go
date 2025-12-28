package controller

import (
	"fmt"

	"github.com/go-go-golems/prescribe/internal/session"
)

// SaveSession saves the current state to a session file
func (c *Controller) SaveSession(path string) error {
	sess := session.NewSession(c.data)
	return sess.Save(path)
}

// LoadSession loads a session file and applies it to the current state
func (c *Controller) LoadSession(path string) error {
	sess, err := session.Load(path)
	if err != nil {
		return fmt.Errorf("failed to load session: %w", err)
	}

	// Verify branches match
	if sess.SourceBranch != c.data.SourceBranch {
		return fmt.Errorf("session source branch (%s) doesn't match current branch (%s)",
			sess.SourceBranch, c.data.SourceBranch)
	}

	// Apply session to data
	if err := sess.ApplyToData(c.data); err != nil {
		return fmt.Errorf("failed to apply session: %w", err)
	}

	return nil
}

// GetDefaultSessionPath returns the default session path
func (c *Controller) GetDefaultSessionPath() string {
	return session.GetDefaultSessionPath(c.repoPath)
}
