package schemas

import "embed"

// CUESchemas embeds all CUE schema files from the schemas package.
// This allows the schemas to be bundled into the binary and loaded at runtime.
//
//go:embed *.cue
var CUESchemas embed.FS
