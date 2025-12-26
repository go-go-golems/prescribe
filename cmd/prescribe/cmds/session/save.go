package session

import (
	"fmt"

	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/helpers"
	"github.com/spf13/cobra"
)

var SaveCmd = &cobra.Command{
	Use:   "save [path]",
	Short: "Save current session to YAML file",
	Long:  `Save the current PR builder session to a YAML file.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmdCmd *cobra.Command, args []string) error {
		ctrl, err := helpers.NewInitializedController(cmdCmd)
		if err != nil {
			return err
		}

		// Prefer loading an existing session so this command reflects current state.
		helpers.LoadDefaultSessionIfExists(ctrl)

		// Determine save path
		savePath := ctrl.GetDefaultSessionPath()
		if len(args) > 0 {
			savePath = args[0]
		}

		// Save session
		if err := ctrl.SaveSession(savePath); err != nil {
			return fmt.Errorf("failed to save session: %w", err)
		}

		fmt.Printf("Session saved to: %s\n", savePath)
		return nil
	},
}
