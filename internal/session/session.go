package session

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-go-golems/prescribe/internal/domain"
	"github.com/go-go-golems/prescribe/internal/presets"
	"github.com/go-go-golems/prescribe/internal/tokens"
	"gopkg.in/yaml.v3"
)

// Session represents a PR builder session configuration
type Session struct {
	// Metadata
	Version string `yaml:"version"`

	// Git info
	SourceBranch string `yaml:"source_branch"`
	TargetBranch string `yaml:"target_branch"`

	// Derived git history configuration
	GitHistory *GitHistoryConfig `yaml:"git_history,omitempty"`
	// Derived git context item configuration (reference-based; materialized at generation time)
	GitContext []GitContextItemConfig `yaml:"git_context,omitempty"`

	// Optional PR metadata
	Title       string `yaml:"title,omitempty"`
	Description string `yaml:"description,omitempty"`

	// File configuration
	Files []FileConfig `yaml:"files"`

	// Filters
	Filters []FilterConfig `yaml:"filters,omitempty"`

	// Additional context
	Context []ContextConfig `yaml:"context,omitempty"`

	// Prompt
	Prompt PromptConfig `yaml:"prompt"`
}

// GitHistoryConfig represents the persisted git history settings in the session.
type GitHistoryConfig struct {
	Enabled        bool `yaml:"enabled"`
	MaxCommits     int  `yaml:"max_commits"`
	IncludeMerges  bool `yaml:"include_merges"`
	FirstParent    bool `yaml:"first_parent"`
	IncludeNumstat bool `yaml:"include_numstat"`
}

type GitContextItemConfig struct {
	Kind  string   `yaml:"kind"`
	Ref   string   `yaml:"ref,omitempty"`
	From  string   `yaml:"from,omitempty"`
	To    string   `yaml:"to,omitempty"`
	Path  string   `yaml:"path,omitempty"`
	Paths []string `yaml:"paths,omitempty"`
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
	defaultGitHistory := domain.DefaultGitHistoryConfig()
	effectiveGitHistory := defaultGitHistory
	if data.GitHistory != nil {
		effectiveGitHistory = *data.GitHistory
	}

	session := &Session{
		Version:      "1.0",
		SourceBranch: data.SourceBranch,
		TargetBranch: data.TargetBranch,
		GitHistory: &GitHistoryConfig{
			Enabled:        effectiveGitHistory.Enabled,
			MaxCommits:     effectiveGitHistory.MaxCommits,
			IncludeMerges:  effectiveGitHistory.IncludeMerges,
			FirstParent:    effectiveGitHistory.FirstParent,
			IncludeNumstat: effectiveGitHistory.IncludeNumstat,
		},
		Title:       data.Title,
		Description: data.Description,
		Files:       make([]FileConfig, 0),
		Filters:     make([]FilterConfig, 0),
		Context:     make([]ContextConfig, 0),
	}

	// Convert git context items
	if len(data.GitContext) > 0 {
		session.GitContext = make([]GitContextItemConfig, 0, len(data.GitContext))
		for _, item := range data.GitContext {
			session.GitContext = append(session.GitContext, GitContextItemConfig{
				Kind:  string(item.Kind),
				Ref:   item.Ref,
				From:  item.From,
				To:    item.To,
				Path:  item.Path,
				Paths: append([]string{}, item.Paths...),
			})
		}
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
// repoPath is required to resolve project presets
func (s *Session) ApplyToData(data *domain.PRData, repoPath string) error {
	// Apply PR metadata
	data.Title = s.Title
	data.Description = s.Description

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

			// Recompute tokens based on selected mode so the TUI totals are consistent
			// immediately after loading a session.
			switch file.Type {
			case domain.FileTypeDiff:
				file.Tokens = tokens.Count(file.Diff)
			case domain.FileTypeFull:
				switch file.Version {
				case domain.FileVersionBefore:
					file.Tokens = tokens.Count(file.FullBefore)
				case domain.FileVersionAfter:
					file.Tokens = tokens.Count(file.FullAfter)
				case domain.FileVersionBoth:
					file.Tokens = tokens.Count(file.FullBefore) + tokens.Count(file.FullAfter)
				default:
					// Best effort: count both if version is unspecified.
					file.Tokens = tokens.Count(file.FullBefore) + tokens.Count(file.FullAfter)
				}
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
		tokens_ := tokens.Count(cc.Content)
		data.AdditionalContext = append(data.AdditionalContext, domain.ContextItem{
			Type:    domain.ContextType(cc.Type),
			Path:    cc.Path,
			Content: cc.Content,
			Tokens:  tokens_,
		})
	}

	// Apply git history config (compatibility: missing block => nil, handled as "enabled defaults" downstream)
	if s.GitHistory != nil {
		data.GitHistory = &domain.GitHistoryConfig{
			Enabled:        s.GitHistory.Enabled,
			MaxCommits:     s.GitHistory.MaxCommits,
			IncludeMerges:  s.GitHistory.IncludeMerges,
			FirstParent:    s.GitHistory.FirstParent,
			IncludeNumstat: s.GitHistory.IncludeNumstat,
		}
	} else {
		data.GitHistory = nil
	}

	// Apply git context items (reference-based)
	data.GitContext = make([]domain.GitContextItem, 0, len(s.GitContext))
	for _, item := range s.GitContext {
		data.GitContext = append(data.GitContext, domain.GitContextItem{
			Kind:  domain.GitContextItemKind(item.Kind),
			Ref:   item.Ref,
			From:  item.From,
			To:    item.To,
			Path:  item.Path,
			Paths: append([]string{}, item.Paths...),
		})
	}

	// Apply prompt
	if s.Prompt.Preset != "" {
		// Find and apply preset (checks builtin, project, and global presets)
		preset, err := presets.ResolvePromptPreset(s.Prompt.Preset, repoPath)
		if err == nil {
			data.SetPrompt(preset.Template, preset)
			return nil
		}
		// If preset not found, fall through to template if available
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
