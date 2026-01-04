package filter

import (
	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/filter/preset"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// NewFilterCmd groups all filter-related subcommands.
func NewFilterCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "filter",
		Short: "Manage file filters",
		Long:  "Create, test, list, and manage file filters in the current session.",
	}

	addCmd, err := NewAddCobraCommand()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build filter add command")
	}
	listCmd, err := NewListCobraCommand()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build filter list command")
	}
	removeCmd, err := NewRemoveCobraCommand()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build filter remove command")
	}
	clearCmd, err := NewClearCobraCommand()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build filter clear command")
	}
	testCmd, err := NewTestCobraCommand()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build filter test command")
	}
	showCmd, err := NewShowCobraCommand()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build filter show command")
	}
	presetCmd, err := preset.NewPresetCmd()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build filter preset command")
	}

	cmd.AddCommand(
		addCmd,
		presetCmd,
		listCmd,
		removeCmd,
		clearCmd,
		testCmd,
		showCmd,
	)
	return cmd, nil
}
