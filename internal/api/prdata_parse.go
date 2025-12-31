package api

import (
	"regexp"
	"strings"

	"github.com/go-go-golems/geppetto/pkg/events/structuredsink/parsehelpers"
	geppettoparse "github.com/go-go-golems/geppetto/pkg/steps/parse"
	"github.com/go-go-golems/prescribe/internal/domain"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// ParseGeneratedPRDataFromAssistantText attempts to extract and parse the structured PR YAML
// from an assistant output blob.
//
// Strategy:
// 1) Prefer the last fenced ```yaml``` block in the text (via geppetto parse helpers).
// 2) Fallback: strip a single fenced block using parsehelpers and parse the remaining body.
func ParseGeneratedPRDataFromAssistantText(assistantText string) (*domain.GeneratedPRData, error) {
	raw := strings.TrimSpace(assistantText)
	if raw == "" {
		return nil, errors.New("empty assistant output")
	}

	blocks, err := geppettoparse.ExtractYAMLBlocks(raw)
	if err == nil && len(blocks) > 0 {
		// Pinocchio-style robustness: scan from the end and pick the first candidate
		// that parses AND satisfies our expected schema constraints.
		for i := len(blocks) - 1; i >= 0; i-- {
			candidate := strings.TrimSpace(blocks[i])
			if candidate == "" {
				continue
			}
			out, err := parseGeneratedPRDataYAML([]byte(candidate))
			if err != nil {
				continue
			}
			if isGeneratedPRDataValid(out) {
				return out, nil
			}
		}
	}

	// Only strip code fences if the *entire* assistant output is fenced.
	//
	// Important: The structured YAML contract commonly includes fenced code blocks
	// inside YAML block scalars (e.g., release_notes.body contains ```bash ... ```).
	// A naive "find first ``` anywhere" approach will corrupt the YAML by stripping
	// from the first inner fence instead of an outer wrapper.
	body := []byte(raw)
	if strings.HasPrefix(raw, "```") {
		_, body = parsehelpers.StripCodeFenceBytes([]byte(raw))
	}
	if out, err := parseGeneratedPRDataYAML(body); err == nil && isGeneratedPRDataValid(out) {
		return out, nil
	}

	// Heuristic salvage: if the model emitted prose around the YAML (common in some providers),
	// attempt to parse from the last "title:" block to the end.
	salvaged, ok := trySalvageYAMLFromTitleBlock(string(body))
	if ok {
		if out, err := parseGeneratedPRDataYAML([]byte(salvaged)); err == nil && isGeneratedPRDataValid(out) {
			return out, nil
		}
	}

	return parseGeneratedPRDataYAML(body)
}

func parseGeneratedPRDataYAML(b []byte) (*domain.GeneratedPRData, error) {
	body := strings.TrimSpace(string(b))
	if body == "" {
		return nil, errors.New("empty YAML body")
	}
	// Common model failure mode: emits a key without ":" (e.g. "body" instead of "body: |").
	// Try a minimal repair before unmarshalling.
	body = repairCommonYAMLFormatting(body)
	var out domain.GeneratedPRData
	if err := yaml.Unmarshal([]byte(body), &out); err != nil {
		return nil, errors.Wrap(err, "failed to parse PR YAML")
	}
	return &out, nil
}

var reYAMLTitleStart = regexp.MustCompile(`(?m)^[ \t]*title:[ \t]*`)
var reBareKeyLine = regexp.MustCompile(`(?m)^(body|changelog|release_notes|release-notes)[ \t]*$`)

func isGeneratedPRDataValid(d *domain.GeneratedPRData) bool {
	if d == nil {
		return false
	}
	// Minimal schema check: at least title + body should be present.
	// (changelog/release_notes are contractually expected, but we keep validation permissive
	// to avoid rejecting partial-yet-useful outputs.)
	return strings.TrimSpace(d.Title) != "" && strings.TrimSpace(d.Body) != ""
}

func repairCommonYAMLFormatting(s string) string {
	// Replace bare key lines like:
	//   body
	// with a minimal placeholder that remains valid YAML:
	//   body: ""
	//
	// This lets us parse partial outputs and surface a clearer follow-up path
	// (and avoids hard parse failures on trivial formatting mistakes).
	return reBareKeyLine.ReplaceAllString(s, `$1: ""`)
}

func trySalvageYAMLFromTitleBlock(s string) (string, bool) {
	raw := strings.TrimSpace(s)
	if raw == "" {
		return "", false
	}

	locs := reYAMLTitleStart.FindAllStringIndex(raw, -1)
	if len(locs) == 0 {
		return "", false
	}
	// Prefer the last occurrence (models often show an example earlier and the real output later).
	start := locs[len(locs)-1][0]
	out := strings.TrimSpace(raw[start:])
	if out == "" {
		return "", false
	}
	return out, true
}
