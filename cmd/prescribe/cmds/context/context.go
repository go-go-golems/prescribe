package context

import "github.com/spf13/cobra"

// ContextCmd groups all additional-context related subcommands.
var ContextCmd = &cobra.Command{
	Use:   "context",
	Short: "Manage additional context",
	Long:  "Add additional context (files and notes) to the current session.",
}

func init() {
	ContextCmd.AddCommand(
		AddCmd,
	)
}
