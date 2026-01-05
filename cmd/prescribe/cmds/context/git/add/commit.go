package add

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
	"github.com/go-go-golems/prescribe/internal/domain"
	prescribe_layers "github.com/go-go-golems/prescribe/pkg/layers"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type CommitSettings struct {
	Ref string `glazed.parameter:"ref"`
}

type CommitCommand struct {
	*cmds.CommandDescription
}

var _ cmds.BareCommand = &CommitCommand{}

func NewCommitCommand() (*CommitCommand, error) {
	repoLayer, err := prescribe_layers.NewRepositoryLayer()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository layer")
	}
	repoLayerExisting, err := prescribe_layers.WrapAsExistingCobraFlagsLayer(repoLayer)
	if err != nil {
		return nil, errors.Wrap(err, "failed to wrap repository layer as existing flags layer")
	}

	defaultLayer, err := schema.NewSection(
		schema.DefaultSlug,
		"Default",
		schema.WithArguments(
			fields.New(
				"ref",
				fields.TypeString,
				fields.WithHelp("Git ref (commit-ish)"),
				fields.WithRequired(true),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	cmdDesc := cmds.NewCommandDescription(
		"commit",
		cmds.WithShort("Add a commit metadata item"),
		cmds.WithLong("Add a commit metadata git_context item to session.yaml."),
		cmds.WithLayersList(repoLayerExisting, defaultLayer),
	)

	return &CommitCommand{CommandDescription: cmdDesc}, nil
}

func (c *CommitCommand) Run(ctx context.Context, parsedLayers *glazed_layers.ParsedLayers) error {
	_ = ctx

	settings := &CommitSettings{}
	if err := parsedLayers.InitializeStruct(schema.DefaultSlug, settings); err != nil {
		return errors.Wrap(err, "failed to initialize commit settings")
	}

	ref := strings.TrimSpace(settings.Ref)
	if ref == "" {
		return fmt.Errorf("ref is required")
	}

	ctrl, err := helpers.NewInitializedControllerFromParsedLayers(parsedLayers)
	if err != nil {
		return err
	}
	helpers.LoadDefaultSessionIfExists(ctrl)

	data := ctrl.GetData()
	data.GitContext = append(data.GitContext, domain.GitContextItem{
		Kind: domain.GitContextItemKindCommit,
		Ref:  ref,
	})

	savePath := ctrl.GetDefaultSessionPath()
	if err := ctrl.SaveSession(savePath); err != nil {
		return err
	}
	fmt.Printf("Added git_context commit ref=%s\n", ref)
	fmt.Printf("Session saved: %s\n", savePath)
	return nil
}

func NewCommitCobraCommand() (*cobra.Command, error) {
	glazedCmd, err := NewCommitCommand()
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
