package cmds

import (
	"fmt"
	"os"

	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/context"
	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/file"
	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/filter"
	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/session"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "prescribe",
	Short: "A TUI for building GitHub PR descriptions",
	Long: `Prescribe is a CLI/TUI application for generating pull request descriptions using LLMs.
	
It allows you to:
- View and filter PR diffs
- Toggle file inclusion and replace diffs with full files
- Apply filters with glob patterns
- Customize prompts with presets
- Generate AI-powered PR descriptions`,
	Version: "0.1.0",
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringP("repo", "r", ".", "Path to git repository")
	rootCmd.PersistentFlags().StringP("target", "t", "", "Target branch (default: main or master)")

	// Register subdirectory commands
	registerCommands()
}

func registerCommands() {
	// Command groups
	rootCmd.AddCommand(filter.FilterCmd)
	rootCmd.AddCommand(session.SessionCmd)
	rootCmd.AddCommand(file.FileCmd)
	rootCmd.AddCommand(context.ContextCmd)

	// Root-level commands (generate, tui)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(tuiCmd)
}
