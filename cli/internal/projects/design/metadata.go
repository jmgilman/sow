package design

import _ "embed"

// Embedded CUE metadata schemas for phase validation

//go:embed cue/design_metadata.cue
var designMetadataSchema string

//go:embed cue/finalization_metadata.cue
var finalizationMetadataSchema string
