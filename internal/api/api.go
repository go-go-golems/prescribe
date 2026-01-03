package api

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"strings"

	"github.com/go-go-golems/geppetto/pkg/events"
	geppettoengine "github.com/go-go-golems/geppetto/pkg/inference/engine"
	"github.com/go-go-golems/geppetto/pkg/inference/engine/factory"
	"github.com/go-go-golems/geppetto/pkg/inference/middleware"
	"github.com/go-go-golems/geppetto/pkg/steps/ai/settings"
	"github.com/go-go-golems/geppetto/pkg/turns"
	"github.com/go-go-golems/prescribe/internal/domain"
	"github.com/go-go-golems/prescribe/internal/tokens"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
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
	SourceCommit      string
	TargetCommit      string
	Title             string
	Description       string
	Files             []domain.FileChange
	AdditionalContext []domain.ContextItem
	Prompt            string
}

// GenerateDescriptionResponse contains the generated PR description
type GenerateDescriptionResponse struct {
	Description string
	Parsed      *domain.GeneratedPRData
	ParseError  string
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

	debugLogTurnSeed(req, seed, systemPrompt, userPrompt)

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

	debugLogAssistantText(updatedTurn, description)

	var parsed *domain.GeneratedPRData
	parseErrStr := ""
	if p, err := ParseGeneratedPRDataFromAssistantText(description); err == nil {
		parsed = p
	} else {
		parsed = p // best-effort: keep partial struct if available
		parseErrStr = err.Error()
	}
	if parsed != nil && !isGeneratedPRDataValid(parsed) {
		if parseErrStr == "" {
			parseErrStr = "parsed PR YAML is missing required fields (title/body)"
		}
	}

	// Best-effort retry: if the model likely hit a max token limit and produced invalid/partial YAML,
	// rerun once with a higher output token budget and a short corrective instruction.
	if parseErrStr != "" && isLikelyMaxTokensStopReason(getTurnStopReason(updatedTurn)) {
		retryMax := computeRetryMaxResponseTokens(s.stepSettings)
		if retryMax > 0 {
			log.Debug().
				Str("stop_reason", getTurnStopReason(updatedTurn)).
				Int("retry_max_response_tokens", retryMax).
				Msg("api: retrying inference due to invalid YAML + max-tokens stop reason")

			retrySettings := s.stepSettings.Clone()
			retrySettings.Chat.MaxResponseTokens = &retryMax

			retrySeed := turns.NewTurnBuilder().
				WithSystemPrompt(systemPrompt).
				WithUserPrompt(userPrompt + "\n\n" + yamlRepairRetrySuffix()).
				Build()

			debugLogTurnSeed(req, retrySeed, systemPrompt, userPrompt+"\n\n"+yamlRepairRetrySuffix())

			retryEng, err := factory.NewEngineFromStepSettings(retrySettings)
			if err != nil {
				log.Debug().Err(err).Msg("api: retry engine creation failed; keeping first attempt")
			} else if retryTurn, err := retryEng.RunInference(ctx, retrySeed); err != nil {
				log.Debug().Err(err).Msg("api: retry inference failed; keeping first attempt")
			} else {
				retryDesc := extractLastAssistantText(retryTurn)
				if strings.TrimSpace(retryDesc) != "" {
					debugLogAssistantText(retryTurn, retryDesc)
					rParsed, rErrStr := parseAndValidateGeneratedPRData(retryDesc)
					if rErrStr == "" {
						updatedTurn = retryTurn
						description = retryDesc
						parsed = rParsed
						parseErrStr = ""
					} else {
						log.Debug().Str("retry_parse_error", rErrStr).Msg("api: retry output still invalid; keeping first attempt")
					}
				}
			}
		}
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
		Parsed:      parsed,
		ParseError:  parseErrStr,
		TokensUsed:  tokensUsed,
		Model:       model,
	}, nil
}

// GenerateDescriptionStreaming runs inference with an attached event sink and prints streaming
// events to the provided writer while still returning the final result.
//
// This is intended for stdio streaming (CLI). TUI streaming can re-use the same plumbing with
// a different router handler.
func (s *Service) GenerateDescriptionStreaming(ctx context.Context, req GenerateDescriptionRequest, w io.Writer) (*GenerateDescriptionResponse, error) {
	if s.stepSettings == nil {
		return nil, errors.New("no AI StepSettings configured (configure provider/model flags higher up)")
	}

	systemPrompt, userPrompt, err := compilePrompt(req)
	if err != nil {
		return nil, err
	}

	seed := turns.NewTurnBuilder().
		WithSystemPrompt(systemPrompt).
		WithUserPrompt(userPrompt).
		Build()

	debugLogTurnSeed(req, seed, systemPrompt, userPrompt)

	router, err := events.NewEventRouter()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create event router")
	}
	defer func() {
		_ = router.Close()
	}()

	router.AddHandler("chat", "chat", events.StepPrinterFunc("", w))

	watermillSink := middleware.NewWatermillSink(router.Publisher, "chat")
	eng, err := factory.NewEngineFromStepSettings(s.stepSettings, geppettoengine.WithSink(watermillSink))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create engine from step settings")
	}

	eg := errgroup.Group{}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	eg.Go(func() error {
		err := router.Run(ctx)
		if errors.Is(err, context.Canceled) {
			return nil
		}
		return err
	})

	var updatedTurn *turns.Turn
	eg.Go(func() error {
		defer cancel()
		<-router.Running()
		t, err := eng.RunInference(ctx, seed)
		if err != nil {
			return errors.Wrap(err, "inference failed")
		}
		updatedTurn = t
		return nil
	})

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	description := extractLastAssistantText(updatedTurn)
	if strings.TrimSpace(description) == "" {
		description = "<no assistant text produced>"
	}

	debugLogAssistantText(updatedTurn, description)

	var parsed *domain.GeneratedPRData
	parseErrStr := ""
	if p, err := ParseGeneratedPRDataFromAssistantText(description); err == nil {
		parsed = p
	} else {
		parsed = p // best-effort: keep partial struct if available
		parseErrStr = err.Error()
	}
	if parsed != nil && !isGeneratedPRDataValid(parsed) {
		if parseErrStr == "" {
			parseErrStr = "parsed PR YAML is missing required fields (title/body)"
		}
	}

	// Same best-effort retry policy as non-streaming. Note: streaming output to `w` will include
	// the first attempt's deltas; we still aim to produce a valid final result for stdout + parsed summary.
	if parseErrStr != "" && isLikelyMaxTokensStopReason(getTurnStopReason(updatedTurn)) {
		retryMax := computeRetryMaxResponseTokens(s.stepSettings)
		if retryMax > 0 {
			log.Debug().
				Str("stop_reason", getTurnStopReason(updatedTurn)).
				Int("retry_max_response_tokens", retryMax).
				Msg("api: retrying inference (streaming) due to invalid YAML + max-tokens stop reason")

			retrySettings := s.stepSettings.Clone()
			retrySettings.Chat.MaxResponseTokens = &retryMax

			retrySeed := turns.NewTurnBuilder().
				WithSystemPrompt(systemPrompt).
				WithUserPrompt(userPrompt + "\n\n" + yamlRepairRetrySuffix()).
				Build()

			debugLogTurnSeed(req, retrySeed, systemPrompt, userPrompt+"\n\n"+yamlRepairRetrySuffix())

			retryEng, err := factory.NewEngineFromStepSettings(retrySettings, geppettoengine.WithSink(watermillSink))
			if err != nil {
				log.Debug().Err(err).Msg("api: retry engine creation failed; keeping first attempt")
			} else if retryTurn, err := retryEng.RunInference(ctx, retrySeed); err != nil {
				log.Debug().Err(err).Msg("api: retry inference failed; keeping first attempt")
			} else {
				retryDesc := extractLastAssistantText(retryTurn)
				if strings.TrimSpace(retryDesc) != "" {
					debugLogAssistantText(retryTurn, retryDesc)
					rParsed, rErrStr := parseAndValidateGeneratedPRData(retryDesc)
					if rErrStr == "" {
						updatedTurn = retryTurn
						description = retryDesc
						parsed = rParsed
						parseErrStr = ""
					} else {
						log.Debug().Str("retry_parse_error", rErrStr).Msg("api: retry output still invalid; keeping first attempt")
					}
				}
			}
		}
	}

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
		Parsed:      parsed,
		ParseError:  parseErrStr,
		TokensUsed:  tokensUsed,
		Model:       model,
	}, nil
}

func debugLogTurnSeed(req GenerateDescriptionRequest, seed *turns.Turn, systemPrompt, userPrompt string) {
	sysSum := summarizeForDebug(systemPrompt, 4000)
	userSum := summarizeForDebug(userPrompt, 4000)
	sysHash := fmt.Sprintf("%x", sha256.Sum256([]byte(systemPrompt)))
	userHash := fmt.Sprintf("%x", sha256.Sum256([]byte(userPrompt)))

	log.Debug().
		Str("source_branch", req.SourceBranch).
		Str("target_branch", req.TargetBranch).
		Str("model_request", "").
		Str("turn_id", func() string {
			if seed == nil {
				return ""
			}
			return seed.ID
		}()).
		Int("turn_blocks", func() int {
			if seed == nil {
				return 0
			}
			return len(seed.Blocks)
		}()).
		Int("system_len", len(systemPrompt)).
		Str("system_sha256", sysHash).
		Str("system_preview", sysSum).
		Int("user_len", len(userPrompt)).
		Str("user_sha256", userHash).
		Str("user_preview", userSum).
		Msg("api: seed turn prepared for inference")
}

func debugLogAssistantText(t *turns.Turn, assistantText string) {
	preview := summarizeForDebug(assistantText, 4000)
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(assistantText)))
	model := ""
	stopReason := ""
	var usage *events.Usage
	if t != nil && t.Metadata != nil {
		if v, ok := t.Metadata[turns.TurnMetaKeyModel]; ok {
			if s, ok := v.(string); ok {
				model = s
			}
		}
		if v, ok := t.Metadata[turns.TurnMetaKeyStopReason]; ok && v != nil {
			if s, ok := v.(string); ok {
				stopReason = s
			} else {
				stopReason = fmt.Sprintf("%v", v)
			}
		}
		if v, ok := t.Metadata[turns.TurnMetaKeyUsage]; ok && v != nil {
			switch u := v.(type) {
			case *events.Usage:
				usage = u
			case events.Usage:
				uu := u
				usage = &uu
			case map[string]any:
				// tolerate map payloads if they came from serialization boundaries
				uu := events.Usage{}
				if x, ok := u["input_tokens"].(int); ok {
					uu.InputTokens = x
				} else if x, ok := u["input_tokens"].(int64); ok {
					uu.InputTokens = int(x)
				}
				if x, ok := u["output_tokens"].(int); ok {
					uu.OutputTokens = x
				} else if x, ok := u["output_tokens"].(int64); ok {
					uu.OutputTokens = int(x)
				}
				if uu.InputTokens != 0 || uu.OutputTokens != 0 {
					usage = &uu
				}
			}
		}
	}
	e := log.Debug().
		Str("model", model).
		Int("assistant_len", len(assistantText)).
		Str("assistant_sha256", hash).
		Str("assistant_preview", preview)
	if strings.TrimSpace(stopReason) != "" {
		e = e.Str("stop_reason", stopReason)
	}
	if usage != nil {
		e = e.Int("input_tokens", usage.InputTokens).Int("output_tokens", usage.OutputTokens)
	}
	e.Msg("api: assistant raw output (last assistant text block)")
}

func summarizeForDebug(s string, maxLen int) string {
	const ellipsis = "\n...\n"
	if maxLen <= 0 {
		return ""
	}
	ss := strings.TrimSpace(s)
	if len(ss) <= maxLen {
		return ss
	}
	// Keep a prefix and suffix to help detect truncation/format issues.
	prefixLen := maxLen * 2 / 3
	suffixLen := maxLen - prefixLen - len(ellipsis)
	if suffixLen < 0 {
		suffixLen = 0
	}
	if prefixLen < 0 {
		prefixLen = 0
	}
	prefix := ss[:prefixLen]
	suffix := ""
	if suffixLen > 0 {
		suffix = ss[len(ss)-suffixLen:]
	}
	return prefix + ellipsis + suffix
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

	if strings.TrimSpace(req.Title) != "" {
		b.WriteString("## Proposed PR title\n\n")
		b.WriteString(strings.TrimSpace(req.Title))
		b.WriteString("\n\n")
	}

	if strings.TrimSpace(req.Description) != "" {
		b.WriteString("## PR description / notes\n\n")
		b.WriteString(strings.TrimSpace(req.Description))
		b.WriteString("\n\n")
	}

	if strings.TrimSpace(req.SourceCommit) != "" || strings.TrimSpace(req.TargetCommit) != "" {
		b.WriteString("## Commit refs\n\n")
		if strings.TrimSpace(req.SourceCommit) != "" {
			b.WriteString(fmt.Sprintf("- Source commit: %s\n", req.SourceCommit))
		}
		if strings.TrimSpace(req.TargetCommit) != "" {
			b.WriteString(fmt.Sprintf("- Target commit: %s\n", req.TargetCommit))
		}
		b.WriteString("\n")
	}

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
		case domain.FileTypeDiff:
			diff := f.Diff
			b.WriteString("```diff\n")
			b.WriteString(strings.TrimRight(diff, "\n"))
			b.WriteString("\n```\n\n")
		}
	}

	gitHistory := ""
	nonHistoryContext := make([]domain.ContextItem, 0, len(req.AdditionalContext))
	for _, ctx := range req.AdditionalContext {
		if ctx.Type == domain.ContextTypeGitHistory && strings.TrimSpace(ctx.Content) != "" {
			if gitHistory == "" {
				gitHistory = strings.TrimRight(ctx.Content, "\n")
			} else {
				gitHistory += "\n\n" + strings.TrimRight(ctx.Content, "\n")
			}
			continue
		}
		nonHistoryContext = append(nonHistoryContext, ctx)
	}

	if strings.TrimSpace(gitHistory) != "" {
		b.WriteString("## Git history\n\n")
		b.WriteString("```text\n")
		b.WriteString(strings.TrimRight(gitHistory, "\n"))
		b.WriteString("\n```\n\n")
	}

	if len(nonHistoryContext) > 0 {
		b.WriteString(fmt.Sprintf("## Additional context (%d)\n\n", len(nonHistoryContext)))
		for _, ctx := range nonHistoryContext {
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
