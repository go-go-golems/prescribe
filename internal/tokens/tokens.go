package tokens

import (
	"os"
	"strings"
	"sync"

	"github.com/tiktoken-go/tokenizer"
)

// Count returns the token count for s using a real tokenizer (tiktoken-style BPE).
//
// This is intended for "preflight" budgeting in the prescribe TUI, i.e. counting
// tokens before calling any provider. It is not provider-billing-authoritative.
//
// Encoding selection:
// - default: cl100k_base
// - override with PRESCRIBE_TOKEN_ENCODING (supported: cl100k_base, o200k_base, r50k_base, p50k_base, p50k_edit)
func Count(s string) int {
	codec := getCodec()
	if codec == nil {
		// Fallback: keep old behavior if tokenizer init fails for any reason.
		return len(s) / 4
	}

	n, err := codec.Count(s)
	if err != nil {
		// Best-effort fallback; we don't want token counting failures to break the UI.
		return len(s) / 4
	}
	return n
}

func EncodingName() string {
	_, name := getCodecAndName()
	return name
}

var (
	codecOnce sync.Once
	codec     tokenizer.Codec
	codecName string
)

func getCodec() tokenizer.Codec {
	c, _ := getCodecAndName()
	return c
}

func getCodecAndName() (tokenizer.Codec, string) {
	codecOnce.Do(func() {
		enc := tokenizer.Cl100kBase
		if v, ok := os.LookupEnv("PRESCRIBE_TOKEN_ENCODING"); ok {
			switch strings.TrimSpace(strings.ToLower(v)) {
			case string(tokenizer.O200kBase):
				enc = tokenizer.O200kBase
			case string(tokenizer.Cl100kBase):
				enc = tokenizer.Cl100kBase
			case string(tokenizer.R50kBase):
				enc = tokenizer.R50kBase
			case string(tokenizer.P50kBase):
				enc = tokenizer.P50kBase
			case string(tokenizer.P50kEdit):
				enc = tokenizer.P50kEdit
			}
		}

		c, err := tokenizer.Get(enc)
		if err != nil {
			// Leave codec nil; Count will fall back to heuristic.
			codec = nil
			codecName = ""
			return
		}
		codec = c
		codecName = c.GetName()
	})
	return codec, codecName
}


