package git

import (
	"fmt"
	"strconv"
	"strings"

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

func newGitContextListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List configured git_context items",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctrl, err := helpers.NewInitializedController(cmd)
			if err != nil {
				return err
			}
			helpers.LoadDefaultSessionIfExists(ctrl)

			items := ctrl.GetData().GitContext
			if len(items) == 0 {
				fmt.Println("No git_context items configured")
				return nil
			}

			for i, it := range items {
				switch it.Kind {
				case domain.GitContextItemKindCommit, domain.GitContextItemKindCommitPatch:
					fmt.Printf("[%d] %s ref=%s", i, it.Kind, it.Ref)
					if len(it.Paths) > 0 {
						fmt.Printf(" paths=%s", strings.Join(it.Paths, ","))
					}
					fmt.Printf("\n")
				case domain.GitContextItemKindFileAtRef:
					fmt.Printf("[%d] %s ref=%s path=%s\n", i, it.Kind, it.Ref, it.Path)
				case domain.GitContextItemKindFileDiff:
					fmt.Printf("[%d] %s from=%s to=%s path=%s\n", i, it.Kind, it.From, it.To, it.Path)
				default:
					fmt.Printf("[%d] %s\n", i, it.Kind)
				}
			}
			return nil
		},
	}
}

func newGitContextRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <index>",
		Short: "Remove a git_context item by index",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			idx, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid index %q", args[0])
			}

			ctrl, err := helpers.NewInitializedController(cmd)
			if err != nil {
				return err
			}
			helpers.LoadDefaultSessionIfExists(ctrl)

			data := ctrl.GetData()
			if idx < 0 || idx >= len(data.GitContext) {
				return fmt.Errorf("index out of range: %d", idx)
			}

			data.GitContext = append(data.GitContext[:idx], data.GitContext[idx+1:]...)
			savePath := ctrl.GetDefaultSessionPath()
			if err := ctrl.SaveSession(savePath); err != nil {
				return err
			}
			fmt.Printf("Removed git_context item %d\n", idx)
			fmt.Printf("Session saved: %s\n", savePath)
			return nil
		},
	}
}

func newGitContextClearCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clear",
		Short: "Clear all git_context items",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctrl, err := helpers.NewInitializedController(cmd)
			if err != nil {
				return err
			}
			helpers.LoadDefaultSessionIfExists(ctrl)

			data := ctrl.GetData()
			data.GitContext = nil

			savePath := ctrl.GetDefaultSessionPath()
			if err := ctrl.SaveSession(savePath); err != nil {
				return err
			}
			fmt.Printf("Cleared git_context\n")
			fmt.Printf("Session saved: %s\n", savePath)
			return nil
		},
	}
}

func newGitContextAddCmd() *cobra.Command {
	addCmd := &cobra.Command{
		Use:   "add",
		Short: "Add a git_context item",
	}

	addCmd.AddCommand(
		newGitContextAddCommitCmd(),
		newGitContextAddCommitPatchCmd(),
		newGitContextAddFileAtCmd(),
		newGitContextAddFileDiffCmd(),
	)
	return addCmd
}

func newGitContextAddCommitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "commit <ref>",
		Short: "Add a commit metadata item",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref := strings.TrimSpace(args[0])
			if ref == "" {
				return fmt.Errorf("ref is required")
			}

			ctrl, err := helpers.NewInitializedController(cmd)
			if err != nil {
				return err
			}
			helpers.LoadDefaultSessionIfExists(ctrl)

			data := ctrl.GetData()
			data.GitContext = append(data.GitContext, domain.GitContextItem{
				Kind: domain.GitContextItemKindCommit,
				Ref:  ref,
			})

			savePath := ctrl.GetDefaultSessionPath()
			if err := ctrl.SaveSession(savePath); err != nil {
				return err
			}
			fmt.Printf("Added git_context commit ref=%s\n", ref)
			fmt.Printf("Session saved: %s\n", savePath)
			return nil
		},
	}
}

func newGitContextAddCommitPatchCmd() *cobra.Command {
	var paths []string
	cmd := &cobra.Command{
		Use:   "commit-patch <ref>",
		Short: "Add a commit patch item (diff text)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref := strings.TrimSpace(args[0])
			if ref == "" {
				return fmt.Errorf("ref is required")
			}

			ctrl, err := helpers.NewInitializedController(cmd)
			if err != nil {
				return err
			}
			helpers.LoadDefaultSessionIfExists(ctrl)

			data := ctrl.GetData()
			data.GitContext = append(data.GitContext, domain.GitContextItem{
				Kind:  domain.GitContextItemKindCommitPatch,
				Ref:   ref,
				Paths: append([]string{}, paths...),
			})

			savePath := ctrl.GetDefaultSessionPath()
			if err := ctrl.SaveSession(savePath); err != nil {
				return err
			}
			fmt.Printf("Added git_context commit_patch ref=%s\n", ref)
			fmt.Printf("Session saved: %s\n", savePath)
			return nil
		},
	}
	cmd.Flags().StringArrayVar(&paths, "path", nil, "Optional path filter (can be repeated)")
	return cmd
}

func newGitContextAddFileAtCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "file-at <ref> <path>",
		Short: "Add file content at a ref",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref := strings.TrimSpace(args[0])
			path := strings.TrimSpace(args[1])
			if ref == "" || path == "" {
				return fmt.Errorf("ref and path are required")
			}

			ctrl, err := helpers.NewInitializedController(cmd)
			if err != nil {
				return err
			}
			helpers.LoadDefaultSessionIfExists(ctrl)

			data := ctrl.GetData()
			data.GitContext = append(data.GitContext, domain.GitContextItem{
				Kind: domain.GitContextItemKindFileAtRef,
				Ref:  ref,
				Path: path,
			})

			savePath := ctrl.GetDefaultSessionPath()
			if err := ctrl.SaveSession(savePath); err != nil {
				return err
			}
			fmt.Printf("Added git_context file_at_ref ref=%s path=%s\n", ref, path)
			fmt.Printf("Session saved: %s\n", savePath)
			return nil
		},
	}
}

func newGitContextAddFileDiffCmd() *cobra.Command {
	var (
		from string
		to   string
		path string
	)
	cmd := &cobra.Command{
		Use:   "file-diff",
		Short: "Add a single-file diff between two refs",
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(from) == "" || strings.TrimSpace(to) == "" || strings.TrimSpace(path) == "" {
				return fmt.Errorf("--from, --to, and --path are required")
			}

			ctrl, err := helpers.NewInitializedController(cmd)
			if err != nil {
				return err
			}
			helpers.LoadDefaultSessionIfExists(ctrl)

			data := ctrl.GetData()
			data.GitContext = append(data.GitContext, domain.GitContextItem{
				Kind: domain.GitContextItemKindFileDiff,
				From: strings.TrimSpace(from),
				To:   strings.TrimSpace(to),
				Path: strings.TrimSpace(path),
			})

			savePath := ctrl.GetDefaultSessionPath()
			if err := ctrl.SaveSession(savePath); err != nil {
				return err
			}
			fmt.Printf("Added git_context file_diff from=%s to=%s path=%s\n", from, to, path)
			fmt.Printf("Session saved: %s\n", savePath)
			return nil
		},
	}
	cmd.Flags().StringVar(&from, "from", "", "From ref (required)")
	cmd.Flags().StringVar(&to, "to", "", "To ref (required)")
	cmd.Flags().StringVar(&path, "path", "", "File path (required)")
	_ = cmd.MarkFlagRequired("from")
	_ = cmd.MarkFlagRequired("to")
	_ = cmd.MarkFlagRequired("path")
	return cmd
}
