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

const commitPatchSlug = "context-git-add-commit-patch"

type CommitPatchSettings struct {
	Paths []string `glazed.parameter:"path"`
}

type CommitPatchDefaultSettings struct {
	Ref string `glazed.parameter:"ref"`
}

type CommitPatchCommand struct {
	*cmds.CommandDescription
}

var _ cmds.BareCommand = &CommitPatchCommand{}

func NewCommitPatchCommand() (*CommitPatchCommand, error) {
	repoLayer, err := prescribe_layers.NewRepositoryLayer()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository layer")
	}
	repoLayerExisting, err := prescribe_layers.WrapAsExistingCobraFlagsLayer(repoLayer)
	if err != nil {
		return nil, errors.Wrap(err, "failed to wrap repository layer as existing flags layer")
	}

	pathsLayer, err := schema.NewSection(
		commitPatchSlug,
		"Commit Patch",
		schema.WithFields(
			fields.New(
				"path",
				fields.TypeStringList,
				fields.WithDefault([]string{}),
				fields.WithHelp("Optional path filter (can be repeated)"),
			),
		),
	)
	if err != nil {
		return nil, err
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
		"commit-patch",
		cmds.WithShort("Add a commit patch item (diff text)"),
		cmds.WithLong("Add a commit patch git_context item to session.yaml (optionally path-filtered)."),
		cmds.WithLayersList(repoLayerExisting, pathsLayer, defaultLayer),
	)

	return &CommitPatchCommand{CommandDescription: cmdDesc}, nil
}

func (c *CommitPatchCommand) Run(ctx context.Context, parsedLayers *glazed_layers.ParsedLayers) error {
	_ = ctx

	settings := &CommitPatchSettings{}
	if err := parsedLayers.InitializeStruct(commitPatchSlug, settings); err != nil {
		return errors.Wrap(err, "failed to initialize commit-patch settings")
	}
	defaultSettings := &CommitPatchDefaultSettings{}
	if err := parsedLayers.InitializeStruct(schema.DefaultSlug, defaultSettings); err != nil {
		return errors.Wrap(err, "failed to initialize commit-patch default settings")
	}

	ref := strings.TrimSpace(defaultSettings.Ref)
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
		Kind:  domain.GitContextItemKindCommitPatch,
		Ref:   ref,
		Paths: append([]string{}, settings.Paths...),
	})

	savePath := ctrl.GetDefaultSessionPath()
	if err := ctrl.SaveSession(savePath); err != nil {
		return err
	}
	fmt.Printf("Added git_context commit_patch ref=%s\n", ref)
	fmt.Printf("Session saved: %s\n", savePath)
	return nil
}

func NewCommitPatchCobraCommand() (*cobra.Command, error) {
	glazedCmd, err := NewCommitPatchCommand()
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
