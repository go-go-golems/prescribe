package filter

import (
	"context"
	"fmt"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	glazed_layers "github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/helpers"
	"github.com/go-go-golems/prescribe/internal/domain"
	prescribe_layers "github.com/go-go-golems/prescribe/pkg/layers"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var AddFilterCmd *cobra.Command

const filterAddSlug = "filter-add"

type FilterAddSettings struct {
	Name        string   `glazed.parameter:"name"`
	Description string   `glazed.parameter:"description"`
	Exclude     []string `glazed.parameter:"exclude"`
	Include     []string `glazed.parameter:"include"`
}

type FilterAddCommand struct {
	*cmds.CommandDescription
}

var _ cmds.BareCommand = &FilterAddCommand{}

func NewFilterAddCommand() (*FilterAddCommand, error) {
	repoLayer, err := prescribe_layers.NewRepositoryLayer()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository layer")
	}
	// repo/target are persistent flags on the root command.
	repoLayerExisting, err := prescribe_layers.WrapAsExistingCobraFlagsLayer(repoLayer)
	if err != nil {
		return nil, errors.Wrap(err, "failed to wrap repository layer as existing flags layer")
	}

	addLayer, err := schema.NewSection(
		filterAddSlug,
		"Filter Add",
		schema.WithFields(
			fields.New(
				"name",
				fields.TypeString,
				fields.WithDefault(""),
				fields.WithHelp("Filter name (required)"),
				fields.WithShortFlag("n"),
				fields.WithRequired(true),
			),
			fields.New(
				"description",
				fields.TypeString,
				fields.WithDefault(""),
				fields.WithHelp("Filter description"),
				fields.WithShortFlag("d"),
			),
			fields.New(
				"exclude",
				fields.TypeStringList,
				fields.WithDefault([]string{}),
				fields.WithHelp("Exclude patterns (can specify multiple)"),
				fields.WithShortFlag("e"),
			),
			fields.New(
				"include",
				fields.TypeStringList,
				fields.WithDefault([]string{}),
				fields.WithHelp("Include patterns (can specify multiple)"),
				fields.WithShortFlag("i"),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	cmdDesc := cmds.NewCommandDescription(
		"add",
		cmds.WithShort("Add a filter to the session"),
		cmds.WithLong("Add a file filter to the current session."),
		cmds.WithLayersList(
			repoLayerExisting,
			addLayer,
		),
	)

	return &FilterAddCommand{CommandDescription: cmdDesc}, nil
}

func (c *FilterAddCommand) Run(ctx context.Context, parsedLayers *glazed_layers.ParsedLayers) error {
	_ = ctx

	settings := &FilterAddSettings{}
	if err := parsedLayers.InitializeStruct(filterAddSlug, settings); err != nil {
		return errors.Wrap(err, "failed to initialize filter add settings")
	}

	if len(settings.Exclude) == 0 && len(settings.Include) == 0 {
		return errors.New("at least one pattern is required (--exclude or --include)")
	}

	ctrl, err := helpers.NewInitializedControllerFromParsedLayers(parsedLayers)
	if err != nil {
		return err
	}

	// Load existing session if present so we don't clobber it on save.
	helpers.LoadDefaultSessionIfExists(ctrl)

	rules := make([]domain.FilterRule, 0)
	for i, pattern := range settings.Exclude {
		rules = append(rules, domain.FilterRule{
			Type:    domain.FilterTypeExclude,
			Pattern: pattern,
			Order:   i,
		})
	}
	for i, pattern := range settings.Include {
		rules = append(rules, domain.FilterRule{
			Type:    domain.FilterTypeInclude,
			Pattern: pattern,
			Order:   len(settings.Exclude) + i,
		})
	}

	filter := domain.Filter{
		Name:        settings.Name,
		Description: settings.Description,
		Rules:       rules,
	}
	ctrl.AddFilter(filter)

	savePath := ctrl.GetDefaultSessionPath()
	if err := ctrl.SaveSession(savePath); err != nil {
		return errors.Wrap(err, "failed to save session")
	}

	fmt.Printf("Filter '%s' added and saved to session\n", settings.Name)
	data := ctrl.GetData()
	fmt.Printf("  Files now filtered: %d\n", len(data.GetFilteredFiles()))

	return nil
}

func InitAddFilterCmd() error {
	glazedCmd, err := NewFilterAddCommand()
	if err != nil {
		return err
	}
	cobraCmd, err := cli.BuildCobraCommand(glazedCmd,
		cli.WithParserConfig(cli.CobraParserConfig{
			MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
		}),
	)
	if err != nil {
		return err
	}

	AddFilterCmd = cobraCmd
	return nil
}
