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
	"github.com/go-go-golems/prescribe/internal/domain"
	prescribe_layers "github.com/go-go-golems/prescribe/pkg/layers"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func InitTestFilterCmd() error {
	cmd, err := buildTestFilterCmd()
	if err != nil {
		return err
	}
	TestFilterCmd = cmd
	return nil
}

var TestFilterCmd *cobra.Command

const filterTestSlug = "filter-test"

type FilterTestSettings struct {
	Name    string   `glazed.parameter:"name"`
	Exclude []string `glazed.parameter:"exclude"`
	Include []string `glazed.parameter:"include"`
}

type FilterTestCommand struct {
	*cmds.CommandDescription
}

var _ cmds.GlazeCommand = &FilterTestCommand{}

func NewFilterTestCommand() (*FilterTestCommand, error) {
	repoLayer, err := prescribe_layers.NewRepositoryLayer()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository layer")
	}
	repoLayerExisting, err := prescribe_layers.WrapAsExistingCobraFlagsLayer(repoLayer)
	if err != nil {
		return nil, errors.Wrap(err, "failed to wrap repository layer as existing flags layer")
	}

	testLayer, err := schema.NewSection(
		filterTestSlug,
		"Filter Test",
		schema.WithFields(
			fields.New(
				"name",
				fields.TypeString,
				fields.WithDefault("test"),
				fields.WithHelp("Filter name for display purposes"),
				fields.WithShortFlag("n"),
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
		"test",
		cmds.WithShort("Test a filter pattern without applying it"),
		cmds.WithLong("Test how a filter would affect files without actually applying it to the session."),
		cmds.WithLayersList(
			repoLayerExisting,
			testLayer,
		),
	)

	return &FilterTestCommand{CommandDescription: cmdDesc}, nil
}

func (c *FilterTestCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *glazed_layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &FilterTestSettings{}
	if err := parsedLayers.InitializeStruct(filterTestSlug, settings); err != nil {
		return errors.Wrap(err, "failed to initialize filter test settings")
	}

	if len(settings.Exclude) == 0 && len(settings.Include) == 0 {
		return errors.New("at least one pattern is required (--exclude or --include)")
	}

	ctrl, err := helpers.NewInitializedControllerFromParsedLayers(parsedLayers)
	if err != nil {
		return err
	}

	rules := buildRules(settings.Exclude, settings.Include)
	filter := domain.Filter{
		Name:  settings.Name,
		Rules: rules,
	}

	matched, unmatched := ctrl.TestFilter(filter)

	total := len(matched) + len(unmatched)
	for _, path := range matched {
		row := types.NewRow(
			types.MRP("filter_name", settings.Name),
			types.MRP("file_path", path),
			types.MRP("matched", true),
			types.MRP("total_files", total),
			types.MRP("matched_files", len(matched)),
			types.MRP("filtered_files", len(unmatched)),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	for _, path := range unmatched {
		row := types.NewRow(
			types.MRP("filter_name", settings.Name),
			types.MRP("file_path", path),
			types.MRP("matched", false),
			types.MRP("total_files", total),
			types.MRP("matched_files", len(matched)),
			types.MRP("filtered_files", len(unmatched)),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

func buildRules(exclude []string, include []string) []domain.FilterRule {
	rules := make([]domain.FilterRule, 0)
	for i, pattern := range exclude {
		rules = append(rules, domain.FilterRule{
			Type:    domain.FilterTypeExclude,
			Pattern: pattern,
			Order:   i,
		})
	}
	for i, pattern := range include {
		rules = append(rules, domain.FilterRule{
			Type:    domain.FilterTypeInclude,
			Pattern: pattern,
			Order:   len(exclude) + i,
		})
	}
	return rules
}

func buildTestFilterCmd() (*cobra.Command, error) {
	glazedCmd, err := NewFilterTestCommand()
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
