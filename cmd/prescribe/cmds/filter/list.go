package filter

import (
	"fmt"

	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/helpers"
	"github.com/spf13/cobra"
)

var ListFiltersCmd = &cobra.Command{
	Use:   "list",
	Short: "List all active filters",
	Long:  `Display all active filters in the current session.`,
	RunE: func(cmdCmd *cobra.Command, args []string) error {
		ctrl, err := helpers.NewInitializedController(cmdCmd)
		if err != nil {
			return err
		}

		helpers.LoadDefaultSessionIfExists(ctrl)

		// Get filters
		filters := ctrl.GetFilters()

		if len(filters) == 0 {
			fmt.Println("No active filters")
			return nil
		}

		fmt.Printf("Active Filters (%d)\n", len(filters))
		fmt.Println("==================")

		for i, filter := range filters {
			fmt.Printf("\n[%d] %s\n", i, filter.Name)
			if filter.Description != "" {
				fmt.Printf("    Description: %s\n", filter.Description)
			}
			fmt.Printf("    Rules: %d\n", len(filter.Rules))
			for j, rule := range filter.Rules {
				fmt.Printf("      [%d] %s: %s\n", j, rule.Type, rule.Pattern)
			}
		}

		// Show impact
		data := ctrl.GetData()
		fmt.Printf("\nImpact:\n")
		fmt.Printf("  Total files: %d\n", len(data.ChangedFiles))
		fmt.Printf("  Visible files: %d\n", len(data.GetVisibleFiles()))
		fmt.Printf("  Filtered files: %d\n", len(data.GetFilteredFiles()))

		return nil
	},
}
