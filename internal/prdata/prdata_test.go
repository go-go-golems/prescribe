package prdata

import (
	"path/filepath"
	"testing"

	"github.com/go-go-golems/prescribe/internal/domain"
)

func TestWriteThenLoadGeneratedPRDataYAML(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "last-generated-pr.yaml")

	in := &domain.GeneratedPRData{
		Title: "T",
		Body:  "B",
	}

	if err := WriteGeneratedPRDataToYAMLFile(p, in); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	out, err := LoadGeneratedPRDataFromYAMLFile(p)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if out.Title != in.Title {
		t.Fatalf("title mismatch: got %q want %q", out.Title, in.Title)
	}
	if out.Body != in.Body {
		t.Fatalf("body mismatch: got %q want %q", out.Body, in.Body)
	}
}
