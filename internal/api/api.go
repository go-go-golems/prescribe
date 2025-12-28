package api

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-go-golems/geppetto/pkg/inference/engine/factory"
	"github.com/go-go-golems/geppetto/pkg/steps/ai/settings"
	"github.com/go-go-golems/geppetto/pkg/turns"
	"github.com/go-go-golems/prescribe/internal/domain"
	"github.com/go-go-golems/prescribe/internal/tokens"
	"github.com/pkg/errors"
)

// Service provides API operations for generating PR descriptions
type Service struct {
	stepSettings *settings.StepSettings
}

// NewService creates a new API service
func NewService() *Service {
	return &Service{}
}

// SetStepSettings configures the inference engine settings.
// Parsing is expected to happen at a higher layer (CLI/TUI), not inside this service.
func (s *Service) SetStepSettings(stepSettings *settings.StepSettings) {
	s.stepSettings = stepSettings
}

// GenerateDescriptionRequest contains the request data for generating a PR description
type GenerateDescriptionRequest struct {
	SourceBranch      string
	TargetBranch      string
	Files             []domain.FileChange
	AdditionalContext []domain.ContextItem
	Prompt            string
}

// GenerateDescriptionResponse contains the generated PR description
type GenerateDescriptionResponse struct {
	Description string
	TokensUsed  int
	Model       string
}

// GenerateDescription generates a PR description using a real geppetto engine.
// The engine is created from the configured StepSettings; no parsing happens here.
func (s *Service) GenerateDescription(ctx context.Context, req GenerateDescriptionRequest) (*GenerateDescriptionResponse, error) {
	if s.stepSettings == nil {
		return nil, errors.New("no AI StepSettings configured (configure provider/model flags higher up)")
	}

	systemPrompt, userPrompt, err := compilePrompt(req)
	if err != nil {
		return nil, err
	}

	// Use Turns directly (no conversation.Manager)
	seed := turns.NewTurnBuilder().
		WithSystemPrompt(systemPrompt).
		WithUserPrompt(userPrompt).
		Build()

	eng, err := factory.NewEngineFromStepSettings(s.stepSettings)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create engine from step settings")
	}

	updatedTurn, err := eng.RunInference(ctx, seed)
	if err != nil {
		return nil, errors.Wrap(err, "inference failed")
	}

	description := extractLastAssistantText(updatedTurn)
	if strings.TrimSpace(description) == "" {
		// Preserve a minimal signal for callers/debugging
		description = "<no assistant text produced>"
	}

	// Best-effort token usage (provider usage might be in Turn.Metadata; we can wire later).
	tokensUsed := tokens.Count(description)

	model := ""
	if updatedTurn != nil && updatedTurn.Metadata != nil {
		if v, ok := updatedTurn.Metadata[turns.TurnMetaKeyModel]; ok {
			if s, ok := v.(string); ok {
				model = s
			}
		}
	}

	return &GenerateDescriptionResponse{
		Description: description,
		TokensUsed:  tokensUsed,
		Model:       model,
	}, nil
}

// ValidateRequest validates a generate description request
func (s *Service) ValidateRequest(req GenerateDescriptionRequest) error {
	if req.SourceBranch == "" {
		return fmt.Errorf("source branch is required")
	}
	if req.TargetBranch == "" {
		return fmt.Errorf("target branch is required")
	}
	if len(req.Files) == 0 {
		return fmt.Errorf("no files to generate description from")
	}

	// Check if at least one file is included
	hasIncluded := false
	for _, file := range req.Files {
		if file.Included {
			hasIncluded = true
			break
		}
	}
	if !hasIncluded {
		return fmt.Errorf("at least one file must be included")
	}

	return nil
}

func buildUserContext(req GenerateDescriptionRequest) string {
	var b strings.Builder

	b.WriteString("# Prescribe generation context\n\n")
	b.WriteString("## Branches\n\n")
	b.WriteString(fmt.Sprintf("- Source: %s\n", req.SourceBranch))
	b.WriteString(fmt.Sprintf("- Target: %s\n\n", req.TargetBranch))

	b.WriteString(fmt.Sprintf("## Included files (%d)\n\n", len(req.Files)))
	for _, f := range req.Files {
		b.WriteString(fmt.Sprintf("### %s\n\n", f.Path))

		switch f.Type {
		case domain.FileTypeFull:
			content := f.FullAfter
			if content == "" {
				content = f.FullBefore
			}
			if content == "" {
				content = f.Diff
			}
			b.WriteString("```text\n")
			b.WriteString(strings.TrimRight(content, "\n"))
			b.WriteString("\n```\n\n")
		default:
			diff := f.Diff
			b.WriteString("```diff\n")
			b.WriteString(strings.TrimRight(diff, "\n"))
			b.WriteString("\n```\n\n")
		}
	}

	if len(req.AdditionalContext) > 0 {
		b.WriteString(fmt.Sprintf("## Additional context (%d)\n\n", len(req.AdditionalContext)))
		for _, ctx := range req.AdditionalContext {
			switch ctx.Type {
			case domain.ContextTypeNote:
				b.WriteString("- ")
				b.WriteString(strings.TrimSpace(ctx.Content))
				b.WriteString("\n")
			case domain.ContextTypeFile:
				label := ctx.Path
				if label == "" {
					label = "file"
				}
				b.WriteString(fmt.Sprintf("### %s\n\n", label))
				b.WriteString("```text\n")
				b.WriteString(strings.TrimRight(ctx.Content, "\n"))
				b.WriteString("\n```\n\n")
			default:
				b.WriteString("- ")
				b.WriteString(strings.TrimSpace(ctx.Content))
				b.WriteString("\n")
			}
		}
		b.WriteString("\n")
	}

	return b.String()
}

func extractLastAssistantText(t *turns.Turn) string {
	if t == nil {
		return ""
	}
	for i := len(t.Blocks) - 1; i >= 0; i-- {
		b := t.Blocks[i]
		if b.Kind != turns.BlockKindLLMText {
			continue
		}
		if b.Role != turns.RoleAssistant {
			continue
		}
		if txt, ok := b.Payload[turns.PayloadKeyText].(string); ok {
			return txt
		}
	}
	return ""
}
