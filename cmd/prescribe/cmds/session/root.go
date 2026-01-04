package session

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// NewSessionCmd groups all session-related subcommands.
func NewSessionCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "session",
		Short: "Manage sessions",
		Long:  "Initialize, inspect, save, and load PR builder sessions.",
	}

	initCmd, err := NewInitCobraCommand()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build session init command")
	}
	saveCmd, err := NewSaveCobraCommand()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build session save command")
	}
	loadCmd, err := NewLoadCobraCommand()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build session load command")
	}
	showCmd, err := NewShowCobraCommand()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build session show command")
	}
	tokenCountCmd, err := NewTokenCountCobraCommand()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build session token-count command")
	}

	cmd.AddCommand(initCmd, saveCmd, loadCmd, showCmd, tokenCountCmd)
	return cmd, nil
}
