package filter

import (
	"fmt"

	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/helpers"
	"github.com/spf13/cobra"
)

var ClearFiltersCmd = &cobra.Command{
	Use:   "clear",
	Short: "Remove all filters from the session",
	Long:  `Remove all active filters, making all files visible.`,
	RunE: func(cmdCmd *cobra.Command, args []string) error {
		ctrl, err := helpers.NewInitializedController(cmdCmd)
		if err != nil {
			return err
		}

		if err := helpers.LoadDefaultSession(ctrl); err != nil {
			return err
		}

		// Get current filter count
		filterCount := len(ctrl.GetFilters())
		if filterCount == 0 {
			fmt.Println("No filters to clear")
			return nil
		}

		// Clear filters
		ctrl.ClearFilters()

		// Save session
		savePath := ctrl.GetDefaultSessionPath()
		if err := ctrl.SaveSession(savePath); err != nil {
			return fmt.Errorf("failed to save session: %w", err)
		}

		fmt.Printf("Cleared %d filter(s) and saved session\n", filterCount)

		// Show new state
		data := ctrl.GetData()
		fmt.Printf("  All files now visible: %d\n", len(data.ChangedFiles))

		return nil
	},
}
