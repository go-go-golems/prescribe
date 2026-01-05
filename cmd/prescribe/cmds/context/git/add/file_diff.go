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

const fileDiffSlug = "context-git-add-file-diff"

type FileDiffSettings struct {
	From string `glazed.parameter:"from"`
	To   string `glazed.parameter:"to"`
	Path string `glazed.parameter:"path"`
}

type FileDiffCommand struct {
	*cmds.CommandDescription
}

var _ cmds.BareCommand = &FileDiffCommand{}

func NewFileDiffCommand() (*FileDiffCommand, error) {
	repoLayer, err := prescribe_layers.NewRepositoryLayer()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository layer")
	}
	repoLayerExisting, err := prescribe_layers.WrapAsExistingCobraFlagsLayer(repoLayer)
	if err != nil {
		return nil, errors.Wrap(err, "failed to wrap repository layer as existing flags layer")
	}

	diffLayer, err := schema.NewSection(
		fileDiffSlug,
		"File Diff",
		schema.WithFields(
			fields.New(
				"from",
				fields.TypeString,
				fields.WithHelp("From ref (required)"),
				fields.WithRequired(true),
			),
			fields.New(
				"to",
				fields.TypeString,
				fields.WithHelp("To ref (required)"),
				fields.WithRequired(true),
			),
			fields.New(
				"path",
				fields.TypeString,
				fields.WithHelp("File path (required)"),
				fields.WithRequired(true),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	cmdDesc := cmds.NewCommandDescription(
		"file-diff",
		cmds.WithShort("Add a single-file diff between two refs"),
		cmds.WithLong("Add a file-diff git_context item to session.yaml."),
		cmds.WithLayersList(repoLayerExisting, diffLayer),
	)

	return &FileDiffCommand{CommandDescription: cmdDesc}, nil
}

func (c *FileDiffCommand) Run(ctx context.Context, parsedLayers *glazed_layers.ParsedLayers) error {
	_ = ctx

	settings := &FileDiffSettings{}
	if err := parsedLayers.InitializeStruct(fileDiffSlug, settings); err != nil {
		return errors.Wrap(err, "failed to initialize file-diff settings")
	}

	from := strings.TrimSpace(settings.From)
	to := strings.TrimSpace(settings.To)
	path := strings.TrimSpace(settings.Path)
	if from == "" || to == "" || path == "" {
		return fmt.Errorf("--from, --to, and --path are required")
	}

	ctrl, err := helpers.NewInitializedControllerFromParsedLayers(parsedLayers)
	if err != nil {
		return err
	}
	helpers.LoadDefaultSessionIfExists(ctrl)

	data := ctrl.GetData()
	data.GitContext = append(data.GitContext, domain.GitContextItem{
		Kind: domain.GitContextItemKindFileDiff,
		From: from,
		To:   to,
		Path: path,
	})

	savePath := ctrl.GetDefaultSessionPath()
	if err := ctrl.SaveSession(savePath); err != nil {
		return err
	}
	fmt.Printf("Added git_context file_diff from=%s to=%s path=%s\n", from, to, path)
	fmt.Printf("Session saved: %s\n", savePath)
	return nil
}

func NewFileDiffCobraCommand() (*cobra.Command, error) {
	glazedCmd, err := NewFileDiffCommand()
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
