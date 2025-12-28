package doc

import "embed"

// FS embeds the Prescribe help/documentation markdown sections.
//
// Files are loaded into the Glazed help system during CLI initialization.
// See cmd/prescribe/main.go.
//
//go:embed topics/*.md
var FS embed.FS
