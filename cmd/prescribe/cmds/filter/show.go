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

// ShowFilteredCmd is built by InitShowFilteredCmd() and registered by filter/filter.go.
var ShowFilteredCmd *cobra.Command

type FilterShowCommand struct {
	*cmds.CommandDescription
}

var _ cmds.GlazeCommand = &FilterShowCommand{}

func NewFilterShowCommand() (*FilterShowCommand, error) {
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
		cmds.WithShort("Show files that are filtered out"),
		cmds.WithLong("Display all files that are being filtered out by active filters."),
		cmds.WithLayersList(
			repoLayerExisting,
		),
	)

	return &FilterShowCommand{CommandDescription: cmdDesc}, nil
}

func (c *FilterShowCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *glazed_layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	ctrl, err := helpers.NewInitializedControllerFromParsedLayers(parsedLayers)
	if err != nil {
		return err
	}

	helpers.LoadDefaultSessionIfExists(ctrl)

	filtered := ctrl.GetFilteredFiles()
	visible := ctrl.GetVisibleFiles()
	data := ctrl.GetData()

	for _, f := range filtered {
		row := types.NewRow(
			types.MRP("file_path", f.Path),
			types.MRP("additions", f.Additions),
			types.MRP("deletions", f.Deletions),
			types.MRP("tokens", f.Tokens),
			types.MRP("total_files", len(data.ChangedFiles)),
			types.MRP("visible_files", len(visible)),
			types.MRP("filtered_files", len(filtered)),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

func runFilterShowClassic(ctx context.Context, parsedLayers *glazed_layers.ParsedLayers) error {
	_ = ctx

	ctrl, err := helpers.NewInitializedControllerFromParsedLayers(parsedLayers)
	if err != nil {
		return err
	}

	helpers.LoadDefaultSessionIfExists(ctrl)

	// Get filtered files
	filtered := ctrl.GetFilteredFiles()
	visible := ctrl.GetVisibleFiles()
	data := ctrl.GetData()

	fmt.Printf("File Status\n")
	fmt.Println("==================")
	fmt.Printf("Total files: %d\n", len(data.ChangedFiles))
	fmt.Printf("Visible files: %d\n", len(visible))
	fmt.Printf("Filtered files: %d\n\n", len(filtered))

	if len(filtered) == 0 {
		fmt.Println("No files are being filtered out")
		return nil
	}

	fmt.Printf("Filtered Files:\n")
	for _, file := range filtered {
		fmt.Printf("  ✗ %s (+%d -%d, %dt)\n",
			file.Path,
			file.Additions,
			file.Deletions,
			file.Tokens)
	}

	// Show which filters are active
	filters := ctrl.GetFilters()
	if len(filters) > 0 {
		fmt.Printf("\nActive Filters:\n")
		for i, filter := range filters {
			fmt.Printf("  [%d] %s\n", i, filter.Name)
			for _, rule := range filter.Rules {
				fmt.Printf("      %s: %s\n", rule.Type, rule.Pattern)
			}
		}
	}

	return nil
}

func InitShowFilteredCmd() error {
	glazedCmd, err := NewFilterShowCommand()
	if err != nil {
		return err
	}

	cobraCmd, err := cli.BuildCobraCommandFromCommandAndFunc(
		glazedCmd,
		runFilterShowClassic,
		cli.WithDualMode(true),
		cli.WithParserConfig(cli.CobraParserConfig{
			MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
		}),
	)
	if err != nil {
		return err
	}

	ShowFilteredCmd = cobraCmd
	return nil
}
