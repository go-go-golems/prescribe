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

func TestCompilePrompt_pinocchioStyleCombinedPrompt_omitsEmptyDescriptionBlock(t *testing.T) {
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
			{Type: domain.ContextTypeFile, Path: "README.md", Content: "# Hello\n"},
		},
	}

	_, user, err := compilePrompt(req)
	if err != nil {
		t.Fatalf("compilePrompt error: %v", err)
	}

	// When no note-based description was provided, the prompt must not emit a dangling marker.
	if strings.Contains(user, "The description of the pull request is:") {
		t.Fatalf("expected prompt to omit legacy description line, got:\n%s", user)
	}
	if strings.Contains(user, "Pull request description / notes provided by the user:") {
		t.Fatalf("expected prompt to omit description block header when empty, got:\n%s", user)
	}
	if strings.Contains(user, "BEGIN DESCRIPTION") || strings.Contains(user, "END DESCRIPTION") {
		t.Fatalf("expected prompt to omit description block when empty, got:\n%s", user)
	}
	if strings.Contains(user, "is: .") {
		t.Fatalf("expected prompt to avoid empty description marker, got:\n%s", user)
	}
}

func TestCompilePrompt_pinocchioStyleCombinedPrompt_rendersTitleAndDescriptionVars(t *testing.T) {
	req := GenerateDescriptionRequest{
		SourceBranch: "feature",
		TargetBranch: "main",
		Title:        "Provided title",
		Description:  "Provided description",
		Prompt:       prompts.DefaultPrompt(),
		Files: []domain.FileChange{
			{
				Path:     "a.go",
				Type:     domain.FileTypeDiff,
				Included: true,
				Diff:     "diff --git a/a.go b/a.go\n+added\n",
			},
		},
	}

	_, user, err := compilePrompt(req)
	if err != nil {
		t.Fatalf("compilePrompt error: %v", err)
	}
	if !strings.Contains(user, "Provided title") {
		t.Fatalf("expected rendered prompt to contain provided title, got:\n%s", user)
	}
	if !strings.Contains(user, "Provided description") {
		t.Fatalf("expected rendered prompt to contain provided description, got:\n%s", user)
	}
}
