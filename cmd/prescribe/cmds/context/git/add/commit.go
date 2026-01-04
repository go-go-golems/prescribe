package add

import (
	"fmt"
	"strings"

	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/helpers"
	"github.com/go-go-golems/prescribe/internal/domain"
	"github.com/spf13/cobra"
)

func newCommitCmd() *cobra.Command {
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
