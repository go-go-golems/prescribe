package session

import (
	"sync"

	"github.com/spf13/cobra"
)

// SessionCmd groups all session-related subcommands.
var SessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Manage sessions",
	Long:  "Initialize, inspect, save, and load PR builder sessions.",
}

var initOnce sync.Once
var initErr error

func Init() error {
	initOnce.Do(func() {
		if err := InitInitCmd(); err != nil {
			initErr = err
			return
		}
		if err := InitSaveCmd(); err != nil {
			initErr = err
			return
		}
		if err := InitLoadCmd(); err != nil {
			initErr = err
			return
		}
		if err := InitShowCmd(); err != nil {
			initErr = err
			return
		}
		if err := InitTokenCountCmd(); err != nil {
			initErr = err
			return
		}

		SessionCmd.AddCommand(
			InitCmd,
			SaveCmd,
			LoadCmd,
			ShowCmd,
			TokenCountCmd,
		)
	})
	return initErr
}
