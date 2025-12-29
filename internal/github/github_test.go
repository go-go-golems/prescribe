package github

import (
	"reflect"
	"strings"
	"testing"
)

func TestBuildGhCreatePRArgs(t *testing.T) {
	t.Run("missing title", func(t *testing.T) {
		_, err := BuildGhCreatePRArgs(CreatePROptions{Body: "b"})
		if err == nil || !strings.Contains(err.Error(), "missing PR title") {
			t.Fatalf("expected missing title error, got %v", err)
		}
	})

	t.Run("missing body", func(t *testing.T) {
		_, err := BuildGhCreatePRArgs(CreatePROptions{Title: "t"})
		if err == nil || !strings.Contains(err.Error(), "missing PR body") {
			t.Fatalf("expected missing body error, got %v", err)
		}
	})

	t.Run("base + draft", func(t *testing.T) {
		args, err := BuildGhCreatePRArgs(CreatePROptions{
			Title: "t",
			Body:  "b",
			Base:  "main",
			Draft: true,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := []string{"pr", "create", "--title", "t", "--body", "b", "--base", "main", "--draft"}
		if !reflect.DeepEqual(args, want) {
			t.Fatalf("args mismatch:\n got: %#v\nwant: %#v", args, want)
		}
	})
}

func TestRedactGhArgs(t *testing.T) {
	in := []string{"pr", "create", "--title", "t", "--body", "secret", "--base", "main"}
	out := RedactGhArgs(in)
	if strings.Join(out, " ") != "pr create --title t --body <omitted> --base main" {
		t.Fatalf("unexpected redaction: %v", out)
	}
	// Ensure original slice is not modified.
	if in[5] != "secret" {
		t.Fatalf("input modified unexpectedly: %v", in)
	}
}
