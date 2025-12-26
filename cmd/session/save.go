package session

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/go-go-golems/prescribe/internal/controller"
)

var SaveCmd = &cobra.Command{
	Use:   "save [path]",
	Short: "Save current session to YAML file",
	Long:  `Save the current PR builder session to a YAML file.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmdCmd *cobra.Command, args []string) error {
		// Get flags from parent command
		repoPath, _ := cmdCmd.Flags().GetString("repo")
		targetBranch, _ := cmdCmd.Flags().GetString("target")
		if repoPath == "" {
			repoPath = "."
		}
		// Create controller
		ctrl, err := controller.NewController(repoPath)
		if err != nil {
			return fmt.Errorf("failed to create controller: %w", err)
		}
		
		// Initialize
		if err := ctrl.Initialize(targetBranch); err != nil {
			return fmt.Errorf("failed to initialize: %w", err)
		}
		
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

