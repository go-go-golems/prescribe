package cmds

import (
	"context"
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	glazed_layers "github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/helpers"
	prescribe_layers "github.com/go-go-golems/prescribe/pkg/layers"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var generateCmd *cobra.Command

type GenerateCommand struct {
	*cmds.CommandDescription
}

var _ cmds.BareCommand = &GenerateCommand{}

func NewGenerateCommand() (*GenerateCommand, error) {
	repoLayer, err := prescribe_layers.NewRepositoryLayer()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository layer")
	}
	repoLayerExisting, err := prescribe_layers.WrapAsExistingCobraFlagsLayer(repoLayer)
	if err != nil {
		return nil, errors.Wrap(err, "failed to wrap repository layer as existing flags layer")
	}

	generationLayer, err := prescribe_layers.NewGenerationLayer()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create generation layer")
	}

	cmdDesc := cmds.NewCommandDescription(
		"generate",
		cmds.WithShort("Generate PR description"),
		cmds.WithLong("Generate a PR description using AI based on the current session."),
		cmds.WithLayersList(
			repoLayerExisting,
			generationLayer,
		),
	)

	return &GenerateCommand{CommandDescription: cmdDesc}, nil
}

func (c *GenerateCommand) Run(ctx context.Context, parsedLayers *glazed_layers.ParsedLayers) error {
	_ = ctx

	genSettings, err := prescribe_layers.GetGenerationSettings(parsedLayers)
	if err != nil {
		return err
	}

	ctrl, err := helpers.NewInitializedControllerFromParsedLayers(parsedLayers)
	if err != nil {
		return err
	}

	// Prefer loading an existing default session so this command reflects current state.
	helpers.LoadDefaultSessionIfExists(ctrl)

	// Load session if specified (overrides default session).
	if genSettings.LoadSession != "" {
		if err := ctrl.LoadSession(genSettings.LoadSession); err != nil {
			return errors.Wrap(err, "failed to load session")
		}
		fmt.Fprintf(os.Stderr, "Loaded session from: %s\n", genSettings.LoadSession)
	}

	// Override prompt if specified
	if genSettings.Prompt != "" {
		ctrl.SetPrompt(genSettings.Prompt, nil)
	} else if genSettings.Preset != "" {
		if err := ctrl.LoadPromptPreset(genSettings.Preset); err != nil {
			return errors.Wrap(err, "failed to load preset")
		}
	}

	// Generate description
	fmt.Fprintf(os.Stderr, "Generating PR description...\n")
	description, err := ctrl.GenerateDescription()
	if err != nil {
		return errors.Wrap(err, "failed to generate description")
	}

	// Output description
	if genSettings.OutputFile != "" {
		if err := os.WriteFile(genSettings.OutputFile, []byte(description), 0644); err != nil {
			return errors.Wrap(err, "failed to write output file")
		}
		fmt.Fprintf(os.Stderr, "Description written to %s\n", genSettings.OutputFile)
	} else {
		fmt.Println(description)
	}

	return nil
}

func InitGenerateCmd() error {
	glazedCmd, err := NewGenerateCommand()
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

	generateCmd = cobraCmd
	return nil
}
