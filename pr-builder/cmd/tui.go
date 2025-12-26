package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/user/pr-builder/internal/controller"
	"github.com/user/pr-builder/internal/tui"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch interactive TUI",
	Long:  `Launch the interactive Terminal User Interface for building PR descriptions.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create controller
		ctrl, err := controller.NewController(repoPath)
		if err != nil {
			return fmt.Errorf("failed to create controller: %w", err)
		}
		
		// Initialize
		if err := ctrl.Initialize(targetBranch); err != nil {
			return fmt.Errorf("failed to initialize: %w", err)
		}
		
		// Try to load existing session
		sessionPath := ctrl.GetDefaultSessionPath()
		_ = ctrl.LoadSession(sessionPath) // Ignore error if no session exists
		
		// Create and run TUI
		p := tea.NewProgram(tui.NewEnhancedModel(ctrl), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("failed to run TUI: %w", err)
		}
		
		return nil
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
