package history

import "github.com/go-go-golems/prescribe/internal/domain"

func effectiveGitHistoryConfig(data *domain.PRData) (domain.GitHistoryConfig, bool) {
	if data != nil && data.GitHistory != nil {
		return *data.GitHistory, true
	}
	return domain.DefaultGitHistoryConfig(), false
}
