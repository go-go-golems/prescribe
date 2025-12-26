package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/go-go-golems/prescribe/internal/controller"
)

var (
	outputFile   string
	promptText   string
	presetID     string
	loadSession  string
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate PR description",
	Long:  `Generate a PR description using AI based on the current session.`,
	RunE: func(cmdCmd *cobra.Command, args []string) error {
		// Get flags from parent command
		repoPath, _ := cmdCmd.Flags().GetString("repo")
		targetBranch, _ := cmdCmd.Flags().GetString("target")
		if repoPath == "" {
			repoPath = "."
		}
		
		// Create controller
		ctrl, err := controller.NewController(repoPath)
		if err != nil {
			return fmt.Errorf("failed to create controller: %w", err)
		}
		
		// Initialize
		if err := ctrl.Initialize(targetBranch); err != nil {
			return fmt.Errorf("failed to initialize: %w", err)
		}
		
		// Load session if specified
		if loadSession != "" {
			if err := ctrl.LoadSession(loadSession); err != nil {
				return fmt.Errorf("failed to load session: %w", err)
			}
			fmt.Fprintf(os.Stderr, "Loaded session from: %s\n", loadSession)
		}
		
		// Override prompt if specified
		if promptText != "" {
			ctrl.SetPrompt(promptText, nil)
		} else if presetID != "" {
			if err := ctrl.LoadPromptPreset(presetID); err != nil {
				return fmt.Errorf("failed to load preset: %w", err)
			}
		}
		
		// Generate description
		fmt.Fprintf(os.Stderr, "Generating PR description...\n")
		description, err := ctrl.GenerateDescription()
		if err != nil {
			return fmt.Errorf("failed to generate description: %w", err)
		}
		
		// Output description
		if outputFile != "" {
			if err := os.WriteFile(outputFile, []byte(description), 0644); err != nil {
				return fmt.Errorf("failed to write output file: %w", err)
			}
			fmt.Fprintf(os.Stderr, "Description written to %s\n", outputFile)
		} else {
			fmt.Println(description)
		}
		
		return nil
	},
}

func init() {
	generateCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (default: stdout)")
	generateCmd.Flags().StringVarP(&promptText, "prompt", "p", "", "Custom prompt text")
	generateCmd.Flags().StringVar(&presetID, "preset", "", "Prompt preset ID")
	generateCmd.Flags().StringVarP(&loadSession, "session", "s", "", "Load session file before generating")
}
