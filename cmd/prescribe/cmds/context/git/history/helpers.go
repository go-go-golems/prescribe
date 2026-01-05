package history

import (
	glazed_layers "github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/prescribe/internal/domain"
)

func effectiveGitHistoryConfig(data *domain.PRData) (domain.GitHistoryConfig, bool) {
	if data != nil && data.GitHistory != nil {
		return *data.GitHistory, true
	}
	return domain.DefaultGitHistoryConfig(), false
}

func parameterWasSet(parsedLayers *glazed_layers.ParsedLayers, slug, key string) bool {
	p, ok := parsedLayers.GetParameter(slug, key)
	if !ok {
		return false
	}
	if len(p.Log) == 0 {
		return false
	}
	return p.Log[0].Source != "default"
}
