package cmd

import (
	"fmt"
	"os"

	"github.com/go-go-golems/prescribe/cmd/file"
	"github.com/go-go-golems/prescribe/cmd/filter"
	"github.com/go-go-golems/prescribe/cmd/session"
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
	// Filter commands
	rootCmd.AddCommand(filter.AddFilterCmd)
	rootCmd.AddCommand(filter.ListFiltersCmd)
	rootCmd.AddCommand(filter.RemoveFilterCmd)
	rootCmd.AddCommand(filter.ClearFiltersCmd)
	rootCmd.AddCommand(filter.TestFilterCmd)
	rootCmd.AddCommand(filter.ShowFilteredCmd)

	// Session commands
	rootCmd.AddCommand(session.InitCmd)
	rootCmd.AddCommand(session.SaveCmd)
	rootCmd.AddCommand(session.LoadCmd)
	rootCmd.AddCommand(session.ShowCmd)

	// File commands
	rootCmd.AddCommand(file.ToggleFileCmd)
	rootCmd.AddCommand(file.AddContextCmd)

	// Root-level commands (generate, tui)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(tuiCmd)
}
