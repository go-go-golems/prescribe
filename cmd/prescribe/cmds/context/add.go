package context

import (
	"fmt"

	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/helpers"
	"github.com/spf13/cobra"
)

var (
	contextNote string
)

var AddCmd = &cobra.Command{
	Use:   "add [file-path]",
	Short: "Add additional context to session",
	Long:  "Add a file or note as additional context for PR description generation.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmdCmd *cobra.Command, args []string) error {
		if contextNote == "" && len(args) == 0 {
			return fmt.Errorf("either pass a file path argument or use --note")
		}
		if contextNote != "" && len(args) > 0 {
			return fmt.Errorf("use either a file path argument or --note (not both)")
		}

		ctrl, err := helpers.NewInitializedController(cmdCmd)
		if err != nil {
			return err
		}

		helpers.LoadDefaultSessionIfExists(ctrl)

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
	AddCmd.Flags().StringVar(&contextNote, "note", "", "Add a note as context")
}
