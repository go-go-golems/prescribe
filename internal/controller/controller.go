package controller

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/prescribe/internal/api"
	"github.com/go-go-golems/prescribe/internal/domain"
	"github.com/go-go-golems/prescribe/internal/git"
	"gopkg.in/yaml.v3"
)

// Controller coordinates between domain data, git, and API services
type Controller struct {
	data       *domain.PRData
	gitService *git.Service
	apiService *api.Service
	repoPath   string
}

// NewController creates a new controller
func NewController(repoPath string) (*Controller, error) {
	gitService, err := git.NewService(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize git service: %w", err)
	}
	
	return &Controller{
		data:       domain.NewPRData(),
		gitService: gitService,
		apiService: api.NewService(),
		repoPath:   repoPath,
	}, nil
}

// Initialize loads the PR data from git
func (c *Controller) Initialize(targetBranch string) error {
	// Get current branch
	sourceBranch, err := c.gitService.GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}
	
	// If no target branch specified, use default
	if targetBranch == "" {
		targetBranch, err = c.gitService.GetDefaultBranch()
		if err != nil {
			return fmt.Errorf("failed to get default branch: %w", err)
		}
	}
	
	c.data.SourceBranch = sourceBranch
	c.data.TargetBranch = targetBranch
	
	// Get changed files
	files, err := c.gitService.GetChangedFiles(sourceBranch, targetBranch)
	if err != nil {
		return fmt.Errorf("failed to get changed files: %w", err)
	}
	
	c.data.ChangedFiles = files
	
	return nil
}

// GetData returns the current domain data
func (c *Controller) GetData() *domain.PRData {
	return c.data
}

// ToggleFileInclusion toggles file inclusion
func (c *Controller) ToggleFileInclusion(index int) error {
	return c.data.ToggleFileInclusion(index)
}

// ReplaceWithFullFile replaces a file's diff with full content
func (c *Controller) ReplaceWithFullFile(index int, version domain.FileVersion) error {
	return c.data.ReplaceWithFullFile(index, version)
}

// RestoreToDiff restores a file to diff view
func (c *Controller) RestoreToDiff(index int) error {
	return c.data.RestoreToDiff(index)
}

// AddFilter adds a filter
func (c *Controller) AddFilter(filter domain.Filter) {
	c.data.AddFilter(filter)
}

// RemoveFilter removes a filter
func (c *Controller) RemoveFilter(index int) error {
	return c.data.RemoveFilter(index)
}

// AddContextFile adds a file from the repository as context
func (c *Controller) AddContextFile(path string) error {
	// Get file content from current branch
	content, err := c.gitService.GetFileContent(c.data.SourceBranch, path)
	if err != nil {
		return fmt.Errorf("failed to get file content: %w", err)
	}
	
	tokens := len(content) / 4 // rough estimate
	
	c.data.AddContextItem(domain.ContextItem{
		Type:    domain.ContextTypeFile,
		Path:    path,
		Content: content,
		Tokens:  tokens,
	})
	
	return nil
}

// AddContextNote adds a text note as context
func (c *Controller) AddContextNote(content string) {
	tokens := len(content) / 4 // rough estimate
	
	c.data.AddContextItem(domain.ContextItem{
		Type:    domain.ContextTypeNote,
		Content: content,
		Tokens:  tokens,
	})
}

// RemoveContextItem removes a context item
func (c *Controller) RemoveContextItem(index int) error {
	return c.data.RemoveContextItem(index)
}

// SetPrompt sets the current prompt
func (c *Controller) SetPrompt(prompt string, preset *domain.PromptPreset) {
	c.data.SetPrompt(prompt, preset)
}

// LoadPromptPreset loads a prompt preset
func (c *Controller) LoadPromptPreset(presetID string) error {
	// Check built-in presets
	builtins := domain.GetBuiltinPresets()
	for _, preset := range builtins {
		if preset.ID == presetID {
			c.data.SetPrompt(preset.Template, &preset)
			return nil
		}
	}
	
	// Check project presets
	projectPresets, err := c.LoadProjectPresets()
	if err == nil {
		for _, preset := range projectPresets {
			if preset.ID == presetID {
				c.data.SetPrompt(preset.Template, &preset)
				return nil
			}
		}
	}
	
	// Check global presets
	globalPresets, err := c.LoadGlobalPresets()
	if err == nil {
		for _, preset := range globalPresets {
			if preset.ID == presetID {
				c.data.SetPrompt(preset.Template, &preset)
				return nil
			}
		}
	}
	
	return fmt.Errorf("preset not found: %s", presetID)
}

// LoadProjectPresets loads presets from project directory
func (c *Controller) LoadProjectPresets() ([]domain.PromptPreset, error) {
	presetDir := filepath.Join(c.repoPath, ".pr-builder", "prompts")
	return c.loadPresetsFromDir(presetDir, domain.PresetLocationProject)
}

// LoadGlobalPresets loads presets from global directory
func (c *Controller) LoadGlobalPresets() ([]domain.PromptPreset, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	presetDir := filepath.Join(homeDir, ".pr-builder", "prompts")
	return c.loadPresetsFromDir(presetDir, domain.PresetLocationGlobal)
}

// loadPresetsFromDir loads presets from a directory
func (c *Controller) loadPresetsFromDir(dir string, location domain.PresetLocation) ([]domain.PromptPreset, error) {
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

// SavePromptPreset saves a prompt preset
func (c *Controller) SavePromptPreset(name, description, template string, location domain.PresetLocation) error {
	var dir string
	
	if location == domain.PresetLocationProject {
		dir = filepath.Join(c.repoPath, ".pr-builder", "prompts")
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		dir = filepath.Join(homeDir, ".pr-builder", "prompts")
	}
	
	// Create directory if it doesn't exist
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	
	// Create filename from name
	filename := strings.ToLower(strings.ReplaceAll(name, " ", "_")) + ".yaml"
	filePath := filepath.Join(dir, filename)
	
	// Create preset data
	preset := struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
		Template    string `yaml:"template"`
	}{
		Name:        name,
		Description: description,
		Template:    template,
	}
	
	// Marshal to YAML
	data, err := yaml.Marshal(preset)
	if err != nil {
		return err
	}
	
	// Write to file
	return os.WriteFile(filePath, data, 0644)
}

// GenerateDescription generates a PR description using the API
func (c *Controller) GenerateDescription() (string, error) {
	// Validate we have content to generate from
	visibleFiles := c.data.GetVisibleFiles()
	includedFiles := make([]domain.FileChange, 0)
	for _, file := range visibleFiles {
		if file.Included {
			includedFiles = append(includedFiles, file)
		}
	}
	
	if len(includedFiles) == 0 {
		return "", fmt.Errorf("no files included for generation")
	}
	
	// Build request
	req := api.GenerateDescriptionRequest{
		SourceBranch:      c.data.SourceBranch,
		TargetBranch:      c.data.TargetBranch,
		Files:             includedFiles,
		AdditionalContext: c.data.AdditionalContext,
		Prompt:            c.data.CurrentPrompt,
	}
	
	// Validate request
	if err := c.apiService.ValidateRequest(req); err != nil {
		return "", err
	}
	
	// Generate description
	resp, err := c.apiService.GenerateDescription(req)
	if err != nil {
		return "", err
	}
	
	c.data.GeneratedDescription = resp.Description
	return resp.Description, nil
}

// GetRepoFiles returns all files in the repository
func (c *Controller) GetRepoFiles() ([]string, error) {
	return c.gitService.ListFiles(c.data.SourceBranch)
}

// GetFilters returns all active filters
func (c *Controller) GetFilters() []domain.Filter {
	return c.data.ActiveFilters
}

// ClearFilters removes all active filters
func (c *Controller) ClearFilters() {
	c.data.ActiveFilters = make([]domain.Filter, 0)
}

// TestFilter tests a filter pattern against files without applying it
func (c *Controller) TestFilter(filter domain.Filter) (matched []string, unmatched []string) {
	matched = make([]string, 0)
	unmatched = make([]string, 0)
	
	for _, file := range c.data.ChangedFiles {
		passes := true
		for _, rule := range filter.Rules {
			// Test pattern matching
			matches := domain.TestPattern(file.Path, rule.Pattern)
			
			if rule.Type == domain.FilterTypeExclude && matches {
				passes = false
				break
			}
			if rule.Type == domain.FilterTypeInclude && !matches {
				passes = false
				break
			}
		}
		
		if passes {
			matched = append(matched, file.Path)
		} else {
			unmatched = append(unmatched, file.Path)
		}
	}
	
	return matched, unmatched
}

// GetFilteredFiles returns files that are filtered out
func (c *Controller) GetFilteredFiles() []domain.FileChange {
	return c.data.GetFilteredFiles()
}

// GetVisibleFiles returns files that pass filters
func (c *Controller) GetVisibleFiles() []domain.FileChange {
	return c.data.GetVisibleFiles()
}
