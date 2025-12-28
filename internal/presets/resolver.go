package presets

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-go-golems/prescribe/internal/domain"
	"gopkg.in/yaml.v3"
)

// ResolvePromptPreset resolves a preset ID by checking builtin, project, and global presets.
// Returns the preset if found, or an error if not found.
func ResolvePromptPreset(presetID, repoPath string) (*domain.PromptPreset, error) {
	// Check built-in presets
	builtins := domain.GetBuiltinPresets()
	for _, preset := range builtins {
		if preset.ID == presetID {
			return &preset, nil
		}
	}

	// Check project presets
	projectPresets, err := LoadProjectPresets(repoPath)
	if err == nil {
		for _, preset := range projectPresets {
			if preset.ID == presetID {
				return &preset, nil
			}
		}
	}

	// Check global presets
	globalPresets, err := LoadGlobalPresets()
	if err == nil {
		for _, preset := range globalPresets {
			if preset.ID == presetID {
				return &preset, nil
			}
		}
	}

	return nil, fmt.Errorf("preset not found: %s", presetID)
}

// LoadProjectPresets loads presets from project directory
func LoadProjectPresets(repoPath string) ([]domain.PromptPreset, error) {
	presetDir := filepath.Join(repoPath, ".pr-builder", "prompts")
	return loadPresetsFromDir(presetDir, domain.PresetLocationProject)
}

// LoadGlobalPresets loads presets from global directory
func LoadGlobalPresets() ([]domain.PromptPreset, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	presetDir := filepath.Join(homeDir, ".pr-builder", "prompts")
	return loadPresetsFromDir(presetDir, domain.PresetLocationGlobal)
}

// loadPresetsFromDir loads presets from a directory
func loadPresetsFromDir(dir string, location domain.PresetLocation) ([]domain.PromptPreset, error) {
	presets := make([]domain.PromptPreset, 0)

	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return presets, nil
	}

	// Read all YAML files
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

		var preset struct {
			Name        string `yaml:"name"`
			Description string `yaml:"description"`
			Template    string `yaml:"template"`
		}

		if err := yaml.Unmarshal(data, &preset); err != nil {
			continue
		}

		presets = append(presets, domain.PromptPreset{
			ID:          entry.Name(),
			Name:        preset.Name,
			Description: preset.Description,
			Template:    preset.Template,
			Location:    location,
		})
	}

	return presets, nil
}
