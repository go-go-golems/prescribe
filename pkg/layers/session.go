package layers

import (
	glazed_layers "github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/pkg/errors"
)

const SessionSlug = "session"

type SessionSettings struct {
	SessionPath string `glazed.parameter:"session-path"`
	AutoSave    bool   `glazed.parameter:"auto-save"`
}

func NewSessionLayer() (schema.Section, error) {
	return schema.NewSection(
		SessionSlug,
		"Session Configuration",
		schema.WithFields(
			fields.New(
				"session-path",
				fields.TypeString,
				fields.WithDefault(""),
				fields.WithHelp("Path to session file (default: app default session path)"),
				fields.WithShortFlag("p"),
			),
			fields.New(
				"auto-save",
				fields.TypeBool,
				fields.WithDefault(false),
				fields.WithHelp("Automatically save session after operations"),
			),
		),
	)
}

func GetSessionSettings(parsedLayers *glazed_layers.ParsedLayers) (*SessionSettings, error) {
	if parsedLayers == nil {
		return nil, errors.New("parsedLayers is nil")
	}

	settings := &SessionSettings{}
	if err := parsedLayers.InitializeStruct(SessionSlug, settings); err != nil {
		return nil, errors.Wrap(err, "failed to initialize session settings")
	}

	return settings, nil
}


