package git

import (
	"fmt"
	"strings"

	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/helpers"
	"github.com/go-go-golems/prescribe/internal/domain"
	"github.com/spf13/cobra"
)

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
