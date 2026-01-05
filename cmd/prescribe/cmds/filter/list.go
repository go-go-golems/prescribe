package filter

import (
	"context"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	glazed_layers "github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/helpers"
	prescribe_layers "github.com/go-go-golems/prescribe/pkg/layers"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type FilterListCommand struct {
	*cmds.CommandDescription
}

var _ cmds.GlazeCommand = &FilterListCommand{}

func NewFilterListCommand() (*FilterListCommand, error) {
	repoLayer, err := prescribe_layers.NewRepositoryLayer()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository layer")
	}
	// repo/target are currently persistent flags on the root command. Wrapping prevents
	// "flag redefined" errors while still allowing parsing from inherited flags.
	repoLayerExisting, err := prescribe_layers.WrapAsExistingCobraFlagsLayer(repoLayer)
	if err != nil {
		return nil, errors.Wrap(err, "failed to wrap repository layer as existing flags layer")
	}

	cmdDesc := cmds.NewCommandDescription(
		"list",
		cmds.WithShort("List all active filters"),
		cmds.WithLong("Display all active filters in the current session."),
		cmds.WithLayersList(
			repoLayerExisting,
		),
	)

	return &FilterListCommand{CommandDescription: cmdDesc}, nil
}

func (c *FilterListCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *glazed_layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	ctrl, err := helpers.NewInitializedControllerFromParsedLayers(parsedLayers)
	if err != nil {
		return err
	}

	helpers.LoadDefaultSessionIfExists(ctrl)

	filters := ctrl.GetFilters()
	for i, f := range filters {
		if len(f.Rules) == 0 {
			row := types.NewRow(
				types.MRP("filter_index", i),
				types.MRP("filter_name", f.Name),
				types.MRP("filter_description", f.Description),
				types.MRP("rule_index", nil),
				types.MRP("rule_type", nil),
				types.MRP("rule_pattern", nil),
			)
			if err := gp.AddRow(ctx, row); err != nil {
				return err
			}
			continue
		}

		for j, r := range f.Rules {
			row := types.NewRow(
				types.MRP("filter_index", i),
				types.MRP("filter_name", f.Name),
				types.MRP("filter_description", f.Description),
				types.MRP("rule_index", j),
				types.MRP("rule_type", r.Type),
				types.MRP("rule_pattern", r.Pattern),
			)
			if err := gp.AddRow(ctx, row); err != nil {
				return err
			}
		}
	}

	return nil
}

func NewListCobraCommand() (*cobra.Command, error) {
	glazedCmd, err := NewFilterListCommand()
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
