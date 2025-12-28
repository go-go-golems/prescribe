package result

import "testing"

func TestModel_SetSize_clamps(t *testing.T) {
	m := New()
	m.SetSize(-1, -2)
	// Just ensure we don't panic and we clamp to non-negative.
	if m.vp.Width != 0 || m.vp.Height != 0 {
		t.Fatalf("expected 0x0 viewport, got %dx%d", m.vp.Width, m.vp.Height)
	}
}
