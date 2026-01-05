package git

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

type GitContextRemoveSettings struct {
	Index int `glazed.parameter:"index"`
}

type GitContextRemoveCommand struct {
	*cmds.CommandDescription
}

var _ cmds.BareCommand = &GitContextRemoveCommand{}

func NewGitContextRemoveCommand() (*GitContextRemoveCommand, error) {
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
				"index",
				fields.TypeInteger,
				fields.WithHelp("Index of the git_context item to remove"),
				fields.WithRequired(true),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	cmdDesc := cmds.NewCommandDescription(
		"remove",
		cmds.WithShort("Remove a git_context item by index"),
		cmds.WithLong("Remove a git_context item from session.yaml by index."),
		cmds.WithLayersList(
			repoLayerExisting,
			defaultLayer,
		),
	)

	return &GitContextRemoveCommand{CommandDescription: cmdDesc}, nil
}

func (c *GitContextRemoveCommand) Run(ctx context.Context, parsedLayers *glazed_layers.ParsedLayers) error {
	_ = ctx

	settings := &GitContextRemoveSettings{}
	if err := parsedLayers.InitializeStruct(schema.DefaultSlug, settings); err != nil {
		return errors.Wrap(err, "failed to initialize git context remove settings")
	}

	idx := settings.Index

	ctrl, err := helpers.NewInitializedControllerFromParsedLayers(parsedLayers)
	if err != nil {
		return err
	}
	helpers.LoadDefaultSessionIfExists(ctrl)

	data := ctrl.GetData()
	if idx < 0 || idx >= len(data.GitContext) {
		return fmt.Errorf("index out of range: %d", idx)
	}

	data.GitContext = append(data.GitContext[:idx], data.GitContext[idx+1:]...)
	savePath := ctrl.GetDefaultSessionPath()
	if err := ctrl.SaveSession(savePath); err != nil {
		return err
	}
	fmt.Printf("Removed git_context item %d\n", idx)
	fmt.Printf("Session saved: %s\n", savePath)
	return nil
}

func NewRemoveCobraCommand() (*cobra.Command, error) {
	glazedCmd, err := NewGitContextRemoveCommand()
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
