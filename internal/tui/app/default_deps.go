package app

import (
	"time"

	"github.com/atotto/clipboard"
)

// DefaultDeps is the production dependency set for the TUI.
//
// Clipboard behavior depends on the host OS and available clipboard utilities;
// callers should surface any returned error to the user via a toast.
type DefaultDeps struct{}

func (DefaultDeps) Now() time.Time { return time.Now() }

func (DefaultDeps) ClipboardWriteAll(text string) error {
	return clipboard.WriteAll(text)
}
