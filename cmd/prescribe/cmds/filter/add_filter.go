package filter

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/go-go-golems/prescribe/internal/controller"
	"github.com/go-go-golems/prescribe/internal/domain"
)

var (
	filterName        string
	filterDescription string
	excludePatterns   []string
	includePatterns   []string
)

var AddFilterCmd = &cobra.Command{
	Use:   "add-filter",
	Short: "Add a filter to the session",
	Long:  `Add a file filter to the current session.`,
	RunE: func(cmdCmd *cobra.Command, args []string) error {
		if filterName == "" {
			return fmt.Errorf("filter name is required (--name)")
		}
		
		if len(excludePatterns) == 0 && len(includePatterns) == 0 {
			return fmt.Errorf("at least one pattern is required (--exclude or --include)")
		}
		
		// Get flags from parent command
		repoPath, _ := cmdCmd.Flags().GetString("repo")
		targetBranch, _ := cmdCmd.Flags().GetString("target")
		if repoPath == "" {
			repoPath = "."
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
	AddFilterCmd.Flags().StringVarP(&filterName, "name", "n", "", "Filter name (required)")
	AddFilterCmd.Flags().StringVarP(&filterDescription, "description", "d", "", "Filter description")
	AddFilterCmd.Flags().StringSliceVarP(&excludePatterns, "exclude", "e", []string{}, "Exclude patterns (can specify multiple)")
	AddFilterCmd.Flags().StringSliceVarP(&includePatterns, "include", "i", []string{}, "Include patterns (can specify multiple)")
	AddFilterCmd.MarkFlagRequired("name")
}

