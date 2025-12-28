package session

import (
	"context"
	"fmt"
	"strings"

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

var InitCmd *cobra.Command

const sessionInitSlug = "session-init"

type SessionInitSettings struct {
	Save        bool   `glazed.parameter:"save"`
	Path        string `glazed.parameter:"path"`
	Title       string `glazed.parameter:"title"`
	Description string `glazed.parameter:"description"`
}

type SessionInitCommand struct {
	*cmds.CommandDescription
}

var _ cmds.BareCommand = &SessionInitCommand{}

func NewSessionInitCommand() (*SessionInitCommand, error) {
	repoLayer, err := prescribe_layers.NewRepositoryLayer()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository layer")
	}
	repoLayerExisting, err := prescribe_layers.WrapAsExistingCobraFlagsLayer(repoLayer)
	if err != nil {
		return nil, errors.Wrap(err, "failed to wrap repository layer as existing flags layer")
	}

	initLayer, err := schema.NewSection(
		sessionInitSlug,
		"Session Init",
		schema.WithFields(
			fields.New(
				"save",
				fields.TypeBool,
				fields.WithDefault(false),
				fields.WithHelp("Save session to disk after initialization"),
			),
			fields.New(
				"path",
				fields.TypeString,
				fields.WithDefault(""),
				fields.WithHelp("Path to save session (default: app default session path)"),
				fields.WithShortFlag("p"),
			),
			fields.New(
				"title",
				fields.TypeString,
				fields.WithDefault(""),
				fields.WithHelp("PR title to persist into session.yaml (only takes effect with --save)"),
			),
			fields.New(
				"description",
				fields.TypeString,
				fields.WithDefault(""),
				fields.WithHelp("PR description/notes to persist into session.yaml (only takes effect with --save)"),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	cmdDesc := cmds.NewCommandDescription(
		"init",
		cmds.WithShort("Initialize a new PR builder session"),
		cmds.WithLong("Initialize a new PR builder session from the current git state."),
		cmds.WithLayersList(
			repoLayerExisting,
			initLayer,
		),
	)

	return &SessionInitCommand{CommandDescription: cmdDesc}, nil
}

func (c *SessionInitCommand) Run(ctx context.Context, parsedLayers *glazed_layers.ParsedLayers) error {
	_ = ctx

	settings := &SessionInitSettings{}
	if err := parsedLayers.InitializeStruct(sessionInitSlug, settings); err != nil {
		return errors.Wrap(err, "failed to initialize session init settings")
	}

	ctrl, err := helpers.NewInitializedControllerFromParsedLayers(parsedLayers)
	if err != nil {
		return err
	}

	// Apply repo defaults (if configured) as part of explicit session initialization.
	// This keeps TUI startup strict (must have a saved session) while still supporting repo-level defaults.
	n, err := ctrl.ApplyDefaultFilterPresetsFromRepoConfig()
	if err != nil {
		return errors.Wrap(err, "failed to apply repo default filter presets")
	}

	data := ctrl.GetData()
	if strings.TrimSpace(settings.Title) != "" {
		data.Title = settings.Title
	}
	if strings.TrimSpace(settings.Description) != "" {
		data.Description = settings.Description
	}

	fmt.Printf("Initialized PR builder session\n")
	fmt.Printf("  Source: %s\n", data.SourceBranch)
	fmt.Printf("  Target: %s\n", data.TargetBranch)
	fmt.Printf("  Files: %d\n", len(data.ChangedFiles))
	if n > 0 {
		fmt.Printf("  Defaults: applied %d filter preset(s)\n", n)
	}

	if settings.Save {
		savePath := settings.Path
		if savePath == "" {
			savePath = ctrl.GetDefaultSessionPath()
		}

		if err := ctrl.SaveSession(savePath); err != nil {
			return errors.Wrap(err, "failed to save session")
		}

		fmt.Printf("\nSession saved to: %s\n", savePath)
	}

	return nil
}

func InitInitCmd() error {
	glazedCmd, err := NewSessionInitCommand()
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

	InitCmd = cobraCmd
	return nil
}
