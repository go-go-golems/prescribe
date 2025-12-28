package controller

import (
	"testing"

	"github.com/go-go-golems/prescribe/internal/domain"
)

func TestController_SetFileIncludedByPath(t *testing.T) {
	c := &Controller{
		data: &domain.PRData{
			ChangedFiles: []domain.FileChange{
				{Path: "a.go", Included: false},
				{Path: "b.go", Included: true},
			},
		},
	}

	if err := c.SetFileIncludedByPath("a.go", true); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got := c.data.ChangedFiles[0].Included; got != true {
		t.Fatalf("expected a.go included=true, got %v", got)
	}

	if err := c.SetFileIncludedByPath("b.go", false); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got := c.data.ChangedFiles[1].Included; got != false {
		t.Fatalf("expected b.go included=false, got %v", got)
	}

	if err := c.SetFileIncludedByPath("missing.go", true); err == nil {
		t.Fatalf("expected error for missing file, got nil")
	}
}

func TestController_SetAllVisibleIncluded_respectsFilters(t *testing.T) {
	c := &Controller{
		data: &domain.PRData{
			ChangedFiles: []domain.FileChange{
				{Path: "a.go", Included: false},
				{Path: "b_test.go", Included: false},
			},
			ActiveFilters: []domain.Filter{
				{
					Name: "Exclude Tests",
					Rules: []domain.FilterRule{
						{Type: domain.FilterTypeExclude, Pattern: "**/*test*"},
					},
				},
			},
		},
	}

	n, err := c.SetAllVisibleIncluded(true)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1 visible file, got %d", n)
	}

	if got := c.data.ChangedFiles[0].Included; got != true {
		t.Fatalf("expected a.go included=true, got %v", got)
	}
	if got := c.data.ChangedFiles[1].Included; got != false {
		t.Fatalf("expected b_test.go unchanged (included=false), got %v", got)
	}
}

func TestController_BuildGenerateDescriptionRequest(t *testing.T) {
	c := &Controller{
		data: &domain.PRData{
			SourceBranch: "feature",
			TargetBranch: "main",
			ChangedFiles: []domain.FileChange{
				{Path: "a.go", Included: true},
				{Path: "b.go", Included: false},
			},
			AdditionalContext: []domain.ContextItem{
				{Type: domain.ContextTypeNote, Content: "note", Tokens: 1},
			},
			CurrentPrompt: "prompt",
		},
	}

	req, err := c.BuildGenerateDescriptionRequest()
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if req.SourceBranch != "feature" || req.TargetBranch != "main" {
		t.Fatalf("unexpected branches: %#v", req)
	}
	if req.Prompt != "prompt" {
		t.Fatalf("expected prompt to be set")
	}
	if len(req.AdditionalContext) != 1 {
		t.Fatalf("expected 1 additional context item, got %d", len(req.AdditionalContext))
	}
	if len(req.Files) != 1 || req.Files[0].Path != "a.go" {
		t.Fatalf("expected only included files in request, got %#v", req.Files)
	}
}

func TestController_BuildGenerateDescriptionRequest_requiresIncludedFiles(t *testing.T) {
	c := &Controller{
		data: &domain.PRData{
			SourceBranch: "feature",
			TargetBranch: "main",
			ChangedFiles: []domain.FileChange{
				{Path: "a.go", Included: false},
			},
		},
	}

	_, err := c.BuildGenerateDescriptionRequest()
	if err == nil {
		t.Fatalf("expected error when no files included, got nil")
	}
}
