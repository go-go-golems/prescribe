package helpers

import (
	"fmt"

	"github.com/go-go-golems/prescribe/internal/controller"
	"github.com/spf13/cobra"
)

// RepoParams represents the common repo/target parameters shared across most commands.
type RepoParams struct {
	RepoPath     string
	TargetBranch string
}

func GetRepoParams(cmd *cobra.Command) RepoParams {
	repoPath, _ := cmd.Flags().GetString("repo")
	targetBranch, _ := cmd.Flags().GetString("target")
	if repoPath == "" {
		repoPath = "."
	}

	return RepoParams{
		RepoPath:     repoPath,
		TargetBranch: targetBranch,
	}
}

// NewInitializedController creates a controller from Cobra flags and runs Initialize().
func NewInitializedController(cmd *cobra.Command) (*controller.Controller, error) {
	params := GetRepoParams(cmd)

	ctrl, err := controller.NewController(params.RepoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create controller: %w", err)
	}

	if err := ctrl.Initialize(params.TargetBranch); err != nil {
		return nil, fmt.Errorf("failed to initialize: %w", err)
	}

	return ctrl, nil
}
