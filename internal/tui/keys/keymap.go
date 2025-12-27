package keys

import (
	"github.com/charmbracelet/bubbles/key"
)

// KeyMap defines the centralized key bindings for the new modular TUI.
//
// It is designed to be used with bubbles/help.Model (implements the KeyMap shape it expects).
// Keep this aligned with the existing key contract in internal/tui/model_enhanced.go.
type KeyMap struct {
	// Global / app-level
	Quit key.Binding
	Help key.Binding

	// Navigation
	Up   key.Binding
	Down key.Binding

	// Main screen actions
	ToggleIncluded      key.Binding
	ToggleFilteredView  key.Binding
	OpenFilters         key.Binding
	Generate            key.Binding
	Back                key.Binding
	SelectAllVisible    key.Binding
	UnselectAllVisible  key.Binding
	CopyContext         key.Binding

	// Filter screen actions
	DeleteFilter key.Binding
	ClearFilters key.Binding
	Preset1      key.Binding
	Preset2      key.Binding
	Preset3      key.Binding
}

func Default() KeyMap {
	return KeyMap{
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		ToggleIncluded: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "toggle"),
		),
		ToggleFilteredView: key.NewBinding(
			key.WithKeys("v"),
			key.WithHelp("v", "view filtered"),
		),
		OpenFilters: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "filters"),
		),
		Generate: key.NewBinding(
			key.WithKeys("g"),
			key.WithHelp("g", "generate"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		SelectAllVisible: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "select all"),
		),
		UnselectAllVisible: key.NewBinding(
			key.WithKeys("A"),
			key.WithHelp("A", "unselect all"),
		),
		CopyContext: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "copy context"),
		),
		DeleteFilter: key.NewBinding(
			key.WithKeys("d", "x"),
			key.WithHelp("d", "delete filter"),
		),
		ClearFilters: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "clear filters"),
		),
		Preset1: key.NewBinding(
			key.WithKeys("1"),
			key.WithHelp("1", "preset 1"),
		),
		Preset2: key.NewBinding(
			key.WithKeys("2"),
			key.WithHelp("2", "preset 2"),
		),
		Preset3: key.NewBinding(
			key.WithKeys("3"),
			key.WithHelp("3", "preset 3"),
		),
	}
}

// ShortHelp returns keybindings to show in the condensed help view.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Up, k.Down,
		k.ToggleIncluded,
		k.OpenFilters,
		k.Generate,
		k.Quit,
	}
}

// FullHelp returns keybindings to show in the expanded help view.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.ToggleIncluded, k.ToggleFilteredView},
		{k.OpenFilters, k.Back, k.Generate, k.CopyContext},
		{k.SelectAllVisible, k.UnselectAllVisible},
		{k.DeleteFilter, k.ClearFilters, k.Preset1, k.Preset2, k.Preset3},
		{k.Help, k.Quit},
	}
}


