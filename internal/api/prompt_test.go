package api

import (
	"strings"
	"testing"

	"github.com/go-go-golems/prescribe/internal/domain"
	"github.com/go-go-golems/prescribe/internal/prompts"
)

func TestCompilePrompt_pinocchioStyleCombinedPrompt_rendersTemplates(t *testing.T) {
	req := GenerateDescriptionRequest{
		SourceBranch: "feature",
		TargetBranch: "main",
		Prompt:       prompts.DefaultPrompt(),
		Files: []domain.FileChange{
			{
				Path:     "a.go",
				Type:     domain.FileTypeDiff,
				Included: true,
				Diff:     "diff --git a/a.go b/a.go\n+added\n",
			},
		},
		AdditionalContext: []domain.ContextItem{
			{Type: domain.ContextTypeNote, Content: "note-1"},
			{Type: domain.ContextTypeFile, Path: "README.md", Content: "# Hello\n"},
		},
	}

	sys, user, err := compilePrompt(req)
	if err != nil {
		t.Fatalf("compilePrompt error: %v", err)
	}
	if strings.Contains(sys, "{{") || strings.Contains(user, "{{") {
		t.Fatalf("expected templates to be rendered (no raw {{...}} left), got:\nSYS:\n%s\n\nUSER:\n%s", sys, user)
	}
	if !strings.Contains(sys, "experienced software engineer") {
		t.Fatalf("expected system prompt to contain base role text, got:\n%s", sys)
	}
	if !strings.Contains(user, "diff --git a/a.go b/a.go") {
		t.Fatalf("expected rendered prompt to contain diff, got:\n%s", user)
	}
	if !strings.Contains(user, "note-1") {
		t.Fatalf("expected rendered prompt to contain description/note mapping, got:\n%s", user)
	}
	if !strings.Contains(user, "README.md") || !strings.Contains(user, "# Hello") {
		t.Fatalf("expected rendered prompt to contain context file mapping, got:\n%s", user)
	}
}

func TestCompilePrompt_plainPrompt_fallsBackToUserContext(t *testing.T) {
	req := GenerateDescriptionRequest{
		SourceBranch: "feature",
		TargetBranch: "main",
		Prompt:       "system only",
		Files: []domain.FileChange{
			{
				Path:     "a.go",
				Type:     domain.FileTypeDiff,
				Included: true,
				Diff:     "diff --git a/a.go b/a.go\n+added\n",
			},
		},
	}

	sys, user, err := compilePrompt(req)
	if err != nil {
		t.Fatalf("compilePrompt error: %v", err)
	}
	if sys != "system only" {
		t.Fatalf("expected system prompt to match input, got %q", sys)
	}
	if !strings.Contains(user, "# Prescribe generation context") {
		t.Fatalf("expected fallback user context output, got:\n%s", user)
	}
}


