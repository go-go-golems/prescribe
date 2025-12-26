package session

import "github.com/spf13/cobra"

// SessionCmd groups all session-related subcommands.
var SessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Manage sessions",
	Long:  "Initialize, inspect, save, and load PR builder sessions.",
}

func init() {
	SessionCmd.AddCommand(
		InitCmd,
		SaveCmd,
		LoadCmd,
		ShowCmd,
	)
}
