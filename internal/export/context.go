package export

import (
	"fmt"
	"strings"

	"github.com/go-go-golems/prescribe/internal/api"
	"github.com/go-go-golems/prescribe/internal/domain"
)

// SeparatorType controls how the context payload is delimited.
// This mirrors catter's delimiter idea (xml/markdown/simple/begin-end/default).
type SeparatorType string

const (
	SeparatorXML      SeparatorType = "xml"
	SeparatorMarkdown SeparatorType = "markdown"
	SeparatorSimple   SeparatorType = "simple"
	SeparatorBeginEnd SeparatorType = "begin-end"
	SeparatorDefault  SeparatorType = "default"
)

// BuildGenerationContext formats the canonical generation inputs into a single text blob.
// Default separator: XML.
func BuildGenerationContext(req api.GenerateDescriptionRequest, sep SeparatorType) string {
	if sep == "" {
		sep = SeparatorXML
	}
	switch sep {
	case SeparatorMarkdown:
		return buildMarkdown(req)
	case SeparatorSimple:
		return buildSimple(req)
	case SeparatorBeginEnd:
		return buildBeginEnd(req)
	case SeparatorDefault:
		return buildDefault(req)
	case SeparatorXML:
		fallthrough
	default:
		return buildXML(req)
	}
}

// BuildRenderedLLMPayload compiles and renders the pinocchio-style prompt (if applicable)
// and returns the exact (system,user) payload that would seed the Turn (no inference).
//
// This is intended for CLI export/debugging: "what we would send to the model".
func BuildRenderedLLMPayload(req api.GenerateDescriptionRequest, sep SeparatorType) (string, error) {
	systemPrompt, userPrompt, err := api.CompilePrompt(req)
	if err != nil {
		return "", err
	}
	if sep == "" {
		sep = SeparatorXML
	}

	switch sep {
	case SeparatorMarkdown:
		var b strings.Builder
		b.WriteString("# Prescribe LLM payload (rendered)\n\n")
		b.WriteString("## System\n\n```text\n")
		b.WriteString(strings.TrimRight(systemPrompt, "\n"))
		b.WriteString("\n```\n\n")
		b.WriteString("## User\n\n```text\n")
		b.WriteString(strings.TrimRight(userPrompt, "\n"))
		b.WriteString("\n```\n")
		return b.String(), nil

	case SeparatorSimple:
		var b strings.Builder
		b.WriteString("--- START SYSTEM PROMPT ---\n")
		b.WriteString(strings.TrimRight(systemPrompt, "\n"))
		b.WriteString("\n--- END SYSTEM PROMPT ---\n\n")
		b.WriteString("--- START USER PROMPT ---\n")
		b.WriteString(strings.TrimRight(userPrompt, "\n"))
		b.WriteString("\n--- END USER PROMPT ---\n")
		return b.String(), nil

	case SeparatorBeginEnd:
		var b strings.Builder
		b.WriteString("--- BEGIN SYSTEM PROMPT ---\n")
		b.WriteString(strings.TrimRight(systemPrompt, "\n"))
		b.WriteString("\n--- END SYSTEM PROMPT ---\n\n")
		b.WriteString("--- BEGIN USER PROMPT ---\n")
		b.WriteString(strings.TrimRight(userPrompt, "\n"))
		b.WriteString("\n--- END USER PROMPT ---\n")
		return b.String(), nil

	case SeparatorDefault:
		var b strings.Builder
		b.WriteString("System:\n")
		b.WriteString(strings.TrimRight(systemPrompt, "\n"))
		b.WriteString("\n\nUser:\n")
		b.WriteString(strings.TrimRight(userPrompt, "\n"))
		b.WriteString("\n")
		return b.String(), nil

	case SeparatorXML:
		fallthrough
	default:
		var b strings.Builder
		b.WriteString("<prescribe>\n")
		b.WriteString("<branches>\n")
		b.WriteString(fmt.Sprintf("<source>%s</source>\n", xmlEscape(req.SourceBranch)))
		b.WriteString(fmt.Sprintf("<target>%s</target>\n", xmlEscape(req.TargetBranch)))
		b.WriteString("</branches>\n")

		b.WriteString("<llm_payload>\n")
		b.WriteString("<system><![CDATA[")
		b.WriteString(xmlCDATA(systemPrompt))
		b.WriteString("]]></system>\n")
		b.WriteString("<user><![CDATA[")
		b.WriteString(xmlCDATA(userPrompt))
		b.WriteString("]]></user>\n")
		b.WriteString("</llm_payload>\n")
		b.WriteString("</prescribe>\n")
		return b.String(), nil
	}
}

func buildXML(req api.GenerateDescriptionRequest) string {
	var b strings.Builder

	b.WriteString("<prescribe>\n")
	b.WriteString("<branches>\n")
	b.WriteString(fmt.Sprintf("<source>%s</source>\n", xmlEscape(req.SourceBranch)))
	b.WriteString(fmt.Sprintf("<target>%s</target>\n", xmlEscape(req.TargetBranch)))
	b.WriteString("</branches>\n")

	b.WriteString("<prompt>\n")
	b.WriteString("<text>\n")
	b.WriteString(xmlEscape(strings.TrimSpace(req.Prompt)))
	b.WriteString("\n</text>\n")
	b.WriteString("</prompt>\n")

	b.WriteString(fmt.Sprintf("<files count=\"%d\">\n", len(req.Files)))
	for _, f := range req.Files {
		b.WriteString(fmt.Sprintf("<file name=\"%s\" type=\"%s\">\n", xmlEscape(f.Path), xmlEscape(string(f.Type))))
		switch f.Type {
		case domain.FileTypeFull:
			content := f.FullAfter
			if content == "" {
				content = f.FullBefore
			}
			if content == "" {
				content = f.Diff
			}
			b.WriteString("<content>\n")
			b.WriteString(xmlEscape(strings.TrimRight(content, "\n")))
			b.WriteString("\n</content>\n")
		default:
			b.WriteString("<diff>\n")
			b.WriteString(xmlEscape(strings.TrimRight(f.Diff, "\n")))
			b.WriteString("\n</diff>\n")
		}
		b.WriteString("</file>\n")
	}
	b.WriteString("</files>\n")

	if len(req.AdditionalContext) > 0 {
		b.WriteString(fmt.Sprintf("<context count=\"%d\">\n", len(req.AdditionalContext)))
		for _, ctx := range req.AdditionalContext {
			switch ctx.Type {
			case domain.ContextTypeNote:
				b.WriteString("<item type=\"note\">\n<text>\n")
				b.WriteString(xmlEscape(strings.TrimSpace(ctx.Content)))
				b.WriteString("\n</text>\n</item>\n")
			case domain.ContextTypeFile:
				b.WriteString(fmt.Sprintf("<item type=\"file\" path=\"%s\">\n<content>\n", xmlEscape(ctx.Path)))
				b.WriteString(xmlEscape(strings.TrimRight(ctx.Content, "\n")))
				b.WriteString("\n</content>\n</item>\n")
			default:
				b.WriteString("<item>\n<text>\n")
				b.WriteString(xmlEscape(strings.TrimSpace(ctx.Content)))
				b.WriteString("\n</text>\n</item>\n")
			}
		}
		b.WriteString("</context>\n")
	}

	b.WriteString("</prescribe>\n")
	return b.String()
}

func xmlCDATA(s string) string {
	// Escape the CDATA terminator by splitting into multiple CDATA sections.
	// This preserves the original bytes while keeping the XML well-formed.
	return strings.ReplaceAll(s, "]]>", "]]]]><![CDATA[>")
}

func buildMarkdown(req api.GenerateDescriptionRequest) string {
	// Markdown is our "clipboard-friendly" representation; keep it stable.
	// (call sites can choose markdown explicitly)
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
			b.WriteString("```diff\n")
			b.WriteString(strings.TrimRight(f.Diff, "\n"))
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

func buildSimple(req api.GenerateDescriptionRequest) string {
	var b strings.Builder
	b.WriteString("--- START PRESCRIBE CONTEXT ---\n")
	b.WriteString(fmt.Sprintf("SOURCE: %s\nTARGET: %s\n\n", req.SourceBranch, req.TargetBranch))
	b.WriteString("--- START PROMPT ---\n")
	b.WriteString(strings.TrimSpace(req.Prompt))
	b.WriteString("\n--- END PROMPT ---\n\n")
	for _, f := range req.Files {
		b.WriteString(fmt.Sprintf("--- START FILE: %s ---\n", f.Path))
		if f.Type == domain.FileTypeFull {
			content := f.FullAfter
			if content == "" {
				content = f.FullBefore
			}
			if content == "" {
				content = f.Diff
			}
			b.WriteString(strings.TrimRight(content, "\n"))
		} else {
			b.WriteString(strings.TrimRight(f.Diff, "\n"))
		}
		b.WriteString(fmt.Sprintf("\n--- END FILE: %s ---\n\n", f.Path))
	}
	b.WriteString("--- END PRESCRIBE CONTEXT ---\n")
	return b.String()
}

func buildBeginEnd(req api.GenerateDescriptionRequest) string {
	var b strings.Builder
	b.WriteString("--- BEGIN PRESCRIBE CONTEXT ---\n")
	b.WriteString(fmt.Sprintf("SOURCE: %s\nTARGET: %s\n\n", req.SourceBranch, req.TargetBranch))
	b.WriteString("--- BEGIN PROMPT ---\n")
	b.WriteString(strings.TrimSpace(req.Prompt))
	b.WriteString("\n--- END PROMPT ---\n\n")
	for _, f := range req.Files {
		b.WriteString(fmt.Sprintf("--- BEGIN FILE: %s ---\n", f.Path))
		if f.Type == domain.FileTypeFull {
			content := f.FullAfter
			if content == "" {
				content = f.FullBefore
			}
			if content == "" {
				content = f.Diff
			}
			b.WriteString(strings.TrimRight(content, "\n"))
		} else {
			b.WriteString(strings.TrimRight(f.Diff, "\n"))
		}
		b.WriteString(fmt.Sprintf("\n--- END FILE: %s ---\n\n", f.Path))
	}
	b.WriteString("--- END PRESCRIBE CONTEXT ---\n")
	return b.String()
}

func buildDefault(req api.GenerateDescriptionRequest) string {
	var b strings.Builder
	b.WriteString("Prescribe context\n\n")
	b.WriteString(fmt.Sprintf("Source: %s\nTarget: %s\n\n", req.SourceBranch, req.TargetBranch))
	b.WriteString("Prompt:\n")
	b.WriteString(strings.TrimSpace(req.Prompt))
	b.WriteString("\n\n")
	for _, f := range req.Files {
		b.WriteString(fmt.Sprintf("File: %s\n", f.Path))
		if f.Type == domain.FileTypeFull {
			content := f.FullAfter
			if content == "" {
				content = f.FullBefore
			}
			if content == "" {
				content = f.Diff
			}
			b.WriteString(strings.TrimRight(content, "\n"))
		} else {
			b.WriteString(strings.TrimRight(f.Diff, "\n"))
		}
		b.WriteString("\n\n")
	}
	return b.String()
}

func xmlEscape(s string) string {
	// Minimal XML escaping for content safety.
	// (We avoid pulling in encoding/xml here to keep control of output.)
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}
