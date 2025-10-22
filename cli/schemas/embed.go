package schemas

import "embed"

// CUESchemas embeds all CUE schema files from the schemas package.
// This allows the schemas to be bundled into the binary and loaded at runtime.
// Includes subdirectories (phases/, projects/) and cue.mod for import resolution.
//
//go:embed *.cue phases/*.cue projects/*.cue cue.mod/module.cue
var CUESchemas embed.FS
