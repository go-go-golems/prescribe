package history

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

const historySetSlug = "context-git-history-set"

type SetSettings struct {
	EnabledStr     string `glazed.parameter:"enabled"`
	MaxCommits     int    `glazed.parameter:"max-commits"`
	IncludeMerges  bool   `glazed.parameter:"include-merges"`
	FirstParent    bool   `glazed.parameter:"first-parent"`
	IncludeNumstat bool   `glazed.parameter:"include-numstat"`
}

type SetCommand struct {
	*cmds.CommandDescription
}

var _ cmds.BareCommand = &SetCommand{}

func NewSetCommand() (*SetCommand, error) {
	repoLayer, err := prescribe_layers.NewRepositoryLayer()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository layer")
	}
	repoLayerExisting, err := prescribe_layers.WrapAsExistingCobraFlagsLayer(repoLayer)
	if err != nil {
		return nil, errors.Wrap(err, "failed to wrap repository layer as existing flags layer")
	}

	setLayer, err := schema.NewSection(
		historySetSlug,
		"Git History Set",
		schema.WithFields(
			fields.New(
				"enabled",
				fields.TypeString,
				fields.WithDefault(""),
				fields.WithHelp("Set enabled (true/false)"),
			),
			fields.New(
				"max-commits",
				fields.TypeInteger,
				fields.WithDefault(0),
				fields.WithHelp("Set max_commits (positive integer)"),
			),
			fields.New(
				"include-merges",
				fields.TypeBool,
				fields.WithDefault(false),
				fields.WithHelp("Set include_merges (true/false)"),
			),
			fields.New(
				"first-parent",
				fields.TypeBool,
				fields.WithDefault(false),
				fields.WithHelp("Set first_parent (true/false)"),
			),
			fields.New(
				"include-numstat",
				fields.TypeBool,
				fields.WithDefault(false),
				fields.WithHelp("Set include_numstat (true/false)"),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	cmdDesc := cmds.NewCommandDescription(
		"set",
		cmds.WithShort("Update git history config"),
		cmds.WithLong("Update git history config fields in session.yaml (only provided flags are applied)."),
		cmds.WithLayersList(repoLayerExisting, setLayer),
	)

	return &SetCommand{CommandDescription: cmdDesc}, nil
}

func (c *SetCommand) Run(ctx context.Context, parsedLayers *glazed_layers.ParsedLayers) error {
	_ = ctx

	settings := &SetSettings{}
	if err := parsedLayers.InitializeStruct(historySetSlug, settings); err != nil {
		return errors.Wrap(err, "failed to initialize history set settings")
	}

	ctrl, err := helpers.NewInitializedControllerFromParsedLayers(parsedLayers)
	if err != nil {
		return err
	}
	helpers.LoadDefaultSessionIfExists(ctrl)

	data := ctrl.GetData()
	cfg, _ := effectiveGitHistoryConfig(data)

	if parameterWasSet(parsedLayers, historySetSlug, "enabled") {
		v, err := strconv.ParseBool(settings.EnabledStr)
		if err != nil {
			return fmt.Errorf("invalid --enabled value %q (expected true/false)", settings.EnabledStr)
		}
		cfg.Enabled = v
	}
	if parameterWasSet(parsedLayers, historySetSlug, "max-commits") {
		if settings.MaxCommits <= 0 {
			return fmt.Errorf("--max-commits must be > 0 (got %d)", settings.MaxCommits)
		}
		cfg.MaxCommits = settings.MaxCommits
	}
	if parameterWasSet(parsedLayers, historySetSlug, "include-merges") {
		cfg.IncludeMerges = settings.IncludeMerges
	}
	if parameterWasSet(parsedLayers, historySetSlug, "first-parent") {
		cfg.FirstParent = settings.FirstParent
	}
	if parameterWasSet(parsedLayers, historySetSlug, "include-numstat") {
		cfg.IncludeNumstat = settings.IncludeNumstat
	}

	data.GitHistory = &cfg

	savePath := ctrl.GetDefaultSessionPath()
	if err := ctrl.SaveSession(savePath); err != nil {
		return err
	}

	fmt.Printf("Git history config updated\n")
	fmt.Printf("Session saved: %s\n", savePath)
	return nil
}

func NewSetCobraCommand() (*cobra.Command, error) {
	glazedCmd, err := NewSetCommand()
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
