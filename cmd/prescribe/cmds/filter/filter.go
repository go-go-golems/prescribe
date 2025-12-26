package filter

import "github.com/spf13/cobra"

// FilterCmd groups all filter-related subcommands.
var FilterCmd = &cobra.Command{
	Use:   "filter",
	Short: "Manage file filters",
	Long:  "Create, test, list, and manage file filters in the current session.",
}

func init() {
	FilterCmd.AddCommand(
		AddFilterCmd,
		ListFiltersCmd,
		RemoveFilterCmd,
		ClearFiltersCmd,
		TestFilterCmd,
		ShowFilteredCmd,
	)
}
