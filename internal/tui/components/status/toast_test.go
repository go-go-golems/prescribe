package status

import (
	"testing"

	"github.com/go-go-golems/prescribe/internal/tui/events"
)

func TestToastState_ExpireIsIdSafe(t *testing.T) {
	var s ToastState

	id1 := s.Show("first", events.ToastInfo)
	id2 := s.Show("second", events.ToastInfo)

	if s.Current() == nil || s.Current().ID != id2 {
		t.Fatalf("expected current toast to be id2")
	}

	// Expiring the older toast must not clear the newer toast.
	if cleared := s.Expire(id1); cleared {
		t.Fatalf("expected old id not to clear current toast")
	}
	if s.Current() == nil || s.Current().ID != id2 {
		t.Fatalf("expected current toast to still be id2")
	}

	// Expiring the current id should clear.
	if cleared := s.Expire(id2); !cleared {
		t.Fatalf("expected current id to clear toast")
	}
	if s.Current() != nil {
		t.Fatalf("expected toast to be cleared")
	}
}
