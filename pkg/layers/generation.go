package layers

import (
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	glazed_layers "github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/pkg/errors"
)

const GenerationSlug = "generation"

type GenerationSettings struct {
	Prompt      string `glazed.parameter:"prompt"`
	Preset      string `glazed.parameter:"preset"`
	LoadSession string `glazed.parameter:"load-session"`
	OutputFile  string `glazed.parameter:"output-file"`
	Title       string `glazed.parameter:"title"`
	Description string `glazed.parameter:"description"`
}

func NewGenerationLayer() (schema.Section, error) {
	return schema.NewSection(
		GenerationSlug,
		"Generation Configuration",
		schema.WithFields(
			fields.New(
				"prompt",
				fields.TypeString,
				fields.WithDefault(""),
				fields.WithHelp("Custom prompt text"),
				fields.WithShortFlag("p"),
			),
			fields.New(
				"preset",
				fields.TypeString,
				fields.WithDefault(""),
				fields.WithHelp("Prompt preset ID"),
			),
			fields.New(
				"load-session",
				fields.TypeString,
				fields.WithDefault(""),
				fields.WithHelp("Load session file before generating"),
				fields.WithShortFlag("s"),
			),
			fields.New(
				"output-file",
				fields.TypeString,
				fields.WithDefault(""),
				fields.WithHelp("Output file (default: stdout)"),
				fields.WithShortFlag("o"),
			),
			fields.New(
				"title",
				fields.TypeString,
				fields.WithDefault(""),
				fields.WithHelp("Proposed PR title (overrides session title)"),
			),
			fields.New(
				"description",
				fields.TypeString,
				fields.WithDefault(""),
				fields.WithHelp("PR description/notes (overrides session description)"),
			),
		),
	)
}

func GetGenerationSettings(parsedLayers *glazed_layers.ParsedLayers) (*GenerationSettings, error) {
	if parsedLayers == nil {
		return nil, errors.New("parsedLayers is nil")
	}

	settings := &GenerationSettings{}
	if err := parsedLayers.InitializeStruct(GenerationSlug, settings); err != nil {
		return nil, errors.Wrap(err, "failed to initialize generation settings")
	}

	return settings, nil
}
