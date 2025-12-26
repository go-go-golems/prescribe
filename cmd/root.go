package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	repoPath     string
	targetBranch string
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "pr-builder",
	Short: "A TUI for building GitHub PR descriptions",
	Long: `PR Builder is a CLI/TUI application for generating pull request descriptions using LLMs.
	
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
	rootCmd.PersistentFlags().StringVarP(&repoPath, "repo", "r", ".", "Path to git repository")
	rootCmd.PersistentFlags().StringVarP(&targetBranch, "target", "t", "", "Target branch (default: main or master)")
}
