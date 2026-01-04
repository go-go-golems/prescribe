package add

import (
	"fmt"
	"strings"

	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/helpers"
	"github.com/go-go-golems/prescribe/internal/domain"
	"github.com/spf13/cobra"
)

func newFileDiffCmd() *cobra.Command {
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
