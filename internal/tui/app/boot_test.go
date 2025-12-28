package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-go-golems/prescribe/internal/controller"
	"github.com/go-go-golems/prescribe/internal/tui/events"
)

func TestBootCmd_MissingSession_Fails(t *testing.T) {
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}

	ctrl, err := controller.NewController(repo)
	if err != nil {
		t.Fatalf("NewController: %v", err)
	}

	msg := bootCmd(ctrl)()
	failed, ok := msg.(events.SessionLoadFailedMsg)
	if !ok {
		t.Fatalf("expected SessionLoadFailedMsg, got %T", msg)
	}

	if failed.Path == "" {
		t.Fatalf("expected Path to be set")
	}
}
