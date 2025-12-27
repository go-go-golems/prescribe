package prompts

import (
	_ "embed"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// DefaultPrompt returns prescribe's default prompt template.
//
// This prompt is adapted from:
// pinocchio/cmd/pinocchio/prompts/code/create-pull-request.yaml
func DefaultPrompt() string {
	return strings.TrimSpace(get().CombinedText)
}

type pinocchioPromptYAML struct {
	SystemPrompt string `yaml:"system-prompt"`
	Prompt       string `yaml:"prompt"`
}

type loaded struct {
	CombinedText string
}

var (
	once   sync.Once
	cached loaded
)

//go:embed assets/create-pull-request.yaml
var createPullRequestYAML []byte

func get() loaded {
	once.Do(func() {
		var p pinocchioPromptYAML
		if err := yaml.Unmarshal(createPullRequestYAML, &p); err != nil {
			// Should never happen (embedded file). Fall back to a minimal prompt.
			cached = loaded{
				CombinedText: "Generate a pull request description based on the provided diff and context.",
			}
			return
		}

		// prescribe stores a single prompt string; we combine system + prompt.
		cached = loaded{
			CombinedText: strings.TrimSpace(p.SystemPrompt) + "\n\n" + strings.TrimSpace(p.Prompt),
		}
	})
	return cached
}


