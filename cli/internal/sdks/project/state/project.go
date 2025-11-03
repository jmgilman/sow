package state

import (
	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/jmgilman/sow/cli/schemas/project"
)

// Project wraps the CUE-generated ProjectState with runtime behavior.
// It embeds ProjectState for serialization and adds runtime-only fields
// that are not persisted (config, machine, and ctx).
type Project struct {
	project.ProjectState

	// Runtime-only fields (not serialized)
	config  *ProjectTypeConfig
	machine *Machine
	ctx     *sow.Context // Context for FS operations
}

// ProjectTypeConfig holds project-type-specific configuration.
// This will be expanded in future tasks to include OnAdvance handlers,
// guards, and other project-type-specific behavior.
type ProjectTypeConfig struct {
	// TODO: Add fields in future tasks
}

// Machine represents the state machine for the project.
// This is a placeholder for the state machine that will be
// integrated in future tasks.
type Machine struct {
	currentState State
}

// State returns the current state of the machine.
// This is a stub that will be implemented in Unit 3.
func (m *Machine) State() State {
	if m == nil {
		return State("")
	}
	return m.currentState
}

// Helper methods for common guard patterns

// PhaseOutputApproved checks if a phase has an approved artifact of the given type.
// Returns false if phase not found, artifact not found, or artifact not approved.
// This is a read-only method used in state machine guards.
func (p *Project) PhaseOutputApproved(phaseName, outputType string) bool {
	phase, exists := p.Phases[phaseName]
	if !exists {
		return false // Phase not found
	}

	for _, artifact := range phase.Outputs {
		if artifact.Type == outputType && artifact.Approved {
			return true
		}
	}

	return false
}

// PhaseMetadataBool reads a boolean value from phase metadata.
// Returns false if phase not found, key not found, or value is not a boolean.
// This is a read-only method used in state machine guards.
func (p *Project) PhaseMetadataBool(phaseName, key string) bool {
	phase, exists := p.Phases[phaseName]
	if !exists {
		return false // Phase not found
	}

	if phase.Metadata == nil {
		return false // No metadata
	}

	value, ok := phase.Metadata[key]
	if !ok {
		return false // Key not found
	}

	boolValue, ok := value.(bool)
	if !ok {
		return false // Wrong type
	}

	return boolValue
}

// AllTasksComplete checks if all tasks across all phases are completed.
// Returns true if all tasks have status "completed", or if no tasks exist (vacuous truth).
// This is a read-only method used in state machine guards.
func (p *Project) AllTasksComplete() bool {
	for _, phase := range p.Phases {
		for _, task := range phase.Tasks {
			if task.Status != "completed" {
				return false
			}
		}
	}
	return true // All tasks completed (or no tasks exist)
}
