package session

import (
	"context"
	"fmt"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	glazed_layers "github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/helpers"
	prescribe_layers "github.com/go-go-golems/prescribe/pkg/layers"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var SaveCmd *cobra.Command

type SessionSaveSettings struct {
	Path string `glazed.parameter:"path"`
}

type SessionSaveCommand struct {
	*cmds.CommandDescription
}

var _ cmds.BareCommand = &SessionSaveCommand{}

func NewSessionSaveCommand() (*SessionSaveCommand, error) {
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
				"path",
				fields.TypeString,
				fields.WithHelp("Path to YAML session file (default: app default session path)"),
				fields.WithRequired(false),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	cmdDesc := cmds.NewCommandDescription(
		"save",
		cmds.WithShort("Save current session to YAML file"),
		cmds.WithLong("Save the current PR builder session to a YAML file."),
		cmds.WithLayersList(
			repoLayerExisting,
			defaultLayer,
		),
	)

	return &SessionSaveCommand{CommandDescription: cmdDesc}, nil
}

func (c *SessionSaveCommand) Run(ctx context.Context, parsedLayers *glazed_layers.ParsedLayers) error {
	_ = ctx

	settings := &SessionSaveSettings{}
	if err := parsedLayers.InitializeStruct(schema.DefaultSlug, settings); err != nil {
		return errors.Wrap(err, "failed to initialize session save settings")
	}

	ctrl, err := helpers.NewInitializedControllerFromParsedLayers(parsedLayers)
	if err != nil {
		return err
	}

	// Prefer loading an existing session so this command reflects current state.
	helpers.LoadDefaultSessionIfExists(ctrl)

	savePath := ctrl.GetDefaultSessionPath()
	if settings.Path != "" {
		savePath = settings.Path
	}

	if err := ctrl.SaveSession(savePath); err != nil {
		return errors.Wrap(err, "failed to save session")
	}

	fmt.Printf("Session saved to: %s\n", savePath)
	return nil
}

func InitSaveCmd() error {
	glazedCmd, err := NewSessionSaveCommand()
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

	SaveCmd = cobraCmd
	return nil
}
