package helpers

import (
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/prescribe/internal/controller"
	prescribe_layers "github.com/go-go-golems/prescribe/pkg/layers"
	"github.com/pkg/errors"
)

// NewInitializedControllerFromParsedLayers creates a controller from Glazed parsed layers
// and runs Initialize().
//
// This is intended for Glazed-based commands that have access to `*layers.ParsedLayers`
// rather than Cobra flags.
func NewInitializedControllerFromParsedLayers(parsedLayers *layers.ParsedLayers) (*controller.Controller, error) {
	repoSettings, err := prescribe_layers.GetRepositorySettings(parsedLayers)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get repository settings")
	}

	ctrl, err := controller.NewController(repoSettings.RepoPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create controller")
	}

	if err := ctrl.Initialize(repoSettings.TargetBranch); err != nil {
		return nil, errors.Wrap(err, "failed to initialize")
	}

	return ctrl, nil
}


