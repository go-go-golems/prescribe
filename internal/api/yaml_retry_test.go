package api

import (
	"testing"

	"github.com/go-go-golems/geppetto/pkg/steps/ai/settings"
)

func TestIsLikelyMaxTokensStopReason(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"", false},
		{"stop", false},
		{"max_tokens", true},
		{"MAX_TOKENS", true},
		{"FinishReasonMaxTokens", true},
		{"finishreasonmaxtokens", true},
		{"Max Tokens", true},
	}

	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			if got := isLikelyMaxTokensStopReason(tc.in); got != tc.want {
				t.Fatalf("expected %v, got %v for %q", tc.want, got, tc.in)
			}
		})
	}
}

func TestComputeRetryMaxResponseTokens(t *testing.T) {
	mk := func(v *int) *settings.StepSettings {
		ss, err := settings.NewStepSettings()
		if err != nil {
			t.Fatalf("NewStepSettings: %v", err)
		}
		ss.Chat.MaxResponseTokens = v
		return ss
	}

	t.Run("nil max tokens -> 2048", func(t *testing.T) {
		ss := mk(nil)
		if got := computeRetryMaxResponseTokens(ss); got != 2048 {
			t.Fatalf("expected 2048, got %d", got)
		}
	})

	t.Run("tiny -> 2048", func(t *testing.T) {
		v := 20
		ss := mk(&v)
		if got := computeRetryMaxResponseTokens(ss); got != 2048 {
			t.Fatalf("expected 2048, got %d", got)
		}
	})

	t.Run("reasonable -> 4x", func(t *testing.T) {
		v := 512
		ss := mk(&v)
		if got := computeRetryMaxResponseTokens(ss); got != 2048 {
			t.Fatalf("expected 2048, got %d", got)
		}
	})

	t.Run("cap at 8192", func(t *testing.T) {
		v := 3000
		ss := mk(&v)
		if got := computeRetryMaxResponseTokens(ss); got != 8192 {
			t.Fatalf("expected 8192, got %d", got)
		}
	})
}

func TestParseAndValidateGeneratedPRData(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		in := "title: ok\nbody: |\n  hi\nchangelog: |\n  c\nrelease_notes:\n  title: rn\n  body: |\n    rn body\n"
		_, errStr := parseAndValidateGeneratedPRData(in)
		if errStr != "" {
			t.Fatalf("expected no error, got %q", errStr)
		}
	})

	t.Run("missing body treated as invalid", func(t *testing.T) {
		in := "title: ok\nbody: \"\"\nchangelog: |\n  c\n"
		_, errStr := parseAndValidateGeneratedPRData(in)
		if errStr == "" {
			t.Fatalf("expected error, got empty")
		}
	})
}
