package cmds

import (
	"context"
	"fmt"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	glazed_layers "github.com/go-go-golems/glazed/pkg/cmds/layers"
	cmd_middlewares "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	prescribe_layers "github.com/go-go-golems/prescribe/pkg/layers"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var createCmd *cobra.Command

type CreateCommand struct {
	*cmds.CommandDescription
}

var _ cmds.BareCommand = &CreateCommand{}

type CreateExtraSettings struct {
	UseLast  bool   `glazed.parameter:"use-last"`
	YAMLFile string `glazed.parameter:"yaml-file"`
	Title    string `glazed.parameter:"title"`
	Body     string `glazed.parameter:"body"`
	Draft    bool   `glazed.parameter:"draft"`
	DryRun   bool   `glazed.parameter:"dry-run"`
	Base     string `glazed.parameter:"base"`
}

func NewCreateCommand() (*CreateCommand, error) {
	repoLayer, err := prescribe_layers.NewRepositoryLayer()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository layer")
	}
	repoLayerExisting, err := prescribe_layers.WrapAsExistingCobraFlagsLayer(repoLayer)
	if err != nil {
		return nil, errors.Wrap(err, "failed to wrap repository layer as existing flags layer")
	}

	useLastFlag := parameters.NewParameterDefinition(
		"use-last",
		parameters.ParameterTypeBool,
		parameters.WithHelp("Use the last generated PR data from session"),
		parameters.WithDefault(false),
	)
	yamlFileFlag := parameters.NewParameterDefinition(
		"yaml-file",
		parameters.ParameterTypeString,
		parameters.WithHelp("Path to YAML file containing GeneratedPRData"),
		parameters.WithDefault(""),
	)
	titleFlag := parameters.NewParameterDefinition(
		"title",
		parameters.ParameterTypeString,
		parameters.WithHelp("Override PR title"),
		parameters.WithDefault(""),
	)
	bodyFlag := parameters.NewParameterDefinition(
		"body",
		parameters.ParameterTypeString,
		parameters.WithHelp("Override PR body"),
		parameters.WithDefault(""),
	)
	draftFlag := parameters.NewParameterDefinition(
		"draft",
		parameters.ParameterTypeBool,
		parameters.WithHelp("Create PR as draft"),
		parameters.WithDefault(false),
	)
	dryRunFlag := parameters.NewParameterDefinition(
		"dry-run",
		parameters.ParameterTypeBool,
		parameters.WithHelp("Show what would be created without actually creating PR"),
		parameters.WithDefault(false),
	)
	baseFlag := parameters.NewParameterDefinition(
		"base",
		parameters.ParameterTypeString,
		parameters.WithHelp("Base branch for PR (default: main or detected default branch)"),
		parameters.WithDefault(""),
	)

	layersList := []glazed_layers.ParameterLayer{
		repoLayerExisting,
	}

	cmdDesc := cmds.NewCommandDescription(
		"create",
		cmds.WithShort("Create a GitHub PR"),
		cmds.WithLong("Create a GitHub PR using generated PR data or from a YAML file."),
		cmds.WithFlags(useLastFlag, yamlFileFlag, titleFlag, bodyFlag, draftFlag, dryRunFlag, baseFlag),
		cmds.WithLayersList(
			layersList...,
		),
	)

	return &CreateCommand{CommandDescription: cmdDesc}, nil
}

func (c *CreateCommand) Run(ctx context.Context, parsedLayers *glazed_layers.ParsedLayers) error {
	extra := &CreateExtraSettings{}
	if err := parsedLayers.InitializeStruct(glazed_layers.DefaultSlug, extra); err != nil {
		return errors.Wrap(err, "failed to decode create extra settings")
	}

	// TODO: Implement PR creation logic in subsequent tasks
	fmt.Println("Create command called with flags:")
	fmt.Printf("  --use-last: %v\n", extra.UseLast)
	fmt.Printf("  --yaml-file: %s\n", extra.YAMLFile)
	fmt.Printf("  --title: %s\n", extra.Title)
	fmt.Printf("  --body: %s\n", extra.Body)
	fmt.Printf("  --draft: %v\n", extra.Draft)
	fmt.Printf("  --dry-run: %v\n", extra.DryRun)
	fmt.Printf("  --base: %s\n", extra.Base)

	return fmt.Errorf("PR creation not yet implemented")
}

func InitCreateCmd() error {
	glazedCmd, err := NewCreateCommand()
	if err != nil {
		return errors.Wrap(err, "failed to create create command")
	}

	createMiddlewares := func(parsedCommandLayers *glazed_layers.ParsedLayers, cmd *cobra.Command, args []string) ([]cmd_middlewares.Middleware, error) {
		middlewares_ := []cmd_middlewares.Middleware{
			cmd_middlewares.ParseFromCobraCommand(cmd, parameters.WithParseStepSource("cobra")),
			cmd_middlewares.GatherArguments(args, parameters.WithParseStepSource("arguments")),
			cmd_middlewares.SetFromDefaults(parameters.WithParseStepSource(parameters.SourceDefaults)),
		}
		return middlewares_, nil
	}

	cobraCmd, err := cli.BuildCobraCommand(
		glazedCmd,
		cli.WithParserConfig(cli.CobraParserConfig{
			EnableProfileSettingsLayer: false,
			MiddlewaresFunc:            createMiddlewares,
		}),
	)
	if err != nil {
		return errors.Wrap(err, "failed to build cobra command")
	}

	createCmd = cobraCmd
	return nil
}
