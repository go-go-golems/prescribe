package session

import (
	"fmt"

	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/helpers"
	"github.com/spf13/cobra"
)

var (
	sessionPath string
	autoSave    bool
)

var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new PR builder session",
	Long:  `Initialize a new PR builder session from the current git state.`,
	RunE: func(cmdCmd *cobra.Command, args []string) error {
		ctrl, err := helpers.NewInitializedController(cmdCmd)
		if err != nil {
			return err
		}

		data := ctrl.GetData()

		fmt.Printf("Initialized PR builder session\n")
		fmt.Printf("  Source: %s\n", data.SourceBranch)
		fmt.Printf("  Target: %s\n", data.TargetBranch)
		fmt.Printf("  Files: %d\n", len(data.ChangedFiles))

		// Auto-save if requested
		if autoSave {
			savePath := sessionPath
			if savePath == "" {
				savePath = ctrl.GetDefaultSessionPath()
			}

			if err := ctrl.SaveSession(savePath); err != nil {
				return fmt.Errorf("failed to save session: %w", err)
			}

			fmt.Printf("\nSession saved to: %s\n", savePath)
		}

		return nil
	},
}

func InitInitCmd() error {
	InitCmd.Flags().BoolVar(&autoSave, "save", false, "Save session to disk after initialization")
	InitCmd.Flags().StringVarP(&sessionPath, "path", "p", "", "Path to save session (default: app default session path)")
	return nil
}
