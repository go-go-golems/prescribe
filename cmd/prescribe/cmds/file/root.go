package file

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// NewFileCmd groups all file-related subcommands.
func NewFileCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "file",
		Short: "File operations",
		Long:  "Operations that act on individual changed files in the current session.",
	}

	toggleCmd, err := NewToggleCobraCommand()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build file toggle command")
	}

	cmd.AddCommand(toggleCmd)
	return cmd, nil
}
