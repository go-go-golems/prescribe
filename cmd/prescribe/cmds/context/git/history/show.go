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

type ShowCommand struct {
	*cmds.CommandDescription
}

var _ cmds.BareCommand = &ShowCommand{}

func NewShowCommand() (*ShowCommand, error) {
	repoLayer, err := prescribe_layers.NewRepositoryLayer()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository layer")
	}
	repoLayerExisting, err := prescribe_layers.WrapAsExistingCobraFlagsLayer(repoLayer)
	if err != nil {
		return nil, errors.Wrap(err, "failed to wrap repository layer as existing flags layer")
	}

	cmdDesc := cmds.NewCommandDescription(
		"show",
		cmds.WithShort("Show effective git history config"),
		cmds.WithLong("Show the effective git_history config (defaults vs session.yaml) and the derived range."),
		cmds.WithLayersList(repoLayerExisting),
	)

	return &ShowCommand{CommandDescription: cmdDesc}, nil
}

func (c *ShowCommand) Run(ctx context.Context, parsedLayers *glazed_layers.ParsedLayers) error {
	_ = ctx

	ctrl, err := helpers.NewInitializedControllerFromParsedLayers(parsedLayers)
	if err != nil {
		return err
	}
	helpers.LoadDefaultSessionIfExists(ctrl)

	data := ctrl.GetData()
	cfg, explicit := effectiveGitHistoryConfig(data)
	rangeSpec := fmt.Sprintf("%s..%s", data.TargetBranch, data.SourceBranch)

	src := "defaults (missing in session.yaml)"
	if explicit {
		src = "session.yaml"
	}

	fmt.Printf("Git history config (%s)\n", src)
	fmt.Printf("  enabled: %v\n", cfg.Enabled)
	fmt.Printf("  max_commits: %d\n", cfg.MaxCommits)
	fmt.Printf("  include_merges: %v\n", cfg.IncludeMerges)
	fmt.Printf("  first_parent: %v\n", cfg.FirstParent)
	fmt.Printf("  include_numstat: %v\n", cfg.IncludeNumstat)
	fmt.Printf("  range: %s\n", rangeSpec)
	return nil
}

func NewShowCobraCommand() (*cobra.Command, error) {
	glazedCmd, err := NewShowCommand()
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
