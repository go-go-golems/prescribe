package history

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

type EnableCommand struct {
	*cmds.CommandDescription
}

var _ cmds.BareCommand = &EnableCommand{}

func NewEnableCommand() (*EnableCommand, error) {
	repoLayer, err := prescribe_layers.NewRepositoryLayer()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository layer")
	}
	repoLayerExisting, err := prescribe_layers.WrapAsExistingCobraFlagsLayer(repoLayer)
	if err != nil {
		return nil, errors.Wrap(err, "failed to wrap repository layer as existing flags layer")
	}

	cmdDesc := cmds.NewCommandDescription(
		"enable",
		cmds.WithShort("Enable derived git history"),
		cmds.WithLong("Enable derived git history by writing git_history.enabled=true to session.yaml."),
		cmds.WithLayersList(repoLayerExisting),
	)

	return &EnableCommand{CommandDescription: cmdDesc}, nil
}

func (c *EnableCommand) Run(ctx context.Context, parsedLayers *glazed_layers.ParsedLayers) error {
	_ = ctx

	ctrl, err := helpers.NewInitializedControllerFromParsedLayers(parsedLayers)
	if err != nil {
		return err
	}
	helpers.LoadDefaultSessionIfExists(ctrl)

	data := ctrl.GetData()
	cfg, _ := effectiveGitHistoryConfig(data)
	cfg.Enabled = true
	data.GitHistory = &cfg

	savePath := ctrl.GetDefaultSessionPath()
	if err := ctrl.SaveSession(savePath); err != nil {
		return err
	}
	fmt.Printf("Git history enabled\n")
	fmt.Printf("Session saved: %s\n", savePath)
	return nil
}

func NewEnableCobraCommand() (*cobra.Command, error) {
	glazedCmd, err := NewEnableCommand()
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
