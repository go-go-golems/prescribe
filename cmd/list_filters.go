package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/user/pr-builder/internal/controller"
)

var listFiltersCmd = &cobra.Command{
	Use:   "list-filters",
	Short: "List all active filters",
	Long:  `Display all active filters in the current session.`,
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

		// Load session if exists
		sessionPath := ctrl.GetDefaultSessionPath()
		if err := ctrl.LoadSession(sessionPath); err == nil {
			// Session loaded
		}

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

func init() {
	rootCmd.AddCommand(listFiltersCmd)
}
