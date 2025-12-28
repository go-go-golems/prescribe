package filter

import (
	"sync"

	"github.com/spf13/cobra"
)

// FilterCmd groups all filter-related subcommands.
var FilterCmd = &cobra.Command{
	Use:   "filter",
	Short: "Manage file filters",
	Long:  "Create, test, list, and manage file filters in the current session.",
}

var initOnce sync.Once
var initErr error

func Init() error {
	initOnce.Do(func() {
		if err := InitAddFilterCmd(); err != nil {
			initErr = err
			return
		}
		if err := InitPresetCmd(); err != nil {
			initErr = err
			return
		}
		if err := InitRemoveFilterCmd(); err != nil {
			initErr = err
			return
		}
		if err := InitClearFiltersCmd(); err != nil {
			initErr = err
			return
		}
		if err := InitTestFilterCmd(); err != nil {
			initErr = err
			return
		}
		if err := InitListFiltersCmd(); err != nil {
			initErr = err
			return
		}
		if err := InitShowFilteredCmd(); err != nil {
			initErr = err
			return
		}

		FilterCmd.AddCommand(
			AddFilterCmd,
			PresetCmd,
			ListFiltersCmd,
			RemoveFilterCmd,
			ClearFiltersCmd,
			TestFilterCmd,
			ShowFilteredCmd,
		)
	})
	return initErr
}
