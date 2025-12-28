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

var LoadCmd *cobra.Command

type SessionLoadSettings struct {
	Path string `glazed.parameter:"path"`
}

type SessionLoadCommand struct {
	*cmds.CommandDescription
}

var _ cmds.BareCommand = &SessionLoadCommand{}

func NewSessionLoadCommand() (*SessionLoadCommand, error) {
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
		"load",
		cmds.WithShort("Load session from YAML file"),
		cmds.WithLong("Load a PR builder session from a YAML file."),
		cmds.WithLayersList(
			repoLayerExisting,
			defaultLayer,
		),
	)

	return &SessionLoadCommand{CommandDescription: cmdDesc}, nil
}

func (c *SessionLoadCommand) Run(ctx context.Context, parsedLayers *glazed_layers.ParsedLayers) error {
	_ = ctx

	settings := &SessionLoadSettings{}
	if err := parsedLayers.InitializeStruct(schema.DefaultSlug, settings); err != nil {
		return errors.Wrap(err, "failed to initialize session load settings")
	}

	ctrl, err := helpers.NewInitializedControllerFromParsedLayers(parsedLayers)
	if err != nil {
		return err
	}

	loadPath := ctrl.GetDefaultSessionPath()
	if settings.Path != "" {
		loadPath = settings.Path
	}

	if err := ctrl.LoadSession(loadPath); err != nil {
		return errors.Wrap(err, "failed to load session")
	}

	data := ctrl.GetData()
	fmt.Printf("Session loaded from: %s\n", loadPath)
	fmt.Printf("  Source: %s\n", data.SourceBranch)
	fmt.Printf("  Target: %s\n", data.TargetBranch)
	fmt.Printf("  Files: %d (%d included)\n", len(data.ChangedFiles), len(data.GetVisibleFiles()))
	fmt.Printf("  Filters: %d active\n", len(data.ActiveFilters))
	fmt.Printf("  Context: %d items\n", len(data.AdditionalContext))

	return nil
}

func InitLoadCmd() error {
	glazedCmd, err := NewSessionLoadCommand()
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

	LoadCmd = cobraCmd
	return nil
}
