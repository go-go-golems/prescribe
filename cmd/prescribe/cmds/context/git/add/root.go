package add

import "github.com/spf13/cobra"

// NewAddCmd groups all git_context "add" verbs.
func NewAddCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a git_context item",
	}

	cmd.AddCommand(
		newCommitCmd(),
		newCommitPatchCmd(),
		newFileAtCmd(),
		newFileDiffCmd(),
	)

	return cmd, nil
}
