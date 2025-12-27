package export

import (
	"fmt"
	"strings"

	"github.com/go-go-golems/prescribe/internal/api"
	"github.com/go-go-golems/prescribe/internal/domain"
)

// BuildGenerationContextText formats the canonical generation inputs into a clipboard-friendly text blob.
//
// NOTE: This is intentionally "human readable first" (markdown-ish) so it can be pasted into PRs, chat, or LLM UIs.
func BuildGenerationContextText(req api.GenerateDescriptionRequest) string {
	var b strings.Builder

	b.WriteString("# Prescribe generation context\n\n")
	b.WriteString("## Branches\n\n")
	b.WriteString(fmt.Sprintf("- Source: %s\n", req.SourceBranch))
	b.WriteString(fmt.Sprintf("- Target: %s\n\n", req.TargetBranch))

	b.WriteString("## Prompt\n\n")
	b.WriteString("```text\n")
	b.WriteString(strings.TrimSpace(req.Prompt))
	b.WriteString("\n```\n\n")

	b.WriteString(fmt.Sprintf("## Included files (%d)\n\n", len(req.Files)))
	for _, f := range req.Files {
		b.WriteString(fmt.Sprintf("### %s\n\n", f.Path))

		switch f.Type {
		case domain.FileTypeFull:
			// Prefer FullAfter if available; fall back to Diff if the file doesn't have full content populated.
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
