package export

import (
	"strings"
	"testing"

	"github.com/go-go-golems/prescribe/internal/api"
	"github.com/go-go-golems/prescribe/internal/domain"
)

func TestBuildGenerationContext_markdown_basic(t *testing.T) {
	req := api.GenerateDescriptionRequest{
		SourceBranch: "feature",
		TargetBranch: "main",
		SourceCommit: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		TargetCommit: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		Prompt:       "please write a PR description",
		Files: []domain.FileChange{
			{
				Path:     "a.go",
				Type:     domain.FileTypeDiff,
				Included: true,
				Diff:     "diff --git a/a.go b/a.go\n+added\n",
			},
		},
		AdditionalContext: []domain.ContextItem{
			{Type: domain.ContextTypeNote, Content: "remember to mention perf"},
		},
	}

	out := BuildGenerationContext(req, SeparatorMarkdown)
	for _, want := range []string{
		"Source: feature",
		"Target: main",
		"Source commit: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"Target commit: bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		"please write a PR description",
		"### a.go",
		"diff --git a/a.go b/a.go",
		"remember to mention perf",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
}
