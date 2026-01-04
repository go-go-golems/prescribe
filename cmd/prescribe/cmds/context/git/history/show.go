package history

import (
	"fmt"

	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/helpers"
	"github.com/spf13/cobra"
)

func newShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show effective git history config",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctrl, err := helpers.NewInitializedController(cmd)
			if err != nil {
				return err
			}
			helpers.LoadDefaultSessionIfExists(ctrl)

			data := ctrl.GetData()
			cfg, explicit := effectiveGitHistoryConfig(data)
			rangeSpec := fmt.Sprintf("%s..%s", data.TargetBranch, data.SourceBranch)

			src := "defaults (missing in session.yaml)"
			if explicit {
				src = "session.yaml"
			}

			fmt.Printf("Git history config (%s)\n", src)
			fmt.Printf("  enabled: %v\n", cfg.Enabled)
			fmt.Printf("  max_commits: %d\n", cfg.MaxCommits)
			fmt.Printf("  include_merges: %v\n", cfg.IncludeMerges)
			fmt.Printf("  first_parent: %v\n", cfg.FirstParent)
			fmt.Printf("  include_numstat: %v\n", cfg.IncludeNumstat)
			fmt.Printf("  range: %s\n", rangeSpec)
			return nil
		},
	}
}
