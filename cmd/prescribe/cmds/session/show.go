package session

import (
	"context"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	glazed_layers "github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/helpers"
	prescribe_layers "github.com/go-go-golems/prescribe/pkg/layers"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// ShowCmd is built by InitShowCmd() and registered by session/session.go.
var ShowCmd *cobra.Command

type SessionShowSettings struct {
	// No settings for now. Keep the slug in case we add flags later.
}

type SessionShowCommand struct {
	*cmds.CommandDescription
}

var _ cmds.GlazeCommand = &SessionShowCommand{}

func NewSessionShowCommand() (*SessionShowCommand, error) {
	repoLayer, err := prescribe_layers.NewRepositoryLayer()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository layer")
	}
	repoLayerExisting, err := prescribe_layers.WrapAsExistingCobraFlagsLayer(repoLayer)
	if err != nil {
		return nil, errors.Wrap(err, "failed to wrap repository layer as existing flags layer")
	}

	cmdDesc := cmds.NewCommandDescription(
		"show",
		cmds.WithShort("Show current session state"),
		cmds.WithLong("Display the current PR builder session state."),
		cmds.WithLayersList(
			repoLayerExisting,
		),
	)

	return &SessionShowCommand{CommandDescription: cmdDesc}, nil
}

func (c *SessionShowCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *glazed_layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	ctrl, err := helpers.NewInitializedControllerFromParsedLayers(parsedLayers)
	if err != nil {
		return err
	}
	helpers.LoadDefaultSessionIfExists(ctrl)

	data := ctrl.GetData()
	visibleFiles := data.GetVisibleFiles()
	includedCount := 0
	for _, f := range visibleFiles {
		if f.Included {
			includedCount++
		}
	}

	var presetID any = nil
	var presetName any = nil
	var promptPreview any = nil
	if data.CurrentPreset != nil {
		presetID = data.CurrentPreset.ID
		presetName = data.CurrentPreset.Name
	} else {
		preview := data.CurrentPrompt
		if len(preview) > 100 {
			preview = preview[:100] + "..."
		}
		promptPreview = preview
	}

	row := types.NewRow(
		types.MRP("source_branch", data.SourceBranch),
		types.MRP("target_branch", data.TargetBranch),
		types.MRP("total_files", len(data.ChangedFiles)),
		types.MRP("visible_files", len(visibleFiles)),
		types.MRP("included_files", includedCount),
		types.MRP("filtered_files", len(data.GetFilteredFiles())),
		types.MRP("active_filters", len(data.ActiveFilters)),
		types.MRP("additional_context_items", len(data.AdditionalContext)),
		types.MRP("token_count", data.GetTotalTokens()),
		types.MRP("preset_id", presetID),
		types.MRP("preset_name", presetName),
		types.MRP("prompt_preview", promptPreview),
	)
	return gp.AddRow(ctx, row)
}

func InitShowCmd() error {
	glazedCmd, err := NewSessionShowCommand()
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

	ShowCmd = cobraCmd
	return nil
}
