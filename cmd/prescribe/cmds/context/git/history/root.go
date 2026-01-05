package history

import "github.com/spf13/cobra"

// NewHistoryCmd groups all derived-git-history verbs.
func NewHistoryCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "history",
		Short: "Manage derived git history",
		Long:  "Show and configure derived git history inclusion for generation (stored in session.yaml).",
	}

	showCmd, err := NewShowCobraCommand()
	if err != nil {
		return nil, err
	}
	enableCmd, err := NewEnableCobraCommand()
	if err != nil {
		return nil, err
	}
	disableCmd, err := NewDisableCobraCommand()
	if err != nil {
		return nil, err
	}
	setCmd, err := NewSetCobraCommand()
	if err != nil {
		return nil, err
	}

	cmd.AddCommand(showCmd, enableCmd, disableCmd, setCmd)

	return cmd, nil
}
