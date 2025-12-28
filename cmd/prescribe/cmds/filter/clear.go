package filter

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

var ClearFiltersCmd *cobra.Command

type FilterClearCommand struct {
	*cmds.CommandDescription
}

var _ cmds.BareCommand = &FilterClearCommand{}

func NewFilterClearCommand() (*FilterClearCommand, error) {
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
		cmds.WithShort("Remove all filters from the session"),
		cmds.WithLong("Remove all active filters, making all files visible."),
		cmds.WithLayersList(
			repoLayerExisting,
		),
	)

	return &FilterClearCommand{CommandDescription: cmdDesc}, nil
}

func (c *FilterClearCommand) Run(ctx context.Context, parsedLayers *glazed_layers.ParsedLayers) error {
	_ = ctx

	ctrl, err := helpers.NewInitializedControllerFromParsedLayers(parsedLayers)
	if err != nil {
		return err
	}

	if err := helpers.LoadDefaultSession(ctrl); err != nil {
		return err
	}

	filterCount := len(ctrl.GetFilters())
	if filterCount == 0 {
		fmt.Println("No filters to clear")
		return nil
	}

	ctrl.ClearFilters()

	savePath := ctrl.GetDefaultSessionPath()
	if err := ctrl.SaveSession(savePath); err != nil {
		return errors.Wrap(err, "failed to save session")
	}

	fmt.Printf("Cleared %d filter(s) and saved session\n", filterCount)
	data := ctrl.GetData()
	fmt.Printf("  All files now visible: %d\n", len(data.ChangedFiles))

	return nil
}

func InitClearFiltersCmd() error {
	glazedCmd, err := NewFilterClearCommand()
	if err != nil {
		return err
	}
	cobraCmd, err := cli.BuildCobraCommand(
		glazedCmd,
		cli.WithParserConfig(cli.CobraParserConfig{
			MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
		}),
	)
	if err != nil {
		return err
	}

	ClearFiltersCmd = cobraCmd
	return nil
}
