package api

import (
	"fmt"
	"strings"

	"github.com/go-go-golems/geppetto/pkg/steps/ai/settings"
	"github.com/go-go-golems/geppetto/pkg/turns"
	"github.com/go-go-golems/prescribe/internal/domain"
)

func parseAndValidateGeneratedPRData(assistantText string) (*domain.GeneratedPRData, string) {
	p, err := ParseGeneratedPRDataFromAssistantText(assistantText)
	if err != nil {
		return p, err.Error()
	}
	if p == nil {
		return nil, "failed to parse PR YAML (no data)"
	}
	if !isGeneratedPRDataValid(p) {
		return p, "parsed PR YAML is missing required fields (title/body)"
	}
	return p, ""
}

func getTurnStopReason(t *turns.Turn) string {
	if t == nil || t.Metadata == nil {
		return ""
	}
	v, ok := t.Metadata[turns.TurnMetaKeyStopReason]
	if !ok || v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

func isLikelyMaxTokensStopReason(stopReason string) bool {
	s := strings.ToLower(strings.TrimSpace(stopReason))
	if s == "" {
		return false
	}
	// Provider variants:
	// - OpenAI: "max_tokens"
	// - Claude: "max_tokens"
	// - Gemini (SDK fmt): "FinishReasonMaxTokens"
	return strings.Contains(s, "maxtokens") ||
		strings.Contains(s, "max_tokens") ||
		strings.Contains(s, "max tokens")
}

func computeRetryMaxResponseTokens(ss *settings.StepSettings) int {
	if ss == nil || ss.Chat == nil {
		return 0
	}
	if ss.Chat.MaxResponseTokens == nil {
		return 2048
	}
	cur := *ss.Chat.MaxResponseTokens
	if cur <= 0 {
		return 2048
	}
	// If it was very small, jump to a reasonable value for YAML outputs.
	if cur < 256 {
		return 2048
	}
	// Otherwise, scale up but keep a cap.
	next := cur * 4
	if next > 8192 {
		next = 8192
	}
	return next
}

func yamlRepairRetrySuffix() string {
	// Keep this short (we don't want to bloat the already-large prompt).
	// The base prompt already contains the schema; we just restate the hard constraints.
	return strings.TrimSpace(`
IMPORTANT: Your previous response was incomplete/invalid. Re-output the complete YAML now.

Rules:
- Output YAML only (no markdown, no code fences, no prose).
- Include all required keys exactly: title, body, changelog, release_notes (with title/body).
- Ensure body, changelog, and release_notes.body are non-empty.
`)
}
