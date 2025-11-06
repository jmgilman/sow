package exploration

import _ "embed"

// Embedded CUE metadata schemas for phase validation

//go:embed cue/exploration_metadata.cue
var explorationMetadataSchema string

//go:embed cue/finalization_metadata.cue
var finalizationMetadataSchema string
