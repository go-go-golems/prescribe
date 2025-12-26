package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/user/pr-builder/internal/controller"
)

var saveCmd = &cobra.Command{
	Use:   "save [path]",
	Short: "Save current session to YAML file",
	Long:  `Save the current PR builder session to a YAML file.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
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

func init() {
	rootCmd.AddCommand(saveCmd)
}
