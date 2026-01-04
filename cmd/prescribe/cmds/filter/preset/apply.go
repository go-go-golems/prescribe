package preset

import (
	"context"
	"fmt"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	glazed_layers "github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/helpers"
	"github.com/go-go-golems/prescribe/internal/domain"
	prescribe_layers "github.com/go-go-golems/prescribe/pkg/layers"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type ApplySettings struct {
	PresetID string `glazed.parameter:"preset-id"`
}

type ApplyCommand struct {
	*cmds.CommandDescription
}

var _ cmds.BareCommand = &ApplyCommand{}

func NewApplyCommand() (*ApplyCommand, error) {
	repoLayer, err := prescribe_layers.NewRepositoryLayer()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository layer")
	}
	repoLayerExisting, err := prescribe_layers.WrapAsExistingCobraFlagsLayer(repoLayer)
	if err != nil {
		return nil, errors.Wrap(err, "failed to wrap repository layer as existing flags layer")
	}

	defaultLayer, err := schema.NewSection(
		schema.DefaultSlug,
		"Default",
		schema.WithArguments(
			fields.New(
				"preset-id",
				fields.TypeString,
				fields.WithHelp("Preset ID (filename)"),
				fields.WithRequired(true),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	cmdDesc := cmds.NewCommandDescription(
		"apply",
		cmds.WithShort("Apply a filter preset to the current session"),
		cmds.WithLong("Load a filter preset by ID (filename) and add it to active filters, then save the session."),
		cmds.WithLayersList(repoLayerExisting, defaultLayer),
	)

	return &ApplyCommand{CommandDescription: cmdDesc}, nil
}

func (c *ApplyCommand) Run(ctx context.Context, parsedLayers *glazed_layers.ParsedLayers) error {
	_ = ctx

	settings := &ApplySettings{}
	if err := parsedLayers.InitializeStruct(schema.DefaultSlug, settings); err != nil {
		return errors.Wrap(err, "failed to initialize preset apply settings")
	}

	ctrl, err := helpers.NewInitializedControllerFromParsedLayers(parsedLayers)
	if err != nil {
		return err
	}

	helpers.LoadDefaultSessionIfExists(ctrl)

	p, err := ctrl.LoadFilterPresetByID(settings.PresetID)
	if err != nil {
		return errors.Wrap(err, "failed to load filter preset")
	}

	ctrl.AddFilter(domain.Filter{
		Name:        p.Name,
		Description: p.Description,
		Rules:       p.Rules,
	})

	savePath := ctrl.GetDefaultSessionPath()
	if err := ctrl.SaveSession(savePath); err != nil {
		return errors.Wrap(err, "failed to save session")
	}

	fmt.Printf("Applied filter preset %q and saved session\n", settings.PresetID)
	return nil
}

func NewApplyCobraCommand() (*cobra.Command, error) {
	glazedCmd, err := NewApplyCommand()
	if err != nil {
		return nil, err
	}
	return cli.BuildCobraCommand(
		glazedCmd,
		cli.WithParserConfig(cli.CobraParserConfig{
			MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
		}),
	)
}
