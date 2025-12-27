package filter

import (
	"sync"

	"github.com/spf13/cobra"
)

// PresetCmd groups all filter preset-related subcommands.
var PresetCmd = &cobra.Command{
	Use:   "preset",
	Short: "Manage filter presets",
	Long:  "List, save, and apply named filter presets.",
}

var presetInitOnce sync.Once
var presetInitErr error

func InitPresetCmd() error {
	presetInitOnce.Do(func() {
		if err := InitFilterPresetListCmd(); err != nil {
			presetInitErr = err
			return
		}
		if err := InitFilterPresetSaveCmd(); err != nil {
			presetInitErr = err
			return
		}
		if err := InitFilterPresetApplyCmd(); err != nil {
			presetInitErr = err
			return
		}

		PresetCmd.AddCommand(
			FilterPresetListCmd,
			FilterPresetSaveCmd,
			FilterPresetApplyCmd,
		)
	})
	return presetInitErr
}
