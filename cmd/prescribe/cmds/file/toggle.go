package file

import (
	"fmt"

	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/helpers"
	"github.com/spf13/cobra"
)

var ToggleFileCmd = &cobra.Command{
	Use:   "toggle <path>",
	Short: "Toggle file inclusion in session",
	Long:  `Toggle whether a file is included in the PR description context.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmdCmd *cobra.Command, args []string) error {
		filePath := args[0]

		ctrl, err := helpers.NewInitializedController(cmdCmd)
		if err != nil {
			return err
		}

		helpers.LoadDefaultSessionIfExists(ctrl)

		// Find file and toggle
		data := ctrl.GetData()
		found := false
		for i, file := range data.ChangedFiles {
			if file.Path == filePath {
				if err := ctrl.ToggleFileInclusion(i); err != nil {
					return fmt.Errorf("failed to toggle file: %w", err)
				}
				found = true

				// Show new state
				newState := "excluded"
				if data.ChangedFiles[i].Included {
					newState = "included"
				}
				fmt.Printf("File '%s' is now %s\n", filePath, newState)
				break
			}
		}

		if !found {
			return fmt.Errorf("file not found: %s", filePath)
		}

		// Save session
		savePath := ctrl.GetDefaultSessionPath()
		if err := ctrl.SaveSession(savePath); err != nil {
			return fmt.Errorf("failed to save session: %w", err)
		}

		fmt.Printf("Session saved\n")

		return nil
	},
}
