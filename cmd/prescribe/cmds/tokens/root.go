package tokens

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// NewTokensCmd groups token/debug-related subcommands.
func NewTokensCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "tokens",
		Short: "Token counting utilities",
		Long:  "Token counting and debugging utilities (e.g. post-hoc counting of exported XML-ish payloads).",
	}

	countXMLCmd, err := NewCountXMLCobraCommand()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build tokens count-xml command")
	}

	cmd.AddCommand(countXMLCmd)
	return cmd, nil
}
