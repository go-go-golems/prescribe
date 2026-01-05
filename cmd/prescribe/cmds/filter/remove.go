package filter

import (
	"context"
	"fmt"
	"strconv"

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

type FilterRemoveSettings struct {
	IndexOrName string `glazed.parameter:"index-or-name"`
}

type FilterRemoveCommand struct {
	*cmds.CommandDescription
}

var _ cmds.BareCommand = &FilterRemoveCommand{}

func NewFilterRemoveCommand() (*FilterRemoveCommand, error) {
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
				"index-or-name",
				fields.TypeString,
				fields.WithHelp("Filter index (0-based) or filter name"),
				fields.WithRequired(true),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	cmdDesc := cmds.NewCommandDescription(
		"remove",
		cmds.WithShort("Remove a filter from the session"),
		cmds.WithLong("Remove a filter by index (0-based) or name."),
		cmds.WithLayersList(
			repoLayerExisting,
			defaultLayer,
		),
	)

	return &FilterRemoveCommand{CommandDescription: cmdDesc}, nil
}

func (c *FilterRemoveCommand) Run(ctx context.Context, parsedLayers *glazed_layers.ParsedLayers) error {
	_ = ctx

	settings := &FilterRemoveSettings{}
	if err := parsedLayers.InitializeStruct(schema.DefaultSlug, settings); err != nil {
		return errors.Wrap(err, "failed to initialize filter remove settings")
	}

	ctrl, err := helpers.NewInitializedControllerFromParsedLayers(parsedLayers)
	if err != nil {
		return err
	}

	if err := helpers.LoadDefaultSession(ctrl); err != nil {
		return err
	}

	filters := ctrl.GetFilters()
	if len(filters) == 0 {
		return errors.New("no filters to remove")
	}

	selector := settings.IndexOrName

	// Try to parse as index
	index, err := strconv.Atoi(selector)
	if err != nil {
		// Not a number, try to find by name
		found := false
		for i, filter := range filters {
			if filter.Name == selector {
				index = i
				found = true
				break
			}
		}
		if !found {
			return errors.Errorf("filter not found: %s", selector)
		}
	}

	// Validate index
	if index < 0 || index >= len(filters) {
		return errors.Errorf("invalid filter index: %d (valid range: 0-%d)", index, len(filters)-1)
	}

	filterName := filters[index].Name

	if err := ctrl.RemoveFilter(index); err != nil {
		return errors.Wrap(err, "failed to remove filter")
	}

	savePath := ctrl.GetDefaultSessionPath()
	if err := ctrl.SaveSession(savePath); err != nil {
		return errors.Wrap(err, "failed to save session")
	}

	fmt.Printf("Filter '%s' removed and session saved\n", filterName)

	data := ctrl.GetData()
	fmt.Printf("  Visible files: %d\n", len(data.GetVisibleFiles()))
	fmt.Printf("  Filtered files: %d\n", len(data.GetFilteredFiles()))

	return nil
}

func NewRemoveCobraCommand() (*cobra.Command, error) {
	glazedCmd, err := NewFilterRemoveCommand()
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
