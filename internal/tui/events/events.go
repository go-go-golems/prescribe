package events

import "time"

// NOTE: This package is intentionally dependency-light.
// It exists to provide a shared, cycle-free message vocabulary between:
// - internal/tui/app (root orchestrator), and
// - internal/tui/components/* (UI-only component models).

// --- Boot/session lifecycle ---------------------------------------------------

// SessionLoadedMsg indicates the default session was loaded successfully.
type SessionLoadedMsg struct{ Path string }

// SessionLoadSkippedMsg indicates session loading was intentionally skipped (e.g. missing file).
type SessionLoadSkippedMsg struct{ Path string }

// SessionLoadFailedMsg indicates session loading failed (YAML error, branch mismatch, etc.).
type SessionLoadFailedMsg struct {
	Path string
	Err  error
}

// SessionSavedMsg indicates the current session was persisted successfully.
type SessionSavedMsg struct{ Path string }

// SessionSaveFailedMsg indicates session persistence failed.
type SessionSaveFailedMsg struct {
	Path string
	Err  error
}

// --- Intents (user actions) ---------------------------------------------------

// ToggleFileIncludedRequested toggles the "included" bit for a file identified by its stable path.
type ToggleFileIncludedRequested struct{ Path string }

// SetAllVisibleIncludedRequested is the canonical "select all / unselect all".
type SetAllVisibleIncludedRequested struct{ Included bool }

// ToggleShowFilteredRequested toggles whether the UI shows filtered files (vs visible files).
type ToggleShowFilteredRequested struct{}

type OpenFiltersRequested struct{}
type CloseFiltersRequested struct{}

type RemoveFilterRequested struct{ Index int }
type ClearFiltersRequested struct{}

// AddFilterPresetRequested requests adding a predefined filter preset.
// PresetID is app-defined (e.g. "exclude-tests", "exclude-docs", ...).
type AddFilterPresetRequested struct{ PresetID string }

type GenerateRequested struct{}

// CopyContextRequested requests exporting the generation context to clipboard.
type CopyContextRequested struct{}

// --- Results (side-effects) ---------------------------------------------------

type DescriptionGeneratedMsg struct{ Text string }
type DescriptionGenerationFailedMsg struct{ Err error }

type ClipboardCopiedMsg struct {
	What  string
	Bytes int
}
type ClipboardCopyFailedMsg struct{ Err error }

// --- UX (toasts) --------------------------------------------------------------

type ToastLevel int

const (
	ToastInfo ToastLevel = iota
	ToastSuccess
	ToastWarning
	ToastError
)

// ShowToastMsg requests displaying a toast for a bounded duration.
type ShowToastMsg struct {
	Text     string
	Level    ToastLevel
	Duration time.Duration
}

// ToastExpiredMsg clears a toast if the ID matches the current toast.
// This prevents an older timer from clearing a newer toast.
type ToastExpiredMsg struct{ ID int64 }


