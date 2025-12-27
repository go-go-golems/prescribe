package filter

import (
	"context"
	"fmt"

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

// ListFiltersCmd is built by buildListFiltersCmd() and registered by filter/filter.go.
//
// NOTE: Do not rely on init() ordering across files in this package.
var ListFiltersCmd *cobra.Command

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

func runFilterListClassic(ctx context.Context, parsedLayers *glazed_layers.ParsedLayers) error {
	_ = ctx

	ctrl, err := helpers.NewInitializedControllerFromParsedLayers(parsedLayers)
	if err != nil {
		return err
	}

	helpers.LoadDefaultSessionIfExists(ctrl)

	// Get filters
	filters := ctrl.GetFilters()

	if len(filters) == 0 {
		fmt.Println("No active filters")
		return nil
	}

	fmt.Printf("Active Filters (%d)\n", len(filters))
	fmt.Println("==================")

	for i, filter := range filters {
		fmt.Printf("\n[%d] %s\n", i, filter.Name)
		if filter.Description != "" {
			fmt.Printf("    Description: %s\n", filter.Description)
		}
		fmt.Printf("    Rules: %d\n", len(filter.Rules))
		for j, rule := range filter.Rules {
			fmt.Printf("      [%d] %s: %s\n", j, rule.Type, rule.Pattern)
		}
	}

	// Show impact
	data := ctrl.GetData()
	fmt.Printf("\nImpact:\n")
	fmt.Printf("  Total files: %d\n", len(data.ChangedFiles))
	fmt.Printf("  Visible files: %d\n", len(data.GetVisibleFiles()))
	fmt.Printf("  Filtered files: %d\n", len(data.GetFilteredFiles()))

	return nil
}

func buildListFiltersCmd() (*cobra.Command, error) {
	glazedCmd, err := NewFilterListCommand()
	if err != nil {
		return nil, err
	}

	cobraCmd, err := cli.BuildCobraCommandFromCommandAndFunc(
		glazedCmd,
		runFilterListClassic,
		cli.WithDualMode(true),
		cli.WithParserConfig(cli.CobraParserConfig{
			MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
			// NOTE: repo/target are inherited persistent flags, so we don't add them here.
		}),
	)
	if err != nil {
		return nil, err
	}

	return cobraCmd, nil
}

func InitListFiltersCmd() error {
	cmd, err := buildListFiltersCmd()
	if err != nil {
		return err
	}
	ListFiltersCmd = cmd
	return nil
}
