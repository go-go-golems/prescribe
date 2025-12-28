package layout

// Layout captures the computed sizes for the TUI.
//
// The design intent is:
// - root model stores terminal Width/Height
// - header/footer heights are derived (measured) from rendered chrome
// - body sizes are the remainder and are pushed into child components.
//
// This struct is deliberately small; we can extend it later if we add split panes.
type Layout struct {
	Width, Height int

	HeaderH int
	FooterH int

	BodyW int
	BodyH int
}

// Compute derives a stable Layout for a given window size and chrome heights.
//
// Invariants:
// - BodyH is never negative
// - BodyW equals Width (single-pane body, for now)
func Compute(width, height, headerH, footerH int) Layout {
	bodyH := height - headerH - footerH
	if bodyH < 0 {
		bodyH = 0
	}

	return Layout{
		Width:   width,
		Height:  height,
		HeaderH: headerH,
		FooterH: footerH,
		BodyW:   width,
		BodyH:   bodyH,
	}
}
