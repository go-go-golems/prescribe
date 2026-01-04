package git

import (
	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/context/git/add"
	"github.com/spf13/cobra"
)

// NewGitCmd groups all git-derived context subcommands.
func NewGitCmd() (*cobra.Command, error) {
	gitCmd := &cobra.Command{
		Use:   "git",
		Short: "Manage git-derived context",
		Long:  "Manage git-derived context (history and explicit git artifacts) for the current session.",
	}

	historyCmd := &cobra.Command{
		Use:   "history",
		Short: "Manage derived git history",
		Long:  "Show and configure derived git history inclusion for generation (stored in session.yaml).",
	}

	historyCmd.AddCommand(
		newGitHistoryShowCmd(),
		newGitHistoryEnableCmd(true),
		newGitHistoryEnableCmd(false),
		newGitHistorySetCmd(),
	)

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
