package model

import (
	"fmt"
	"strings"
)

// FileChange represents a changed file in the PR
type FileChange struct {
	Path      string
	Included  bool
	Additions int
	Deletions int
	Tokens    int
	Type      FileType
	Version   FileVersion
	Diff      string
	FullBefore string
	FullAfter string
}

type FileType string

const (
	FileTypeDiff FileType = "diff"
	FileTypeFull FileType = "full_file"
)

type FileVersion string

const (
	FileVersionBefore FileVersion = "before"
	FileVersionAfter  FileVersion = "after"
	FileVersionBoth   FileVersion = "both"
)

// FilterRule represents a file filter rule
type FilterRule struct {
	Type    FilterType
	Pattern string
	Order   int
}

type FilterType string

const (
	FilterTypeInclude FilterType = "include"
	FilterTypeExclude FilterType = "exclude"
)

// Filter represents a named filter with rules
type Filter struct {
	Name        string
	Description string
	Rules       []FilterRule
}

// ContextItem represents additional context (file or note)
type ContextItem struct {
	Type    ContextType
	Path    string
	Content string
	Tokens  int
}

type ContextType string

const (
	ContextTypeFile ContextType = "file"
	ContextTypeNote ContextType = "note"
)

// PromptPreset represents a prompt template
type PromptPreset struct {
	ID          string
	Name        string
	Description string
	Template    string
	Location    PresetLocation
}

type PresetLocation string

const (
	PresetLocationBuiltin PresetLocation = "builtin"
	PresetLocationProject PresetLocation = "project"
	PresetLocationGlobal  PresetLocation = "global"
)

// PRBuilderModel is the core model for the application
type PRBuilderModel struct {
	// Git information
	SourceBranch string
	TargetBranch string
	
	// Files
	ChangedFiles []FileChange
	AdditionalContext []ContextItem
	
	// Filters
	ActiveFilters []Filter
	
	// Prompt
	CurrentPrompt string
	CurrentPreset *PromptPreset
	
	// Generated description
	GeneratedDescription string
}

// NewPRBuilderModel creates a new model
func NewPRBuilderModel() *PRBuilderModel {
	return &PRBuilderModel{
		ChangedFiles: make([]FileChange, 0),
		AdditionalContext: make([]ContextItem, 0),
		ActiveFilters: make([]Filter, 0),
		CurrentPrompt: GetDefaultPrompt(),
	}
}

// GetDefaultPrompt returns the default prompt template
func GetDefaultPrompt() string {
	return "Generate a clear PR description with: summary of changes, motivation, key changes, testing notes, and breaking changes if any."
}

// GetVisibleFiles returns files that pass the active filters
func (m *PRBuilderModel) GetVisibleFiles() []FileChange {
	if len(m.ActiveFilters) == 0 {
		return m.ChangedFiles
	}
	
	visible := make([]FileChange, 0)
	for _, file := range m.ChangedFiles {
		if m.passesFilters(file.Path) {
			visible = append(visible, file)
		}
	}
	return visible
}

// GetFilteredFiles returns files that don't pass the active filters
func (m *PRBuilderModel) GetFilteredFiles() []FileChange {
	if len(m.ActiveFilters) == 0 {
		return []FileChange{}
	}
	
	filtered := make([]FileChange, 0)
	for _, file := range m.ChangedFiles {
		if !m.passesFilters(file.Path) {
			filtered = append(filtered, file)
		}
	}
	return filtered
}

// passesFilters checks if a file path passes all active filters
func (m *PRBuilderModel) passesFilters(path string) bool {
	// If no filters, all files pass
	if len(m.ActiveFilters) == 0 {
		return true
	}
	
	// Apply each filter's rules
	for _, filter := range m.ActiveFilters {
		for _, rule := range filter.Rules {
			matches := matchesPattern(path, rule.Pattern)
			
			if rule.Type == FilterTypeExclude && matches {
				return false
			}
			if rule.Type == FilterTypeInclude && !matches {
				return false
			}
		}
	}
	
	return true
}

// matchesPattern performs simple glob-style pattern matching
func matchesPattern(path, pattern string) bool {
	// Simple implementation - just check if pattern is contained
	// In a real implementation, use filepath.Match or a glob library
	if strings.HasPrefix(pattern, "*") && strings.HasSuffix(pattern, "*") {
		return strings.Contains(path, strings.Trim(pattern, "*"))
	}
	if strings.HasPrefix(pattern, "*") {
		return strings.HasSuffix(path, strings.TrimPrefix(pattern, "*"))
	}
	if strings.HasSuffix(pattern, "*") {
		return strings.HasPrefix(path, strings.TrimSuffix(pattern, "*"))
	}
	return strings.Contains(path, pattern)
}

// GetTotalTokens calculates total tokens for all included content
func (m *PRBuilderModel) GetTotalTokens() int {
	total := 0
	
	// Count visible files
	for _, file := range m.GetVisibleFiles() {
		if file.Included {
			total += file.Tokens
		}
	}
	
	// Count additional context
	for _, ctx := range m.AdditionalContext {
		total += ctx.Tokens
	}
	
	return total
}

// ToggleFileInclusion toggles whether a file is included in the context
func (m *PRBuilderModel) ToggleFileInclusion(index int) error {
	if index < 0 || index >= len(m.ChangedFiles) {
		return fmt.Errorf("invalid file index: %d", index)
	}
	m.ChangedFiles[index].Included = !m.ChangedFiles[index].Included
	return nil
}

// ReplaceWithFullFile replaces a file's diff with full file content
func (m *PRBuilderModel) ReplaceWithFullFile(index int, version FileVersion) error {
	if index < 0 || index >= len(m.ChangedFiles) {
		return fmt.Errorf("invalid file index: %d", index)
	}
	
	file := &m.ChangedFiles[index]
	file.Type = FileTypeFull
	file.Version = version
	
	// Recalculate tokens based on version
	// In a real implementation, this would use actual token counting
	switch version {
	case FileVersionBefore:
		file.Tokens = len(file.FullBefore) / 4 // rough estimate
	case FileVersionAfter:
		file.Tokens = len(file.FullAfter) / 4
	case FileVersionBoth:
		file.Tokens = (len(file.FullBefore) + len(file.FullAfter)) / 4
	}
	
	return nil
}

// RestoreToDiff restores a file from full file back to diff
func (m *PRBuilderModel) RestoreToDiff(index int) error {
	if index < 0 || index >= len(m.ChangedFiles) {
		return fmt.Errorf("invalid file index: %d", index)
	}
	
	file := &m.ChangedFiles[index]
	file.Type = FileTypeDiff
	file.Version = ""
	file.Tokens = len(file.Diff) / 4 // rough estimate
	
	return nil
}

// AddFilter adds a filter to the active filters
func (m *PRBuilderModel) AddFilter(filter Filter) {
	m.ActiveFilters = append(m.ActiveFilters, filter)
}

// RemoveFilter removes a filter by index
func (m *PRBuilderModel) RemoveFilter(index int) error {
	if index < 0 || index >= len(m.ActiveFilters) {
		return fmt.Errorf("invalid filter index: %d", index)
	}
	m.ActiveFilters = append(m.ActiveFilters[:index], m.ActiveFilters[index+1:]...)
	return nil
}

// AddContextItem adds a context item (file or note)
func (m *PRBuilderModel) AddContextItem(item ContextItem) {
	m.AdditionalContext = append(m.AdditionalContext, item)
}

// RemoveContextItem removes a context item by index
func (m *PRBuilderModel) RemoveContextItem(index int) error {
	if index < 0 || index >= len(m.AdditionalContext) {
		return fmt.Errorf("invalid context item index: %d", index)
	}
	m.AdditionalContext = append(m.AdditionalContext[:index], m.AdditionalContext[index+1:]...)
	return nil
}

// SetPrompt sets the current prompt template
func (m *PRBuilderModel) SetPrompt(prompt string, preset *PromptPreset) {
	m.CurrentPrompt = prompt
	m.CurrentPreset = preset
}

// GetBuiltinPresets returns the built-in prompt presets
func GetBuiltinPresets() []PromptPreset {
	return []PromptPreset{
		{
			ID:          "default",
			Name:        "Default",
			Description: "Standard PR description format",
			Template:    GetDefaultPrompt(),
			Location:    PresetLocationBuiltin,
		},
		{
			ID:          "detailed",
			Name:        "Detailed",
			Description: "Comprehensive PR description with executive summary",
			Template:    "Create a comprehensive PR description including: Executive summary, detailed changes by component, rationale, testing strategy, deployment notes, and rollback plan.",
			Location:    PresetLocationBuiltin,
		},
		{
			ID:          "concise",
			Name:        "Concise",
			Description: "Brief and to-the-point PR description",
			Template:    "Write a brief PR description: What changed, why, and how to test.",
			Location:    PresetLocationBuiltin,
		},
		{
			ID:          "conventional",
			Name:        "Conventional Commits",
			Description: "Follows conventional commits format",
			Template:    "Generate PR description following conventional commits format with type, scope, breaking changes, and footer.",
			Location:    PresetLocationBuiltin,
		},
	}
}
