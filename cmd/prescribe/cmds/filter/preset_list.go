package filter

import (
	"context"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	glazed_layers "github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/helpers"
	prescribe_layers "github.com/go-go-golems/prescribe/pkg/layers"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var FilterPresetListCmd *cobra.Command

const filterPresetListSlug = "filter-preset-list"

type FilterPresetListSettings struct {
	Project bool `glazed.parameter:"project"`
	Global  bool `glazed.parameter:"global"`
	All     bool `glazed.parameter:"all"`
}

type FilterPresetListCommand struct {
	*cmds.CommandDescription
}

var _ cmds.GlazeCommand = &FilterPresetListCommand{}

func NewFilterPresetListCommand() (*FilterPresetListCommand, error) {
	repoLayer, err := prescribe_layers.NewRepositoryLayer()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository layer")
	}
	repoLayerExisting, err := prescribe_layers.WrapAsExistingCobraFlagsLayer(repoLayer)
	if err != nil {
		return nil, errors.Wrap(err, "failed to wrap repository layer as existing flags layer")
	}

	listLayer, err := schema.NewSection(
		filterPresetListSlug,
		"Filter Preset List",
		schema.WithFields(
			fields.New(
				"project",
				fields.TypeBool,
				fields.WithDefault(false),
				fields.WithHelp("List project presets (<repo>/.pr-builder/filters)"),
			),
			fields.New(
				"global",
				fields.TypeBool,
				fields.WithDefault(false),
				fields.WithHelp("List global presets (~/.pr-builder/filters)"),
			),
			fields.New(
				"all",
				fields.TypeBool,
				fields.WithDefault(false),
				fields.WithHelp("List both project and global presets (default when no scope flags are set)"),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	cmdDesc := cmds.NewCommandDescription(
		"list",
		cmds.WithShort("List filter presets"),
		cmds.WithLong("List named filter presets from project/global locations."),
		cmds.WithLayersList(
			repoLayerExisting,
			listLayer,
		),
	)

	return &FilterPresetListCommand{CommandDescription: cmdDesc}, nil
}

func (c *FilterPresetListCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *glazed_layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &FilterPresetListSettings{}
	if err := parsedLayers.InitializeStruct(filterPresetListSlug, settings); err != nil {
		return errors.Wrap(err, "failed to initialize preset list settings")
	}

	ctrl, err := helpers.NewInitializedControllerFromParsedLayers(parsedLayers)
	if err != nil {
		return err
	}

	wantProject := settings.Project
	wantGlobal := settings.Global
	wantAll := settings.All

	// Default behavior: if no scope flags are set, list all.
	if !wantProject && !wantGlobal && !wantAll {
		wantAll = true
	}
	if wantAll {
		wantProject = true
		wantGlobal = true
	}

	if wantProject {
		ps, err := ctrl.LoadProjectFilterPresets()
		if err != nil {
			return errors.Wrap(err, "failed to load project filter presets")
		}
		for _, p := range ps {
			if len(p.Rules) == 0 {
				row := types.NewRow(
					types.MRP("preset_id", p.ID),
					types.MRP("preset_name", p.Name),
					types.MRP("preset_description", p.Description),
					types.MRP("preset_location", p.Location),
					types.MRP("rule_index", nil),
					types.MRP("rule_type", nil),
					types.MRP("rule_pattern", nil),
				)
				if err := gp.AddRow(ctx, row); err != nil {
					return err
				}
				continue
			}
			for i, r := range p.Rules {
				row := types.NewRow(
					types.MRP("preset_id", p.ID),
					types.MRP("preset_name", p.Name),
					types.MRP("preset_description", p.Description),
					types.MRP("preset_location", p.Location),
					types.MRP("rule_index", i),
					types.MRP("rule_type", r.Type),
					types.MRP("rule_pattern", r.Pattern),
				)
				if err := gp.AddRow(ctx, row); err != nil {
					return err
				}
			}
		}
	}

	if wantGlobal {
		ps, err := ctrl.LoadGlobalFilterPresets()
		if err != nil {
			return errors.Wrap(err, "failed to load global filter presets")
		}
		for _, p := range ps {
			if len(p.Rules) == 0 {
				row := types.NewRow(
					types.MRP("preset_id", p.ID),
					types.MRP("preset_name", p.Name),
					types.MRP("preset_description", p.Description),
					types.MRP("preset_location", p.Location),
					types.MRP("rule_index", nil),
					types.MRP("rule_type", nil),
					types.MRP("rule_pattern", nil),
				)
				if err := gp.AddRow(ctx, row); err != nil {
					return err
				}
				continue
			}
			for i, r := range p.Rules {
				row := types.NewRow(
					types.MRP("preset_id", p.ID),
					types.MRP("preset_name", p.Name),
					types.MRP("preset_description", p.Description),
					types.MRP("preset_location", p.Location),
					types.MRP("rule_index", i),
					types.MRP("rule_type", r.Type),
					types.MRP("rule_pattern", r.Pattern),
				)
				if err := gp.AddRow(ctx, row); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func InitFilterPresetListCmd() error {
	glazedCmd, err := NewFilterPresetListCommand()
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
	FilterPresetListCmd = cobraCmd
	return nil
}
