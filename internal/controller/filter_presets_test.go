package controller

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-go-golems/prescribe/internal/domain"
)

func TestFilterPresets_Project_SaveThenLoad(t *testing.T) {
	tmp := t.TempDir()
	c := &Controller{
		repoPath: tmp,
		data:     domain.NewPRData(),
	}

	rules := []domain.FilterRule{
		{Type: domain.FilterTypeExclude, Pattern: "**/*test*", Order: 0},
		{Type: domain.FilterTypeExclude, Pattern: "**/*spec*", Order: 1},
	}

	if err := c.SaveFilterPreset("Exclude Tests", "Exclude test files", rules, domain.PresetLocationProject); err != nil {
		t.Fatalf("SaveFilterPreset: %v", err)
	}

	presetPath := filepath.Join(tmp, ".pr-builder", "filters", "exclude_tests.yaml")
	b, err := os.ReadFile(presetPath)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", presetPath, err)
	}
	if len(b) == 0 {
		t.Fatalf("expected preset file to be non-empty")
	}

	presets, err := c.LoadProjectFilterPresets()
	if err != nil {
		t.Fatalf("LoadProjectFilterPresets: %v", err)
	}
	if len(presets) != 1 {
		t.Fatalf("expected 1 preset, got %d", len(presets))
	}

	got := presets[0]
	if got.ID != "exclude_tests.yaml" {
		t.Fatalf("expected ID exclude_tests.yaml, got %q", got.ID)
	}
	if got.Location != domain.PresetLocationProject {
		t.Fatalf("expected project location, got %q", got.Location)
	}
	if got.Name != "Exclude Tests" {
		t.Fatalf("expected name %q, got %q", "Exclude Tests", got.Name)
	}
	if got.Description != "Exclude test files" {
		t.Fatalf("expected description %q, got %q", "Exclude test files", got.Description)
	}
	if len(got.Rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(got.Rules))
	}
	if got.Rules[0].Type != domain.FilterTypeExclude || got.Rules[0].Pattern != "**/*test*" {
		t.Fatalf("unexpected rule[0]: %+v", got.Rules[0])
	}
	if got.Rules[1].Type != domain.FilterTypeExclude || got.Rules[1].Pattern != "**/*spec*" {
		t.Fatalf("unexpected rule[1]: %+v", got.Rules[1])
	}
}

func TestFilterPresets_Global_SaveThenLoad(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	c := &Controller{
		repoPath: t.TempDir(),
		data:     domain.NewPRData(),
	}

	rules := []domain.FilterRule{
		{Type: domain.FilterTypeExclude, Pattern: "**/*.md", Order: 0},
	}

	if err := c.SaveFilterPreset("Exclude Docs", "Exclude documentation files", rules, domain.PresetLocationGlobal); err != nil {
		t.Fatalf("SaveFilterPreset: %v", err)
	}

	presets, err := c.LoadGlobalFilterPresets()
	if err != nil {
		t.Fatalf("LoadGlobalFilterPresets: %v", err)
	}
	if len(presets) != 1 {
		t.Fatalf("expected 1 preset, got %d", len(presets))
	}

	got := presets[0]
	if got.Location != domain.PresetLocationGlobal {
		t.Fatalf("expected global location, got %q", got.Location)
	}
	if got.ID != "exclude_docs.yaml" {
		t.Fatalf("expected ID exclude_docs.yaml, got %q", got.ID)
	}

	// Ensure file exists where we expect it.
	presetPath := filepath.Join(home, ".pr-builder", "filters", "exclude_docs.yaml")
	if _, err := os.Stat(presetPath); err != nil {
		t.Fatalf("expected %s to exist: %v", presetPath, err)
	}
}
