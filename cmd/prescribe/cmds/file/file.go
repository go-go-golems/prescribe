package file

import (
	"sync"

	"github.com/spf13/cobra"
)

// FileCmd groups all file-related subcommands.
var FileCmd = &cobra.Command{
	Use:   "file",
	Short: "File operations",
	Long:  "Operations that act on individual changed files in the current session.",
}

var initOnce sync.Once

func Init() error {
	initOnce.Do(func() {
		FileCmd.AddCommand(ToggleFileCmd)
	})
	return nil
}
