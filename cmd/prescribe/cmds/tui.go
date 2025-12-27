package cmds

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	glazed_layers "github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/helpers"
	"github.com/go-go-golems/prescribe/internal/tui/app"
	prescribe_layers "github.com/go-go-golems/prescribe/pkg/layers"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var tuiCmd *cobra.Command

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

	cmdDesc := cmds.NewCommandDescription(
		"tui",
		cmds.WithShort("Launch interactive TUI"),
		cmds.WithLong("Launch the interactive Terminal User Interface for building PR descriptions."),
		cmds.WithLayersList(
			repoLayerExisting,
		),
	)

	return &TuiCommand{CommandDescription: cmdDesc}, nil
}

func (c *TuiCommand) Run(ctx context.Context, parsedLayers *glazed_layers.ParsedLayers) error {
	_ = ctx

	ctrl, err := helpers.NewInitializedControllerFromParsedLayers(parsedLayers)
	if err != nil {
		return err
	}

	helpers.LoadDefaultSessionIfExists(ctrl)

	p := tea.NewProgram(app.New(ctrl, app.DefaultDeps{}), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return errors.Wrap(err, "failed to run TUI")
	}

	return nil
}

func InitTuiCmd() error {
	glazedCmd, err := NewTuiCommand()
	if err != nil {
		return err
	}
	cobraCmd, err := cli.BuildCobraCommand(
		glazedCmd,
		cli.WithParserConfig(cli.CobraParserConfig{
			MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
		}),
	)
	if err != nil {
		return err
	}

	tuiCmd = cobraCmd
	return nil
}
