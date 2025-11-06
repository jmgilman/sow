package breakdown

import _ "embed"

// Embedded CUE metadata schemas for phase validation

//go:embed cue/breakdown_metadata.cue
var breakdownMetadataSchema string
