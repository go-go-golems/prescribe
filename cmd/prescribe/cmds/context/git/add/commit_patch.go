package add

import (
	"fmt"
	"strings"

	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/helpers"
	"github.com/go-go-golems/prescribe/internal/domain"
	"github.com/spf13/cobra"
)

func newCommitPatchCmd() *cobra.Command {
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
