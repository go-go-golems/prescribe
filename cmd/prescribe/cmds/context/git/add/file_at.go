package add

import (
	"fmt"
	"strings"

	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/helpers"
	"github.com/go-go-golems/prescribe/internal/domain"
	"github.com/spf13/cobra"
)

func newFileAtCmd() *cobra.Command {
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
