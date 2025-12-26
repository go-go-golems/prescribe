package session

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-go-golems/prescribe/internal/domain"
	"gopkg.in/yaml.v3"
)

// Session represents a PR builder session configuration
type Session struct {
	// Metadata
	Version string `yaml:"version"`
	
	// Git info
	SourceBranch string `yaml:"source_branch"`
	TargetBranch string `yaml:"target_branch"`
	
	// File configuration
	Files []FileConfig `yaml:"files"`
	
	// Filters
	Filters []FilterConfig `yaml:"filters,omitempty"`
	
	// Additional context
	Context []ContextConfig `yaml:"context,omitempty"`
	
	// Prompt
	Prompt PromptConfig `yaml:"prompt"`
}

// FileConfig represents a file's configuration in the session
type FileConfig struct {
	Path     string `yaml:"path"`
	Included bool   `yaml:"included"`
	Mode     string `yaml:"mode"` // "diff", "full_before", "full_after", "full_both"
}

// FilterConfig represents a filter in the session
type FilterConfig struct {
	Name        string       `yaml:"name"`
	Description string       `yaml:"description,omitempty"`
	Rules       []FilterRule `yaml:"rules"`
}

// FilterRule represents a single filter rule
type FilterRule struct {
	Type    string `yaml:"type"`    // "include" or "exclude"
	Pattern string `yaml:"pattern"` // glob pattern
}

// ContextConfig represents additional context
type ContextConfig struct {
	Type    string `yaml:"type"` // "file" or "note"
	Path    string `yaml:"path,omitempty"`
	Content string `yaml:"content,omitempty"`
}

// PromptConfig represents the prompt configuration
type PromptConfig struct {
	Preset   string `yaml:"preset,omitempty"`   // preset ID if using preset
	Template string `yaml:"template,omitempty"` // custom template if not using preset
}

// NewSession creates a new session from PR data
func NewSession(data *domain.PRData) *Session {
	session := &Session{
		Version:      "1.0",
		SourceBranch: data.SourceBranch,
		TargetBranch: data.TargetBranch,
		Files:        make([]FileConfig, 0),
		Filters:      make([]FilterConfig, 0),
		Context:      make([]ContextConfig, 0),
	}
	
	// Convert files
	for _, file := range data.ChangedFiles {
		mode := "diff"
		if file.Type == domain.FileTypeFull {
			switch file.Version {
			case domain.FileVersionBefore:
				mode = "full_before"
			case domain.FileVersionAfter:
				mode = "full_after"
			case domain.FileVersionBoth:
				mode = "full_both"
			}
		}
		
		session.Files = append(session.Files, FileConfig{
			Path:     file.Path,
			Included: file.Included,
			Mode:     mode,
		})
	}
	
	// Convert filters
	for _, filter := range data.ActiveFilters {
		rules := make([]FilterRule, 0)
		for _, rule := range filter.Rules {
			rules = append(rules, FilterRule{
				Type:    string(rule.Type),
				Pattern: rule.Pattern,
			})
		}
		
		session.Filters = append(session.Filters, FilterConfig{
			Name:        filter.Name,
			Description: filter.Description,
			Rules:       rules,
		})
	}
	
	// Convert context
	for _, ctx := range data.AdditionalContext {
		session.Context = append(session.Context, ContextConfig{
			Type:    string(ctx.Type),
			Path:    ctx.Path,
			Content: ctx.Content,
		})
	}
	
	// Convert prompt
	if data.CurrentPreset != nil {
		session.Prompt = PromptConfig{
			Preset: data.CurrentPreset.ID,
		}
	} else {
		session.Prompt = PromptConfig{
			Template: data.CurrentPrompt,
		}
	}
	
	return session
}

// ApplyToData applies the session configuration to PR data
func (s *Session) ApplyToData(data *domain.PRData) error {
	// Apply file configurations
	fileMap := make(map[string]FileConfig)
	for _, fc := range s.Files {
		fileMap[fc.Path] = fc
	}
	
	for i := range data.ChangedFiles {
		file := &data.ChangedFiles[i]
		if fc, ok := fileMap[file.Path]; ok {
			file.Included = fc.Included
			
			// Apply mode
			switch fc.Mode {
			case "diff":
				file.Type = domain.FileTypeDiff
			case "full_before":
				file.Type = domain.FileTypeFull
				file.Version = domain.FileVersionBefore
			case "full_after":
				file.Type = domain.FileTypeFull
				file.Version = domain.FileVersionAfter
			case "full_both":
				file.Type = domain.FileTypeFull
				file.Version = domain.FileVersionBoth
			}
		}
	}
	
	// Apply filters
	data.ActiveFilters = make([]domain.Filter, 0)
	for _, fc := range s.Filters {
		rules := make([]domain.FilterRule, 0)
		for _, rule := range fc.Rules {
			rules = append(rules, domain.FilterRule{
				Type:    domain.FilterType(rule.Type),
				Pattern: rule.Pattern,
			})
		}
		
		data.ActiveFilters = append(data.ActiveFilters, domain.Filter{
			Name:        fc.Name,
			Description: fc.Description,
			Rules:       rules,
		})
	}
	
	// Apply context
	data.AdditionalContext = make([]domain.ContextItem, 0)
	for _, cc := range s.Context {
		tokens := len(cc.Content) / 4
		data.AdditionalContext = append(data.AdditionalContext, domain.ContextItem{
			Type:    domain.ContextType(cc.Type),
			Path:    cc.Path,
			Content: cc.Content,
			Tokens:  tokens,
		})
	}
	
	// Apply prompt
	if s.Prompt.Preset != "" {
		// Find and apply preset
		builtins := domain.GetBuiltinPresets()
		for _, preset := range builtins {
			if preset.ID == s.Prompt.Preset {
				data.SetPrompt(preset.Template, &preset)
				return nil
			}
		}
	}
	
	if s.Prompt.Template != "" {
		data.SetPrompt(s.Prompt.Template, nil)
	}
	
	return nil
}

// Save saves the session to a YAML file
func (s *Session) Save(path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Marshal to YAML
	data, err := yaml.Marshal(s)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}
	
	// Write to file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}
	
	return nil
}

// Load loads a session from a YAML file
func Load(path string) (*Session, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}
	
	var session Session
	if err := yaml.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}
	
	return &session, nil
}

// GetDefaultSessionPath returns the default session file path for a repo
func GetDefaultSessionPath(repoPath string) string {
	return filepath.Join(repoPath, ".pr-builder", "session.yaml")
}
