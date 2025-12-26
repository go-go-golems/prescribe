package session

import (
	"fmt"

	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/helpers"
	"github.com/go-go-golems/prescribe/internal/session"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	showYAML bool
)

var ShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current session state",
	Long:  `Display the current PR builder session state.`,
	RunE: func(cmdCmd *cobra.Command, args []string) error {
		ctrl, err := helpers.NewInitializedController(cmdCmd)
		if err != nil {
			return err
		}

		helpers.LoadDefaultSessionIfExists(ctrl)

		data := ctrl.GetData()

		if showYAML {
			// Show as YAML
			sess := session.NewSession(data)
			yamlData, err := yaml.Marshal(sess)
			if err != nil {
				return fmt.Errorf("failed to marshal session: %w", err)
			}
			fmt.Print(string(yamlData))
		} else {
			// Show human-readable format
			fmt.Printf("PR Builder Session\n")
			fmt.Printf("==================\n\n")

			fmt.Printf("Branches:\n")
			fmt.Printf("  Source: %s\n", data.SourceBranch)
			fmt.Printf("  Target: %s\n", data.TargetBranch)
			fmt.Printf("\n")

			fmt.Printf("Files: %d total\n", len(data.ChangedFiles))
			visibleFiles := data.GetVisibleFiles()
			includedCount := 0
			for _, f := range visibleFiles {
				if f.Included {
					includedCount++
				}
			}
			fmt.Printf("  Visible: %d\n", len(visibleFiles))
			fmt.Printf("  Included: %d\n", includedCount)
			fmt.Printf("  Filtered: %d\n", len(data.GetFilteredFiles()))
			fmt.Printf("\n")

			if len(data.ActiveFilters) > 0 {
				fmt.Printf("Active Filters:\n")
				for _, filter := range data.ActiveFilters {
					fmt.Printf("  - %s: %s\n", filter.Name, filter.Description)
					for _, rule := range filter.Rules {
						fmt.Printf("      %s: %s\n", rule.Type, rule.Pattern)
					}
				}
				fmt.Printf("\n")
			}

			if len(data.AdditionalContext) > 0 {
				fmt.Printf("Additional Context:\n")
				for _, ctx := range data.AdditionalContext {
					if ctx.Type == "file" {
						fmt.Printf("  - File: %s\n", ctx.Path)
					} else {
						preview := ctx.Content
						if len(preview) > 60 {
							preview = preview[:60] + "..."
						}
						fmt.Printf("  - Note: %s\n", preview)
					}
				}
				fmt.Printf("\n")
			}

			fmt.Printf("Prompt:\n")
			if data.CurrentPreset != nil {
				fmt.Printf("  Preset: %s (%s)\n", data.CurrentPreset.Name, data.CurrentPreset.ID)
			} else {
				preview := data.CurrentPrompt
				if len(preview) > 100 {
					preview = preview[:100] + "..."
				}
				fmt.Printf("  Template: %s\n", preview)
			}
			fmt.Printf("\n")

			fmt.Printf("Token Count: %d\n", data.GetTotalTokens())
		}

		return nil
	},
}

func init() {
	ShowCmd.Flags().BoolVarP(&showYAML, "yaml", "y", false, "Show session as YAML")
}
