package api

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-go-golems/prescribe/internal/domain"
)

// Service provides API operations for generating PR descriptions
type Service struct {
	// In a real implementation, this would have API credentials, HTTP client, etc.
}

// NewService creates a new API service
func NewService() *Service {
	return &Service{}
}

// GenerateDescriptionRequest contains the request data for generating a PR description
type GenerateDescriptionRequest struct {
	SourceBranch      string
	TargetBranch      string
	Files             []domain.FileChange
	AdditionalContext []domain.ContextItem
	Prompt            string
}

// GenerateDescriptionResponse contains the generated PR description
type GenerateDescriptionResponse struct {
	Description string
	TokensUsed  int
	Model       string
}

// GenerateDescription generates a PR description using an LLM (mock implementation)
func (s *Service) GenerateDescription(req GenerateDescriptionRequest) (*GenerateDescriptionResponse, error) {
	// Simulate API call delay
	time.Sleep(2 * time.Second)
	
	// Build a mock description based on the input
	var sb strings.Builder
	
	sb.WriteString("# Pull Request: ")
	sb.WriteString(req.SourceBranch)
	sb.WriteString(" → ")
	sb.WriteString(req.TargetBranch)
	sb.WriteString("\n\n")
	
	sb.WriteString("## Summary\n\n")
	sb.WriteString("This PR includes changes across ")
	sb.WriteString(fmt.Sprintf("%d files", len(req.Files)))
	sb.WriteString(", implementing new features and improvements.\n\n")
	
	sb.WriteString("## Changes\n\n")
	
	// List changed files
	for _, file := range req.Files {
		if !file.Included {
			continue
		}
		sb.WriteString(fmt.Sprintf("- **%s**: ", file.Path))
		if file.Additions > 0 {
			sb.WriteString(fmt.Sprintf("+%d lines", file.Additions))
		}
		if file.Deletions > 0 {
			if file.Additions > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("-%d lines", file.Deletions))
		}
		sb.WriteString("\n")
	}
	
	sb.WriteString("\n## Key Changes\n\n")
	
	// Generate mock key changes based on file names
	for i, file := range req.Files {
		if !file.Included || i >= 3 { // Only show first 3
			continue
		}
		
		// Extract meaningful info from path
		parts := strings.Split(file.Path, "/")
		fileName := parts[len(parts)-1]
		
		if strings.Contains(fileName, "auth") {
			sb.WriteString("- Enhanced authentication system with improved security measures\n")
		} else if strings.Contains(fileName, "test") {
			sb.WriteString("- Added comprehensive test coverage for new functionality\n")
		} else if strings.Contains(fileName, "api") {
			sb.WriteString("- Updated API endpoints with new features and validation\n")
		} else if strings.Contains(fileName, "middleware") {
			sb.WriteString("- Improved middleware with better error handling\n")
		} else {
			sb.WriteString(fmt.Sprintf("- Updated %s with new functionality\n", fileName))
		}
	}
	
	sb.WriteString("\n## Testing\n\n")
	sb.WriteString("- All existing tests pass\n")
	sb.WriteString("- New tests added for changed functionality\n")
	sb.WriteString("- Manual testing completed for critical paths\n")
	
	sb.WriteString("\n## Breaking Changes\n\n")
	sb.WriteString("None\n")
	
	if len(req.AdditionalContext) > 0 {
		sb.WriteString("\n## Additional Context\n\n")
		for _, ctx := range req.AdditionalContext {
			if ctx.Type == domain.ContextTypeNote {
				sb.WriteString(fmt.Sprintf("- %s\n", ctx.Content))
			}
		}
	}
	
	// Calculate mock token usage
	tokensUsed := len(sb.String()) / 4
	
	return &GenerateDescriptionResponse{
		Description: sb.String(),
		TokensUsed:  tokensUsed,
		Model:       "mock-gpt-4",
	}, nil
}

// ValidateRequest validates a generate description request
func (s *Service) ValidateRequest(req GenerateDescriptionRequest) error {
	if req.SourceBranch == "" {
		return fmt.Errorf("source branch is required")
	}
	if req.TargetBranch == "" {
		return fmt.Errorf("target branch is required")
	}
	if len(req.Files) == 0 {
		return fmt.Errorf("no files to generate description from")
	}
	
	// Check if at least one file is included
	hasIncluded := false
	for _, file := range req.Files {
		if file.Included {
			hasIncluded = true
			break
		}
	}
	if !hasIncluded {
		return fmt.Errorf("at least one file must be included")
	}
	
	return nil
}
