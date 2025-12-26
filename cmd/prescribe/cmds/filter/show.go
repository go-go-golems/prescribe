package filter

import (
	"fmt"

	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/helpers"
	"github.com/spf13/cobra"
)

var ShowFilteredCmd = &cobra.Command{
	Use:   "show",
	Short: "Show files that are filtered out",
	Long:  `Display all files that are being filtered out by active filters.`,
	RunE: func(cmdCmd *cobra.Command, args []string) error {
		ctrl, err := helpers.NewInitializedController(cmdCmd)
		if err != nil {
			return err
		}

		helpers.LoadDefaultSessionIfExists(ctrl)

		// Get filtered files
		filtered := ctrl.GetFilteredFiles()
		visible := ctrl.GetVisibleFiles()
		data := ctrl.GetData()

		fmt.Printf("File Status\n")
		fmt.Println("==================")
		fmt.Printf("Total files: %d\n", len(data.ChangedFiles))
		fmt.Printf("Visible files: %d\n", len(visible))
		fmt.Printf("Filtered files: %d\n\n", len(filtered))

		if len(filtered) == 0 {
			fmt.Println("No files are being filtered out")
			return nil
		}

		fmt.Printf("Filtered Files:\n")
		for _, file := range filtered {
			fmt.Printf("  ✗ %s (+%d -%d, %dt)\n",
				file.Path,
				file.Additions,
				file.Deletions,
				file.Tokens)
		}

		// Show which filters are active
		filters := ctrl.GetFilters()
		if len(filters) > 0 {
			fmt.Printf("\nActive Filters:\n")
			for i, filter := range filters {
				fmt.Printf("  [%d] %s\n", i, filter.Name)
				for _, rule := range filter.Rules {
					fmt.Printf("      %s: %s\n", rule.Type, rule.Pattern)
				}
			}
		}

		return nil
	},
}
