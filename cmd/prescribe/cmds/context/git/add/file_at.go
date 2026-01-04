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

type FileAtSettings struct {
	Ref  string `glazed.parameter:"ref"`
	Path string `glazed.parameter:"path"`
}

type FileAtCommand struct {
	*cmds.CommandDescription
}

var _ cmds.BareCommand = &FileAtCommand{}

func NewFileAtCommand() (*FileAtCommand, error) {
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
				fields.WithHelp("Git ref"),
				fields.WithRequired(true),
			),
			fields.New(
				"path",
				fields.TypeString,
				fields.WithHelp("File path"),
				fields.WithRequired(true),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	cmdDesc := cmds.NewCommandDescription(
		"file-at",
		cmds.WithShort("Add file content at a ref"),
		cmds.WithLong("Add a file-at-ref git_context item to session.yaml."),
		cmds.WithLayersList(repoLayerExisting, defaultLayer),
	)

	return &FileAtCommand{CommandDescription: cmdDesc}, nil
}

func (c *FileAtCommand) Run(ctx context.Context, parsedLayers *glazed_layers.ParsedLayers) error {
	_ = ctx

	settings := &FileAtSettings{}
	if err := parsedLayers.InitializeStruct(schema.DefaultSlug, settings); err != nil {
		return errors.Wrap(err, "failed to initialize file-at settings")
	}

	ref := strings.TrimSpace(settings.Ref)
	path := strings.TrimSpace(settings.Path)
	if ref == "" || path == "" {
		return fmt.Errorf("ref and path are required")
	}

	ctrl, err := helpers.NewInitializedControllerFromParsedLayers(parsedLayers)
	if err != nil {
		return err
	}
	helpers.LoadDefaultSessionIfExists(ctrl)

	data := ctrl.GetData()
	data.GitContext = append(data.GitContext, domain.GitContextItem{
		Kind: domain.GitContextItemKindFileAtRef,
		Ref:  ref,
		Path: path,
	})

	savePath := ctrl.GetDefaultSessionPath()
	if err := ctrl.SaveSession(savePath); err != nil {
		return err
	}
	fmt.Printf("Added git_context file_at_ref ref=%s path=%s\n", ref, path)
	fmt.Printf("Session saved: %s\n", savePath)
	return nil
}

func NewFileAtCobraCommand() (*cobra.Command, error) {
	glazedCmd, err := NewFileAtCommand()
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
