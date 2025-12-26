package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/user/pr-builder/internal/controller"
)

var clearFiltersCmd = &cobra.Command{
	Use:   "clear-filters",
	Short: "Remove all filters from the session",
	Long:  `Remove all active filters, making all files visible.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create controller
		ctrl, err := controller.NewController(repoPath)
		if err != nil {
			return fmt.Errorf("failed to create controller: %w", err)
		}

		// Initialize
		if err := ctrl.Initialize(targetBranch); err != nil {
			return fmt.Errorf("failed to initialize: %w", err)
		}

		// Load session
		sessionPath := ctrl.GetDefaultSessionPath()
		if err := ctrl.LoadSession(sessionPath); err != nil {
			return fmt.Errorf("failed to load session: %w", err)
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
		if err := ctrl.SaveSession(sessionPath); err != nil {
			return fmt.Errorf("failed to save session: %w", err)
		}

		fmt.Printf("Cleared %d filter(s) and saved session\n", filterCount)

		// Show new state
		data := ctrl.GetData()
		fmt.Printf("  All files now visible: %d\n", len(data.ChangedFiles))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(clearFiltersCmd)
}
