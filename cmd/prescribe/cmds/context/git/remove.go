package git

import (
	"fmt"
	"strconv"

	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/helpers"
	"github.com/spf13/cobra"
)

func newGitContextRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <index>",
		Short: "Remove a git_context item by index",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			idx, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid index %q", args[0])
			}

			ctrl, err := helpers.NewInitializedController(cmd)
			if err != nil {
				return err
			}
			helpers.LoadDefaultSessionIfExists(ctrl)

			data := ctrl.GetData()
			if idx < 0 || idx >= len(data.GitContext) {
				return fmt.Errorf("index out of range: %d", idx)
			}

			data.GitContext = append(data.GitContext[:idx], data.GitContext[idx+1:]...)
			savePath := ctrl.GetDefaultSessionPath()
			if err := ctrl.SaveSession(savePath); err != nil {
				return err
			}
			fmt.Printf("Removed git_context item %d\n", idx)
			fmt.Printf("Session saved: %s\n", savePath)
			return nil
		},
	}
}
