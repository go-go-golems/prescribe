package cmds

import (
	"context"
	stderrors "errors"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	geppettolayers "github.com/go-go-golems/geppetto/pkg/layers"
	gepsettings "github.com/go-go-golems/geppetto/pkg/steps/ai/settings"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	glazed_layers "github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/helpers"
	"github.com/go-go-golems/prescribe/internal/tui/app"
	prescribe_layers "github.com/go-go-golems/prescribe/pkg/layers"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type TuiCommand struct {
	*cmds.CommandDescription
}

var _ cmds.BareCommand = &TuiCommand{}

func NewTuiCommand() (*TuiCommand, error) {
	repoLayer, err := prescribe_layers.NewRepositoryLayer()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository layer")
	}
	repoLayerExisting, err := prescribe_layers.WrapAsExistingCobraFlagsLayer(repoLayer)
	if err != nil {
		return nil, errors.Wrap(err, "failed to wrap repository layer as existing flags layer")
	}

	geppettoLayers, err := geppettolayers.CreateGeppettoLayers()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create geppetto parameter layers")
	}

	layersList := []glazed_layers.ParameterLayer{
		repoLayerExisting,
	}
	layersList = append(layersList, geppettoLayers...)

	cmdDesc := cmds.NewCommandDescription(
		"tui",
		cmds.WithShort("Launch interactive TUI"),
		cmds.WithLong("Launch the interactive Terminal User Interface for building PR descriptions."),
		cmds.WithLayersList(
			layersList...,
		),
	)

	return &TuiCommand{CommandDescription: cmdDesc}, nil
}

func (c *TuiCommand) Run(ctx context.Context, parsedLayers *glazed_layers.ParsedLayers) error {
	ctrl, err := helpers.NewInitializedControllerFromParsedLayers(parsedLayers)
	if err != nil {
		return err
	}

	stepSettings, err := gepsettings.NewStepSettingsFromParsedLayers(parsedLayers)
	if err != nil {
		return errors.Wrap(err, "failed to build AI step settings from parsed layers")
	}
	ctrl.SetStepSettings(stepSettings)

	// The TUI requires an initialized, persisted session.
	// This ensures users explicitly capture their working set (filters + included files) before interacting.
	sessionPath := ctrl.GetDefaultSessionPath()
	if err := ctrl.LoadSession(sessionPath); err != nil {
		if stderrors.Is(err, os.ErrNotExist) {
			return errors.Errorf("no session found at %s; run 'prescribe session init --save' first", sessionPath)
		}
		return errors.Wrap(err, "failed to load session")
	}

	p := tea.NewProgram(app.New(ctrl, app.DefaultDeps{}), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return errors.Wrap(err, "failed to run TUI")
	}

	return nil
}

func NewTuiCobraCommand() (*cobra.Command, error) {
	glazedCmd, err := NewTuiCommand()
	if err != nil {
		return nil, err
	}
	cobraCmd, err := cli.BuildCobraCommand(
		glazedCmd,
		cli.WithParserConfig(cli.CobraParserConfig{
			MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
		}),
	)
	if err != nil {
		return nil, err
	}

	return cobraCmd, nil
}
