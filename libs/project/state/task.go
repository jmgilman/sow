package state

import (
	"github.com/jmgilman/sow/libs/schemas/project"
)

// Task wraps the CUE-generated TaskState.
// This is a pure data wrapper with no additional runtime fields.
type Task struct {
	project.TaskState
}
