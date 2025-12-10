package state

import (
	"github.com/jmgilman/sow/libs/schemas/project"
)

// Artifact wraps the CUE-generated ArtifactState.
// This is a pure data wrapper with no additional runtime fields.
type Artifact struct {
	project.ArtifactState
}
