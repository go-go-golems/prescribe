package cmds

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	glazed_layers "github.com/go-go-golems/glazed/pkg/cmds/layers"
	cmd_middlewares "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/prescribe/internal/domain"
	"github.com/go-go-golems/prescribe/internal/git"
	"github.com/go-go-golems/prescribe/internal/github"
	"github.com/go-go-golems/prescribe/internal/prdata"
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
		parameters.WithHelp("Use the last generated PR data saved under .pr-builder"),
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
		parameters.WithHelp("Base branch for PR"),
		parameters.WithDefault("main"),
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

	repoSettings, err := prescribe_layers.GetRepositorySettings(parsedLayers)
	if err != nil {
		return err
	}

	var sourceDesc string
	title := strings.TrimSpace(extra.Title)
	body := strings.TrimSpace(extra.Body)

	if strings.TrimSpace(extra.YAMLFile) != "" {
		p, err := prdata.LoadGeneratedPRDataFromYAMLFile(extra.YAMLFile)
		if err != nil {
			return err
		}
		sourceDesc = "yaml-file:" + extra.YAMLFile
		title = p.Title
		body = p.Body
	} else if extra.UseLast {
		path := prdata.LastGeneratedPRDataPath(repoSettings.RepoPath)
		p, err := prdata.LoadGeneratedPRDataFromYAMLFile(path)
		if err != nil {
			return errors.Wrapf(err, "failed to load last generated PR data (expected at %s)", path)
		}
		sourceDesc = "use-last:" + path
		title = p.Title
		body = p.Body
	} else {
		sourceDesc = "flags"
	}

	// Allow explicit overrides even when loading from YAML/use-last.
	if strings.TrimSpace(extra.Title) != "" {
		title = extra.Title
	}
	if strings.TrimSpace(extra.Body) != "" {
		body = extra.Body
	}

	opts := github.CreatePROptions{
		Title: title,
		Body:  body,
		Base:  extra.Base,
		Draft: extra.Draft,
	}

	args, err := github.BuildGhCreatePRArgs(opts)
	if err != nil {
		return err
	}

	if extra.DryRun {
		fmt.Println("Dry-run: would push branch and create PR via GitHub CLI:")
		fmt.Printf("  repo: %s\n", repoSettings.RepoPath)
		fmt.Printf("  source: %s\n", sourceDesc)
		fmt.Printf("  command: git push\n")
		fmt.Printf("  command: gh %s\n", strings.Join(github.RedactGhArgs(args), " "))
		fmt.Printf("  title_len=%d body_len=%d base=%q draft=%v\n", len(opts.Title), len(opts.Body), opts.Base, opts.Draft)
		return nil
	}

	gitSvc, err := git.NewService(repoSettings.RepoPath)
	if err != nil {
		return err
	}
	if err := gitSvc.PushCurrentBranch(ctx); err != nil {
		failPath := prdata.FailurePRDataPath(repoSettings.RepoPath, time.Now())
		_ = prdata.WriteGeneratedPRDataToYAMLFile(failPath, &domain.GeneratedPRData{Title: opts.Title, Body: opts.Body})
		fmt.Fprintf(os.Stderr, "PR creation failed during git push; saved PR data to %s\n", failPath)
		return err
	}

	svc := github.NewService(repoSettings.RepoPath)
	out, err := svc.CreatePR(ctx, opts)
	if err != nil {
		failPath := prdata.FailurePRDataPath(repoSettings.RepoPath, time.Now())
		_ = prdata.WriteGeneratedPRDataToYAMLFile(failPath, &domain.GeneratedPRData{Title: opts.Title, Body: opts.Body})
		fmt.Fprintf(os.Stderr, "PR creation failed during gh pr create; saved PR data to %s\n", failPath)
		return err
	}

	fmt.Print(out)
	return nil
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
