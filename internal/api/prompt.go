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
	var commitsParts []string

	for _, f := range req.Files {
		switch f.Type {
		case domain.FileTypeFull:
			if f.Version == domain.FileVersionBoth {
				// Include both versions when FileVersionBoth is set
				if strings.TrimSpace(f.FullBefore) != "" {
					codeFiles = append(codeFiles, templateFile{
						Path:    f.Path + ":before",
						Content: strings.TrimRight(f.FullBefore, "\n"),
					})
				}
				if strings.TrimSpace(f.FullAfter) != "" {
					codeFiles = append(codeFiles, templateFile{
						Path:    f.Path + ":after",
						Content: strings.TrimRight(f.FullAfter, "\n"),
					})
				}
			} else {
				// Single version logic (existing behavior)
				content := f.FullAfter
				if f.Version == domain.FileVersionBefore || content == "" {
					content = f.FullBefore
				}
				if content == "" {
					content = f.Diff
				}
				if strings.TrimSpace(content) != "" {
					codeFiles = append(codeFiles, templateFile{
						Path:    f.Path,
						Content: strings.TrimRight(content, "\n"),
					})
				}
			}
		case domain.FileTypeDiff:
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
		case domain.ContextTypeGitHistory:
			if strings.TrimSpace(c.Content) != "" {
				commitsParts = append(commitsParts, strings.TrimRight(c.Content, "\n"))
			}
		default:
			if strings.TrimSpace(c.Content) != "" {
				noteParts = append(noteParts, strings.TrimSpace(c.Content))
			}
		}
	}

	notes := strings.TrimSpace(strings.Join(noteParts, "\n"))
	descParts := make([]string, 0, 2)
	if strings.TrimSpace(req.Description) != "" {
		descParts = append(descParts, strings.TrimSpace(req.Description))
	}
	if notes != "" {
		descParts = append(descParts, notes)
	}
	description := strings.TrimSpace(strings.Join(descParts, "\n\n"))
	diff := strings.TrimSpace(strings.Join(diffParts, "\n\n"))
	title := strings.TrimSpace(req.Title)

	return map[string]any{
		// Pinocchio-style prompt variables (subset)
		"diff":              diff,
		"code":              codeFiles,
		"context":           contextFiles,
		"description":       description,
		"title":             title,
		"issue":             "",
		"commits":           strings.TrimSpace(strings.Join(commitsParts, "\n\n")),
		"additional_system": "",
		"additional":        []string{},

		// Defaults from the embedded prompt pack (keep behavior stable)
		"without_files": true,
		"concise":       false,
		"use_bullets":   false,
		"use_keywords":  false,
	}
}

// CompilePrompt is an exported wrapper to compile the prompt into the exact (system,user) strings
// that will be used to seed a Turn. This is useful for CLI export-only workflows.
func CompilePrompt(req GenerateDescriptionRequest) (string, string, error) {
	return compilePrompt(req)
}

func splitCombinedPinocchioPrompt(prompt string) (string, string, bool) {
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

func compilePrompt(req GenerateDescriptionRequest) (string, string, error) {
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
