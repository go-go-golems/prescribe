package app

import (
	"sort"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/prescribe/internal/controller"
	"github.com/go-go-golems/prescribe/internal/tui/events"
)

func loadFilterPresetsCmd(ctrl *controller.Controller) tea.Cmd {
	return func() tea.Msg {
		project, err := ctrl.LoadProjectFilterPresets()
		if err != nil {
			return events.FilterPresetsLoadFailedMsg{Err: err}
		}
		global, err := ctrl.LoadGlobalFilterPresets()
		if err != nil {
			return events.FilterPresetsLoadFailedMsg{Err: err}
		}

		presets := make([]events.FilterPresetSummary, 0, len(project)+len(global))
		for _, p := range project {
			presets = append(presets, events.FilterPresetSummary{
				ID:          p.ID,
				Name:        p.Name,
				Description: p.Description,
				Location:    "project",
			})
		}
		for _, p := range global {
			presets = append(presets, events.FilterPresetSummary{
				ID:          p.ID,
				Name:        p.Name,
				Description: p.Description,
				Location:    "global",
			})
		}

		// Stable ordering: project first, then global; within each, sort by Name then ID.
		sort.SliceStable(presets, func(i, j int) bool {
			if presets[i].Location != presets[j].Location {
				return presets[i].Location < presets[j].Location
			}
			if presets[i].Name != presets[j].Name {
				return presets[i].Name < presets[j].Name
			}
			return presets[i].ID < presets[j].ID
		})

		return events.FilterPresetsLoadedMsg{Presets: presets}
	}
}
