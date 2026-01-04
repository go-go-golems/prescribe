package git

import (
	"fmt"
	"strconv"

	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/helpers"
	"github.com/go-go-golems/prescribe/internal/domain"
	"github.com/spf13/cobra"
)

func effectiveGitHistoryConfig(data *domain.PRData) (domain.GitHistoryConfig, bool) {
	if data != nil && data.GitHistory != nil {
		return *data.GitHistory, true
	}
	return domain.DefaultGitHistoryConfig(), false
}

func newGitHistoryShowCmd() *cobra.Command {
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

func newGitHistoryEnableCmd(enable bool) *cobra.Command {
	verb := "enable"
	short := "Enable derived git history"
	if !enable {
		verb = "disable"
		short = "Disable derived git history"
	}

	return &cobra.Command{
		Use:   verb,
		Short: short,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctrl, err := helpers.NewInitializedController(cmd)
			if err != nil {
				return err
			}
			helpers.LoadDefaultSessionIfExists(ctrl)

			data := ctrl.GetData()
			cfg, _ := effectiveGitHistoryConfig(data)
			cfg.Enabled = enable
			data.GitHistory = &cfg

			savePath := ctrl.GetDefaultSessionPath()
			if err := ctrl.SaveSession(savePath); err != nil {
				return err
			}
			fmt.Printf("Git history %sd\n", verb)
			fmt.Printf("Session saved: %s\n", savePath)
			return nil
		},
	}
}

func newGitHistorySetCmd() *cobra.Command {
	var (
		enabledStr        string
		maxCommits        int
		includeMerges     bool
		firstParent       bool
		includeNumstat    bool
		includeMergesSet  bool
		firstParentSet    bool
		includeNumstatSet bool
		maxCommitsSet     bool
	)

	cmd := &cobra.Command{
		Use:   "set",
		Short: "Update git history config",
		Long:  "Update git history config fields in session.yaml (only provided flags are applied).",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctrl, err := helpers.NewInitializedController(cmd)
			if err != nil {
				return err
			}
			helpers.LoadDefaultSessionIfExists(ctrl)

			data := ctrl.GetData()
			cfg, _ := effectiveGitHistoryConfig(data)

			if cmd.Flags().Changed("enabled") {
				v, err := strconv.ParseBool(enabledStr)
				if err != nil {
					return fmt.Errorf("invalid --enabled value %q (expected true/false)", enabledStr)
				}
				cfg.Enabled = v
			}
			if maxCommitsSet {
				if maxCommits <= 0 {
					return fmt.Errorf("--max-commits must be > 0 (got %d)", maxCommits)
				}
				cfg.MaxCommits = maxCommits
			}
			if includeMergesSet {
				cfg.IncludeMerges = includeMerges
			}
			if firstParentSet {
				cfg.FirstParent = firstParent
			}
			if includeNumstatSet {
				cfg.IncludeNumstat = includeNumstat
			}

			data.GitHistory = &cfg

			savePath := ctrl.GetDefaultSessionPath()
			if err := ctrl.SaveSession(savePath); err != nil {
				return err
			}

			fmt.Printf("Git history config updated\n")
			fmt.Printf("Session saved: %s\n", savePath)
			return nil
		},
	}

	cmd.Flags().StringVar(&enabledStr, "enabled", "", "Set enabled (true/false)")
	cmd.Flags().IntVar(&maxCommits, "max-commits", 0, "Set max_commits (positive integer)")
	cmd.Flags().BoolVar(&includeMerges, "include-merges", false, "Set include_merges (true/false)")
	cmd.Flags().BoolVar(&firstParent, "first-parent", false, "Set first_parent (true/false)")
	cmd.Flags().BoolVar(&includeNumstat, "include-numstat", false, "Set include_numstat (true/false)")

	cmd.Flags().Lookup("include-merges").NoOptDefVal = "true"
	cmd.Flags().Lookup("first-parent").NoOptDefVal = "true"
	cmd.Flags().Lookup("include-numstat").NoOptDefVal = "true"

	cmd.Flags().Lookup("max-commits").NoOptDefVal = "30"

	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		includeMergesSet = cmd.Flags().Changed("include-merges")
		firstParentSet = cmd.Flags().Changed("first-parent")
		includeNumstatSet = cmd.Flags().Changed("include-numstat")
		maxCommitsSet = cmd.Flags().Changed("max-commits")
		return nil
	}

	return cmd
}
