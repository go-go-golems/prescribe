package controller

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/prescribe/internal/domain"
	"gopkg.in/yaml.v3"
)

type filterPresetYAML struct {
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description,omitempty"`
	Rules       []filterPresetRuleYAML `yaml:"rules"`
}

type filterPresetRuleYAML struct {
	Type    string `yaml:"type"`    // "include" or "exclude"
	Pattern string `yaml:"pattern"` // glob pattern
}

// LoadProjectFilterPresets loads filter presets from the project directory: <repo>/.pr-builder/filters
func (c *Controller) LoadProjectFilterPresets() ([]domain.FilterPreset, error) {
	presetDir := filepath.Join(c.repoPath, ".pr-builder", "filters")
	return c.loadFilterPresetsFromDir(presetDir, domain.PresetLocationProject)
}

// LoadGlobalFilterPresets loads filter presets from the global directory: ~/.pr-builder/filters
func (c *Controller) LoadGlobalFilterPresets() ([]domain.FilterPreset, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	presetDir := filepath.Join(homeDir, ".pr-builder", "filters")
	return c.loadFilterPresetsFromDir(presetDir, domain.PresetLocationGlobal)
}

func (c *Controller) loadFilterPresetsFromDir(dir string, location domain.PresetLocation) ([]domain.FilterPreset, error) {
	presets := make([]domain.FilterPreset, 0)

	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return presets, nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".yaml" && filepath.Ext(entry.Name()) != ".yml" {
			continue
		}

		filePath := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		var preset filterPresetYAML
		if err := yaml.Unmarshal(data, &preset); err != nil {
			continue
		}

		rules := make([]domain.FilterRule, 0, len(preset.Rules))
		for i, r := range preset.Rules {
			rules = append(rules, domain.FilterRule{
				Type:    domain.FilterType(r.Type),
				Pattern: r.Pattern,
				Order:   i,
			})
		}

		presets = append(presets, domain.FilterPreset{
			ID:          entry.Name(),
			Name:        preset.Name,
			Description: preset.Description,
			Rules:       rules,
			Location:    location,
		})
	}

	return presets, nil
}

// SaveFilterPreset saves a filter preset under either the project or global preset directory.
func (c *Controller) SaveFilterPreset(name, description string, rules []domain.FilterRule, location domain.PresetLocation) error {
	var dir string
	if location == domain.PresetLocationProject {
		dir = filepath.Join(c.repoPath, ".pr-builder", "filters")
	} else if location == domain.PresetLocationGlobal {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		dir = filepath.Join(homeDir, ".pr-builder", "filters")
	} else {
		return fmt.Errorf("unsupported preset location: %s", location)
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	filename := strings.ToLower(strings.ReplaceAll(name, " ", "_")) + ".yaml"
	filePath := filepath.Join(dir, filename)

	yamlRules := make([]filterPresetRuleYAML, 0, len(rules))
	for _, r := range rules {
		yamlRules = append(yamlRules, filterPresetRuleYAML{
			Type:    string(r.Type),
			Pattern: r.Pattern,
		})
	}

	preset := filterPresetYAML{
		Name:        name,
		Description: description,
		Rules:       yamlRules,
	}

	data, err := yaml.Marshal(preset)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}
