package api

import "testing"

func TestParseGeneratedPRDataFromAssistantText_prefersLastYAMLBlock(t *testing.T) {
	in := "Some analysis\n\n```yaml\ntitle: wrong\nbody: |\n  nope\n```\n\nMore text\n\n```yaml\ntitle: right\nbody: |\n  yep\nchangelog: |\n  Add thing\nrelease_notes:\n  title: rn\n  body: |\n    rn body\n```\n"

	got, err := ParseGeneratedPRDataFromAssistantText(in)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got.Title != "right" {
		t.Fatalf("expected title=right, got %q", got.Title)
	}
	if got.ReleaseNotes == nil || got.ReleaseNotes.Title != "rn" {
		t.Fatalf("expected release_notes.title=rn, got %#v", got.ReleaseNotes)
	}
}

func TestParseGeneratedPRDataFromAssistantText_fallback_stripsFence(t *testing.T) {
	in := "```yaml\ntitle: ok\nbody: |\n  hi\nchangelog: |\n  c\n```"
	got, err := ParseGeneratedPRDataFromAssistantText(in)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got.Title != "ok" {
		t.Fatalf("expected title=ok, got %q", got.Title)
	}
}
