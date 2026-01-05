package context

import (
	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/context/git"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// NewContextCmd groups all additional-context related subcommands.
func NewContextCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "context",
		Short: "Manage additional context",
		Long:  "Add additional context (files and notes) to the current session.",
	}

	addCmd, err := NewAddCobraCommand()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build context add command")
	}

	gitCmd, err := git.NewGitCmd()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build context git command")
	}

	cmd.AddCommand(addCmd, gitCmd)
	return cmd, nil
}
