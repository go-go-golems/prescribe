package filter

import (
	"fmt"
	"strconv"

	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/helpers"
	"github.com/spf13/cobra"
)

var RemoveFilterCmd = &cobra.Command{
	Use:   "remove <index|name>",
	Short: "Remove a filter from the session",
	Long:  `Remove a filter by index or name.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmdCmd *cobra.Command, args []string) error {
		ctrl, err := helpers.NewInitializedController(cmdCmd)
		if err != nil {
			return err
		}

		if err := helpers.LoadDefaultSession(ctrl); err != nil {
			return err
		}

		// Get filters
		filters := ctrl.GetFilters()
		if len(filters) == 0 {
			return fmt.Errorf("no filters to remove")
		}

		// Try to parse as index
		index, err := strconv.Atoi(args[0])
		if err != nil {
			// Not a number, try to find by name
			found := false
			for i, filter := range filters {
				if filter.Name == args[0] {
					index = i
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("filter not found: %s", args[0])
			}
		}

		// Validate index
		if index < 0 || index >= len(filters) {
			return fmt.Errorf("invalid filter index: %d (valid range: 0-%d)", index, len(filters)-1)
		}

		// Get filter name before removing
		filterName := filters[index].Name

		// Remove filter
		if err := ctrl.RemoveFilter(index); err != nil {
			return fmt.Errorf("failed to remove filter: %w", err)
		}

		// Save session
		savePath := ctrl.GetDefaultSessionPath()
		if err := ctrl.SaveSession(savePath); err != nil {
			return fmt.Errorf("failed to save session: %w", err)
		}

		fmt.Printf("Filter '%s' removed and session saved\n", filterName)

		// Show new impact
		data := ctrl.GetData()
		fmt.Printf("  Visible files: %d\n", len(data.GetVisibleFiles()))
		fmt.Printf("  Filtered files: %d\n", len(data.GetFilteredFiles()))

		return nil
	},
}
