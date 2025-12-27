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

var FilterPresetSaveCmd *cobra.Command

const filterPresetSaveSlug = "filter-preset-save"

type FilterPresetSaveSettings struct {
	Name        string   `glazed.parameter:"name"`
	Description string   `glazed.parameter:"description"`
	Project     bool     `glazed.parameter:"project"`
	Global      bool     `glazed.parameter:"global"`
	FromIndex   int      `glazed.parameter:"from_filter_index"`
	Exclude     []string `glazed.parameter:"exclude"`
	Include     []string `glazed.parameter:"include"`
}

type FilterPresetSaveCommand struct {
	*cmds.CommandDescription
}

var _ cmds.BareCommand = &FilterPresetSaveCommand{}

func NewFilterPresetSaveCommand() (*FilterPresetSaveCommand, error) {
	repoLayer, err := prescribe_layers.NewRepositoryLayer()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository layer")
	}
	repoLayerExisting, err := prescribe_layers.WrapAsExistingCobraFlagsLayer(repoLayer)
	if err != nil {
		return nil, errors.Wrap(err, "failed to wrap repository layer as existing flags layer")
	}

	saveLayer, err := schema.NewSection(
		filterPresetSaveSlug,
		"Filter Preset Save",
		schema.WithFields(
			fields.New(
				"name",
				fields.TypeString,
				fields.WithDefault(""),
				fields.WithHelp("Preset name (required)"),
				fields.WithShortFlag("n"),
				fields.WithRequired(true),
			),
			fields.New(
				"description",
				fields.TypeString,
				fields.WithDefault(""),
				fields.WithHelp("Preset description"),
				fields.WithShortFlag("d"),
			),
			fields.New(
				"project",
				fields.TypeBool,
				fields.WithDefault(false),
				fields.WithHelp("Save preset to project presets (<repo>/.pr-builder/filters)"),
			),
			fields.New(
				"global",
				fields.TypeBool,
				fields.WithDefault(false),
				fields.WithHelp("Save preset to global presets (~/.pr-builder/filters)"),
			),
			fields.New(
				"from_filter_index",
				fields.TypeInteger,
				fields.WithDefault(-1),
				fields.WithHelp("Save rules from an existing active filter (by index in `prescribe filter list`)"),
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
		"save",
		cmds.WithShort("Save a filter preset"),
		cmds.WithLong("Save a named filter preset to project or global presets."),
		cmds.WithLayersList(
			repoLayerExisting,
			saveLayer,
		),
	)

	return &FilterPresetSaveCommand{CommandDescription: cmdDesc}, nil
}

func (c *FilterPresetSaveCommand) Run(ctx context.Context, parsedLayers *glazed_layers.ParsedLayers) error {
	_ = ctx

	settings := &FilterPresetSaveSettings{}
	if err := parsedLayers.InitializeStruct(filterPresetSaveSlug, settings); err != nil {
		return errors.Wrap(err, "failed to initialize preset save settings")
	}

	if settings.Project && settings.Global {
		return errors.New("choose exactly one location: --project or --global")
	}
	if !settings.Project && !settings.Global {
		return errors.New("missing location: choose --project or --global")
	}

	location := domain.PresetLocationProject
	if settings.Global {
		location = domain.PresetLocationGlobal
	}

	ctrl, err := helpers.NewInitializedControllerFromParsedLayers(parsedLayers)
	if err != nil {
		return err
	}

	rules := make([]domain.FilterRule, 0)
	if settings.FromIndex >= 0 {
		helpers.LoadDefaultSessionIfExists(ctrl)
		filters := ctrl.GetFilters()
		if settings.FromIndex >= len(filters) {
			return errors.Errorf("invalid --from-filter-index %d (have %d active filters)", settings.FromIndex, len(filters))
		}
		rules = append(rules, filters[settings.FromIndex].Rules...)
	} else {
		if len(settings.Exclude) == 0 && len(settings.Include) == 0 {
			return errors.New("at least one rule source is required: either --from-filter-index or --exclude/--include")
		}
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
	}

	if err := ctrl.SaveFilterPreset(settings.Name, settings.Description, rules, location); err != nil {
		return errors.Wrap(err, "failed to save filter preset")
	}

	fmt.Printf("Filter preset '%s' saved (%s)\n", settings.Name, location)
	return nil
}

func InitFilterPresetSaveCmd() error {
	glazedCmd, err := NewFilterPresetSaveCommand()
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

	FilterPresetSaveCmd = cobraCmd
	return nil
}
