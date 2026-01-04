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

	listCmd, err := NewListCobraCommand()
	if err != nil {
		return nil, err
	}
	removeCmd, err := NewRemoveCobraCommand()
	if err != nil {
		return nil, err
	}
	clearCmd, err := NewClearCobraCommand()
	if err != nil {
		return nil, err
	}
	gitCmd.AddCommand(
		listCmd,
		removeCmd,
		clearCmd,
		addCmd,
	)
	return gitCmd, nil
}
