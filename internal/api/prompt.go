package api

import (
	"fmt"
	"strings"

	"github.com/go-go-golems/glazed/pkg/helpers/templating"
	"github.com/go-go-golems/prescribe/internal/domain"
	"github.com/pkg/errors"
)

type templateFile struct {
	Path    string
	Content string
}

func renderTemplateString(name, text string, vars map[string]any) (string, error) {
	if strings.TrimSpace(text) == "" {
		return text, nil
	}
	tpl, err := templating.CreateTemplate(name).Parse(text)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse template")
	}
	var b strings.Builder
	if err := tpl.Execute(&b, vars); err != nil {
		return "", errors.Wrap(err, "failed to execute template")
	}
	return b.String(), nil
}

func buildTemplateVars(req GenerateDescriptionRequest) map[string]any {
	// Map prescribe state into pinocchio-style variables used by the embedded prompt template.
	//
	// We intentionally keep this dependency-light: use a map + small structs so templates can access
	// `.diff`, `.code` (with `.Path`/`.Content`), `.context` (with `.Path`/`.Content`), etc.

	var diffParts []string
	codeFiles := make([]templateFile, 0)
	contextFiles := make([]templateFile, 0)
	var noteParts []string

	for _, f := range req.Files {
		switch f.Type {
		case domain.FileTypeFull:
			content := f.FullAfter
			if content == "" {
				content = f.FullBefore
			}
			if content == "" {
				content = f.Diff
			}
			if strings.TrimSpace(content) != "" {
				codeFiles = append(codeFiles, templateFile{Path: f.Path, Content: strings.TrimRight(content, "\n")})
			}
		default:
			if strings.TrimSpace(f.Diff) != "" {
				// Keep diffs well-delimited per file to avoid “smashed together” ambiguity.
				// We mirror the XML-ish boundary style used in the export-context separator approach.
				diffParts = append(diffParts, fmt.Sprintf(
					"<file name=\"%s\" type=\"diff\">\n<diff>\n%s\n</diff>\n</file>",
					xmlEscapeAttr(f.Path),
					strings.TrimRight(f.Diff, "\n"),
				))
			}
		}
	}

	for _, c := range req.AdditionalContext {
		switch c.Type {
		case domain.ContextTypeFile:
			if strings.TrimSpace(c.Content) != "" {
				contextFiles = append(contextFiles, templateFile{Path: c.Path, Content: strings.TrimRight(c.Content, "\n")})
			}
		case domain.ContextTypeNote:
			if strings.TrimSpace(c.Content) != "" {
				noteParts = append(noteParts, strings.TrimSpace(c.Content))
			}
		default:
			if strings.TrimSpace(c.Content) != "" {
				noteParts = append(noteParts, strings.TrimSpace(c.Content))
			}
		}
	}

	description := strings.TrimSpace(strings.Join(noteParts, "\n"))
	diff := strings.TrimSpace(strings.Join(diffParts, "\n\n"))

	return map[string]any{
		// Pinocchio-style prompt variables (subset)
		"diff":             diff,
		"code":             codeFiles,
		"context":          contextFiles,
		"description":      description,
		"title":            "",
		"issue":            "",
		"commits":          "",
		"additional_system": "",
		"additional":        []string{},

		// Defaults from the embedded prompt pack (keep behavior stable)
		"bracket":       true,
		"without_files": true,
		"concise":       false,
		"use_bullets":   false,
		"use_keywords":  false,
	}
}

// CompilePrompt is an exported wrapper to compile the prompt into the exact (system,user) strings
// that will be used to seed a Turn. This is useful for CLI export-only workflows.
func CompilePrompt(req GenerateDescriptionRequest) (systemPrompt string, userPrompt string, err error) {
	return compilePrompt(req)
}

func splitCombinedPinocchioPrompt(prompt string) (systemTemplate string, userTemplate string, ok bool) {
	// prescribe stores a single "combined" string. Our default prompt is derived from pinocchio's
	// create-pull-request prompt pack, which defines a "context" template. We treat the text before
	// that definition as system prompt, and the remainder as the user prompt template to render.
	p := prompt
	idx := strings.Index(p, "{{ define \"context\"")
	if idx < 0 {
		idx = strings.Index(p, "{{define \"context\"")
	}
	if idx < 0 {
		return "", "", false
	}
	return strings.TrimSpace(p[:idx]), strings.TrimSpace(p[idx:]), true
}

func compilePrompt(req GenerateDescriptionRequest) (systemPrompt string, userPrompt string, err error) {
	combined := strings.TrimSpace(req.Prompt)
	if combined == "" {
		return "", buildUserContext(req), nil
	}

	sysTmpl, userTmpl, ok := splitCombinedPinocchioPrompt(combined)
	if !ok {
		// Legacy behavior: treat prompt as an already-rendered system prompt and send the context as the user message.
		return combined, buildUserContext(req), nil
	}

	vars := buildTemplateVars(req)
	sys, err := renderTemplateString("system-prompt", sysTmpl, vars)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to render system prompt template")
	}
	user, err := renderTemplateString("prompt", userTmpl, vars)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to render user prompt template")
	}
	return strings.TrimSpace(sys), strings.TrimSpace(user), nil
}

func xmlEscapeAttr(s string) string {
	// Minimal XML escaping for attribute safety.
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}


