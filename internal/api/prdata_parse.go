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
		candidate := strings.TrimSpace(blocks[len(blocks)-1])
		if candidate != "" {
			return parseGeneratedPRDataYAML([]byte(candidate))
		}
	}

	_, body := parsehelpers.StripCodeFenceBytes([]byte(raw))
	if out, err := parseGeneratedPRDataYAML(body); err == nil {
		return out, nil
	}

	// Heuristic salvage: if the model emitted prose around the YAML (common in some providers),
	// attempt to parse from the last "title:" block to the end.
	salvaged, ok := trySalvageYAMLFromTitleBlock(string(body))
	if ok {
		return parseGeneratedPRDataYAML([]byte(salvaged))
	}

	return parseGeneratedPRDataYAML(body)
}

func parseGeneratedPRDataYAML(b []byte) (*domain.GeneratedPRData, error) {
	body := strings.TrimSpace(string(b))
	if body == "" {
		return nil, errors.New("empty YAML body")
	}
	var out domain.GeneratedPRData
	if err := yaml.Unmarshal([]byte(body), &out); err != nil {
		return nil, errors.Wrap(err, "failed to parse PR YAML")
	}
	return &out, nil
}

var reYAMLTitleStart = regexp.MustCompile(`(?m)^[ \t]*title:[ \t]*`)

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
