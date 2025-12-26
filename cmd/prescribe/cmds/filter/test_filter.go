package filter

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/go-go-golems/prescribe/internal/controller"
	"github.com/go-go-golems/prescribe/internal/domain"
)

var (
	testFilterName        string
	testExcludePatterns   []string
	testIncludePatterns   []string
)

var TestFilterCmd = &cobra.Command{
	Use:   "test-filter",
	Short: "Test a filter pattern without applying it",
	Long:  `Test how a filter would affect files without actually applying it to the session.`,
	RunE: func(cmdCmd *cobra.Command, args []string) error {
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
		for i, pattern := range testExcludePatterns {
			rules = append(rules, domain.FilterRule{
				Type:    domain.FilterTypeExclude,
				Pattern: pattern,
				Order:   i,
			})
		}
		for i, pattern := range testIncludePatterns {
			rules = append(rules, domain.FilterRule{
				Type:    domain.FilterTypeInclude,
				Pattern: pattern,
				Order:   len(testExcludePatterns) + i,
			})
		}

		// Create test filter
		filter := domain.Filter{
			Name:  testFilterName,
			Rules: rules,
		}

		// Test filter
		matched, unmatched := ctrl.TestFilter(filter)

		// Display results
		fmt.Printf("Filter Test: %s\n", testFilterName)
		fmt.Println("==================")
		
		fmt.Printf("\nRules:\n")
		for _, rule := range rules {
			fmt.Printf("  %s: %s\n", rule.Type, rule.Pattern)
		}

		fmt.Printf("\nMatched Files (%d):\n", len(matched))
		for _, path := range matched {
			fmt.Printf("  ✓ %s\n", path)
		}

		fmt.Printf("\nFiltered Files (%d):\n", len(unmatched))
		for _, path := range unmatched {
			fmt.Printf("  ✗ %s\n", path)
		}

		fmt.Printf("\nSummary:\n")
		fmt.Printf("  Total files: %d\n", len(matched)+len(unmatched))
		fmt.Printf("  Would be visible: %d\n", len(matched))
		fmt.Printf("  Would be filtered: %d\n", len(unmatched))

		return nil
	},
}

