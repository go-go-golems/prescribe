package filter

import (
	"fmt"

	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/helpers"
	"github.com/go-go-golems/prescribe/internal/domain"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var FilterPresetApplyCmd *cobra.Command

func InitFilterPresetApplyCmd() error {
	FilterPresetApplyCmd = &cobra.Command{
		Use:   "apply PRESET_ID",
		Short: "Apply a filter preset to the current session",
		Long:  "Loads a filter preset by ID (filename) and adds it to the active filters, then saves the session.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctrl, err := helpers.NewInitializedController(cmd)
			if err != nil {
				return err
			}

			helpers.LoadDefaultSessionIfExists(ctrl)

			presetID := args[0]
			p, err := ctrl.LoadFilterPresetByID(presetID)
			if err != nil {
				return errors.Wrap(err, "failed to load filter preset")
			}

			ctrl.AddFilter(domain.Filter{
				Name:        p.Name,
				Description: p.Description,
				Rules:       p.Rules,
			})

			savePath := ctrl.GetDefaultSessionPath()
			if err := ctrl.SaveSession(savePath); err != nil {
				return errors.Wrap(err, "failed to save session")
			}

			fmt.Printf("Applied filter preset %q and saved session\n", presetID)
			return nil
		},
	}
	return nil
}
