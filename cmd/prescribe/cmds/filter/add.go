package filter

import (
	"fmt"

	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/helpers"
	"github.com/go-go-golems/prescribe/internal/domain"
	"github.com/spf13/cobra"
)

var (
	filterName        string
	filterDescription string
	excludePatterns   []string
	includePatterns   []string
)

var AddFilterCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a filter to the session",
	Long:  `Add a file filter to the current session.`,
	RunE: func(cmdCmd *cobra.Command, args []string) error {
		if filterName == "" {
			return fmt.Errorf("filter name is required (--name)")
		}

		if len(excludePatterns) == 0 && len(includePatterns) == 0 {
			return fmt.Errorf("at least one pattern is required (--exclude or --include)")
		}

		ctrl, err := helpers.NewInitializedController(cmdCmd)
		if err != nil {
			return err
		}

		// Load existing session if present so we don't clobber it on save.
		helpers.LoadDefaultSessionIfExists(ctrl)

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
