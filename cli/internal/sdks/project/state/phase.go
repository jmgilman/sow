package state

import (
	"github.com/jmgilman/sow/cli/schemas/project"
)

// Phase wraps the CUE-generated PhaseState.
// This is a pure data wrapper with no additional runtime fields.
type Phase struct {
	project.PhaseState
}
