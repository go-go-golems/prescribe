package file

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/go-go-golems/prescribe/internal/controller"
)

var (
	contextNote string
)

var AddContextCmd = &cobra.Command{
	Use:   "add-context [file-path]",
	Short: "Add additional context to session",
	Long:  `Add a file or note as additional context for PR description generation.`,
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
		
		// Add context
		if contextNote != "" {
			// Add note
			ctrl.AddContextNote(contextNote)
			fmt.Printf("Added note to context\n")
		} else {
			// Add file
			filePath := args[0]
			if err := ctrl.AddContextFile(filePath); err != nil {
				return fmt.Errorf("failed to add file: %w", err)
			}
			fmt.Printf("Added file '%s' to context\n", filePath)
		}
		
		// Save session
		savePath := ctrl.GetDefaultSessionPath()
		if err := ctrl.SaveSession(savePath); err != nil {
			return fmt.Errorf("failed to save session: %w", err)
		}
		
		fmt.Printf("Session saved\n")
		
		// Show token count
		data := ctrl.GetData()
		fmt.Printf("Total tokens: %d\n", data.GetTotalTokens())
		
		return nil
	},
}

func init() {
	AddContextCmd.Flags().StringVar(&contextNote, "note", "", "Add a note as context")
}

