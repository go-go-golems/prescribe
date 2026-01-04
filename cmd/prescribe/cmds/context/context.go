package context

import (
	"sync"

	"github.com/spf13/cobra"
)

// ContextCmd groups all additional-context related subcommands.
var ContextCmd = &cobra.Command{
	Use:   "context",
	Short: "Manage additional context",
	Long:  "Add additional context (files and notes) to the current session.",
}

var initOnce sync.Once
var initErr error

func Init() error {
	initOnce.Do(func() {
		if err := InitAddCmd(); err != nil {
			initErr = err
			return
		}

		if err := InitGitCmd(); err != nil {
			initErr = err
			return
		}

		ContextCmd.AddCommand(AddCmd, GitCmd)
	})
	return initErr
}
