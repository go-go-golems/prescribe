package cmds

import "testing"

func TestResolveCreateBase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		explicitBase  string
		sessionTarget string
		want          string
	}{
		{
			name:          "explicit base wins",
			explicitBase:  "main",
			sessionTarget: "develop",
			want:          "main",
		},
		{
			name:          "fallback to session target when empty",
			explicitBase:  "",
			sessionTarget: "develop",
			want:          "develop",
		},
		{
			name:          "trim whitespace",
			explicitBase:  "  ",
			sessionTarget: " master ",
			want:          "master",
		},
		{
			name:          "explicit base trims whitespace",
			explicitBase:  "  release ",
			sessionTarget: "develop",
			want:          "release",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := resolveCreateBase(tt.explicitBase, tt.sessionTarget)
			if got != tt.want {
				t.Fatalf("resolveCreateBase(%q, %q) = %q, want %q", tt.explicitBase, tt.sessionTarget, got, tt.want)
			}
		})
	}
}
