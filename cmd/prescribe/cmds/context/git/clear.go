package git

import (
	"context"
	"fmt"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	glazed_layers "github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/helpers"
	prescribe_layers "github.com/go-go-golems/prescribe/pkg/layers"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type GitContextClearCommand struct {
	*cmds.CommandDescription
}

var _ cmds.BareCommand = &GitContextClearCommand{}

func NewGitContextClearCommand() (*GitContextClearCommand, error) {
	repoLayer, err := prescribe_layers.NewRepositoryLayer()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository layer")
	}
	repoLayerExisting, err := prescribe_layers.WrapAsExistingCobraFlagsLayer(repoLayer)
	if err != nil {
		return nil, errors.Wrap(err, "failed to wrap repository layer as existing flags layer")
	}

	cmdDesc := cmds.NewCommandDescription(
		"clear",
		cmds.WithShort("Clear all git_context items"),
		cmds.WithLong("Clear all git_context items in session.yaml."),
		cmds.WithLayersList(repoLayerExisting),
	)

	return &GitContextClearCommand{CommandDescription: cmdDesc}, nil
}

func (c *GitContextClearCommand) Run(ctx context.Context, parsedLayers *glazed_layers.ParsedLayers) error {
	_ = ctx

	ctrl, err := helpers.NewInitializedControllerFromParsedLayers(parsedLayers)
	if err != nil {
		return err
	}
	helpers.LoadDefaultSessionIfExists(ctrl)

	data := ctrl.GetData()
	data.GitContext = nil

	savePath := ctrl.GetDefaultSessionPath()
	if err := ctrl.SaveSession(savePath); err != nil {
		return err
	}
	fmt.Printf("Cleared git_context\n")
	fmt.Printf("Session saved: %s\n", savePath)
	return nil
}

func NewClearCobraCommand() (*cobra.Command, error) {
	glazedCmd, err := NewGitContextClearCommand()
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
