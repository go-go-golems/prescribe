package layout

import "testing"

func TestCompute_BodyNeverNegative(t *testing.T) {
	l := Compute(80, 10, 8, 5)
	if l.BodyH != 0 {
		t.Fatalf("expected BodyH=0, got %d", l.BodyH)
	}
}

func TestCompute_PreservesDimensions(t *testing.T) {
	l := Compute(123, 45, 2, 3)
	if l.Width != 123 || l.Height != 45 {
		t.Fatalf("expected width/height 123/45, got %d/%d", l.Width, l.Height)
	}
	if l.BodyW != 123 {
		t.Fatalf("expected BodyW=123, got %d", l.BodyW)
	}
	if l.BodyH != 40 {
		t.Fatalf("expected BodyH=40, got %d", l.BodyH)
	}
}
