package session

import (
	"context"
	"fmt"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	glazed_layers "github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/helpers"
	"github.com/go-go-golems/prescribe/internal/session"
	prescribe_layers "github.com/go-go-golems/prescribe/pkg/layers"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// ShowCmd is built by InitShowCmd() and registered by session/session.go.
var ShowCmd *cobra.Command

const sessionShowSlug = "session-show"

type SessionShowSettings struct {
	YAML bool `glazed.parameter:"yaml"`
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

	showLayer, err := schema.NewSection(
		sessionShowSlug,
		"Session Show",
		schema.WithFields(
			fields.New(
				"yaml",
				fields.TypeBool,
				fields.WithDefault(false),
				fields.WithHelp("Show session as YAML (classic mode only; use --output yaml with --with-glaze-output)"),
				fields.WithShortFlag("y"),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	cmdDesc := cmds.NewCommandDescription(
		"show",
		cmds.WithShort("Show current session state"),
		cmds.WithLong("Display the current PR builder session state."),
		cmds.WithLayersList(
			repoLayerExisting,
			showLayer,
		),
	)

	return &SessionShowCommand{CommandDescription: cmdDesc}, nil
}

func (c *SessionShowCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *glazed_layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &SessionShowSettings{}
	if err := parsedLayers.InitializeStruct(sessionShowSlug, settings); err != nil {
		return errors.Wrap(err, "failed to initialize session show settings")
	}
	if settings.YAML {
		return errors.New("cannot use --yaml with --with-glaze-output (use --output yaml)")
	}

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

func runSessionShowClassic(ctx context.Context, parsedLayers *glazed_layers.ParsedLayers) error {
	_ = ctx

	settings := &SessionShowSettings{}
	if err := parsedLayers.InitializeStruct(sessionShowSlug, settings); err != nil {
		return errors.Wrap(err, "failed to initialize session show settings")
	}

	ctrl, err := helpers.NewInitializedControllerFromParsedLayers(parsedLayers)
	if err != nil {
		return err
	}

	helpers.LoadDefaultSessionIfExists(ctrl)

	data := ctrl.GetData()

	if settings.YAML {
		// Show as YAML
		sess := session.NewSession(data)
		yamlData, err := yaml.Marshal(sess)
		if err != nil {
			return fmt.Errorf("failed to marshal session: %w", err)
		}
		fmt.Print(string(yamlData))
		return nil
	}

	// Show human-readable format
	fmt.Printf("PR Builder Session\n")
	fmt.Printf("==================\n\n")

	fmt.Printf("Branches:\n")
	fmt.Printf("  Source: %s\n", data.SourceBranch)
	fmt.Printf("  Target: %s\n", data.TargetBranch)
	fmt.Printf("\n")

	fmt.Printf("Files: %d total\n", len(data.ChangedFiles))
	visibleFiles := data.GetVisibleFiles()
	includedCount := 0
	for _, f := range visibleFiles {
		if f.Included {
			includedCount++
		}
	}
	fmt.Printf("  Visible: %d\n", len(visibleFiles))
	fmt.Printf("  Included: %d\n", includedCount)
	fmt.Printf("  Filtered: %d\n", len(data.GetFilteredFiles()))
	fmt.Printf("\n")

	if len(data.ActiveFilters) > 0 {
		fmt.Printf("Active Filters:\n")
		for _, filter := range data.ActiveFilters {
			fmt.Printf("  - %s: %s\n", filter.Name, filter.Description)
			for _, rule := range filter.Rules {
				fmt.Printf("      %s: %s\n", rule.Type, rule.Pattern)
			}
		}
		fmt.Printf("\n")
	}

	if len(data.AdditionalContext) > 0 {
		fmt.Printf("Additional Context:\n")
		for _, ctx := range data.AdditionalContext {
			if ctx.Type == "file" {
				fmt.Printf("  - File: %s\n", ctx.Path)
			} else {
				preview := ctx.Content
				if len(preview) > 60 {
					preview = preview[:60] + "..."
				}
				fmt.Printf("  - Note: %s\n", preview)
			}
		}
		fmt.Printf("\n")
	}

	fmt.Printf("Prompt:\n")
	if data.CurrentPreset != nil {
		fmt.Printf("  Preset: %s (%s)\n", data.CurrentPreset.Name, data.CurrentPreset.ID)
	} else {
		preview := data.CurrentPrompt
		if len(preview) > 100 {
			preview = preview[:100] + "..."
		}
		fmt.Printf("  Template: %s\n", preview)
	}
	fmt.Printf("\n")

	fmt.Printf("Token Count: %d\n", data.GetTotalTokens())

	return nil
}

func InitShowCmd() error {
	glazedCmd, err := NewSessionShowCommand()
	if err != nil {
		return err
	}

	cobraCmd, err := cli.BuildCobraCommandFromCommandAndFunc(
		glazedCmd,
		runSessionShowClassic,
		cli.WithDualMode(true),
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
