package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/user/pr-builder/internal/controller"
)

var (
	sessionPath string
	autoSave    bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new PR builder session",
	Long:  `Initialize a new PR builder session from the current git state.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create controller
		ctrl, err := controller.NewController(repoPath)
		if err != nil {
			return fmt.Errorf("failed to create controller: %w", err)
		}
		
		// Initialize from git
		if err := ctrl.Initialize(targetBranch); err != nil {
			return fmt.Errorf("failed to initialize: %w", err)
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

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVarP(&sessionPath, "output", "o", "", "Session file path (default: .pr-builder/session.yaml)")
	initCmd.Flags().BoolVarP(&autoSave, "save", "s", false, "Automatically save session after init")
}
