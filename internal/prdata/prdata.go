package prdata

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-go-golems/prescribe/internal/domain"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func LastGeneratedPRDataPath(repoPath string) string {
	return filepath.Join(repoPath, ".pr-builder", "last-generated-pr.yaml")
}

func FailurePRDataPath(repoPath string, now time.Time) string {
	// Use a timestamped filename so repeated failures don't overwrite.
	ts := now.UTC().Format("20060102-150405")
	return filepath.Join(repoPath, ".pr-builder", fmt.Sprintf("pr-data-%s.yaml", ts))
}

func LoadGeneratedPRDataFromYAMLFile(path string) (*domain.GeneratedPRData, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read PR data YAML file")
	}

	var out domain.GeneratedPRData
	if err := yaml.Unmarshal(b, &out); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal PR data YAML")
	}

	if strings.TrimSpace(out.Title) == "" {
		return nil, errors.New("invalid PR data YAML: missing title")
	}
	if strings.TrimSpace(out.Body) == "" {
		return nil, errors.New("invalid PR data YAML: missing body")
	}

	return &out, nil
}

func WriteGeneratedPRDataToYAMLFile(path string, data *domain.GeneratedPRData) error {
	if data == nil {
		return errors.New("generated PR data is nil")
	}
	if strings.TrimSpace(data.Title) == "" {
		return errors.New("generated PR data missing title")
	}
	if strings.TrimSpace(data.Body) == "" {
		return errors.New("generated PR data missing body")
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return errors.Wrap(err, "failed to create directory for PR data YAML")
	}

	b, err := yaml.Marshal(data)
	if err != nil {
		return errors.Wrap(err, "failed to marshal PR data YAML")
	}

	if err := os.WriteFile(path, b, 0o644); err != nil {
		return errors.Wrap(err, "failed to write PR data YAML file")
	}

	return nil
}
