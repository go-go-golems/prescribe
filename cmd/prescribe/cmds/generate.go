package cmds

import (
	"context"
	"fmt"
	"os"

	geppettolayers "github.com/go-go-golems/geppetto/pkg/layers"
	gepsettings "github.com/go-go-golems/geppetto/pkg/steps/ai/settings"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	glazed_layers "github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/helpers"
	pexport "github.com/go-go-golems/prescribe/internal/export"
	prescribe_layers "github.com/go-go-golems/prescribe/pkg/layers"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var generateCmd *cobra.Command

type GenerateCommand struct {
	*cmds.CommandDescription
}

var _ cmds.BareCommand = &GenerateCommand{}

type GenerateExtraSettings struct {
	ExportContext  bool   `glazed.parameter:"export-context"`
	ExportRendered bool   `glazed.parameter:"export-rendered"`
	Separator      string `glazed.parameter:"separator"`
}

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

	geppettoLayers, err := geppettolayers.CreateGeppettoLayers()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create geppetto parameter layers")
	}

	extraFlags := parameters.NewParameterDefinition(
		"export-context",
		parameters.ParameterTypeBool,
		parameters.WithHelp("Print the full generation context (prompt + files + context) and exit (no inference)"),
		parameters.WithDefault(false),
	)
	exportRenderedFlag := parameters.NewParameterDefinition(
		"export-rendered",
		parameters.ParameterTypeBool,
		parameters.WithHelp("Print the rendered LLM payload (system+user) and exit (no inference)"),
		parameters.WithDefault(false),
	)
	separatorFlag := parameters.NewParameterDefinition(
		"separator",
		parameters.ParameterTypeString,
		parameters.WithHelp("Separator format for export flags: xml (default), markdown, simple, begin-end, default"),
		parameters.WithDefault("xml"),
	)

	layersList := []glazed_layers.ParameterLayer{
		repoLayerExisting,
		generationLayer,
	}
	layersList = append(layersList, geppettoLayers...)

	cmdDesc := cmds.NewCommandDescription(
		"generate",
		cmds.WithShort("Generate PR description"),
		cmds.WithLong("Generate a PR description using AI based on the current session."),
		cmds.WithFlags(extraFlags, exportRenderedFlag, separatorFlag),
		cmds.WithLayersList(
			layersList...,
		),
	)

	return &GenerateCommand{CommandDescription: cmdDesc}, nil
}

func (c *GenerateCommand) Run(ctx context.Context, parsedLayers *glazed_layers.ParsedLayers) error {
	genSettings, err := prescribe_layers.GetGenerationSettings(parsedLayers)
	if err != nil {
		return err
	}

	extra := &GenerateExtraSettings{}
	if err := parsedLayers.InitializeStruct(glazed_layers.DefaultSlug, extra); err != nil {
		return errors.Wrap(err, "failed to decode generate extra settings")
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

	// Export-only path (no inference).
	if extra.ExportContext && extra.ExportRendered {
		return errors.New("flags --export-context and --export-rendered are mutually exclusive")
	}
	if extra.ExportContext || extra.ExportRendered {
		req, err := ctrl.BuildGenerateDescriptionRequest()
		if err != nil {
			return err
		}
		sep := pexport.SeparatorType(extra.Separator)
		text := ""
		if extra.ExportRendered {
			rendered, err := pexport.BuildRenderedLLMPayload(req, sep)
			if err != nil {
				return err
			}
			text = rendered
		} else {
			text = pexport.BuildGenerationContext(req, sep)
		}
		if genSettings.OutputFile != "" {
			if err := os.WriteFile(genSettings.OutputFile, []byte(text), 0644); err != nil {
				return errors.Wrap(err, "failed to write output file")
			}
			what := "Context"
			if extra.ExportRendered {
				what = "Rendered payload"
			}
			fmt.Fprintf(os.Stderr, "%s written to %s\n", what, genSettings.OutputFile)
			return nil
		}
		fmt.Print(text)
		return nil
	}

	// StepSettings parsing happens here (higher up), then injected into API service.
	stepSettings, err := gepsettings.NewStepSettingsFromParsedLayers(parsedLayers)
	if err != nil {
		return errors.Wrap(err, "failed to build AI step settings from parsed layers")
	}
	ctrl.SetStepSettings(stepSettings)

	// Generate description
	fmt.Fprintf(os.Stderr, "Generating PR description...\n")
	description, err := ctrl.GenerateDescription(ctx)
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
