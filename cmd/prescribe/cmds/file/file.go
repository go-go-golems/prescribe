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
var initErr error

func Init() error {
	initOnce.Do(func() {
		if err := InitToggleFileCmd(); err != nil {
			initErr = err
			return
		}
		FileCmd.AddCommand(ToggleFileCmd)
	})
	return initErr
}
