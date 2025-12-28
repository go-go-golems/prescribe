package context

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

var AddCmd *cobra.Command

const contextAddSlug = "context-add"

type ContextAddSettings struct {
	Note string `glazed.parameter:"note"`
}

type ContextAddDefaultSettings struct {
	FilePath string `glazed.parameter:"file-path"`
}

type ContextAddCommand struct {
	*cmds.CommandDescription
}

var _ cmds.BareCommand = &ContextAddCommand{}

func NewContextAddCommand() (*ContextAddCommand, error) {
	repoLayer, err := prescribe_layers.NewRepositoryLayer()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository layer")
	}
	repoLayerExisting, err := prescribe_layers.WrapAsExistingCobraFlagsLayer(repoLayer)
	if err != nil {
		return nil, errors.Wrap(err, "failed to wrap repository layer as existing flags layer")
	}

	addLayer, err := schema.NewSection(
		contextAddSlug,
		"Context Add",
		schema.WithFields(
			fields.New(
				"note",
				fields.TypeString,
				fields.WithDefault(""),
				fields.WithHelp("Add a note as additional context (mutually exclusive with file-path argument)"),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	defaultLayer, err := schema.NewSection(
		schema.DefaultSlug,
		"Default",
		schema.WithArguments(
			fields.New(
				"file-path",
				fields.TypeString,
				fields.WithHelp("Add a file as additional context (mutually exclusive with --note)"),
				fields.WithRequired(false),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	cmdDesc := cmds.NewCommandDescription(
		"add",
		cmds.WithShort("Add additional context to session"),
		cmds.WithLong("Add a file or note as additional context for PR description generation."),
		cmds.WithLayersList(
			repoLayerExisting,
			addLayer,
			defaultLayer,
		),
	)

	return &ContextAddCommand{CommandDescription: cmdDesc}, nil
}

func (c *ContextAddCommand) Run(ctx context.Context, parsedLayers *glazed_layers.ParsedLayers) error {
	_ = ctx

	settings := &ContextAddSettings{}
	if err := parsedLayers.InitializeStruct(contextAddSlug, settings); err != nil {
		return errors.Wrap(err, "failed to initialize context add settings")
	}

	defaultSettings := &ContextAddDefaultSettings{}
	if err := parsedLayers.InitializeStruct(schema.DefaultSlug, defaultSettings); err != nil {
		return errors.Wrap(err, "failed to initialize context add default settings")
	}

	if settings.Note == "" && defaultSettings.FilePath == "" {
		return errors.New("either pass a file-path argument or use --note")
	}
	if settings.Note != "" && defaultSettings.FilePath != "" {
		return errors.New("use either a file-path argument or --note (not both)")
	}

	ctrl, err := helpers.NewInitializedControllerFromParsedLayers(parsedLayers)
	if err != nil {
		return err
	}

	helpers.LoadDefaultSessionIfExists(ctrl)

	if settings.Note != "" {
		ctrl.AddContextNote(settings.Note)
		fmt.Printf("Added note to context\n")
	} else {
		if err := ctrl.AddContextFile(defaultSettings.FilePath); err != nil {
			return errors.Wrap(err, "failed to add file")
		}
		fmt.Printf("Added file '%s' to context\n", defaultSettings.FilePath)
	}

	savePath := ctrl.GetDefaultSessionPath()
	if err := ctrl.SaveSession(savePath); err != nil {
		return errors.Wrap(err, "failed to save session")
	}

	fmt.Printf("Session saved\n")
	data := ctrl.GetData()
	fmt.Printf("Total tokens: %d\n", data.GetTotalTokens())

	return nil
}

func InitAddCmd() error {
	glazedCmd, err := NewContextAddCommand()
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

	AddCmd = cobraCmd
	return nil
}
