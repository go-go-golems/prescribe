package cmds

import (
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/context"
	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/file"
	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/filter"
	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/session"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	// rootCmd represents the base command
	rootCmd := &cobra.Command{
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
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Ensure logging is initialized before any subcommand runs.
			return logging.InitLoggerFromCobra(cmd)
		},
	}
	return rootCmd
}

func InitRootCmd(rootCmd *cobra.Command) error {
	if rootCmd == nil {
		return errors.New("rootCmd is nil")
	}

	// Global flags
	rootCmd.PersistentFlags().StringP("repo", "r", ".", "Path to git repository")
	rootCmd.PersistentFlags().StringP("target", "t", "", "Target branch (default: main or master)")

	// Explicit initialization of subcommand trees (no init() ordering reliance).
	if err := filter.Init(); err != nil {
		return errors.Wrap(err, "failed to init filter commands")
	}
	if err := session.Init(); err != nil {
		return errors.Wrap(err, "failed to init session commands")
	}
	if err := file.Init(); err != nil {
		return errors.Wrap(err, "failed to init file commands")
	}
	if err := context.Init(); err != nil {
		return errors.Wrap(err, "failed to init context commands")
	}
	if err := InitGenerateCmd(); err != nil {
		return errors.Wrap(err, "failed to init generate command")
	}
	if err := InitTuiCmd(); err != nil {
		return errors.Wrap(err, "failed to init tui command")
	}

	// Command groups
	rootCmd.AddCommand(filter.FilterCmd)
	rootCmd.AddCommand(session.SessionCmd)
	rootCmd.AddCommand(file.FileCmd)
	rootCmd.AddCommand(context.ContextCmd)

	// Root-level commands (generate, tui)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(tuiCmd)

	return nil
}

// Execute executes the provided root command.
func Execute(rootCmd *cobra.Command) {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
