package github

import (
	"context"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

type CreatePROptions struct {
	Title string
	Body  string

	// Base is the base branch (e.g., "main").
	Base string

	// Draft creates a draft PR when true.
	Draft bool
}

type Service struct {
	repoPath string
}

func NewService(repoPath string) *Service {
	return &Service{repoPath: repoPath}
}

func BuildGhCreatePRArgs(opts CreatePROptions) ([]string, error) {
	if strings.TrimSpace(opts.Title) == "" {
		return nil, errors.New("missing PR title")
	}
	if strings.TrimSpace(opts.Body) == "" {
		return nil, errors.New("missing PR body")
	}

	args := []string{"pr", "create", "--title", opts.Title, "--body", opts.Body}

	if strings.TrimSpace(opts.Base) != "" {
		args = append(args, "--base", opts.Base)
	}
	if opts.Draft {
		args = append(args, "--draft")
	}

	return args, nil
}

func RedactGhArgs(args []string) []string {
	out := make([]string, len(args))
	copy(out, args)
	for i := 0; i < len(out); i++ {
		if out[i] == "--body" && i+1 < len(out) {
			out[i+1] = "<omitted>"
		}
	}
	return out
}

func (s *Service) CreatePR(ctx context.Context, opts CreatePROptions) (string, error) {
	args, err := BuildGhCreatePRArgs(opts)
	if err != nil {
		return "", err
	}

	cmd := exec.CommandContext(ctx, "gh", args...)
	cmd.Dir = s.repoPath
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.Wrapf(err, "gh %s failed: %s", strings.Join(RedactGhArgs(args), " "), strings.TrimSpace(string(out)))
	}

	return string(out), nil
}
