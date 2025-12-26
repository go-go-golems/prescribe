package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/user/pr-builder/internal/controller"
	"github.com/user/pr-builder/internal/domain"
)

var (
	filterName        string
	filterDescription string
	excludePatterns   []string
	includePatterns   []string
)

var addFilterCmd = &cobra.Command{
	Use:   "add-filter",
	Short: "Add a filter to the session",
	Long:  `Add a file filter to the current session.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if filterName == "" {
			return fmt.Errorf("filter name is required (--name)")
		}
		
		if len(excludePatterns) == 0 && len(includePatterns) == 0 {
			return fmt.Errorf("at least one pattern is required (--exclude or --include)")
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
		
		// Build filter rules
		rules := make([]domain.FilterRule, 0)
		for i, pattern := range excludePatterns {
			rules = append(rules, domain.FilterRule{
				Type:    domain.FilterTypeExclude,
				Pattern: pattern,
				Order:   i,
			})
		}
		for i, pattern := range includePatterns {
			rules = append(rules, domain.FilterRule{
				Type:    domain.FilterTypeInclude,
				Pattern: pattern,
				Order:   len(excludePatterns) + i,
			})
		}
		
		// Create and add filter
		filter := domain.Filter{
			Name:        filterName,
			Description: filterDescription,
			Rules:       rules,
		}
		
		ctrl.AddFilter(filter)
		
		// Save session
		savePath := ctrl.GetDefaultSessionPath()
		if err := ctrl.SaveSession(savePath); err != nil {
			return fmt.Errorf("failed to save session: %w", err)
		}
		
		fmt.Printf("Filter '%s' added and saved to session\n", filterName)
		
		// Show impact
		data := ctrl.GetData()
		fmt.Printf("  Files now filtered: %d\n", len(data.GetFilteredFiles()))
		
		return nil
	},
}

func init() {
	rootCmd.AddCommand(addFilterCmd)
	addFilterCmd.Flags().StringVarP(&filterName, "name", "n", "", "Filter name (required)")
	addFilterCmd.Flags().StringVarP(&filterDescription, "description", "d", "", "Filter description")
	addFilterCmd.Flags().StringSliceVarP(&excludePatterns, "exclude", "e", []string{}, "Exclude patterns (can specify multiple)")
	addFilterCmd.Flags().StringSliceVarP(&includePatterns, "include", "i", []string{}, "Include patterns (can specify multiple)")
	addFilterCmd.MarkFlagRequired("name")
}
