package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-go-golems/prescribe/internal/controller"
	"github.com/go-go-golems/prescribe/internal/tui/events"
)

func TestBootCmd_MissingSession_AppliesRepoDefaultFilterPresets(t *testing.T) {
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}

	// Preset: <repo>/.pr-builder/filters/exclude_tests.yaml
	if err := os.MkdirAll(filepath.Join(repo, ".pr-builder", "filters"), 0755); err != nil {
		t.Fatalf("mkdir preset dir: %v", err)
	}
	preset := []byte(`name: Exclude Tests
description: Exclude test files
rules:
  - type: exclude
    pattern: "**/*test*"
`)
	if err := os.WriteFile(filepath.Join(repo, ".pr-builder", "filters", "exclude_tests.yaml"), preset, 0644); err != nil {
		t.Fatalf("write preset: %v", err)
	}

	// Repo config defaults: <repo>/.pr-builder/config.yaml
	cfg := []byte(`defaults:
  filter_presets:
    - exclude_tests.yaml
`)
	if err := os.WriteFile(filepath.Join(repo, ".pr-builder", "config.yaml"), cfg, 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	ctrl, err := controller.NewController(repo)
	if err != nil {
		t.Fatalf("NewController: %v", err)
	}

	msg := bootCmd(ctrl)()
	applied, ok := msg.(events.DefaultFiltersAppliedMsg)
	if !ok {
		t.Fatalf("expected DefaultFiltersAppliedMsg, got %T", msg)
	}
	if applied.Count != 1 {
		t.Fatalf("expected Count=1, got %d", applied.Count)
	}

	filters := ctrl.GetFilters()
	if len(filters) != 1 {
		t.Fatalf("expected 1 active filter, got %d", len(filters))
	}
	if filters[0].Name != "Exclude Tests" {
		t.Fatalf("unexpected filter name: %q", filters[0].Name)
	}
}
