package git

import (
	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/context/git/add"
	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/context/git/history"
	"github.com/spf13/cobra"
)

// NewGitCmd groups all git-derived context subcommands.
func NewGitCmd() (*cobra.Command, error) {
	gitCmd := &cobra.Command{
		Use:   "git",
		Short: "Manage git-derived context",
		Long:  "Manage git-derived context (history and explicit git artifacts) for the current session.",
	}

	historyCmd, err := history.NewHistoryCmd()
	if err != nil {
		return nil, err
	}

	gitCmd.AddCommand(historyCmd)

	addCmd, err := add.NewAddCmd()
	if err != nil {
		return nil, err
	}
	gitCmd.AddCommand(
		newGitContextListCmd(),
		newGitContextRemoveCmd(),
		newGitContextClearCmd(),
		addCmd,
	)
	return gitCmd, nil
}
