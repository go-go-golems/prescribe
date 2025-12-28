package filelist

import (
	"testing"

	"github.com/go-go-golems/prescribe/internal/domain"
	"github.com/go-go-golems/prescribe/internal/tui/keys"
	"github.com/go-go-golems/prescribe/internal/tui/styles"
)

func TestModel_SetFilesAndSelect(t *testing.T) {
	m := New(keys.Default(), styles.Default())
	m.SetFiles([]domain.FileChange{
		{Path: "a.go", Included: true},
		{Path: "b.go", Included: false},
	})
	m.SetSelectedIndex(1)

	path, ok := m.SelectedPath()
	if !ok || path != "b.go" {
		t.Fatalf("expected selected path b.go, got %q ok=%v", path, ok)
	}
	inc, ok := m.SelectedIncluded()
	if !ok || inc != false {
		t.Fatalf("expected selected included=false, got %v ok=%v", inc, ok)
	}
}
