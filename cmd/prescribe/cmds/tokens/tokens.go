package tokens

import (
	"sync"

	"github.com/spf13/cobra"
)

// TokensCmd groups token/debug-related subcommands.
var TokensCmd = &cobra.Command{
	Use:   "tokens",
	Short: "Token counting utilities",
	Long:  "Token counting and debugging utilities (e.g. post-hoc counting of exported XML-ish payloads).",
}

var initOnce sync.Once
var initErr error

func Init() error {
	initOnce.Do(func() {
		if err := InitCountXMLCmd(); err != nil {
			initErr = err
			return
		}
		TokensCmd.AddCommand(
			CountXMLCmd,
		)
	})
	return initErr
}
