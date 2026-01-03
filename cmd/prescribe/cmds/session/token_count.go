package session

import (
	"context"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	glazed_layers "github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds/helpers"
	"github.com/go-go-golems/prescribe/internal/domain"
	"github.com/go-go-golems/prescribe/internal/tokens"
	prescribe_layers "github.com/go-go-golems/prescribe/pkg/layers"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// TokenCountCmd is built by InitTokenCountCmd() and registered by session/session.go.
var TokenCountCmd *cobra.Command

type SessionTokenCountSettings struct {
	All             bool `glazed.parameter:"all"`
	IncludeFiltered bool `glazed.parameter:"include-filtered"`
}

type SessionTokenCountCommand struct {
	*cmds.CommandDescription
}

var _ cmds.GlazeCommand = &SessionTokenCountCommand{}

func NewSessionTokenCountCommand() (*SessionTokenCountCommand, error) {
	repoLayer, err := prescribe_layers.NewRepositoryLayer()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create repository layer")
	}
	repoLayerExisting, err := prescribe_layers.WrapAsExistingCobraFlagsLayer(repoLayer)
	if err != nil {
		return nil, errors.Wrap(err, "failed to wrap repository layer as existing flags layer")
	}

	allFlag := parameters.NewParameterDefinition(
		"all",
		parameters.ParameterTypeBool,
		parameters.WithHelp("Include not-included files in the breakdown (still respects filters unless --include-filtered is set)"),
		parameters.WithDefault(false),
	)
	includeFilteredFlag := parameters.NewParameterDefinition(
		"include-filtered",
		parameters.ParameterTypeBool,
		parameters.WithHelp("Include files that are currently filtered out (in addition to visible files)"),
		parameters.WithDefault(false),
	)

	cmdDesc := cmds.NewCommandDescription(
		"token-count",
		cmds.WithShort("Show token breakdown for current session context"),
		cmds.WithLong("Display a per-element token breakdown for the current session context (included files + additional context)."),
		cmds.WithFlags(allFlag, includeFilteredFlag),
		cmds.WithLayersList(
			repoLayerExisting,
		),
	)

	return &SessionTokenCountCommand{CommandDescription: cmdDesc}, nil
}

func fileModeString(f domain.FileChange) string {
	if f.Type == domain.FileTypeDiff {
		return "diff"
	}
	switch f.Version {
	case domain.FileVersionBefore:
		return "full_before"
	case domain.FileVersionAfter:
		return "full_after"
	case domain.FileVersionBoth:
		return "full_both"
	default:
		return "full"
	}
}

// effectiveFileContent mirrors the selection logic used when building generation context/prompt vars:
// prefer after/before content for full-file mode; fall back to diff as a best-effort.
func effectiveFileContent(f domain.FileChange) (string, string) {
	if f.Type == domain.FileTypeDiff {
		return strings.TrimRight(f.Diff, "\n"), "diff"
	}
	// Full-file mode
	switch f.Version {
	case domain.FileVersionBefore:
		if f.FullBefore != "" {
			return strings.TrimRight(f.FullBefore, "\n"), "full_before"
		}
	case domain.FileVersionAfter:
		if f.FullAfter != "" {
			return strings.TrimRight(f.FullAfter, "\n"), "full_after"
		}
	case domain.FileVersionBoth:
		// For "both", treat before+after as two contributions.
		// We'll represent this as a concatenation with a newline separator for "effective" counting.
		// (This is only a diagnostic; the actual prompt template may format these separately.)
		before := strings.TrimRight(f.FullBefore, "\n")
		after := strings.TrimRight(f.FullAfter, "\n")
		if before != "" && after != "" {
			return before + "\n" + after, "full_both"
		}
		if before != "" {
			return before, "full_before"
		}
		if after != "" {
			return after, "full_after"
		}
	}

	// Best-effort fallback (keeps diagnostics resilient to missing full content)
	if strings.TrimSpace(f.FullAfter) != "" {
		return strings.TrimRight(f.FullAfter, "\n"), "full_after"
	}
	if strings.TrimSpace(f.FullBefore) != "" {
		return strings.TrimRight(f.FullBefore, "\n"), "full_before"
	}
	return strings.TrimRight(f.Diff, "\n"), "diff_fallback"
}

func effectiveContextContent(c domain.ContextItem) string {
	// Generation prompt vars/export trim trailing newlines for file content.
	return strings.TrimRight(c.Content, "\n")
}

func (c *SessionTokenCountCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *glazed_layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &SessionTokenCountSettings{}
	if err := parsedLayers.InitializeStruct(glazed_layers.DefaultSlug, settings); err != nil {
		return errors.Wrap(err, "failed to decode token-count settings")
	}

	ctrl, err := helpers.NewInitializedControllerFromParsedLayers(parsedLayers)
	if err != nil {
		return err
	}
	helpers.LoadDefaultSessionIfExists(ctrl)

	data := ctrl.GetData()
	encoding := tokens.EncodingName()

	visibleFiles := data.GetVisibleFiles()
	visibleSet := map[string]bool{}
	for _, f := range visibleFiles {
		visibleSet[f.Path] = true
	}

	filesToReport := make([]domain.FileChange, 0)
	if settings.IncludeFiltered {
		filesToReport = append(filesToReport, data.ChangedFiles...)
	} else {
		filesToReport = append(filesToReport, visibleFiles...)
	}

	storedTotal := 0
	effectiveTotal := 0

	for _, f := range filesToReport {
		isVisible := visibleSet[f.Path]
		visibility := "filtered"
		if isVisible {
			visibility = "visible"
		}

		if !settings.All && !f.Included {
			continue
		}

		effContent, effMode := effectiveFileContent(f)
		effTokens := tokens.Count(effContent)

		// "stored" tokens are what PRData.GetTotalTokens() uses (for visible+included files).
		// For filtered files, we still report f.Tokens but do not add it to the storedTotal.
		if isVisible && f.Included {
			storedTotal += f.Tokens
			effectiveTotal += effTokens
		}

		row := types.NewRow(
			types.MRP("kind", "file"),
			types.MRP("encoding", encoding),
			types.MRP("visibility", visibility),
			types.MRP("path", f.Path),
			types.MRP("mode", fileModeString(f)),
			types.MRP("effective_mode", effMode),
			types.MRP("included", f.Included),
			types.MRP("additions", f.Additions),
			types.MRP("deletions", f.Deletions),
			types.MRP("tokens_stored", f.Tokens),
			types.MRP("tokens_effective", effTokens),
			types.MRP("tokens_delta", f.Tokens-effTokens),
			types.MRP("bytes_effective", len(effContent)),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	// Additional context (always counted in stored total)
	for i, ac := range data.AdditionalContext {
		eff := effectiveContextContent(ac)
		effTokens := tokens.Count(eff)
		storedTotal += ac.Tokens
		effectiveTotal += effTokens

		row := types.NewRow(
			types.MRP("kind", "context"),
			types.MRP("encoding", encoding),
			types.MRP("context_index", i),
			types.MRP("context_type", string(ac.Type)),
			types.MRP("path", ac.Path),
			types.MRP("tokens_stored", ac.Tokens),
			types.MRP("tokens_effective", effTokens),
			types.MRP("tokens_delta", ac.Tokens-effTokens),
			types.MRP("bytes_effective", len(eff)),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	// Git history is currently derived at generation time (not persisted in session.yaml).
	// To keep token-count aligned with what `generate` will send, we compute it via the canonical request builder.
	if req, err := ctrl.BuildGenerateDescriptionRequest(); err == nil {
		for _, ac := range req.AdditionalContext {
			if ac.Type != domain.ContextTypeGitHistory {
				continue
			}
			eff := effectiveContextContent(ac)
			if strings.TrimSpace(eff) == "" {
				continue
			}
			effTokens := tokens.Count(eff)
			storedTotal += ac.Tokens
			effectiveTotal += effTokens

			row := types.NewRow(
				types.MRP("kind", "git_history"),
				types.MRP("encoding", encoding),
				types.MRP("path", ac.Path),
				types.MRP("tokens_stored", ac.Tokens),
				types.MRP("tokens_effective", effTokens),
				types.MRP("tokens_delta", ac.Tokens-effTokens),
				types.MRP("bytes_effective", len(eff)),
			)
			if err := gp.AddRow(ctx, row); err != nil {
				return err
			}
		}
	}

	// Summary row (matches session show semantics for "stored_total": visible+included files + all additional context)
	summary := types.NewRow(
		types.MRP("kind", "total"),
		types.MRP("encoding", encoding),
		types.MRP("stored_total", storedTotal),
		types.MRP("effective_total", effectiveTotal),
		types.MRP("delta", storedTotal-effectiveTotal),
	)
	return gp.AddRow(ctx, summary)
}

func InitTokenCountCmd() error {
	glazedCmd, err := NewSessionTokenCountCommand()
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

	TokenCountCmd = cobraCmd
	return nil
}
