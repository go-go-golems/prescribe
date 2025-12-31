package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-go-golems/prescribe/internal/domain"
	"github.com/go-go-golems/prescribe/internal/tokens"
	"github.com/pkg/errors"
)

// Service provides git operations
type Service struct {
	repoPath string
}

// NewService creates a new git service
func NewService(repoPath string) (*Service, error) {
	// Verify it's a git repository
	gitDir := filepath.Join(repoPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("not a git repository: %s", repoPath)
	}

	return &Service{repoPath: repoPath}, nil
}

// ResolveCommit resolves a git ref (branch name, tag, SHA, etc) to a full commit SHA.
func (s *Service) ResolveCommit(ref string) (string, error) {
	cmd := exec.Command("git", "rev-parse", ref)
	cmd.Dir = s.repoPath
	output, err := cmd.Output()
	if err != nil {
		return "", errors.Wrap(err, "failed to resolve ref to commit")
	}
	return strings.TrimSpace(string(output)), nil
}

// GetCurrentBranch returns the current branch name
func (s *Service) GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = s.repoPath
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetDefaultBranch returns the default branch (main or master)
func (s *Service) GetDefaultBranch() (string, error) {
	// Try to get the default branch from remote
	cmd := exec.Command("git", "symbolic-ref", "refs/remotes/origin/HEAD")
	cmd.Dir = s.repoPath
	output, err := cmd.Output()
	if err == nil {
		branch := strings.TrimSpace(string(output))
		branch = strings.TrimPrefix(branch, "refs/remotes/origin/")
		return branch, nil
	}

	// Fallback: check if main exists, otherwise use master
	cmd = exec.Command("git", "rev-parse", "--verify", "main")
	cmd.Dir = s.repoPath
	if err := cmd.Run(); err == nil {
		return "main", nil
	}

	return "master", nil
}

// GetDiff returns the diff between two branches
func (s *Service) GetDiff(sourceBranch, targetBranch string) (string, error) {
	cmd := exec.Command("git", "diff", fmt.Sprintf("%s...%s", targetBranch, sourceBranch))
	cmd.Dir = s.repoPath
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get diff: %w", err)
	}
	return string(output), nil
}

// GetChangedFiles returns a list of changed files between two branches
func (s *Service) GetChangedFiles(sourceBranch, targetBranch string) ([]domain.FileChange, error) {
	// Get list of changed files with stats
	cmd := exec.Command("git", "diff", "--numstat", fmt.Sprintf("%s...%s", targetBranch, sourceBranch))
	cmd.Dir = s.repoPath
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get changed files: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	files := make([]domain.FileChange, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		additions := 0
		deletions := 0
		path := parts[2]

		// Parse additions and deletions
		if parts[0] != "-" {
			v, err := strconv.Atoi(parts[0])
			if err != nil {
				return nil, errors.Wrapf(err, "failed to parse additions for %s: %q", path, parts[0])
			}
			additions = v
		}
		if parts[1] != "-" {
			v, err := strconv.Atoi(parts[1])
			if err != nil {
				return nil, errors.Wrapf(err, "failed to parse deletions for %s: %q", path, parts[1])
			}
			deletions = v
		}

		// Get the diff for this file
		diff, err := s.GetFileDiff(sourceBranch, targetBranch, path)
		if err != nil {
			diff = ""
		}

		// Get full file content (before and after)
		fullBefore, _ := s.GetFileContent(targetBranch, path)
		fullAfter, _ := s.GetFileContent(sourceBranch, path)

		// Count tokens using tokenizer (preflight estimate)
		tokens_ := tokens.Count(diff)

		files = append(files, domain.FileChange{
			Path:       path,
			Included:   true, // Include by default
			Additions:  additions,
			Deletions:  deletions,
			Tokens:     tokens_,
			Type:       domain.FileTypeDiff,
			Diff:       diff,
			FullBefore: fullBefore,
			FullAfter:  fullAfter,
		})
	}

	return files, nil
}

// GetFileDiff returns the diff for a specific file
func (s *Service) GetFileDiff(sourceBranch, targetBranch, filePath string) (string, error) {
	cmd := exec.Command("git", "diff", fmt.Sprintf("%s...%s", targetBranch, sourceBranch), "--", filePath)
	cmd.Dir = s.repoPath
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get file diff: %w", err)
	}
	return string(output), nil
}

// GetFileContent returns the content of a file at a specific branch/commit
func (s *Service) GetFileContent(ref, filePath string) (string, error) {
	cmd := exec.Command("git", "show", fmt.Sprintf("%s:%s", ref, filePath))
	cmd.Dir = s.repoPath
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get file content: %w", err)
	}
	return string(output), nil
}

// ListFiles returns all files in the repository at a given ref
func (s *Service) ListFiles(ref string) ([]string, error) {
	cmd := exec.Command("git", "ls-tree", "-r", "--name-only", ref)
	cmd.Dir = s.repoPath
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	files := make([]string, 0, len(lines))
	for _, line := range lines {
		if line != "" {
			files = append(files, line)
		}
	}

	return files, nil
}

// PushCurrentBranch pushes the current branch to its upstream.
//
// Note: This intentionally does NOT set upstream (no -u). If the branch has no
// upstream configured, this will fail and the caller should surface a helpful
// message.
func (s *Service) PushCurrentBranch(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "git", "push")
	cmd.Dir = s.repoPath
	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "git push failed: %s", strings.TrimSpace(string(out)))
	}
	return nil
}
