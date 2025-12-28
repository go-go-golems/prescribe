package layers

import (
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	glazed_layers "github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/pkg/errors"
)

const RepositorySlug = "repository"

type RepositorySettings struct {
	RepoPath     string `glazed.parameter:"repo"`
	TargetBranch string `glazed.parameter:"target"`
}

func NewRepositoryLayer() (schema.Section, error) {
	return schema.NewSection(
		RepositorySlug,
		"Repository Configuration",
		schema.WithFields(
			fields.New(
				"repo",
				fields.TypeString,
				fields.WithDefault("."),
				fields.WithHelp("Path to git repository"),
				fields.WithShortFlag("r"),
			),
			fields.New(
				"target",
				fields.TypeString,
				fields.WithDefault(""),
				fields.WithHelp("Target branch (default: main or master)"),
				fields.WithShortFlag("t"),
			),
		),
	)
}

func GetRepositorySettings(parsedLayers *glazed_layers.ParsedLayers) (*RepositorySettings, error) {
	if parsedLayers == nil {
		return nil, errors.New("parsedLayers is nil")
	}

	settings := &RepositorySettings{}
	if err := parsedLayers.InitializeStruct(RepositorySlug, settings); err != nil {
		return nil, errors.Wrap(err, "failed to initialize repository settings")
	}

	return settings, nil
}
