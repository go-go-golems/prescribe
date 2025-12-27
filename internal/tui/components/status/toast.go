package status

import "github.com/go-go-golems/prescribe/internal/tui/events"

// Toast represents a transient status message (\"help bubble\") rendered in the footer area.
type Toast struct {
	ID    int64
	Text  string
	Level events.ToastLevel
}

// ToastState holds the current toast and enforces ID-safe expiry.
//
// The ID is critical: it prevents a timer from clearing a newer toast.
type ToastState struct {
	current *Toast
	nextID  int64
}

func (s *ToastState) Current() *Toast {
	return s.current
}

func (s *ToastState) Show(text string, level events.ToastLevel) int64 {
	s.nextID++
	id := s.nextID
	s.current = &Toast{
		ID:    id,
		Text:  text,
		Level: level,
	}
	return id
}

// Expire clears the current toast only if id matches the current toast ID.
// It returns true if a toast was cleared.
func (s *ToastState) Expire(id int64) bool {
	if s.current == nil {
		return false
	}
	if s.current.ID != id {
		return false
	}
	s.current = nil
	return true
}


