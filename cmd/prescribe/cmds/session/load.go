package session

import (
	"fmt"

	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/helpers"
	"github.com/spf13/cobra"
)

var LoadCmd = &cobra.Command{
	Use:   "load [path]",
	Short: "Load session from YAML file",
	Long:  `Load a PR builder session from a YAML file.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmdCmd *cobra.Command, args []string) error {
		ctrl, err := helpers.NewInitializedController(cmdCmd)
		if err != nil {
			return err
		}

		// Determine load path
		loadPath := ctrl.GetDefaultSessionPath()
		if len(args) > 0 {
			loadPath = args[0]
		}

		// Load session
		if err := ctrl.LoadSession(loadPath); err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}

		data := ctrl.GetData()

		fmt.Printf("Session loaded from: %s\n", loadPath)
		fmt.Printf("  Source: %s\n", data.SourceBranch)
		fmt.Printf("  Target: %s\n", data.TargetBranch)
		fmt.Printf("  Files: %d (%d included)\n", len(data.ChangedFiles), len(data.GetVisibleFiles()))
		fmt.Printf("  Filters: %d active\n", len(data.ActiveFilters))
		fmt.Printf("  Context: %d items\n", len(data.AdditionalContext))

		return nil
	},
}
