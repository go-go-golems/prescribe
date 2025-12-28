package filterpane

import (
	"testing"

	"github.com/go-go-golems/prescribe/internal/domain"
	"github.com/go-go-golems/prescribe/internal/tui/keys"
	"github.com/go-go-golems/prescribe/internal/tui/styles"
)

func TestModel_SetFiltersAndSelect(t *testing.T) {
	m := New(keys.Default(), styles.Default())
	m.SetFilters([]domain.Filter{
		{Name: "A"},
		{Name: "B"},
	})
	m.SetSelectedIndex(1)

	f, ok := m.SelectedFilter()
	if !ok || f.Name != "B" {
		t.Fatalf("expected selected filter B, got %#v ok=%v", f, ok)
	}
}
