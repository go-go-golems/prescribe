package history

import "github.com/spf13/cobra"

// NewHistoryCmd groups all derived-git-history verbs.
func NewHistoryCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "history",
		Short: "Manage derived git history",
		Long:  "Show and configure derived git history inclusion for generation (stored in session.yaml).",
	}

	cmd.AddCommand(
		newShowCmd(),
		newEnableCmd(),
		newDisableCmd(),
		newSetCmd(),
	)

	return cmd, nil
}
