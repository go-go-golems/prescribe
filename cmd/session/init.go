package session

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/go-go-golems/prescribe/internal/controller"
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

