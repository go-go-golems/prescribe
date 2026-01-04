package add

import "github.com/spf13/cobra"

// NewAddCmd groups all git_context "add" verbs.
func NewAddCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a git_context item",
	}

	commitCmd, err := NewCommitCobraCommand()
	if err != nil {
		return nil, err
	}
	commitPatchCmd, err := NewCommitPatchCobraCommand()
	if err != nil {
		return nil, err
	}
	fileAtCmd, err := NewFileAtCobraCommand()
	if err != nil {
		return nil, err
	}
	fileDiffCmd, err := NewFileDiffCobraCommand()
	if err != nil {
		return nil, err
	}

	cmd.AddCommand(commitCmd, commitPatchCmd, fileAtCmd, fileDiffCmd)

	return cmd, nil
}
