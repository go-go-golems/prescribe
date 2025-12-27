package controller

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-go-golems/prescribe/internal/domain"
	"gopkg.in/yaml.v3"
)

type repoConfigYAML struct {
	Defaults repoDefaultsYAML `yaml:"defaults,omitempty"`
}

type repoDefaultsYAML struct {
	FilterPresets []string `yaml:"filter_presets,omitempty"`
}

// ApplyDefaultFilterPresetsFromRepoConfig loads <repo>/.pr-builder/config.yaml and applies any configured
// default filter presets to the current controller state.
//
// This is intended for "new session" behavior (i.e. when session.yaml is missing).
// It does not save a session automatically.
func (c *Controller) ApplyDefaultFilterPresetsFromRepoConfig() (int, error) {
	cfgPath := filepath.Join(c.repoPath, ".pr-builder", "config.yaml")

	b, err := os.ReadFile(cfgPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, nil
		}
		return 0, fmt.Errorf("read repo config: %w", err)
	}

	var cfg repoConfigYAML
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return 0, fmt.Errorf("unmarshal repo config: %w", err)
	}

	if len(cfg.Defaults.FilterPresets) == 0 {
		return 0, nil
	}

	applied := 0
	for _, presetID := range cfg.Defaults.FilterPresets {
		preset, err := c.findFilterPresetByID(presetID)
		if err != nil {
			return applied, err
		}

		c.AddFilter(domain.Filter{
			Name:        preset.Name,
			Description: preset.Description,
			Rules:       preset.Rules,
		})
		applied++
	}

	return applied, nil
}

func (c *Controller) findFilterPresetByID(presetID string) (domain.FilterPreset, error) {
	projectPresets, err := c.LoadProjectFilterPresets()
	if err == nil {
		for _, p := range projectPresets {
			if p.ID == presetID {
				return p, nil
			}
		}
	}

	globalPresets, err := c.LoadGlobalFilterPresets()
	if err == nil {
		for _, p := range globalPresets {
			if p.ID == presetID {
				return p, nil
			}
		}
	}

	return domain.FilterPreset{}, fmt.Errorf("filter preset not found: %s", presetID)
}
