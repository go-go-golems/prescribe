package layers

import (
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	glazed_layers "github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/pkg/errors"
)

const FilterSlug = "filter"

type FilterSettings struct {
	Name            string   `glazed.parameter:"filter-name"`
	Description     string   `glazed.parameter:"filter-description"`
	ExcludePatterns []string `glazed.parameter:"exclude-patterns"`
	IncludePatterns []string `glazed.parameter:"include-patterns"`
}

func NewFilterLayer() (schema.Section, error) {
	return schema.NewSection(
		FilterSlug,
		"Filter Configuration",
		schema.WithFields(
			fields.New(
				"filter-name",
				fields.TypeString,
				fields.WithDefault(""),
				fields.WithHelp("Filter name"),
				fields.WithShortFlag("n"),
			),
			fields.New(
				"filter-description",
				fields.TypeString,
				fields.WithDefault(""),
				fields.WithHelp("Filter description"),
				fields.WithShortFlag("d"),
			),
			fields.New(
				"exclude-patterns",
				fields.TypeStringList,
				fields.WithDefault([]string{}),
				fields.WithHelp("Exclude patterns (glob syntax)"),
				fields.WithShortFlag("e"),
			),
			fields.New(
				"include-patterns",
				fields.TypeStringList,
				fields.WithDefault([]string{}),
				fields.WithHelp("Include patterns (glob syntax)"),
				fields.WithShortFlag("i"),
			),
		),
	)
}

func GetFilterSettings(parsedLayers *glazed_layers.ParsedLayers) (*FilterSettings, error) {
	if parsedLayers == nil {
		return nil, errors.New("parsedLayers is nil")
	}

	settings := &FilterSettings{}
	if err := parsedLayers.InitializeStruct(FilterSlug, settings); err != nil {
		return nil, errors.Wrap(err, "failed to initialize filter settings")
	}

	return settings, nil
}
