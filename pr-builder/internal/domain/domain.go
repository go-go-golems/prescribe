package domain

import (
	"fmt"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
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

// PRData is the core domain data for the application
type PRData struct {
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

// NewPRData creates a new PR data instance
func NewPRData() *PRData {
	return &PRData{
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
func (d *PRData) GetVisibleFiles() []FileChange {
	if len(d.ActiveFilters) == 0 {
		return d.ChangedFiles
	}
	
	visible := make([]FileChange, 0)
	for _, file := range d.ChangedFiles {
		if d.passesFilters(file.Path) {
			visible = append(visible, file)
		}
	}
	return visible
}

// GetFilteredFiles returns files that don't pass the active filters
func (d *PRData) GetFilteredFiles() []FileChange {
	if len(d.ActiveFilters) == 0 {
		return []FileChange{}
	}
	
	filtered := make([]FileChange, 0)
	for _, file := range d.ChangedFiles {
		if !d.passesFilters(file.Path) {
			filtered = append(filtered, file)
		}
	}
	return filtered
}

// passesFilters checks if a file path passes all active filters
func (d *PRData) passesFilters(path string) bool {
	// If no filters, all files pass
	if len(d.ActiveFilters) == 0 {
		return true
	}
	
	// Apply each filter's rules
	for _, filter := range d.ActiveFilters {
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

// matchesPattern performs glob-style pattern matching using doublestar
// Supports *, **, ?, [abc], and other glob patterns
func matchesPattern(path, pattern string) bool {
	// Use doublestar for proper glob matching
	matched, err := doublestar.Match(pattern, path)
	if err != nil {
		// If pattern is invalid, fall back to substring matching
		return strings.Contains(path, pattern)
	}
	return matched
}

// GetTotalTokens calculates total tokens for all included content
func (d *PRData) GetTotalTokens() int {
	total := 0
	
	// Count visible files
	for _, file := range d.GetVisibleFiles() {
		if file.Included {
			total += file.Tokens
		}
	}
	
	// Count additional context
	for _, ctx := range d.AdditionalContext {
		total += ctx.Tokens
	}
	
	return total
}

// ToggleFileInclusion toggles whether a file is included in the context
func (d *PRData) ToggleFileInclusion(index int) error {
	if index < 0 || index >= len(d.ChangedFiles) {
		return fmt.Errorf("invalid file index: %d", index)
	}
	d.ChangedFiles[index].Included = !d.ChangedFiles[index].Included
	return nil
}

// ReplaceWithFullFile replaces a file's diff with full file content
func (d *PRData) ReplaceWithFullFile(index int, version FileVersion) error {
	if index < 0 || index >= len(d.ChangedFiles) {
		return fmt.Errorf("invalid file index: %d", index)
	}
	
	file := &d.ChangedFiles[index]
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
func (d *PRData) RestoreToDiff(index int) error {
	if index < 0 || index >= len(d.ChangedFiles) {
		return fmt.Errorf("invalid file index: %d", index)
	}
	
	file := &d.ChangedFiles[index]
	file.Type = FileTypeDiff
	file.Version = ""
	file.Tokens = len(file.Diff) / 4 // rough estimate
	
	return nil
}

// AddFilter adds a filter to the active filters
func (d *PRData) AddFilter(filter Filter) {
	d.ActiveFilters = append(d.ActiveFilters, filter)
}

// RemoveFilter removes a filter by index
func (d *PRData) RemoveFilter(index int) error {
	if index < 0 || index >= len(d.ActiveFilters) {
		return fmt.Errorf("invalid filter index: %d", index)
	}
	d.ActiveFilters = append(d.ActiveFilters[:index], d.ActiveFilters[index+1:]...)
	return nil
}

// AddContextItem adds a context item (file or note)
func (d *PRData) AddContextItem(item ContextItem) {
	d.AdditionalContext = append(d.AdditionalContext, item)
}

// RemoveContextItem removes a context item by index
func (d *PRData) RemoveContextItem(index int) error {
	if index < 0 || index >= len(d.AdditionalContext) {
		return fmt.Errorf("invalid context item index: %d", index)
	}
	d.AdditionalContext = append(d.AdditionalContext[:index], d.AdditionalContext[index+1:]...)
	return nil
}

// SetPrompt sets the current prompt template
func (d *PRData) SetPrompt(prompt string, preset *PromptPreset) {
	d.CurrentPrompt = prompt
	d.CurrentPreset = preset
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

// TestPattern tests if a path matches a pattern (exported for testing)
func TestPattern(path, pattern string) bool {
	return matchesPattern(path, pattern)
}
