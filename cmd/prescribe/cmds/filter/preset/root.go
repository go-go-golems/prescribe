package preset

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// NewPresetCmd groups all filter preset-related subcommands.
func NewPresetCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "preset",
		Short: "Manage filter presets",
		Long:  "List, save, and apply named filter presets.",
	}

	listCmd, err := NewListCobraCommand()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build filter preset list command")
	}
	saveCmd, err := NewSaveCobraCommand()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build filter preset save command")
	}
	applyCmd, err := NewApplyCobraCommand()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build filter preset apply command")
	}

	cmd.AddCommand(listCmd, saveCmd, applyCmd)
	return cmd, nil
}
