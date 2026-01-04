package file

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

type FileToggleSettings struct {
	Path string `glazed.parameter:"path"`
}

type FileToggleCommand struct {
	*cmds.CommandDescription
}

var _ cmds.BareCommand = &FileToggleCommand{}

func NewFileToggleCommand() (*FileToggleCommand, error) {
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
				fields.WithHelp("Path of the file to toggle (must match a changed file path in the current session)"),
				fields.WithRequired(true),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	cmdDesc := cmds.NewCommandDescription(
		"toggle",
		cmds.WithShort("Toggle file inclusion in session"),
		cmds.WithLong("Toggle whether a file is included in the PR description context."),
		cmds.WithLayersList(
			repoLayerExisting,
			defaultLayer,
		),
	)

	return &FileToggleCommand{CommandDescription: cmdDesc}, nil
}

func (c *FileToggleCommand) Run(ctx context.Context, parsedLayers *glazed_layers.ParsedLayers) error {
	_ = ctx

	settings := &FileToggleSettings{}
	if err := parsedLayers.InitializeStruct(schema.DefaultSlug, settings); err != nil {
		return errors.Wrap(err, "failed to initialize file toggle settings")
	}

	filePath := settings.Path

	ctrl, err := helpers.NewInitializedControllerFromParsedLayers(parsedLayers)
	if err != nil {
		return err
	}

	helpers.LoadDefaultSessionIfExists(ctrl)

	// Find file and toggle
	data := ctrl.GetData()
	found := false
	for i, file := range data.ChangedFiles {
		if file.Path == filePath {
			if err := ctrl.ToggleFileInclusion(i); err != nil {
				return errors.Wrap(err, "failed to toggle file")
			}
			found = true

			// Re-read to ensure we're printing the updated state.
			data = ctrl.GetData()
			newState := "excluded"
			if data.ChangedFiles[i].Included {
				newState = "included"
			}
			fmt.Printf("File '%s' is now %s\n", filePath, newState)
			break
		}
	}

	if !found {
		return errors.Errorf("file not found: %s", filePath)
	}

	// Save session
	savePath := ctrl.GetDefaultSessionPath()
	if err := ctrl.SaveSession(savePath); err != nil {
		return errors.Wrap(err, "failed to save session")
	}

	fmt.Printf("Session saved\n")

	return nil
}

func NewToggleCobraCommand() (*cobra.Command, error) {
	glazedCmd, err := NewFileToggleCommand()
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
