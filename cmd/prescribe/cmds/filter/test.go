package filter

import (
	"fmt"

	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/helpers"
	"github.com/go-go-golems/prescribe/internal/domain"
	"github.com/spf13/cobra"
)

var (
	testFilterName      string
	testExcludePatterns []string
	testIncludePatterns []string
)

var TestFilterCmd = &cobra.Command{
	Use:   "test",
	Short: "Test a filter pattern without applying it",
	Long:  `Test how a filter would affect files without actually applying it to the session.`,
	RunE: func(cmdCmd *cobra.Command, args []string) error {
		if testFilterName == "" {
			testFilterName = "test"
		}

		if len(testExcludePatterns) == 0 && len(testIncludePatterns) == 0 {
			return fmt.Errorf("at least one pattern is required (--exclude or --include)")
		}

		ctrl, err := helpers.NewInitializedController(cmdCmd)
		if err != nil {
			return err
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

func InitTestFilterCmd() error {
	TestFilterCmd.Flags().StringVarP(&testFilterName, "name", "n", "test", "Filter name for display purposes")
	TestFilterCmd.Flags().StringSliceVarP(&testExcludePatterns, "exclude", "e", []string{}, "Exclude patterns (can specify multiple)")
	TestFilterCmd.Flags().StringSliceVarP(&testIncludePatterns, "include", "i", []string{}, "Include patterns (can specify multiple)")
	return nil
}
