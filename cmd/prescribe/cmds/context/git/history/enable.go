package history

import (
	"fmt"

	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/helpers"
	"github.com/spf13/cobra"
)

func newEnableCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "enable",
		Short: "Enable derived git history",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctrl, err := helpers.NewInitializedController(cmd)
			if err != nil {
				return err
			}
			helpers.LoadDefaultSessionIfExists(ctrl)

			data := ctrl.GetData()
			cfg, _ := effectiveGitHistoryConfig(data)
			cfg.Enabled = true
			data.GitHistory = &cfg

			savePath := ctrl.GetDefaultSessionPath()
			if err := ctrl.SaveSession(savePath); err != nil {
				return err
			}
			fmt.Printf("Git history enabled\n")
			fmt.Printf("Session saved: %s\n", savePath)
			return nil
		},
	}
}
