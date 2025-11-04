package standard

import _ "embed"

// Embedded CUE metadata schemas for phase validation

//go:embed cue/implementation_metadata.cue
var implementationMetadataSchema string

//go:embed cue/review_metadata.cue
var reviewMetadataSchema string

//go:embed cue/finalize_metadata.cue
var finalizeMetadataSchema string
