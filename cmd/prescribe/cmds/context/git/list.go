package git

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	glazed_layers "github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/helpers"
	"github.com/go-go-golems/prescribe/internal/domain"
	prescribe_layers "github.com/go-go-golems/prescribe/pkg/layers"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type GitContextListCommand struct {
	*cmds.CommandDescription
}

var _ cmds.BareCommand = &GitContextListCommand{}

func NewGitContextListCommand() (*GitContextListCommand, error) {
	repoLayer, err := prescribe_layers.NewRepositoryLayer()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository layer")
	}
	repoLayerExisting, err := prescribe_layers.WrapAsExistingCobraFlagsLayer(repoLayer)
	if err != nil {
		return nil, errors.Wrap(err, "failed to wrap repository layer as existing flags layer")
	}

	cmdDesc := cmds.NewCommandDescription(
		"list",
		cmds.WithShort("List configured git_context items"),
		cmds.WithLong("List configured git_context items from session.yaml."),
		cmds.WithLayersList(repoLayerExisting),
	)

	return &GitContextListCommand{CommandDescription: cmdDesc}, nil
}

func (c *GitContextListCommand) Run(ctx context.Context, parsedLayers *glazed_layers.ParsedLayers) error {
	_ = ctx

	ctrl, err := helpers.NewInitializedControllerFromParsedLayers(parsedLayers)
	if err != nil {
		return err
	}
	helpers.LoadDefaultSessionIfExists(ctrl)

	items := ctrl.GetData().GitContext
	if len(items) == 0 {
		fmt.Println("No git_context items configured")
		return nil
	}

	for i, it := range items {
		switch it.Kind {
		case domain.GitContextItemKindCommit, domain.GitContextItemKindCommitPatch:
			fmt.Printf("[%d] %s ref=%s", i, it.Kind, it.Ref)
			if len(it.Paths) > 0 {
				fmt.Printf(" paths=%s", strings.Join(it.Paths, ","))
			}
			fmt.Printf("\n")
		case domain.GitContextItemKindFileAtRef:
			fmt.Printf("[%d] %s ref=%s path=%s\n", i, it.Kind, it.Ref, it.Path)
		case domain.GitContextItemKindFileDiff:
			fmt.Printf("[%d] %s from=%s to=%s path=%s\n", i, it.Kind, it.From, it.To, it.Path)
		default:
			fmt.Printf("[%d] %s\n", i, it.Kind)
		}
	}

	return nil
}

func NewListCobraCommand() (*cobra.Command, error) {
	glazedCmd, err := NewGitContextListCommand()
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
