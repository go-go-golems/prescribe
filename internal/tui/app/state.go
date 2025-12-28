package app

import (
	"github.com/go-go-golems/prescribe/internal/controller"
	"github.com/go-go-golems/prescribe/internal/tui/components/filelist"
	"github.com/go-go-golems/prescribe/internal/tui/components/filterpane"
	"github.com/go-go-golems/prescribe/internal/tui/components/result"
	"github.com/go-go-golems/prescribe/internal/tui/components/status"
	"github.com/go-go-golems/prescribe/internal/tui/events"
	"github.com/go-go-golems/prescribe/internal/tui/keys"
	"github.com/go-go-golems/prescribe/internal/tui/layout"
	"github.com/go-go-golems/prescribe/internal/tui/styles"
)

// Mode is the app-level screen/state machine mode.
type Mode int

const (
	ModeMain Mode = iota
	ModeFilters
	ModeGenerating
	ModeResult
)

// Model is the root Bubbletea model for the modular TUI.
//
// Phase 2 goal: get a behavior-compatible root model in place that can later
// delegate list/result/filter logic into components.
type Model struct {
	ctrl *controller.Controller
	deps Deps

	mode Mode

	err error

	// view flags
	showFiltered bool
	showFullHelp bool

	// selection
	selectedIndex int
	filterIndex   int

	// quick preset UX (loaded from preset dirs; used for keymap.Preset{1,2,3})
	filterPresets []events.FilterPresetSummary

	// terminal + layout
	width  int
	height int
	layout layout.Layout

	// generation/result
	generatedDesc string
	result        result.Model
	filelist      filelist.Model
	filterpane    filterpane.Model

	// shared UI primitives
	keymap keys.KeyMap
	styles styles.Styles
	status status.Model
}
